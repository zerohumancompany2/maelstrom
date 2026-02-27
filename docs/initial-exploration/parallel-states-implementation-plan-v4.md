# Parallel States Implementation Plan v4

**Date:** 2026-02-27
**Status:** Draft
**Theme:** Recursive Runtime Pattern - "A parallel region is just a runtime wrapped in a goroutine"

---

## 1. Core Concept: Recursive Runtime Pattern

This plan unifies the architecture by treating parallel regions as **lightweight instances of the same state machine core** used by top-level runtimes. The pattern is recursive: a parallel state's parent spawns N region runtimes, each with its own goroutine, event queue, and lifecycle.

```
ChartRuntime (parent, parallel state)
├── EventRouter (coordinates all regions)
│   ├── External events → routed to specific regions
│   ├── Region messages → coordinated
│   └── Exit coordination → waits for all regions "done"
│
├── RegionRuntime A (goroutine)
│   ├── StateMachine Core (reused from engine)
│   ├── Event Queue (chan from parent)
│   └── Parent Channel (signals completion, emits events)
│
├── RegionRuntime B (goroutine)
│   └── ... same pattern ...
│
└── RegionRuntime N (goroutine)
    └── ... same pattern ...
```

### 1.1 Key Insight

The state machine logic (event evaluation, transition execution, entry/exit actions) is **identical** for top-level runtimes and parallel regions. Only the **communication plumbing** differs:

| Aspect            | Top-Level Runtime          | Parallel Region                |
| ----------------- | -------------------------- | ------------------------------ |
| Event source      | External `Dispatch()` call | Parent EventRouter channel     |
| Completion signal | `stopRuntime()` called     | `done` message to parent       |
| State reporting   | `Snapshot()` API           | `transition` message to parent |
| Event emission    | Direct to queue            | Via parent router              |

---

## 2. Architecture Components

### 2.1 State Machine Core (Extracted)

The reusable state machine engine, decoupled from runtime management:

```go
// StateMachine executes statechart logic for a single region.
// Used by both top-level ChartRuntime and parallel RegionRuntime.
type StateMachine struct {
    definition   ChartDefinition    // Region's subtree
    activeState  string             // Current state path
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
    // 7. Return result (transitioned?, isFinal?, emitEvent?)
}

// Entry/Exit action execution (same as current engine)
func (sm *StateMachine) executeEntryActions(statePath string, ev Event) error
func (sm *StateMachine) executeExitActions(statePath string, ev Event) error
```

**Reused from current:** `engine.go` logic lines 250-377.

### 2.2 RegionRuntime (Lightweight Wrapper)

A parallel region is a state machine core wrapped in a goroutine with channel I/O:

```go
// RegionRuntime represents one parallel region executing in isolation.
type RegionRuntime struct {
    name       string              // Region identifier (e.g., "regionA")
    stateMachine *StateMachine     // Core state machine (reused)

    // Communication channels (parent-mediated)
    eventQueue chan Event         // Receives from parent router
    parentChan chan RegionMessage // Sends to parent router

    // Lifecycle
    state      RegionState        // running, paused, stopping, done
    mu         sync.Mutex
}

type RegionState int
const (
    RegionStateRunning RegionState = iota
    RegionStatePaused
    RegionStateStopping
    RegionStateDone
)

// Run executes the region's event loop.
// This runs in its own goroutine.
func (rr *RegionRuntime) Run() {
    for {
        select {
        case ev := <-rr.eventQueue:
            if rr.state == RegionStatePaused {
                // Queue or drop based on policy
                continue
            }
            rr.handleEvent(ev)

        case <-rr.stopSignal:
            rr.gracefulShutdown()
            return
        }
    }
}

func (rr *RegionRuntime) handleEvent(ev Event) {
    result := rr.stateMachine.ProcessEvent(ev)

    // Report transition to parent
    if result.Transitioned {
        rr.parentChan <- RegionMessage{
            Type:     "transition",
            Region:   rr.name,
            NewState: rr.stateMachine.activeState,
        }
    }

    // Report completion if final state reached
    if result.IsFinalState {
        rr.state = RegionStateDone
        rr.parentChan <- RegionMessage{
            Type:   "done",
            Region: rr.name,
        }
    }

    // Route emitted event through parent
    if result.EmitEvent != nil {
        rr.parentChan <- RegionMessage{
            Type:    "event",
            Region:  rr.name,
            Event:   *result.EmitEvent,
        }
    }
}
```

### 2.3 EventResult (Communication Contract)

The state machine core returns structured results for parent coordination:

