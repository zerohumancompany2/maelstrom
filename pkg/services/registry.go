// Package services provides service registry functionality.
// Spec Reference: Section 7.3
package services

import (
	"errors"
	"sync"
)

// ErrAlreadyRegistered is returned when a service is registered twice.
var ErrAlreadyRegistered = errors.New("service already registered")

// ErrNotFound is returned when a service is not found.
var ErrNotFound = errors.New("service not found")

// Service represents a service in the system.
type Service interface {
	Name() string
	HandleMail(mail any) (any, error)
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
	// TODO: implement
	return nil
}

// Get retrieves a service by name.
// Returns service and true if found, nil and false otherwise.
func (sr *ServiceRegistry) Get(name string) (Service, bool) {
	// TODO: implement
	return nil, false
}

// List returns all registered service names.
func (sr *ServiceRegistry) List() []string {
	// TODO: implement
	return nil
}

// TODO: implement lifecycle tracking (registered, running, stopped)
// TODO: implement thread-safe operations
// TODO: implement well-known ID addressing for sys:* services
