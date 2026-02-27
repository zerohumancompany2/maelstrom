# Parallel States Implementation Plan v5

**Date:** 2026-02-27
**Status:** Approved by CTO
**Theme:** Recursive Runtime Pattern with Unified Event Transport

---

## 1. Core Concept: Recursive Runtime Pattern

This plan unifies the architecture by treating parallel regions as **lightweight instances of the same state machine core** used by top-level runtimes. All communication uses a single `Event` type with `sys:` prefixes for runtime coordination.

```
ChartRuntime (parent, parallel state)
├── EventRouter (coordinates all regions)
│   ├── External events → routed to specific regions
│   ├── System events (sys:done, sys:transition) → coordinated
│   └── Exit coordination → waits for all regions "sys:done"
│
├── RegionRuntime A (goroutine)
│   ├── StateMachine Core (reused from engine)
│   ├── Event Queue (chan Event from parent)
│   └── Output Channel (chan Event to parent)
│
├── RegionRuntime B (goroutine)
│   └── ... same pattern ...
│
└── RegionRuntime N (goroutine)
    └── ... same pattern ...
```

### 1.1 Key Insight: Unified Event Transport

**All channels use `chan Event`.** Runtime coordination uses `sys:` prefixes (matching arch-v1 convention):

| Event Type | Examples | Direction |
|------------|----------|-----------|
| User events | `user_query`, `startWorker`, `nextA` | External → Regions |
| System events | `sys:enter`, `sys:done`, `sys:transition` | Regions → Parent |
| Control events | `sys:pause`, `sys:resume`, `sys:exit` | Parent → Regions |

**Two trivial helpers** (single source of truth):

```go
func (e Event) IsSystem() bool { return strings.HasPrefix(e.Type, "sys:") }
func (e Event) IsUser() bool   { return !e.IsSystem() }
```

---

## 2. System Event Constants

```go
// System event types for parallel region coordination
const (
    // Parent → Region: Lifecycle control
    SysEnter  = "sys:enter"  // Region should enter initial state
    SysExit   = "sys:exit"   // Region should process toward final state
    SysPause  = "sys:pause"  // Region should pause processing
    SysResume = "sys:resume" // Region should resume processing

    // Region → Parent: Coordination signals
    SysDone       = "sys:done"       // Region reached final state
    SysTransition = "sys:transition" // Region changed state
    SysPanic      = "sys:panic"      // Region goroutine panicked
    SysBackpressure = "sys:backpressure" // Region queue full
)
```

---

## 3. Architecture Components

### 3.1 State Machine Core (Extracted)

The reusable state machine engine, decoupled from runtime management:

```go
// StateMachine executes statechart logic for a single region.
type StateMachine struct {
    definition   ChartDefinition
    activeState  string
    actions      map[string]ActionFn
    guards       map[string]GuardFn
    appCtx       ApplicationContext
}

// ProcessEvent evaluates a single event, returns result for parent coordination.
func (sm *StateMachine) ProcessEvent(ev Event) EventResult {
    // 1. Find current node
    // 2. Evaluate transitions (with guards)
    // 3. Execute exit actions
    // 4. Execute transition actions
    // 5. Update active state
    // 6. Execute entry actions
    // 7. Return result
}
```

### 3.2 RegionRuntime (Unified Event Channels)

A parallel region is a state machine core wrapped in a goroutine with **symmetric `chan Event` I/O**:

