# Parallel States Implementation Plan v3

Hierarchical event routing model. Events flow up to parent or down to children. No sideways region-region communication.

---

## Core Architecture

### Event Flow Directions

| Direction | Path             | Use Case                                               |
| --------- | ---------------- | ------------------------------------------------------ |
| **Up**    | Child → Parent   | Completion, errors, state changes parent must know     |
| **Down**  | Parent → Child   | External events, cross-region signals, parent commands |
| **Self**  | Within same node | Local transition evaluation                            |

### Invariants

1. **No region-region channels** — All cross-region communication routes through parent Event Router
2. **Parent mediates coordination** — If Region A needs Region B to act, A signals parent; parent dispatches to B
3. **Consistent event handling** — Atomic, compound, parallel nodes use same processing logic; only routing differs
4. **Hierarchical addressing** — `region:foo` resolves through parent, not direct connection

---

## Event Router

Central component in parallel state parent that handles all event distribution:

```go
// EventRouter manages event distribution for a parallel state
type EventRouter struct {
    parentID     string           // the parallel state's runtime ID
    regions      map[string]*RegionRuntime
    completion   map[string]bool  // tracks done status
    parentChan   chan RegionMessage  // receives from regions (up)
}

type RouterCommand struct {
    Type      string      // "dispatch", "regionDone", "regionTransition"
    Target    string      // region name for "dispatch"
    Event     Event       // event to deliver
    Region    string      // source region for "regionDone", "regionTransition"
    NewState  string      // new state path for "regionTransition"
}

func (er *EventRouter) Run(incomingEvents chan Event) {
    for {
        select {
        case ev := <-incomingEvents:
            er.routeEvent(ev)

        case msg := <-er.parentChan:
            er.handleRegionMessage(msg)
        }
    }
}

func (er *EventRouter) routeEvent(ev Event) {
    target, _ := parseTargetPath(ev.TargetPath)

    switch {
    case target.IsSelf:
        // "self" with no region context = broadcast to all regions
        // "self" with region context = route to that specific region
        if ev.SourceRegion != "" {
            er.dispatchToRegion(ev.SourceRegion, ev)
        } else {
            er.broadcastToAll(ev)
        }

    case target.Region != "":
        // Explicit region targeting
        er.dispatchToRegion(target.Region, ev)

    case target.Child != "":
        // Future: sub-chart routing (deferred to v3+)
        er.dispatchToChild(target.Child, ev)
    }
}

func (er *EventRouter) handleRegionMessage(msg RegionMessage) {
    switch msg.Type {
    case "done":
        er.completion[msg.Region] = true
        if er.allRegionsComplete() {
            er.signalParentCompletion()
        }

    case "transition":
        // Track state change for parent-level transitions
        er.parentStateUpdate(msg.Region, msg.NewState)

    case "event":
        // Region wants to emit event (possibly to another region)
        // Parent router handles redirection
        er.routeEvent(msg.Event)
    }
}
```

---

## Region Runtime (Revised)

Regions receive events from parent, send messages up to parent. No awareness of other regions.

```go
type RegionRuntime struct {
    name         string
    currentState string
    eventQueue   chan Event       // receives from parent router (down)
    parentChan   chan RegionMessage  // sends to parent router (up)

    // Same eval engine as atomic/compound states
    actions      map[string]ActionFn
    guards       map[string]GuardFn
}

func (rr *RegionRuntime) run() {
    for ev := range rr.eventQueue {
        result := rr.processEvent(ev)

        if result.Transitioned && result.IsFinalState {
            rr.parentChan <- RegionMessage{
                Type:   "done",
                Region: rr.name,
            }
        }

        if result.EmitEvent != nil {
            // Event needs routing through parent
            rr.parentChan <- RegionMessage{
                Type:    "event",
                Region:  rr.name,
                Event:   *result.EmitEvent,
            }
        }
    }
}

func (rr *RegionRuntime) processEvent(ev Event) EventResult {
    // Identical to atomic/compound state processing
    // Evaluate guards, execute actions, determine transition
}
```

---

## Event Addressing Specification

| TargetPath               | Semantics                          | Example Use                   |
| ------------------------ | ---------------------------------- | ----------------------------- |
| `""` or `"."`            | Self (current context)             | Local transition, same region |
| `"region:foo"`           | Specific region                    | Cross-region signal           |
| `"child:bar"`            | Sub-chart (deferred)               | Spawn sub-agent               |
| `"child:bar/region:baz"` | Region within sub-chart (deferred) | Nested parallel               |

