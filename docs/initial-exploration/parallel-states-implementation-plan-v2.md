# Parallel States Implementation Plan v2

Clarified semantics for self-routing and timeout configuration. This document supersedes the ambiguous sections of v1.

---

## Clarified Semantics

### 1. "Self" Routing in Parallel States

**Rule:** Omitted `TargetPath` or `"."` routes to the **region that dispatched the event**.

**Rationale:** Parallel regions are independent state machines. Broadcasting violates independence—an event dispatched from Region A's "decide" state shouldn't automatically be evaluated in Region B's "monitor" state.

**Implementation:**

```go
// Event carries implicit source region context
type Event struct {
    Type          string
    Payload       any
    CorrelationID string
    Source        string
    TargetPath    string  // explicit routing
    SourceRegion  string  // NEW: implicit, set by Dispatch
}

func (e *Engine) Dispatch(id RuntimeID, ev Event) error {
    runtime := e.getRuntime(id)

    target, err := parseTargetPath(ev.TargetPath)
    if err != nil {
        return err
    }

    if target.IsSelf {
        // Route to source region (must be set by caller context)
        if ev.SourceRegion == "" {
            return fmt.Errorf("self-routing requires SourceRegion context")
        }
        region, exists := runtime.regions[ev.SourceRegion]
        if !exists {
            return fmt.Errorf("source region %s not found", ev.SourceRegion)
        }
        region.eventQueue <- ev
    } else if target.Region != "" {
        // Explicit cross-region routing
        region, exists := runtime.regions[target.Region]
        if !exists {
            return fmt.Errorf("region %s not found", target.Region)
        }
        region.eventQueue <- ev
    }

    return nil
}
```

**Cross-Region Communication Pattern:**

```yaml
# Region A wants to signal Region B
nodes:
  root:
    type: parallel
    regions:
      - name: regionA
        child:
          waiting:
            type: atomic
            transitions:
              - event: "proceed"
                action: ["signalRegionB"]  # dispatches with TargetPath: "region:regionB"
                to: "done"
          done:
            type: final  # emits implicit done
      - name: regionB
        child:
          idle:
            type: atomic
            transitions:
              - event: "startWork"  # sent from regionA via explicit routing
                to: "working"
          working:
            type: final
```

---

### 2. Timeout Configuration (Force-Stop Policy)

**Rule:** Timeout policies are **ChartDefinition configuration**, not hardcoded values. The Library implements the mechanism; the Chart specifies the policy.

**Rationale:** Different charts have different requirements. A monitoring agent might tolerate 30s for graceful shutdown; a real-time trading agent needs 100ms force-stop.

**Configuration Location:**

```yaml
# ChartDefinition spec
spec:
  stabilityPolicy:
    maxReconfigDepth: 3
    collapseAfter: 5

  # NEW: Parallel state timeout policy
  parallelPolicy:
    defaultTimeoutMs: 30000        # default for graceful stop
    forceStopTimeoutMs: 5000       # hard limit before SIGKILL equivalent

    # Per-state overrides (optional)
    stateTimeouts:
      - statePath: "root/orient"
        timeoutMs: 10000
        forceStopMs: 2000
```

**Library API Changes:**

```go
// ChartDefinition adds ParallelPolicy
type ChartDefinition struct {
    // ... existing fields
    ParallelPolicy ParallelPolicy
}

type ParallelPolicy struct {
    DefaultTimeoutMs   int
    ForceStopTimeoutMs int
    StateTimeouts      []StateTimeout
}

type StateTimeout struct {
    StatePath      string
    TimeoutMs      int
    ForceStopMs    int
}

// Control command accepts timeout override
type ControlCmd struct {
    Type      ControlCmdType
    TimeoutMs *int  // optional override for this operation
}
```

**Implementation in Exit Flow:**

```go
func (e *Engine) exitParallelState(runtime *ChartRuntime, ev Event, policy ParallelPolicy) error {
    // Determine effective timeout
    timeoutMs := policy.DefaultTimeoutMs
    forceStopMs := policy.ForceStopTimeoutMs

    // Check for state-specific override
    currentPath := runtime.getCurrentParallelPath()
    for _, st := range policy.StateTimeouts {
        if st.StatePath == currentPath {
            timeoutMs = st.TimeoutMs
            forceStopMs = st.ForceStopMs
            break
        }
    }

    // Phase 1: Graceful stop signal
    ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeoutMs)*time.Millisecond)
    defer cancel()

    for _, region := range runtime.regions {
        region.gracefulStop <- struct{}{}
    }

    // Phase 2: Wait for completion or timeout
    done := make(chan struct{})
    go func() {
        for _, region := range runtime.regions {
            <-region.done
        }
        close(done)
    }()

    select {
    case <-done:
        // Graceful completion
    case <-ctx.Done():
        // Phase 3: Force stop
        log.Printf("Timeout exceeded, force-stopping regions")
        for _, region := range runtime.regions {
            close(region.forceStop) // immediate termination
        }

        // Hard limit on force-stop
        forceCtx, forceCancel := context.WithTimeout(context.Background(), time.Duration(forceStopMs)*time.Millisecond)
        defer forceCancel()

        select {
        case <-done:
        case <-forceCtx.Done():
            return fmt.Errorf("force-stop timeout exceeded, regions may be leaked")
        }
    }

    // Execute exit actions...
    return nil
}
```

---

## Key Differences from v1

| Aspect             | v1 (Ambiguous)                  | v2 (Clarified)                                                |
| ------------------ | ------------------------------- | ------------------------------------------------------------- |
| Self-routing       | Unclear (broadcast vs targeted) | Routes to source region only; requires `SourceRegion` context |
| Cross-region comms | Not specified                   | Explicit `TargetPath: "region:name"` required                 |
| Timeout config     | Hardcoded values                | ChartDefinition `parallelPolicy` block                        |
| Force-stop         | Binary on/off                   | Two-phase: graceful timeout → force-stop timeout              |
| Per-state timeouts | Not supported                   | `stateTimeouts` array in policy                               |

---

## Open Questions (Still)

1. **SourceRegion injection**: Should the Engine automatically set `SourceRegion` on all dispatches, or should actions explicitly include it?

2. **Region-local transitions**: When a region transitions, should the parent be notified synchronously (for ActiveStates tracking) or is async via `RegionMessage` acceptable?

3. **Exit action ordering**: The arch says "no order guarantee" for parallel exit actions. Is concurrent execution acceptable, or should we attempt best-effort synchronization?

---

*End of v2 Clarifications*