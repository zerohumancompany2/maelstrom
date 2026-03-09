# Phase 3.2: Mail Router

## Goal
Implement the MailRouter that routes messages to agent inboxes, topics, or system services.

## Scope
- Create `pkg/mail/router.go` with MailRouter struct
- Implement Route() method
- Implement SubscribeAgent, SubscribeTopic, SubscribeService methods
- Thread-safe routing with mutex protection
- 5 tests, 5 commits

## Current Status

### Implementation Status
| Component | Status | Notes |
|-----------|--------|-------|
| `MailRouter` | ❌ Missing | Core routing logic |
| `Route()` | ❌ Missing | Route to agent/topic/service |
| `SubscribeAgent()` | ❌ Missing | Subscribe to agent inbox |
| `SubscribeTopic()` | ❌ Missing | Subscribe to topic |
| `SubscribeService()` | ❌ Missing | Subscribe to sys service |

### Files Status
| File | Status |
|------|-------|
| `pkg/mail/router.go` | ❌ MISSING - create |

## Required Implementation

### MailRouter
```go
// pkg/mail/router.go
package mail

import (
    "errors"
    "sync"
)

type AddressType int

const (
    AddressTypeUnknown AddressType = iota
    AddressTypeAgent
    AddressTypeTopic
    AddressTypeSys
)

type MailRouter struct {
    agents   map[string]*AgentInbox
    topics   map[string]*Topic
    services map[string]*ServiceInbox
    mu       sync.RWMutex
}

func NewMailRouter() *MailRouter {
    return &MailRouter{
        agents:   make(map[string]*AgentInbox),
        topics:   make(map[string]*Topic),
        services: make(map[string]*ServiceInbox),
    }
}

func (r *MailRouter) Route(mail Mail) error {
    addrType, id, err := ParseAddress(mail.Target)
    if err != nil {
        return err
    }
    
    r.mu.Lock()
    defer r.mu.Unlock()
    
    switch addrType {
    case AddressTypeAgent:
        inbox, exists := r.agents[id]
        if !exists {
            return errors.New("agent not found: " + id)
        }
        return inbox.Push(mail)
    case AddressTypeTopic:
        topic, exists := r.topics[id]
        if !exists {
            return errors.New("topic not found: " + id)
        }
        return topic.Publish(mail)
    case AddressTypeSys:
        inbox, exists := r.services[id]
        if !exists {
            return errors.New("service not found: " + id)
        }
        return inbox.Push(mail)
    default:
        return errors.New("unknown address type")
    }
}

func (r *MailRouter) SubscribeAgent(id string, inbox *AgentInbox) error {
    r.mu.Lock()
    defer r.mu.Unlock()
    r.agents[id] = inbox
    return nil
}

func (r *MailRouter) SubscribeTopic(name string, topic *Topic) error {
    r.mu.Lock()
    defer r.mu.Unlock()
    r.topics[name] = topic
    return nil
}

func (r *MailRouter) SubscribeService(name string, inbox *ServiceInbox) error {
    r.mu.Lock()
    defer r.mu.Unlock()
    r.services[name] = inbox
    return nil
}
```

## Tests to Write (5 tests, 5 commits)

### Test 1: TestMailRouter_SubscribeMethods
```go
func TestMailRouter_SubscribeMethods(t *testing.T) {
    router := NewMailRouter()
    
    // Test SubscribeAgent
    agentInbox := &AgentInbox{ID: "test-agent"}
    err := router.SubscribeAgent("test-agent", agentInbox)
    if err != nil {
        t.Errorf("Expected nil error, got %v", err)
    }
    
    // Test SubscribeTopic
    topic := &Topic{Name: "test-topic"}
    err = router.SubscribeTopic("test-topic", topic)
    if err != nil {
        t.Errorf("Expected nil error, got %v", err)
    }
    
    // Test SubscribeService
    serviceInbox := &ServiceInbox{ID: "test-service"}
    err = router.SubscribeService("test-service", serviceInbox)
    if err != nil {
        t.Errorf("Expected nil error, got %v", err)
    }
}
```
**Acceptance Criteria:**
- SubscribeAgent registers agent inbox
- SubscribeTopic registers topic subscribers
- SubscribeService registers service inbox

### Test 2: TestMailRouter_RouteToAgent
```go
func TestMailRouter_RouteToAgent(t *testing.T) {
    router := NewMailRouter()
    
    agentInbox := &AgentInbox{ID: "recommendation-agent"}
    router.SubscribeAgent("recommendation-agent", agentInbox)
    
    mail := Mail{
        ID:     "msg-001",
        Source: "agent:user-agent",
        Target: "agent:recommendation-agent",
        Type:   MailTypeUser,
    }
    
    err := router.Route(mail)
    if err != nil {
        t.Errorf("Expected nil error, got %v", err)
    }
    
    // Verify message was pushed to inbox
    agentInbox.mu.RLock()
    if len(agentInbox.Messages) != 1 {
        t.Errorf("Expected 1 message in inbox, got %d", len(agentInbox.Messages))
    }
    agentInbox.mu.RUnlock()
}
```
**Acceptance Criteria:**
- Mail to `agent:<id>` delivered to correct inbox
- Returns nil error on success
- Agent must be registered first

