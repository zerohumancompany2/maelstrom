# Phase 2.4: Lifecycle Service Enhancement

## Goal
Enhance Lifecycle Service with proper runtime listing and control methods.

## Scope
- Update `List()` to return `[]RuntimeInfo`
- Implement `Control()` method
- Track spawned runtimes
- 7 tests, 7 commits

## Current Status

### Implementation Status
| Test | Status | Notes |
|------|--------|-------|
| TestLifecycleService_ListReturnsRuntimeInfo | ❌ Missing | Returns []string instead of []RuntimeInfo |
| TestLifecycleService_ControlStart | ❌ Missing | Method doesn't exist |
| TestLifecycleService_ControlStop | ❌ Missing | Method doesn't exist |
| TestLifecycleService_SpawnTracksRuntime | ❌ Missing | Spawn doesn't track |

### Files Status
| File | Status |
|------|--------|
| `pkg/services/lifecycle/service.go` | ⚠️ Partial - update List() signature |

## Required Implementation

### Updated List Method
```go
// pkg/services/lifecycle/service.go
func (l *LifecycleService) List() ([]RuntimeInfo, error) {
    // Return list of runtime info
}
```

### Control Method
```go
// pkg/services/lifecycle/service.go
func (l *LifecycleService) Control(id statechart.RuntimeID, cmd statechart.ControlCmd) error {
    // Forward control command to engine
}
```

## Tests to Write (7 tests, 7 commits)

### Test 1: TestLifecycleService_ListReturnsRuntimeInfo
```go
func TestLifecycleService_ListReturnsRuntimeInfo(t *testing.T)
```
**Acceptance Criteria:**
- List() returns []RuntimeInfo
- Each RuntimeInfo has ID, DefinitionID, Boundary

### Test 2: TestLifecycleService_ListEmptyWhenNoRuntimes
```go
func TestLifecycleService_ListEmptyWhenNoRuntimes(t *testing.T)
```
**Acceptance Criteria:**
- List() returns empty slice when no runtimes
- No error returned

### Test 3: TestLifecycleService_SpawnTracksRuntime
```go
func TestLifecycleService_SpawnTracksRuntime(t *testing.T)
```
**Acceptance Criteria:**
- Spawn() adds runtime to internal tracking
- List() includes spawned runtime

### Test 4: TestLifecycleService_ControlStart
```go
func TestLifecycleService_ControlStart(t *testing.T)
```
**Acceptance Criteria:**
- Control() with CmdStart succeeds
- Returns nil error

### Test 5: TestLifecycleService_ControlStop
```go
func TestLifecycleService_ControlStop(t *testing.T)
```
**Acceptance Criteria:**
- Control() with CmdStop succeeds
- Returns nil error

### Test 6: TestLifecycleService_ControlNotFoundReturnsError
```go
func TestLifecycleService_ControlNotFoundReturnsError(t *testing.T)
```
**Acceptance Criteria:**
- Control() with non-existent ID returns error
- Error indicates runtime not found

### Test 7: TestLifecycleService_RuntimeInfoHasActiveStates
```go
func TestLifecycleService_RuntimeInfoHasActiveStates(t *testing.T)
```
**Acceptance Criteria:**
- RuntimeInfo.ActiveStates is populated
- RuntimeInfo.IsRunning is true for active runtimes

## Dependencies

### Test Dependencies
```
Test 1 → Test 2 (List behavior)
Test 3 (Spawn tracking)
Test 4 → Test 5, Test 6 (Control behavior)
Test 7 (RuntimeInfo details)
```

### Phase Dependencies
- **Phase 2.1** - Type definitions (RuntimeInfo must exist)

## Files to Create/Modify

| File | Operation | Changes |
|------|-----------|---------|
| `pkg/services/lifecycle/service.go` | MODIFY | Update List(), add Control(), add runtime tracking |

## Implementation Order (TDD)

1. **Test 1**: Write test → verify RED → Update List() to return []RuntimeInfo → verify GREEN → commit
2. **Test 2**: Write test → verify RED → Handle empty runtime list → verify GREEN → commit
3. **Test 3**: Write test → verify RED → Add runtime tracking in Spawn() → verify GREEN → commit
4. **Test 4**: Write test → verify RED → Implement Control() with CmdStart → verify GREEN → commit
5. **Test 5**: Write test → verify RED → Implement Control() with CmdStop → verify GREEN → commit
6. **Test 6**: Write test → verify RED → Add error handling for missing runtime → verify GREEN → commit
7. **Test 7**: Write test → verify RED → Populate ActiveStates and IsRunning → verify GREEN → commit

**Total: 7 tests, 7 commits**

## Deliverables
- ✅ `List()` returns []RuntimeInfo
- ✅ `Control()` implemented
- ✅ Runtime tracking in Spawn()
- ✅ 7 commits (one per test per TDD workflow)