### Cross-Region Communication Pattern

```yaml
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
                  targetPath: "region:worker"  # explicit routing
            transitions:
              - event: "workerDone"
                to: "complete"
          complete:
            type: final

      - name: worker
        child:
          idle:
            type: atomic
            transitions:
              - event: "startWorker"  # receives via parent router
                to: "processing"
          processing:
            type: atomic
            onEnter:
              - action: doWork
            transitions:
              - event: "workComplete"
                action:
                  - emitEvent:
                      type: "workerDone"
                      targetPath: "region:coordinator"  # back to coordinator
                to: "done"
          done:
            type: final
```

---

## Parent-Region Message Protocol

Messages flow up from regions to parent via `parentChan`:

```go
type RegionMessage struct {
    Type     string      // "done", "transition", "event", "error"
    Region   string      // source region name

    // For "transition"
    NewState string

    // For "event" (region emitting event to be routed)
    Event    Event

    // For "error"
    Error    error
}
```

**Message Types:**

| Type         | Sent When                          | Parent Action                                  |
| ------------ | ---------------------------------- | ---------------------------------------------- |
| `done`       | Region enters final state          | Track completion; exit parallel when all done  |
| `transition` | Region changes state               | Update ActiveStates; check parent-level guards |
| `event`      | Region emits event with TargetPath | Route via EventRouter                          |
| `error`      | Region processing error            | Log, propagate, or initiate force-stop         |

---

## Parallel State Lifecycle

### Entry

```go
func (e *Engine) enterParallelState(runtime *ChartRuntime, node *Node) error {
    router := &EventRouter{
        parentID:   runtime.id,
        regions:    make(map[string]*RegionRuntime),
        completion: make(map[string]bool),
        parentChan: make(chan RegionMessage, 100),
    }

    // Create and start all regions
    for regionName, childNode := range node.Children {
        region := &RegionRuntime{
            name:         regionName,
            currentState: findInitialState(childNode),
            eventQueue:   make(chan Event, 100),
            parentChan:   router.parentChan, // same channel for all
            actions:      runtime.actions,
            guards:       runtime.guards,
        }

        router.regions[regionName] = region
        router.completion[regionName] = false

        go region.run()

        // Send initial entry event
        region.eventQueue <- Event{Type: "sys:enter"}
    }

    runtime.router = router
    go router.Run(runtime.externalEvents) // pump events through router

    return nil
}
```

### Exit

```go
func (e *Engine) exitParallelState(runtime *ChartRuntime, ev Event) error {
    router := runtime.router

    // Phase 1: Request graceful stop
    for _, region := range router.regions {
        region.eventQueue <- Event{Type: "sys:exit"}
    }

    // Phase 2: Wait with timeout (configured in ChartDefinition)
    if !router.waitForCompletion(timeoutMs) {
        // Phase 3: Force stop (close queues, abandon goroutines)
        for _, region := range router.regions {
            close(region.eventQueue)
        }
    }

    // Phase 4: Execute exit actions for all regions (concurrent, no order guarantee)
    // ...

    return nil
}
```

---

## ActiveStates Representation

```go
// For parallel states, ActiveStates contains one entry per region
type ChartRuntime struct {
    // ...
    activeStates []StatePath
    router       *EventRouter
}

// StatePath format: "root/parallelState/regionName/currentNode"
// Example parallel state ActiveStates:
// ["root/orient/shortTerm/conversationScan", "root/orient/longTerm/ragQuery"]
```

---

## Differences from v2

| Aspect              | v2                              | v3                                                  |
| ------------------- | ------------------------------- | --------------------------------------------------- |
| Region-region comms | Implied direct (via TargetPath) | Explicitly through parent EventRouter               |
| "self" routing      | SourceRegion context            | Broadcast if no context, specific region if context |
| Cross-region events | Dispatched directly             | Routed through parent                               |
| Event flow model    | Mixed                           | Strictly hierarchical (up/down)                     |
| Parent awareness    | Passive                         | Active EventRouter mediates all                     |

---

## Clarified Semantics (from discussion)

### Parent Transitions on Parallel States

**Rule:** Parent transition signals *intent* to exit, but parallel state **cannot complete exit until all regions emit `done`**.

