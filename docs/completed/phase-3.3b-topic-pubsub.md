# Phase 3.3B: Topic Pub/Sub

## Goal
Implement Topic data structure for pub-sub messaging.

## Scope
- Add Topic to `pkg/mail/inbox.go`
- Implement Topic with Publish/Subscribe/Unsubscribe
- Channel-based pub/sub
- 3 tests, 3 commits

## Current Status

### Implementation Status
| Component | Status | Notes |
|-----------|--------|-------|
| `Topic` | ❌ Missing | Pub-sub topic |

### Files Status
| File | Status |
|------|-------|
| `pkg/mail/inbox.go` | ⚠️ Partial - Topic missing (AgentInbox, ServiceInbox from 3.3a) |

## Required Implementation

### Topic
```go
// pkg/mail/inbox.go (add to existing file)
package mail

import (
    "errors"
    "sync"
)

type Subscriber interface {
    Receive() <-chan Mail
}

type Topic struct {
    Name        string
    Subscribers []Subscriber
    mu          sync.RWMutex
}

func (t *Topic) Publish(mail Mail) error {
    t.mu.RLock()
    defer t.mu.RUnlock()
    
    for _, sub := range t.Subscribers {
        ch := sub.Receive()
        select {
        case ch <- mail:
            // Delivered
        default:
            // Subscriber buffer full, skip
        }
    }
    return nil
}

func (t *Topic) Subscribe(sub Subscriber) error {
    t.mu.Lock()
    defer t.mu.Unlock()
    t.Subscribers = append(t.Subscribers, sub)
    return nil
}

func (t *Topic) Unsubscribe(sub Subscriber) error {
    t.mu.Lock()
    defer t.mu.Unlock()
    
    for i, s := range t.Subscribers {
        if s == sub {
            t.Subscribers = append(t.Subscribers[:i], t.Subscribers[i+1:]...)
            return nil
        }
    }
    
    return errors.New("subscriber not found")
}
```

## Tests to Write (3 tests, 3 commits)

### Test 1: TestTopic_SubscribeUnsubscribe
```go
func TestTopic_SubscribeUnsubscribe(t *testing.T) {
    topic := &Topic{Name: "test-topic"}
    
    // Create mock subscriber
    type mockSubscriber struct {
        ch chan Mail
    }
    func (m *mockSubscriber) Receive() <-chan Mail {
        return m.ch
    }
    
    sub1 := &mockSubscriber{ch: make(chan Mail, 10)}
    sub2 := &mockSubscriber{ch: make(chan Mail, 10)}
    
    // Subscribe first subscriber
    err := topic.Subscribe(sub1)
    if err != nil {
        t.Errorf("Expected nil error, got %v", err)
    }
    
    // Subscribe second subscriber
    err = topic.Subscribe(sub2)
    if err != nil {
        t.Errorf("Expected nil error, got %v", err)
    }
    
    // Unsubscribe first subscriber
    err = topic.Unsubscribe(sub1)
    if err != nil {
        t.Errorf("Expected nil error, got %v", err)
    }
    
    // Verify only sub2 remains
    topic.mu.RLock()
    if len(topic.Subscribers) != 1 {
        t.Errorf("Expected 1 subscriber, got %d", len(topic.Subscribers))
    }
    if topic.Subscribers[0] != sub2 {
        t.Error("Expected sub2 to remain")
    }
    topic.mu.RUnlock()
}
```
**Acceptance Criteria:**
- Subscribe adds subscriber
- Unsubscribe removes subscriber
- Returns nil error on success

### Test 2: TestTopic_Broadcast
```go
func TestTopic_Broadcast(t *testing.T) {
    topic := &Topic{Name: "market-data"}
    
    // Create mock subscribers
    type mockSubscriber struct {
        ch chan Mail
    }
    func (m *mockSubscriber) Receive() <-chan Mail {
        return m.ch
    }
    
    sub1 := &mockSubscriber{ch: make(chan Mail, 10)}
    sub2 := &mockSubscriber{ch: make(chan Mail, 10)}
    sub3 := &mockSubscriber{ch: make(chan Mail, 10)}
    
    topic.Subscribe(sub1)
    topic.Subscribe(sub2)
    topic.Subscribe(sub3)
    
    // Publish message
    mail := Mail{ID: "msg-001", Type: MailTypeAssistant}
    err := topic.Publish(mail)
    if err != nil {
        t.Errorf("Expected nil error, got %v", err)
    }
    
    // Verify all subscribers received the message
    for i, sub := range []*mockSubscriber{sub1, sub2, sub3} {
        select {
        case received := <-sub.ch:
            if received.ID != "msg-001" {
                t.Errorf("Subscriber %d: Expected msg-001, got %s", i, received.ID)
            }
        default:
            t.Errorf("Subscriber %d: Expected to receive message", i)
        }
    }
}
```
**Acceptance Criteria:**
- Publish delivers to all subscribers
- Each subscriber receives mail once
- Returns nil error on success

### Test 3: TestTopic_UnsubscribeNotFound
```go
func TestTopic_UnsubscribeNotFound(t *testing.T) {
    topic := &Topic{Name: "test-topic"}
    
    // Create subscriber that's not subscribed
    type mockSubscriber struct {
        ch chan Mail
    }
    func (m *mockSubscriber) Receive() <-chan Mail {
        return m.ch
    }
    
    unsubscribedSub := &mockSubscriber{ch: make(chan Mail, 10)}
    
    // Try to unsubscribe non-subscribed subscriber
    err := topic.Unsubscribe(unsubscribedSub)
    if err == nil {
        t.Error("Expected error for unsubscribing non-subscriber")
    }
    
    // Verify error message
    if err.Error() != "subscriber not found" {
        t.Errorf("Expected 'subscriber not found', got '%s'", err.Error())
    }
}
```
**Acceptance Criteria:**
- Unsubscribe non-existent subscriber returns error
- Error indicates subscriber not found

## Dependencies

### Test Dependencies
```
Test 1 → Test 2, Test 3 (Subscribe before Broadcast/Unsubscribe)
```

### Phase Dependencies
- **Phase 3.1** must complete first (Mail types)
- **Phase 3.2** depends on this phase (MailRouter needs topics)
- **Phase 3.3a** is sibling phase (AgentInbox, ServiceInbox)

## Files to Create/Modify

| File | Operation | Changes |
|------|-----------|---------|
| `pkg/mail/inbox.go` | MODIFY | Add Topic struct and methods |

## Implementation Order (TDD)

1. **Test 1**: Write test → verify RED → Implement Topic Subscribe/Unsubscribe → verify GREEN → commit
2. **Test 2**: Write test → verify RED → Implement Topic Publish → verify GREEN → commit
3. **Test 3**: Write test → verify RED → Add error handling to Unsubscribe → verify GREEN → commit

**Total: 3 tests, 3 commits**

## Deliverables
- ✅ Topic in `pkg/mail/inbox.go`
- ✅ Subscribe/Unsubscribe/Publish methods
- ✅ 3 commits (one per test per TDD workflow)

## Recursive Analysis Decision

**Decision: Split from original Phase 3.3**

**Reasoning:**
- Original Phase 3.3 had 6 tests across 3 independent inbox types
- Topic is fundamentally different from AgentInbox/ServiceInbox (pub-sub vs queue)
- 3 tests is within recommended range (2-5)
- Can be developed independently of 3.3a

**Benefits of split:**
- Each phase is now within 2-5 test range
- Independent features separated
- Can be developed in parallel with 3.3a
- Clear separation of concerns (queue vs pub-sub)
- Topic has different semantics (broadcast vs FIFO)