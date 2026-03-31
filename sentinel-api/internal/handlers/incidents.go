package handlers

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/sentinel-proxy/api/internal/storage"
)

type IncidentsHandler struct {
	storage *storage.Storage
}

func NewIncidentsHandler(s *storage.Storage) *IncidentsHandler {
	return &IncidentsHandler{storage: s}
}

func (h *IncidentsHandler) GetIncidents(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	status := r.URL.Query().Get("status")

	incidents, err := h.storage.GetIncidents(ctx, status)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"incidents": incidents})
}

func (h *IncidentsHandler) CloseIncident(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	// Extract ID from path like /incidents/{id}/close
	path := r.URL.Path
	path = strings.TrimPrefix(path, "/incidents/")
	path = strings.TrimSuffix(path, "/close")
	id := strings.TrimSuffix(path, "/")
	if id == "" {
		http.Error(w, "incident ID required", http.StatusBadRequest)
		return
	}

	err := h.storage.CloseIncident(ctx, id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "closed"})
}
