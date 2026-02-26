# Session Summary: Core Statechart Library v1

## Overview
Implemented the core statechart library (`pkg/statechart`) following the architecture specification in `docs/arch-v1.md` (Section 5).

## Files Created

| File | Description |
|------|-------------|
| `pkg/statechart/types.go` | Core type definitions (RuntimeID, Event, Node, Transition, etc.) |
| `pkg/statechart/interfaces.go` | Library and ApplicationContext interfaces |
| `pkg/statechart/runtime.go` | ChartRuntime struct for live statechart instances |
| `pkg/statechart/engine.go` | Engine implementing the Library interface |
| `pkg/statechart/statechart_test.go` | Comprehensive test suite (40+ tests) |
| `internal/testutil/mock_context.go` | Mock ApplicationContext for testing |
| `go.mod` | Go module definition |

## Implementation Details

### Node Types (Unified Primitive)
- **Atomic** (0 children): Leaf states with no substates
- **Compound** (1 child): Sequential substates with default initial state
- **Parallel** (≥2 children): Concurrent regions (structure only, full routing deferred)

### Core Features Implemented
1. **Lifecycle Management**: Spawn, Start, Pause, Resume, Stop
2. **Event Dispatch**: FIFO queue processing with run-to-completion semantics
3. **Transitions**: Event-driven state transitions with guards
4. **Action/Guard Registry**: Named function registration for chart definitions
5. **Compound States**: Automatic entry to default child states
6. **Entry/Exit Actions**: Proper ordering (exit before entry)
7. **Snapshot/Restore**: Runtime state capture and restoration

### Design Decisions

#### Action Error Handling
Actions that fail do NOT block transitions. The transition completes, and errors are logged (currently ignored). This matches the robustness principle for statecharts.

#### Unregistered Actions
Transition actions referencing unregistered names are silently skipped. This allows chart definitions to reference actions that may be registered later or in different contexts.

#### State Path Resolution
- Paths may optionally include the root node ID
- Compound state transitions automatically resolve to default children
- Path format: `"parent/child/grandchild"`

### Deferred to v2
- Sub-chart support (transient ChartRuntime spawn/join/abort)
- Dynamic reclassification (atomic ↔ compound/parallel at runtime)
- Parallel region event routing (`TargetPath` handling)
- ReplaceDefinition control command
- Trace callbacks for observability

## Test Coverage

| Category | Tests |
|----------|-------|
| Core Lifecycle | Spawn, Start, Pause, Resume, Stop |
| Event Dispatch | FIFO ordering, run-to-completion |
| Transitions | Guards, multiple transitions, no-match handling |
| Actions | Registration, execution, error handling |
| Compound States | Default child entry, nested transitions |
| Parallel States | Structure recognition (routing deferred) |
| Snapshot/Restore | State capture, ID uniqueness |
| Edge Cases | Non-existent runtime, invalid commands |

## API Surface

```go
type Library interface {
    Spawn(def ChartDefinition, initialAppCtx ApplicationContext) (RuntimeID, error)
    Control(id RuntimeID, cmd ControlCmd) error
    Dispatch(id RuntimeID, ev Event) error
    RegisterAction(name string, fn ActionFn)
    RegisterGuard(name string, fn GuardFn)
    Snapshot(id RuntimeID) (Snapshot, error)
    Restore(snap Snapshot) (RuntimeID, error)
}
```

## Compliance with arch-v1.md

### SHALL Requirements (Section 5.3)
- ✅ Evaluate events synchronously inside compound/parallel boundaries
- ❌ Run parallel regions in isolated goroutines (structure only, routing deferred)
- ❌ Fire trace callback (deferred to v2)
- ❌ Support dynamic reclassification (deferred to v2)
- ❌ Provide `replaceDefinition` hook (deferred to v2)
- ❌ Support transient sub-charts (deferred to v2)

### SHALL NOT Requirements (Section 5.4)
- ✅ SHALL NOT know about Mail, LLM, taints, boundaries, files, or agents
- ✅ SHALL NOT perform any I/O
- ✅ SHALL NOT implement persistence (only serialize/deserialize snapshots)

## Known Limitations
1. Parallel regions are recognized structurally but events are not routed to specific regions
2. Dynamic reclassification is not supported
3. Sub-charts (transient runtimes) are not supported
4. No trace/observability hooks
5. Snapshot ApplicationState is not serialized (nil in current implementation)

## Next Steps
1. Implement parallel region event routing with goroutine isolation
2. Add sub-chart support for transient ChartRuntime lifecycle
3. Implement dynamic reclassification for meta-programming
4. Add trace callbacks for observability
5. Implement ApplicationContext serialization for snapshots
