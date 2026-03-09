# Phase 3.4: Publisher/Subscriber Interfaces

## Goal
Implement Publisher and Subscriber interfaces with concrete implementations.

## Scope
- Create `pkg/mail/publisher.go` with Publisher interface
- Create `pkg/mail/subscriber.go` with Subscriber interface
- Implement concrete types that satisfy interfaces
- Integration with MailRouter
- 4 tests, 4 commits

## Current Status

### Implementation Status
| Component | Status | Notes |
|-----------|--------|-------|
| `Publisher` interface | ❌ Missing | Publish method |
| `Subscriber` interface | ❌ Missing | Receive method |
| Concrete implementations | ❌ Missing | Channel-based |

### Files Status
| File | Status |
|------|-------|
| `pkg/mail/publisher.go` | ❌ MISSING - create |
| `pkg/mail/subscriber.go` | ❌ MISSING - create |

## Required Implementation

### Publisher Interface
```go
// pkg/mail/publisher.go
package mail

type Publisher interface {
    Publish(mail Mail) (Ack, error)
}

type RouterPublisher struct {
    router *MailRouter
}

func NewRouterPublisher(router *MailRouter) *RouterPublisher {
    return &RouterPublisher{router: router}
}

func (p *RouterPublisher) Publish(mail Mail) (Ack, error) {
    err := p.router.Route(mail)
    return Ack{
        CorrelationID: mail.CorrelationID,
        DeliveredAt:   time.Now(),
    }, err
}
```

### Subscriber Interface
```go
// pkg/mail/subscriber.go
package mail

type Subscriber interface {
    Receive() <-chan Mail
}

type ChannelSubscriber struct {
    ch chan Mail
}

func NewChannelSubscriber(bufferSize int) *ChannelSubscriber {
    return &ChannelSubscriber{
        ch: make(chan Mail, bufferSize),
    }
}

func (s *ChannelSubscriber) Receive() <-chan Mail {
    return s.ch
}

func (s *ChannelSubscriber) GetMail() (Mail, bool) {
    select {
    case mail := <-s.ch:
        return mail, true
    default:
        return Mail{}, false
    }
}
```

## Tests to Write (4 tests, 4 commits)

### Test 1: TestPublisher_Publish
```go
func TestPublisher_Publish(t *testing.T) {
    router := NewMailRouter()
    publisher := NewRouterPublisher(router)
    
    // Setup: register an agent
    inbox := &AgentInbox{ID: "test-agent"}
    router.SubscribeAgent("test-agent", inbox)
    
    // Publish mail
    mail := Mail{
        ID:     "msg-001",
        Source: "agent:user",
        Target: "agent:test-agent",
        Type:   MailTypeUser,
    }
    
    ack, err := publisher.Publish(mail)
    if err != nil {
        t.Errorf("Expected nil error, got %v", err)
    }
    
    if ack.CorrelationID != mail.CorrelationID {
        t.Errorf("Expected CorrelationID to match, got %s", ack.CorrelationID)
    }
    
    if ack.DeliveredAt.IsZero() {
        t.Error("Expected DeliveredAt to be set")
    }
}
```
**Acceptance Criteria:**
- Publisher.Publish sends mail correctly
- Returns Ack with correct fields
- Returns nil error on success

### Test 2: TestSubscriber_Receives
```go
func TestSubscriber_Receives(t *testing.T) {
    sub := NewChannelSubscriber(10)
    
    ch := sub.Receive()
    if ch == nil {
        t.Error("Expected non-nil channel")
    }
    
    // Send mail through channel
    mail := Mail{ID: "msg-001", Type: MailTypeUser}
    ch <- mail
    
    // Receive mail
    received, ok := sub.GetMail()
    if !ok {
        t.Error("Expected to receive mail")
    }
    
    if received.ID != "msg-001" {
        t.Errorf("Expected msg-001, got %s", received.ID)
    }
}
```
**Acceptance Criteria:**
- Subscriber.Receive returns channel
- Mail received on channel after publish
- GetMail retrieves mail without blocking when available

