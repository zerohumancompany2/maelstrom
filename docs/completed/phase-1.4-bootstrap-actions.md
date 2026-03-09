# Phase 1.4: Bootstrap Actions - TDD Implementation Plan

**Branch**: `feat/bootstrap-actions`  
**Parent Spec**: `docs/planning/layer-01-kernel-bootstrap.md` Section 3  
**Status**: Ready for TDD execution

---

## Executive Summary

This plan implements the bootstrap actions referenced in `pkg/bootstrap/chart.go` that currently only log. Each action will spawn a service runtime via the statechart engine and signal completion by dispatching readiness events.

---

## Part 1: Current State Analysis

### 1.1 How Services Are Currently "Loaded" (kernel.go)

**Current Implementation** (lines 92-135):
```go
func (k *Kernel) onBootstrapStateEnter(ctx context.Context, state string, bootstrapRTID statechart.RuntimeID) error {
    switch state {
    case "security":
        log.Println("[kernel] Loading sys:security service")
        go func() { seq.HandleEvent(ctx, "SECURITY_READY") }()
    // ... similar for communication, observability, lifecycle, handoff
    }
    return nil
}
```

**Problem**: Just logs and immediately fires READY events without actually loading anything.

### 1.2 Actions Referenced in YAML (chart.go)

| State | Entry Action | Current Status |
|-------|-------------|----------------|
| security | `loadSecurityService` | Not registered |
| communication | `loadCommunicationService` | Not registered |
| observability | `loadObservabilityService` | Not registered |
| lifecycle | `loadLifecycleService` | Not registered |
| handoff | `signalKernelReady` | Not registered |

### 1.3 ActionFn Signature (interfaces.go)

```go
type ActionFn func(runtimeCtx RuntimeContext, appCtx ApplicationContext, event Event) error
```

**Parameters**:
- `runtimeCtx`: Read-only context (ChartID, RuntimeID, ActiveStates, StartTime, Metrics)
- `appCtx`: Read-write application context with namespace isolation
- `event`: The triggering event (Type, Payload, CorrelationID, Source)

### 1.4 How Actions Integrate with Statechart Engine

From `engine.go`:
1. Actions registered via `engine.RegisterAction(name, fn)`
2. On spawn, engine copies actions to runtime's local registry
3. Entry actions executed via `executeEntryActions()` (lines 451-477)
4. Actions can access runtime context, app context, and triggering event

---

## Part 2: Bootstrap Actions Specification

### 2.1 Action Signatures

All bootstrap actions follow the same pattern:

```go
// Package-level constants for action names
const (
    ActionLoadSecurityService     = "loadSecurityService"
    ActionLoadCommunicationService = "loadCommunicationService"
    ActionLoadObservabilityService = "loadObservabilityService"
    ActionLoadLifecycleService    = "loadLifecycleService"
    ActionSignalKernelReady       = "signalKernelReady"
)

// ServiceLoadParams - parameters passed to load actions via appCtx
type ServiceLoadParams struct {
    ServiceID     string   // e.g., "sys:security"
    BootstrapRTID string   // parent bootstrap runtime ID
    Dependencies  []string // required services (for dependency ordering)
}

// KernelReadyParams - parameters for signalKernelReady
type KernelReadyParams struct {
    BootstrapRTID string
    LoadedServices []string // list of successfully loaded service IDs
}
```

### 2.2 Action Specifications

#### Action 1: `loadSecurityService`

**Purpose**: Spawn sys:security runtime, wait for READY, signal bootstrap

**Parameters** (retrieved from appCtx):
```go
type SecurityLoadParams struct {
    ServiceID     string // "sys:security"
    BootstrapRTID string // parent bootstrap runtime ID
}
```

**What it does**:
1. Load security chart definition (from embedded YAML or file)
2. Spawn security runtime via `engine.SpawnTransient(def, appCtx, bootstrapRTID)`
3. Start the runtime via `engine.Control(securityRTID, CmdStart)`
4. Store securityRTID in appCtx for future reference
5. Dispatch `SECURITY_READY` event to bootstrap parent

**Events Emitted**:
- `SECURITY_READY` to bootstrap runtime (via engine.Dispatch)

