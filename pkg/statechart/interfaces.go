package statechart

// ApplicationContext provides user-scoped, read-write data bag with namespace isolation.
// This interface is implemented by the Maelstrom application layer and passed to actions/guards.
type ApplicationContext interface {
	// Get retrieves a value and its associated taints from the context.
	// The callerBoundary is used for access control (enforced by Maelstrom, not the library).
	Get(key string, callerBoundary string) (any, []string, error)

	// Set stores a value with associated taints in the context.
	// The callerBoundary is used for access control (enforced by Maelstrom, not the library).
	Set(key string, value any, taints []string, callerBoundary string) error

	// Namespace returns the namespace identifier for this context.
	Namespace() string
}

// ActionFn is the signature for actions executed on entry/exit or transitions.
// The action receives runtime context (read-only), application context (read-write),
// and the initiating event.
type ActionFn func(runtimeCtx RuntimeContext, appCtx ApplicationContext, event Event) error

// GuardFn is the signature for guards that determine if a transition may fire.
// Guards receive read-only application context and the initiating event.
type GuardFn func(appCtx ApplicationContext, event Event) bool

// Library is the pure statechart library interface.
// It provides lifecycle management, event dispatch, action/guard registration,
// and persistence (snapshot/restore) for statechart runtimes.
type Library interface {
	// Spawn creates a new ChartRuntime from the given definition.
	// Returns the unique RuntimeID for the created instance.
	Spawn(def ChartDefinition, initialAppCtx ApplicationContext) (RuntimeID, error)

	// Control sends a lifecycle command to a ChartRuntime.
	// Commands: start, pause, resume, stop, snapshot, injectEvent, replaceDefinition
	Control(id RuntimeID, cmd ControlCmd) error

	// Dispatch sends an event to a ChartRuntime's event queue.
	// Events are processed according to the runtime's current state and configuration.
	Dispatch(id RuntimeID, ev Event) error

	// RegisterAction registers a named action function that can be referenced in chart definitions.
	RegisterAction(name string, fn ActionFn)

	// RegisterGuard registers a named guard function that can be referenced in chart definitions.
	RegisterGuard(name string, fn GuardFn)

	// Snapshot captures the current state of a ChartRuntime.
	Snapshot(id RuntimeID) (Snapshot, error)

	// Restore creates a new ChartRuntime from a previously captured snapshot.
	// Returns the new RuntimeID (different from the original).
	Restore(snap Snapshot) (RuntimeID, error)
}

// TraceCallback is called by the library on lifecycle events for observability.
// This is an optional extension point for monitoring/debugging.
type TraceCallback interface {
	OnTransition(runtimeID RuntimeID, fromState, toState string, event Event)
	OnStateEntry(runtimeID RuntimeID, state string, event Event)
	OnStateExit(runtimeID RuntimeID, state string, event Event)
	OnEventDispatch(runtimeID RuntimeID, event Event)
}
