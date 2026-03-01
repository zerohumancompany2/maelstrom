package runtime

import (
	"errors"
	"fmt"

	"github.com/maelstrom/v3/pkg/registry"
	"github.com/maelstrom/v3/pkg/statechart"
)

var ErrChartNotFound = errors.New("chart not found")

// Factory creates ChartRuntimes from ChartDefinitions.
type Factory struct {
	reg           *registry.Registry
	engine        statechart.Library
	statechartIDs map[string]statechart.RuntimeID
}

// NewFactory creates a factory bound to a registry.
func NewFactory(reg *registry.Registry) *Factory {
	return &Factory{
		reg:           reg,
		statechartIDs: make(map[string]statechart.RuntimeID),
	}
}

// WithEngine attaches a statechart engine (fluent config).
func (f *Factory) WithEngine(engine statechart.Library) *Factory {
	f.engine = engine
	return f
}

// Create instantiates a single runtime from a definition.
func (f *Factory) Create(key string) (*ChartRuntime, error) {
	val, err := f.reg.Get(key)
	if err != nil {
		return nil, fmt.Errorf("failed to get %s from registry: %w", key, err)
	}

	def, ok := val.(statechart.ChartDefinition)
	if !ok {
		return nil, fmt.Errorf("definition for %s is not a ChartDefinition", key)
	}

	if f.engine != nil {
		id, err := f.engine.Spawn(def, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to spawn in engine: %w", err)
		}
		f.statechartIDs[key] = id
	}

	rt, err := NewChartRuntime(key, val)
	if err != nil {
		return nil, err
	}

	return rt, nil
}

// CreateAll instantiates runtimes for all charts in the registry.
func (f *Factory) CreateAll() (map[string]*ChartRuntime, error) {
	result := make(map[string]*ChartRuntime)

	f.reg.CloneUnderLock(func(snapshot map[string]interface{}) {
		for key, def := range snapshot {
			def, ok := def.(statechart.ChartDefinition)
			if !ok {
				continue // Skip non-chart definitions
			}

			if f.engine != nil {
				id, err := f.engine.Spawn(def, nil)
				if err != nil {
					continue // Skip failed spawns
				}
				f.statechartIDs[key] = id
			}

			rt, err := NewChartRuntime(key, def)
			if err != nil {
				continue // Skip invalid definitions
			}
			result[key] = rt
		}
	})

	return result, nil
}

// GetStatechartID returns the RuntimeID for a chart.
func (f *Factory) GetStatechartID(chartID string) (statechart.RuntimeID, bool) {
	id, ok := f.statechartIDs[chartID]
	return id, ok
}

// DispatchEvent sends an event to a runtime via the engine.
func (f *Factory) DispatchEvent(chartID string, evt Event) error {
	if f.engine == nil {
		return fmt.Errorf("engine not configured")
	}

	runtimeID, ok := f.statechartIDs[chartID]
	if !ok {
		return fmt.Errorf("chart not found: %s", chartID)
	}

	statechartEvt := statechart.Event{
		Type:          evt.Type,
		Payload:       evt.Payload,
		CorrelationID: evt.CorrelationID,
		Source:        evt.Source,
	}

	return f.engine.Dispatch(runtimeID, statechartEvt)
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
