package auth

import (
	"encoding/json"
	"net/http"

	"root/internal/domain/token"
	"root/internal/service/auth"
	"root/pkg/logger"
)

// Handler handles authentication HTTP requests
type Handler struct {
	service *auth.Service
	logger  logger.Logger
}

// NewHandler creates a new auth handler
func NewHandler(service *auth.Service, logger logger.Logger) *Handler {
	return &Handler{
		service: service,
		logger:  logger,
	}
}

// IssueTokenRequest represents the request to issue a token
type IssueTokenRequest struct {
	Subject  string         `json:"subject"`
	Roles    []string       `json:"roles,omitempty"`
	Audience string         `json:"audience,omitempty"`
	Metadata map[string]any `json:"metadata,omitempty"`
}

// IssueToken handles token issuance requests
func (h *Handler) IssueToken(w http.ResponseWriter, r *http.Request) {
	var req IssueTokenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	claims := token.Claims{
		Subject:  req.Subject,
		Roles:    req.Roles,
		Audience: req.Audience,
		Metadata: req.Metadata,
	}

	tok, err := h.service.IssueToken(r.Context(), claims)
	if err != nil {
		h.logger.Error("failed to issue token", map[string]any{"error": err})
		http.Error(w, "failed to issue token", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tok)
}

// ValidateTokenRequest represents the request to validate a token
type ValidateTokenRequest struct {
	Token string `json:"token"`
}

// ValidateToken handles token validation requests
func (h *Handler) ValidateToken(w http.ResponseWriter, r *http.Request) {
	var req ValidateTokenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	claims, err := h.service.ValidateToken(r.Context(), req.Token)
	if err != nil {
		http.Error(w, "invalid token", http.StatusUnauthorized)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(claims)
}

// RefreshTokenRequest represents the request to refresh a token
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token"`
}

// RefreshToken handles token refresh requests
func (h *Handler) RefreshToken(w http.ResponseWriter, r *http.Request) {
	var req RefreshTokenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	tok, err := h.service.RefreshToken(r.Context(), req.RefreshToken)
	if err != nil {
		http.Error(w, "failed to refresh token", http.StatusUnauthorized)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tok)
}

// RevokeTokenRequest represents the request to revoke a token
type RevokeTokenRequest struct {
	Token string `json:"token"`
}

// RevokeToken handles token revocation requests
func (h *Handler) RevokeToken(w http.ResponseWriter, r *http.Request) {
	var req RevokeTokenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if err := h.service.RevokeToken(r.Context(), req.Token); err != nil {
		h.logger.Error("failed to revoke token", map[string]any{"error": err})
		http.Error(w, "failed to revoke token", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
