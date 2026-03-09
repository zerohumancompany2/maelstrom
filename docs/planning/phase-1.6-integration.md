# Phase 1.6: Integration - End-to-End Bootstrap Verification

**Branch**: `feat/integration-tests`  
**Parent Spec**: `docs/planning/layer-01-kernel-bootstrap.md` Section 5.6  
**Status**: Ready for TDD execution

---

## Executive Summary

This phase implements end-to-end integration tests that verify the complete bootstrap flow from kernel start to dormant state. All previous phases must be complete before starting this phase.

---

## 1. Current State Analysis

### 1.1 Kernel Start() Flow (pkg/kernel/kernel.go)
```
Start(ctx) → LoadBootstrapChart() → Spawn(bootstrapRTID) → Control(CmdStart)
           → NewSequence() → OnStateEnter() → OnComplete() → seq.Start()
           → Wait for ctx.Done()
```

**Key Observations:**
- Bootstrap chart is loaded and spawned with statechart engine
- Sequence tracks state transitions (security → communication → observability → lifecycle → handoff → complete)
- Services are "loaded" via onBootstrapStateEnter (currently just logs + async event emission)
- onBootstrapComplete() logs "going dormant" but has no actual implementation

### 1.2 Bootstrap Chart Structure (pkg/bootstrap/chart.go)
- **Chart ID**: `sys:bootstrap` v1.0.0
- **States**: initializing → security → communication → observability → lifecycle → handoff → complete
- **4 Core Services**: sys:security, sys:communication, sys:observability, sys:lifecycle
- **Events**: SECURITY_READY, COMMUNICATION_READY, OBSERVABILITY_READY, LIFECYCLE_READY, KERNEL_READY

### 1.3 Existing Tests Status
| Test File | Tests | Status |
|-----------|-------|--------|
| kernel_test.go | TestKernel_StartsBootstrapAndWaits | Uses `time.Sleep(100ms)` - flaky |
| kernel_test.go | TestKernel_BootstrapCompletes | Uses `time.Sleep(800ms)` - flaky |
| kernel_test.go | TestKernel_SpawnsAllServices | Placeholder |
| kernel_test.go | TestKernel_ServicesReady | Placeholder |
| kernel_test.go | TestKernel_KernelReadyEvent | Placeholder |
| kernel_test.go | TestKernel_MailSystemRequired | Placeholder |
| sequence_test.go | TestBootstrapSequence_CompleteSequence | Manual event simulation - good foundation |

---

## 2. Definition: What "Complete Bootstrap Flow" Means

A **complete bootstrap flow** is verified when ALL of the following conditions are met:

```
┌─────────────────────────────────────────────────────────────────────────┐
│ COMPLETE BOOTSTRAP FLOW VERIFICATION CRITERIA                           │
├─────────────────────────────────────────────────────────────────────────┤
│ 1. KERNEL STARTS                                                        │
│    - Kernel.Start(ctx) returns without immediate error                  │
│    - Bootstrap chart loaded successfully                                │
│    - Bootstrap runtime spawned (if engine provided)                     │
├─────────────────────────────────────────────────────────────────────────┤
│ 2. BOOTSTRAP CHART SPAWNS                                               │
│    - Runtime ID assigned to bootstrap runtime                           │
│    - Runtime state transitions: Created → Running                       │
│    - Initial state is "initializing"                                    │
├─────────────────────────────────────────────────────────────────────────┤
│ 3. SERVICES LOAD IN ORDER                                               │
│    - sys:security loads first (state: security)                         │
│    - sys:communication loads second (state: communication)              │
│    - sys:observability loads third (state: observability)               │
│    - sys:lifecycle loads fourth (state: lifecycle)                      │
│    - Each service emits READY event within timeout (e.g., 100ms)        │
├─────────────────────────────────────────────────────────────────────────┤
│ 4. STATE TRANSITIONS COMPLETE                                           │
│    - security → communication (on SECURITY_READY)                       │
│    - communication → observability (on COMMUNICATION_READY)             │
│    - observability → lifecycle (on OBSERVABILITY_READY)                 │
│    - lifecycle → handoff (on LIFECYCLE_READY)                           │
│    - handoff → complete (on KERNEL_READY)                               │
├─────────────────────────────────────────────────────────────────────────┤
│ 5. KERNEL_READY SIGNAL EMITTED                                          │
│    - KERNEL_READY event sent from handoff state                         │
│    - onBootstrapComplete() callback invoked                             │
│    - Kernel.IsBootstrapComplete() returns true                          │
├─────────────────────────────────────────────────────────────────────────┤
│ 6. KERNEL DORMANT                                                       │
│    - onBootstrapComplete() logs "going dormant"                         │
│    - Kernel ready to hand off to ChartRegistry                          │
└─────────────────────────────────────────────────────────────────────────┘
```

