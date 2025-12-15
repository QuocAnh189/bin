package registry

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"root/internal/domain/service"
	"root/pkg/logger"
)

// RegistryRepository defines the interface for service registry storage
type RegistryRepository interface {
	Register(ctx context.Context, svc *service.Service) error
	Deregister(ctx context.Context, id string) error
	Get(ctx context.Context, id string) (*service.Service, error)
	List(ctx context.Context) ([]*service.Service, error)
	Update(ctx context.Context, svc *service.Service) error
}

// Service handles service registry operations
type Service struct {
	repo                RegistryRepository
	healthCheckInterval time.Duration
	healthCheckTimeout  time.Duration
	logger              logger.Logger
	httpClient          *http.Client
}

// Config holds registry service configuration
type Config struct {
	Repository          RegistryRepository
	HealthCheckInterval time.Duration
	HealthCheckTimeout  time.Duration
	Logger              logger.Logger
}

// New creates a new registry service
func New(cfg Config) *Service {
	return &Service{
		repo:                cfg.Repository,
		healthCheckInterval: cfg.HealthCheckInterval,
		healthCheckTimeout:  cfg.HealthCheckTimeout,
		logger:              cfg.Logger,
		httpClient: &http.Client{
			Timeout: cfg.HealthCheckTimeout,
		},
	}
}

// Register registers a new service
func (s *Service) Register(ctx context.Context, svc *service.Service) error {
	now := time.Now()
	svc.RegisteredAt = now
	svc.LastHeartbeat = now
	svc.Status = service.StatusHealthy

	if err := s.repo.Register(ctx, svc); err != nil {
		s.logger.Error("failed to register service", map[string]any{"error": err, "service": svc.Name})
		return fmt.Errorf("register service: %w", err)
	}

	s.logger.Info("service registered", map[string]any{
		"service_id": svc.ID,
		"name":       svc.Name,
		"version":    svc.Version,
	})

	return nil
}

// Deregister removes a service from the registry
func (s *Service) Deregister(ctx context.Context, id string) error {
	if err := s.repo.Deregister(ctx, id); err != nil {
		s.logger.Error("failed to deregister service", map[string]any{"error": err, "service_id": id})
		return fmt.Errorf("deregister service: %w", err)
	}

	s.logger.Info("service deregistered", map[string]any{"service_id": id})
	return nil
}

// List returns all registered services
func (s *Service) List(ctx context.Context) ([]*service.Service, error) {
	services, err := s.repo.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("list services: %w", err)
	}
	return services, nil
}

// Discover finds services matching the given criteria
func (s *Service) Discover(ctx context.Context, capability string) ([]*service.Service, error) {
	services, err := s.List(ctx)
	if err != nil {
		return nil, err
	}

	var matched []*service.Service
	for _, svc := range services {
		if svc.Status == service.StatusHealthy && hasCapability(svc, capability) {
			matched = append(matched, svc)
		}
	}

	return matched, nil
}

// Heartbeat updates the last heartbeat timestamp for a service
func (s *Service) Heartbeat(ctx context.Context, id string) error {
	svc, err := s.repo.Get(ctx, id)
	if err != nil {
		return fmt.Errorf("get service: %w", err)
	}

	svc.UpdateHeartbeat()

	if err := s.repo.Update(ctx, svc); err != nil {
		s.logger.Error("failed to update heartbeat", map[string]any{"error": err, "service_id": id})
		return fmt.Errorf("update heartbeat: %w", err)
	}

	return nil
}

// StartHealthChecks starts background health checks for all registered services
func (s *Service) StartHealthChecks(ctx context.Context) {
	ticker := time.NewTicker(s.healthCheckInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			s.logger.Info("stopping health checks", make(map[string]any))
			return
		case <-ticker.C:
			s.performHealthChecks(ctx)
		}
	}
}

func (s *Service) performHealthChecks(ctx context.Context) {
	services, err := s.repo.List(ctx)
	if err != nil {
		s.logger.Error("failed to list services for health check", map[string]any{"error": err})
		return
	}

	for _, svc := range services {
		if svc.HealthCheckURL == "" {
			// Use heartbeat-based health check
			if !svc.IsHealthy(s.healthCheckInterval * 2) {
				svc.MarkUnhealthy()
				s.repo.Update(ctx, svc)
				s.logger.Warn("service marked unhealthy (heartbeat timeout)", map[string]any{
					"service_id": svc.ID,
					"name":       svc.Name,
				})
			}
			continue
		}

		// Perform HTTP health check
		go s.checkServiceHealth(ctx, svc)
	}
}

func (s *Service) checkServiceHealth(ctx context.Context, svc *service.Service) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, svc.HealthCheckURL, nil)
	if err != nil {
		s.logger.Error("failed to create health check request", map[string]any{"error": err, "service_id": svc.ID})
		return
	}

	resp, err := s.httpClient.Do(req)
	if err != nil || resp.StatusCode != http.StatusOK {
		svc.MarkUnhealthy()
		s.repo.Update(ctx, svc)
		s.logger.Warn("service health check failed", map[string]any{
			"service_id": svc.ID,
			"name":       svc.Name,
			"error":      err,
		})
		return
	}
	defer resp.Body.Close()

	if svc.Status != service.StatusHealthy {
		svc.Status = service.StatusHealthy
		s.repo.Update(ctx, svc)
		s.logger.Info("service recovered", map[string]any{
			"service_id": svc.ID,
			"name":       svc.Name,
		})
	}
}

func hasCapability(svc *service.Service, capability string) bool {
	if capability == "" {
		return true
	}
	for _, cap := range svc.Capabilities {
		if cap == capability {
			return true
		}
	}
	return false
}
