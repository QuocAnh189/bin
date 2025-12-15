package session

import (
	"time"
)

// Session represents a user session
type Session struct {
	ID        string         `json:"id"`
	UserID    string         `json:"user_id"`
	ServiceID string         `json:"service_id"`
	Data      map[string]any `json:"data"`
	CreatedAt time.Time      `json:"created_at"`
	ExpiresAt time.Time      `json:"expires_at"`
	UpdatedAt time.Time      `json:"updated_at"`
}

// IsExpired checks if the session has expired
func (s *Session) IsExpired() bool {
	return time.Now().After(s.ExpiresAt)
}

// IsActive checks if the session is currently active
func (s *Session) IsActive() bool {
	return !s.IsExpired()
}

// Touch updates the session's UpdatedAt timestamp
func (s *Session) Touch() {
	s.UpdatedAt = time.Now()
}
