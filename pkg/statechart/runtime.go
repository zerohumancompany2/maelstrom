package statechart

import (
	"sync"
)

// ChartRuntime represents a live instance of a ChartDefinition.
type ChartRuntime struct {
	id          RuntimeID
	definition  ChartDefinition
	state       RuntimeState
	activeState string              // Current state path for atomic/compound
	regionStates map[string]string  // Current state per region for parallel
	eventQueue  []Event
	appCtx      ApplicationContext
	runtimeCtx  RuntimeContext
	mu          sync.RWMutex

	// References to engine registries (set during creation)
	actions map[string]ActionFn
	guards  map[string]GuardFn
}

// ID returns the runtime's unique identifier.
func (cr *ChartRuntime) ID() RuntimeID {
	return cr.id
}

// State returns the current lifecycle state of the runtime.
func (cr *ChartRuntime) State() RuntimeState {
	cr.mu.RLock()
	defer cr.mu.RUnlock()
	return cr.state
}