```go
type EventResult struct {
    Transitioned bool        // Did a transition fire?
    FromState    string      // Previous state
    ToState      string      // New state
    IsFinalState bool        // Is new state a final state?
    EmitEvent    *Event      // Event to route (cross-region or external)
    Error        error       // Processing error (if any)
}
```

### 2.4 EventRouter (Unified Coordination)

The EventRouter mediates all communication within a parallel state. It replaces the direct dispatch mechanism for parallel regions.

```go
// EventRouter manages event distribution for a parallel state.
// Runs in its own goroutine, owned by the parent ChartRuntime.
type EventRouter struct {
    // Configuration
    parentID    string                      // Owning runtime ID
    regions     map[string]*RegionRuntime   // Active regions

    // Communication
    incoming    chan Event                  // External events
    regionChan  chan RegionMessage          // From child regions

    // State tracking
    completion  map[string]bool             // Region -> done?
    activeStates map[string]string          // Region -> current state path

    // Lifecycle
    state       RouterState                 // active, paused, exiting
    mu          sync.RWMutex
}

type RouterState int
const (
    RouterStateActive RouterState = iota
    RouterStatePaused
    RouterStateExiting
    RouterStateStopped
)

// Run is the EventRouter's main loop.
func (er *EventRouter) Run() {
    for {
        select {
        case ev := <-er.incoming:
            er.routeEvent(ev)

        case msg := <-er.regionChan:
            er.handleRegionMessage(msg)

        case <-er.pauseSignal:
            er.state = RouterStatePaused
            er.broadcastToRegions(Event{Type: "sys:pause"})

        case <-er.resumeSignal:
            er.state = RouterStateActive
            er.broadcastToRegions(Event{Type: "sys:resume"})

        case <-er.exitSignal:
            er.initiateGracefulExit()
        }
    }
}

// routeEvent distributes events based on TargetPath.
func (er *EventRouter) routeEvent(ev Event) {
    target := parseTargetPath(ev.TargetPath)

    switch {
    case target.IsSelf || target.TargetPath == "":
        // Broadcast to all regions (external event)
        er.broadcastToRegions(ev)

    case target.Region != "":
        // Target specific region
        if region, ok := er.regions[target.Region]; ok {
            region.eventQueue <- ev
        } else {
            // Log error: region not found
        }

    case target.Child != "":
        // Future: nested parallel states (v5+)
        // Route to region which contains the child
    }
}

// handleRegionMessage processes messages from child regions.
func (er *EventRouter) handleRegionMessage(msg RegionMessage) {
    switch msg.Type {
    case "done":
        er.completion[msg.Region] = true
        if er.allRegionsComplete() {
            er.signalParentCompletion()
        }

    case "transition":
        er.activeStates[msg.Region] = msg.NewState
        // Update parent's view of ActiveStates
        er.updateParentActiveStates()

    case "event":
        // Region wants to emit event (possibly to another region)
        // Re-route through this same router
        er.routeEvent(msg.Event)
    }
}

func (er *EventRouter) broadcastToRegions(ev Event) {
    for _, region := range er.regions {
        // Non-blocking send with timeout?
        select {
        case region.eventQueue <- ev:
        default:
            // Queue full - backpressure handling
        }
    }
}

func (er *EventRouter) allRegionsComplete() bool {
    for _, done := range er.completion {
        if !done {
            return false
        }
    }
    return true
}
```

### 2.5 TargetPath Parsing

```go
type TargetPath struct {
    TargetPath string // Raw path
    IsSelf     bool   // "." or ""
    Region     string // "region:foo" → "foo"
    Child      string // "child:bar" → "bar" (future: sub-charts)
}

func parseTargetPath(path string) TargetPath {
    if path == "" || path == "." {
        return TargetPath{IsSelf: true}
    }

    // Parse "region:name" format
    if strings.HasPrefix(path, "region:") {
        return TargetPath{Region: strings.TrimPrefix(path, "region:")}
    }

    // Parse "child:name" format (deferred to sub-chart support)
    if strings.HasPrefix(path, "child:") {
        return TargetPath{Child: strings.TrimPrefix(path, "child:")}
    }

    return TargetPath{TargetPath: path}
}
```

---

## 3. Integration with ChartRuntime

### 3.1 Modified ChartRuntime for Parallel Support

