package ratelimit

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/go-redis/redis/v8"
)

type RateLimiter struct {
	redis *redis.Client
	ttl   time.Duration
}

func New(redisAddr string, ttl time.Duration) (*RateLimiter, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr: redisAddr,
	})
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("redis connection failed: %w", err)
	}
	return &RateLimiter{redis: rdb, ttl: ttl}, nil
}

func (rl *RateLimiter) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := r.Context().Value("clientIP").(string)
		method := r.Method
		path := r.URL.Path

		ctx := r.Context()

		// Increment counters
		key1 := fmt.Sprintf("rate:%s:%s:%s", ip, method, path)
		key2 := fmt.Sprintf("rate:%s:%s", ip, path)

		rl.redis.Incr(ctx, key1)
		rl.redis.Expire(ctx, key1, rl.ttl)

		rl.redis.Incr(ctx, key2)
		rl.redis.Expire(ctx, key2, rl.ttl)

		// Store in context for later use (status-based counters)
		r = r.WithContext(context.WithValue(ctx, "rateKeys", []string{key1, key2}))

		next.ServeHTTP(w, r)
	})
}

// IncrementStatusCounter increments status-based counters (called after response)
func (rl *RateLimiter) IncrementStatusCounter(ctx context.Context, ip, path string, statusCode int) {
	statusClass := getStatusClass(statusCode)
	key := fmt.Sprintf("rate_status:%s:%s:%s", ip, path, statusClass)
	rl.redis.Incr(ctx, key)
	rl.redis.Expire(ctx, key, rl.ttl)
}

func getStatusClass(code int) string {
	switch {
	case code >= 200 && code < 300:
		return "2xx"
	case code >= 300 && code < 400:
		return "3xx"
	case code >= 400 && code < 500:
		return "4xx"
	case code >= 500:
		return "5xx"
	default:
		return "unknown"
	}
}
