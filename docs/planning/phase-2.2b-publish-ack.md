# Phase 2.2B: Publish Returns Ack

## Goal
Implement Publish() to return Ack with correct fields, using existing pkg/mail interfaces.

## Scope
- Update `Publish()` signature to return (mail.Ack, error)
- Populate Ack with MailID, CorrelationID, DeliveredAt, Success
- Handle publish to non-existent address (Ack.Success = false)
- Use existing `mail.Publisher` interface from pkg/mail
- 3 tests, 3 commits

## Current Status

### Implementation Status
| Test | Status | Notes |
|------|--------|-------|
| TestCommunicationService_PublishReturnsAck | ❌ Missing | Publish returns error only |
| TestCommunicationService_PublishAckHasCorrelationID | ❌ Missing | Ack fields not populated |
| TestCommunicationService_PublishToNonExistentAddress | ❌ Missing | No-subscriber case not handled |

### Files Status
| File | Status |
|------|--------|
| `pkg/services/communication/service.go` | ⚠️ MODIFY - update Publish() signature and implementation |
| `pkg/services/communication/service_test.go` | ⚠️ MODIFY - add 3 new tests |

## Required Implementation

### Updated Publish Method
```go
// pkg/services/communication/service.go
func (c *CommunicationService) Publish(mail mail.Mail) (ack mail.Ack, err error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	ack.MailID = mail.ID
	ack.CorrelationID = mail.CorrelationID
	ack.DeliveredAt = time.Now()

	subscribers, exists := c.subscribers[mail.Target]
	if !exists || len(subscribers) == 0 {
		ack.Success = false
		ack.ErrorMessage = "no subscribers"
		return ack, nil
	}

	for _, ch := range subscribers {
		select {
		case ch <- mail:
			ack.Success = true
		default:
		}
	}

	return ack, nil
}
```

## Tests to Write (3 tests, 3 commits)

### Test 1: TestCommunicationService_PublishReturnsAck
```go
func TestCommunicationService_PublishReturnsAck(t *testing.T) {
	svc := NewCommunicationService()
	
	ch, _ := svc.Subscribe("test-topic")
	m := mail.Mail{Source: "test", Target: "test-topic"}
	
	ack, err := svc.Publish(m)
	
	if err != nil {
		t.Errorf("Expected nil error, got %v", err)
	}
	if ack.MailID != m.ID {
		t.Errorf("Expected MailID %s, got %s", m.ID, ack.MailID)
	}
	if !ack.Success {
		t.Error("Expected Success to be true")
	}
	_ = ch
}
```
**Acceptance Criteria:**
- Publish() returns (mail.Ack, error)
- Ack.MailID matches published mail.ID
- Ack.Success is true when delivered

### Test 2: TestCommunicationService_PublishAckHasCorrelationID
```go
func TestCommunicationService_PublishAckHasCorrelationID(t *testing.T) {
	svc := NewCommunicationService()
	
	ch, _ := svc.Subscribe("test-topic")
	correlationID := "test-correlation-123"
	m := mail.Mail{Source: "test", Target: "test-topic", CorrelationID: correlationID}
	
	ack, err := svc.Publish(m)
	
	if err != nil {
		t.Errorf("Expected nil error, got %v", err)
	}
	if ack.CorrelationID != correlationID {
		t.Errorf("Expected CorrelationID %s, got %s", correlationID, ack.CorrelationID)
	}
	if ack.DeliveredAt.IsZero() {
		t.Error("Expected DeliveredAt to be set")
	}
	_ = ch
}
```
**Acceptance Criteria:**
- Ack.CorrelationID matches mail.CorrelationID
- Ack.DeliveredAt is set to current time (non-zero)

### Test 3: TestCommunicationService_PublishToNonExistentAddress
```go
func TestCommunicationService_PublishToNonExistentAddress(t *testing.T) {
	svc := NewCommunicationService()
	
	m := mail.Mail{Source: "test", Target: "non-existent:address"}
	
	ack, err := svc.Publish(m)
	
	if err != nil {
		t.Errorf("Expected nil error, got %v", err)
	}
	if ack.Success {
		t.Error("Expected Success to be false for non-existent address")
	}
	if ack.ErrorMessage != "no subscribers" {
		t.Errorf("Expected ErrorMessage 'no subscribers', got %s", ack.ErrorMessage)
	}
}
```
**Acceptance Criteria:**
- Publish() to address with no subscribers succeeds (no error)
- Ack.Success is false (no delivery)
- Ack.ErrorMessage indicates "no subscribers"

## Dependencies

### Test Dependencies
```
Test 1 → Test 2 (Ack structure)
Test 3 (Publish edge case - independent)
```

### Phase Dependencies
- **Phase 2.2a** must complete first (existing tests updated)
- **Phase 2.1** - Type definitions (mail.Ack must exist)

## Files to Create/Modify

| File | Operation | Changes |
|------|-----------|---------|
| `pkg/services/communication/service.go` | MODIFY | Update Publish() to return (mail.Ack, error), populate Ack fields |
| `pkg/services/communication/service_test.go` | MODIFY | Add 3 new tests |

## Implementation Order (TDD)

1. **Test 1**: Write test → verify RED → Update Publish() to return Ack with MailID and Success → verify GREEN → commit
2. **Test 2**: Write test → verify RED → Populate Ack.CorrelationID and Ack.DeliveredAt → verify GREEN → commit
3. **Test 3**: Write test → verify RED → Handle no-subscriber case with Success=false and ErrorMessage → verify GREEN → commit

**Total: 3 tests, 3 commits**

## Deliverables
- ✅ `Publish()` returns (mail.Ack, error)
- ✅ Ack populated with MailID, CorrelationID, DeliveredAt, Success, ErrorMessage
- ✅ No-subscriber case handled correctly
- ✅ 3 commits (one per test per TDD workflow)