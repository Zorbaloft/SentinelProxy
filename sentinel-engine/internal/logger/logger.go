package logger

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type LogEntry struct {
	RequestTime  time.Time              `bson:"requestTime"`
	ResponseTime time.Time              `bson:"responseTime"`
	DurationMs   int64                  `bson:"durationMs"`
	Timestamp    time.Time              `bson:"timestamp"` // For TTL index
	Request      RequestData            `bson:"request"`
	Response     ResponseData           `bson:"response"`
	Client       ClientData             `bson:"client"`
	Meta         MetaData               `bson:"meta"`
}

type RequestData struct {
	Method    string            `bson:"method"`
	Scheme    string            `bson:"scheme"`
	Host      string            `bson:"host"`
	Path      string            `bson:"path"`
	Query     string            `bson:"query"`
	Headers   map[string]string `bson:"headers"`
	Body      string            `bson:"body,omitempty"`
	BodyHash  string            `bson:"bodyHash,omitempty"`
	BodySize  int64             `bson:"bodySize"`
	Truncated bool              `bson:"truncated"`
}

type ResponseData struct {
	Status     int               `bson:"status"`
	Headers    map[string]string `bson:"headers"`
	Body       string            `bson:"body,omitempty"`
	BodyHash   string            `bson:"bodyHash,omitempty"`
	BodySize   int64             `bson:"bodySize"`
	Truncated  bool              `bson:"truncated"`
	BodyPrefix string            `bson:"bodyPrefix,omitempty"`
}

type ClientData struct {
	IP       string `bson:"ip"`
	UserAgent string `bson:"userAgent"`
	Referer  string `bson:"referer"`
}

type MetaData struct {
	RequestID string `bson:"requestId"`
	Upstream  string `bson:"upstream"`
	Error     string `bson:"error,omitempty"`
}

type Logger struct {
	client      *mongo.Client
	collection  *mongo.Collection
	logChan     chan *LogEntry
	bufferSize  int
	dropCount   int64
	bodyMaxBytes int64
}

var redactedHeaders = map[string]bool{
	"authorization": true,
	"cookie":        true,
	"set-cookie":    true,
}

func New(mongoURI string, bodyMaxBytes int64) (*Logger, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(mongoURI))
	if err != nil {
		return nil, fmt.Errorf("mongo connection failed: %w", err)
	}

	db := client.Database("sentinel")
	collection := db.Collection("logs")

	logger := &Logger{
		client:      client,
		collection:  collection,
		logChan:     make(chan *LogEntry, 1000),
		bufferSize:  1000,
		bodyMaxBytes: bodyMaxBytes,
	}

	// Start worker goroutines
	for i := 0; i < 5; i++ {
		go logger.worker()
	}

	return logger, nil
}

func (l *Logger) LogRequestResponse(r *http.Request, w *ResponseWriter, startTime time.Time, upstreamURL string) {
	entry := &LogEntry{
		RequestTime:  startTime,
		ResponseTime: time.Now(),
		Timestamp:    startTime, // For TTL index
		DurationMs:   time.Since(startTime).Milliseconds(),
	}

	// Request data
	entry.Request = RequestData{
		Method:    r.Method,
		Scheme:    r.URL.Scheme,
		Host:      r.Host,
		Path:      r.URL.Path,
		Query:     r.URL.RawQuery,
		Headers:   redactHeaders(r.Header),
		BodySize:  r.ContentLength,
	}

	// Read request body
	if r.Body != nil {
		bodyBytes, err := io.ReadAll(r.Body)
		if err == nil {
			r.Body = io.NopCloser(bytes.NewReader(bodyBytes))
			entry.Request.BodySize = int64(len(bodyBytes))
			if int64(len(bodyBytes)) <= l.bodyMaxBytes {
				entry.Request.Body = string(bodyBytes)
			} else {
				hash := sha256.Sum256(bodyBytes)
				entry.Request.BodyHash = fmt.Sprintf("%x", hash)
				entry.Request.Truncated = true
			}
		}
	}

	// Response data
	entry.Response = ResponseData{
		Status:    w.statusCode,
		Headers:   redactHeaders(w.Header()),
		BodySize:  int64(len(w.body)),
	}

	if int64(len(w.body)) <= l.bodyMaxBytes {
		entry.Response.Body = string(w.body)
	} else {
		hash := sha256.Sum256(w.body)
		entry.Response.BodyHash = fmt.Sprintf("%x", hash)
		entry.Response.Truncated = true
		// Store prefix (up to 4096 bytes)
		prefixLen := 4096
		if len(w.body) < prefixLen {
			prefixLen = len(w.body)
		}
		entry.Response.BodyPrefix = base64.StdEncoding.EncodeToString(w.body[:prefixLen])
	}

	// Client data
	entry.Client = ClientData{
		IP:       r.Context().Value("clientIP").(string),
		UserAgent: r.Header.Get("User-Agent"),
		Referer:  r.Header.Get("Referer"),
	}

	// Meta data
	entry.Meta = MetaData{
		RequestID: r.Context().Value("requestID").(string),
		Upstream:  upstreamURL,
	}

	// Non-blocking send
	select {
	case l.logChan <- entry:
	default:
		// Buffer full, drop log
		l.dropCount++
	}
}

func (l *Logger) worker() {
	for entry := range l.logChan {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		_, err := l.collection.InsertOne(ctx, entry)
		cancel()
		if err != nil {
			// Log error but don't crash
			fmt.Printf("Failed to insert log: %v\n", err)
		}
	}
}

func redactHeaders(headers http.Header) map[string]string {
	result := make(map[string]string)
	for key, values := range headers {
		lowerKey := http.CanonicalHeaderKey(key)
		if redactedHeaders[lowerKey] {
			result[key] = "[REDACTED]"
		} else {
			result[key] = values[0] // Take first value
		}
	}
	return result
}

// EnsureTTLIndex creates TTL index on timestamp field (idempotent)
func (l *Logger) EnsureTTLIndex(ttlDays int) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	indexModel := mongo.IndexModel{
		Keys: map[string]interface{}{
			"timestamp": 1,
		},
		Options: options.Index().SetExpireAfterSeconds(int32(ttlDays * 86400)),
	}

	_, err := l.collection.Indexes().CreateOne(ctx, indexModel)
	if err != nil {
		// Check if index already exists
		if mongo.IsDuplicateKeyError(err) {
			return nil
		}
		return fmt.Errorf("failed to create TTL index: %w", err)
	}
	return nil
}

// ResponseWriter wraps http.ResponseWriter to capture response
type ResponseWriter struct {
	http.ResponseWriter
	statusCode int
	body       []byte
}

func NewResponseWriter(w http.ResponseWriter) *ResponseWriter {
	return &ResponseWriter{
		ResponseWriter: w,
		statusCode:     http.StatusOK,
		body:           make([]byte, 0),
	}
}

func (rw *ResponseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func (rw *ResponseWriter) Write(b []byte) (int, error) {
	rw.body = append(rw.body, b...)
	return rw.ResponseWriter.Write(b)
}

func (rw *ResponseWriter) StatusCode() int {
	return rw.statusCode
}
