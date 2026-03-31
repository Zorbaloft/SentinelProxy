package proxy

import (
	"context"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type Proxy struct {
	upstreamURL *url.URL
	upstreamHost string
	client      *http.Client
}

func New(upstreamURL, upstreamHost string) (*Proxy, error) {
	u, err := url.Parse(upstreamURL)
	if err != nil {
		return nil, err
	}

	return &Proxy{
		upstreamURL: u,
		upstreamHost: upstreamHost,
		client: &http.Client{
			Timeout: 30 * time.Second,
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				return http.ErrUseLastResponse
			},
		},
	}, nil
}

func (p *Proxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Build upstream URL
	targetURL := *p.upstreamURL
	targetURL.Path = r.URL.Path
	targetURL.RawQuery = r.URL.RawQuery

	// Create request
	req, err := http.NewRequest(r.Method, targetURL.String(), r.Body)
	if err != nil {
		http.Error(w, "Bad Gateway", http.StatusBadGateway)
		return
	}

	// Copy headers
	for key, values := range r.Header {
		// Skip hop-by-hop headers
		if key == "Connection" || key == "Keep-Alive" || key == "Proxy-Authenticate" ||
			key == "Proxy-Authorization" || key == "Te" || key == "Trailers" ||
			key == "Transfer-Encoding" || key == "Upgrade" {
			continue
		}
		for _, value := range values {
			req.Header.Add(key, value)
		}
	}

	// Set X-Forwarded-For
	clientIP := r.Context().Value("clientIP").(string)
	if existing := req.Header.Get("X-Forwarded-For"); existing != "" {
		req.Header.Set("X-Forwarded-For", existing+", "+clientIP)
	} else {
		req.Header.Set("X-Forwarded-For", clientIP)
	}

	// Set X-Request-ID
	requestID := r.Context().Value("requestID").(string)
	req.Header.Set("X-Request-ID", requestID)

	// Determine Host header: preserve incoming Host if it matches domain pattern, otherwise use default
	incomingHost := r.Host
	if incomingHost == "" {
		incomingHost = r.Header.Get("Host")
	}
	
	// Check if incoming Host matches the domain pattern (*.3cket.local or 3cket.local)
	hostToUse := p.upstreamHost
	if incomingHost != "" {
		// Remove port if present (e.g., "api.3cket.local:443" -> "api.3cket.local")
		hostWithoutPort := strings.Split(incomingHost, ":")[0]
		// Check if it ends with .3cket.local or is exactly 3cket.local
		if strings.HasSuffix(hostWithoutPort, ".3cket.local") || hostWithoutPort == "3cket.local" {
			hostToUse = hostWithoutPort
		}
	}
	
	// Set Host header
	req.Host = hostToUse
	req.Header.Set("Host", hostToUse)

	// Execute request
	resp, err := p.client.Do(req)
	if err != nil {
		http.Error(w, "Bad Gateway", http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	// Copy response headers (before WriteHeader)
	for key, values := range resp.Header {
		// Skip hop-by-hop headers
		if key == "Connection" || key == "Keep-Alive" || key == "Transfer-Encoding" {
			continue
		}
		for _, value := range values {
			w.Header().Add(key, value)
		}
	}

	// Set status code (must be before body copy)
	w.WriteHeader(resp.StatusCode)

	// Copy response body
	io.Copy(w, resp.Body)

	// Store response info in context for logger
	r = r.WithContext(r.Context())
	if r.Context().Value("responseStatus") == nil {
		r = r.WithContext(context.WithValue(r.Context(), "responseStatus", resp.StatusCode))
	}
}
