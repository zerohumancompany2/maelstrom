# Parallel States Implementation Plan

## Overview
Implement full parallel state support for the Maelstrom statechart library, including region isolation, event routing, completion tracking, and force-stop semantics.

## Semantics Reference (from arch-v1.md)

### Event Routing
- **Omitted TargetPath** = "self" (current region context), not broadcast
- **Explicit routing** via `region:foo`, `child:bar`, `child:bar/region:baz`
- Cross-region communication = explicit event dispatch with TargetPath

### Parallel State Lifecycle
| Phase            | Semantics                                                                     |
| ---------------- | ----------------------------------------------------------------------------- |
| **Entry**        | ALL regions simultaneously enter their initial states                         |
| **Execution**    | Regions run in isolated goroutines, process events independently              |
| **Natural Exit** | ALL regions reach final state → automatic done events → parent can transition |
| **Forced Exit**  | timeout/abort via ControlCmd → force-stop all regions                         |

### State Representation
- `ActiveStates []string` contains one StatePath per orthogonal region
- `regionStates map[string]string` tracks current state per region name

### Completion Model
- **Automatic completion events** - regions emit implicit `done` when reaching final state
- **Transient sub-charts** can be aborted without final state
- **Persistent regions** may never complete (e.g., monitoring loops)

### Cross-Region Coordination
- **No direct coupling** - transitions in regionA don't directly affect regionB
- **Event propagation** - regions communicate via explicit events routed through parent
- **Exit barrier** - parent waits for all regions' done events OR force-stop

---

## Implementation Steps

### Step 1: Add Final State Support to Node

**File**: `pkg/statechart/types.go`

Add `IsFinal` field to Node struct:
```go
type Node struct {
    ID            string
    Children      map[string]*Node
    RegionNames   []string
    Transitions   []Transition
    EntryActions  []string
    ExitActions   []string
    IsInitial     bool
    IsFinal       bool  // NEW: marks a final state
}
```

When a region enters a final state, it emits an implicit completion event to the parent.

---

### Step 2: Extend Event Structure for Completion

**File**: `pkg/statechart/types.go`

Add implicit event types:
```go
const (
    EventTypeImplicitDone = "sys:done"  // emitted when region reaches final state
)
```

The Event struct already has `TargetPath`; completion events from regions will target the parent parallel state.

---

### Step 3: Create RegionRuntime Structure

**File**: `pkg/statechart/runtime.go` (or new file)

```go
// RegionRuntime represents a single parallel region's execution context
type RegionRuntime struct {
    name         string
    currentState string
    eventQueue   chan Event
    done         chan struct{}
    parentChan   chan RegionMessage
    runtimeCtx   RuntimeContext
    appCtx       ApplicationContext
    actions      map[string]ActionFn
    guards       map[string]GuardFn
}

type RegionMessage struct {
    Type     string      // "transition", "done", "error"
    Region   string
    Payload  interface{}
}
```

---

### Step 4: Extend ChartRuntime for Parallel Support

**File**: `pkg/statechart/runtime.go`

```go
type ChartRuntime struct {
    id            RuntimeID
    definition    ChartDefinition
    state         RuntimeState
    activeState   string                     // for atomic/compound (backward compat)
    regionStates  map[string]string          // NEW: state per region
    regions       map[string]*RegionRuntime  // NEW: region goroutines
    eventQueue    []Event
    appCtx        ApplicationContext
    runtimeCtx    RuntimeContext
    mu            sync.RWMutex
    actions       map[string]ActionFn
    guards        map[string]GuardFn

    // NEW: completion tracking
    completedRegions map[string]bool  // tracks which regions reached final state
    forceStop        bool             // signal to force exit without waiting for completion
}
```

---

### Step 5: Implement TargetPath Parsing

**File**: `pkg/statechart/engine.go`

