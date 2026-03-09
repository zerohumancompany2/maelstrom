# Phase 1.1B: Kernel Advanced

## Goal
Implement advanced kernel functionality with 4 tests building on Phase 1.1A.

## Scope
- Implement bootstrap completion verification
- Implement shutdown with context cancellation
- Implement service runtime ID retrieval
- Implement integration test with engine spawn and control
- Create 4 tests in `pkg/kernel/kernel_test.go`

## Current Status

### Implementation Status
| Test | Status | Notes |
|------|--------|-------|
| Test 8: IsBootstrapComplete_AfterBootstrap | ✅ Partially Implemented | `IsBootstrapComplete()` exists but needs verification |
| Test 11: Shutdown_ContextCancellation | ❌ Missing | Needs `Shutdown()` from 1.1A + context handling |
| Test 12: GetServiceRuntimeID_ReturnsCorrectID | ❌ Missing | Needs `services` map from 1.1A + getter method |
| Test 13: Start_WithEngine_SpawnsAndControls | ✅ Partially Implemented | `Start()` spawns but needs service tracking |

### Dependencies on Phase 1.1A
- Test 8 depends on Test 7 (bootstrap completion mechanism)
- Test 11 depends on Tests 9-10 (basic shutdown)
- Test 12 depends on Test 6 (services map)
- Test 13 depends on Tests 3-6 (actions registration, spawn, tracking)

## Required Implementation

### IsBootstrapComplete Enhancement
```go
// Already exists, verify it works correctly
func (k *Kernel) IsBootstrapComplete() bool {
    k.mu.RLock()
    defer k.mu.RUnlock()
    if k.sequence == nil {
        return false
    }
    return k.sequence.IsComplete()
}
```

### GetServiceRuntimeID
```go
func (k *Kernel) GetServiceRuntimeID(name string) (statechart.RuntimeID, bool) {
    k.mu.RLock()
    defer k.mu.RUnlock()
    id, ok := k.services[name]
    return id, ok
}
```

### Shutdown with Context Cancellation
```go
func (k *Kernel) Shutdown(ctx context.Context) error {
    if k.engine == nil {
        return nil
    }
    k.mu.RLock()
    services := make(map[string]statechart.RuntimeID, len(k.services))
    for name, id := range k.services {
        services[name] = id
    }
    k.mu.RUnlock()
    
    for name, id := range services {
        select {
        case <-ctx.Done():
            return ctx.Err()
        default:
        }
        if err := k.engine.Control(id, statechart.CmdStop); err != nil {
            log.Printf("[kernel] failed to stop %s: %v", name, err)
        }
    }
    return nil
}
```

### Start() Service Tracking
```go
// In Start() method, after spawn:
bootstrapRTID, err := k.engine.Spawn(def, nil)
if err != nil {
    return fmt.Errorf("failed to spawn bootstrap runtime: %w", err)
}
k.mu.Lock()
k.services["sys:bootstrap"] = bootstrapRTID
k.mu.Unlock()
```

## Tests to Write (4 tests, 4 commits)

### Test 8: IsBootstrapComplete_AfterBootstrap
```go
func TestKernel_IsBootstrapComplete_AfterBootstrap()
```
**Acceptance Criteria:**
1. Kernel starts with `New()` or `NewWithEngine()`
2. Bootstrap sequence completes (reaches "complete" state)
3. `IsBootstrapComplete()` returns `true` after completion
4. `IsBootstrapComplete()` returns `false` before completion

### Test 11: Shutdown_ContextCancellation
```go
func TestKernel_Shutdown_ContextCancellation()
```
**Acceptance Criteria:**
1. `Shutdown()` accepts context parameter
2. When context is cancelled during shutdown, method returns without panic
3. Services already stopped before cancellation are not restarted
4. Returns appropriate error or nil based on cancellation state

### Test 12: GetServiceRuntimeID_ReturnsCorrectID
```go
func TestKernel_GetServiceRuntimeID_ReturnsCorrectID()
```
**Acceptance Criteria:**
1. `GetServiceRuntimeID("sys:bootstrap")` returns correct RuntimeID after bootstrap spawns
2. Returns `(zero value, false)` for non-existent service names
3. Thread-safe concurrent access (uses RWMutex)
4. Returns consistent ID across multiple calls for same service

### Test 13: Start_WithEngine_SpawnsAndControls
```go
func TestKernel_Start_WithEngine_SpawnsAndControls()
```
**Acceptance Criteria:**
1. Kernel with engine spawns exactly one chart (bootstrap)
2. Kernel sends `CmdStart` control command to spawned runtime
3. Service ID is tracked in services map
4. No errors when engine is properly configured

## Dependencies

### Test Dependencies
```
Phase 1.1A Must Complete First:
├── Test 6 (services map) → Test 12 (GetServiceRuntimeID)
├── Test 7 (readyChan) → Test 8 (IsBootstrapComplete)
├── Test 9-10 (Shutdown) → Test 11 (Shutdown with context)
└── Test 3-6 (actions + spawn) → Test 13 (integration)
```

### Phase Dependencies
- **Phase 1.1A** - All 9 tests must complete first
- **Phase 1.2.1-1.2.3** - Bootstrap chart and sequence must be functional

## Files to Create/Modify

| File | Operation | Changes |
|------|-----------|---------|
| `pkg/kernel/kernel.go` | MODIFY | Add `GetServiceRuntimeID()`, enhance `Shutdown()` with context, update `Start()` to track services |
| `pkg/kernel/kernel_test.go` | MODIFY | Add 4 new tests (Tests 8, 11, 12, 13) |

## Implementation Order (TDD)

1. **Test 8**: Write test → verify RED → Verify/fix `IsBootstrapComplete()` → verify GREEN → commit
2. **Test 11**: Write test → verify RED → Enhance `Shutdown()` with context cancellation → verify GREEN → commit
3. **Test 12**: Write test → verify RED → Implement `GetServiceRuntimeID()` → verify GREEN → commit
4. **Test 13**: Write test → verify RED → Update `Start()` to track services → verify GREEN → commit

**Total: 4 tests, 4 commits**

## Deliverables
- Modified `pkg/kernel/kernel.go` with enhanced methods
- Updated `pkg/kernel/kernel_test.go` with 4 passing tests
- 4 commits (one per test per TDD workflow)

## Notes

### Test 14 - REMOVED
Test 14 (`TestKernel_FullBootstrapSequence`) is a placeholder test that violates TDD principles. It is **explicitly removed** from this phase and should not be implemented.

### Clarification on Test 7 Channel Mechanism
The channel/signal mechanism for bootstrap completion works as follows:
- `bootstrap.Sequence` has `onComplete` callback
- When sequence reaches "complete" state, callback fires
- `IsBootstrapComplete()` checks `seq.IsComplete()` which returns `true` when `currentState == "complete"`
- Phase 1.1A Test 7 must fix the blocking mechanism (currently waits on context cancellation, NOT bootstrap completion)