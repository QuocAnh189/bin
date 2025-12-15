package session

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/aq189/bin/internal/service/session"
	"github.com/aq189/bin/pkg/logger"
)

// Handler handles session HTTP requests
type Handler struct {
	service *session.Service
	logger  logger.Logger
}

// NewHandler creates a new session handler
func NewHandler(service *session.Service, logger logger.Logger) *Handler {
	return &Handler{
		service: service,
		logger:  logger,
	}
}

// CreateSessionRequest represents the request to create a session
type CreateSessionRequest struct {
	UserID    string         `json:"user_id"`
	ServiceID string         `json:"service_id"`
	Data      map[string]any `json:"data"`
	TTL       int            `json:"ttl"` // minutes
}

// Create handles session creation requests
func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	var req CreateSessionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	ttl := time.Duration(req.TTL) * time.Minute
	sess, err := h.service.Create(r.Context(), req.UserID, req.ServiceID, req.Data, ttl)
	if err != nil {
		h.logger.Error("failed to create session", map[string]any{"error": err})
		http.Error(w, "failed to create session", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(sess)
}

// Get handles session retrieval requests
func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	id := extractID(r.URL.Path)
	if id == "" {
		http.Error(w, "session id required", http.StatusBadRequest)
		return
	}

	sess, err := h.service.Get(r.Context(), id)
	if err != nil {
		http.Error(w, "session not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(sess)
}

// UpdateSessionRequest represents the request to update a session
type UpdateSessionRequest struct {
	Data map[string]any `json:"data"`
}

// Update handles session update requests
func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	id := extractID(r.URL.Path)
	if id == "" {
		http.Error(w, "session id required", http.StatusBadRequest)
		return
	}

	var req UpdateSessionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if err := h.service.Update(r.Context(), id, req.Data); err != nil {
		h.logger.Error("failed to update session", map[string]any{"error": err, "session_id": id})
		http.Error(w, "failed to update session", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// Delete handles session deletion requests
func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	id := extractID(r.URL.Path)
	if id == "" {
		http.Error(w, "session id required", http.StatusBadRequest)
		return
	}

	if err := h.service.Delete(r.Context(), id); err != nil {
		h.logger.Error("failed to delete session", map[string]any{"error": err, "session_id": id})
		http.Error(w, "failed to delete session", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func extractID(path string) string {
	parts := strings.Split(path, "/")
	if len(parts) > 0 {
		return parts[len(parts)-1]
	}
	return ""
}