**AppCtx Writes**:
- Key: `"bootstrap:security:runtimeID"` → Value: securityRTID string

---

#### Action 2: `loadCommunicationService`

**Purpose**: Spawn sys:communication runtime, wait for READY, signal bootstrap

**Parameters**:
```go
type CommunicationLoadParams struct {
    ServiceID     string // "sys:communication"
    BootstrapRTID string
    SecurityRTID  string // dependency - must be loaded first
}
```

**What it does**:
1. Load communication chart definition
2. Spawn communication runtime (transient child of bootstrap)
3. Start the runtime
4. Store communicationRTID in appCtx
5. Dispatch `COMMUNICATION_READY` event to bootstrap

**Events Emitted**:
- `COMMUNICATION_READY` to bootstrap runtime

**AppCtx Writes**:
- Key: `"bootstrap:communication:runtimeID"` → Value: communicationRTID string

---

#### Action 3: `loadObservabilityService`

**Purpose**: Spawn sys:observability runtime, wait for READY, signal bootstrap

**Parameters**:
```go
type ObservabilityLoadParams struct {
    ServiceID     string // "sys:observability"
    BootstrapRTID string
    SecurityRTID  string // dependency
}
```

**What it does**:
1. Load observability chart definition
2. Spawn observability runtime
3. Start the runtime
4. Store observabilityRTID in appCtx
5. Dispatch `OBSERVABILITY_READY` event to bootstrap

**Events Emitted**:
- `OBSERVABILITY_READY` to bootstrap runtime

**AppCtx Writes**:
- Key: `"bootstrap:observability:runtimeID"` → Value: observabilityRTID string

---

#### Action 4: `loadLifecycleService`

**Purpose**: Spawn sys:lifecycle runtime, wait for READY, signal bootstrap

**Parameters**:
```go
type LifecycleLoadParams struct {
    ServiceID     string // "sys:lifecycle"
    BootstrapRTID string
    SecurityRTID  string // dependency
    CommRTID      string // dependency
}
```

**What it does**:
1. Load lifecycle chart definition
2. Spawn lifecycle runtime
3. Start the runtime
4. Store lifecycleRTID in appCtx
5. Dispatch `LIFECYCLE_READY` event to bootstrap

**Events Emitted**:
- `LIFECYCLE_READY` to bootstrap runtime

**AppCtx Writes**:
- Key: `"bootstrap:lifecycle:runtimeID"` → Value: lifecycleRTID string

---

#### Action 5: `signalKernelReady`

**Purpose**: Aggregate loaded services, signal kernel bootstrap complete

**Parameters**:
```go
type KernelReadyParams struct {
    BootstrapRTID string
}
```

**What it does**:
1. Read all service runtime IDs from appCtx
2. Build list of loaded services
3. Store loaded services list in appCtx for kernel to read
4. Dispatch `KERNEL_READY` event to bootstrap

**Events Emitted**:
- `KERNEL_READY` to bootstrap runtime

**AppCtx Writes**:
- Key: `"bootstrap:loaded:services"` → Value: []string of service IDs

---

### 2.3 Action Integration with Kernel

The kernel must:
1. **Register all actions** with the engine before spawning bootstrap
2. **Set up appCtx** with bootstrap runtime ID for actions to use
3. **Track loaded services** in its own state (copy from appCtx after bootstrap)

**Kernel Integration Pattern**:
```go
func (k *Kernel) Start(ctx context.Context) error {
    // ... existing code ...
    
    // NEW: Register bootstrap actions with engine
    k.registerBootstrapActions()
    
    // ... spawn bootstrap ...
}

func (k *Kernel) registerBootstrapActions() {
    k.engine.RegisterAction(ActionLoadSecurityService, loadSecurityService)
    k.engine.RegisterAction(ActionLoadCommunicationService, loadCommunicationService)
    k.engine.RegisterAction(ActionLoadObservabilityService, loadObservabilityService)
    k.engine.RegisterAction(ActionLoadLifecycleService, loadLifecycleService)
    k.engine.RegisterAction(ActionSignalKernelReady, signalKernelReady)
}
```

