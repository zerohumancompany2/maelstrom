# Phase 3.8: Integration & CommunicationService

## Goal
Wire Mail system to CommunicationService and verify full integration.

## Scope
- Update CommunicationService to use MailRouter
- Verify full mail flow end-to-end
- Integration tests for all components
- 2 tests, 2 commits

## Current Status

### Implementation Status
| Component | Status | Notes |
|-----------|--------|-------|
| CommunicationService integration | ⚠️ Partial | Stub exists in Layer 2 |
| Full mail flow | ❌ Missing | End-to-end test |

### Files Status
| File | Status |
|------|-------|
| `pkg/services/communication/service.go` | ⚠️ MODIFY - integrate MailRouter |
| `pkg/mail/mail_test.go` | ❌ MISSING - create comprehensive tests |

## Required Implementation

### CommunicationService Integration
```go
// pkg/services/communication/service.go
package communication

import (
    "maelstrom/pkg/mail"
    "sync"
    "time"
)

type CommunicationService struct {
    id          string
    router      *mail.MailRouter
    publisher   mail.Publisher
    subscribers map[string][]mail.Subscriber
    mu          sync.RWMutex
}

func NewCommunicationService() *CommunicationService {
    router := mail.NewMailRouter()
    return &CommunicationService{
        id:          "sys:communication",
        router:      router,
        publisher:   mail.NewRouterPublisher(router),
        subscribers: make(map[string][]mail.Subscriber),
    }
}

func (s *CommunicationService) ID() string {
    return s.id
}

func (s *CommunicationService) Publish(m mail.Mail) (ack mail.Ack, err error) {
    return s.publisher.Publish(m)
}

func (s *CommunicationService) Subscribe(address string) (chan mail.Mail, error) {
    ch := make(chan mail.Mail, 100)
    
    sub := &channelSubscriber{ch: ch}
    
    // Create inbox for this address if needed
    if mail.IsValidAgentAddress(address) || mail.IsValidSysAddress(address) {
        inbox := &mail.AgentInbox{ID: address}
        s.router.SubscribeAgent(address, inbox)
    } else if mail.IsValidTopicAddress(address) {
        topic := &mail.Topic{Name: address}
        s.router.SubscribeTopic(address, topic)
    }
    
    s.mu.Lock()
    s.subscribers[address] = append(s.subscribers[address], sub)
    s.mu.Unlock()
    
    return ch, nil
}

func (s *CommunicationService) Unsubscribe(address string, stream chan mail.Mail) error {
    s.mu.Lock()
    defer s.mu.Unlock()
    
    subs, exists := s.subscribers[address]
    if !exists {
        return errors.New("no subscribers for address: " + address)
    }
    
    for i, sub := range subs {
        if sub.(*channelSubscriber).ch == stream {
            s.subscribers[address] = append(subs[:i], subs[i+1:]...)
            return nil
        }
    }
    
    return errors.New("stream not found")
}

func (s *CommunicationService) UpgradeToStream(sessionID string, lastEventID *string) (chan mail.StreamChunk, error) {
    return mail.UpgradeToStream(sessionID, lastEventID)
}

func (s *CommunicationService) HandleMail(m mail.Mail) error {
    return s.router.Route(m)
}

type channelSubscriber struct {
    ch chan mail.Mail
}

func (c *channelSubscriber) Receive() <-chan mail.Mail {
    return c.ch
}
```

