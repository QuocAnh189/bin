package memory

import (
	"fmt"
	"sync"
)

// ConfigRepository implements in-memory configuration storage
type ConfigRepository struct {
	mu      sync.RWMutex
	configs map[string]map[string]map[string]any // serviceID -> version -> config
}

// NewConfigRepository creates a new in-memory config repository
func NewConfigRepository() *ConfigRepository {
	return &ConfigRepository{
		configs: make(map[string]map[string]map[string]any),
	}
}

// Get retrieves configuration for a service and version
func (r *ConfigRepository) Get(serviceID, version string) (map[string]any, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	versions, exists := r.configs[serviceID]
	if !exists {
		return nil, fmt.Errorf("service not found")
	}

	config, exists := versions[version]
	if !exists {
		return nil, fmt.Errorf("version not found")
	}

	return config, nil
}

// Set stores configuration for a service and version
func (r *ConfigRepository) Set(serviceID, version string, config map[string]any) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.configs[serviceID]; !exists {
		r.configs[serviceID] = make(map[string]map[string]any)
	}

	r.configs[serviceID][version] = config
	return nil
}

// Delete removes configuration for a service and version
func (r *ConfigRepository) Delete(serviceID, version string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if versions, exists := r.configs[serviceID]; exists {
		delete(versions, version)
		if len(versions) == 0 {
			delete(r.configs, serviceID)
		}
	}

	return nil
}

// List returns all versions for a service
func (r *ConfigRepository) List(serviceID string) ([]string, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	versions, exists := r.configs[serviceID]
	if !exists {
		return []string{}, nil
	}

	result := make([]string, 0, len(versions))
	for version := range versions {
		result = append(result, version)
	}

	return result, nil
}
