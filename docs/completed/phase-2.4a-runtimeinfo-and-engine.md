# Phase 2.4A: RuntimeInfo Type and Engine Wiring

## Goal
Add RuntimeInfo type definition and wire engine into LifecycleService without breaking existing tests, following strict TDD workflow.

## Scope
- Create `RuntimeInfo` type in `pkg/services/lifecycle/types.go`
- Add `engine` field to `LifecycleService` struct
- Add `NewLifecycleService(engine)` constructor (keep backward-compatible no-arg version)
- Add runtime tracking map to struct
- 2 tests, 2 commits
- **NO breaking changes** - existing tests must pass

## Current Status

### Implementation Status
| Item | Status | Notes |
|------|--------|-------|
| `RuntimeInfo` type | ❌ Missing | Needed for `List()` return type |
| Engine field in LifecycleService | ❌ Missing | Needed for `Control()` forwarding |
| NewLifecycleService(engine) | ❌ Missing | Needed for dependency injection |
| Runtime tracking map | ❌ Missing | Needed for `List()` implementation |

### Files Status
| File | Status |
|------|--------|
| `pkg/services/lifecycle/types.go` | ❌ MISSING - create with RuntimeInfo |
| `pkg/services/lifecycle/service.go` | ⚠️ Partial - add engine field, constructor, tracking map |

## Required Implementation

### RuntimeInfo Type
```go
// pkg/services/lifecycle/types.go
type RuntimeInfo struct {
    ID           string
    DefinitionID string
    Boundary     mail.BoundaryType
    ActiveStates []string
    IsRunning    bool
}
```

### Updated LifecycleService Struct
```go
// pkg/services/lifecycle/service.go
type LifecycleService struct {
    mu         sync.Mutex
    engine     statechart.Library
    runtimes   map[statechart.RuntimeID]RuntimeInfo
}
```

### New Constructor
```go
// pkg/services/lifecycle/service.go
func NewLifecycleService(engine statechart.Library) *LifecycleService {
    return &LifecycleService{
        engine:   engine,
        runtimes: make(map[statechart.RuntimeID]RuntimeInfo),
    }
}

// Backward-compatible constructor (for existing tests)
func NewLifecycleService() *LifecycleService {
    return &LifecycleService{
        runtimes: make(map[statechart.RuntimeID]RuntimeInfo),
    }
}
```

## Tests to Write (2 tests, 2 commits)

### Test 1: TestLifecycle_RuntimeInfoHasRequiredFields
```go
func TestLifecycle_RuntimeInfoHasRequiredFields(t *testing.T)
```
**Acceptance Criteria:**
- RuntimeInfo type exists in `pkg/services/lifecycle/types.go`
- Has ID (string), DefinitionID (string), Boundary (mail.BoundaryType), ActiveStates ([]string), IsRunning (bool) fields
- Can instantiate with: `RuntimeInfo{ID: "test", DefinitionID: "def1", Boundary: mail.InnerBoundary}`

### Test 2: TestLifecycleService_NewWithEngineReturnsNonNil
```go
func TestLifecycleService_NewWithEngineReturnsNonNil(t *testing.T)
```
**Acceptance Criteria:**
- `NewLifecycleService(engine)` accepts statechart.Library parameter
- Returns non-nil LifecycleService
- Service has initialized runtimes map (not nil)
- **Existing test `TestLifecycleService_NewLifecycleServiceReturnsNonNil` still passes**

## Dependencies

### Test Dependencies
```
Test 1 (RuntimeInfo type) → Independent
Test 2 (New constructor) → Depends on Test 1 (uses RuntimeInfo type in struct)
```

### Phase Dependencies
- **Phase 2.1A** should have created RuntimeInfo type (but it didn't - this phase completes that work)
- **Phase 2.4B** depends on this phase completing first (needs RuntimeInfo and engine wiring)

## Files to Create/Modify

| File | Operation | Changes |
|------|-----------|---------|
| `pkg/services/lifecycle/types.go` | CREATE | Create file with RuntimeInfo type |
| `pkg/services/lifecycle/service.go` | MODIFY | Add engine field, runtimes map, NewLifecycleService(engine) constructor |

## Implementation Order (TDD)

1. **Test 1**: Write test → verify RED → Create lifecycle/types.go with RuntimeInfo → verify GREEN → commit
2. **Test 2**: Write test → verify RED → Add engine field, runtimes map, NewLifecycleService(engine) to service.go → verify GREEN → commit

**Total: 2 tests, 2 commits**

## Deliverables
- ✅ `RuntimeInfo` type in `pkg/services/lifecycle/types.go`
- ✅ `engine` field in `LifecycleService` struct
- ✅ `runtimes` map in `LifecycleService` struct
- ✅ `NewLifecycleService(engine)` constructor
- ✅ Backward-compatible `NewLifecycleService()` constructor preserved
- ✅ 2 commits (one per test per TDD workflow)
- ✅ NO breaking changes - existing tests still pass