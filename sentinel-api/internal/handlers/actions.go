package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/go-redis/redis/v8"
)

type ActionsHandler struct {
	redis *redis.Client
}

func NewActionsHandler(redisAddr string) (*ActionsHandler, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr: redisAddr,
	})
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("redis connection failed: %w", err)
	}
	return &ActionsHandler{redis: rdb}, nil
}

type BlockRequest struct {
	IP      string `json:"ip"`
	TTLSec  int    `json:"ttlSec"`
	Reason  string `json:"reason"`
}

type RedirectRequest struct {
	IP       string `json:"ip"`
	TargetURL string `json:"targetUrl"`
	TTLSec   int     `json:"ttlSec"`
	Reason   string  `json:"reason"`
}

func (h *ActionsHandler) Block(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var req BlockRequest
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := json.Unmarshal(body, &req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if req.IP == "" {
		http.Error(w, "ip required", http.StatusBadRequest)
		return
	}

	ttlSec := req.TTLSec
	if ttlSec == 0 {
		ttlSec = 3600 // Default 1 hour
	}

	blockEntry := map[string]interface{}{
		"reason":    req.Reason,
		"createdAt": time.Now(),
		"expiresAt": time.Now().Add(time.Duration(ttlSec) * time.Second),
	}

	blockJSON, _ := json.Marshal(blockEntry)
	key := fmt.Sprintf("blocklist:%s", req.IP)

	if err := h.redis.Set(ctx, key, blockJSON, time.Duration(ttlSec)*time.Second).Err(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "blocked"})
}

func (h *ActionsHandler) Unblock(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var req struct {
		IP string `json:"ip"`
	}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := json.Unmarshal(body, &req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if req.IP == "" {
		http.Error(w, "ip required", http.StatusBadRequest)
		return
	}

	key := fmt.Sprintf("blocklist:%s", req.IP)
	if err := h.redis.Del(ctx, key).Err(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "unblocked"})
}

func (h *ActionsHandler) Redirect(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var req RedirectRequest
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := json.Unmarshal(body, &req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if req.IP == "" || req.TargetURL == "" {
		http.Error(w, "ip and targetUrl required", http.StatusBadRequest)
		return
	}

	ttlSec := req.TTLSec
	if ttlSec == 0 {
		ttlSec = 3600 // Default 1 hour
	}

	redirectEntry := map[string]interface{}{
		"targetUrl": req.TargetURL,
		"reason":    req.Reason,
		"createdAt": time.Now(),
		"expiresAt": time.Now().Add(time.Duration(ttlSec) * time.Second),
	}

	redirectJSON, _ := json.Marshal(redirectEntry)
	key := fmt.Sprintf("redirect:%s", req.IP)

	if err := h.redis.Set(ctx, key, redirectJSON, time.Duration(ttlSec)*time.Second).Err(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "redirected"})
}

func (h *ActionsHandler) Unredirect(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var req struct {
		IP string `json:"ip"`
	}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := json.Unmarshal(body, &req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if req.IP == "" {
		http.Error(w, "ip required", http.StatusBadRequest)
		return
	}

	key := fmt.Sprintf("redirect:%s", req.IP)
	if err := h.redis.Del(ctx, key).Err(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "unredirected"})
}
