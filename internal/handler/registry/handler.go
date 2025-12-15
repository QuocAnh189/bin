package registry

import (
	"encoding/json"
	"net/http"
	"strings"

	"root/internal/domain/service"
	"root/internal/service/registry"
	"root/pkg/logger"
)

// Handler handles service registry HTTP requests
type Handler struct {
	service *registry.Service
	logger  logger.Logger
}

// NewHandler creates a new registry handler
func NewHandler(service *registry.Service, logger logger.Logger) *Handler {
	return &Handler{
		service: service,
		logger:  logger,
	}
}

// RegisterRequest represents the request to register a service
type RegisterRequest struct {
	ID             string            `json:"id"`
	Name           string            `json:"name"`
	Version        string            `json:"version"`
	Endpoints      []string          `json:"endpoints"`
	Capabilities   []string          `json:"capabilities"`
	Metadata       map[string]string `json:"metadata,omitempty"`
	HealthCheckURL string            `json:"health_check_url,omitempty"`
}

// Register handles service registration requests
func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	svc := &service.Service{
		ID:             req.ID,
		Name:           req.Name,
		Version:        req.Version,
		Endpoints:      req.Endpoints,
		Capabilities:   req.Capabilities,
		Metadata:       req.Metadata,
		HealthCheckURL: req.HealthCheckURL,
	}

	if err := h.service.Register(r.Context(), svc); err != nil {
		h.logger.Error("failed to register service", map[string]any{"error": err})
		http.Error(w, "failed to register service", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(svc)
}

// Deregister handles service deregistration requests
func (h *Handler) Deregister(w http.ResponseWriter, r *http.Request) {
	id := extractID(r.URL.Path)
	if id == "" {
		http.Error(w, "service id required", http.StatusBadRequest)
		return
	}

	if err := h.service.Deregister(r.Context(), id); err != nil {
		h.logger.Error("failed to deregister service", map[string]any{"error": err, "service_id": id})
		http.Error(w, "failed to deregister service", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// ListServices handles listing all registered services
func (h *Handler) ListServices(w http.ResponseWriter, r *http.Request) {
	services, err := h.service.List(r.Context())
	if err != nil {
		h.logger.Error("failed to list services", map[string]any{"error": err})
		http.Error(w, "failed to list services", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(services)
}

// Discover handles service discovery requests
func (h *Handler) Discover(w http.ResponseWriter, r *http.Request) {
	capability := r.URL.Query().Get("capability")

	services, err := h.service.Discover(r.Context(), capability)
	if err != nil {
		h.logger.Error("failed to discover services", map[string]any{"error": err})
		http.Error(w, "failed to discover services", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(services)
}

// Heartbeat handles service heartbeat requests
func (h *Handler) Heartbeat(w http.ResponseWriter, r *http.Request) {
	id := extractID(r.URL.Path)
	if id == "" {
		http.Error(w, "service id required", http.StatusBadRequest)
		return
	}

	if err := h.service.Heartbeat(r.Context(), id); err != nil {
		http.Error(w, "failed to update heartbeat", http.StatusInternalServerError)
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