### Test 3: TestPublisher_SubscriberIntegration
```go
func TestPublisher_SubscriberIntegration(t *testing.T) {
    router := NewMailRouter()
    publisher := NewRouterPublisher(router)
    
    // Setup subscriber
    sub := NewChannelSubscriber(10)
    inbox := &AgentInbox{ID: "test-agent"}
    router.SubscribeAgent("test-agent", inbox)
    
    // Publish mail
    mail := Mail{
        ID:     "msg-001",
        Source: "agent:user",
        Target: "agent:test-agent",
        Type:   MailTypeUser,
    }
    
    _, err := publisher.Publish(mail)
    if err != nil {
        t.Errorf("Expected nil error, got %v", err)
    }
    
    // Verify mail reached inbox
    inbox.mu.RLock()
    if len(inbox.Messages) != 1 {
        t.Errorf("Expected 1 message in inbox, got %d", len(inbox.Messages))
    }
    inbox.mu.RUnlock()
}
```
**Acceptance Criteria:**
- Publisher and Subscriber work together
- Mail flows from publisher through router to inbox

### Test 4: TestPublisherInterface_ImplementedByRouter
```go
func TestPublisherInterface_ImplementedByRouter(t *testing.T) {
    router := NewMailRouter()
    publisher := NewRouterPublisher(router)
    
    // Verify RouterPublisher implements Publisher interface
    var _ Publisher = publisher
    
    // Use as Publisher
    var p Publisher = publisher
    if p == nil {
        t.Error("Expected non-nil Publisher")
    }
    
    // Also verify MailRouter can be used directly
    // (if MailRouter implements Publisher directly)
    // var _ Publisher = router  // Uncomment if MailRouter implements Publisher
}
```
**Acceptance Criteria:**
- MailRouter (via RouterPublisher) implements Publisher interface
- Can use RouterPublisher as Publisher
- Type assertion succeeds

## Dependencies

### Test Dependencies
```
Test 1 → Test 2 (Publish before Receive)
Test 3 (Integration - depends on 1 and 2)
Test 4 (Interface check - depends on 1)
```

### Phase Dependencies
- **Phase 3.1** must complete first (Mail types)
- **Phase 3.2** must complete first (MailRouter)
- **Phase 3.3a, 3.3b** must complete first (Inboxes)

## Files to Create/Modify

| File | Operation | Changes |
|------|-----------|---------|
| `pkg/mail/publisher.go` | CREATE | Publisher interface and RouterPublisher |
| `pkg/mail/subscriber.go` | CREATE | Subscriber interface and ChannelSubscriber |

## Implementation Order (TDD)

1. **Test 1**: Write test → verify RED → Implement Publisher interface and RouterPublisher → verify GREEN → commit
2. **Test 2**: Write test → verify RED → Implement Subscriber interface and ChannelSubscriber → verify GREEN → commit
3. **Test 3**: Write test → verify RED → Verify integration works → verify GREEN → commit
4. **Test 4**: Write test → verify RED → Verify interface implementation → verify GREEN → commit

**Total: 4 tests, 4 commits**

## Deliverables
- ✅ Publisher interface in `pkg/mail/publisher.go`
- ✅ Subscriber interface in `pkg/mail/subscriber.go`
- ✅ RouterPublisher implementation
- ✅ ChannelSubscriber implementation
- ✅ 4 commits (one per test per TDD workflow)

## Recursive Analysis Decision

**Decision: Keep as-is (no split)**

**Reasoning:**
- 4 tests is within recommended range (2-5)
- Tests are tightly coupled around pub/sub interfaces
- Single coherent feature: Publisher/Subscriber pattern
- Splitting would create unnecessary fragmentation

**Alternative (if split needed):**
- 3.4a: Publisher interface + RouterPublisher - 2 tests
- 3.4b: Subscriber interface + ChannelSubscriber - 2 tests