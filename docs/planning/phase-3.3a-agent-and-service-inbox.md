# Phase 3.3A: AgentInbox & ServiceInbox

## Goal
Implement AgentInbox and ServiceInbox data structures for message storage and delivery.

## Scope
- Create `pkg/mail/inbox.go` with inbox implementations
- Implement AgentInbox with Push/Pop/Subscribe
- Implement ServiceInbox (similar to AgentInbox)
- Channel-based pub/sub
- 3 tests, 3 commits

## Current Status

### Implementation Status
| Component | Status | Notes |
|-----------|--------|-------|
| `AgentInbox` | ❌ Missing | Agent message queue |
| `ServiceInbox` | ❌ Missing | System service queue |

### Files Status
| File | Status |
|------|-------|
| `pkg/mail/inbox.go` | ❌ MISSING - create |

## Required Implementation

### AgentInbox
```go
// pkg/mail/inbox.go
package mail

import (
    "errors"
    "sync"
)

type AgentInbox struct {
    ID       string
    Messages []Mail
    mu       sync.RWMutex
}

func (a *AgentInbox) Push(mail Mail) error {
    a.mu.Lock()
    defer a.mu.Unlock()
    a.Messages = append(a.Messages, mail)
    return nil
}

func (a *AgentInbox) Pop() (Mail, error) {
    a.mu.Lock()
    defer a.mu.Unlock()
    
    if len(a.Messages) == 0 {
        return Mail{}, errors.New("inbox is empty")
    }
    
    mail := a.Messages[0]
    a.Messages = a.Messages[1:]
    return mail, nil
}

func (a *AgentInbox) Subscribe() <-chan Mail {
    ch := make(chan Mail, 100)
    // In a real implementation, this would spawn a goroutine
    // that sends messages to the channel when they arrive
    return ch
}
```

### ServiceInbox
```go
type ServiceInbox struct {
    ID       string
    Messages []Mail
    mu       sync.RWMutex
}

func (s *ServiceInbox) Push(mail Mail) error {
    s.mu.Lock()
    defer s.mu.Unlock()
    s.Messages = append(s.Messages, mail)
    return nil
}

func (s *ServiceInbox) Pop() (Mail, error) {
    s.mu.Lock()
    defer s.mu.Unlock()
    
    if len(s.Messages) == 0 {
        return Mail{}, errors.New("inbox is empty")
    }
    
    mail := s.Messages[0]
    s.Messages = s.Messages[1:]
    return mail, nil
}
```

## Tests to Write (3 tests, 3 commits)

### Test 1: TestAgentInbox_PushPop
```go
func TestAgentInbox_PushPop(t *testing.T) {
    inbox := &AgentInbox{ID: "test-agent"}
    
    // Push first message
    mail1 := Mail{ID: "msg-001", Type: MailTypeUser}
    err := inbox.Push(mail1)
    if err != nil {
        t.Errorf("Expected nil error, got %v", err)
    }
    
    // Push second message
    mail2 := Mail{ID: "msg-002", Type: MailTypeAssistant}
    err = inbox.Push(mail2)
    if err != nil {
        t.Errorf("Expected nil error, got %v", err)
    }
    
    // Pop first message (FIFO)
    popped, err := inbox.Pop()
    if err != nil {
        t.Errorf("Expected nil error, got %v", err)
    }
    if popped.ID != "msg-001" {
        t.Errorf("Expected msg-001, got %s", popped.ID)
    }
    
    // Pop second message
    popped, err = inbox.Pop()
    if err != nil {
        t.Errorf("Expected nil error, got %v", err)
    }
    if popped.ID != "msg-002" {
        t.Errorf("Expected msg-002, got %s", popped.ID)
    }
    
    // Inbox should be empty now
    _, err = inbox.Pop()
    if err == nil {
        t.Error("Expected error for empty inbox")
    }
}
```
**Acceptance Criteria:**
- Messages pushed to inbox
- Messages popped in FIFO order
- Returns nil error on success
- Returns error when inbox is empty

### Test 2: TestAgentInbox_Subscribe
```go
func TestAgentInbox_Subscribe(t *testing.T) {
    inbox := &AgentInbox{ID: "test-agent"}
    
    ch := inbox.Subscribe()
    if ch == nil {
        t.Error("Expected non-nil channel")
    }
    
    // Verify channel is buffered (can send without blocking)
    select {
    case ch <- Mail{ID: "msg-001"}:
        // Success
    default:
        t.Error("Expected channel to accept message without blocking")
    }
}
```
**Acceptance Criteria:**
- Subscribe returns a channel
- Channel is non-nil
- Channel is buffered (non-blocking send)

### Test 3: TestServiceInbox_PushPop
```go
func TestServiceInbox_PushPop(t *testing.T) {
    inbox := &ServiceInbox{ID: "test-service"}
    
    // Push message
    mail := Mail{ID: "msg-001", Type: MailTypeHeartbeat}
    err := inbox.Push(mail)
    if err != nil {
        t.Errorf("Expected nil error, got %v", err)
    }
    
    // Pop message
    popped, err := inbox.Pop()
    if err != nil {
        t.Errorf("Expected nil error, got %v", err)
    }
    if popped.ID != "msg-001" {
        t.Errorf("Expected msg-001, got %s", popped.ID)
    }
    
    // Verify FIFO ordering with multiple messages
    inbox.Push(Mail{ID: "msg-002"})
    inbox.Push(Mail{ID: "msg-003"})
    
    popped, _ = inbox.Pop()
    if popped.ID != "msg-002" {
        t.Errorf("Expected msg-002, got %s", popped.ID)
    }
    
    popped, _ = inbox.Pop()
    if popped.ID != "msg-003" {
        t.Errorf("Expected msg-003, got %s", popped.ID)
    }
}
```
**Acceptance Criteria:**
- ServiceInbox Push/Pop works like AgentInbox
- FIFO ordering maintained
- Returns nil error on success

## Dependencies

### Test Dependencies
```
Test 1 → Test 2 (Push/Pop before Subscribe)
Test 3 (Independent - ServiceInbox)
```

### Phase Dependencies
- **Phase 3.1** must complete first (Mail types)
- **Phase 3.2** depends on this phase (MailRouter needs inboxes)
- **Phase 3.3b** is sibling phase (Topic implementation)

## Files to Create/Modify

| File | Operation | Changes |
|------|-----------|---------|
| `pkg/mail/inbox.go` | CREATE | AgentInbox, ServiceInbox (Part 1) |

## Implementation Order (TDD)

1. **Test 1**: Write test → verify RED → Implement AgentInbox Push/Pop → verify GREEN → commit
2. **Test 2**: Write test → verify RED → Implement AgentInbox Subscribe → verify GREEN → commit
3. **Test 3**: Write test → verify RED → Implement ServiceInbox → verify GREEN → commit

**Total: 3 tests, 3 commits**

## Deliverables
- ✅ AgentInbox in `pkg/mail/inbox.go`
- ✅ ServiceInbox in `pkg/mail/inbox.go`
- ✅ 3 commits (one per test per TDD workflow)

## Recursive Analysis Decision

**Decision: Split from original Phase 3.3**

**Reasoning:**
- Original Phase 3.3 had 6 tests across 3 independent inbox types
- AgentInbox and ServiceInbox are similar (both queues)
- Topic is different (pub-sub vs queue)
- Splitting creates more focused phases:
  - 3.3a: AgentInbox + ServiceInbox (3 tests)
  - 3.3b: Topic (3 tests)

**Benefits of split:**
- Each phase is now within 2-5 test range
- Independent features separated
- Can be developed in parallel
- Clear separation of concerns (queue vs pub-sub)