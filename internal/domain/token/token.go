package token

import (
	"time"
)

// Type represents the type of token
type Type string

const (
	TypeAccess  Type = "access"
	TypeRefresh Type = "refresh"
)

// Token represents a JWT token with metadata
type Token struct {
	Value     string            `json:"token"`
	Type      Type              `json:"type"`
	ExpiresAt time.Time         `json:"expires_at"`
	IssuedAt  time.Time         `json:"issued_at"`
	Claims    map[string]any    `json:"claims,omitempty"`
}

// Claims represents standard JWT claims
type Claims struct {
	Subject   string         `json:"sub"`           // User ID
	Issuer    string         `json:"iss"`           // Issuer
	Audience  string         `json:"aud,omitempty"` // Intended audience
	ExpiresAt time.Time      `json:"exp"`           // Expiration time
	IssuedAt  time.Time      `json:"iat"`           // Issued at
	NotBefore time.Time      `json:"nbf,omitempty"` // Not before
	TokenID   string         `json:"jti,omitempty"` // Token ID
	Roles     []string       `json:"roles,omitempty"`
	Metadata  map[string]any `json:"metadata,omitempty"`
}

// IsExpired checks if the token is expired
func (t *Token) IsExpired() bool {
	return time.Now().After(t.ExpiresAt)
}

// IsValid checks basic token validity
func (t *Token) IsValid() bool {
	return !t.IsExpired() && t.Value != ""
}
