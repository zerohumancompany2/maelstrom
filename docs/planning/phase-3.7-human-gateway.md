# Phase 3.7: Human Gateway Service

## Goal
Implement sys:human-gateway service for human chat with running agents.

## Scope
- Create `pkg/services/humangateway/service.go`
- Implement HTTPS chat session at `/chat/{agentId}`
- Session receives sanitized ContextMap snapshot
- Messages sent as human_feedback or user mail type
- Optional action item shorthand (@pause, @inject-memory)
- 4 tests, 4 commits

## Current Status

### Implementation Status
| Component | Status | Notes |
|-----------|--------|-------|
| `HumanGatewayService` | ❌ Missing | Chat service |
| Chat session handling | ❌ Missing | HTTPS endpoint |
| Action item parsing | ❌ Missing | @pause, @inject-memory |

### Files Status
| File | Status |
|------|-------|
| `pkg/services/humangateway/service.go` | ❌ MISSING - create |
| `pkg/services/humangateway/chat.go` | ❌ MISSING - create |

## Required Implementation

### HumanGatewayService
```go
// pkg/services/humangateway/service.go
package humangateway

import (
    "maelstrom/pkg/mail"
    "sync"
)

type HumanGatewayService struct {
    id       string
    sessions map[string]*ChatSession
    mu       sync.RWMutex
}

func NewHumanGatewayService() *HumanGatewayService {
    return &HumanGatewayService{
        id:       "sys:human-gateway",
        sessions: make(map[string]*ChatSession),
    }
}

func (h *HumanGatewayService) ID() string {
    return h.id
}

func (h *HumanGatewayService) HandleChat(agentID, message string) (mail.Mail, error) {
    // Parse for action items
    actionItems, _ := h.ParseActionItem(message)
    
    // Create mail
    mailType := mail.MailTypeUser
    if len(actionItems) > 0 {
        mailType = mail.MailTypeHumanFeedback
    }
    
    return mail.Mail{
        ID:     generateID(),
        Type:   mailType,
        Source: "human:" + agentID,
        Target: "agent:" + agentID,
        Content: map[string]any{
            "message":     message,
            "actionItems": actionItems,
        },
        Metadata: mail.MailMetadata{
            Boundary: mail.InnerBoundary,
            Taints:   []string{"USER_SUPPLIED"},
        },
    }, nil
}

func (h *HumanGatewayService) GetSession(agentID string) *ChatSession {
    h.mu.RLock()
    defer h.mu.RUnlock()
    return h.sessions[agentID]
}

func (h *HumanGatewayService) CreateSession(agentID string) *ChatSession {
    h.mu.Lock()
    defer h.mu.Unlock()
    
    session := &ChatSession{
        AgentID:    agentID,
        Messages:   make([]mail.Mail, 0),
        ContextMap: make(ContextMapSnapshot),
    }
    h.sessions[agentID] = session
    return session
}
```

### Chat and Action Items
```go
// pkg/services/humangateway/chat.go
package humangateway

import (
    "maelstrom/pkg/mail"
    "strings"
)

type ChatSession struct {
    AgentID    string
    Messages   []mail.Mail
    ContextMap ContextMapSnapshot
}

type ContextMapSnapshot map[string]any

type ActionItem struct {
    Type    string
    Target  string
    Payload any
}

func (h *HumanGatewayService) ParseActionItem(message string) ([]ActionItem, error) {
    var items []ActionItem
    
    // Check for @pause
    if strings.Contains(message, "@pause") {
        items = append(items, ActionItem{
            Type:   "pause",
            Target: "",
            Payload: nil,
        })
    }
    
    // Check for @inject-memory <content>
    if idx := strings.Index(message, "@inject-memory"); idx >= 0 {
        content := strings.TrimSpace(message[idx+16:])
        items = append(items, ActionItem{
            Type:   "inject-memory",
            Target: "",
            Payload: content,
        })
    }
    
    return items, nil
}

func SanitizeContextForBoundary(ctx ContextMapSnapshot, boundary mail.BoundaryType) ContextMapSnapshot {
    // For Phase 1, return as-is
    // Future: apply boundary rules to strip forbidden data
    sanitized := make(ContextMapSnapshot)
    for k, v := range ctx {
        sanitized[k] = v
    }
    return sanitized
}
```

## Tests to Write (4 tests, 4 commits)

### Test 1: TestHumanGatewayService_HandleChat
```go
func TestHumanGatewayService_HandleChat(t *testing.T) {
    svc := NewHumanGatewayService()
    
    mail, err := svc.HandleChat("recommendation-agent", "What do you think?")
    if err != nil {
        t.Errorf("Expected nil error, got %v", err)
    }
    
    if mail.Type != mail.MailTypeUser {
        t.Errorf("Expected MailTypeUser, got %s", mail.Type)
    }
    
    if mail.Source != "human:recommendation-agent" {
        t.Errorf("Expected source 'human:recommendation-agent', got '%s'", mail.Source)
    }
    
    if mail.Target != "agent:recommendation-agent" {
        t.Errorf("Expected target 'agent:recommendation-agent', got '%s'", mail.Target)
    }
    
    content, ok := mail.Content.(map[string]any)
    if !ok {
        t.Error("Expected content to be map[string]any")
    }
    if content["message"] != "What do you think?" {
        t.Error("Expected message preserved in content")
    }
}
```
**Acceptance Criteria:**
- HandleChat creates mail for agent
- Mail type is human_feedback or user
- Returns nil error on success
- Source set to human:{agentID}
- Target set to agent:{agentID}