### Comprehensive Mail Tests
```go
// pkg/mail/mail_test.go
package mail

import (
    "testing"
    "time"
)

func TestFullMailFlow(t *testing.T) {
    // Setup
    router := NewMailRouter()
    publisher := NewRouterPublisher(router)
    
    // Create subscriber inbox
    inbox := &AgentInbox{ID: "test-agent"}
    router.SubscribeAgent("test-agent", inbox)
    
    // Create mail with all fields
    originalMail := Mail{
        ID:            "msg-001",
        CorrelationID: "corr-001",
        Type:          MailTypeUser,
        CreatedAt:     time.Now(),
        Source:        "agent:user-agent",
        Target:        "agent:test-agent",
        Content:       map[string]any{"text": "hello"},
        Metadata: MailMetadata{
            Tokens:   10,
            Model:    "gpt-4",
            Cost:     0.01,
            Boundary: OuterBoundary,
            Taints:   []string{"USER_SUPPLIED"},
        },
    }
    
    // Publish
    ack, err := publisher.Publish(originalMail)
    if err != nil {
        t.Fatalf("Publish failed: %v", err)
    }
    
    // Verify Ack
    if ack.CorrelationID != originalMail.CorrelationID {
        t.Errorf("Expected CorrelationID '%s', got '%s'", 
            originalMail.CorrelationID, ack.CorrelationID)
    }
    
    if ack.DeliveredAt.IsZero() {
        t.Error("Expected DeliveredAt to be set")
    }
    
    // Verify delivery to inbox
    inbox.mu.RLock()
    if len(inbox.Messages) != 1 {
        t.Errorf("Expected 1 message in inbox, got %d", len(inbox.Messages))
    }
    deliveredMail := inbox.Messages[0]
    inbox.mu.RUnlock()
    
    // Verify mail integrity
    if deliveredMail.ID != originalMail.ID {
        t.Errorf("Expected ID '%s', got '%s'", originalMail.ID, deliveredMail.ID)
    }
    if deliveredMail.Type != originalMail.Type {
        t.Errorf("Expected Type '%s', got '%s'", originalMail.Type, deliveredMail.Type)
    }
    if deliveredMail.Source != originalMail.Source {
        t.Errorf("Expected Source '%s', got '%s'", originalMail.Source, deliveredMail.Source)
    }
}

func TestCommunicationService_Integration(t *testing.T) {
    // This test requires the communication package
    // Import it and test full integration
    
    // For now, test the core components work together
    router := NewMailRouter()
    
    // Test agent-to-agent routing
    agent1 := &AgentInbox{ID: "agent1"}
    agent2 := &AgentInbox{ID: "agent2"}
    router.SubscribeAgent("agent1", agent1)
    router.SubscribeAgent("agent2", agent2)
    
    mail1 := Mail{
        ID:     "msg-001",
        Source: "agent:agent1",
        Target: "agent:agent2",
        Type:   MailTypeUser,
    }
    
    err := router.Route(mail1)
    if err != nil {
        t.Fatalf("Route failed: %v", err)
    }
    
    agent2.mu.RLock()
    if len(agent2.Messages) != 1 {
        t.Errorf("Expected 1 message, got %d", len(agent2.Messages))
    }
    agent2.mu.RUnlock()
    
    // Test topic publishing
    topic := &Topic{Name: "events"}
    router.SubscribeTopic("events", topic)
    
    sub1 := &ChannelSubscriber{ch: make(chan Mail, 10)}
    sub2 := &ChannelSubscriber{ch: make(chan Mail, 10)}
    topic.Subscribe(sub1)
    topic.Subscribe(sub2)
    
    mail2 := Mail{
        ID:     "msg-002",
        Source: "sys:events",
        Target: "topic:events",
        Type:   MailTypeAssistant,
    }
    
    err = router.Route(mail2)
    if err != nil {
        t.Fatalf("Topic route failed: %v", err)
    }
    
    // Verify both subscribers received the mail
    select {
    case received := <-sub1.ch:
        if received.ID != "msg-002" {
            t.Errorf("Expected msg-002, got %s", received.ID)
        }
    default:
        t.Error("Subscriber 1 did not receive mail")
    }
    
    select {
    case received := <-sub2.ch:
        if received.ID != "msg-002" {
            t.Errorf("Expected msg-002, got %s", received.ID)
        }
    default:
        t.Error("Subscriber 2 did not receive mail")
    }
    
    // Test sys service routing
    serviceInbox := &ServiceInbox{ID: "heartbeat"}
    router.SubscribeService("heartbeat", serviceInbox)
    
    mail3 := Mail{
        ID:     "msg-003",
        Source: "agent:scheduler",
        Target: "sys:heartbeat",
        Type:   MailTypeHeartbeat,
    }
    
    err = router.Route(mail3)
    if err != nil {
        t.Fatalf("Service route failed: %v", err)
    }
    
    serviceInbox.mu.RLock()
    if len(serviceInbox.Messages) != 1 {
        t.Errorf("Expected 1 message in service inbox, got %d", len(serviceInbox.Messages))
    }
    serviceInbox.mu.RUnlock()
}
```

## Tests to Write (2 tests, 2 commits)

### Test 1: TestFullMailFlow
**Acceptance Criteria:**
- Mail created with all fields
- Mail routed through MailRouter
- Mail delivered to subscriber
- Ack returned with correct fields
- Mail integrity preserved through routing

### Test 2: TestCommunicationService_Integration
**Acceptance Criteria:**
- CommunicationService uses MailRouter
- Agent-to-agent routing works
- Topic publishing works (broadcast)
- Sys service routing works
- All components integrated correctly

## Dependencies

### Test Dependencies
```
Test 1 → Test 2 (Full flow before Integration)
```

### Phase Dependencies
- **Phase 3.2** must complete first (MailRouter)
- **Phase 3.3a, 3.3b** must complete first (Inboxes, Topic)
- **Phase 3.4** must complete first (Publisher/Subscriber)
- **Phase 3.5** must complete first (Streaming)

## Files to Create/Modify

| File | Operation | Changes |
|------|-----------|---------|
| `pkg/services/communication/service.go` | MODIFY | Integrate MailRouter, update methods |
| `pkg/mail/mail_test.go` | CREATE | Comprehensive integration tests |

## Implementation Order (TDD)

1. **Test 1**: Write test → verify RED → Wire MailRouter into CommunicationService → verify GREEN → commit
2. **Test 2**: Write test → verify RED → Verify full integration (agent, topic, service) → verify GREEN → commit

**Total: 2 tests, 2 commits**

## Deliverables
- ✅ CommunicationService integrated with MailRouter
- ✅ Full mail flow working end-to-end
- ✅ Agent-to-agent routing
- ✅ Topic broadcast
- ✅ Sys service routing
- ✅ 2 commits (one per test per TDD workflow)

## Recursive Analysis Decision

**Decision: Keep as-is (no split)**

**Reasoning:**
- 2 tests is within recommended range (2-5)
- Tests are integration tests that verify everything works together
- Single coherent feature: Full integration
- Splitting would defeat the purpose of integration tests

**Alternative (if split needed):**
- 3.8a: CommunicationService wiring - 1 test
- 3.8b: Full integration verification - 1 test