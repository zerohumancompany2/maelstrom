# Phase 1.2.3: Bootstrap Sequence Integration

## Goal
Wire the bootstrap chart to the kernel so actions can register services and handle errors properly.

## Scope
- Add kernel reference to bootstrap sequence
- Add `RegisterService()` method to kernel
- Wire actions to receive kernel reference
- Implement error handling with `failed` state transitions
- Create 4 integration tests

## Current Status

### Current Bootstrap Sequence Flow
```
Kernel.Start()
  ├── Loads bootstrap chart (pkg/bootstrap/chart.go)
  ├── Spawns bootstrap runtime via statechart.Library.Spawn()
  ├── Creates bootstrap.Sequence (pkg/bootstrap/sequence.go)
  │   ├── Sets OnStateEnter callback → Kernel.onBootstrapStateEnter()
  │   └── Sets OnComplete callback → Kernel.onBootstrapComplete()
  └── Starts sequence → transitions through states:
       initializing → security → communication → observability → lifecycle → handoff → complete
```

### Missing Integration Points

#### 1. No Kernel Reference Passed to Sequence
**Spec Requirement:** Actions should call `kernel.RegisterService()`
**Current State:** `bootstrap.Sequence` has no reference to `Kernel`
**File:** `pkg/bootstrap/sequence.go` - `Sequence` struct lacks `kernel *kernel.Kernel` field

#### 2. No Service Registration Method on Kernel
**Spec Requirement:** `Kernel` should have `RegisterService()` method
**Current State:** `Kernel` has `runtimes` map but no public `RegisterService()` method
**File:** `pkg/kernel/kernel.go` - missing `RegisterService(name string, rt *runtime.ChartRuntime)` method

#### 3. No Action Implementation Wiring
**Spec Requirement:** Entry actions should be registered with the statechart engine
**Current State:** Bootstrap YAML references actions but they are never registered with `statechart.Library.RegisterAction()`
**Files:** 
- `pkg/bootstrap/chart.go` - actions defined in YAML but not implemented
- `pkg/kernel/kernel.go` - no action registration in `Start()`

#### 4. No Error Handling / Failed State
**Spec Requirement:** Bootstrap chart should have `failed` state for error transitions
**Current State:** Bootstrap YAML has no `failed` state; no error transitions defined
**File:** `pkg/bootstrap/chart.go` - missing `failed` state and error transitions

#### 5. Services Not Actually Spawned
**Spec Requirement:** Each state's entry action spawns a core service chart
**Current State:** `onBootstrapStateEnter()` only logs and fires simulated READY events
**File:** `pkg/kernel/kernel.go` - `onBootstrapStateEnter()` is a stub

## Required Changes

### 1. Add Kernel Reference to Sequence
```go
// pkg/bootstrap/sequence.go
type Sequence struct {
    kernel       *kernel.Kernel  // ADD THIS
    library      *statechart.Library
    onStateEnter func(state string)
    onComplete   func()
    // ...
}

func NewSequence(kernel *kernel.Kernel, library *statechart.Library) *Sequence {
    return &Sequence{
        kernel:  kernel,  // PASS KERNEL
        library: library,
        // ...
    }
}
```

### 2. Add RegisterService Method to Kernel
```go
// pkg/kernel/kernel.go
func (k *Kernel) RegisterService(name string, rt *runtime.ChartRuntime) error {
    k.mu.Lock()
    defer k.mu.Unlock()
    k.services[name] = rt  // Add services map field
    return nil
}
```

### 3. Wire Actions to Receive Kernel Reference
```go
// pkg/bootstrap/actions.go
var kernelRef *kernel.Kernel  // Global or passed via context

func spawnSecurityService(ctx context.Context, params map[string]interface{}) error {
    // Use kernelRef.RegisterService() here
    kernelRef.RegisterService("security", securityRuntime)
    return nil
}
```

### 4. Implement Error Transitions
Actions should return errors that trigger transitions to `failed` state:
```go
func spawnSecurityService(ctx context.Context, params map[string]interface{}) error {
    if err := validateParams(params); err != nil {
        return errors.New("security bootstrap failed: " + err.Error())
    }
    // ... spawn service
    return nil
}
```

## Tests to Write (4 tests)

### Test 1: Sequence passes kernel to actions
```go
func TestSequence_PassesKernelToActions()
```
**Acceptance Criteria:**
- `Sequence` struct has `kernel` field
- `NewSequence()` accepts and stores kernel reference
- Kernel reference is accessible in `onStateEnter` callback

### Test 2: Actions receive kernel reference
```go
func TestActions_ReceiveKernelReference()
```
**Acceptance Criteria:**
- Bootstrap actions can access kernel reference
- Actions can call `kernel.RegisterService()`
- Service registration succeeds

### Test 3: Error path triggers failed state
```go
func TestErrorPath_TriggersFailedState()
```
**Acceptance Criteria:**
- When action returns error, statechart transitions to `failed` state
- `failed` state entry action is executed
- Error is logged

### Test 4: Success path triggers ready state
```go
func TestSuccessPath_TriggersReadyState()
```
**Acceptance Criteria:**
- When all actions succeed, statechart transitions to `ready` state
- `ready` state entry action is executed
- `kernel_ready` event is emitted

## Files to Modify

| File | Changes Required |
|------|------------------|
| `pkg/bootstrap/sequence.go` | Add `kernel` field to `Sequence` struct; pass kernel to `onStateEnter` callback |
| `pkg/kernel/kernel.go` | Add `RegisterService()` method; add `services` map field; implement action registration in `Start()` |
| `pkg/bootstrap/actions.go` | Modify actions to use kernel reference for service registration |
| `pkg/bootstrap/chart.go` | Add `failed` state to YAML; add error transitions from each service state to `failed` |

## Implementation Order (TDD)

1. **Test 1**: Write test → verify RED → Add kernel field to Sequence → verify GREEN → commit
2. **Test 2**: Write test → verify RED → Add RegisterService() to Kernel → Wire actions to use it → verify GREEN → commit
3. **Test 3**: Write test → verify RED → Add failed state to YAML → Add error transitions → Make actions return errors → verify GREEN → commit
4. **Test 4**: Write test → verify RED → Ensure success path reaches ready state → verify GREEN → commit

## Dependencies
- Phase 1.2.1 (YAML must have `failed` and `ready` states)
- Phase 1.2.2 (Action stubs must exist)

## Deliverables
- Updated `pkg/bootstrap/sequence.go` with kernel reference
- Updated `pkg/kernel/kernel.go` with `RegisterService()` method
- Updated `pkg/bootstrap/actions.go` to use kernel reference
- Updated `pkg/bootstrap/chart.go` with error handling
- New integration tests (4 tests in existing or new test file)
- 4 commits (one per test per TDD workflow)