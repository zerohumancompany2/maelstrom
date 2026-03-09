# Phase 1.3.1: Security Service Stub

## Goal
Create stub implementation for the security service with 5 pass-through methods.

## Scope
- Create `pkg/services/security/service.go`
- Implement 5 stub methods
- Write 5 tests in `pkg/services/security/service_test.go`
- All methods return nil/empty (no actual boundary enforcement)

## Current Status

### File Status
| File | Status |
|------|--------|
| `pkg/services/security/bootstrap.go` | ✅ EXISTS - BootstrapChart() stub |
| `pkg/services/security/service.go` | ❌ **MISSING** - Must create |
| `pkg/services/security/service_test.go` | ✅ EXISTS - Has placeholder tests |

### Implementation Status
- **0/5 methods implemented**
- Service struct does not exist
- Only bootstrap.go exists

## Required Implementation

### Security Service Struct
```go
package security

type SecurityService struct{}
```

### Constructor
```go
func NewSecurityService() *SecurityService {
    return &SecurityService{}
}
```

### 5 Methods

| # | Method | Signature | Return Value |
|---|--------|-----------|--------------|
| 1 | `ID()` | `func (s *SecurityService) ID() string` | `"sys:security"` |
| 2 | `HandleMail()` | `func (s *SecurityService) HandleMail(mail mail.Mail) error` | `nil` |
| 3 | `ValidateBoundary()` | `func (s *SecurityService) ValidateBoundary(source, target mail.BoundaryType) error` | `nil` |
| 4 | `Start()` | `func (s *SecurityService) Start() error` | `nil` |
| 5 | `Stop()` | `func (s *SecurityService) Stop() error` | `nil` |

## Tests to Write (5 tests, 5 commits)

### Test 1: Constructor returns non-nil
```go
func TestSecurityService_NewSecurityServiceReturnsNonNil()
```
**Acceptance Criteria:**
- `NewSecurityService()` returns a non-nil `*SecurityService`

### Test 2: ID returns correct string
```go
func TestSecurityService_IDReturnsCorrectString()
```
**Acceptance Criteria:**
- `ID()` returns exactly `"sys:security"`

### Test 3: HandleMail returns nil
```go
func TestSecurityService_HandleMailReturnsNil()
```
**Acceptance Criteria:**
- `HandleMail(mail)` returns `nil` error for any input

### Test 4: ValidateBoundary returns nil
```go
func TestSecurityService_ValidateBoundaryReturnsNil()
```
**Acceptance Criteria:**
- `ValidateBoundary(source, target)` returns `nil` error for any boundary types

### Test 5: Start returns nil
```go
func TestSecurityService_StartReturnsNil()
```
**Acceptance Criteria:**
- `Start()` returns `nil` error

### Test 6: Stop returns nil
```go
func TestSecurityService_StopReturnsNil()
```
**Acceptance Criteria:**
- `Stop()` returns `nil` error

## Dependencies

| Package | Types Used | Purpose |
|---------|------------|---------|
| `github.com/maelstrom/v3/pkg/mail` | `Mail`, `BoundaryType` | `HandleMail()` parameter and `ValidateBoundary()` parameters |

## Files to Create/Modify

| File | Operation | Changes |
|------|-----------|---------|
| `pkg/services/security/service.go` | CREATE | SecurityService struct + 5 stub methods |
| `pkg/services/security/service_test.go` | MODIFY | Add 5 tests |

## Implementation Order (TDD)

1. **Test 1**: Write test → verify RED → Implement `NewSecurityService()` → verify GREEN → commit
2. **Test 2**: Write test → verify RED → Implement `ID()` → verify GREEN → commit
3. **Test 3**: Write test → verify RED → Implement `HandleMail()` → verify GREEN → commit
4. **Test 4**: Write test → verify RED → Implement `ValidateBoundary()` → verify GREEN → commit
5. **Test 5**: Write test → verify RED → Implement `Start()` → verify GREEN → commit
6. **Test 6**: Write test → verify RED → Implement `Stop()` → verify GREEN → commit

**Total: 6 tests, 6 commits**

## Deliverables
- New `pkg/services/security/service.go` with 5 stub methods
- Updated `pkg/services/security/service_test.go` with 6 passing tests
- 6 commits (one per test per TDD workflow)