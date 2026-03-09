# Phase 1.3.2: Communication Service Stub

## Goal
Create stub implementation for the communication service with 7 pass-through methods.

## Scope
- Create `pkg/services/communication/service.go`
- Implement 7 stub methods
- Write 7 tests in `pkg/services/communication/service_test.go`
- All methods return nil/empty (no actual pub/sub functionality)

## Current Status

### File Status
| File | Status |
|------|--------|
| `pkg/services/communication/bootstrap.go` | ✅ EXISTS - BootstrapChart() stub |
| `pkg/services/communication/service.go` | ❌ **MISSING** - Must create |
| `pkg/services/communication/service_test.go` | ✅ EXISTS - Has placeholder tests |

### Implementation Status
- **0/7 methods implemented**
- Service struct does not exist
- Only bootstrap.go exists

## Required Implementation

### Communication Service Struct
```go
package communication

type CommunicationService struct{}
```

### Constructor
```go
func NewCommunicationService() *CommunicationService {
    return &CommunicationService{}
}
```

### 7 Methods

| # | Method | Signature | Return Value |
|---|--------|-----------|--------------|
| 1 | `ID()` | `func (c *CommunicationService) ID() string` | `"sys:communication"` |
| 2 | `HandleMail()` | `func (c *CommunicationService) HandleMail(mail Mail) error` | `nil` |
| 3 | `Publish()` | `func (c *CommunicationService) Publish(mail Mail) error` | `nil` |
| 4 | `Subscribe()` | `func (c *CommunicationService) Subscribe(address string) (<-chan Mail, error)` | `nil, nil` |
| 5 | `Start()` | `func (c *CommunicationService) Start() error` | `nil` |
| 6 | `Stop()` | `func (c *CommunicationService) Stop() error` | `nil` |

## Tests to Write (7 tests, 7 commits)

### Test 1: Constructor returns non-nil
```go
func TestCommunicationService_NewCommunicationServiceReturnsNonNil()
```
**Acceptance Criteria:**
- `NewCommunicationService()` returns a non-nil `*CommunicationService`

### Test 2: ID returns correct string
```go
func TestCommunicationService_IDReturnsCorrectString()
```
**Acceptance Criteria:**
- `ID()` returns exactly `"sys:communication"`

### Test 3: HandleMail returns nil
```go
func TestCommunicationService_HandleMailReturnsNil()
```
**Acceptance Criteria:**
- `HandleMail(mail)` returns `nil` error

### Test 4: Publish returns nil
```go
func TestCommunicationService_PublishReturnsNil()
```
**Acceptance Criteria:**
- `Publish(mail)` returns `nil` error

### Test 5: Subscribe returns nil channel and error
```go
func TestCommunicationService_SubscribeReturnsNilChannelAndError()
```
**Acceptance Criteria:**
- `Subscribe(address)` returns `(nil, nil)`

### Test 6: Start returns nil
```go
func TestCommunicationService_StartReturnsNil()
```
**Acceptance Criteria:**
- `Start()` returns `nil` error

### Test 7: Stop returns nil
```go
func TestCommunicationService_StopReturnsNil()
```
**Acceptance Criteria:**
- `Stop()` returns `nil` error

## Dependencies

| Package | Types Used | Purpose |
|---------|------------|---------|
| `github.com/maelstrom/v3/pkg/mail` | `Mail` | Parameter type for `HandleMail()`, `Publish()`, return channel type for `Subscribe()` |

### Mail Type Structure (from `pkg/mail/types.go`)
```go
type Mail struct {
    ID            string
    CorrelationID string
    Type          MailType
    CreatedAt     time.Time
    Source        string
    Target        string
    Content       any
    Metadata      MailMetadata
}
```

## Files to Create/Modify

| File | Operation | Changes |
|------|-----------|---------|
| `pkg/services/communication/service.go` | CREATE | CommunicationService struct + 7 stub methods |
| `pkg/services/communication/service_test.go` | MODIFY | Add 7 tests |

## Implementation Order (TDD)

1. **Test 1**: Write test → verify RED → Implement `NewCommunicationService()` → verify GREEN → commit
2. **Test 2**: Write test → verify RED → Implement `ID()` → verify GREEN → commit
3. **Test 3**: Write test → verify RED → Implement `HandleMail()` → verify GREEN → commit
4. **Test 4**: Write test → verify RED → Implement `Publish()` → verify GREEN → commit
5. **Test 5**: Write test → verify RED → Implement `Subscribe()` → verify GREEN → commit
6. **Test 6**: Write test → verify RED → Implement `Start()` → verify GREEN → commit
7. **Test 7**: Write test → verify RED → Implement `Stop()` → verify GREEN → commit

**Total: 7 tests, 7 commits**

## Deliverables
- New `pkg/services/communication/service.go` with 7 stub methods
- Updated `pkg/services/communication/service_test.go` with 7 passing tests
- 7 commits (one per test per TDD workflow)