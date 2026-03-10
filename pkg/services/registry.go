// Package services provides service registry functionality.
// Spec Reference: Section 7.3
package services

import (
	"errors"
	"sort"
	"sync"

	"github.com/maelstrom/v3/pkg/mail"
)

// ErrAlreadyRegistered is returned when a service is registered twice.
var ErrAlreadyRegistered = errors.New("service already registered")

// ErrNotFound is returned when a service is not found.
var ErrNotFound = errors.New("service not found")

// Service represents a service in the system.
type Service interface {
	ID() string
	HandleMail(mail mail.Mail) *OutcomeEvent
	Start() error
	Stop() error
}

// ServiceRegistry manages service registration and lookup.
type ServiceRegistry struct {
	services     map[string]Service
	mu           sync.RWMutex
	lifecycles   map[string]string   // service -> lifecycle state
	capabilities map[string][]string // service -> capabilities
	capIndex     map[string][]string // capability -> service IDs
	health       map[string]string   // service -> health status
	dependencies map[string][]string // service -> dependencies
	dependents   map[string][]string // service -> dependents (reverse index)
}

// NewServiceRegistry creates a new ServiceRegistry.
func NewServiceRegistry() *ServiceRegistry {
	return &ServiceRegistry{
		services:     make(map[string]Service),
		lifecycles:   make(map[string]string),
		capabilities: make(map[string][]string),
		capIndex:     make(map[string][]string),
		health:       make(map[string]string),
		dependencies: make(map[string][]string),
		dependents:   make(map[string][]string),
	}
}

// Register registers a service with the registry.
// Returns error if service name already exists.
func (sr *ServiceRegistry) Register(name string, svc Service) error {
	sr.mu.Lock()
	defer sr.mu.Unlock()
	if _, exists := sr.services[name]; exists {
		return ErrAlreadyRegistered
	}
	sr.services[name] = svc
	return nil
}

// RegisterWithState registers a service with an initial lifecycle state.
// Returns error if service is already registered.
func (sr *ServiceRegistry) RegisterWithState(svc Service, initialState string) error {
	sr.mu.Lock()
	defer sr.mu.Unlock()
	id := svc.ID()
	if _, exists := sr.services[id]; exists {
		return ErrAlreadyRegistered
	}
	sr.services[id] = svc
	sr.lifecycles[id] = initialState
	return nil
}

// RegisterWithCapabilities registers a service with the given capabilities.
// Returns error if service is already registered.
func (sr *ServiceRegistry) RegisterWithCapabilities(name string, svc Service, caps []string) error {
	sr.mu.Lock()
	defer sr.mu.Unlock()
	if _, exists := sr.services[name]; exists {
		return ErrAlreadyRegistered
	}
	sr.services[name] = svc
	sr.capabilities[name] = caps
	for _, cap := range caps {
		sr.capIndex[cap] = append(sr.capIndex[cap], name)
	}
	return nil
}

// FindByCapability returns all services that have the given capability.
func (sr *ServiceRegistry) FindByCapability(capability string) []Service {
	sr.mu.RLock()
	defer sr.mu.RUnlock()
	serviceIDs := sr.capIndex[capability]
	services := make([]Service, 0, len(serviceIDs))
	for _, id := range serviceIDs {
		if svc, ok := sr.services[id]; ok {
			services = append(services, svc)
		}
	}
	return services
}

// GetHealthStatus returns the health status of a service.
func (sr *ServiceRegistry) GetHealthStatus(serviceID string) string {
	sr.mu.RLock()
	defer sr.mu.RUnlock()
	if status, ok := sr.health[serviceID]; ok {
		return status
	}
	return "unknown"
}

// UpdateHealthStatus updates the health status of a service.
func (sr *ServiceRegistry) UpdateHealthStatus(serviceID string, status string) {
	sr.mu.Lock()
	defer sr.mu.Unlock()
	sr.health[serviceID] = status
}

