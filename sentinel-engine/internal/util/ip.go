package util

import (
	"net/http"
	"strings"
)

// ExtractClientIP determines the client IP using the trusted header order:
// 1. CF-Connecting-IP
// 2. first IP in X-Forwarded-For
// 3. X-Real-IP
// 4. remote addr
func ExtractClientIP(r *http.Request) string {
	// 1. Cloudflare
	if ip := r.Header.Get("CF-Connecting-IP"); ip != "" {
		return strings.TrimSpace(ip)
	}

	// 2. X-Forwarded-For (first IP)
	if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
		ips := strings.Split(forwarded, ",")
		if len(ips) > 0 {
			return strings.TrimSpace(ips[0])
		}
	}

	// 3. X-Real-IP
	if ip := r.Header.Get("X-Real-IP"); ip != "" {
		return strings.TrimSpace(ip)
	}

	// 4. Remote address
	ip := r.RemoteAddr
	if idx := strings.LastIndex(ip, ":"); idx != -1 {
		ip = ip[:idx]
	}
	return ip
}
