# Phase 8.4: Human Chat Completion

## Goal
Implement complete human-in-the-loop chat interface with running agents per spec (arch-v1.md L725-736).

## Scope
- Implement HTTPS chat session at `/chat/{agentId}` (arch-v1.md L728)
- Implement read-only ContextMap snapshot for sessions (arch-v1.md L731)
- Implement last N messages sanitization by boundary rules (arch-v1.md L731)
- Implement human messages as mail_received (arch-v1.md L732)
- Implement agent replies via normal mail flow (arch-v1.md L733)
- Implement action item shorthands (arch-v1.md L734)
- 6 tests, 6 commits

## Current Status

### Implementation Status
| Component | Status | Notes |
|-----------|--------|-------|
| `ChatSession` | ❌ Missing | Need to implement chat session management |
| `ContextMap snapshot` | ⚠️ Partial | ContextMap exists; need read-only snapshot |
| `Message sanitization` | ❌ Missing | Need to implement boundary-based sanitization |
| `Human feedback mail` | ❌ Missing | Need to convert chat messages to mail_received |
| `Action item parser` | ❌ Missing | Need to implement @pause, @inject-memory parsing |

### Files Status
| File | Status |
|------|--------|
| `pkg/services/gateway/chat.go` | ❌ Create - chat session management |
| `pkg/services/gateway/chat_test.go` | ❌ Create - chat tests |
| `pkg/mail/types.go` | ⚠️ Modify - add chat-related types |
| `pkg/security/contextmap.go` | ⚠️ Modify - add read-only snapshot |

## Required Implementation

### HTTPS Chat Endpoint (arch-v1.md L728)
```go
// pkg/services/gateway/chat.go
type ChatSession struct {
    AgentID    string
    SessionID  string
    ContextMap *ContextMapSnapshot
    Messages   []ChatMessage
}

func (s *GatewayService) CreateChatSession(agentID string) (*ChatSession, error) {
    // Create chat session for long-running top-level Agent
    // Session at /chat/{agentId}
}
```

### Read-Only ContextMap Snapshot (arch-v1.md L731)
```go
// pkg/security/contextmap.go
type ContextMapSnapshot struct {
    Blocks     []*ContextBlock
    TokenCount int
    ReadOnly   bool
}

func (cm *ContextMap) Snapshot() *ContextMapSnapshot {
    // Create read-only snapshot of current ContextMap
    // Session receives this snapshot
}
```

### Last N Messages Sanitization (arch-v1.md L731)
```go
// pkg/services/gateway/chat.go
func (s *ChatSession) GetLastNMessages(n int) []ChatMessage {
    // Return last N messages sanitized by boundary rules
    // Remove SECRET, PII, and inner-boundary content
}
```

### Human Messages as Mail Received (arch-v1.md L732)
```go
// pkg/services/gateway/chat.go
func (s *GatewayService) SendHumanMessage(agentID string, message string) (*mail.Mail, error) {
    // Create mail_received (type: human_feedback or user)
    // Deliver to Agent's inbox via CommunicationService
}
```

### Agent Replies via Normal Mail Flow (arch-v1.md L733)
```go
// pkg/services/gateway/chat.go
func (s *GatewayService) RenderAgentReply(mail *mail.Mail) ChatMessage {
    // Agent replies via normal mail → rendered back in chat UI
    // assistant or partial_assistant mail types
}
```

### Action Item Shorthand (arch-v1.md L734)
```go
// pkg/services/gateway/chat.go
type ActionItem struct {
    Type    string // pause, inject-memory
    Payload any
}

func (s *GatewayService) ParseActionItem(message string) (*ActionItem, error) {
    // Parse @pause, @inject-memory X, etc.
    // Become special Mail messages
}
```

## Tests to Write (6 tests, 6 commits)