```go
type RegionRuntime struct {
    name         string
    stateMachine *StateMachine

    // Unified event channels (both chan Event)
    inputChan  chan Event  // Receives from parent router
    outputChan chan Event  // Sends to parent router

    state      RegionState
    mu         sync.Mutex
}

func (rr *RegionRuntime) Run() {
    for {
        select {
        case ev := <-rr.inputChan:
            rr.handleEvent(ev)

        case <-rr.stopSignal:
            rr.gracefulShutdown()
            return
        }
    }
}

func (rr *RegionRuntime) handleEvent(ev Event) {
    // Handle system events directly
    if ev.IsSystem() {
        rr.handleSystemEvent(ev)
        return
    }

    // Process user event through state machine
    result := rr.stateMachine.ProcessEvent(ev)

    // Report transition to parent via system event
    if result.Transitioned {
        rr.outputChan <- Event{
            Type:   SysTransition,
            Source: "region:" + rr.name,
            Payload: TransitionPayload{
                From: result.FromState,
                To:   result.ToState,
            },
        }
    }

    // Report completion if final state reached
    if result.IsFinalState {
        rr.state = RegionStateDone
        rr.outputChan <- Event{
            Type:   SysDone,
            Source: "region:" + rr.name,
        }
    }

    // Route emitted user event through parent
    if result.EmitEvent != nil {
        rr.outputChan <- Event{
            Type:       result.EmitEvent.Type,
            Payload:    result.EmitEvent.Payload,
            Source:     "region:" + rr.name,
            TargetPath: result.EmitEvent.TargetPath,
        }
    }
}

func (rr *RegionRuntime) handleSystemEvent(ev Event) {
    switch ev.Type {
    case SysEnter:
        // Execute entry actions for initial state
        rr.stateMachine.executeEntryActions(rr.stateMachine.activeState, ev)

    case SysExit:
        // Signal to process toward final state
        rr.state = RegionStateExiting

    case SysPause:
        rr.state = RegionStatePaused

    case SysResume:
        rr.state = RegionStateRunning
    }
}
```

### 3.3 EventResult (Unchanged)

```go
type EventResult struct {
    Transitioned bool
    FromState    string
    ToState      string
    IsFinalState bool
    EmitEvent    *Event  // User event to route
    Error        error
}
```

### 3.4 EventRouter (Handles System Events)

```go
type EventRouter struct {
    parentID     string
    regions      map[string]*RegionRuntime
    inputChan    chan Event  // External events + region output
    regionOutput chan Event  // Alias for same channel

    completion   map[string]bool
    activeStates map[string]string
    state        RouterState
    mu           sync.RWMutex
}

func (er *EventRouter) Run() {
    for {
        select {
        case ev := <-er.inputChan:
            er.handleEvent(ev)

        case <-er.pauseSignal:
            er.broadcastToRegions(Event{Type: SysPause})

        case <-er.resumeSignal:
            er.broadcastToRegions(Event{Type: SysResume})

        case <-er.exitSignal:
            er.initiateGracefulExit()
        }
    }
}

func (er *EventRouter) handleEvent(ev Event) {
    // System events from regions
    if ev.IsSystem() && strings.HasPrefix(ev.Source, "region:") {
        er.handleRegionSystemEvent(ev)
        return
    }

    // User events: route based on TargetPath
    er.routeUserEvent(ev)
}

func (er *EventRouter) handleRegionSystemEvent(ev Event) {
    regionName := strings.TrimPrefix(ev.Source, "region:")

    switch ev.Type {
    case SysDone:
        er.completion[regionName] = true
        if er.allRegionsComplete() {
            er.signalParentCompletion()
        }

    case SysTransition:
        payload := ev.Payload.(TransitionPayload)
        er.activeStates[regionName] = payload.To
        er.updateParentActiveStates()

    case SysPanic:
        // Handle region panic - log and potentially fail fast
        er.handleRegionPanic(regionName, ev.Payload)

    case SysBackpressure:
        // Handle queue full scenario
        er.handleBackpressure(regionName)
    }
}

func (er *EventRouter) routeUserEvent(ev Event) {
    target := parseTargetPath(ev.TargetPath)

    switch {
    case target.IsSelf || target.TargetPath == "":
        // Broadcast to all regions
        er.broadcastToRegions(ev)

    case target.Region != "":
        // Target specific region
        if region, ok := er.regions[target.Region]; ok {
            region.inputChan <- ev
        }

    case target.Child != "":
        // Future: nested parallel states
    }
}

func (er *EventRouter) broadcastToRegions(ev Event) {
    for _, region := range er.regions {
        select {
        case region.inputChan <- ev:
        default:
            // Queue full - emit backpressure
            region.outputChan <- Event{
                Type:   SysBackpressure,
                Source: "router",
            }
        }
    }
}

func (er *EventRouter) initiateGracefulExit() {
    er.state = RouterStateExiting
    er.broadcastToRegions(Event{Type: SysExit})
}
```

