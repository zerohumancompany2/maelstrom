# Statechart Library

A pure Go statechart library for Maelstrom supporting atomic, compound, and parallel states with lifecycle management, event dispatch, and persistence.

## Overview

This library provides a type-safe, concurrent state machine implementation that handles complex state hierarchies. Charts are defined as immutable YAML specifications and executed as isolated runtimes managed by an engine.

## Key Concepts

- **`Engine`**: Main entry point that manages chart runtimes, action/guard registries
- **`ChartRuntime`**: Live instance of a chart with its own state, event queue, and lifecycle
- **`ChartDefinition`**: Immutable, hydrated chart specification (parsed from YAML)
- **`Node`**: State primitive that can be atomic, compound, or parallel
- **`Event`**: Triggers transitions; can be user events or system events (`sys:*`)
- **`ActionFn` / `GuardFn`**: Callback types for side effects and transition guards

## Quick Start

```go
package main

import (
    "github.com/maelstrom/v3/pkg/statechart"
)

func main() {
    // Create engine
    engine := statechart.NewEngine()

    // Register actions
    engine.RegisterAction("logEntry", func(ctx statechart.RuntimeContext, 
        appCtx statechart.ApplicationContext, ev statechart.Event) error {
        // Side effect logic
        return nil
    })

    // Register guards
    engine.RegisterGuard("isLoggedIn", func(appCtx statechart.ApplicationContext, 
        ev statechart.Event) bool {
        // Guard logic
        return true
    })

    // Define chart
    def := statechart.ChartDefinition{
        ID:       "user-flow",
        Version:  "1.0.0",
        Root: &statechart.Node{
            ID: "root",
            Children: map[string]*statechart.Node{
                "idle": {
                    ID: "idle",
                    Transitions: []statechart.Transition{
                        {Event: "login", Target: "authenticated"},
                    },
                },
                "authenticated": {
                    ID: "authenticated",
                    IsInitial: true,
                },
            },
        },
        InitialState: "root/authenticated",
    }

    // Spawn and control runtime
    rtID, err := engine.Spawn(def, appCtx)
    if err != nil {
        // handle error
    }

    engine.Control(rtID, statechart.CmdStart)
    engine.Dispatch(rtID, statechart.Event{Type: "login"})
}
```

## Core Features

### Lifecycle Management

```go
// Spawn a new runtime
rtID, _ := engine.Spawn(def, appCtx)

// Control lifecycle
engine.Control(rtID, statechart.CmdStart)   // created → running
engine.Control(rtID, statechart.CmdPause)   // running → paused
engine.Control(rtID, statechart.CmdResume)  // paused → running
engine.Control(rtID, statechart.CmdStop)    // any → stopped
```

### Parallel States

When a chart enters a parallel state (node with multiple children), regions are automatically coordinated:

```go
def := statechart.ChartDefinition{
    Root: &statechart.Node{
        ID: "parallel",
        Children: map[string]*statechart.Node{
            "regionA": { /* ... */ },
            "regionB": { /* ... */ },
        },
    },
}

// Enter parallel state - regions run concurrently
engine.Dispatch(rtID, statechart.Event{Type: "enterParallel"})

// Target specific region
engine.Dispatch(rtID, statechart.Event{
    Type: "regionEvent",
    TargetPath: "region:regionA",
})
```

### Snapshot & Restore

```go
// Capture state
snap, _ := engine.Snapshot(rtID)

// Restore later (creates new runtime ID)
newRtID, _ := engine.Restore(snap, def, appCtx)
```

### Hot Reload

```go
// Replace definition without stopping runtime
engine.ReplaceDefinition(rtID, newDef)
```

### Transient Sub-charts

```go
// Child lifecycle bound to parent
childID, _ := engine.SpawnTransient(childDef, appCtx, parentID)
// Stopping parent automatically stops all transient children
```

## State Types

| Type | Children | Behavior |
|------|----------|----------|
| **Atomic** | None | Leaf state; no nested states |
| **Compound** | Single | Auto-enters default child (marked `IsInitial`) |
| **Parallel** | Multiple | All children (regions) run concurrently |

## System Events

The library uses `sys:*` prefixed events for internal coordination:

