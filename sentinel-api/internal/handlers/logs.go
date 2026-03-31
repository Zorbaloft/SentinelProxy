package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"go.mongodb.org/mongo-driver/bson"

	"github.com/sentinel-proxy/api/internal/storage"
)

type LogsHandler struct {
	storage *storage.Storage
}

func NewLogsHandler(s *storage.Storage) *LogsHandler {
	return &LogsHandler{storage: s}
}

func (h *LogsHandler) GetLogs(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Parse query parameters
	filter := bson.M{}

	if ip := r.URL.Query().Get("ip"); ip != "" {
		filter["client.ip"] = ip
	}

	if path := r.URL.Query().Get("path"); path != "" {
		filter["request.path"] = bson.M{"$regex": path, "$options": "i"}
	}

	if statusStr := r.URL.Query().Get("status"); statusStr != "" {
		if status, err := strconv.Atoi(statusStr); err == nil {
			filter["response.status"] = status
		}
	}

	if fromStr := r.URL.Query().Get("from"); fromStr != "" {
		if from, err := time.Parse(time.RFC3339, fromStr); err == nil {
			if filter["timestamp"] == nil {
				filter["timestamp"] = bson.M{}
			}
			filter["timestamp"].(bson.M)["$gte"] = from
		}
	}

	if toStr := r.URL.Query().Get("to"); toStr != "" {
		if to, err := time.Parse(time.RFC3339, toStr); err == nil {
			if filter["timestamp"] == nil {
				filter["timestamp"] = bson.M{}
			}
			filter["timestamp"].(bson.M)["$lte"] = to
		}
	}

	limit := int64(100)
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if l, err := strconv.ParseInt(limitStr, 10, 64); err == nil && l > 0 && l <= 1000 {
			limit = l
		}
	}

	cursor := r.URL.Query().Get("cursor")

	logs, nextCursor, err := h.storage.GetLogs(ctx, filter, limit, cursor)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"logs":       logs,
		"nextCursor": nextCursor,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