// GetUnhealthyServices returns all services with unhealthy status.
func (sr *ServiceRegistry) GetUnhealthyServices() []Service {
	sr.mu.RLock()
	defer sr.mu.RUnlock()
	unhealthy := make([]Service, 0)
	for id, status := range sr.health {
		if status == "unhealthy" {
			if svc, ok := sr.services[id]; ok {
				unhealthy = append(unhealthy, svc)
			}
		}
	}
	return unhealthy
}

// GetState retrieves the lifecycle state of a service.
// Returns state and true if found, empty string and false otherwise.
func (sr *ServiceRegistry) GetState(serviceID string) (string, bool) {
	sr.mu.RLock()
	defer sr.mu.RUnlock()
	state, ok := sr.lifecycles[serviceID]
	return state, ok
}

// UpdateState updates the lifecycle state of a service.
// Returns ErrNotFound if service is not registered.
func (sr *ServiceRegistry) UpdateState(serviceID string, newState string) error {
	sr.mu.Lock()
	defer sr.mu.Unlock()
	if _, exists := sr.lifecycles[serviceID]; !exists {
		return ErrNotFound
	}
	sr.lifecycles[serviceID] = newState
	return nil
}

// QueryByState returns all services matching the given lifecycle state.
func (sr *ServiceRegistry) QueryByState(state string) []Service {
	sr.mu.RLock()
	defer sr.mu.RUnlock()
	result := make([]Service, 0)
	for id, s := range sr.services {
		if sr.lifecycles[id] == state {
			result = append(result, s)
		}
	}
	return result
}

// Get retrieves a service by name.
// Returns service and true if found, nil and false otherwise.
func (sr *ServiceRegistry) Get(name string) (Service, bool) {
	sr.mu.RLock()
	defer sr.mu.RUnlock()
	svc, ok := sr.services[name]
	return svc, ok
}

// List returns all registered service names.
func (sr *ServiceRegistry) List() []string {
	sr.mu.RLock()
	defer sr.mu.RUnlock()
	names := make([]string, 0, len(sr.services))
	for name := range sr.services {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// DiscoverServices returns all registered services.
func (sr *ServiceRegistry) DiscoverServices() []Service {
	sr.mu.RLock()
	defer sr.mu.RUnlock()
	services := make([]Service, 0, len(sr.services))
	for _, svc := range sr.services {
		services = append(services, svc)
	}
	return services
}

// RegisterWithDependencies registers a service with the given dependencies.
// Returns error if service is already registered.
func (sr *ServiceRegistry) RegisterWithDependencies(name string, svc Service, deps []string) error {
	sr.mu.Lock()
	defer sr.mu.Unlock()
	if _, exists := sr.services[name]; exists {
		return ErrAlreadyRegistered
	}
	sr.services[name] = svc
	sr.dependencies[name] = deps
	for _, dep := range deps {
		sr.dependents[dep] = append(sr.dependents[dep], name)
	}
	return nil
}

// GetDependencies returns the dependencies of a service.
func (sr *ServiceRegistry) GetDependencies(serviceID string) []string {
	sr.mu.RLock()
	defer sr.mu.RUnlock()
	deps, ok := sr.dependencies[serviceID]
	if !ok {
		return []string{}
	}
	result := make([]string, len(deps))
	copy(result, deps)
	return result
}

// GetDependents returns all services that depend on the given service.
func (sr *ServiceRegistry) GetDependents(serviceID string) []string {
	sr.mu.RLock()
	defer sr.mu.RUnlock()
	deps, ok := sr.dependents[serviceID]
	if !ok {
		return []string{}
	}
	result := make([]string, len(deps))
	copy(result, deps)
	return result
}

// TODO: implement lifecycle tracking (registered, running, stopped)
// TODO: implement thread-safe operations
// TODO: implement well-known ID addressing for sys:* services
