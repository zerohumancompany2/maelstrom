# Phase 1.6.1: Observability Infrastructure

## Goal
Create observability infrastructure for kernel and bootstrap sequence to enable integration testing.

## Scope
- Add `bootstrapRTID` field to Kernel struct
- Add `GetBootstrapRuntimeID()` getter
- Add `GetCurrentState()` getter
- Implement 3 tests in `pkg/kernel/kernel_test.go`

## Current Status

### Implementation Status
| Test | Status | Notes |
|------|--------|-------|
| Test 1: LoadBootstrapChart | ✅ Ready | `LoadBootstrapChart()` already exists |
| Test 2: SpawnsBootstrapRuntime | ❌ Missing | `bootstrapRTID` is local variable, not stored |
| Test 3: BootstrapSequenceStartsWithSecurity | ⚠️ Partial | `Sequence.CurrentState()` exists but not exposed |

### Files Status
| File | Status |
|------|--------|
| `pkg/bootstrap/chart.go` | ✅ Ready - `LoadBootstrapChart()` exists |
| `pkg/bootstrap/sequence.go` | ✅ Ready - `CurrentState()` exists |
| `pkg/kernel/kernel.go` | ⚠️ Partial - needs getters and field storage |
| `pkg/kernel/kernel_test.go` | ❌ Missing - needs 3 new tests |

## Required Implementation

### Kernel Struct Modification
```go
type Kernel struct {
    engine        statechart.Library
    factory       *runtime.Factory
    sequence      *bootstrap.Sequence
    bootstrapRTID statechart.RuntimeID  // ADD THIS
    runtimes      map[string]*runtime.ChartRuntime
    mu            sync.RWMutex
}
```

### Store bootstrapRTID in Start()
```go
// In Start() method, after spawn:
bootstrapRTID, err := k.engine.Spawn(def, nil)
if err != nil {
    return fmt.Errorf("failed to spawn bootstrap runtime: %w", err)
}
k.mu.Lock()
k.bootstrapRTID = bootstrapRTID
k.mu.Unlock()
log.Printf("[kernel] Spawning bootstrap runtime: %s", bootstrapRTID)
```

### GetBootstrapRuntimeID
```go
func (k *Kernel) GetBootstrapRuntimeID() statechart.RuntimeID {
    k.mu.RLock()
    defer k.mu.RUnlock()
    return k.bootstrapRTID
}
```

### GetCurrentState
```go
func (k *Kernel) GetCurrentState() string {
    k.mu.RLock()
    seq := k.sequence
    k.mu.RUnlock()
    if seq == nil {
        return ""
    }
    return seq.CurrentState()
}
```

## Tests to Write (3 tests, 3 commits)

### Test 1: LoadBootstrapChart
```go
func TestKernel_LoadsBootstrapChart()
```
**Acceptance Criteria:**
1. `err` from `LoadBootstrapChart()` is `nil`
2. `def.ID` is `"sys:bootstrap"`
3. `def.Version` is `"1.0.0"`
4. `def.InitialState` is `"initializing"`

### Test 2: SpawnsBootstrapRuntime
```go
func TestKernel_SpawnsBootstrapRuntime()
```
**Acceptance Criteria:**
1. `NewWithEngine(engine)` creates kernel with engine stored
2. `kernel.Start(ctx)` completes without immediate error
3. `GetBootstrapRuntimeID()` returns non-empty `statechart.RuntimeID`

### Test 3: BootstrapSequenceStartsWithSecurity
```go
func TestKernel_BootstrapSequenceStartsWithSecurity()
```
**Acceptance Criteria:**
1. After `Start()` begins, sequence is created
2. First state transition is `"security"` (from `"initializing"`)
3. `GetCurrentState()` returns `"security"` after first transition

## Dependencies

### Test Dependencies
```
Test 1 (LoadBootstrapChart) → No dependencies
    ↓
Test 2 (SpawnsBootstrapRuntime) → Needs Test 1
    ↓
Test 3 (StartsWithSecurity) → Needs Test 2
```

### Phase Dependencies
- **Phase 1.1A** - Kernel must have basic structure
- **Phase 1.2.1** - Bootstrap YAML must have correct states
- **Phase 1.2.3** - Bootstrap sequence must be functional

## Files to Create/Modify

| File | Operation | Changes |
|------|-----------|---------|
| `pkg/kernel/kernel.go` | MODIFY | Add `bootstrapRTID` field, store it in `Start()`, add `GetBootstrapRuntimeID()` and `GetCurrentState()` getters |
| `pkg/kernel/kernel_test.go` | MODIFY | Add 3 new tests |

## Implementation Order (TDD)

1. **Test 1**: Write test → verify RED → Verify `LoadBootstrapChart()` works → verify GREEN → commit
2. **Test 2**: Write test → verify RED → Add `bootstrapRTID` field → Store in `Start()` → Add getter → verify GREEN → commit
3. **Test 3**: Write test → verify RED → Add `GetCurrentState()` getter → verify GREEN → commit

**Total: 3 tests, 3 commits**

## Deliverables
- Modified `pkg/kernel/kernel.go` with observability field and getters
- Updated `pkg/kernel/kernel_test.go` with 3 passing tests
- 3 commits (one per test per TDD workflow)