| Event | Direction | Purpose |
|-------|-----------|---------|
| `sys:enter` | Parent → Region | Initialize region |
| `sys:exit` | Parent → Region | Shut down region |
| `sys:pause` | Parent → Region | Pause region processing |
| `sys:resume` | Parent → Region | Resume region processing |
| `sys:done` | Region → Parent | Region reached final state |
| `sys:transition` | Region → Parent | Report state change |
| `sys:panic` | Region → Parent | Region error |
| `sys:backpressure` | Region → Parent | Region overwhelmed |

## Events

Events are the primary mechanism for driving state transitions in Maelstrom. They flow through the statechart engine synchronously within compound/parallel boundaries, and asynchronously across parallel regions.

### Event Structure

```go
type Event struct {
    Type          string      // Event name (e.g., "login", "sys:enter")
    Payload       any         // Arbitrary data attached to the event
    CorrelationID string      // For tracing/request correlation
    Source        string      // Where the event originated
    TargetPath    string      // Where the event should be routed
}
```

**Fields**:

- **`Type`**: The event name that triggers transitions. Must match the `Event` field on a `Transition`.
- **`Payload`**: Arbitrary data (typically `map[string]interface{}`) passed to actions and guards.
- **`CorrelationID`**: UUID for tracing events across boundaries, regions, and runtimes.
- **`Source`**: Identifies the origin (e.g., `"web-frontend"`, `"region:payment"`, `"sys:heartbeat"`).
- **`TargetPath`**: Optional routing hint for parallel states (see [Event Addressing](#event-addressing)).

### Event Lifecycle

```
1. Dispatch → Event queued in ChartRuntime.eventQueue
2. Process   → Event dequeued, evaluated against current state
3. Match     → Transition found (guard evaluated if present)
4. Execute   → Exit actions → Transition actions → Entry actions
5. Complete  → State updated, trace callbacks fired
```

### Event Types

#### User Events

User-defined events that drive application logic:

```go
engine.Dispatch(rtID, statechart.Event{
    Type: "user_login",
    Payload: map[string]string{
        "username": "alice",
        "method": "oauth",
    },
    CorrelationID: "req-123",
    Source: "web-frontend",
})
```

**Characteristics**:

- Defined by chart authors in YAML or code
- Match transitions via `Type` field
- Can carry arbitrary payload data
- May be targeted to specific regions via `TargetPath`

#### System Events

Reserved events (prefixed with `sys:`) for internal coordination:

```go
const (
    // Parent → Region: Lifecycle control
    SysEnter  = "sys:enter"
    SysExit   = "sys:exit"
    SysPause  = "sys:pause"
    SysResume = "sys:resume"

    // Region → Parent: Coordination signals
    SysDone         = "sys:done"
    SysTransition   = "sys:transition"
    SysPanic        = "sys:panic"
    SysBackpressure = "sys:backpressure"
)
```

**Characteristics**:

- Automatically generated by the library
- Handle parallel region lifecycle and coordination
- Cannot be manually dispatched (ignored if sent)
- Use `Event.IsSystem()` to detect

### Event Addressing

Events can be targeted to specific destinations using the `TargetPath` field. The routing behavior depends on whether the runtime is in an atomic/compound state or a parallel state.

#### TargetPath Syntax

| Format | Description | Example |
|--------|-------------|---------|
| *(empty)* | Broadcast to current state or all regions | `""` |
| `region:<name>` | Target a specific parallel region | `"region:payment"` |
| `child:<name>` | Target a specific child state | `"child:processing"` |
| `<path>` | Full state path | `"root/orders/pending"` |

#### Atomic/Compound States

When the runtime is in an atomic or compound state, events are evaluated against the current state's transitions:

```go
// No TargetPath: event evaluated against current state
engine.Dispatch(rtID, statechart.Event{Type: "login"})

// With TargetPath: ignored in atomic/compound states
// (TargetPath is only meaningful for parallel states)
engine.Dispatch(rtID, statechart.Event{
    Type: "login",
    TargetPath: "child:authenticated", // Ignored
})
```

#### Parallel States

When in a parallel state, events are routed based on `TargetPath`:

```go
// Broadcast: sent to ALL regions
engine.Dispatch(rtID, statechart.Event{
    Type: "shutdown",
    // TargetPath empty → all regions receive it
})

// Targeted: sent to specific region only
engine.Dispatch(rtID, statechart.Event{
    Type: "processPayment",
    TargetPath: "region:payment", // Only payment region receives it
})

// Child targeting: sent to specific child within a region
engine.Dispatch(rtID, statechart.Event{
    Type: "updateStatus",
    TargetPath: "child:processing", // Targets child state
})
```

### Event Ordering Guarantees

**Synchronous** (guaranteed order):

- Within a single atomic/compound state
- Events processed FIFO from the queue
- No interleaving with other events

**Asynchronous** (eventual consistency):

- Across parallel regions
- Between parent and child runtimes
- Events may be processed out of order

### Event Queue Behavior

```go
// Events are queued and processed sequentially
engine.Dispatch(rtID, statechart.Event{Type: "event1"})
engine.Dispatch(rtID, statechart.Event{Type: "event2"})
engine.Dispatch(rtID, statechart.Event{Type: "event3"})

// Processing order: event1 → event2 → event3
// Each event completes before the next starts
```

**Key behaviors**:

- FIFO ordering within a runtime
- New events added to queue don't interrupt processing
- Queue is part of snapshot/restore
- Empty queue is a condition for quiescence

### Practical Examples

#### Example 1: Order Processing with Parallel Regions

```go
// Chart has parallel regions: inventory, payment, shipping
def := statechart.ChartDefinition{
    Root: &statechart.Node{
        ID: "order",
        Children: map[string]*statechart.Node{
            "inventory": { /* checks stock */ },
            "payment":   { /* processes payment */ },
            "shipping":  { /* arranges delivery */ },
        },
    },
}

// Broadcast cancel to all regions
engine.Dispatch(rtID, statechart.Event{
    Type: "cancel",
    Payload: map[string]string{"reason": "customer_request"},
})

// Only process payment (inventory and shipping continue independently)
engine.Dispatch(rtID, statechart.Event{
    Type: "processPayment",
    TargetPath: "region:payment",
    Payload: map[string]interface{}{
        "amount": 99.99,
        "currency": "USD",
    },
})

// Update shipping without affecting other regions
engine.Dispatch(rtID, statechart.Event{
    Type: "updateAddress",
    TargetPath: "region:shipping",
    Payload: map[string]string{
        "street": "123 Main St",
        "city": "Springfield",
    },
})
```

#### Example 2: Cross-Region Communication

Regions can emit events that are routed to other regions:

```go
// Register an action that emits a cross-region event
engine.RegisterAction("paymentComplete", func(ctx RuntimeContext, 
    appCtx ApplicationContext, ev Event) error {
    
    // Emit event targeted at shipping region
    // This would be handled by the runtime to dispatch to the target
    emitEvent := Event{
        Type: "startShipping",
        TargetPath: "region:shipping",
        CorrelationID: ev.CorrelationID,
    }
    
    // In practice, this would be queued for the target region
    // The library handles the routing automatically
    return nil
})
```

#### Example 3: Correlation IDs for Tracing

```go
// Use CorrelationID to track events across regions
reqID := "req-" + uuid.New().String()

engine.Dispatch(rtID, statechart.Event{
    Type: "processOrder",
    CorrelationID: reqID,
    Payload: map[string]interface{}{
        "orderId": "ORD-12345",
    },
})

// All child events and transitions will carry this ID
// for distributed tracing
```

#### Example 4: Event Source Tracking

```go
// Track where events originate
engine.Dispatch(rtID, statechart.Event{
    Type: "userAction",
    Source: "web-frontend",
    Payload: map[string]string{
        "action": "click",
        "element": "checkout-button",
    },
})

// System events automatically set Source to region name
// e.g., "region:payment" for events from the payment region
```

### Addressing Rules Summary

| Scenario | TargetPath | Routing Result |
|----------|------------|----------------|
| Atomic state, empty | `""` | Event evaluated against current state |
| Atomic state, non-empty | `"region:x"` | TargetPath ignored; event evaluated against current state |
| Parallel state, empty | `""` | Broadcast to ALL regions |
| Parallel state, `region:<name>` | `"region:payment"` | Sent only to named region |
| Parallel state, `child:<name>` | `"child:x"` | Sent to child state within regions |
| Region doesn't exist | `"region:unknown"` | Event dropped (no error) |

### System Event Addressing

System events (`sys:*`) have special routing:

- `sys:enter`, `sys:exit`, `sys:pause`, `sys:resume`: Sent by parent to regions
- `sys:done`, `sys:transition`, `sys:panic`, `sys:backpressure`: Sent by regions to parent

These are automatically handled by the library and don't require manual `TargetPath` configuration.

### Events vs Mail

**Important**: Events are local to a single ChartRuntime. For inter-agent communication, use the Mail system (not part of this library):

| Events | Mail |
|--------|------|
| Local to ChartRuntime | Cross-runtime communication |
| Synchronous within boundaries | Async, at-least-once delivery |
| Drive state transitions | Carry messages between agents |
| No global addressing | Topics, agent inboxes, sys:* services |
| Part of statechart library | Part of Maelstrom application layer |

See [arch-v1.md Section 9](../../docs/arch-v1.md#9-inter-agent-communication) for Mail system details.

## Transitions

Transitions define how states change in response to events. Each transition consists of an event trigger, optional guard, target state, and optional actions.

### Transition Structure

```go
type Transition struct {
    Event   string   // Event type that triggers this transition
    Guard   string   // Name of registered guard (empty = no guard)
    Actions []string // Action names to execute during transition
    Target  string   // Target state path
}
```

### Transition Types

#### Internal Transitions

Transitions within the same state (target = current state):

```go
node := &statechart.Node{
    ID: "processing",
    Transitions: []statechart.Transition{
        {
            Event:  "retry",
            Target: "processing", // Same state
            Actions: []string{"logRetry", "incrementCounter"},
        },
    },
}
```

**Behavior:**

- Exit actions are NOT executed
- Transition actions ARE executed
- Entry actions are NOT executed
- Useful for side effects without state change

#### External Transitions

Transitions between different states:

```go
node := &statechart.Node{
    ID: "pending",
    Transitions: []statechart.Transition{
        {
            Event:  "approved",
            Target: "processing",
            Actions: []string{"notifyUser", "startTimer"},
        },
    },
}
```

**Behavior:**

1. Exit actions for source state are executed
2. Transition actions are executed
3. State is updated to target
4. Entry actions for target state are executed

#### Self-Transitions

Special case where target equals source state but full lifecycle runs:

```go
node := &statechart.Node{
    ID: "active",
    EntryActions: []string{"refreshData"},
    Transitions: []statechart.Transition{
        {
            Event:  "refresh",
            Target: "active", // Same state, but entry actions run
        },
    },
}
```

**Note:** In this library, self-transitions behave as internal transitions (no entry/exit actions). To force full lifecycle, use a temporary intermediate state.

### Transition Execution Order

For external transitions, actions execute in this order:

```
1. Exit actions (source state)
2. Transition actions (defined on transition)
3. Entry actions (target state)
```

Example:

```go
def := statechart.ChartDefinition{
    Root: &statechart.Node{
        ID: "root",
        Children: map[string]*statechart.Node{
            "idle": {
                ID: "idle",
                ExitActions: []string{"cleanupIdle"},
                Transitions: []statechart.Transition{
                    {
                        Event:   "start",
                        Target:  "running",
                        Actions: []string{"logTransition", "validateState"},
                    },
                },
            },
            "running": {
                ID: "running",
                EntryActions: []string{"initialize", "subscribeEvents"},
            },
        },
    },
}

// When "start" event fires:
// 1. cleanupIdle() executes
// 2. logTransition() executes
// 3. validateState() executes
// 4. initialize() executes
// 5. subscribeEvents() executes
```

### Guards

Guards are conditional checks that determine if a transition can fire:

```go
// Register guard
engine.RegisterGuard("hasPermission", func(appCtx ApplicationContext, ev Event) bool {
    payload := ev.Payload.(map[string]interface{})
    user := payload["user"].(string)
    return hasPermission(user, "edit")
})

// Use in transition
node := &statechart.Node{
    ID: "view",
    Transitions: []statechart.Transition{
        {
            Event:  "edit",
            Guard:  "hasPermission",
            Target: "editing",
        },
    },
}
```

**Guard Evaluation:**

- If guard returns `false`, transition is blocked (no-op)
- If guard returns `true`, transition proceeds
- If multiple transitions match the event, first matching guard wins
- Transitions without guards are evaluated first

### Multiple Transitions on Same Event

When multiple transitions share the same event, guards determine which fires:

```go
node := &statechart.Node{
    ID: "payment",
    Transitions: []statechart.Transition{
        {
            Event:  "submit",
            Guard:  "isValidCard",
            Target: "processing",
        },
        {
            Event:  "submit",
            Guard:  "isInvalidCard",
            Target: "error",
        },
        {
            Event:  "submit",
            Target: "default", // No guard - fallback
        },
    },
}
```

**Evaluation Order:**

1. Transitions without guards are checked first
2. Transitions with guards are evaluated in order
3. First matching transition fires
4. If no match, event is ignored (no-op)

### Transition Actions

Actions execute side effects during transitions:

```go
// Register action
engine.RegisterAction("sendEmail", func(ctx RuntimeContext, 
    appCtx ApplicationContext, ev Event) error {
    
    email := ev.Payload.(map[string]interface{})["email"].(string)
    return sendConfirmationEmail(email)
})

// Use in transition
node := &statechart.Node{
    ID: "checkout",
    Transitions: []statechart.Transition{
        {
            Event:   "confirm",
            Target:  "confirmed",
            Actions: []string{"sendEmail", "updateInventory", "logOrder"},
        },
    },
}
```

**Action Execution:**

- Actions execute in order (left to right)
- Actions receive runtime context, application context, and triggering event
- Action errors are logged but don't block the transition
- All actions complete before state changes

### Practical Examples

#### Example 1: Order State Machine

```go
def := statechart.ChartDefinition{
    ID: "order-flow",
    Version: "1.0.0",
    Root: &statechart.Node{
        ID: "root",
        Children: map[string]*statechart.Node{
            "created": {
                ID: "created",
                EntryActions: []string{"sendConfirmation"},
                Transitions: []statechart.Transition{
                    {
                        Event:  "payment_received",
                        Guard:  "inventoryAvailable",
                        Target: "processing",
                        Actions: []string{"reserveInventory"},
                    },
                    {
                        Event:  "payment_failed",
                        Target: "cancelled",
                        Actions: []string{"notifyFailure"},
                    },
                },
            },
            "processing": {
                ID: "processing",
                EntryActions: []string{"startShipping"},
                Transitions: []statechart.Transition{
                    {
                        Event:  "shipped",
                        Target: "shipped",
                        Actions: []string{"updateTracking"},
                    },
                    {
                        Event:  "cancelled",
                        Guard:  "canCancel",
                        Target: "cancelled",
                        Actions: []string{"releaseInventory", "refund"},
                    },
                },
            },
            "shipped": {
                ID: "shipped",
                EntryActions: []string{"sendShipmentNotification"},
                Transitions: []statechart.Transition{
                    {
                        Event:  "delivered",
                        Target: "completed",
                    },
                },
            },
            "completed": {
                ID: "completed",
                EntryActions: []string{"sendCompletionSurvey"},
                // No transitions - final state
            },
            "cancelled": {
                ID: "cancelled",
                EntryActions: []string{"cleanup"},
                // No transitions - final state
            },
        },
    },
    InitialState: "root/created",
}
```

#### Example 2: Retry Logic with Guards

```go
// Register retry guard
engine.RegisterGuard("canRetry", func(appCtx ApplicationContext, ev Event) bool {
    ctx := appCtx.(RetryContext)
    return ctx.RetryCount < maxRetries
})

engine.RegisterAction("incrementRetry", func(ctx RuntimeContext, 
    appCtx ApplicationContext, ev Event) error {
    appCtx.Set("retry_count", getRetryCount()+1, nil, "system")
    return nil
})

node := &statechart.Node{
    ID: "processing",
    Transitions: []statechart.Transition{
        {
            Event:   "error",
            Guard:   "canRetry",
            Target:  "processing", // Self-transition
            Actions: []string{"incrementRetry", "logRetry"},
        },
        {
            Event:  "error",
            Target: "failed", // Fallback when canRetry is false
        },
    },
}
```

#### Example 3: Parallel State Transitions

```go
def := statechart.ChartDefinition{
    Root: &statechart.Node{
        ID: "parallel_checkout",
        Children: map[string]*statechart.Node{
            "payment_region": {
                ID: "payment",
                Children: map[string]*statechart.Node{
                    "waiting": {
                        ID: "waiting",
                        Transitions: []statechart.Transition{
                            {
                                Event:  "pay",
                                Target: "processing",
                            },
                        },
                    },
                    "processing": {
                        ID: "processing",
                        Transitions: []statechart.Transition{
                            {
                                Event:  "complete",
                                Target: "done",
                            },
                        },
                    },
                    "done": {
                        ID: "done",
                        // Final state
                    },
                },
            },
            "inventory_region": {
                ID: "inventory",
                Children: map[string]*statechart.Node{
                    "checking": {
                        ID: "checking",
                        Transitions: []statechart.Transition{
                            {
                                Event:  "verified",
                                Target: "reserved",
                            },
                        },
                    },
                    "reserved": {
                        ID: "reserved",
                        // Final state
                    },
                },
            },
        },
    },
}

// Each region transitions independently
// Payment region: waiting → processing → done
// Inventory region: checking → reserved
```

#### Example 4: Compound State Transitions

```go
def := statechart.ChartDefinition{
    Root: &statechart.Node{
        ID: "root",
        Children: map[string]*statechart.Node{
            "order": {
                ID: "order",
                // Compound state - has children
                Children: map[string]*statechart.Node{
                    "draft": {
                        ID: "draft",
                        IsInitial: true,
                        Transitions: []statechart.Transition{
                            {
                                Event:  "submit",
                                Target: "review",
                            },
                        },
                    },
                    "review": {
                        ID: "review",
                        Transitions: []statechart.Transition{
                            {
                                Event:  "approve",
                                Target: "confirmed",
                            },
                            {
                                Event:  "reject",
                                Target: "draft",
                            },
                        },
                    },
                    "confirmed": {
                        ID: "confirmed",
                    },
                },
                // Transitions on compound state affect all children
                Transitions: []statechart.Transition{
                    {
                        Event:  "cancel_order",
                        Target: "cancelled",
                        Actions: []string{"refundAll"},
                    },
                },
            },
            "cancelled": {
                ID: "cancelled",
            },
        },
    },
    InitialState: "root/order/draft", // Auto-enters initial child
}
```

## Chart Definition (YAML)

Charts are typically defined in YAML and hydrated via `DefaultHydrator()`:

```yaml
id: order-processor
version: 1.0.0
root:
  id: root
  children:
    pending:
      id: pending
      transitions:
        - event: paid
          target: processing
          actions:
            - logPayment
          guard: isValidOrder
    processing:
      id: processing
      entry:
        - startProcessing
      transitions:
        - event: complete
          target: shipped
    shipped:
      id: shipping
      isInitial: true
initialState: root/pending
```

Environment variable substitution is supported:

```yaml
id: ${SERVICE_ID:-default-service}
version: ${VERSION:-1.0.0}
```

## Error Handling

Common errors returned by the library:

- `ErrRuntimeNotFound`: Runtime ID doesn't exist
- `ErrInvalidControlCmd`: Invalid lifecycle command
- `ErrRuntimeNotRunning`: Operation requires running runtime
- `ErrInvalidState`: Invalid state transition
- `ErrActionNotFound`: Referenced action not registered
- `ErrGuardNotFound`: Referenced guard not registered

## Concurrency

- Engine is safe for concurrent use across multiple goroutines
- Each runtime processes events sequentially (FIFO queue)
- Parallel regions run in separate goroutines with internal synchronization

## See Also

- [Architecture Documentation](../../docs/arch-v1.md)
- [Usage Examples](../../docs/usage-examples.md)