### 3.5 TargetPath Parsing (Unchanged)

```go
type TargetPath struct {
    TargetPath string
    IsSelf     bool
    Region     string
    Child      string
}

func parseTargetPath(path string) TargetPath {
    if path == "" || path == "." {
        return TargetPath{IsSelf: true}
    }

    if strings.HasPrefix(path, "region:") {
        return TargetPath{Region: strings.TrimPrefix(path, "region:")}
    }

    if strings.HasPrefix(path, "child:") {
        return TargetPath{Child: strings.TrimPrefix(path, "child:")}
    }

    return TargetPath{TargetPath: path}
}
```

---

## 4. Integration with ChartRuntime

### 4.1 Modified ChartRuntime

```go
type ChartRuntime struct {
    id           RuntimeID
    definition   ChartDefinition
    state        RuntimeState
    activeState  string
    appCtx       ApplicationContext
    runtimeCtx   RuntimeContext
    mu           sync.RWMutex

    // Parallel state support
    router       *EventRouter
    isParallel   bool

    // Event buffer for non-parallel states
    eventQueue   []Event
    actions      map[string]ActionFn
    guards       map[string]GuardFn
}

func (e *Engine) processEvent(runtime *ChartRuntime, ev Event) error {
    currentNode := e.findNode(runtime.definition.Root, runtime.activeState)

    if currentNode.NodeType() == NodeTypeParallel {
        if runtime.router != nil {
            runtime.router.inputChan <- ev
        }
        return nil
    }

    return e.processAtomicOrCompound(runtime, currentNode, ev)
}
```

### 4.2 Parallel State Entry

```go
func (e *Engine) enterParallelState(runtime *ChartRuntime, node *Node) error {
    runtime.isParallel = true

    // Single channel for all events (external + region output)
    eventChan := make(chan Event, 100)

    router := &EventRouter{
        parentID:     string(runtime.id),
        regions:      make(map[string]*RegionRuntime),
        inputChan:    eventChan,
        regionOutput: eventChan, // Same channel
        completion:   make(map[string]bool),
        activeStates: make(map[string]string),
        state:        RouterStateActive,
    }
    runtime.router = router

    // Create and start regions
    for regionName, childNode := range node.Children {
        region := &RegionRuntime{
            name:       regionName,
            inputChan:  make(chan Event, 100),
            outputChan: eventChan, // Routes back to router
            stateMachine: &StateMachine{
                definition: ChartDefinition{
                    Root:         childNode,
                    InitialState: findInitialState(childNode),
                },
                actions: runtime.actions,
                guards:  runtime.guards,
                appCtx:  runtime.appCtx,
            },
            state: RegionStateRunning,
        }

        router.regions[regionName] = region
        router.completion[regionName] = false

        go region.Run()

        // Send enter event
        region.inputChan <- Event{Type: SysEnter}
    }

    go router.Run()
    runtime.updateParallelActiveStates()

    return nil
}
```

### 4.3 Parallel State Exit

```go
func (e *Engine) exitParallelState(runtime *ChartRuntime, ev Event) error {
    router := runtime.router
    if router == nil {
        return nil
    }

    router.exitSignal <- true

    select {
    case <-router.completionSignal:
        // All regions done
    case <-time.After(router.exitTimeout):
        e.forceStopParallelState(runtime)
    }

    return e.executeExitActions(runtime, runtime.activeState, ev)
}
```

---

## 5. Communication Flow Examples

### 5.1 External Event to Specific Region

