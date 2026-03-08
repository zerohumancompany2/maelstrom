package datasource

import (
	"github.com/maelstrom/v3/pkg/security"
	"sync"
)

type DataSource interface {
	TagOnWrite(path string, taints []string) error
	GetTaints(path string) ([]string, error)
	ValidateAccess(boundary security.BoundaryType) error
}

type Registry struct {
	mu      sync.RWMutex
	sources map[string]func(map[string]any) (DataSource, error)
}

func NewRegistry() *Registry {
	return &Registry{
		sources: make(map[string]func(map[string]any) (DataSource, error)),
	}
}

func (r *Registry) Register(name string, factory func(map[string]any) (DataSource, error)) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.sources[name] = factory
}

func (r *Registry) Get(name string, config map[string]any) (DataSource, error) {
	r.mu.RLock()
	factory, ok := r.sources[name]
	r.mu.RUnlock()

	if !ok {
		return nil, nil
	}

	return factory(config)
}

func (r *Registry) List() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	names := make([]string, 0, len(r.sources))
	for name := range r.sources {
		names = append(names, name)
	}
	return names
}

var globalRegistry *Registry

func init() {
	globalRegistry = NewRegistry()
}

func Register(name string, factory func(map[string]any) (DataSource, error)) {
	globalRegistry.Register(name, factory)
}

func Get(name string, config map[string]any) (DataSource, error) {
	return globalRegistry.Get(name, config)
}

func List() []string {
	return globalRegistry.List()
}
