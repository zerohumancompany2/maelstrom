package statechart

import (
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

var (
	// ErrRuntimeNotFound is returned when a runtime ID doesn't exist.
	ErrRuntimeNotFound = errors.New("runtime not found")

	// ErrInvalidControlCmd is returned when an invalid control command is issued.
	ErrInvalidControlCmd = errors.New("invalid control command")

	// ErrRuntimeNotRunning is returned when an operation requires a running runtime.
	ErrRuntimeNotRunning = errors.New("runtime not running")

	// ErrInvalidState is returned when a state transition is invalid.
	ErrInvalidState = errors.New("invalid runtime state for operation")

	// ErrActionNotFound is returned when a referenced action is not registered.
	ErrActionNotFound = errors.New("action not found")

	// ErrGuardNotFound is returned when a referenced guard is not registered.
	ErrGuardNotFound = errors.New("guard not found")
)

// Engine implements the Library interface.
type Engine struct {
	runtimes    map[RuntimeID]*ChartRuntime
	runtimeMu   sync.RWMutex
	actions     map[string]ActionFn
	guards      map[string]GuardFn
	actionMu    sync.RWMutex
	guardMu     sync.RWMutex
	idCounter   uint64
}

// NewEngine creates a new statechart engine.
func NewEngine() Library {
	return &Engine{
		runtimes: make(map[RuntimeID]*ChartRuntime),
		actions:  make(map[string]ActionFn),
		guards:   make(map[string]GuardFn),
	}
}

// generateID creates a new unique RuntimeID.
func (e *Engine) generateID() RuntimeID {
	count := atomic.AddUint64(&e.idCounter, 1)
	return RuntimeID(fmt.Sprintf("rt-%d", count))
}

// Spawn creates a new ChartRuntime from the given definition.
func (e *Engine) Spawn(def ChartDefinition, initialAppCtx ApplicationContext) (RuntimeID, error) {
	runtime := &ChartRuntime{
		id:           e.generateID(),
		definition:   def,
		state:        RuntimeStateCreated,
		activeState:  def.InitialState,
		regionStates: make(map[string]string),
		eventQueue:   make([]Event, 0),
		appCtx:       initialAppCtx,
		runtimeCtx: RuntimeContext{
			ChartID:   def.ID,
			RuntimeID: "", // set below
			ActiveStates: []string{def.InitialState},
			StartTime:    time.Time{}, // set on Start
			Metrics:      make(map[string]float64),
		},
	}

	// Copy action/guard registries to runtime
	e.actionMu.RLock()
	e.guardMu.RLock()
	runtime.actions = make(map[string]ActionFn, len(e.actions))
	runtime.guards = make(map[string]GuardFn, len(e.guards))
	for k, v := range e.actions {
		runtime.actions[k] = v
	}
	for k, v := range e.guards {
		runtime.guards[k] = v
	}
	e.guardMu.RUnlock()
	e.actionMu.RUnlock()

	// Set runtime ID in context
	runtime.runtimeCtx.RuntimeID = string(runtime.id)

	// Store runtime
	e.runtimeMu.Lock()
	e.runtimes[runtime.id] = runtime
	e.runtimeMu.Unlock()

	return runtime.id, nil
}

// SpawnTransient creates a sub-chart with lifecycle binding to parent.
// When parent stops, all transient children are automatically stopped.
func (e *Engine) SpawnTransient(def ChartDefinition, appCtx ApplicationContext, parentID RuntimeID) (RuntimeID, error) {
	// Verify parent exists
	e.runtimeMu.RLock()
	parent, exists := e.runtimes[parentID]
	e.runtimeMu.RUnlock()

	if !exists {
		return "", fmt.Errorf("%w: parent %s", ErrRuntimeNotFound, parentID)
	}

	// Create child runtime
	childID, err := e.Spawn(def, appCtx)
	if err != nil {
		return "", err
	}

	// Register child with parent
	parent.mu.Lock()
	parent.children = append(parent.children, childID)
	parent.mu.Unlock()

	return childID, nil
}

// ReplaceDefinition hot-reloads a new chart definition into an existing runtime.
// Preserves runtime state and context where possible.
func (e *Engine) ReplaceDefinition(id RuntimeID, newDef ChartDefinition) error {
	e.runtimeMu.RLock()
	runtime, exists := e.runtimes[id]
	e.runtimeMu.RUnlock()

	if !exists {
		return fmt.Errorf("%w: %s", ErrRuntimeNotFound, id)
	}

	runtime.mu.Lock()
	defer runtime.mu.Unlock()

	// Update definition
	runtime.definition = newDef

	return nil
}

// Control sends a lifecycle command to a ChartRuntime.
func (e *Engine) Control(id RuntimeID, cmd ControlCmd) error {
	e.runtimeMu.RLock()
	runtime, exists := e.runtimes[id]
	e.runtimeMu.RUnlock()

	if !exists {
		return fmt.Errorf("%w: %s", ErrRuntimeNotFound, id)
	}

	switch cmd {
	case CmdStart:
		return e.startRuntime(runtime)
	case CmdPause:
		return e.pauseRuntime(runtime)
	case CmdResume:
		return e.resumeRuntime(runtime)
	case CmdStop:
		return e.stopRuntime(runtime)
	case CmdSnapshot:
		// Snapshot is synchronous, doesn't change state
		return nil
	default:
		return fmt.Errorf("%w: %s", ErrInvalidControlCmd, cmd)
	}
}

func (e *Engine) startRuntime(runtime *ChartRuntime) error {
	runtime.mu.Lock()
	defer runtime.mu.Unlock()

	if runtime.state != RuntimeStateCreated {
		return fmt.Errorf("%w: cannot start from %s", ErrInvalidState, runtime.state)
	}

	runtime.state = RuntimeStateRunning
	runtime.runtimeCtx.StartTime = time.Now()

	// Execute entry actions for initial state
	if err := e.executeEntryActions(runtime, runtime.activeState, Event{}); err != nil {
		return err
	}

	return nil
}

func (e *Engine) pauseRuntime(runtime *ChartRuntime) error {
	runtime.mu.Lock()
	defer runtime.mu.Unlock()

	if runtime.state != RuntimeStateRunning {
		return fmt.Errorf("%w: cannot pause from %s", ErrInvalidState, runtime.state)
	}

	runtime.state = RuntimeStatePaused

	// Broadcast pause to parallel regions
	if runtime.isParallel && runtime.eventRouter != nil {
		for _, region := range runtime.eventRouter.regions {
			region.inputChan <- Event{Type: SysPause}
		}
	}

	return nil
}

func (e *Engine) resumeRuntime(runtime *ChartRuntime) error {
	runtime.mu.Lock()
	defer runtime.mu.Unlock()

	if runtime.state != RuntimeStatePaused {
		return fmt.Errorf("%w: cannot resume from %s", ErrInvalidState, runtime.state)
	}

	runtime.state = RuntimeStateRunning

	// Broadcast resume to parallel regions
	if runtime.isParallel && runtime.eventRouter != nil {
		for _, region := range runtime.eventRouter.regions {
			region.inputChan <- Event{Type: SysResume}
		}
	}

	// Process any queued events
	go e.processEventQueue(runtime)
	return nil
}

func (e *Engine) stopRuntime(runtime *ChartRuntime) error {
	runtime.mu.Lock()

	if runtime.state == RuntimeStateStopped {
		runtime.mu.Unlock()
		return fmt.Errorf("%w: already stopped", ErrInvalidState)
	}

	// Stop all transient children first
	children := make([]RuntimeID, len(runtime.children))
	copy(children, runtime.children)

	// Execute exit actions for current state
	if err := e.executeExitActions(runtime, runtime.activeState, Event{}); err != nil {
		// Log error but continue with stop
		_ = err
	}

	runtime.state = RuntimeStateStopped
	runtime.eventQueue = nil
	runtime.mu.Unlock()

	// Stop children after unlocking to avoid deadlock
	for _, childID := range children {
		if child, exists := e.runtimes[childID]; exists {
			e.stopRuntime(child)
		}
	}

	// Remove runtime from engine after unlocking
	e.runtimeMu.Lock()
	delete(e.runtimes, runtime.id)
	e.runtimeMu.Unlock()

	return nil
}

// Dispatch sends an event to a ChartRuntime's event queue.
func (e *Engine) Dispatch(id RuntimeID, ev Event) error {
	e.runtimeMu.RLock()
	runtime, exists := e.runtimes[id]
	e.runtimeMu.RUnlock()

	if !exists {
		return fmt.Errorf("%w: %s", ErrRuntimeNotFound, id)
	}

	runtime.mu.Lock()

	if runtime.state != RuntimeStateRunning {
		runtime.mu.Unlock()
		return fmt.Errorf("%w: runtime is %s", ErrRuntimeNotRunning, runtime.state)
	}

	// Add to event queue
	runtime.eventQueue = append(runtime.eventQueue, ev)
	isEmpty := len(runtime.eventQueue) == 1
	runtime.mu.Unlock()

	// Only start processing if this was the first event (no other goroutine processing)
	if isEmpty {
		go e.processEventQueue(runtime)
	}

	return nil
}

// processEventQueue processes events in FIFO order.
func (e *Engine) processEventQueue(runtime *ChartRuntime) {
	for {
		runtime.mu.Lock()

		if len(runtime.eventQueue) == 0 || runtime.state != RuntimeStateRunning {
			runtime.mu.Unlock()
			return
		}

		ev := runtime.eventQueue[0]
		runtime.eventQueue = runtime.eventQueue[1:]
		runtime.mu.Unlock()

		if err := e.processEvent(runtime, ev); err != nil {
			// Error handling: log and continue
			_ = err
		}
	}
}

// processEvent evaluates a single event against the current state configuration.
func (e *Engine) processEvent(runtime *ChartRuntime, ev Event) error {
	// Find the current node
	currentPath := runtime.activeState
	node := e.findNode(runtime.definition.Root, currentPath)
	if node == nil {
		return fmt.Errorf("current state node not found: %s", currentPath)
	}

	// Find matching transition on current node
	for _, trans := range node.Transitions {
		if trans.Event != ev.Type {
			continue
		}

		// Evaluate guard if present
		if trans.Guard != "" {
			guard, exists := runtime.guards[trans.Guard]
			if !exists {
				return fmt.Errorf("%w: %s", ErrGuardNotFound, trans.Guard)
			}
			if !guard(runtime.appCtx, ev) {
				continue // Guard blocked transition
			}
		}

		// Execute transition
		return e.executeTransition(runtime, currentPath, trans.Target, ev, trans.Actions)
	}

	// No matching transition on current node
	// If in parallel state, route to parallel regions
	if node.NodeType() == NodeTypeParallel {
		runtime.processParallelEvent(ev)
		return nil
	}

	return nil
}

// executeTransition performs a state transition with entry/exit actions.
// Transition completes even if actions fail - errors are logged but don't block.
func (e *Engine) executeTransition(runtime *ChartRuntime, fromPath, toPath string, ev Event, actions []string) error {
	// Find source node for exit actions
	fromNode := e.findNode(runtime.definition.Root, fromPath)

	// If exiting a parallel state, execute parallel exit actions
	if fromNode != nil && fromNode.NodeType() == NodeTypeParallel {
		runtime.exitParallelState()
		_ = e.executeExitActions(runtime, fromPath, ev)
	} else {
		// Execute exit actions for current state (errors don't block transition)
		_ = e.executeExitActions(runtime, fromPath, ev)
	}

	// Execute transition actions (errors don't block transition)
	for _, actionName := range actions {
		action, exists := runtime.actions[actionName]
		if !exists {
			// Continue even if action not found
			continue
		}
		_ = action(runtime.runtimeCtx, runtime.appCtx, ev)
	}

	// Resolve target state path (handle compound states by entering default child)
	finalPath := e.resolveTargetState(runtime, toPath)

	// Update active state
	runtime.activeState = finalPath
	runtime.runtimeCtx.ActiveStates = []string{finalPath}

	// Execute entry actions for new state (errors don't block transition)
	_ = e.executeEntryActions(runtime, finalPath, ev)

	// Check if we entered a parallel state and initialize it
	targetNode := e.findNode(runtime.definition.Root, finalPath)
	if targetNode != nil && targetNode.NodeType() == NodeTypeParallel {
		runtime.enterParallelState(finalPath)
	}

	return nil
}

// resolveTargetState resolves the target state path.
// For compound states, it enters the default child.
func (e *Engine) resolveTargetState(runtime *ChartRuntime, targetPath string) string {
	node := e.findNode(runtime.definition.Root, targetPath)
	if node == nil {
		return targetPath
	}

	// For compound states (1 child), enter the default child
	if node.NodeType() == NodeTypeCompound && len(node.Children) > 0 {
		// Find the default child (marked with IsInitial or first child)
		for name, child := range node.Children {
			if child.IsInitial {
				return targetPath + "/" + name
			}
			_ = name
		}
		// If no child marked as initial, use first child
		for name := range node.Children {
			return targetPath + "/" + name
		}
	}

	return targetPath
}

// executeEntryActions runs all entry actions for a state.
func (e *Engine) executeEntryActions(runtime *ChartRuntime, statePath string, ev Event) error {
	node := e.findNode(runtime.definition.Root, statePath)
	if node == nil {
		return nil
	}

	for _, actionName := range node.EntryActions {
		action, exists := runtime.actions[actionName]
		if !exists {
			return fmt.Errorf("%w: %s", ErrActionNotFound, actionName)
		}
		if err := action(runtime.runtimeCtx, runtime.appCtx, ev); err != nil {
			return err
		}
	}

	return nil
}

// executeExitActions runs all exit actions for a state.
func (e *Engine) executeExitActions(runtime *ChartRuntime, statePath string, ev Event) error {
	node := e.findNode(runtime.definition.Root, statePath)
	if node == nil {
		return nil
	}

	for _, actionName := range node.ExitActions {
		action, exists := runtime.actions[actionName]
		if !exists {
			return fmt.Errorf("%w: %s", ErrActionNotFound, actionName)
		}
		if err := action(runtime.runtimeCtx, runtime.appCtx, ev); err != nil {
			return err
		}
	}

	return nil
}

// findNode locates a node by its path (e.g., "root/child/grandchild").
// Paths may optionally start with the root node's ID.
func (e *Engine) findNode(root *Node, path string) *Node {
	if path == "" || path == root.ID {
		return root
	}

	parts := splitPath(path)
	current := root

	// If path starts with root ID, skip it
	if len(parts) > 0 && parts[0] == root.ID {
		parts = parts[1:]
	}

	for _, part := range parts {
		if current.Children == nil {
			return nil
		}
		child, exists := current.Children[part]
		if !exists {
			return nil
		}
		current = child
	}

	return current
}

// splitPath breaks a path into components.
func splitPath(path string) []string {
	// Simple implementation - assumes no leading/trailing slashes
	var parts []string
	current := ""
	for _, ch := range path {
		if ch == '/' {
			if current != "" {
				parts = append(parts, current)
				current = ""
			}
		} else {
			current += string(ch)
		}
	}
	if current != "" {
		parts = append(parts, current)
	}
	return parts
}

// RegisterAction registers a named action function.
func (e *Engine) RegisterAction(name string, fn ActionFn) {
	e.actionMu.Lock()
	defer e.actionMu.Unlock()
	e.actions[name] = fn
}

// RegisterGuard registers a named guard function.
func (e *Engine) RegisterGuard(name string, fn GuardFn) {
	e.guardMu.Lock()
	defer e.guardMu.Unlock()
	e.guards[name] = fn
}

// Snapshot captures the current state of a ChartRuntime.
func (e *Engine) Snapshot(id RuntimeID) (Snapshot, error) {
	e.runtimeMu.RLock()
	runtime, exists := e.runtimes[id]
	e.runtimeMu.RUnlock()

	if !exists {
		return Snapshot{}, fmt.Errorf("%w: %s", ErrRuntimeNotFound, id)
	}

	runtime.mu.RLock()
	defer runtime.mu.RUnlock()

	// Copy event queue
	queueCopy := make([]Event, len(runtime.eventQueue))
	copy(queueCopy, runtime.eventQueue)

	// Copy region states for parallel state
	regionStatesCopy := make(map[string]string, len(runtime.regionStates))
	for k, v := range runtime.regionStates {
		regionStatesCopy[k] = v
	}

	return Snapshot{
		RuntimeID:        runtime.id,
		DefinitionID:     runtime.definition.ID,
		ActiveStates:     []string{runtime.activeState},
		EventQueue:       queueCopy,
		RuntimeContext:   runtime.runtimeCtx,
		ApplicationState: nil, // Would be populated by serializing appCtx
		RegionStates:     regionStatesCopy,
		IsParallel:       runtime.isParallel,
	}, nil
}

// Restore creates a new ChartRuntime from a snapshot.
func (e *Engine) Restore(snap Snapshot, def ChartDefinition, appCtx ApplicationContext) (RuntimeID, error) {
	// Create a new runtime with restored state
	newRuntime := &ChartRuntime{
		id:           e.generateID(),
		definition:   def,
		state:        RuntimeStateCreated,
		activeState:  snap.ActiveStates[0],
		regionStates: snap.RegionStates,
		eventQueue:   snap.EventQueue,
		appCtx:       appCtx,
		runtimeCtx:   snap.RuntimeContext,
		actions:      make(map[string]ActionFn),
		guards:       make(map[string]GuardFn),
		isParallel:   snap.IsParallel,
	}

	// Copy action/guard registries from engine
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

	newRuntime.runtimeCtx.RuntimeID = string(newRuntime.id)

	e.runtimeMu.Lock()
	e.runtimes[newRuntime.id] = newRuntime
	e.runtimeMu.Unlock()

	return newRuntime.id, nil
}