```
External Dispatch
    ↓
Engine.Dispatch() → router.inputChan
    ↓
EventRouter.handleEvent() → routeUserEvent()
    ↓
Parse TargetPath="region:worker"
    ↓
worker.inputChan <- ev
    ↓
RegionRuntime.handleEvent() → StateMachine.ProcessEvent()
    ↓
Transition fires → outputChan <- Event{Type: SysTransition}
    ↓
EventRouter.handleEvent() → handleRegionSystemEvent()
    ↓
Update activeStates
```

### 5.2 Cross-Region Event

```
coordinator region emits Event{Type: "startWorker", TargetPath: "region:worker"}
    ↓
outputChan → router.inputChan
    ↓
EventRouter sees user event with TargetPath
    ↓
Route to worker.inputChan
    ↓
worker processes, transitions
```

### 5.3 Graceful Exit

```
Parent signals exit
    ↓
router.exitSignal
    ↓
EventRouter.initiateGracefulExit()
    ↓
broadcastToRegions(Event{Type: SysExit})
    ↓
Regions process toward final states
    ↓
Region A: SysDone → router.inputChan
    ↓
Region B: SysDone → router.inputChan
    ↓
All regions complete → signalParentCompletion()
```

---

## 6. Quiescence Detection

```go
type EventRouter struct {
    // ... existing fields ...

    pendingEvents    int32
    inFlightMessages int32
}

func (er *EventRouter) IsQuiescent() bool {
    return atomic.LoadInt32(&er.pendingEvents) == 0 &&
           atomic.LoadInt32(&er.inFlightMessages) == 0 &&
           len(er.inputChan) == 0
}
```

---

## 7. Snapshot with Parallel States

```go
func (e *Engine) Snapshot(id RuntimeID) (Snapshot, error) {
    runtime := e.getRuntime(id)
    runtime.mu.RLock()
    defer runtime.mu.RUnlock()

    snap := Snapshot{
        RuntimeID:    runtime.id,
        ActiveStates: runtime.getActiveStates(),
    }

    if runtime.isParallel && runtime.router != nil {
        snap.RegionStates = make(map[string]RegionSnapshot)
        for name, region := range runtime.router.regions {
            snap.RegionStates[name] = RegionSnapshot{
                ActiveState: region.stateMachine.activeState,
                IsDone:      region.state == RegionStateDone,
            }
        }
    }

    return snap, nil
}
```

---

## 8. Implementation Phases

| Phase | Task | Tests |
|-------|------|-------|
| 1 | Extract `StateMachine` core | State machine unit tests |
| 2 | Implement `RegionRuntime` with unified channels | Region lifecycle, sys events |
| 3 | Implement `EventRouter` with system event handling | Routing, cross-region events |
| 4 | Integrate with `ChartRuntime` | Full parallel lifecycle |
| 5 | Quiescence & control commands | Pause/resume, snapshot |
| 6 | Edge cases | Panic recovery, backpressure |

---

## 9. Changes from v4

| Aspect | v4 | v5 (This Document) |
|--------|-----|-------------------|
| Region input | `chan Event` | `chan Event` (unchanged) |
| Region output | `chan RegionMessage` | `chan Event` (unified) |
| System signals | `RegionMessage{Type: "done"}` | `Event{Type: SysDone}` |
| EventRouter input | `chan Event` + `chan RegionMessage` | Single `chan Event` |
| Helpers | None | `IsSystem()`, `IsUser()` |
| Constants | Ad-hoc strings | `SysEnter`, `SysDone`, etc. |

---

## 10. Benefits of v5 Approach

1. **Single event type everywhere** - No type switching, no wrapper structs
2. **Symmetric channels** - Input and output both `chan Event`
3. **Aligned with arch-v1** - Uses established `sys:` convention
4. **Minimal boilerplate** - Just two helper methods
5. **Extensible** - New system events = one constant + one case

---

*Approved by CTO. Ready for Phase 1 implementation.*
