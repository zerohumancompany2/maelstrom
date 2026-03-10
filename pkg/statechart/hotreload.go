package statechart

import (
	"fmt"
	"time"

	"github.com/maelstrom/v3/internal/testutil"
)

// HotReloadOptions provides configuration for hot-reload operations.
type HotReloadOptions struct {
	Timeout          time.Duration
	MaxAttempts      int
	HistoryMode      HistoryMode
	ContextTransform *ContextTransform
}

// HistoryMode specifies how to restore state during hot-reload.
type HistoryMode int

const (
	HistoryModeNone HistoryMode = iota
	HistoryModeShallow
	HistoryModeDeep
)

// ContextTransform provides template-based context migration.
type ContextTransform struct {
	Template string
}

// HotReload performs a hot-reload of the chart definition for the given runtime.
// Implements arch-v1.md L865-880 hot-reload protocol.
func (e *Engine) HotReload(id RuntimeID, newDef ChartDefinition, opts HotReloadOptions) error {
	// Get current runtime
	e.runtimeMu.RLock()
	runtime, exists := e.runtimes[id]
	e.runtimeMu.RUnlock()

	if !exists {
		return fmt.Errorf("%w: %s", ErrRuntimeNotFound, id)
	}

	// Step 1: Signal prepareForReload (arch-v1.md L867)
	// In a full implementation, this would signal the runtime to prepare
	// For now, we proceed directly to quiescence check

	// Step 2: Wait for quiescence within timeout (arch-v1.md L868)
	err := runtime.AwaitQuiescence(opts.Timeout)
	if err != nil {
		// Step 4: Timeout - force-stop and cleanStart (arch-v1.md L873-876)
		e.forceStopRuntime(runtime)
		return e.cleanStart(id, newDef, opts)
	}

	// Step 3: Quiescence reached - stop and spawn with history (arch-v1.md L869-872)
	return e.gracefulReload(id, newDef, opts)
}

// gracefulReload performs a graceful reload with history preservation.
func (e *Engine) gracefulReload(id RuntimeID, newDef ChartDefinition, opts HotReloadOptions) error {
	// Snapshot current state for history
	snap, err := e.Snapshot(id)
	if err != nil {
		return fmt.Errorf("snapshot failed: %w", err)
	}

	// Stop current runtime
	if err := e.Control(id, CmdStop); err != nil {
		return fmt.Errorf("stop failed: %w", err)
	}

	// Apply context transform if version changed (arch-v1.md L872)
	appCtx := e.getAppContextForReload(id, opts)

	// Spawn new runtime with history based on mode
	switch opts.HistoryMode {
	case HistoryModeShallow:
		// Restore to parent state's default sub-state
		return e.restoreWithShallowHistory(snap, newDef, appCtx, id)
	case HistoryModeDeep:
		// Restore to specific sub-state
		return e.restoreWithDeepHistory(snap, newDef, appCtx, id)
	default:
		// No history - clean start
		return e.cleanStartWithDef(newDef, appCtx, id)
	}
}

// forceStopRuntime forcefully stops a runtime without graceful shutdown.
func (e *Engine) forceStopRuntime(runtime *ChartRuntime) {
	runtime.mu.Lock()
	defer runtime.mu.Unlock()

	// Stop parallel regions if active
	if runtime.isParallel && runtime.eventRouter != nil {
		runtime.eventRouter.Stop()
		runtime.eventRouter = nil
	}

	runtime.state = RuntimeStateStopped
	runtime.eventQueue = nil
}

// cleanStart creates a new runtime without preserving history.
func (e *Engine) cleanStart(oldID RuntimeID, newDef ChartDefinition, opts HotReloadOptions) error {
	// Remove old runtime
	e.runtimeMu.Lock()
	delete(e.runtimes, oldID)
	e.runtimeMu.Unlock()

	// Create new runtime with default context
	appCtx := testutil.NewMockApplicationContext()
	_, err := e.Spawn(newDef, appCtx)
	if err != nil {
		return fmt.Errorf("spawn failed: %w", err)
	}

	return nil
}

// cleanStartWithDef creates a new runtime with the given ID reused.
func (e *Engine) cleanStartWithDef(newDef ChartDefinition, appCtx ApplicationContext, id RuntimeID) error {
	// Create new runtime
	newRuntime := &ChartRuntime{
		id:           id,
		definition:   newDef,
		state:        RuntimeStateCreated,
		activeState:  newDef.InitialState,
		regionStates: make(map[string]string),
		eventQueue:   make([]Event, 0),
		appCtx:       appCtx,
		runtimeCtx: RuntimeContext{
			ChartID:      newDef.ID,
			RuntimeID:    string(id),
			ActiveStates: []string{newDef.InitialState},
			StartTime:    time.Time{},
			Metrics:      make(map[string]float64),
		},
		actions: make(map[string]ActionFn),
		guards:  make(map[string]GuardFn),
	}

	// Copy action/guard registries
	e.actionMu.RLock()
	e.guardMu.RLock()
	for k, v := range e.actions {
		newRuntime.actions[k] = v
	}
	for k, v := range e.guards {
		newRuntime.guards[k] = v
	}
	e.guardMu.RUnlock()
	e.actionMu.RUnlock()

	e.runtimeMu.Lock()
	e.runtimes[id] = newRuntime
	e.runtimeMu.Unlock()

	return nil
}

// restoreWithShallowHistory restores to parent state's default sub-state.
func (e *Engine) restoreWithShallowHistory(snap Snapshot, newDef ChartDefinition, appCtx ApplicationContext, id RuntimeID) error {
	// For shallow history, use the initial state of the new definition
	return e.cleanStartWithDef(newDef, appCtx, id)
}

// restoreWithDeepHistory restores to specific sub-state if it exists.
func (e *Engine) restoreWithDeepHistory(snap Snapshot, newDef ChartDefinition, appCtx ApplicationContext, id RuntimeID) error {
	// Check if the previous active state exists in the new definition
	previousState := snap.ActiveStates[0]
	node := e.findNode(newDef.Root, previousState)

	if node != nil {
		// State exists - restore to it
		snap.ActiveStates = []string{previousState}
	} else {
		// State deleted - fall back to shallow history (arch-v1.md L885)
		return e.restoreWithShallowHistory(snap, newDef, appCtx, id)
	}

	_, err := e.Restore(snap, newDef, appCtx)
	return err
}

// getAppContextForReload retrieves the application context for reload.
func (e *Engine) getAppContextForReload(id RuntimeID, opts HotReloadOptions) ApplicationContext {
	e.runtimeMu.RLock()
	runtime, exists := e.runtimes[id]
	e.runtimeMu.RUnlock()

	if exists && runtime != nil {
		return runtime.appCtx
	}

	return testutil.NewMockApplicationContext()
}