---

## 3. Integration Tests to Write (in TDD Order)

### Test 1: `TestKernel_LoadsBootstrapChart`

| Aspect | Detail |
|--------|--------|
| **What it verifies** | Bootstrap chart loads without error |
| **Expected failure** | Chart loading returns error or chart is invalid |
| **Minimal implementation** | Call `bootstrap.LoadBootstrapChart()` and assert no error + valid ID |

```go
func TestKernel_LoadsBootstrapChart(t *testing.T) {
    def, err := bootstrap.LoadBootstrapChart()
    assert.NoError(t, err)
    assert.Equal(t, "sys:bootstrap", def.ID)
    assert.Equal(t, "1.0.0", def.Version)
    assert.Equal(t, "initializing", def.InitialState)
}
```

---

### Test 2: `TestKernel_SpawnsBootstrapRuntime`

| Aspect | Detail |
|--------|--------|
| **What it verifies** | Kernel spawns bootstrap runtime with engine |
| **Expected failure** | Spawn returns error or runtime ID is empty |
| **Minimal implementation** | Create kernel with engine, verify Spawn is called |

```go
func TestKernel_SpawnsBootstrapRuntime(t *testing.T) {
    engine := statechart.NewEngine()
    kernel := NewWithEngine(engine)
    
    ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
    defer cancel()
    
    // Need to verify spawn happens during Start()
    // Expected: bootstrapRTID is non-empty
}
```

---

### Test 3: `TestKernel_BootstrapSequenceStartsWithSecurity`

| Aspect | Detail |
|--------|--------|
| **What it verifies** | First state entered is "security" |
| **Expected failure** | First state is not "security" |
| **Minimal implementation** | Track state entry callbacks, assert first is "security" |

```go
func TestKernel_BootstrapSequenceStartsWithSecurity(t *testing.T) {
    kernel := New()
    var firstState string
    
    ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
    defer cancel()
    
    go kernel.Start(ctx)
    time.Sleep(50 * time.Millisecond)
    
    // Assert: kernel.sequence.CurrentState() == "security"
}
```

---

### Test 4: `TestKernel_ServicesLoadInOrder`

| Aspect | Detail |
|--------|--------|
| **What it verifies** | Services load in correct order: security → communication → observability → lifecycle |
| **Expected failure** | States not entered in correct order |
| **Minimal implementation** | Track all state entries, verify order |

```go
func TestKernel_ServicesLoadInOrder(t *testing.T) {
    kernel := New()
    var stateOrder []string
    
    ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
    defer cancel()
    
    go kernel.Start(ctx)
    time.Sleep(1 * time.Second)
    
    expected := []string{"security", "communication", "observability", "lifecycle", "handoff", "complete"}
    // Assert: stateOrder matches expected
}
```

---

### Test 5: `TestKernel_AllServicesEmitReadyEvents`

| Aspect | Detail |
|--------|--------|
| **What it verifies** | All 4 services emit their READY events |
| **Expected failure** | One or more READY events not received |
| **Minimal implementation** | Track events emitted, verify all 4 present |

```go
func TestKernel_AllServicesEmitReadyEvents(t *testing.T) {
    kernel := New()
    var eventsReceived []string
    
    ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
    defer cancel()
    
    go kernel.Start(ctx)
    time.Sleep(1 * time.Second)
    
    expectedEvents := []string{"SECURITY_READY", "COMMUNICATION_READY", 
                                "OBSERVABILITY_READY", "LIFECYCLE_READY"}
    // Assert: all events received
}
```

---

### Test 6: `TestKernel_KernelReadyEventEmitted`