---

## Part 3: TDD Test Plan (Strict TDD Workflow)

### Test Order Rationale
Tests ordered from simplest (action signature) → integration (kernel registration) → end-to-end (full bootstrap flow).

---

### Test 1: Action Signature Stub

**Test Name**: `TestLoadSecurityService_ActionSignature`

**What it verifies**: Action function exists with correct signature and returns `NotImplementedError`

**Expected Failure**: Function doesn't exist yet

**Minimal Implementation**:
```go
func TestLoadSecurityService_ActionSignature(t *testing.T) {
    var fn ActionFn
    fn = loadSecurityService // Must compile with ActionFn signature
    
    // Call with stubs - should return NotImplementedError
    err := fn(RuntimeContext{}, &mockAppCtx{}, Event{})
    if !errors.Is(err, ErrNotImplemented) {
        t.Errorf("expected NotImplementedError, got: %v", err)
    }
}
```

**File**: `pkg/bootstrap/actions_test.go` (new)

---

### Test 2: Action Registration

**Test Name**: `TestKernel_RegisterBootstrapActions`

**What it verifies**: Kernel registers all 5 bootstrap actions with engine

**Expected Failure**: `registerBootstrapActions()` doesn't exist

**Minimal Implementation**:
```go
func TestKernel_RegisterBootstrapActions(t *testing.T) {
    engine := statechart.NewEngine()
    kernel := kernel.NewWithEngine(engine)
    
    // Call registration
    kernel.RegisterBootstrapActions()
    
    // Verify actions are registered (via internal check or behavior)
    // Note: Engine doesn't expose action list, so test via Spawn+Execute
}
```

**File**: `pkg/kernel/kernel_test.go` (new or existing)

---

### Test 3: Security Service Spawn

**Test Name**: `TestLoadSecurityService_SpawnsRuntime`

**What it verifies**: Action spawns a transient runtime for sys:security

**Expected Failure**: Action doesn't actually spawn anything

**Minimal Implementation**:
```go
func TestLoadSecurityService_SpawnsRuntime(t *testing.T) {
    engine := statechart.NewEngine()
    
    // Track spawns via TraceCallback
    var spawnedRTIDs []statechart.RuntimeID
    engine.SetTraceCallback(&traceCallback{
        OnSpawn: func(rtID statechart.RuntimeID) { spawnedRTIDs = append(spawnedRTIDs, rtID) },
    })
    
    // Call action
    err := loadSecurityService(RuntimeContext{RuntimeID: "bootstrap-1"}, &mockAppCtx{}, Event{})
    
    // Verify spawn occurred
    if len(spawnedRTIDs) != 1 {
        t.Errorf("expected 1 spawn, got %d", len(spawnedRTIDs))
    }
}
```

**File**: `pkg/bootstrap/actions_test.go`

---

### Test 4: Security Service Starts

**Test Name**: `TestLoadSecurityService_StartsRuntime`

**What it verifies**: Action starts the spawned runtime (CmdStart)

**Expected Failure**: Action doesn't call Control(CmdStart)

**Minimal Implementation**:
```go
func TestLoadSecurityService_StartsRuntime(t *testing.T) {
    engine := statechart.NewEngine()
    
    // Track control commands
    var controlCmds []statechart.ControlCmd
    // ... use TraceCallback or mock engine ...
    
    err := loadSecurityService(RuntimeContext{RuntimeID: "bootstrap-1"}, &mockAppCtx{}, Event{})
    
    // Verify CmdStart was called
    if !contains(controlCmds, statechart.CmdStart) {
        t.Error("expected CmdStart to be called")
    }
}
```

**File**: `pkg/bootstrap/actions_test.go`

---

### Test 5: Security Service Dispatches Ready Event

**Test Name**: `TestLoadSecurityService_DispatchesReadyEvent`

**What it verifies**: Action dispatches SECURITY_READY event to bootstrap parent

**Expected Failure**: Action doesn't dispatch event

