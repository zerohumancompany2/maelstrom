# Phase 2.5B: ReportTaints and PrepareContextForBoundary

## Goal
Implement ReportTaints stub and PrepareContextForBoundary no-op in Security Service, following strict TDD workflow with consolidated tests.

## Scope
- Implement `ReportTaints()` stub (combines original tests 5+6)
- Implement `PrepareContextForBoundary()` no-op (combines original tests 7+8+9)
- 2 tests, 2 commits
- **NO breaking changes** - existing tests must pass

## Current Status

### Implementation Status
| Item | Status | Notes |
|------|--------|-------|
| `ReportTaints()` | ❌ Missing | Method doesn't exist |
| `PrepareContextForBoundary()` | ❌ Missing | Method doesn't exist |

### Files Status
| File | Status |
|------|--------|
| `pkg/services/security/service.go` | ⚠️ Partial - add new methods |

## Required Implementation

### ReportTaints Method
```go
// pkg/services/security/service.go
func (s *SecurityService) ReportTaints(runtimeId string) (security.TaintMap, error) {
    // Phase 1: return empty
    return security.TaintMap{}, nil
}
```

### PrepareContextForBoundary Method
```go
// pkg/services/security/service.go
func (s *SecurityService) PrepareContextForBoundary(runtimeId string, boundary mail.BoundaryType) error {
    // Phase 1: no-op
    return nil
}
```

## Tests to Write (2 tests, 2 commits)

### Test 1: TestSecurityService_ReportTaintsReturnsEmptyMap
```go
func TestSecurityService_ReportTaintsReturnsEmptyMap(t *testing.T)
```
**Acceptance Criteria:**
- ReportTaints() returns TaintMap
- TaintMap is empty (Phase 1)
- Accepts runtimeId parameter
- Returns nil error

**Consolidates original tests:**
- Test 5: TestSecurityService_ReportTaintsReturnsEmptyMap
- Test 6: TestSecurityService_ReportTaintsWithRuntimeID

### Test 2: TestSecurityService_PrepareContextForBoundaryNoOp
```go
func TestSecurityService_PrepareContextForBoundaryNoOp(t *testing.T)
```
**Acceptance Criteria:**
- PrepareContextForBoundary() returns nil
- Phase 1: no-op
- Works with InnerBoundary type
- Works with OuterBoundary type
- Accepts runtimeId parameter

**Consolidates original tests:**
- Test 7: TestSecurityService_PrepareContextForBoundaryNoOp
- Test 8: TestSecurityService_PrepareContextForBoundaryWithInnerBoundary
- Test 9: TestSecurityService_PrepareContextForBoundaryWithOuterBoundary

## Dependencies

### Test Dependencies
```
Test 1 (ReportTaints) → Independent
Test 2 (PrepareContextForBoundary) → Independent
```

### Phase Dependencies
- **Phase 2.1** - Type definitions (security.TaintMap, mail.BoundaryType exist)
- **Phase 2.5A** must complete first (implements ValidateAndSanitize and TaintPropagate)

## Files to Create/Modify

| File | Operation | Changes |
|------|-----------|---------|
| `pkg/services/security/service.go` | MODIFY | Add ReportTaints(), PrepareContextForBoundary() |

## Implementation Order (TDD)

1. **Test 1**: Write test → verify RED → Implement ReportTaints() stub → verify GREEN → commit
2. **Test 2**: Write test → verify RED → Implement PrepareContextForBoundary() no-op → verify GREEN → commit

**Total: 2 tests, 2 commits**

## Deliverables
- ✅ `ReportTaints()` implemented (stub)
- ✅ `PrepareContextForBoundary()` implemented (no-op)
- ✅ 2 commits (one per test per TDD workflow)
- ✅ NO breaking changes - existing tests still pass