| Aspect | Detail |
|--------|--------|
| **What it verifies** | KERNEL_READY event emitted after all services ready |
| **Expected failure** | KERNEL_READY not received |
| **Minimal implementation** | Track KERNEL_READY event, verify it's last |

```go
func TestKernel_KernelReadyEventEmitted(t *testing.T) {
    kernel := New()
    var kernelReadyReceived bool
    
    ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
    defer cancel()
    
    go kernel.Start(ctx)
    time.Sleep(1 * time.Second)
    
    // Assert: kernelReadyReceived == true
}
```

---

### Test 7: `TestKernel_BootstrapCompletes`

| Aspect | Detail |
|--------|--------|
| **What it verifies** | `kernel.IsBootstrapComplete()` returns true |
| **Expected failure** | `IsBootstrapComplete()` returns false |
| **Minimal implementation** | Poll `IsBootstrapComplete()` until true or timeout |

```go
func TestKernel_BootstrapCompletes(t *testing.T) {
    kernel := New()
    
    ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
    defer cancel()
    
    go kernel.Start(ctx)
    
    // Poll for completion
    for deadline := ctx.Deadline(); time.Now().Before(*deadline); {
        if kernel.IsBootstrapComplete() {
            return // Success
        }
        time.Sleep(10 * time.Millisecond)
    }
    
    t.Error("Bootstrap did not complete within timeout")
}
```

---

### Test 8: `TestKernel_OnCompleteCallbackInvoked`

| Aspect | Detail |
|--------|--------|
| **What it verifies** | `onBootstrapComplete()` callback is invoked |
| **Expected failure** | Callback not invoked |
| **Minimal implementation** | Track callback invocation, assert it happened |

```go
func TestKernel_OnCompleteCallbackInvoked(t *testing.T) {
    kernel := New()
    var completeCalled bool
    
    // Need to hook into kernel's completion tracking
    ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
    defer cancel()
    
    go kernel.Start(ctx)
    time.Sleep(1 * time.Second)
    
    // Assert: completeCalled == true
}
```

---

### Test 9: `TestKernel_GoesDormant`

| Aspect | Detail |
|--------|--------|
| **What it verifies** | Kernel logs "going dormant" after bootstrap |
| **Expected failure** | Dormant message not logged |
| **Minimal implementation** | Capture log output, verify message present |

```go
func TestKernel_GoesDormant(t *testing.T) {
    kernel := New()
    
    // Capture log output
    var logOutput strings.Builder
    
    ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
    defer cancel()
    
    go kernel.Start(ctx)
    time.Sleep(1 * time.Second)
    
    // Assert: logOutput contains "going dormant"
}
```

---

### Test 10: `TestKernel_FullE2EBootstrap`

| Aspect | Detail |
|--------|--------|
| **What it verifies** | Complete end-to-end bootstrap flow |
| **Expected failure** | Any part of flow fails |
| **Minimal implementation** | Combine all assertions into one comprehensive test |

```go
func TestKernel_FullE2EBootstrap(t *testing.T) {
    kernel := New()
    
    // Track everything
    var statesEntered []string
    var eventsReceived []string
    var completeCalled bool
    
    ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
    defer cancel()
    
    go kernel.Start(ctx)
    
    // Wait for completion with polling
    for deadline := ctx.Deadline(); time.Now().Before(*deadline); {
        if kernel.IsBootstrapComplete() {
            break
        }
        time.Sleep(10 * time.Millisecond)
    }
    
    // Verify ALL criteria
    assertStatesInOrder(t, statesEntered)
    assertAllEventsReceived(t, eventsReceived)
    assertCompleteCalled(t, completeCalled)
}
```

---

## 4. Metrics & Assertions Checklist

### 4.1 Service Count Metrics
| Metric | Expected Value | How to Verify |
|--------|---------------|---------------|
| Services loaded | 4 | Count state entries (security, communication, observability, lifecycle) |
| Services ready | 4 | Count READY events received |
| Total states entered | 6 | security, communication, observability, lifecycle, handoff, complete |