**Minimal Implementation**:
```go
func TestLoadSecurityService_DispatchesReadyEvent(t *testing.T) {
    engine := statechart.NewEngine()
    
    // Track dispatched events
    var dispatchedEvents []statechart.Event
    
    err := loadSecurityService(
        RuntimeContext{RuntimeID: "bootstrap-1"}, 
        &mockAppCtx{}, 
        Event{},
    )
    
    // Verify SECURITY_READY was dispatched
    readyEvent := findEvent(dispatchedEvents, "SECURITY_READY")
    if readyEvent == nil {
        t.Error("expected SECURITY_READY event to be dispatched")
    }
}
```

**File**: `pkg/bootstrap/actions_test.go`

---

### Test 6: Security Service Stores Runtime ID

**Test Name**: `TestLoadSecurityService_StoresRuntimeID`

**What it verifies**: Action stores spawned runtime ID in appCtx

**Expected Failure**: Action doesn't write to appCtx

**Minimal Implementation**:
```go
func TestLoadSecurityService_StoresRuntimeID(t *testing.T) {
    engine := statechart.NewEngine()
    mockCtx := &mockAppCtx{}
    
    err := loadSecurityService(
        RuntimeContext{RuntimeID: "bootstrap-1"}, 
        mockCtx, 
        Event{},
    )
    
    // Verify runtime ID was stored
    rtID, _, err := mockCtx.Get("bootstrap:security:runtimeID", "sys:bootstrap")
    if err != nil {
        t.Errorf("expected runtime ID to be stored, got error: %v", err)
    }
    if rtID == "" {
        t.Error("expected non-empty runtime ID")
    }
}
```

**File**: `pkg/bootstrap/actions_test.go`

---

### Test 7: Communication Service Action

**Test Name**: `TestLoadCommunicationService_SpawnsAndStarts`

**What it verifies**: Communication action spawns, starts, and signals ready

**Expected Failure**: Action doesn't exist

**Minimal Implementation**:
```go
func TestLoadCommunicationService_SpawnsAndStarts(t *testing.T) {
    engine := statechart.NewEngine()
    mockCtx := &mockAppCtx{}
    
    // Pre-populate security RTID (dependency)
    mockCtx.Set("bootstrap:security:runtimeID", "rt-100", nil, "sys:bootstrap")
    
    err := loadCommunicationService(
        RuntimeContext{RuntimeID: "bootstrap-1"}, 
        mockCtx, 
        Event{},
    )
    
    // Verify spawn, start, and COMMUNICATION_READY event
    // ... similar to security tests ...
}
```

**File**: `pkg/bootstrap/actions_test.go`

---

### Test 8: Observability Service Action

**Test Name**: `TestLoadObservabilityService_SpawnsAndStarts`

**What it verifies**: Observability action spawns, starts, and signals ready

**Minimal Implementation**: Same pattern as communication test

**File**: `pkg/bootstrap/actions_test.go`

---

### Test 9: Lifecycle Service Action

**Test Name**: `TestLoadLifecycleService_SpawnsAndStarts`

**What it verifies**: Lifecycle action spawns, starts, and signals ready

**Minimal Implementation**: Same pattern as communication test

**File**: `pkg/bootstrap/actions_test.go`

---

### Test 10: Signal Kernel Ready Action

**Test Name**: `TestSignalKernelReady_AggregatesServices`

**What it verifies**: Action reads all service RTIDs and stores loaded services list

**Expected Failure**: Action doesn't exist

**Minimal Implementation**:
```go
func TestSignalKernelReady_AggregatesServices(t *testing.T) {
    mockCtx := &mockAppCtx{}
    
    // Pre-populate all service RTIDs
    mockCtx.Set("bootstrap:security:runtimeID", "rt-1", nil, "sys:bootstrap")
    mockCtx.Set("bootstrap:communication:runtimeID", "rt-2", nil, "sys:bootstrap")
    mockCtx.Set("bootstrap:observability:runtimeID", "rt-3", nil, "sys:bootstrap")
    mockCtx.Set("bootstrap:lifecycle:runtimeID", "rt-4", nil, "sys:bootstrap")
    
    err := signalKernelReady(
        RuntimeContext{RuntimeID: "bootstrap-1"}, 
        mockCtx, 
        Event{},
    )
    
    // Verify loaded services list was stored
    services, _, err := mockCtx.Get("bootstrap:loaded:services", "sys:bootstrap")
    if err != nil {
        t.Errorf("expected loaded services to be stored: %v", err)
    }
    
    serviceList := services.([]string)
    if len(serviceList) != 4 {
        t.Errorf("expected 4 services, got %d", len(serviceList))
    }
}
```

