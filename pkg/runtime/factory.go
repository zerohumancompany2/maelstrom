package runtime

import (
	"fmt"

	"github.com/maelstrom/v3/pkg/registry"
)

// Factory creates ChartRuntimes from ChartDefinitions.
type Factory struct {
	reg *registry.Registry
}

// NewFactory creates a factory bound to a registry.
func NewFactory(reg *registry.Registry) *Factory {
	return &Factory{reg: reg}
}

// Create instantiates a single runtime from a definition.
func (f *Factory) Create(key string) (*ChartRuntime, error) {
	val, err := f.reg.Get(key)
	if err != nil {
		return nil, fmt.Errorf("failed to get %s from registry: %w", key, err)
	}

	return NewChartRuntime(key, val)
}

// CreateAll instantiates runtimes for all charts in the registry.
func (f *Factory) CreateAll() (map[string]*ChartRuntime, error) {
	result := make(map[string]*ChartRuntime)

	f.reg.CloneUnderLock(func(snapshot map[string]interface{}) {
		for key, def := range snapshot {
			rt, err := NewChartRuntime(key, def)
			if err != nil {
				continue // Skip invalid definitions
			}
			result[key] = rt
		}
	})

	return result, nil
}

// Refresh updates runtimes based on registry changes.
func (f *Factory) Refresh(runtimes map[string]*ChartRuntime) error {
	// Get current registry keys
	var registryKeys []string
	f.reg.CloneUnderLock(func(snapshot map[string]interface{}) {
		for key := range snapshot {
			registryKeys = append(registryKeys, key)
		}
	})

	// Find orphaned charts (in registry but not in runtimes)
	for _, key := range registryKeys {
		if _, exists := runtimes[key]; !exists {
			rt, err := f.Create(key)
			if err != nil {
				continue
			}
			runtimes[key] = rt
		}
	}

	// Find stale runtimes (in runtimes but not in registry)
	for key := range runtimes {
		found := false
		for _, regKey := range registryKeys {
			if regKey == key {
				found = true
				break
			}
		}
		if !found {
			runtimes[key].Stop()
			delete(runtimes, key)
		}
	}

	return nil
}
