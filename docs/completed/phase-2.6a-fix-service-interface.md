# Phase 2.6A: Fix Service Interface

## Goal
Fix Service interface to match existing service implementations (ID() instead of Name(), correct HandleMail signature).

## Scope
- Update Service interface to use `ID()` instead of `Name()`
- Update `HandleMail()` signature from `(mail any) (any, error)` to `(mail mail.Mail) error`
- 0 tests, 1 commit (interface fix only - prerequisite for Phase 2.6B)
- **BREAKING CHANGE** - interface only, no implementations yet

## Current Status

### Implementation Status
| Item | Status | Notes |
|------|--------|-------|
| `Service.ID()` | ❌ Mismatch | Interface uses `Name()`, services use `ID()` |
| `Service.HandleMail()` | ❌ Mismatch | Interface uses `(mail any) (any, error)`, services use `(mail mail.Mail) error` |

### Files Status
| File | Status |
|------|--------|
| `pkg/services/registry.go` | ⚠️ Partial - Service interface incorrect |
| `pkg/services/security/service.go` | ✅ Uses `ID()`, `HandleMail(mail mail.Mail) error` |
| `pkg/services/communication/service.go` | ✅ Uses `ID()`, `HandleMail(mail mail.Mail) error` |
| `pkg/services/lifecycle/service.go` | ✅ Uses `ID()`, `HandleMail(mail mail.Mail) error` |
| `pkg/services/observability/service.go` | ✅ Uses `ID()`, `HandleMail(mail mail.Mail) error` |

## Required Implementation

### Service Interface Fix
```go
// pkg/services/registry.go
type Service interface {
	ID() string
	HandleMail(mail mail.Mail) error
	Start() error
	Stop() error
}
```

**Changes:**
- `Name() string` → `ID() string`
- `HandleMail(mail any) (any, error)` → `HandleMail(mail mail.Mail) error`

## Tests to Write (0 tests)

This phase is a prerequisite fix only. No tests are written in this phase.
All 7 tests will be written in Phase 2.6B after interface is fixed.

## Dependencies

### Test Dependencies
```
None - this phase has 0 tests
```

### Phase Dependencies
- **Phase 2.1** - Type definitions (mail.Mail exists)
- **Phase 2.6B** depends on this phase completing first

## Files to Create/Modify

| File | Operation | Changes |
|------|-----------|---------|
| `pkg/services/registry.go` | MODIFY | Fix Service interface (ID, HandleMail signature) |

## Implementation Order

1. **Commit 1**: Update Service interface → verify all services still compile → commit

**Total: 0 tests, 1 commit**

## Deliverables
- ✅ `Service.ID()` interface method (matches existing implementations)
- ✅ `Service.HandleMail(mail mail.Mail) error` interface method (matches existing implementations)
- ✅ 1 commit (interface fix only)
- ✅ All 4 services compile against fixed interface