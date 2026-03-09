# Phase 2.4B: List Returns RuntimeInfo

## Goal
Update List() method to return []RuntimeInfo and implement runtime tracking in Spawn(), following strict TDD workflow.

## Scope
- Update existing test that expects []string to expect []RuntimeInfo
- Implement List() to return []RuntimeInfo
- Implement runtime tracking in Spawn()
- 3 tests, 3 commits
- **BREAKING CHANGE** - List() signature changes from []string to []RuntimeInfo

## Current Status

### Implementation Status
| Item | Status | Notes |
|------|--------|-------|
| List() returns []RuntimeInfo | ❌ Missing | Currently returns []string |
| Runtime tracking in Spawn() | ❌ Missing | Spawn() doesn't track runtimes |
| List() returns tracked runtimes | ❌ Missing | No runtime data to return |

### Files Status
| File | Status |
|------|--------|
| `pkg/services/lifecycle/service.go` | ⚠️ Partial - update List() signature, implement tracking |
| `pkg/services/lifecycle/service_test.go` | ⚠️ Partial - update existing List test |

## Required Implementation

### Updated List Method
```go
// pkg/services/lifecycle/service.go
func (l *LifecycleService) List() ([]RuntimeInfo, error) {
    l.mu.Lock()
    defer l.mu.Unlock()
    
    result := make([]RuntimeInfo, 0, len(l.runtimes))
    for _, info := range l.runtimes {
        result = append(result, info)
    }
    return result, nil
}
```

### Updated Spawn Method with Tracking
```go
// pkg/services/lifecycle/service.go
func (l *LifecycleService) Spawn(def statechart.ChartDefinition) (statechart.RuntimeID, error) {
    id, err := l.engine.Spawn(def, nil)
    if err != nil {
        return "", err
    }
    
    l.mu.Lock()
    l.runtimes[id] = RuntimeInfo{
        ID:           string(id),
        DefinitionID: def.ID,
        Boundary:     mail.InnerBoundary,
        ActiveStates: []string{def.InitialState},
        IsRunning:    false,
    }
    l.mu.Unlock()
    
    return id, nil
}
```

## Tests to Write (3 tests, 3 commits)

### Test 1: TestLifecycleService_ListReturnsRuntimeInfo
```go
func TestLifecycleService_ListReturnsRuntimeInfo(t *testing.T)
```
**Acceptance Criteria:**
- List() returns []RuntimeInfo (not []string)
- Each RuntimeInfo has ID, DefinitionID, Boundary fields populated
- Returns nil error

### Test 2: TestLifecycleService_ListEmptyWhenNoRuntimes
```go
func TestLifecycleService_ListEmptyWhenNoRuntimes(t *testing.T)
```
**Acceptance Criteria:**
- List() returns empty slice when no runtimes spawned
- Returns nil error
- Slice is not nil (len == 0)

### Test 3: TestLifecycleService_SpawnTracksRuntime
```go
func TestLifecycleService_SpawnTracksRuntime(t *testing.T)
```
**Acceptance Criteria:**
- Spawn() adds runtime to internal tracking map
- List() includes the spawned runtime
- RuntimeInfo has correct ID, DefinitionID, Boundary values

## Dependencies

### Test Dependencies
```
Test 1 (List returns RuntimeInfo) → Independent
Test 2 (List empty) → Independent
Test 3 (Spawn tracks) → Depends on Test 1 (List must return RuntimeInfo)
```

### Phase Dependencies
- **Phase 2.4A** must complete first (provides RuntimeInfo type and engine wiring)
- **Phase 2.4C** depends on this phase (Control() may need runtime tracking)

## Files to Create/Modify

| File | Operation | Changes |
|------|-----------|---------|
| `pkg/services/lifecycle/service.go` | MODIFY | Update List() signature, implement List(), update Spawn() with tracking |
| `pkg/services/lifecycle/service_test.go` | MODIFY | Update TestLifecycleService_ListReturnsNilSlice to expect []RuntimeInfo |

## Implementation Order (TDD)

1. **Test 1**: Write test → verify RED → Update List() to return []RuntimeInfo → verify GREEN → commit
2. **Test 2**: Write test → verify RED → Ensure List() returns empty slice when no runtimes → verify GREEN → commit
3. **Test 3**: Write test → verify RED → Add runtime tracking in Spawn() → verify GREEN → commit

**Total: 3 tests, 3 commits**

## Breaking Changes

### Existing Test Must Be Updated
```go
// OLD: TestLifecycleService_ListReturnsNilSlice
func TestLifecycleService_ListReturnsNilSlice(t *testing.T) {
    svc := NewLifecycleService()
    list, err := svc.List()  // Returns []string
    
    if err != nil {
        t.Errorf("Expected List to return nil error, got %v", err)
    }
    
    if list != nil {
        t.Error("Expected List to return nil slice")
    }
}

// NEW: TestLifecycleService_ListReturnsRuntimeInfo
func TestLifecycleService_ListReturnsRuntimeInfo(t *testing.T) {
    svc := NewLifecycleService()
    list, err := svc.List()  // Returns []RuntimeInfo
    
    if err != nil {
        t.Errorf("Expected List to return nil error, got %v", err)
    }
    
    if len(list) != 0 {
        t.Errorf("Expected empty slice, got %d items", len(list))
    }
}
```

## Deliverables
- ✅ `List()` returns []RuntimeInfo
- ✅ `Spawn()` tracks runtimes in internal map
- ✅ Updated existing test for new List() signature
- ✅ 3 commits (one per test per TDD workflow)
- ✅ Breaking change documented and test updated