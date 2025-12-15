package service

import (
	"time"
)

// Status represents the health status of a service
type Status string

const (
	StatusHealthy   Status = "healthy"
	StatusUnhealthy Status = "unhealthy"
	StatusUnknown   Status = "unknown"
)

// Service represents a registered project server
type Service struct {
	ID           string            `json:"id"`
	Name         string            `json:"name"`
	Version      string            `json:"version"`
	Endpoints    []string          `json:"endpoints"`
	Capabilities []string          `json:"capabilities"`
	Metadata     map[string]string `json:"metadata,omitempty"`
	Status       Status            `json:"status"`
	RegisteredAt time.Time         `json:"registered_at"`
	LastHeartbeat time.Time        `json:"last_heartbeat"`
	HealthCheckURL string          `json:"health_check_url,omitempty"`
}

// IsHealthy checks if the service is healthy based on heartbeat
func (s *Service) IsHealthy(timeout time.Duration) bool {
	if s.Status == StatusUnhealthy {
		return false
	}
	return time.Since(s.LastHeartbeat) < timeout
}

// UpdateHeartbeat updates the last heartbeat timestamp
func (s *Service) UpdateHeartbeat() {
	s.LastHeartbeat = time.Now()
	s.Status = StatusHealthy
}

// MarkUnhealthy marks the service as unhealthy
func (s *Service) MarkUnhealthy() {
	s.Status = StatusUnhealthy
}
