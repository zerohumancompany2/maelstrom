package statechart

import (
	"errors"
	"strings"
	"sync"
	"time"
)

// ErrQuiescenceTimeout is returned when awaiting quiescence times out.
var ErrQuiescenceTimeout = errors.New("timeout awaiting quiescence")

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

	// Parallel state support
	eventRouter *EventRouter // Active when in parallel state
	isParallel  bool         // Whether currently in parallel state

	// Transient sub-chart support (lifecycle binding to parent)
	children []RuntimeID // IDs of child runtimes spawned via SpawnTransient
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

// enterParallelState initializes parallel region execution.
func (cr *ChartRuntime) enterParallelState(parallelPath string) {
	cr.mu.Lock()
	defer cr.mu.Unlock()

	// Find the parallel node
	node := cr.findNode(cr.definition.Root, parallelPath)
	if node == nil || node.NodeType() != NodeTypeParallel {
		return
	}

	// Build region definitions from children
	regionDefs := make(map[string]ChartDefinition)
	for name, childNode := range node.Children {
		regionDefs[name] = ChartDefinition{
			ID:           name,
			Root:         childNode,
			InitialState: childNode.ID,
		}
	}

	// Create and start event router
	cr.eventRouter = NewEventRouter(regionDefs, cr.actions, cr.guards, cr.appCtx)
	cr.isParallel = true

	// Start router in background
	go cr.eventRouter.Run()
}

// processParallelEvent routes an event to parallel regions.
func (cr *ChartRuntime) processParallelEvent(ev Event) {
	cr.mu.RLock()
	router := cr.eventRouter
	cr.mu.RUnlock()

	if router == nil {
		return
	}

	router.Send(ev)
}

// exitParallelState cleans up parallel region execution.
func (cr *ChartRuntime) exitParallelState() {
	cr.mu.Lock()
	defer cr.mu.Unlock()

	if cr.eventRouter != nil {
		cr.eventRouter.Stop()
		cr.eventRouter = nil
	}
	cr.isParallel = false
}

// findNode locates a node by its path (helper for runtime).
func (cr *ChartRuntime) findNode(root *Node, path string) *Node {
	if path == "" || path == root.ID {
		return root
	}

	parts := splitPath(path)
	current := root

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

// IsQuiescent returns true if the runtime is in a stable state for snapshot.
// Quiescence requires: empty event queue, no active parallel regions processing.
func (cr *ChartRuntime) IsQuiescent() bool {
	cr.mu.RLock()
	defer cr.mu.RUnlock()

	// Check event queue is empty
	if len(cr.eventQueue) > 0 {
		return false
	}

	// Check parallel regions are idle
	if cr.isParallel && cr.eventRouter != nil {
		// TODO: Check if regions have pending work
		// For now, parallel state is considered non-quiescent
		return false
	}

	return true
}

// AwaitQuiescence blocks until the runtime becomes quiescent or timeout.
// Returns error if timeout expires before quiescence.
func (cr *ChartRuntime) AwaitQuiescence(timeout time.Duration) error {
	timer := time.NewTimer(timeout)
	defer timer.Stop()

	ticker := time.NewTicker(10 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-timer.C:
			return ErrQuiescenceTimeout
		case <-ticker.C:
			if cr.IsQuiescent() {
				return nil
			}
		}
	}
}

// ReclassifyState dynamically changes a state's classification at runtime.
// This implements arch-v1.md Section 2.2 dynamic reclassification.
// Pass nil children to make atomic, single child for compound, multiple for parallel.
func (cr *ChartRuntime) ReclassifyState(statePath string, children interface{}) {
	var needsParallelInit bool

	cr.mu.Lock()

	node := cr.findNode(cr.definition.Root, statePath)
	if node == nil {
		cr.mu.Unlock()
		return
	}

	switch v := children.(type) {
	case nil:
		// Atomic: remove all children
		node.Children = nil
	case *Node:
		// Compound: single child
		node.Children = map[string]*Node{v.ID: v}
	case map[string]*Node:
		// Parallel: multiple children
		node.Children = v
		// Check if we need to initialize parallel regions
		needsParallelInit = (cr.activeState == statePath || strings.HasPrefix(cr.activeState, statePath+"/")) && len(v) >= 2
	}

	// Update active state based on new node type
	// Check if activeState is this state or a descendant (starts with statePath)
	if cr.activeState == statePath || strings.HasPrefix(cr.activeState, statePath+"/") {
		switch node.NodeType() {
		case NodeTypeCompound:
			// Find initial child
			for _, child := range node.Children {
				if child.IsInitial {
					cr.activeState = statePath + "/" + child.ID
					break
				}
			}
		case NodeTypeAtomic:
			// Strip any child suffix
			cr.activeState = statePath
		}
	}

	cr.mu.Unlock()

	// Initialize parallel regions after releasing lock
	if needsParallelInit {
		cr.enterParallelState(statePath)
	}
}