### Test 1: TestHumanChat_HTTPSEndpointAgentId
```go
func TestHumanChat_HTTPSEndpointAgentId(t *testing.T) {
    svc := NewGatewayService()
    agentID := "agent:dmz"
    
    // Create chat session at /chat/{agentId} (arch-v1.md L728)
    session, err := svc.CreateChatSession(agentID)
    if err != nil {
        t.Fatalf("Expected no error creating chat session, got %v", err)
    }
    
    // Verify session created (arch-v1.md L728)
    if session == nil {
        t.Error("Expected session to be created")
    }
    
    // Verify agent ID matches (arch-v1.md L728)
    if session.AgentID != agentID {
        t.Errorf("Expected agentID '%s', got '%s'", agentID, session.AgentID)
    }
    
    // Verify session ID generated
    if session.SessionID == "" {
        t.Error("Expected session ID to be generated")
    }
    
    // Verify HTTPS endpoint path format (arch-v1.md L728)
    expectedPath := "/chat/" + agentID
    actualPath := svc.GetChatPath(agentID)
    if actualPath != expectedPath {
        t.Errorf("Expected path '%s', got '%s'", expectedPath, actualPath)
    }
    
    // Verify session is for long-running top-level Agent (arch-v1.md L728)
    if session.SessionType != "chat" {
        t.Errorf("Expected session type 'chat', got '%s'", session.SessionType)
    }
}
```
**Acceptance Criteria:**
- HTTPS chat session created at `/chat/{agentId}` (arch-v1.md L728)
- Session has agentID, sessionID, and session type
- Endpoint path format correct

### Test 2: TestHumanChat_ReadOnlyContextMapSnapshot
```go
func TestHumanChat_ReadOnlyContextMapSnapshot(t *testing.T) {
    svc := NewGatewayService()
    agentID := "agent:dmz"
    
    // Setup: Create ContextMap with blocks (arch-v1.md L731)
    blocks := []*ContextBlock{
        {Name: "session", Content: "Current session data", Priority: 1},
        {Name: "memory", Content: "Long-term memory", Priority: 2},
    }
    contextMap := NewContextMap(blocks, 4096)
    
    // Session receives read-only snapshot (arch-v1.md L731)
    snapshot := contextMap.Snapshot()
    if snapshot == nil {
        t.Fatal("Expected snapshot to be created")
    }
    
    // Verify snapshot is read-only (arch-v1.md L731)
    if !snapshot.ReadOnly {
        t.Error("Expected snapshot to be read-only")
    }
    
    // Verify snapshot contains blocks (arch-v1.md L731)
    if len(snapshot.Blocks) != 2 {
        t.Errorf("Expected 2 blocks, got %d", len(snapshot.Blocks))
    }
    
    // Verify snapshot has token count
    if snapshot.TokenCount == 0 {
        t.Error("Expected snapshot to have token count")
    }
    
    // Verify creating chat session receives snapshot (arch-v1.md L731)
    session, err := svc.CreateChatSession(agentID)
    if err != nil {
        t.Fatalf("Expected no error creating session, got %v", err)
    }
    
    if session.ContextMap == nil {
        t.Error("Expected session to have ContextMap snapshot")
    }
    
    if !session.ContextMap.ReadOnly {
        t.Error("Expected session ContextMap to be read-only")
    }
    
    // Verify snapshot is independent (modifying original doesn't affect snapshot)
    originalCount := len(snapshot.Blocks)
    blocks[0].Content = "Modified content"
    if len(snapshot.Blocks) != originalCount {
        t.Error("Expected snapshot to be independent of original")
    }
}
```
**Acceptance Criteria:**
- Session receives read-only snapshot of Agent's ContextMap (arch-v1.md L731)
- Snapshot has ReadOnly flag set to true
- Snapshot contains blocks and token count
- Snapshot is independent of original ContextMap

