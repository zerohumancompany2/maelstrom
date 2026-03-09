# Phase 2.2A: Update Existing Tests for Breaking Change

## Goal
Update 3 existing tests that call Publish() to handle the new signature returning (Ack, error).

## Scope
- Update `TestCommunicationService_PublishReturnsNil` to expect (Ack, error)
- Update `TestCommunicationService_PubSub` to expect (Ack, error)
- Update `TestCommunicationService_RoutesMail` to expect (Ack, error)
- No new functionality, just update existing tests for signature change
- 3 tests updated, 1 commit

## Current Status

### Implementation Status
| Test | Status | Notes |
|------|--------|-------|
| TestCommunicationService_PublishReturnsNil | ⚠️ BREAKING | Uses `err := Publish()` |
| TestCommunicationService_PubSub | ⚠️ BREAKING | Uses `err := Publish()` |
| TestCommunicationService_RoutesMail | ⚠️ BREAKING | Uses `err := Publish()` |

### Files Status
| File | Status |
|------|--------|
| `pkg/services/communication/service_test.go` | ⚠️ MODIFY - update 3 tests |

## Required Implementation

### Updated Test Signatures

#### TestCommunicationService_PublishReturnsNil
```go
func TestCommunicationService_PublishReturnsNil(t *testing.T) {
	svc := NewCommunicationService()

	_, err := svc.Publish(mail.Mail{})

	if err != nil {
		t.Errorf("Expected Publish to return nil error, got %v", err)
	}
}
```

#### TestCommunicationService_PubSub
```go
func TestCommunicationService_PubSub(t *testing.T) {
	svc := NewCommunicationService()
	ch, err := svc.Subscribe("test-topic")
	if err != nil {
		t.Errorf("Subscribe should return nil error, got: %v", err)
	}
	if ch == nil {
		t.Fatal("Subscribe should return non-nil channel")
	}

	mail := mail.Mail{Source: "test", Target: "test-topic"}
	_, err = svc.Publish(mail)
	if err != nil {
		t.Errorf("Publish should return nil error, got: %v", err)
	}
	// ... rest unchanged
}
```

#### TestCommunicationService_RoutesMail
```go
func TestCommunicationService_RoutesMail(t *testing.T) {
	svc := NewCommunicationService()

	agentCh, _ := svc.Subscribe("agent:test-agent")
	topicCh, _ := svc.Subscribe("topic:test-topic")
	sysCh, _ := svc.Subscribe("sys:security")

	_, err := svc.Publish(mail.Mail{Source: "test", Target: "agent:test-agent"})
	if err != nil {
		t.Errorf("Publish to agent failed: %v", err)
	}
	// ... rest unchanged
}
```

## Tests to Update (3 tests, 1 commit)

### Test 1: TestCommunicationService_PublishReturnsNil
**Change:** `err := svc.Publish(mail.Mail{})` → `_, err := svc.Publish(mail.Mail{})`

### Test 2: TestCommunicationService_PubSub
**Change:** `err = svc.Publish(mail)` → `_, err = svc.Publish(mail)`

### Test 3: TestCommunicationService_RoutesMail
**Change:** `err := svc.Publish(...)` → `_, err := svc.Publish(...)` (3 occurrences)

## Dependencies

### Test Dependencies
```
No dependencies - these are existing tests being updated
```

### Phase Dependencies
- **Phase 2.2a** must complete before **Phase 2.2b** (Publish returns Ack)

## Files to Create/Modify

| File | Operation | Changes |
|------|-----------|---------|
| `pkg/services/communication/service_test.go` | MODIFY | Update 3 tests to use `_, err := Publish()` signature |

## Implementation Order (TDD)

1. **Update all 3 tests**: Change signature from `err := Publish()` to `_, err := Publish()` → verify GREEN → commit

**Total: 3 tests updated, 1 commit**

## Deliverables
- ✅ Updated `TestCommunicationService_PublishReturnsNil` with new signature
- ✅ Updated `TestCommunicationService_PubSub` with new signature
- ✅ Updated `TestCommunicationService_RoutesMail` with new signature
- ✅ 1 commit (all 3 test updates together)