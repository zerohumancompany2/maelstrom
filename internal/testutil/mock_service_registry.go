package testutil

import (
	"sync"
)

// MockServiceRegistry provides a mock implementation of services.Registry for testing.
type MockServiceRegistry struct {
	mu         sync.RWMutex
	services   map[string]Service
	registered []string
}

// Service is a simplified service interface for testing.
type Service interface {
	Name() string
}

// MockService is a mock service implementation.
type MockService struct {
	name string
}

// NewMockService creates a new mock service.
func NewMockService(name string) *MockService {
	return &MockService{name: name}
}

// Name returns the service name.
func (s *MockService) Name() string {
	return s.name
}

// NewMockServiceRegistry creates a new mock service registry.
func NewMockServiceRegistry() *MockServiceRegistry {
	return &MockServiceRegistry{
		services:   make(map[string]Service),
		registered: make([]string, 0),
	}
}

// Register registers a service.
func (sr *MockServiceRegistry) Register(name string, svc Service) error {
	sr.mu.Lock()
	defer sr.mu.Unlock()

	if _, exists := sr.services[name]; exists {
		return ErrAlreadyRegistered
	}

	sr.services[name] = svc
	sr.registered = append(sr.registered, name)
	return nil
}

// Get retrieves a service by name.
func (sr *MockServiceRegistry) Get(name string) (Service, bool) {
	sr.mu.RLock()
	defer sr.mu.RUnlock()

	svc, ok := sr.services[name]
	return svc, ok
}

// List returns all registered service names.
func (sr *MockServiceRegistry) List() []string {
	sr.mu.RLock()
	defer sr.mu.RUnlock()

	result := make([]string, len(sr.registered))
	copy(result, sr.registered)
	return result
}

// Count returns the number of registered services.
func (sr *MockServiceRegistry) Count() int {
	sr.mu.RLock()
	defer sr.mu.RUnlock()
	return len(sr.registered)
}

// ErrAlreadyRegistered is returned when a service is registered twice.
var ErrAlreadyRegistered = &Error{"service already registered"}