### Test 3: TestHumanChat_LastNMessagesSanitized
```go
func TestHumanChat_LastNMessagesSanitized(t *testing.T) {
    svc := NewGatewayService()
    agentID := "agent:dmz"
    
    // Setup: Create session with messages (arch-v1.md L731)
    session, err := svc.CreateChatSession(agentID)
    if err != nil {
        t.Fatalf("Expected no error creating session, got %v", err)
    }
    
    // Add messages with various taints and boundaries (arch-v1.md L731)
    messages := []ChatMessage{
        {ID: "msg1", Content: "Normal message", Taints: []string{}, Boundary: "outer", Type: "user"},
        {ID: "msg2", Content: "Secret message", Taints: []string{"SECRET"}, Boundary: "inner", Type: "assistant"},
        {ID: "msg3", Content: "PII message", Taints: []string{"PII"}, Boundary: "dmz", Type: "user"},
        {ID: "msg4", Content: "Tool output", Taints: []string{"TOOL_OUTPUT"}, Boundary: "outer", Type: "assistant"},
        {ID: "msg5", Content: "Clean message", Taints: []string{}, Boundary: "outer", Type: "user"},
    }
    
    for _, msg := range messages {
        session.Messages = append(session.Messages, msg)
    }
    
    // Get last N messages sanitized by boundary rules (arch-v1.md L731)
    sanitized := session.GetLastNMessages(3)
    
    // Verify only 3 messages returned (last N)
    if len(sanitized) != 3 {
        t.Errorf("Expected 3 messages, got %d", len(sanitized))
    }
    
    // Verify SECRET taint removed (arch-v1.md L731)
    for _, msg := range sanitized {
        if slices.Contains(msg.Taints, "SECRET") {
            t.Error("Expected SECRET taint to be sanitized")
        }
    }
    
    // Verify PII taint removed (arch-v1.md L731)
    for _, msg := range sanitized {
        if slices.Contains(msg.Taints, "PII") {
            t.Error("Expected PII taint to be sanitized")
        }
    }
    
    // Verify inner-boundary messages removed (arch-v1.md L731)
    for _, msg := range sanitized {
        if msg.Boundary == "inner" {
            t.Error("Expected inner-boundary messages to be sanitized")
        }
    }
    
    // Verify TOOL_OUTPUT taint preserved (allowed)
    hasToolOutput := false
    for _, msg := range sanitized {
        if slices.Contains(msg.Taints, "TOOL_OUTPUT") {
            hasToolOutput = true
            break
        }
    }
    if !hasToolOutput {
        t.Error("Expected TOOL_OUTPUT taint to be preserved")
    }
    
    // Verify clean messages preserved
    hasClean := false
    for _, msg := range sanitized {
        if msg.Content == "Clean message" {
            hasClean = true
            break
        }
    }
    if !hasClean {
        t.Error("Expected clean message to be preserved")
    }
}
```
**Acceptance Criteria:**
- Last N messages returned (arch-v1.md L731)
- SECRET/PII taints sanitized/removed (arch-v1.md L731)
- Inner-boundary messages removed (arch-v1.md L731)
- Allowed taints (TOOL_OUTPUT) preserved

### Test 4: TestHumanChat_MessageBecomesMailReceived
```go
func TestHumanChat_MessageBecomesMailReceived(t *testing.T) {
    svc := NewGatewayService()
    agentID := "agent:dmz"
    
    // Human message sent (arch-v1.md L732)
    humanMessage := "Hello, agent! How can you help me?"
    mail, err := svc.SendHumanMessage(agentID, humanMessage)
    if err != nil {
        t.Fatalf("Expected no error sending human message, got %v", err)
    }
    
    // Verify mail type is mail_received (arch-v1.md L732)
    if mail.Type != mail.MailTypeMailReceived {
        t.Errorf("Expected mail type 'mail_received', got '%s'", mail.Type)
    }
    
    // Verify mail subtype is human_feedback (arch-v1.md L732)
    if mail.Metadata.HumanFeedbackType != "human_feedback" {
        t.Errorf("Expected human_feedback type, got '%s'", mail.Metadata.HumanFeedbackType)
    }
    
    // Verify delivered to Agent's inbox (arch-v1.md L732)
    if mail.Target != agentID {
        t.Errorf("Expected target '%s', got '%s'", agentID, mail.Target)
    }
    
    // Verify content matches human message
    if mail.Content != humanMessage {
        t.Errorf("Expected content '%s', got '%s'", humanMessage, mail.Content)
    }
    
    // Verify taints include USER_SUPPLIED
    if !slices.Contains(mail.Metadata.Taints, "USER_SUPPLIED") {
        t.Error("Expected USER_SUPPLIED taint")
    }
    
    // Verify boundary is outer (human input)
    if mail.Metadata.Boundary != mail.OuterBoundary {
        t.Errorf("Expected boundary 'outer', got '%s'", mail.Metadata.Boundary)
    }
    
    // Verify mail has ID and timestamp
    if mail.ID == "" {
        t.Error("Expected mail to have ID")
    }
    if mail.CreatedAt.IsZero() {
        t.Error("Expected mail to have timestamp")
    }
}
```
**Acceptance Criteria:**
- Messages become mail_received (arch-v1.md L732)
- Type is human_feedback (arch-v1.md L732)
- Delivered to Agent's inbox via CommunicationService
- Has USER_SUPPLIED taint and outer boundary