```go
type ChartRuntime struct {
    // Existing fields (atomic/compound states)
    id           RuntimeID
    definition   ChartDefinition
    state        RuntimeState
    activeState  string              // For atomic/compound states
    appCtx       ApplicationContext
    runtimeCtx   RuntimeContext
    mu           sync.RWMutex

    // Parallel state support (new)
    router       *EventRouter        // Non-nil if in parallel state
    isParallel   bool                // Is this runtime a parallel state parent?

    // For top-level runtimes only
    eventQueue   []Event             // External event buffer
    actions      map[string]ActionFn
    guards       map[string]GuardFn
}

// processEvent dispatches to appropriate handler based on current state type.
func (e *Engine) processEvent(runtime *ChartRuntime, ev Event) error {
    currentNode := e.findNode(runtime.definition.Root, runtime.activeState)

    if currentNode.NodeType() == NodeTypeParallel {
        // Parallel state: route through EventRouter
        if runtime.router != nil {
            runtime.router.incoming <- ev
        }
        return nil
    }

    // Atomic or compound: direct processing (current behavior)
    return e.processAtomicOrCompound(runtime, currentNode, ev)
}
```

### 3.2 Parallel State Entry

```go
func (e *Engine) enterParallelState(runtime *ChartRuntime, node *Node) error {
    runtime.isParallel = true
    runtime.state = RuntimeStateRunning

    // Create EventRouter
    router := &EventRouter{
        parentID:     string(runtime.id),
        regions:      make(map[string]*RegionRuntime),
        incoming:     make(chan Event, 100),
        regionChan:   make(chan RegionMessage, 100),
        completion:   make(map[string]bool),
        activeStates: make(map[string]string),
        state:        RouterStateActive,
    }
    runtime.router = router

    // Create and start regions
    for regionName, childNode := range node.Children {
        region := &RegionRuntime{
            name:       regionName,
            eventQueue: make(chan Event, 100),
            parentChan: router.regionChan,
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

        // Start region goroutine
        go region.Run()

        // Send initial entry event
        region.eventQueue <- Event{Type: "sys:enter"}
    }

    // Start router goroutine
    go router.Run()

    // Update ActiveStates
    runtime.updateParallelActiveStates()

    return nil
}
```

### 3.3 Parallel State Exit

```go
func (e *Engine) exitParallelState(runtime *ChartRuntime, ev Event) error {
    router := runtime.router
    if router == nil {
        return nil
    }

    // Signal router to initiate exit
    router.exitSignal <- true

    // Wait for completion or timeout
    select {
    case <-router.completionSignal:
        // All regions done
    case <-time.After(router.exitTimeout):
        // Force stop
        e.forceStopParallelState(runtime)
    }

    // Execute parallel state's exit actions (after quiescence)
    return e.executeExitActions(runtime, runtime.activeState, ev)
}

func (er *EventRouter) initiateGracefulExit() {
    er.state = RouterStateExiting

    // Signal all regions to process toward final states
    for _, region := range er.regions {
        region.eventQueue <- Event{Type: "sys:exit"}
    }

    // Wait for all "done" messages (handled in handleRegionMessage)
}
```

---

## 4. Communication Flow Examples

### 4.1 External Event to Specific Region

```
External Dispatch
    ↓
Engine.Dispatch() → ChartRuntime.eventQueue
    ↓
processEvent() sees parallel state → router.incoming
    ↓
EventRouter.routeEvent() parses TargetPath="region:worker"
    ↓
worker.eventQueue <- ev
    ↓
RegionRuntime.handleEvent() → StateMachine.ProcessEvent()
    ↓
Transition fires → parentChan <- RegionMessage{Type:"transition"}
    ↓
EventRouter updates activeStates
```

### 4.2 Cross-Region Event (Region A → Region B)

```yaml
# Chart definition excerpt
nodes:
  root:
    type: parallel
    regions:
      - name: coordinator
        child:
          dispatch:
            type: atomic
            onEnter:
              - action: emitEvent
                event:
                  type: "startWorker"
                  targetPath: "region:worker"  # Targeted!
```

```
coordinator RegionRuntime onEnter action
    ↓
Emit Event{Type:"startWorker", TargetPath:"region:worker"}
    ↓
parentChan <- RegionMessage{Type:"event", Event:...}
    ↓
EventRouter.handleRegionMessage() → routeEvent()
    ↓
Parse TargetPath → region="worker"
    ↓
worker.eventQueue <- Event{Type:"startWorker"}
    ↓
worker processes event, transitions
```

### 4.3 Graceful Exit Flow

```
Parent transition triggers exit
    ↓
exitParallelState() → router.exitSignal
    ↓
EventRouter.initiateGracefulExit()
    ↓
Send sys:exit to all regions
    ↓
Regions process toward final states
    ↓
Region A: final state reached → parentChan <- "done"
    ↓
Region B: final state reached → parentChan <- "done"
    ↓
EventRouter: allRegionsComplete() == true
    ↓
signalParentCompletion()
    ↓
Execute parallel state's onExit actions
    ↓
Transition to target state
```

---

## 5. Quiescence Detection

Quiescence is achieved when:

