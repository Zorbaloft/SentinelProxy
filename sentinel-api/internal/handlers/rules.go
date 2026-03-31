package handlers

import (
	"encoding/json"
	"io"
	"net/http"

	"go.mongodb.org/mongo-driver/bson"

	"github.com/sentinel-proxy/api/internal/storage"
)

type RulesHandler struct {
	storage *storage.Storage
}

func NewRulesHandler(s *storage.Storage) *RulesHandler {
	return &RulesHandler{storage: s}
}

func (h *RulesHandler) GetRules(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	rules, err := h.storage.GetRules(ctx)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"rules": rules})
}

func (h *RulesHandler) CreateRule(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var rule bson.M
	if err := json.Unmarshal(body, &rule); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	id, err := h.storage.CreateRule(ctx, rule)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"id": id.Hex()})
}

func (h *RulesHandler) UpdateRule(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id := r.URL.Path[len("/rules/"):]
	if id == "" {
		http.Error(w, "rule ID required", http.StatusBadRequest)
		return
	}

	// Remove trailing slash if present
	if len(id) > 0 && id[len(id)-1] == '/' {
		id = id[:len(id)-1]
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var rule bson.M
	if err := json.Unmarshal(body, &rule); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = h.storage.UpdateRule(ctx, id, rule)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "updated"})
}

func (h *RulesHandler) DeleteRule(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id := r.URL.Path[len("/rules/"):]
	if id == "" {
		http.Error(w, "rule ID required", http.StatusBadRequest)
		return
	}

	// Remove trailing slash if present
	if len(id) > 0 && id[len(id)-1] == '/' {
		id = id[:len(id)-1]
	}

	err := h.storage.DeleteRule(ctx, id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "deleted"})
}
