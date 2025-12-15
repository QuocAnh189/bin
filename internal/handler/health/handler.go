package health

import (
	"encoding/json"
	"net/http"

	"github.com/aq189/bin/pkg/logger"
)

// Handler handles health check HTTP requests
type Handler struct {
	logger logger.Logger
}

// NewHandler creates a new health handler
func NewHandler(logger logger.Logger) *Handler {
	return &Handler{
		logger: logger,
	}
}

// HealthResponse represents the health check response
type HealthResponse struct {
	Status string `json:"status"`
}

// Health handles liveness probe requests
func (h *Handler) Health(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(HealthResponse{Status: "ok"})
}

// Ready handles readiness probe requests
func (h *Handler) Ready(w http.ResponseWriter, r *http.Request) {
	// TODO: Check dependencies (database, redis, etc.)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(HealthResponse{Status: "ready"})
}
