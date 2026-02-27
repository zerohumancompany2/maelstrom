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

	// Parallel state support
	eventRouter *EventRouter // Active when in parallel state
	isParallel  bool         // Whether currently in parallel state
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