### Test 2: TestHumanGatewayService_ParseActionItem
```go
func TestHumanGatewayService_ParseActionItem(t *testing.T) {
    svc := NewHumanGatewayService()
    
    // Test @pause
    items, err := svc.ParseActionItem("Please @pause processing")
    if err != nil {
        t.Errorf("Expected nil error, got %v", err)
    }
    
    if len(items) != 1 {
        t.Errorf("Expected 1 action item, got %d", len(items))
    }
    if items[0].Type != "pause" {
        t.Errorf("Expected type 'pause', got '%s'", items[0].Type)
    }
    
    // Test @inject-memory
    items, err = svc.ParseActionItem("@inject-memory This is important")
    if err != nil {
        t.Errorf("Expected nil error, got %v", err)
    }
    
    if len(items) != 1 {
        t.Errorf("Expected 1 action item, got %d", len(items))
    }
    if items[0].Type != "inject-memory" {
        t.Errorf("Expected type 'inject-memory', got '%s'", items[0].Type)
    }
    if items[0].Payload != "This is important" {
        t.Errorf("Expected payload 'This is important', got '%v'", items[0].Payload)
    }
    
    // Test no action items
    items, err = svc.ParseActionItem("Just a normal message")
    if err != nil {
        t.Errorf("Expected nil error, got %v", err)
    }
    if len(items) != 0 {
        t.Errorf("Expected 0 action items, got %d", len(items))
    }
}
```
**Acceptance Criteria:**
- @pause parsed correctly
- @inject-memory X parsed correctly with content
- Returns ActionItem slice
- No action items returns empty slice

### Test 3: TestHumanGatewayService_SessionManagement
```go
func TestHumanGatewayService_SessionManagement(t *testing.T) {
    svc := NewHumanGatewayService()
    
    // Create session
    session := svc.CreateSession("test-agent")
    if session == nil {
        t.Error("Expected non-nil session")
    }
    
    if session.AgentID != "test-agent" {
        t.Errorf("Expected AgentID 'test-agent', got '%s'", session.AgentID)
    }
    
    // Get session
    retrieved := svc.GetSession("test-agent")
    if retrieved == nil {
        t.Error("Expected non-nil retrieved session")
    }
    if retrieved != session {
        t.Error("Expected same session instance")
    }
    
    // Get non-existent session
    missing := svc.GetSession("non-existent")
    if missing != nil {
        t.Error("Expected nil for non-existent session")
    }
}
```
**Acceptance Criteria:**
- Chat sessions created per agent
- Sessions store agent ID
- GetSession retrieves existing session
- GetSession returns nil for non-existent session

### Test 4: TestHumanGatewayService_SanitizedContext
```go
func TestHumanGatewayService_SanitizedContext(t *testing.T) {
    ctx := ContextMapSnapshot{
        "conversation": []any{"msg1", "msg2"},
        "memory":       "important data",
        "internal":     "secret data",
    }
    
    // Test sanitization for inner boundary
    sanitized := SanitizeContextForBoundary(ctx, mail.InnerBoundary)
    if len(sanitized) != len(ctx) {
        t.Errorf("Expected %d keys, got %d", len(ctx), len(sanitized))
    }
    
    // Test sanitization for outer boundary
    sanitizedOuter := SanitizeContextForBoundary(ctx, mail.OuterBoundary)
    if len(sanitizedOuter) != len(ctx) {
        t.Errorf("Expected %d keys, got %d", len(ctx), len(sanitizedOuter))
    }
    
    // Phase 1: all data passes through
    // Future: outer boundary should strip internal data
}
```
**Acceptance Criteria:**
- ContextMap snapshot is sanitized
- Boundary rules applied (Phase 1: pass-through)
- Forbidden taints stripped (future)

## Dependencies

### Test Dependencies
```
Test 1 → Test 3 (HandleChat before SessionManagement)
Test 2 (Independent - ActionItem parsing)
Test 4 (Depends on 3 - Context sanitization)
```

### Phase Dependencies
- **Phase 3.1** must complete first (Mail types)
- **Phase 3.6** must complete first (Gateway adapters)

## Files to Create/Modify

| File | Operation | Changes |
|------|-----------|---------|
| `pkg/services/humangateway/service.go` | CREATE | HumanGatewayService |
| `pkg/services/humangateway/chat.go` | CREATE | ChatSession and action items |

## Implementation Order (TDD)

1. **Test 1**: Write test → verify RED → Implement HandleChat method → verify GREEN → commit
2. **Test 2**: Write test → verify RED → Implement ParseActionItem method → verify GREEN → commit
3. **Test 3**: Write test → verify RED → Implement session management → verify GREEN → commit
4. **Test 4**: Write test → verify RED → Implement context sanitization → verify GREEN → commit

**Total: 4 tests, 4 commits**

## Deliverables
- ✅ HumanGatewayService in `pkg/services/humangateway/service.go`
- ✅ ChatSession handling in `pkg/services/humangateway/chat.go`
- ✅ Action item parsing (@pause, @inject-memory)
- ✅ Context sanitization (Phase 1: pass-through)
- ✅ 4 commits (one per test per TDD workflow)

## Recursive Analysis Decision

**Decision: Keep as-is (no split)**

**Reasoning:**
- 4 tests is within recommended range (2-5)
- Tests are tightly coupled around human gateway functionality
- Single coherent feature: Human chat with agents
- Splitting would create unnecessary fragmentation

**Alternative (if split needed):**
- 3.7a: HandleChat + ParseActionItem - 2 tests
- 3.7b: SessionManagement + SanitizedContext - 2 tests