### Test 5: TestHumanChat_AgentRepliesViaMail
```go
func TestHumanChat_AgentRepliesViaMail(t *testing.T) {
    svc := NewGatewayService()
    agentID := "agent:dmz"
    
    // Setup: Create session
    session, err := svc.CreateChatSession(agentID)
    if err != nil {
        t.Fatalf("Expected no error creating session, got %v", err)
    }
    
    // Agent replies via normal mail flow (arch-v1.md L733)
    agentMail := &mail.Mail{
        ID:       "mail-001",
        Type:     mail.MailTypeAssistant,
        Source:   agentID,
        Target:   "user",
        Content:  "I can help you with that request.",
        Metadata: mail.MailMetadata{
            Boundary: mail.OuterBoundary,
            Taints:   []string{},
        },
    }
    
    // Render back in chat UI (arch-v1.md L733)
    chatMessage := svc.RenderAgentReply(agentMail)
    
    // Verify chat message created (arch-v1.md L733)
    if chatMessage == (ChatMessage{}) {
        t.Error("Expected chat message to be rendered")
    }
    
    // Verify content matches mail content (arch-v1.md L733)
    if chatMessage.Content != agentMail.Content {
        t.Errorf("Expected content '%s', got '%s'", agentMail.Content, chatMessage.Content)
    }
    
    // Verify type is assistant (arch-v1.md L733)
    if chatMessage.Type != "assistant" {
        t.Errorf("Expected type 'assistant', got '%s'", chatMessage.Type)
    }
    
    // Verify source is agent
    if chatMessage.Source != agentID {
        t.Errorf("Expected source '%s', got '%s'", agentID, chatMessage.Source)
    }
    
    // Test partial_assistant mail type (arch-v1.md L733)
    partialMail := &mail.Mail{
        ID:    "mail-002",
        Type:  mail.MailTypePartialAssistant,
        Source: agentID,
        Content: "This is a partial response... ",
        Metadata: mail.MailMetadata{
            Boundary: mail.OuterBoundary,
            Stream:   true,
        },
    }
    
    partialMessage := svc.RenderAgentReply(partialMail)
    
    // Verify partial message rendered (arch-v1.md L733)
    if partialMessage.Type != "assistant" {
        t.Errorf("Expected type 'assistant' for partial, got '%s'", partialMessage.Type)
    }
    
    if partialMessage.IsPartial != true {
        t.Error("Expected IsPartial to be true")
    }
    
    // Verify message added to session
    session.Messages = append(session.Messages, chatMessage)
    if len(session.Messages) != 1 {
        t.Errorf("Expected 1 message in session, got %d", len(session.Messages))
    }
}
```
**Acceptance Criteria:**
- Agent replies via normal mail flow (arch-v1.md L733)
- Mail rendered back in chat UI (arch-v1.md L733)
- Both assistant and partial_assistant mail types supported
- Messages added to session