**File**: `pkg/bootstrap/actions_test.go`

---

### Test 11: Kernel Integration Test

**Test Name**: `TestKernel_Start_RegistersAndExecutesActions`

**What it verifies**: Full kernel start flow registers and executes bootstrap actions

**Expected Failure**: Kernel doesn't call registerBootstrapActions()

**Minimal Implementation**:
```go
func TestKernel_Start_RegistersAndExecutesActions(t *testing.T) {
    engine := statechart.NewEngine()
    kernel := kernel.NewWithEngine(engine)
    
    // Start kernel in goroutine
    done := make(chan error)
    go func() {
        done <- kernel.Start(context.Background())
    }()
    
    // Wait for bootstrap to complete
    select {
    case err := <-done:
        if err != nil {
            t.Logf("kernel error (may be expected): %v", err)
        }
    case <-time.After(5 * time.Second):
        t.Fatal("kernel start timed out")
    }
    
    // Verify all services were spawned
    runtimes := kernel.GetRuntimes()
    if len(runtimes) != 4 {
        t.Errorf("expected 4 service runtimes, got %d", len(runtimes))
    }
}
```

**File**: `pkg/kernel/kernel_test.go`

---

### Test 12: Bootstrap Chart Integration

**Test Name**: `TestBootstrapChart_ActionsExecuteInOrder`

**What it verifies**: Bootstrap chart executes actions in correct order when run through engine

**Expected Failure**: Actions not registered or chart not properly wired

**Minimal Implementation**:
```go
func TestBootstrapChart_ActionsExecuteInOrder(t *testing.T) {
    engine := statechart.NewEngine()
    
    // Register bootstrap actions
    registerBootstrapActions(engine)
    
    // Load and spawn bootstrap chart
    def, err := LoadBootstrapChart()
    if err != nil {
        t.Fatalf("failed to load bootstrap chart: %v", err)
    }
    
    mockCtx := &mockAppCtx{}
    rtID, err := engine.Spawn(def, mockCtx)
    if err != nil {
        t.Fatalf("failed to spawn: %v", err)
    }
    
    // Start bootstrap
    err = engine.Control(rtID, statechart.CmdStart)
    if err != nil {
        t.Fatalf("failed to start: %v", err)
    }
    
    // Dispatch START_BOOTSTRAP
    err = engine.Dispatch(rtID, statechart.Event{Type: "START_BOOTSTRAP"})
    if err != nil {
        t.Fatalf("failed to dispatch: %v", err)
    }
    
    // Wait for completion
    time.Sleep(1 * time.Second)
    
    // Verify all services loaded
    services, _, _ := mockCtx.Get("bootstrap:loaded:services", "sys:bootstrap")
    serviceList := services.([]string)
    if len(serviceList) != 4 {
        t.Errorf("expected 4 services, got %d: %v", len(serviceList), services)
    }
}
```

**File**: `pkg/bootstrap/actions_test.go`

---

## Part 4: Implementation Files Structure

```
pkg/bootstrap/
├── actions.go          # NEW: All bootstrap action implementations
├── actions_test.go     # NEW: All action tests
├── chart.go            # EXISTING: Bootstrap chart YAML
├── chart_test.go       # EXISTING
├── sequence.go         # EXISTING
└── sequence_test.go    # EXISTING

pkg/kernel/
├── kernel.go           # EXISTING: Add registerBootstrapActions()
└── kernel_test.go      # EXISTING or NEW: Add integration tests
```

---

## Part 5: Key Design Decisions

### 5.1 Transient vs Independent Runtimes

**Decision**: Use `SpawnTransient` for service runtimes

**Rationale**: 
- Services are children of bootstrap
- When bootstrap completes, services should continue running (not be stopped)
- Actually, reconsider: services should persist after bootstrap. Use `Spawn` not `SpawnTransient`.

