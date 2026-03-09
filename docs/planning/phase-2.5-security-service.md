# Phase 2.5: Security Service Enhancement

## Goal
Enhance Security Service with taint propagation and boundary validation.

## Scope
- Implement `ValidateAndSanitize()` pass-through
- Implement `TaintPropagate()` stub
- Implement `ReportTaints()` stub
- Implement `PrepareContextForBoundary()` stub
- 10 tests, 10 commits

## Current Status

### Implementation Status
| Test | Status | Notes |
|------|--------|-------|
| TestSecurityService_ValidateAndSanitize | ❌ Missing | Method doesn't exist |
| TestSecurityService_TaintPropagate | ❌ Missing | Method doesn't exist |
| TestSecurityService_ReportTaints | ❌ Missing | Method doesn't exist |
| TestSecurityService_PrepareContextForBoundary | ❌ Missing | Method doesn't exist |

### Files Status
| File | Status |
|------|--------|
| `pkg/services/security/service.go` | ⚠️ Partial - add new methods |

## Required Implementation

### ValidateAndSanitize Method
```go
// pkg/services/security/service.go
func (s *SecurityService) ValidateAndSanitize(mail mail.Mail, sourceBoundary, targetBoundary mail.BoundaryType) (mail.Mail, error) {
    // Phase 1: pass-through
    return mail, nil
}
```

### TaintPropagate Method
```go
// pkg/services/security/service.go
func (s *SecurityService) TaintPropagate(obj any, newTaints []string) (any, error) {
    // Phase 1: return as-is
    return obj, nil
}
```

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

## Tests to Write (10 tests, 10 commits)

### Test 1: TestSecurityService_ValidateAndSanitizePassThrough
```go
func TestSecurityService_ValidateAndSanitizePassThrough(t *testing.T)
```
**Acceptance Criteria:**
- ValidateAndSanitize() returns mail unchanged
- Returns nil error

### Test 2: TestSecurityService_ValidateAndSanitizeWithBoundaries
```go
func TestSecurityService_ValidateAndSanitizeWithBoundaries(t *testing.T)
```
**Acceptance Criteria:**
- ValidateAndSanitize() accepts boundary types
- Phase 1: no enforcement

### Test 3: TestSecurityService_TaintPropagateReturnsObject
```go
func TestSecurityService_TaintPropagateReturnsObject(t *testing.T)
```
**Acceptance Criteria:**
- TaintPropagate() returns object unchanged
- Returns nil error

### Test 4: TestSecurityService_TaintPropagateWithString
```go
func TestSecurityService_TaintPropagateWithString(t *testing.T)
```
**Acceptance Criteria:**
- TaintPropagate() works with string type
- Returns nil error

### Test 5: TestSecurityService_ReportTaintsReturnsEmptyMap
```go
func TestSecurityService_ReportTaintsReturnsEmptyMap(t *testing.T)
```
**Acceptance Criteria:**
- ReportTaints() returns TaintMap
- TaintMap is empty (Phase 1)

### Test 6: TestSecurityService_ReportTaintsWithRuntimeID
```go
func TestSecurityService_ReportTaintsWithRuntimeID(t *testing.T)
```
**Acceptance Criteria:**
- ReportTaints() accepts runtimeId
- Returns nil error

### Test 7: TestSecurityService_PrepareContextForBoundaryNoOp
```go
func TestSecurityService_PrepareContextForBoundaryNoOp(t *testing.T)
```
**Acceptance Criteria:**
- PrepareContextForBoundary() returns nil
- Phase 1: no-op

### Test 8: TestSecurityService_PrepareContextForBoundaryWithInnerBoundary
```go
func TestSecurityService_PrepareContextForBoundaryWithInnerBoundary(t *testing.T)
```
**Acceptance Criteria:**
- Works with InnerBoundary type
- Returns nil error

### Test 9: TestSecurityService_PrepareContextForBoundaryWithOuterBoundary
```go
func TestSecurityService_PrepareContextForBoundaryWithOuterBoundary(t *testing.T)
```
**Acceptance Criteria:**
- Works with OuterBoundary type
- Returns nil error

### Test 10: TestSecurityService_AllMethodsCoexist
```go
func TestSecurityService_AllMethodsCoexist(t *testing.T)
```
**Acceptance Criteria:**
- All new methods exist on SecurityService
- No conflicts with existing methods

## Dependencies

### Test Dependencies
```
Test 1 → Test 2 (ValidateAndSanitize)
Test 3 → Test 4 (TaintPropagate)
Test 5 → Test 6 (ReportTaints)
Test 7 → Test 8, Test 9 (PrepareContextForBoundary)
Test 10 (All methods coexist)
```

### Phase Dependencies
- **Phase 2.1** - Type definitions (TaintMap exists in pkg/security/types.go)

## Files to Create/Modify

| File | Operation | Changes |
|------|-----------|---------|
| `pkg/services/security/service.go` | MODIFY | Add ValidateAndSanitize(), TaintPropagate(), ReportTaints(), PrepareContextForBoundary() |

## Implementation Order (TDD)

1. **Test 1**: Write test → verify RED → Implement ValidateAndSanitize() pass-through → verify GREEN → commit
2. **Test 2**: Write test → verify RED → Accept boundary parameters → verify GREEN → commit
3. **Test 3**: Write test → verify RED → Implement TaintPropagate() stub → verify GREEN → commit
4. **Test 4**: Write test → verify RED → Handle string type in TaintPropagate() → verify GREEN → commit
5. **Test 5**: Write test → verify RED → Implement ReportTaints() stub → verify GREEN → commit
6. **Test 6**: Write test → verify RED → Accept runtimeId parameter → verify GREEN → commit
7. **Test 7**: Write test → verify RED → Implement PrepareContextForBoundary() stub → verify GREEN → commit
8. **Test 8**: Write test → verify RED → Handle InnerBoundary → verify GREEN → commit
9. **Test 9**: Write test → verify RED → Handle OuterBoundary → verify GREEN → commit
10. **Test 10**: Write test → verify RED → Verify all methods coexist → verify GREEN → commit

**Total: 10 tests, 10 commits**

## Deliverables
- ✅ `ValidateAndSanitize()` implemented (pass-through)
- ✅ `TaintPropagate()` implemented (stub)
- ✅ `ReportTaints()` implemented (stub)
- ✅ `PrepareContextForBoundary()` implemented (stub)
- ✅ 10 commits (one per test per TDD workflow)