### Test 6: TestHumanChat_ActionItemShorthand
```go
func TestHumanChat_ActionItemShorthand(t *testing.T) {
    svc := NewGatewayService()
    agentID := "agent:dmz"
    
    // Test @pause action item (arch-v1.md L734)
    pauseMessage := "@pause"
    actionItem, err := svc.ParseActionItem(pauseMessage)
    if err != nil {
        t.Fatalf("Expected no error parsing @pause, got %v", err)
    }
    
    // Verify action item type (arch-v1.md L734)
    if actionItem.Type != "pause" {
        t.Errorf("Expected action type 'pause', got '%s'", actionItem.Type)
    }
    
    // Send @pause and verify it becomes special Mail (arch-v1.md L734)
    mail, err := svc.SendHumanMessage(agentID, pauseMessage)
    if err != nil {
        t.Fatalf("Expected no error sending @pause, got %v", err)
    }
    
    // Verify mail has action item metadata (arch-v1.md L734)
    if mail.Metadata.ActionItem.Type != "pause" {
        t.Errorf("Expected action item type 'pause' in mail metadata, got '%s'", mail.Metadata.ActionItem.Type)
    }
    
    // Test @inject-memory X action item (arch-v1.md L734)
    injectMessage := "@inject-memory This is important to remember"
    actionItem, err = svc.ParseActionItem(injectMessage)
    if err != nil {
        t.Fatalf("Expected no error parsing @inject-memory, got %v", err)
    }
    
    // Verify action item type (arch-v1.md L734)
    if actionItem.Type != "inject-memory" {
        t.Errorf("Expected action type 'inject-memory', got '%s'", actionItem.Type)
    }
    
    // Verify payload contains the memory content (arch-v1.md L734)
    if actionItem.Payload != "This is important to remember" {
        t.Errorf("Expected payload 'This is important to remember', got '%v'", actionItem.Payload)
    }
    
    // Send @inject-memory and verify it becomes special Mail (arch-v1.md L734)
    mail, err = svc.SendHumanMessage(agentID, injectMessage)
    if err != nil {
        t.Fatalf("Expected no error sending @inject-memory, got %v", err)
    }
    
    // Verify mail has action item metadata (arch-v1.md L734)
    if mail.Metadata.ActionItem.Type != "inject-memory" {
        t.Errorf("Expected action item type 'inject-memory', got '%s'", mail.Metadata.ActionItem.Type)
    }
    
    if mail.Metadata.ActionItem.Payload != "This is important to remember" {
        t.Errorf("Expected action item payload preserved in mail")
    }
    
    // Test regular message (no action item)
    regularMessage := "Just a regular message"
    actionItem, err = svc.ParseActionItem(regularMessage)
    if err != nil {
        t.Fatalf("Expected no error parsing regular message, got %v", err)
    }
    
    // Verify no action item for regular message
    if actionItem != nil {
        t.Error("Expected nil action item for regular message")
    }
    
    // Test unknown action item (should be treated as regular message)
    unknownAction := "@unknown-action test"
    actionItem, err = svc.ParseActionItem(unknownAction)
    if err != nil {
        t.Fatalf("Expected no error parsing unknown action, got %v", err)
    }
    
    if actionItem != nil {
        t.Error("Expected nil action item for unknown action")
    }
}
```
**Acceptance Criteria:**
- @parse, @inject-memory X parsed correctly (arch-v1.md L734)
- Action items become special Mail messages (arch-v1.md L734)
- Action item type and payload preserved in mail metadata
- Regular messages without action items handled correctly

## Dependencies

### Test Dependencies
```
Test 1 (HTTPS endpoint) → Test 2 (ContextMap snapshot)
Test 3 (Last N messages sanitized)
Test 4 (Message becomes mail_received)
Test 5 (Agent replies via mail)
Test 6 (Action item shorthand)
```

### Phase Dependencies
- **Phase 8.2** - Streaming Path Integration (provides mail flow and streaming)
- **Phase 8.3** - HTTP(S) Exposure & OpenAPI (provides HTTPS endpoint infrastructure)

## Files to Create/Modify

| File | Operation | Changes |
|------|-----------|---------|
| `pkg/services/gateway/chat.go` | CREATE | Chat session management, message handling |
| `pkg/services/gateway/chat_test.go` | CREATE | Human chat tests |
| `pkg/mail/types.go` | MODIFY | Add ActionItem metadata, HumanFeedbackType |
| `pkg/security/contextmap.go` | MODIFY | Add Snapshot() method for read-only snapshot |

## Implementation Order (TDD)

1. **Test 1**: Write test → verify RED → Implement HTTPS chat endpoint → verify GREEN → commit
2. **Test 2**: Write test → verify RED → Implement read-only ContextMap snapshot → verify GREEN → commit
3. **Test 3**: Write test → verify RED → Implement last N messages sanitization → verify GREEN → commit
4. **Test 4**: Write test → verify RED → Implement human message to mail_received → verify GREEN → commit
5. **Test 5**: Write test → verify RED → Implement agent reply rendering → verify GREEN → commit
6. **Test 6**: Write test → verify RED → Implement action item shorthand parsing → verify GREEN → commit

**Total: 6 tests, 6 commits**

## Deliverables
- ✅ HTTPS chat session at `/chat/{agentId}` (arch-v1.md L728)
- ✅ Read-only ContextMap snapshot for sessions (arch-v1.md L731)
- ✅ Last N messages sanitized by boundary rules (arch-v1.md L731)
- ✅ Human messages become mail_received (type: human_feedback) (arch-v1.md L732)
- ✅ Agent replies via normal mail flow rendered in chat (arch-v1.md L733)
- ✅ Action item shorthands: @pause, @inject-memory X (arch-v1.md L734)
- ✅ 6 commits (one per test per TDD workflow)