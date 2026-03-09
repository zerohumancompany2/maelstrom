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
	HandleMail(mail mail.Mail) error
	Start() error
	Stop() error
}

// ServiceRegistry manages service registration and lookup.
type ServiceRegistry struct {
	services   map[string]Service
	mu         sync.RWMutex
	lifecycles map[string]string // service -> lifecycle state
}

// NewServiceRegistry creates a new ServiceRegistry.
func NewServiceRegistry() *ServiceRegistry {
	return &ServiceRegistry{
		services:   make(map[string]Service),
		lifecycles: make(map[string]string),
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

// TODO: implement lifecycle tracking (registered, running, stopped)
// TODO: implement thread-safe operations
// TODO: implement well-known ID addressing for sys:* services
