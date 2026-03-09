# Phase 2.5A: ValidateAndSanitize and TaintPropagate

## Goal
Implement ValidateAndSanitize pass-through and TaintPropagate stub in Security Service, following strict TDD workflow with consolidated tests.

## Scope
- Implement `ValidateAndSanitize()` pass-through (combines original tests 1+2)
- Implement `TaintPropagate()` stub (combines original tests 3+4)
- 2 tests, 2 commits
- **NO breaking changes** - existing tests must pass

## Current Status

### Implementation Status
| Item | Status | Notes |
|------|--------|-------|
| `ValidateAndSanitize()` | ❌ Missing | Method doesn't exist |
| `TaintPropagate()` | ❌ Missing | Method doesn't exist |

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

## Tests to Write (2 tests, 2 commits)

### Test 1: TestSecurityService_ValidateAndSanitizePassThrough
```go
func TestSecurityService_ValidateAndSanitizePassThrough(t *testing.T)
```
**Acceptance Criteria:**
- ValidateAndSanitize() returns mail unchanged
- Returns nil error
- Accepts boundary type parameters (sourceBoundary, targetBoundary)
- Phase 1: no enforcement

**Consolidates original tests:**
- Test 1: TestSecurityService_ValidateAndSanitizePassThrough
- Test 2: TestSecurityService_ValidateAndSanitizeWithBoundaries

### Test 2: TestSecurityService_TaintPropagateReturnsObject
```go
func TestSecurityService_TaintPropagateReturnsObject(t *testing.T)
```
**Acceptance Criteria:**
- TaintPropagate() returns object unchanged
- Returns nil error
- Works with any type (test with string)
- Accepts newTaints parameter (Phase 1: ignored)

**Consolidates original tests:**
- Test 3: TestSecurityService_TaintPropagateReturnsObject
- Test 4: TestSecurityService_TaintPropagateWithString

## Dependencies

### Test Dependencies
```
Test 1 (ValidateAndSanitize) → Independent
Test 2 (TaintPropagate) → Independent
```

### Phase Dependencies
- **Phase 2.1** - Type definitions (mail.Mail, mail.BoundaryType exist)
- **Phase 2.5B** depends on this phase completing first

## Files to Create/Modify

| File | Operation | Changes |
|------|-----------|---------|
| `pkg/services/security/service.go` | MODIFY | Add ValidateAndSanitize(), TaintPropagate() |

## Implementation Order (TDD)

1. **Test 1**: Write test → verify RED → Implement ValidateAndSanitize() pass-through → verify GREEN → commit
2. **Test 2**: Write test → verify RED → Implement TaintPropagate() stub → verify GREEN → commit

**Total: 2 tests, 2 commits**

## Deliverables
- ✅ `ValidateAndSanitize()` implemented (pass-through)
- ✅ `TaintPropagate()` implemented (stub)
- ✅ 2 commits (one per test per TDD workflow)
- ✅ NO breaking changes - existing tests still pass