```go
// TargetPath represents a parsed routing address
type TargetPath struct {
    Region   string  // "regionA" if TargetPath was "region:regionA"
    Child    string  // for sub-chart routing (deferred to v3)
    IsSelf   bool    // true if TargetPath omitted or "."
}

func parseTargetPath(path string) (TargetPath, error) {
    if path == "" || path == "." {
        return TargetPath{IsSelf: true}, nil
    }
    // Parse "region:foo" or "child:bar/region:baz"
    // Return error for unsupported formats
}
```

---

### Step 6: Implement Parallel State Entry

**File**: `pkg/statechart/engine.go`

Modify `resolveTargetState` to handle parallel entry:

```go
func (e *Engine) enterParallelState(runtime *ChartRuntime, node *Node, path string, ev Event) error {
    // 1. Initialize ALL regions simultaneously
    for regionName, childNode := range node.Children {
        // Find initial child of this region
        initialChild := findInitialChild(childNode)
        regionPath := path + "/" + regionName + "/" + initialChild.ID

        // Create RegionRuntime
        region := &RegionRuntime{
            name:         regionName,
            currentState: regionPath,
            eventQueue:   make(chan Event, 100),
            done:         make(chan struct{}),
            parentChan:   make(chan RegionMessage, 10),
            // ... copy context and registries
        }

        runtime.regions[regionName] = region
        runtime.regionStates[regionName] = regionPath

        // Start region goroutine
        go e.runRegion(region)

        // Execute entry actions for initial state
        e.executeEntryActions(runtime, regionPath, ev)
    }

    // 2. Update ActiveStates to include all regions
    runtime.updateActiveStates()

    return nil
}
```

---

### Step 7: Implement Region Goroutine

**File**: `pkg/statechart/engine.go`

```go
func (e *Engine) runRegion(region *RegionRuntime) {
    defer close(region.done)

    for {
        select {
        case ev := <-region.eventQueue:
            // Process event against region's current state
            err := e.processRegionEvent(region, ev)
            if err != nil {
                region.parentChan <- RegionMessage{
                    Type:   "error",
                    Region: region.name,
                    Payload: err,
                }
            }

        case <-region.parentChan: // check for shutdown signal
            // Handle force-stop
            return
        }
    }
}

func (e *Engine) processRegionEvent(region *RegionRuntime, ev Event) error {
    // Similar to processEvent but scoped to region
    // If transition leads to final state, emit done event:
    if targetNode.IsFinal {
        region.parentChan <- RegionMessage{
            Type:   "done",
            Region: region.name,
        }
    }
    return nil
}
```

---

### Step 8: Implement Event Routing

**File**: `pkg/statechart/engine.go`

Modify `Dispatch` to route based on TargetPath:

```go
func (e *Engine) Dispatch(id RuntimeID, ev Event) error {
    // ... get runtime

    target, err := parseTargetPath(ev.TargetPath)
    if err != nil {
        return err
    }

    if target.IsSelf {
        // Route to current context - if in parallel state,
        // determine which region based on caller context
        // For now: route to all regions (broadcast) if in parallel state
        // This matches "evaluate against all regions" interpretation
        runtime.routeToAllRegions(ev)
    } else if target.Region != "" {
        // Route to specific region
        region, exists := runtime.regions[target.Region]
        if !exists {
            return fmt.Errorf("region not found: %s", target.Region)
        }
        region.eventQueue <- ev
    }

    return nil
}
```

**DECISION NEEDED**: Does "self" mean broadcast to all regions, or only the region that dispatched the event? Current implementation leans toward broadcast for events that should affect the whole parallel state.

---

### Step 9: Implement Completion Tracking

**File**: `pkg/statechart/engine.go`

```go
func (e *Engine) handleRegionMessage(runtime *ChartRuntime, msg RegionMessage) {
    switch msg.Type {
    case "done":
        runtime.completedRegions[msg.Region] = true

        // Check if ALL regions completed
        if runtime.allRegionsCompleted() {
            // Parent parallel state can now transition
            // Emit internal completion or allow transition guards to check
        }

    case "transition":
        // Update regionStates
        runtime.regionStates[msg.Region] = msg.Payload.(string)
        runtime.updateActiveStates()

    case "error":
        // Handle region error - log, propagate, or stop
    }
}

func (cr *ChartRuntime) allRegionsCompleted() bool {
    for regionName := range cr.regions {
        if !cr.completedRegions[regionName] {
            return false
        }
    }
    return true
}
```

