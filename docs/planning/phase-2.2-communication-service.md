# Phase 2.2: Communication Service Enhancement

## Goal
Enhance Communication Service with Ack support, unsubscribe, and proper pub/sub interfaces.

## Scope
- Update `Publish()` to return `Ack`
- Implement `Unsubscribe()` method
- Create `publisher.go` with Publisher interface
- Create `subscriber.go` with Subscriber interface
- Add streaming support stub
- 8 tests, 8 commits

## Current Status

### Implementation Status
| Test | Status | Notes |
|------|--------|-------|
| TestCommunicationService_PublishReturnsAck | ❌ Missing | Publish returns error only |
| TestCommunicationService_Unsubscribe | ❌ Missing | Method doesn't exist |
| TestCommunicationService_UpgradeToStream | ❌ Missing | Method doesn't exist |
| TestCommunication_PublisherInterface | ❌ Missing | publisher.go missing |
| TestCommunication_SubscriberInterface | ❌ Missing | subscriber.go missing |

### Files Status
| File | Status |
|------|--------|
| `pkg/services/communication/publisher.go` | ❌ MISSING - must create |
| `pkg/services/communication/subscriber.go` | ❌ MISSING - must create |
| `pkg/services/communication/service.go` | ⚠️ Partial - update Publish() signature |

## Required Implementation

### Publisher Interface
```go
// pkg/services/communication/publisher.go
type Publisher interface {
    Publish(mail mail.Mail) (ack mail.Ack, err error)
}
```

### Subscriber Interface
```go
// pkg/services/communication/subscriber.go
type Subscriber interface {
    Subscribe(address string) (<-chan mail.Mail, error)
    Unsubscribe(address string, stream chan mail.Mail) error
}
```

### Updated Publish Method
```go
// pkg/services/communication/service.go
func (c *CommunicationService) Publish(mail mail.Mail) (ack mail.Ack, err error) {
    // Return Ack with delivery confirmation
}
```

## Tests to Write (8 tests, 8 commits)

### Test 1: TestCommunicationService_PublishReturnsAck
```go
func TestCommunicationService_PublishReturnsAck(t *testing.T)
```
**Acceptance Criteria:**
- Publish() returns (mail.Ack, error)
- Ack has MailID matching published mail
- Ack.Success is true when delivered

### Test 2: TestCommunicationService_PublishAckHasCorrelationID
```go
func TestCommunicationService_PublishAckHasCorrelationID(t *testing.T)
```
**Acceptance Criteria:**
- Ack.CorrelationID matches mail.CorrelationID
- Ack.DeliveredAt is set to current time

### Test 3: TestCommunicationService_UnsubscribeRemovesSubscriber
```go
func TestCommunicationService_UnsubscribeRemovesSubscriber(t *testing.T)
```
**Acceptance Criteria:**
- Unsubscribe() removes stream from address
- After unsubscribe, no more mail received
- Returns nil error on success

### Test 4: TestCommunicationService_UnsubscribeNotFoundReturnsError
```go
func TestCommunicationService_UnsubscribeNotFoundReturnsError(t *testing.T)
```
**Acceptance Criteria:**
- Unsubscribe() with non-existent stream returns error
- Error message indicates stream not found

### Test 5: TestCommunication_PublisherInterfaceImplemented
```go
func TestCommunication_PublisherInterfaceImplemented(t *testing.T)
```
**Acceptance Criteria:**
- CommunicationService implements Publisher interface
- Publisher interface defined in publisher.go

### Test 6: TestCommunication_SubscriberInterfaceImplemented
```go
func TestCommunication_SubscriberInterfaceImplemented(t *testing.T)
```
**Acceptance Criteria:**
- CommunicationService implements Subscriber interface
- Subscriber interface defined in subscriber.go

### Test 7: TestCommunicationService_UpgradeToStreamReturnsChannel
```go
func TestCommunicationService_UpgradeToStreamReturnsChannel(t *testing.T)
```
**Acceptance Criteria:**
- UpgradeToStream() returns channel of StreamChunk
- Channel is non-nil

### Test 8: TestCommunicationService_PublishToNonExistentAddress
```go
func TestCommunicationService_PublishToNonExistentAddress(t *testing.T)
```
**Acceptance Criteria:**
- Publish() to address with no subscribers succeeds
- Ack.Success is false (no delivery)
- Ack.ErrorMessage indicates no subscribers

## Dependencies

### Test Dependencies
```
Test 1 → Test 2 (Ack structure)
Test 5 (Publisher interface)
Test 3 → Test 4 (Unsubscribe behavior)
Test 6 (Subscriber interface)
Test 7 (UpgradeToStream)
Test 8 (Publish edge case)
```

### Phase Dependencies
- **Phase 2.1** - Type definitions (StreamChunk must exist in mail package)

## Files to Create/Modify

| File | Operation | Changes |
|------|-----------|---------|
| `pkg/services/communication/publisher.go` | CREATE | Publisher interface |
| `pkg/services/communication/subscriber.go` | CREATE | Subscriber interface |
| `pkg/services/communication/service.go` | MODIFY | Update Publish(), add Unsubscribe(), UpgradeToStream() |

## Implementation Order (TDD)

1. **Test 1**: Write test → verify RED → Update Publish() to return Ack → verify GREEN → commit
2. **Test 2**: Write test → verify RED → Populate Ack fields → verify GREEN → commit
3. **Test 3**: Write test → verify RED → Implement Unsubscribe() → verify GREEN → commit
4. **Test 4**: Write test → verify RED → Add error handling to Unsubscribe() → verify GREEN → commit
5. **Test 5**: Write test → verify RED → Create publisher.go with interface → verify GREEN → commit
6. **Test 6**: Write test → verify RED → Create subscriber.go with interface → verify GREEN → commit
7. **Test 7**: Write test → verify RED → Implement UpgradeToStream() stub → verify GREEN → commit
8. **Test 8**: Write test → verify RED → Handle no-subscriber case in Publish() → verify GREEN → commit

**Total: 8 tests, 8 commits**

## Deliverables
- ✅ `Publisher` interface in `publisher.go`
- ✅ `Subscriber` interface in `subscriber.go`
- ✅ `Publish()` returns Ack
- ✅ `Unsubscribe()` implemented
- ✅ `UpgradeToStream()` stub
- ✅ 8 commits (one per test per TDD workflow)