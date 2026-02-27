package statechart

// StateMachine executes statechart logic for a single region.
// Used by both top-level ChartRuntime and parallel RegionRuntime.
type StateMachine struct {
	definition    ChartDefinition
	activeState   string
	actions       map[string]ActionFn
	guards        map[string]GuardFn
	appCtx        ApplicationContext
	runtimeCtx    RuntimeContext
	TraceCallback TraceCallback // Optional observability hook
}

// EventResult contains the outcome of processing an event.
type EventResult struct {
	Transitioned bool
	FromState    string
	ToState      string
	IsFinalState bool
	EmitEvent    *Event
	Error        error
}

// ProcessEvent evaluates a single event against the current state.
// Returns EventResult for parent coordination.
func (sm *StateMachine) ProcessEvent(ev Event) EventResult {
	// Fire trace callback for event dispatch
	if sm.TraceCallback != nil {
		sm.TraceCallback.OnEventDispatch(RuntimeID(sm.runtimeCtx.RuntimeID), ev)
	}

	// Find the current node
	node := sm.findNode(sm.definition.Root, sm.activeState)
	if node == nil {
		return EventResult{Error: nil}
	}

	// Find matching transition
	for _, trans := range node.Transitions {
		if trans.Event != ev.Type {
			continue
		}

		// Evaluate guard if present
		if trans.Guard != "" {
			guard, exists := sm.guards[trans.Guard]
			if !exists {
				continue // Guard not found, skip transition
			}
			if !guard(sm.appCtx, ev) {
				continue // Guard blocked transition
			}
		}

		// Execute transition
		return sm.executeTransition(sm.activeState, trans.Target, ev, trans.Actions)
	}

	// No matching transition - this is valid (no-op)
	return EventResult{}
}

// executeTransition performs a state transition with entry/exit actions.
func (sm *StateMachine) executeTransition(fromPath, toPath string, ev Event, transitionActions []string) EventResult {
	// Fire trace callback for state exit
	if sm.TraceCallback != nil {
		sm.TraceCallback.OnStateExit(RuntimeID(sm.runtimeCtx.RuntimeID), fromPath, ev)
	}

	// Execute exit actions for current state
	_ = sm.executeExitActions(fromPath, ev)

	// Execute transition actions
	for _, actionName := range transitionActions {
		action, exists := sm.actions[actionName]
		if !exists {
			continue
		}
		_ = action(sm.runtimeCtx, sm.appCtx, ev)
	}

	// Update active state
	sm.activeState = toPath

	// Fire trace callback for transition
	if sm.TraceCallback != nil {
		sm.TraceCallback.OnTransition(RuntimeID(sm.runtimeCtx.RuntimeID), fromPath, toPath, ev)
	}

	// Fire trace callback for state entry
	if sm.TraceCallback != nil {
		sm.TraceCallback.OnStateEntry(RuntimeID(sm.runtimeCtx.RuntimeID), toPath, ev)
	}

	// Execute entry actions for new state
	_ = sm.executeEntryActions(toPath, ev)

	// Check if new state is final (atomic with no transitions)
	targetNode := sm.findNode(sm.definition.Root, toPath)
	isFinal := targetNode != nil && targetNode.NodeType() == NodeTypeAtomic && len(targetNode.Transitions) == 0

	return EventResult{
		Transitioned: true,
		FromState:    fromPath,
		ToState:      toPath,
		IsFinalState: isFinal,
	}
}

// findNode locates a node by its path.
func (sm *StateMachine) findNode(root *Node, path string) *Node {
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

// executeEntryActions runs all entry actions for a state.
func (sm *StateMachine) executeEntryActions(statePath string, ev Event) error {
	node := sm.findNode(sm.definition.Root, statePath)
	if node == nil {
		return nil
	}

	for _, actionName := range node.EntryActions {
		action, exists := sm.actions[actionName]
		if !exists {
			// Action not found, skip
			continue
		}
		if err := action(sm.runtimeCtx, sm.appCtx, ev); err != nil {
			// Error in action - log but don't block
			return err
		}
	}

	return nil
}

// executeExitActions runs all exit actions for a state.
func (sm *StateMachine) executeExitActions(statePath string, ev Event) error {
	node := sm.findNode(sm.definition.Root, statePath)
	if node == nil {
		return nil
	}

	for _, actionName := range node.ExitActions {
		action, exists := sm.actions[actionName]
		if !exists {
			// Action not found, skip
			continue
		}
		if err := action(sm.runtimeCtx, sm.appCtx, ev); err != nil {
			// Error in action - log but don't block
			return err
		}
	}

	return nil
}