---

### Step 10: Implement Parallel State Exit

**File**: `pkg/statechart/engine.go`

```go
func (e *Engine) exitParallelState(runtime *ChartRuntime, ev Event) error {
    // 1. Signal all regions to stop
    for _, region := range runtime.regions {
        close(region.done) // or send shutdown signal
    }

    // 2. Wait for regions to finish (with timeout for force-stop)
    // Or don't wait if forceStop flag is set

    // 3. Execute exit actions for all regions (in parallel, no order guarantee)
    var wg sync.WaitGroup
    for _, region := range runtime.regions {
        wg.Add(1)
        go func(r *RegionRuntime) {
            defer wg.Done()
            e.executeExitActions(runtime, r.currentState, ev)
        }(region)
    }
    wg.Wait()

    // 4. Clean up
    runtime.regions = nil
    runtime.regionStates = nil
    runtime.completedRegions = nil

    return nil
}
```

---

### Step 11: Extend Control Commands

**File**: `pkg/statechart/engine.go`

Add timeout/force-stop support to existing commands:

```go
func (e *Engine) stopRuntime(runtime *ChartRuntime) error {
    runtime.mu.Lock()

    // Set forceStop flag to allow exit before completion
    runtime.forceStop = true

    // If in parallel state, trigger exit
    if runtime.regions != nil {
        e.exitParallelState(runtime, Event{})
    }

    runtime.state = RuntimeStateStopped
    // ... cleanup
}
```

---

### Step 12: Update Snapshot/Restore

**File**: `pkg/statechart/engine.go`

```go
func (e *Engine) Snapshot(id RuntimeID) (Snapshot, error) {
    // ...

    // Include all region states
    regionStatesCopy := make(map[string]string)
    for k, v := range runtime.regionStates {
        regionStatesCopy[k] = v
    }

    return Snapshot{
        // ... existing fields
        RegionStates:     regionStatesCopy,
        CompletedRegions: runtime.completedRegions,
    }, nil
}
```

**Note**: Restoring parallel states requires recreating region goroutines in their saved states.

---

## Test Plan

### Unit Tests
1. **Parallel entry** - All regions enter initial states simultaneously
2. **TargetPath routing** - Events route to specific regions
3. **Final state detection** - Regions emit done when entering final state
4. **Completion barrier** - Parent waits for all regions before exiting
5. **Force-stop** - Control(Stop) terminates regions immediately
6. **Exit actions** - All region exit actions fire (order not guaranteed)
7. **Cross-region events** - RegionA can dispatch to regionB via TargetPath
8. **Snapshot consistency** - All region states captured atomically

### Integration Tests
1. **YAML-defined parallel states** - Full round-trip from definition to execution
2. **Hot-reload with parallel states** - Quiescence detection with regions
3. **Transient sub-chart lifecycle** - Spawn/join/abort patterns

---

## Files Modified

| File                                | Changes                                                 |
| ----------------------------------- | ------------------------------------------------------- |
| `pkg/statechart/types.go`           | Add IsFinal to Node, EventType constants                |
| `pkg/statechart/runtime.go`         | Add RegionRuntime, extend ChartRuntime                  |
| `pkg/statechart/engine.go`          | Parallel entry/exit, event routing, completion tracking |
| `pkg/statechart/statechart_test.go` | Add parallel state tests                                |

---

## Deferred to v3

- Sub-chart routing (`child:bar` TargetPath)
- Nested parallel states (parallel within parallel)
- Dynamic reclassification (atomic ↔ parallel at runtime)
- ReplaceDefinition for parallel states

---

## Open Questions

1. **"Self" routing**: Broadcast to all regions, or only caller's region?
2. **Event evaluation**: Against all regions simultaneously, or only targeted?
3. **Exit action errors**: Continue with other regions, or halt exit?
4. **Region count changes**: Handle if chart is hot-reloaded with different region count?