**Flow:**
1. Event arrives matching parent transition
2. Parent marks parallel state as "exiting"
3. Exit signal sent to all regions (they continue processing toward final states)
4. Regions emit `done` as they reach final states
5. When **all regions done**: Execute parallel state's exit actions, complete transition

**Region handling during exit:** If a region also has a transition on the same event, both fire—the region transitions (possibly to final state), parent initiates exit sequence.

### Exit Synchronization (Quiescence-Based)

**Rule:** Exit is **pause + quiescence + cleanup**.

**Sequence:**
```
Parent exit initiated (event matches parent transition)
    ↓
Pause event dispatch (quiescence phase 1)
    ↓
Wait for regions to finish current events
    ↓
Quiescence achieved (all regions stable)
    ↓
Signal regions to process toward final states
    ↓
Region A emits "done" ──────┐
Region B emits "done" ──────┼── Wait for all
Region C emits "done" ──────┘
    ↓
All regions done
    ↓
Execute parallel state exit actions (deterministic, after quiescence)
    ↓
Transition to target state
```

**Key point:** Exit actions run **after quiescence**, not during active processing. This ensures deterministic cleanup.

**Force-stop:** Bypasses quiescence and `done` waiting—immediate termination for emergency situations.

### External Event Routing

**Rule:** External events route identically to internal events—no special outside addressing.

- TargetPath `"."` or omitted = broadcast to all regions
- TargetPath `"region:foo"` = routed to specific region

---

## Control Commands (Pause/Resume/Stop/Snapshot)

### Pause → Quiescence

**Semantics:** `Control(id, "pause")` halts event dispatch at EventRouter, allowing regions to finish current events. System reaches **quiescence** (arch-v1 Section 12.3).

**Quiescence criteria:**
- EventRouter queue empty (no pending dispatches)
- All regions finished processing (not mid-event)
- No pending cross-region signals in flight

**In quiescence:**
- Parent transitions **also paused** (whole parallel state "stilled")
- State is stable, inspectable, serializable
- Can transition to resume, stop, or exit-with-cleanup

### Resume

**Semantics:** Resume event processing from quiescent state.

**Flow:**
1. Regions resume from current states
2. EventRouter resumes dispatch
3. Queued events (if any) processed

**Note:** If a region reached final state during pause (possible if entry action caused it), `done` was already emitted. Quiescence means all activity completed.

### Graceful Exit (via Parent Transition)

**Rule:** Exit is **pause + quiescence + cleanup + stop**.

**Flow:**
```
Event triggers parent transition
    ↓
Pause event dispatch
    ↓
Wait for quiescence (regions finish current events)
    ↓
All regions in stable states
    ↓
Execute exit actions (deterministic, no racing)
    ↓
Close regions, complete transition
```

**Key point:** Parallel state cannot complete exit until all regions emit `done` (reach final state). Exit actions run after quiescence achieved.

### Hard Stop

**Semantics:** `Control(id, "stop")` bypasses quiescence—immediate termination.

**Rule:** Stop is **not** the same as parent transition exit—it terminates immediately without waiting for `done`.

**Flow:**
1. Close all region queues immediately (regions may be mid-event)
2. Abandon pending events
3. Exit actions: **attempt with short timeout**, force-abandon if they hang
4. Cleanup and terminate

**Use case:** Emergency shutdown, resource exhaustion, unrecoverable error.

### Snapshot

**Semantics:** `Control(id, "snapshot")` is **pause + quiescence + serialize + optional resume**.

**Flow:**
```
Control(id, "snapshot", resume=true|false)
    ↓
Pause event dispatch
    ↓
Wait for quiescence
    ↓
Capture: region states, completion map, EventRouter queue state
    ↓
Resume (if flag set) or remain paused
```

**Note:** Snapshot captures complete state including in-flight events in queue.

---

## Control Command Summary

| Command | Quiescence Required? | Exit Actions? | Waits for `done`? | Use Case |
|---------|---------------------|---------------|-------------------|----------|
| `pause` | Yes (enter) | No | N/A | Inspection, prepare for operation |
| `resume` | N/A (exit) | No | N/A | Resume from paused state |
| Graceful exit (transition) | Yes | Yes (after quiescence) | Yes | Normal lifecycle completion |
| `stop` (hard) | No | Best-effort with timeout | No | Emergency, unrecoverable |
| `snapshot` | Yes | No | N/A | Persistence, migration, debugging |

---

*End of v3 Hierarchical Event Model*