1. EventRouter queue is empty
2. All region event queues are empty
3. No region is mid-event processing
4. No in-flight RegionMessages

```go
type EventRouter struct {
    // ... existing fields ...

    // Quiescence tracking
    pendingEvents    int32  // Atomic counter
    inFlightMessages int32  // Atomic counter
}

func (er *EventRouter) IsQuiescent() bool {
    return atomic.LoadInt32(&er.pendingEvents) == 0 &&
           atomic.LoadInt32(&er.inFlightMessages) == 0 &&
           len(er.incoming) == 0
}

// Incremented when event dispatched to region
func (er *EventRouter) trackPendingEvent() {
    atomic.AddInt32(&er.pendingEvents, 1)
}

// Decremented when region finishes processing
func (er *EventRouter) eventCompleted() {
    atomic.AddInt32(&er.pendingEvents, -1)
}
```

---

## 6. Snapshot with Parallel States

```go
func (e *Engine) Snapshot(id RuntimeID) (Snapshot, error) {
    runtime := e.getRuntime(id)
    runtime.mu.RLock()
    defer runtime.mu.RUnlock()

    snap := Snapshot{
        RuntimeID:    runtime.id,
        ActiveStates: runtime.getActiveStates(), // Includes all regions
    }

    if runtime.isParallel && runtime.router != nil {
        // Capture each region's state
        snap.RegionStates = make(map[string]RegionSnapshot)
        for name, region := range runtime.router.regions {
            snap.RegionStates[name] = RegionSnapshot{
                ActiveState:  region.stateMachine.activeState,
                IsDone:       region.state == RegionStateDone,
                EventQueue:   copyQueue(region.eventQueue),
            }
        }
        // Capture router state
        snap.RouterState = RouterSnapshot{
            Completion:   copyMap(runtime.router.completion),
            QueueState:   copyQueue(runtime.router.incoming),
        }
    }

    return snap, nil
}
```

---

## 7. Implementation Phases

### Phase 1: Extract State Machine Core
- Create `StateMachine` struct from existing engine logic
- Move `ProcessEvent`, `executeEntryActions`, `executeExitActions` to methods
- Ensure no goroutine or channel logic in core (pure state machine)
- **Tests:** State machine unit tests (independent of runtime)

### Phase 2: RegionRuntime
- Implement `RegionRuntime` with channel I/O
- Implement `EventResult` return protocol
- **Tests:** Single region lifecycle, event processing, completion signaling

### Phase 3: EventRouter
- Implement `EventRouter` with routing logic
- Implement `TargetPath` parsing
- Implement broadcast and targeted routing
- **Tests:** Routing scenarios, cross-region events, completion tracking

### Phase 4: ChartRuntime Integration
- Add parallel state detection
- Implement `enterParallelState` with region spawning
- Implement `exitParallelState` with graceful shutdown
- **Tests:** Full parallel lifecycle, entry/exit actions, parent transitions

### Phase 5: Quiescence & Control Commands
- Implement quiescence detection
- Implement pause/resume for parallel states
- Implement snapshot/restore with region state
- **Tests:** Quiescence, pause/resume, snapshot round-trip

### Phase 6: Edge Cases & Optimization
- Force-stop scenarios
- Region panic recovery
- Backpressure handling (full queues)
- **Tests:** Error scenarios, stress tests, property tests

---

## 8. Differences from v3

| Aspect | v3 Plan | v4 Plan (This Document) |
|--------|---------|-------------------------|
| Core pattern | Separate components | **Recursive runtime** - same core reused |
| State machine | Embedded in engine | **Extracted `StateMachine` core** |
| Region communication | Implied direct | **Explicit `EventResult` protocol** |
| EventRouter scope | Complex standalone | **Unified coordination** - simpler, clearer |
| Quiescence | Described but not detailed | **Explicit counters, atomic tracking** |
| Implementation path | Monolithic | **Phased extraction** - lower risk |

---

## 9. Benefits of v4 Approach

1. **Code reuse:** Same state machine core for top-level and regions
2. **Testability:** State machine can be unit tested independently
3. **Clarity:** Clear separation between state logic and coordination
4. **Extensibility:** Sub-charts (future) follow same pattern
5. **Maintainability:** Changes to state machine apply everywhere

---

## 10. Open Questions

1. **Backpressure:** What happens when region event queue is full? Block vs drop vs error?
2. **Region panic:** Should parent restart region or fail entire parallel state?
3. **Nested parallel:** How to route `child:foo/region:bar` when sub-chart is also parallel?
4. **Shared context:** Should regions share `ApplicationContext` or have scoped views?

---

*End of Parallel States Implementation Plan v4*