### 4.2 Events Emitted
| Event | Source State | Target State |
|-------|-------------|--------------|
| SECURITY_READY | security | communication |
| COMMUNICATION_READY | communication | observability |
| OBSERVABILITY_READY | observability | lifecycle |
| LIFECYCLE_READY | lifecycle | handoff |
| KERNEL_READY | handoff | complete |

### 4.3 Timing Assertions
| Metric | Max Expected | How to Measure |
|--------|-------------|----------------|
| Total bootstrap time | < 2 seconds | Time from Start() to IsBootstrapComplete() |
| Per-service ready time | < 500ms | Time from state entry to READY event |
| State transition time | < 100ms | Time between state entries |

### 4.4 State Assertions
| Assertion | Expected Value |
|-----------|---------------|
| Initial state | "initializing" |
| First loaded state | "security" |
| Final state | "complete" |
| IsComplete() at end | true |

---

## 5. Implementation Dependencies (What Needs to Be Built First)

### 5.1 Kernel Observability (Required for Tests)
```go
// Add to kernel.go:
type Kernel struct {
    // ... existing fields
    
    // For testing/observability
    statesEntered   []string
    eventsReceived  []string
    onCompleteCalled bool
}

func (k *Kernel) GetStatesEntered() []string { /* ... */ }
func (k *Kernel) GetEventsReceived() []string { /* ... */ }
func (k *Kernel) GetCompletionStatus() bool { /* ... */ }
```

### 5.2 Sequence Observability (Required for Tests)
```go
// Add to sequence.go:
type Sequence struct {
    // ... existing fields
    
    // For testing/observability
    statesEntered []string
    eventsHandled []string
}

func (s *Sequence) GetStatesEntered() []string { /* ... */ }
func (s *Sequence) GetEventsHandled() []string { /* ... */ }
```

---

## 6. Test Execution Order (TDD Compliance)

```
Phase 1.6 Test Order:
1. TestKernel_LoadsBootstrapChart           → Stub: LoadBootstrapChart()
2. TestKernel_SpawnsBootstrapRuntime        → Stub: Spawn() call
3. TestKernel_BootstrapSequenceStartsWithSecurity → Stub: State tracking
4. TestKernel_ServicesLoadInOrder           → Stub: Order verification
5. TestKernel_AllServicesEmitReadyEvents    → Stub: Event tracking
6. TestKernel_KernelReadyEventEmitted       → Stub: KERNEL_READY tracking
7. TestKernel_BootstrapCompletes            → Stub: IsComplete() polling
8. TestKernel_OnCompleteCallbackInvoked     → Stub: Callback tracking
9. TestKernel_GoesDormant                   → Stub: Log capture
10. TestKernel_FullE2EBootstrap             → Integration of all above
```

---

## 7. Summary of Verification Approach

| Verification Layer | Method | Tools |
|-------------------|--------|-------|
| **Chart Loading** | Unit test | `bootstrap.LoadBootstrapChart()` |
| **Runtime Spawning** | Integration test | `statechart.Engine.Spawn()` |
| **State Transitions** | Integration test | `Sequence.CurrentState()` |
| **Service Loading** | Integration test | State entry callbacks |
| **Event Emission** | Integration test | Event tracking in sequence |
| **Completion** | Integration test | `IsBootstrapComplete()` |
| **Dormant State** | Integration test | Log output capture |
| **Full E2E** | End-to-end test | All above combined |

---

## 8. Key Files to Modify

| File | Changes Needed |
|------|---------------|
| `pkg/kernel/kernel.go` | Add observability methods, track states/events |
| `pkg/bootstrap/sequence.go` | Add observability methods for states/events |
| `pkg/kernel/kernel_test.go` | Replace placeholders with real tests |
| `pkg/integration/e2e_test.go` | Add bootstrap-specific e2e tests |

---

## 9. Summary

| Metric | Value |
|--------|-------|
| Total Integration Tests | 10 |
| Estimated LOC | ~200 |
| Estimated Time | 3-4 hours |
| Prerequisites | Phases 1.1-1.5 complete |

**Definition of Done**:
1. All 10 integration tests pass
2. Complete bootstrap flow verified end-to-end
3. No flaky tests (no time.Sleep without proper synchronization)
4. All timing assertions pass consistently

---

**End of Phase 1.6 Integration Plan**