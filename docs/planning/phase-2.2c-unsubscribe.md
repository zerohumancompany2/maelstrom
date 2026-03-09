# Phase 2.2C: Unsubscribe Implementation

## Goal
Implement Unsubscribe() method to remove subscribers, using existing pkg/mail interfaces.

## Scope
- Implement `Unsubscribe(address string, ch <-chan Mail) error` method
- Remove subscriber channel from address
- Return error when subscriber not found
- Use existing `mail.Subscriber` interface from pkg/mail
- 2 tests, 2 commits

## Current Status

### Implementation Status
| Test | Status | Notes |
|------|--------|-------|
| TestCommunicationService_UnsubscribeRemovesSubscriber | ❌ Missing | Method doesn't exist |
| TestCommunicationService_UnsubscribeNotFoundReturnsError | ❌ Missing | Error handling not implemented |

### Files Status
| File | Status |
|------|--------|
| `pkg/services/communication/service.go` | ⚠️ MODIFY - add Unsubscribe() method |
| `pkg/services/communication/service_test.go` | ⚠️ MODIFY - add 2 new tests |

## Required Implementation

### Unsubscribe Method
```go
// pkg/services/communication/service.go
func (c *CommunicationService) Unsubscribe(address string, ch <-chan Mail) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	subscribers, exists := c.subscribers[address]
	if !exists {
		return fmt.Errorf("no subscribers for address %s", address)
	}

	found := false
	newSubscribers := make([]chan Mail, 0, len(subscribers))
	for _, subCh := range subscribers {
		if subCh == ch {
			found = true
			close(ch)
		} else {
			newSubscribers = append(newSubscribers, subCh)
		}
	}

	if !found {
		return fmt.Errorf("subscriber not found for address %s", address)
	}

	if len(newSubscribers) == 0 {
		delete(c.subscribers, address)
	} else {
		c.subscribers[address] = newSubscribers
	}

	return nil
}
```

## Tests to Write (2 tests, 2 commits)

### Test 1: TestCommunicationService_UnsubscribeRemovesSubscriber
```go
func TestCommunicationService_UnsubscribeRemovesSubscriber(t *testing.T) {
	svc := NewCommunicationService()
	
	ch, err := svc.Subscribe("test-topic")
	if err != nil {
		t.Fatalf("Subscribe failed: %v", err)
	}
	
	m := mail.Mail{Source: "test", Target: "test-topic"}
	_, err = svc.Publish(m)
	if err != nil {
		t.Fatalf("Publish failed: %v", err)
	}
	
	select {
	case <-ch:
	case <-time.After(100 * time.Millisecond):
		t.Fatal("Timeout waiting for mail before unsubscribe")
	}
	
	err = svc.Unsubscribe("test-topic", ch)
	if err != nil {
		t.Errorf("Unsubscribe should return nil error, got %v", err)
	}
	
	_, err = svc.Publish(m)
	if err != nil {
		t.Fatalf("Publish failed: %v", err)
	}
	
	select {
	case <-ch:
		t.Error("Should not receive mail after unsubscribe")
	case <-time.After(10 * time.Millisecond):
	}
}
```
**Acceptance Criteria:**
- Unsubscribe() removes channel from address
- After unsubscribe, no more mail received on that channel
- Returns nil error on success

### Test 2: TestCommunicationService_UnsubscribeNotFoundReturnsError
```go
func TestCommunicationService_UnsubscribeNotFoundReturnsError(t *testing.T) {
	svc := NewCommunicationService()
	
	ch := make(chan mail.Mail)
	
	err := svc.Unsubscribe("non-existent", ch)
	
	if err == nil {
		t.Error("Expected error for non-existent address, got nil")
	}
	if !strings.Contains(err.Error(), "no subscribers") {
		t.Errorf("Expected error mentioning 'no subscribers', got %v", err)
	}
}
```
**Acceptance Criteria:**
- Unsubscribe() with non-existent address returns error
- Error message indicates "no subscribers"

## Dependencies

### Test Dependencies
```
Test 1 → Test 2 (Unsubscribe behavior)
```

### Phase Dependencies
- **Phase 2.2a** must complete first (existing tests updated)
- **Phase 2.2b** should complete first (Publish returns Ack)

## Files to Create/Modify

| File | Operation | Changes |
|------|-----------|---------|
| `pkg/services/communication/service.go` | MODIFY | Add Unsubscribe() method |
| `pkg/services/communication/service_test.go` | MODIFY | Add 2 new tests |

## Implementation Order (TDD)

1. **Test 1**: Write test → verify RED → Implement Unsubscribe() to remove channel and return nil → verify GREEN → commit
2. **Test 2**: Write test → verify RED → Add error handling for non-existent address → verify GREEN → commit

**Total: 2 tests, 2 commits**

## Deliverables
- ✅ `Unsubscribe()` method implemented
- ✅ Returns error when subscriber not found
- ✅ CommunicationService implements mail.Subscriber interface
- ✅ 2 commits (one per test per TDD workflow)