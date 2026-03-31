package guard

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/go-redis/redis/v8"
)

type Guard struct {
	redis *redis.Client
}

type BlockEntry struct {
	Reason    string    `json:"reason"`
	CreatedAt time.Time `json:"createdAt"`
	ExpiresAt time.Time `json:"expiresAt"`
	RuleID    string    `json:"ruleId,omitempty"`
}

type RedirectEntry struct {
	TargetURL string    `json:"targetUrl"`
	CreatedAt time.Time `json:"createdAt"`
	ExpiresAt time.Time `json:"expiresAt"`
	RuleID    string    `json:"ruleId,omitempty"`
}

func New(redisAddr string) (*Guard, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr: redisAddr,
	})
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("redis connection failed: %w", err)
	}
	return &Guard{redis: rdb}, nil
}

func (g *Guard) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := r.Context().Value("clientIP").(string)

		// Check blocklist
		blockKey := fmt.Sprintf("blocklist:%s", ip)
		blockData, err := g.redis.Get(r.Context(), blockKey).Result()
		if err == nil {
			var block BlockEntry
			if json.Unmarshal([]byte(blockData), &block) == nil {
				w.Header().Set("Content-Type", "text/html")
				w.WriteHeader(http.StatusForbidden)
				w.Write([]byte(`<html><body><h1>403 Forbidden</h1><p>Access denied.</p></body></html>`))
				return
			}
		}

		// Check redirect
		redirectKey := fmt.Sprintf("redirect:%s", ip)
		redirectData, err := g.redis.Get(r.Context(), redirectKey).Result()
		if err == nil {
			var redirect RedirectEntry
			if json.Unmarshal([]byte(redirectData), &redirect) == nil {
				targetURL := redirect.TargetURL
				// If targetURL doesn't include path, preserve original path/query
				if targetURL == "" || (!containsPath(targetURL) && r.URL.Path != "") {
					if targetURL == "" {
						targetURL = r.URL.Scheme + "://" + r.Host
					}
					targetURL = targetURL + r.URL.Path
					if r.URL.RawQuery != "" {
						targetURL = targetURL + "?" + r.URL.RawQuery
					}
				}
				w.Header().Set("Location", targetURL)
				w.WriteHeader(http.StatusFound)
				return
			}
		}

		next.ServeHTTP(w, r)
	})
}

func containsPath(url string) bool {
	return len(url) > 0 && (strings.Contains(url, "/") || strings.Contains(url, "?"))
}
