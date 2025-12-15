package memory

import (
	"context"
	"fmt"
	"sync"

	"github.com/aq189/bin/internal/domain/service"
)

// RegistryRepository implements in-memory service registry storage
type RegistryRepository struct {
	mu       sync.RWMutex
	services map[string]*service.Service
}

// NewRegistryRepository creates a new in-memory registry repository
func NewRegistryRepository() *RegistryRepository {
	return &RegistryRepository{
		services: make(map[string]*service.Service),
	}
}

// Register stores a new service
func (r *RegistryRepository) Register(ctx context.Context, svc *service.Service) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.services[svc.ID] = svc
	return nil
}

// Deregister removes a service
func (r *RegistryRepository) Deregister(ctx context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	delete(r.services, id)
	return nil
}

// Get retrieves a service by ID
func (r *RegistryRepository) Get(ctx context.Context, id string) (*service.Service, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	svc, exists := r.services[id]
	if !exists {
		return nil, fmt.Errorf("service not found")
	}

	return svc, nil
}

// List returns all registered services
func (r *RegistryRepository) List(ctx context.Context) ([]*service.Service, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	services := make([]*service.Service, 0, len(r.services))
	for _, svc := range r.services {
		services = append(services, svc)
	}

	return services, nil
}

// Update updates an existing service
func (r *RegistryRepository) Update(ctx context.Context, svc *service.Service) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.services[svc.ID]; !exists {
		return fmt.Errorf("service not found")
	}

	r.services[svc.ID] = svc
	return nil
}
