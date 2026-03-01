# Phase 2 Plan: Runtime-Statechart Integration (Locked Seam Compliant)

## User Story
As a Maelstrom developer, I want the runtime layer to integrate with the statechart engine while maintaining the Locked Seam architecture, so that ChartRuntimes can execute state transitions through the pure statechart library without coupling Maelstrom code to engine internals.

## Acceptance Criteria
1. Factory maintains the mapping between Maelstrom runtime IDs and statechart RuntimeIDs
2. Events flow from Maelstrom code → Factory → Engine (not through runtime objects)
3. The Locked Seam boundary is preserved: `pkg/statechart` never imports Maelstrom code
4. Actions/Guards can be registered on the engine from Maelstrom layer
5. Factory can spawn runtimes in the engine and receive their RuntimeIDs

## Frozen Semantics & API

### Factory API
```go
// Factory creates ChartRuntimes and manages their statechart integration
type Factory struct {
    reg           *registry.Registry
    engine        statechart.Library
    // Mapping: Maelstrom chart ID -> statechart RuntimeID
    statechartIDs map[string]statechart.RuntimeID
}

// WithEngine attaches a statechart engine (fluent config)
func (f *Factory) WithEngine(engine statechart.Library) *Factory

// Create instantiates a runtime AND spawns it in the engine
func (f *Factory) Create(key string) (*ChartRuntime, error)

// DispatchEvent sends an event to a runtime via the engine
func (f *Factory) DispatchEvent(chartID string, evt Event) error

// GetStatechartID returns the RuntimeID for a chart (test visibility)
func (f *Factory) GetStatechartID(chartID string) (statechart.RuntimeID, bool)
```

### ChartRuntime (unchanged - no engine reference)
```go
type ChartRuntime struct {
    ID      string
    Context RuntimeContext
    def     interface{}
    events  chan Event  // internal queue for non-engine events
    done    chan struct{}
    // NO engine reference here - Locked Seam boundary
}
```

## Executable Test Descriptions

### Test 1: Factory_WithEngine_ConfiguresEngine
```go
// Given: A factory without an engine
// When: WithEngine(engine) is called
// Then: The factory returns itself (fluent) and has the engine configured
```

### Test 2: Factory_Create_SpawnsInEngine
```go
// Given: A factory with engine and registry containing a statechart.ChartDefinition
// When: Create(key) is called
// Then:
//   - A ChartRuntime is returned with correct ID
//   - engine.Spawn() was called with the definition
//   - Factory stores the mapping: chartID -> RuntimeID
```

### Test 3: Factory_DispatchEvent_ForwardsToEngine
```go
// Given: Factory with engine, runtime created for chart "test"
// When: DispatchEvent("test", Event{Type: "foo"}) is called
// Then: engine.Dispatch() is called with the correct RuntimeID and event
```

### Test 4: Factory_Create_WithNonChartDefinition_ReturnsError
```go
// Given: Registry contains map[string]interface{} (not statechart.ChartDefinition)
// When: Create() is called
// Then: Returns error indicating type mismatch (cannot spawn in engine)
```

### Test 5: Factory_DispatchEvent_UnknownChart_ReturnsError
```go
// Given: Factory with no runtimes
// When: DispatchEvent("unknown", evt) is called
// Then: Returns error "chart not found: unknown"
```

### Test 6: Factory_CreateAll_MultipleRuntimes
```go
// Given: Registry with multiple statechart.ChartDefinitions
// When: CreateAll() is called
// Then: All runtimes created and spawned in engine, mappings stored
```

### Test 7: FullIntegration_EventCausesTransition
```go
// Given:
//   - Factory with real engine
//   - Runtime created with chart having transition: idle --event:go--> active
//   - Action registered on engine that sets a flag
// When: DispatchEvent(chartID, Event{Type: "go"})
// Then: Action was called, snapshot shows state is "active"
```

## Architecture Decisions

1. **Locked Seam Preservation**: Maelstrom's `ChartRuntime` never holds engine reference. Factory owns the mapping and orchestrates all engine calls.

2. **Type Safety**: Factory only spawns statechart.ChartDefinition types. Other types return error (ChartDefinition needed for engine.Spawn()).

3. **Testability**: Factory exposes GetStatechartID() for test verification without exposing internals.

4. **Backward Compatibility**: Factory works without engine (falls back to no-op or error for engine-dependent operations).

## Files to Modify

- `pkg/runtime/factory.go` - Add engine, mapping, WithEngine, Create, DispatchEvent
- `pkg/runtime/factory_test.go` - Add 7 test cases above
- `pkg/runtime/types.go` - NO CHANGES (keep Locked Seam clean)

## Implementation Order

1. Add engine field and WithEngine() to Factory
2. Add statechartIDs map and GetStatechartID() method
3. Modify Create() to spawn in engine when configured
4. Add DispatchEvent() method
5. Write tests 1-6 (unit tests)
6. Write test 7 (full integration)
7. Verify all 118+ tests pass
8. Commit with message: "feat: runtime-engine integration via Factory (Locked Seam)"
