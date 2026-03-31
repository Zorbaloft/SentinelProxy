package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/sentinel-proxy/engine/internal/config"
	"github.com/sentinel-proxy/engine/internal/guard"
	loggerpkg "github.com/sentinel-proxy/engine/internal/logger"
	"github.com/sentinel-proxy/engine/internal/proxy"
	"github.com/sentinel-proxy/engine/internal/ratelimit"
	"github.com/sentinel-proxy/engine/internal/util"
)

func main() {
	cfg := config.Load()

	// Initialize guard
	g, err := guard.New(cfg.RedisAddr)
	if err != nil {
		log.Fatalf("Failed to initialize guard: %v", err)
	}

	// Initialize rate limiter
	rl, err := ratelimit.New(cfg.RedisAddr, cfg.RateCounterTTL)
	if err != nil {
		log.Fatalf("Failed to initialize rate limiter: %v", err)
	}

	// Initialize proxy
	prx, err := proxy.New(cfg.UpstreamURL, cfg.UpstreamHost)
	if err != nil {
		log.Fatalf("Failed to initialize proxy: %v", err)
	}

	// Initialize logger
	loggerInstance, err := loggerpkg.New(cfg.MongoURI, cfg.LogBodyMaxBytes)
	if err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}

	// Ensure TTL index
	if err := loggerInstance.EnsureTTLIndex(cfg.LogTTLDays); err != nil {
		log.Printf("Warning: Failed to create TTL index: %v", err)
	}

	// Build middleware chain
	mux := http.NewServeMux()

	// Health check endpoint
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// Main handler with middleware pipeline
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Extract client IP and store in context
		clientIP := util.ExtractClientIP(r)
		r = r.WithContext(context.WithValue(r.Context(), "clientIP", clientIP))

		// Get or create request ID
		requestID := util.GetOrCreateRequestID(r)
		r = r.WithContext(context.WithValue(r.Context(), "requestID", requestID))
		r.Header.Set("X-Request-ID", requestID)

		// Wrap response writer to capture response
		rw := loggerpkg.NewResponseWriter(w)
		startTime := time.Now()

		// Middleware pipeline:
		// 1. Guard (blocklist/redirect check)
		guardHandler := g.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// 2. Rate limiting (counters)
			rateHandler := rl.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// 3. Proxy upstream
				prx.ServeHTTP(w, r)

				// 4. Increment status counter
				if rw, ok := w.(*loggerpkg.ResponseWriter); ok {
					rl.IncrementStatusCounter(r.Context(), clientIP, r.URL.Path, rw.StatusCode())
				}
			}))
			rateHandler.ServeHTTP(w, r)
			
			// Log request/response
			if rw, ok := w.(*loggerpkg.ResponseWriter); ok {
				loggerInstance.LogRequestResponse(r, rw, startTime, cfg.UpstreamURL)
			}
		}))
		guardHandler.ServeHTTP(rw, r)

		// If guard blocked/redirected, log using outer rw
		if rw.StatusCode() == http.StatusForbidden || rw.StatusCode() == http.StatusFound {
			loggerInstance.LogRequestResponse(r, rw, startTime, cfg.UpstreamURL)
		}
	})

	mux.Handle("/", handler)

	// Start server
	addr := ":" + cfg.ServerPort
	srv := &http.Server{
		Addr:    addr,
		Handler: mux,
	}

	go func() {
		log.Printf("Sentinel Engine listening on %s", addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed: %v", err)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down...")
	if err := srv.Shutdown(context.Background()); err != nil {
		log.Fatalf("Server shutdown failed: %v", err)
	}
	log.Println("Server stopped")
}