**Correction**: Use `engine.Spawn()` for services, not `SpawnTransient()`. Services are independent runtimes that persist after bootstrap completes.

### 5.2 Synchronous vs Asynchronous Service Loading

**Decision**: Actions are synchronous but spawn services asynchronously

**Rationale**:
- Action executes synchronously (blocks state transition)
- Service spawning is synchronous (Spawn + Start)
- Service READY signal is asynchronous (service emits event when ready)
- Bootstrap waits for READY event before transitioning

### 5.3 Error Handling

**Decision**: Actions return errors that block state transitions

**Rationale**:
- If service fails to spawn, bootstrap should fail
- Error propagates to kernel Start() which returns to caller
- No partial bootstrap states

---

## Part 6: Summary of Tests (In Order)

| # | Test Name | File | Verifies |
|---|-----------|------|----------|
| 1 | TestLoadSecurityService_ActionSignature | actions_test.go | Action exists with correct signature |
| 2 | TestKernel_RegisterBootstrapActions | kernel_test.go | Kernel registers all 5 actions |
| 3 | TestLoadSecurityService_SpawnsRuntime | actions_test.go | Security spawns runtime |
| 4 | TestLoadSecurityService_StartsRuntime | actions_test.go | Security starts runtime |
| 5 | TestLoadSecurityService_DispatchesReadyEvent | actions_test.go | Security dispatches READY |
| 6 | TestLoadSecurityService_StoresRuntimeID | actions_test.go | Security stores RTID in appCtx |
| 7 | TestLoadCommunicationService_SpawnsAndStarts | actions_test.go | Communication action works |
| 8 | TestLoadObservabilityService_SpawnsAndStarts | actions_test.go | Observability action works |
| 9 | TestLoadLifecycleService_SpawnsAndStarts | actions_test.go | Lifecycle action works |
| 10 | TestSignalKernelReady_AggregatesServices | actions_test.go | KernelReady aggregates services |
| 11 | TestKernel_Start_RegistersAndExecutesActions | kernel_test.go | Kernel integration |
| 12 | TestBootstrapChart_ActionsExecuteInOrder | actions_test.go | Full chart integration |

---

## Part 7: Stub Files to Create First

Per TDD workflow, create stubs before tests:

```go
// pkg/bootstrap/actions.go
package bootstrap

import (
    "github.com/maelstrom/v3/pkg/statechart"
)

const (
    ActionLoadSecurityService     = "loadSecurityService"
    ActionLoadCommunicationService = "loadCommunicationService"
    ActionLoadObservabilityService = "loadObservabilityService"
    ActionLoadLifecycleService    = "loadLifecycleService"
    ActionSignalKernelReady       = "signalKernelReady"
)

// Stub implementations - raise NotImplementedError
func loadSecurityService(rc statechart.RuntimeContext, ac statechart.ApplicationContext, ev statechart.Event) error {
    return nil // TODO: implement
}

func loadCommunicationService(rc statechart.RuntimeContext, ac statechart.ApplicationContext, ev statechart.Event) error {
    return nil // TODO: implement
}

func loadObservabilityService(rc statechart.RuntimeContext, ac statechart.ApplicationContext, ev statechart.Event) error {
    return nil // TODO: implement
}

func loadLifecycleService(rc statechart.RuntimeContext, ac statechart.ApplicationContext, ev statechart.Event) error {
    return nil // TODO: implement
}

func signalKernelReady(rc statechart.RuntimeContext, ac statechart.ApplicationContext, ev statechart.Event) error {
    return nil // TODO: implement
}

// Helper to register all actions
func RegisterBootstrapActions(engine statechart.Library) {
    engine.RegisterAction(ActionLoadSecurityService, loadSecurityService)
    engine.RegisterAction(ActionLoadCommunicationService, loadCommunicationService)
    engine.RegisterAction(ActionLoadObservabilityService, loadObservabilityService)
    engine.RegisterAction(ActionLoadLifecycleService, loadLifecycleService)
    engine.RegisterAction(ActionSignalKernelReady, signalKernelReady)
}
```

---

**End of Phase 1.4 TDD Plan**