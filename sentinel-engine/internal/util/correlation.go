package util

import (
	"net/http"

	"github.com/google/uuid"
)

// GetOrCreateRequestID extracts X-Request-ID or generates a new UUID
func GetOrCreateRequestID(r *http.Request) string {
	if id := r.Header.Get("X-Request-ID"); id != "" {
		return id
	}
	return uuid.New().String()
}