### Test 3: TestMailRouter_RouteToTopic
```go
func TestMailRouter_RouteToTopic(t *testing.T) {
    router := NewMailRouter()
    
    topic := &Topic{Name: "market-data"}
    router.SubscribeTopic("market-data", topic)
    
    mail := Mail{
        ID:     "msg-002",
        Source: "sys:data-service",
        Target: "topic:market-data",
        Type:   MailTypeAssistant,
    }
    
    err := router.Route(mail)
    if err != nil {
        t.Errorf("Expected nil error, got %v", err)
    }
}
```
**Acceptance Criteria:**
- Mail to `topic:<name>` broadcast to all subscribers
- Multiple subscribers receive same mail
- Returns nil error on success

### Test 4: TestMailRouter_RouteToService
```go
func TestMailRouter_RouteToService(t *testing.T) {
    router := NewMailRouter()
    
    serviceInbox := &ServiceInbox{ID: "heartbeat"}
    router.SubscribeService("heartbeat", serviceInbox)
    
    mail := Mail{
        ID:     "msg-003",
        Source: "agent:scheduler",
        Target: "sys:heartbeat",
        Type:   MailTypeHeartbeat,
    }
    
    err := router.Route(mail)
    if err != nil {
        t.Errorf("Expected nil error, got %v", err)
    }
    
    // Verify message was pushed to service inbox
    serviceInbox.mu.RLock()
    if len(serviceInbox.Messages) != 1 {
        t.Errorf("Expected 1 message in service inbox, got %d", len(serviceInbox.Messages))
    }
    serviceInbox.mu.RUnlock()
}
```
**Acceptance Criteria:**
- Mail to `sys:<service>` delivered to service inbox
- Returns nil error on success
- Service must be registered first

### Test 5: TestMailRouter_RouteToUnknownAddress
```go
func TestMailRouter_RouteToUnknownAddress(t *testing.T) {
    router := NewMailRouter()
    
    // Test route to unregistered agent
    mail := Mail{
        ID:     "msg-004",
        Source: "agent:user-agent",
        Target: "agent:non-existent",
        Type:   MailTypeUser,
    }
    
    err := router.Route(mail)
    if err == nil {
        t.Error("Expected error for unregistered agent")
    }
    
    // Test route to unregistered topic
    mail.Target = "topic:non-existent"
    err = router.Route(mail)
    if err == nil {
        t.Error("Expected error for unregistered topic")
    }
    
    // Test route to unregistered service
    mail.Target = "sys:non-existent"
    err = router.Route(mail)
    if err == nil {
        t.Error("Expected error for unregistered service")
    }
    
    // Test route to invalid address format
    mail.Target = "invalid-format"
    err = router.Route(mail)
    if err == nil {
        t.Error("Expected error for invalid address format")
    }
}
```
**Acceptance Criteria:**
- Route to unregistered agent returns error
- Route to unregistered topic returns error
- Route to unregistered service returns error
- Route to invalid address format returns error

## Dependencies

### Test Dependencies
```
Test 1 → Test 2, Test 3, Test 4 (Subscribe must work first)
Test 5 (Error handling - independent)
```

### Phase Dependencies
- **Phase 3.1** must complete first (Mail types, address parsing)
- **Phase 3.3** depends on this phase (Inboxes implementation)

## Files to Create/Modify

| File | Operation | Changes |
|------|-----------|---------|
| `pkg/mail/router.go` | CREATE | MailRouter struct and methods |

## Implementation Order (TDD)

1. **Test 1**: Write test → verify RED → Implement SubscribeAgent/Topic/Service → verify GREEN → commit
2. **Test 2**: Write test → verify RED → Implement Route() for agent addresses → verify GREEN → commit
3. **Test 3**: Write test → verify RED → Implement Route() for topic addresses → verify GREEN → commit
4. **Test 4**: Write test → verify RED → Implement Route() for sys addresses → verify GREEN → commit
5. **Test 5**: Write test → verify RED → Add error handling for unknown addresses → verify GREEN → commit

**Total: 5 tests, 5 commits**

## Deliverables
- ✅ MailRouter in `pkg/mail/router.go`
- ✅ Route() method for all address types
- ✅ SubscribeAgent, SubscribeTopic, SubscribeService methods
- ✅ 5 commits (one per test per TDD workflow)

## Recursive Analysis Decision

**Decision: Keep as-is (no split)**

**Reasoning:**
- 5 tests is within recommended range (2-5)
- All tests are tightly coupled around routing functionality
- Single coherent feature: Mail routing
- Splitting would create unnecessary fragmentation

**Alternative (if split needed):**
- 3.2a: Router structure + Subscribe methods - 1 test
- 3.2b: Route to Agent/Topic/Service - 3 tests
- 3.2c: Error handling - 1 test