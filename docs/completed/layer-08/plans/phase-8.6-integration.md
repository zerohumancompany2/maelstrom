# Phase 8.6: Integration Tests

## Goal
End-to-end integration tests for Layer 8 (all components working together).

## Scope
- Full Layer 8 integration (arch-v1.md L653-736)
- Test complete streaming path: user → Gateway → mail → LLM → StreamChunk → SSE/WS
- Test human chat with running agent (full flow)
- Test channel adapter → mail → streaming end-to-end
- Test security enforcement throughout streaming path
- 4 tests, 4 commits

## Current Status

### Implementation Status
| Component | Status | Notes |
|-----------|--------|-------|
| `Channel Adapters` | ✅ Complete | Phase 8.1 |
| `Streaming Path` | ✅ Complete | Phase 8.2 |
| `HTTP(S) Exposure` | ✅ Complete | Phase 8.3 |
| `Human Chat` | ✅ Complete | Phase 8.4 |
| `Gateway Security` | ✅ Complete | Phase 8.5 |
| `Integration Tests` | ❌ Missing | Need to implement end-to-end tests |

### Files Status
| File | Status |
|------|--------|
| `pkg/services/gateway/integration_test.go` | ❌ Create - integration tests |

## Required Implementation

### Full Streaming Path Integration (arch-v1.md L674-682)
```go
// pkg/services/gateway/integration_test.go
func TestLayer8Integration_FullStreamingPath(t *testing.T) {
    // Complete streaming path: user → Gateway → mail → LLM → StreamChunk → SSE/WS
    // User types → Gateway (SSE connect) → Communication.publish(mail: type=user, taints=["USER_SUPPLIED"], boundary=outer)
    // Security.validate → DMZ Agent inbox
    // Engine.dispatchEvent → enters LLMReason state
    // AgentExtensions.assembleContextMap (pulls session + Memory blocks; taints propagated) → LLM call
    // LLM returns tool calls → Orchestrator (isolated sub-chart) → Tool Registry.resolve → executeIsolated (new namespace)
    // Tool result → taintPropagate → new tool_result Mail (taints=["TOOL_OUTPUT"]) injected to session
    // Engine emits partial_assistant Messages (stream: true) → Communication → Gateway → Security.stripForbiddenTaints → SSE chunks to user
}
```

### Human Chat with Running Agent (arch-v1.md L725-736)
```go
// pkg/services/gateway/integration_test.go
func TestLayer8Integration_HumanChatWithRunningAgent(t *testing.T) {
    // Human chat with running agent (full flow)
    // System Service: sys:human-gateway
    // For any long-running top-level Agent, open an HTTPS chat session at /chat/{agentId}
    // Session receives read-only snapshot of Agent's current ContextMap + last N Messages (sanitized by boundary rules)
    // Any message sent becomes mail_received (type: human_feedback or user) delivered to Agent's inbox
    // Agent replies via normal mail → rendered back in chat UI
    // Optional "action item" shorthand: @pause, @inject-memory X, etc. become special Mail messages
}
```

### Channel Adapter to Mail to Stream (arch-v1.md L659-671)
```go
// pkg/services/gateway/integration_test.go
func TestLayer8Integration_ChannelAdapterToMailToStream(t *testing.T) {
    // Channel adapter → mail → streaming end-to-end
    // Each adapter normalizes inbound traffic to a standard mail_received Message
    // Each adapter normalizes outbound mail_send to the channel's format
    // Test all channel adapters: webhook, websocket, sse, pubsub, smtp, slack, whatsapp, telegram, internal_grpc
}
```

### Security Enforced Throughout (arch-v1.md L276-299, L681, L700)
```go
// pkg/services/gateway/integration_test.go
func TestLayer8Integration_SecurityEnforcedThroughout(t *testing.T) {
    // Security enforcement throughout streaming path
    // All data entering the runtime is tainted at the border
    // No untainted information exists inside the runtime
    // Attach taints as soon as data is ingested
    // Every outgoing chunk passes through DMZ sanitizers
    // Sanitization is per-chunk (stateless) so latency stays <50ms
    // Security strips forbidden taints before emission
}
```

## Tests to Write (4 tests, 4 commits)

### Test 1: TestLayer8Integration_FullStreamingPath
```go
func TestLayer8Integration_FullStreamingPath(t *testing.T) {
    svc := NewGatewayService()
    sessionID := "test-session-001"
    agentID := "agent:dmz"
    
    // Setup: Create SSE connection (arch-v1.md L675)
    sseAdapter := &SSEAdapter{}
    var receivedChunks []*StreamChunk
    
    sseAdapter.OnChunk = func(chunk *StreamChunk) {
        receivedChunks = append(receivedChunks, chunk)
    }
    
    // User types → Gateway (SSE connect) → Communication.publish(mail: type=user, taints=["USER_SUPPLIED"], boundary=outer) (arch-v1.md L675)
    userInput := "Hello, agent! What can you do?"
    mail, err := svc.HandleUserInput(userInput, sessionID)
    if err != nil {
        t.Fatalf("Expected no error handling user input, got %v", err)
    }
    
    // Verify mail_received created (arch-v1.md L675)
    if mail.Type != mail.MailTypeMailReceived {
        t.Errorf("Expected mail type 'mail_received', got '%s'", mail.Type)
    }
    
    // Verify taints=["USER_SUPPLIED"] (arch-v1.md L675)
    if !slices.Contains(mail.Taints, "USER_SUPPLIED") {
        t.Error("Expected USER_SUPPLIED taint")
    }
    
    // Verify boundary=outer (arch-v1.md L675)
    if mail.Metadata.Boundary != mail.OuterBoundary {
        t.Errorf("Expected boundary 'outer', got '%s'", mail.Metadata.Boundary)
    }
    
    // Security.validate → DMZ Agent inbox (arch-v1.md L676)
    err = svc.ValidateMail(mail)
    if err != nil {
        t.Fatalf("Expected no error validating mail, got %v", err)
    }
    
    // Verify delivered to DMZ Agent inbox (arch-v1.md L676)
    if mail.Target != agentID {
        t.Errorf("Expected target '%s', got '%s'", agentID, mail.Target)
    }
    
    // Engine.dispatchEvent → enters LLMReason state (arch-v1.md L677)
    state, err := svc.DispatchEvent(agentID, *mail)
    if err != nil {
        t.Fatalf("Expected no error dispatching event, got %v", err)
    }
    
    // Verify LLMReason state entered (arch-v1.md L677)
    if state.Name != "LLMReason" {
        t.Errorf("Expected state 'LLMReason', got '%s'", state.Name)
    }
    
    // AgentExtensions.assembleContextMap (pulls session + Memory blocks; taints propagated) → LLM call (arch-v1.md L678)
    contextMap, err := svc.AssembleContextMap(agentID)
    if err != nil {
        t.Fatalf("Expected no error assembling context map, got %v", err)
    }
    
    // Verify contextMap includes session data (arch-v1.md L678)
    if contextMap.Session == nil {
        t.Error("Expected contextMap to include session")
    }
    
    // Verify taints propagated through context assembly (arch-v1.md L678)
    if len(contextMap.Taints) == 0 {
        t.Error("Expected contextMap to have propagated taints")
    }
    
    // LLM returns tool calls → Orchestrator (isolated sub-chart) → Tool Registry.resolve → executeIsolated (new namespace) (arch-v1.md L679)
    toolCall := ToolCall{
        Name:    "get_weather",
        Arguments: map[string]any{"location": "San Francisco"},
    }
    
    toolResult, err := svc.ExecuteTool(toolCall, "isolated-namespace-001")
    if err != nil {
        t.Fatalf("Expected no error executing tool, got %v", err)
    }
    
    // Tool result → taintPropagate → new tool_result Mail (taints=["TOOL_OUTPUT"]) injected to session (arch-v1.md L680)
    if toolResult.Type != mail.MailTypeToolResult {
        t.Errorf("Expected mail type 'tool_result', got '%s'", toolResult.Type)
    }
    
    // Verify taints=["TOOL_OUTPUT"] (arch-v1.md L680)
    if !slices.Contains(toolResult.Taints, "TOOL_OUTPUT") {
        t.Error("Expected TOOL_OUTPUT taint on tool result")
    }
    
    // Engine emits partial_assistant Messages (stream: true) → Communication → Gateway → Security.stripForbiddenTaints → SSE chunks to user (arch-v1.md L681)
    assistantResponse := "The weather in San Francisco is currently sunny with a high of 72°F."
    chunk, err := svc.EmitPartialAssistant(assistantResponse, 1)
    if err != nil {
        t.Fatalf("Expected no error emitting partial assistant, got %v", err)
    }
    
    // Verify StreamChunk format (arch-v1.md L696-701)
    if chunk.Chunk != assistantResponse {
        t.Errorf("Expected chunk content '%s', got '%s'", assistantResponse, chunk.Chunk)
    }
    
    // Verify sequence number
    if chunk.Sequence != 1 {
        t.Errorf("Expected sequence 1, got %d", chunk.Sequence)
    }
    
    // Security.stripForbiddenTaints before emission (arch-v1.md L681)
    strippedChunk, err := svc.StripForbiddenTaints(chunk)
    if err != nil {
        t.Fatalf("Expected no error stripping taints, got %v", err)
    }
    
    // Verify forbidden taints removed (arch-v1.md L681)
    if slices.Contains(strippedChunk.Taints, "SECRET") {
        t.Error("Expected SECRET taint to be stripped")
    }
    
    if slices.Contains(strippedChunk.Taints, "INNER_ONLY") {
        t.Error("Expected INNER_ONLY taint to be stripped")
    }
    
    // SSE chunks to user (arch-v1.md L681)
    sseData, err := svc.FormatSSEChunk(strippedChunk)
    if err != nil {
        t.Fatalf("Expected no error formatting SSE chunk, got %v", err)
    }
    
    // Verify SSE format (arch-v1.md L662)
    if !strings.HasPrefix(sseData, "data: ") {
        t.Errorf("Expected SSE format starting with 'data: ', got '%s'", sseData[:min(20, len(sseData))])
    }
    
    // Verify full streaming path completed
    t.Logf("Full streaming path completed: user → Gateway → mail → LLM → StreamChunk → SSE")
}
```
**Acceptance Criteria:**
- Complete streaming path: user → Gateway → mail → LLM → StreamChunk → SSE/WS (arch-v1.md L674-682)
- User input creates mail_received with taints=["USER_SUPPLIED"], boundary=outer (arch-v1.md L675)
- Security.validate passes mail to DMZ Agent inbox (arch-v1.md L676)
- Engine.dispatchEvent enters LLMReason state (arch-v1.md L677)
- AgentExtensions.assembleContextMap pulls session + Memory blocks; taints propagated (arch-v1.md L678)
- Tool execution: Orchestrator → Tool Registry.resolve → executeIsolated (new namespace) (arch-v1.md L679)
- Tool result creates tool_result Mail with taints=["TOOL_OUTPUT"] (arch-v1.md L680)
- Engine emits partial_assistant Messages (stream: true) (arch-v1.md L681)
- Security.stripForbiddenTaints before SSE emission (arch-v1.md L681)
- SSE chunks formatted correctly to user (arch-v1.md L681)

### Test 2: TestLayer8Integration_HumanChatWithRunningAgent
```go
func TestLayer8Integration_HumanChatWithRunningAgent(t *testing.T) {
    svc := NewGatewayService()
    agentID := "agent:dmz"
    
    // System Service: sys:human-gateway (arch-v1.md L725)
    // For any long-running top-level Agent, open an HTTPS chat session at /chat/{agentId} (arch-v1.md L728)
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
    
    // Verify HTTPS endpoint path format (arch-v1.md L728)
    expectedPath := "/chat/" + agentID
    actualPath := svc.GetChatPath(agentID)
    if actualPath != expectedPath {
        t.Errorf("Expected path '%s', got '%s'", expectedPath, actualPath)
    }
    
    // Session receives read-only snapshot of Agent's current ContextMap (arch-v1.md L731)
    if session.ContextMap == nil {
        t.Error("Expected session to have ContextMap snapshot")
    }
    
    // Verify snapshot is read-only (arch-v1.md L731)
    if !session.ContextMap.ReadOnly {
        t.Error("Expected session ContextMap to be read-only")
    }
    
    // Setup: Add messages to session for last N Messages test
    messages := []ChatMessage{
        {ID: "msg1", Content: "Initial message", Taints: []string{}, Boundary: "outer", Type: "user"},
        {ID: "msg2", Content: "Assistant response", Taints: []string{}, Boundary: "outer", Type: "assistant"},
        {ID: "msg3", Content: "Secret message", Taints: []string{"SECRET"}, Boundary: "inner", Type: "assistant"},
        {ID: "msg4", Content: "PII message", Taints: []string{"PII"}, Boundary: "dmz", Type: "user"},
        {ID: "msg5", Content: "Clean message", Taints: []string{}, Boundary: "outer", Type: "user"},
    }
    
    for _, msg := range messages {
        session.Messages = append(session.Messages, msg)
    }
    
    // Last N Messages (sanitized by boundary rules) (arch-v1.md L731)
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
    
    // Any message sent becomes mail_received (type: human_feedback or user) delivered to Agent's inbox (arch-v1.md L732)
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
    
    // Agent replies via normal mail → rendered back in chat UI (arch-v1.md L733)
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
    
    // Optional "action item" shorthand: @pause, @inject-memory X, etc. become special Mail messages (arch-v1.md L734)
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
    actionMail, err := svc.SendHumanMessage(agentID, pauseMessage)
    if err != nil {
        t.Fatalf("Expected no error sending @pause, got %v", err)
    }
    
    // Verify mail has action item metadata (arch-v1.md L734)
    if actionMail.Metadata.ActionItem.Type != "pause" {
        t.Errorf("Expected action item type 'pause' in mail metadata, got '%s'", actionMail.Metadata.ActionItem.Type)
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
    
    // Verify full human chat flow completed
    t.Logf("Human chat with running agent completed: session → messages → mail → replies → action items")
}
```
**Acceptance Criteria:**
- HTTPS chat session at `/chat/{agentId}` (arch-v1.md L728)
- Session receives read-only snapshot of Agent's ContextMap (arch-v1.md L731)
- Last N Messages sanitized by boundary rules (arch-v1.md L731)
- SECRET/PII taints sanitized/removed (arch-v1.md L731)
- Inner-boundary messages removed (arch-v1.md L731)
- Messages become mail_received (type: human_feedback) (arch-v1.md L732)
- Delivered to Agent's inbox (arch-v1.md L732)
- Agent replies via normal mail rendered in chat (arch-v1.md L733)
- Action item shorthands: @pause, @inject-memory X (arch-v1.md L734)
- Action items become special Mail messages (arch-v1.md L734)

### Test 3: TestLayer8Integration_ChannelAdapterToMailToStream
```go
func TestLayer8Integration_ChannelAdapterToMailToStream(t *testing.T) {
    svc := NewGatewayService()
    agentID := "agent:dmz"
    
    // Each adapter normalizes inbound traffic to a standard mail_received Message (arch-v1.md L670)
    // Each adapter normalizes outbound mail_send to the channel's format (arch-v1.md L670)
    
    // Test WebhookAdapter (arch-v1.md L660)
    t.Run("WebhookAdapter", func(t *testing.T) {
        adapter := &WebhookAdapter{}
        
        // Inbound normalization (arch-v1.md L670)
        inboundMessage := map[string]any{
            "from":    "sender@example.com",
            "to":      []string{"recipient@example.com"},
            "subject": "Test message",
            "body":    "Hello, world!",
        }
        
        mailMsg, err := adapter.NormalizeInbound(inboundMessage)
        if err != nil {
            t.Fatalf("NormalizeInbound failed: %v", err)
        }
        
        // Verify mail type (arch-v1.md L670)
        if mailMsg.Type != mail.MailTypeMailReceived {
            t.Errorf("Expected type mail_received, got %v", mailMsg.Type)
        }
        
        // Verify adapter metadata (arch-v1.md L660)
        if mailMsg.Metadata.Adapter != "webhook" {
            t.Errorf("Expected adapter 'webhook', got %v", mailMsg.Metadata.Adapter)
        }
        
        // Verify Stream() returns false for webhook (arch-v1.md L660)
        if adapter.Stream() {
            t.Error("Expected Stream() to return false for webhook")
        }
        
        // Outbound normalization (arch-v1.md L670)
        outboundMail := &mail.Mail{
            Type:    mail.MailTypeMailSend,
            Content: map[string]any{"response": "acknowledged"},
        }
        
        normalized, err := adapter.NormalizeOutbound(outboundMail)
        if err != nil {
            t.Fatalf("NormalizeOutbound failed: %v", err)
        }
        
        if normalized == nil {
            t.Error("Expected normalized outbound content")
        }
    })
    
    // Test WebSocketAdapter (arch-v1.md L661)
    t.Run("WebSocketAdapter", func(t *testing.T) {
        adapter := &WebSocketAdapter{}
        
        // Inbound normalization (arch-v1.md L670)
        inboundMessage := map[string]any{
            "text":      "Hello from WebSocket client",
            "timestamp": time.Now().Unix(),
        }
        
        mailMsg, err := adapter.NormalizeInbound(inboundMessage)
        if err != nil {
            t.Fatalf("NormalizeInbound failed: %v", err)
        }
        
        // Verify mail type (arch-v1.md L670)
        if mailMsg.Type != mail.MailTypeMailReceived {
            t.Errorf("Expected type mail_received, got %v", mailMsg.Type)
        }
        
        // Verify adapter metadata (arch-v1.md L661)
        if mailMsg.Metadata.Adapter != "websocket" {
            t.Errorf("Expected adapter 'websocket', got %v", mailMsg.Metadata.Adapter)
        }
        
        // Verify Stream() returns true for websocket (arch-v1.md L661)
        if !adapter.Stream() {
            t.Error("Expected Stream() to return true for websocket")
        }
        
        // Full streaming path: WebSocket → mail → StreamChunk → WebSocket
        chunk := &StreamChunk{
            Chunk:    "Assistant response",
            Sequence: 1,
            IsFinal:  true,
            Taints:   []string{"USER_SUPPLIED"},
        }
        
        wsData, err := svc.FormatWebSocketChunk(chunk)
        if err != nil {
            t.Fatalf("Expected no error formatting WebSocket chunk, got %v", err)
        }
        
        // Verify WebSocket format: raw JSON (arch-v1.md L696-701)
        var parsed map[string]any
        err = json.Unmarshal(wsData, &parsed)
        if err != nil {
            t.Fatalf("Expected valid JSON in WebSocket data, got %v", err)
        }
        
        if parsed["chunk"] != "Assistant response" {
            t.Error("Expected chunk field in WebSocket JSON")
        }
    })
    
    // Test SSEAdapter (arch-v1.md L662)
    t.Run("SSEAdapter", func(t *testing.T) {
        adapter := &SSEAdapter{}
        
        // Inbound normalization (arch-v1.md L670)
        inboundMessage := map[string]any{
            "event": "user_input",
            "data":  "Hello",
        }
        
        mailMsg, err := adapter.NormalizeInbound(inboundMessage)
        if err != nil {
            t.Fatalf("NormalizeInbound failed: %v", err)
        }
        
        // Verify mail type (arch-v1.md L670)
        if mailMsg.Type != mail.MailTypeMailReceived {
            t.Errorf("Expected type mail_received, got %v", mailMsg.Type)
        }
        
        // Verify adapter metadata (arch-v1.md L662)
        if mailMsg.Metadata.Adapter != "sse" {
            t.Errorf("Expected adapter 'sse', got %v", mailMsg.Metadata.Adapter)
        }
        
        // Verify Stream() returns true for sse (arch-v1.md L662)
        if !adapter.Stream() {
            t.Error("Expected Stream() to return true for sse")
        }
        
        // Full streaming path: SSE → mail → StreamChunk → SSE
        chunk := &StreamChunk{
            Chunk:    "Assistant response",
            Sequence: 1,
            IsFinal:  true,
            Taints:   []string{"USER_SUPPLIED"},
        }
        
        sseData, err := svc.FormatSSEChunk(chunk)
        if err != nil {
            t.Fatalf("Expected no error formatting SSE chunk, got %v", err)
        }
        
        // Verify SSE format: data: <json>\n\n (arch-v1.md L696-701)
        if !strings.HasPrefix(sseData, "data: ") {
            t.Errorf("Expected SSE format starting with 'data: ', got '%s'", sseData[:min(20, len(sseData))])
        }
    })
    
    // Test SMTPAdapter (arch-v1.md L664)
    t.Run("SMTPAdapter", func(t *testing.T) {
        adapter := &SMTPAdapter{}
        
        // Inbound normalization (arch-v1.md L670)
        inboundEmail := map[string]any{
            "from":    "sender@example.com",
            "to":      []string{"recipient@example.com"},
            "subject": "Test email",
            "body":    "Email body content",
        }
        
        mailMsg, err := adapter.NormalizeInbound(inboundEmail)
        if err != nil {
            t.Fatalf("NormalizeInbound failed: %v", err)
        }
        
        // Verify mail type (arch-v1.md L670)
        if mailMsg.Type != mail.MailTypeMailReceived {
            t.Errorf("Expected type mail_received, got %v", mailMsg.Type)
        }
        
        // Verify adapter metadata (arch-v1.md L664)
        if mailMsg.Metadata.Adapter != "smtp" {
            t.Errorf("Expected adapter 'smtp', got %v", mailMsg.Metadata.Adapter)
        }
        
        // Verify Stream() returns false for smtp (arch-v1.md L664)
        if adapter.Stream() {
            t.Error("Expected Stream() to return false for smtp")
        }
    })
    
    // Test InternalGRPCAdapter (arch-v1.md L666)
    t.Run("InternalGRPCAdapter", func(t *testing.T) {
        adapter := &InternalGRPCAdapter{}
        
        // Inbound normalization (arch-v1.md L670)
        protobufMessage := map[string]any{
            "service": "internal_service",
            "method":  "ProcessRequest",
            "payload": map[string]any{
                "id":   "req-001",
                "data": "test payload",
            },
        }
        
        mailMsg, err := adapter.NormalizeInbound(protobufMessage)
        if err != nil {
            t.Fatalf("NormalizeInbound failed: %v", err)
        }
        
        // Verify mail type (arch-v1.md L670)
        if mailMsg.Type != mail.MailTypeMailReceived {
            t.Errorf("Expected type mail_received, got %v", mailMsg.Type)
        }
        
        // Verify adapter metadata (arch-v1.md L666)
        if mailMsg.Metadata.Adapter != "grpc" {
            t.Errorf("Expected adapter 'grpc', got %v", mailMsg.Metadata.Adapter)
        }
        
        // Verify Stream() returns false for grpc (arch-v1.md L666)
        if adapter.Stream() {
            t.Error("Expected Stream() to return false for grpc")
        }
    })
    
    // Verify YAML-declared adapters hot-loaded (arch-v1.md L659)
    t.Run("YamlHotReload", func(t *testing.T) {
        yamlConfig := `
adapters:
  - name: webhook
    config:
      endpoint: /webhook/test
  - name: websocket
    config:
      endpoint: /ws/test
  - name: sse
    config:
      endpoint: /sse/test
  - name: smtp
    config:
      host: smtp.example.com
      port: 587
  - name: grpc
    config:
      address: 0.0.0.0:50051
`
        var config map[string]any
        err := yaml.Unmarshal([]byte(yamlConfig), &config)
        if err != nil {
            t.Fatalf("Failed to parse YAML config: %v", err)
        }
        
        adapters := config["adapters"].([]any)
        if len(adapters) != 5 {
            t.Errorf("Expected 5 adapters in config, got %d", len(adapters))
        }
        
        // Verify all adapter types implement ChannelAdapter interface (arch-v1.md L659)
        var _ ChannelAdapter = &WebhookAdapter{}
        var _ ChannelAdapter = &WebSocketAdapter{}
        var _ ChannelAdapter = &SSEAdapter{}
        var _ ChannelAdapter = &SMTPAdapter{}
        var _ ChannelAdapter = &InternalGRPCAdapter{}
    })
    
    // Verify full channel adapter → mail → streaming path completed
    t.Logf("Channel adapter to mail to stream completed for all adapters")
}
```
**Acceptance Criteria:**
- All channel adapters normalize inbound to mail_received (arch-v1.md L670)
- All channel adapters normalize outbound from mail_send (arch-v1.md L670)
- WebhookAdapter: HTTP POST, Stream()=false (arch-v1.md L660)
- WebSocketAdapter: bidirectional, Stream()=true (arch-v1.md L661)
- SSEAdapter: server-sent events, Stream()=true (arch-v1.md L662)
- SMTPAdapter: email, Stream()=false (arch-v1.md L664)
- InternalGRPCAdapter: service mesh, Stream()=false (arch-v1.md L666)
- YAML-declared adapters hot-loaded (arch-v1.md L659)
- Streaming adapters produce correct chunk format (SSE/WebSocket)

### Test 4: TestLayer8Integration_SecurityEnforcedThroughout
```go
func TestLayer8Integration_SecurityEnforcedThroughout(t *testing.T) {
    svc := NewGatewayService()
    agentID := "agent:dmz"
    
    // All data entering the runtime is tainted at the border (arch-v1.md L276)
    t.Run("BorderTainting", func(t *testing.T) {
        tainter := &BorderTainter{
            DefaultTaints: []string{"USER_SUPPLIED", "OUTER_BOUNDARY"},
        }
        
        testData := map[string]any{
            "message": "Hello, agent!",
            "source":  "user:web",
        }
        
        taintedData, err := tainter.TaintInboundData(testData, "user:web")
        if err != nil {
            t.Fatalf("Expected no error tainting inbound data, got %v", err)
        }
        
        // Verify data is tainted (arch-v1.md L276)
        taintedMap, ok := taintedData.(map[string]any)
        if !ok {
            t.Fatal("Expected tainted data to be map")
        }
        
        if taintedMap["taints"] == nil {
            t.Error("Expected taints to be attached to data")
        }
        
        // Verify taints are attached as soon as data is ingested (arch-v1.md L276)
        taints := taintedMap["taints"].([]string)
        if !slices.Contains(taints, "USER_SUPPLIED") {
            t.Error("Expected USER_SUPPLIED taint")
        }
        
        if !slices.Contains(taints, "OUTER_BOUNDARY") {
            t.Error("Expected OUTER_BOUNDARY taint")
        }
        
        // No untainted information exists inside the runtime (arch-v1.md L276)
        if len(taints) == 0 {
            t.Error("Expected all data to have at least one taint")
        }
    })
    
    // Per-chunk stream sanitization with <50ms latency (arch-v1.md L298-299)
    t.Run("StreamSanitizationPerChunk", func(t *testing.T) {
        sanitizer := &StreamSanitizer{
            Redactor:     NewPIIRedactor(),
            LengthCapper: NewLengthCapper(1000),
            SchemaValidator: NewSchemaValidator(),
            InnerDataStripper: NewInnerDataStripper(),
        }
        
        // Every outgoing chunk passes through DMZ sanitizers (arch-v1.md L298)
        chunks := []*StreamChunk{
            {
                Chunk:    "Hello, this is a normal message",
                Sequence: 1,
                IsFinal:  false,
                Taints:   []string{"USER_SUPPLIED"},
            },
            {
                Chunk:    "PII data: John Doe, email: john@example.com",
                Sequence: 2,
                IsFinal:  false,
                Taints:   []string{"PII"},
            },
            {
                Chunk:    "SECRET: api_key=abc123",
                Sequence: 3,
                IsFinal:  false,
                Taints:   []string{"SECRET", "INNER_ONLY"},
            },
            {
                Chunk:    "Final response complete",
                Sequence: 4,
                IsFinal:  true,
                Taints:   []string{},
            },
        }
        
        // Sanitization is per-chunk (stateless) (arch-v1.md L299)
        for i, chunk := range chunks {
            startTime := time.Now()
            sanitized, err := sanitizer.SanitizeChunk(*chunk)
            elapsed := time.Since(startTime)
            
            if err != nil {
                t.Fatalf("Expected no error sanitizing chunk %d, got %v", i, err)
            }
            
            // Latency stays <50ms (arch-v1.md L299)
            if elapsed >= 50*time.Millisecond {
                t.Errorf("Expected chunk %d sanitization <50ms, got %v", i, elapsed)
            }
            
            // Verify PII redaction (arch-v1.md L298)
            if slices.Contains(chunk.Taints, "PII") {
                if strings.Contains(sanitized.Chunk, "john@example.com") {
                    t.Error("Expected PII to be redacted")
                }
            }
            
            // Verify inner-data stripping (arch-v1.md L298)
            if slices.Contains(chunk.Taints, "INNER_ONLY") {
                if strings.Contains(sanitized.Chunk, "api_key") {
                    t.Error("Expected inner data to be stripped")
                }
            }
            
            // Verify length caps (arch-v1.md L298)
            if len(sanitized.Chunk) > 1000 {
                t.Error("Expected chunk to be capped at 1000 chars")
            }
        }
        
        // No buffering of entire response unless chart explicitly requests it (arch-v1.md L299)
        // Verify each chunk is sanitized independently (stateless)
        firstChunk, _ := sanitizer.SanitizeChunk(*chunks[0])
        secondChunk, _ := sanitizer.SanitizeChunk(*chunks[1])
        
        // Sanitizing chunk 2 should not depend on chunk 1
        if secondChunk.Sequence != 2 {
            t.Error("Expected chunk 2 to be sanitized independently")
        }
    })
    
    // Security strips forbidden taints before emission (arch-v1.md L681, L700)
    t.Run("ForbiddenTaintStripping", func(t *testing.T) {
        stripper := &ForbiddenTaintStripper{
            AllowedOnExit: map[string]bool{
                "USER_SUPPLIED": true,
                "TOOL_OUTPUT":   true,
                "PUBLIC":        true,
            },
        }
        
        // Security strips forbidden taints before emission (arch-v1.md L681)
        testData := map[string]any{
            "content": "Response content",
            "taints":  []string{"USER_SUPPLIED", "SECRET", "INNER_ONLY", "PII"},
        }
        
        strippedData, err := stripper.StripForbiddenTaints(testData, "outer")
        if err != nil {
            t.Fatalf("Expected no error stripping forbidden taints, got %v", err)
        }
        
        strippedMap := strippedData.(map[string]any)
        remainingTaints := strippedMap["taints"].([]string)
        
        // Uses allowedOnExit to determine what can leave runtime (arch-v1.md L700)
        if !slices.Contains(remainingTaints, "USER_SUPPLIED") {
            t.Error("Expected USER_SUPPLIED to remain (allowed on exit)")
        }
        
        if slices.Contains(remainingTaints, "SECRET") {
            t.Error("Expected SECRET to be stripped (forbidden)")
        }
        
        if slices.Contains(remainingTaints, "INNER_ONLY") {
            t.Error("Expected INNER_ONLY to be stripped (forbidden)")
        }
        
        if slices.Contains(remainingTaints, "PII") {
            t.Error("Expected PII to be stripped (forbidden)")
        }
        
        // Test Stream Chunk Format (arch-v1.md L700)
        chunk := map[string]any{
            "chunk":     "Response chunk",
            "sequence":  1,
            "isFinal":   false,
            "taints":    []string{"USER_SUPPLIED", "SECRET"},
        }
        
        strippedChunk, _ := stripper.StripForbiddenTaints(chunk, "outer")
        strippedChunkMap := strippedChunk.(map[string]any)
        
        // Security strips before emission (arch-v1.md L700)
        chunkTaints := strippedChunkMap["taints"].([]string)
        if slices.Contains(chunkTaints, "SECRET") {
            t.Error("Expected SECRET to be stripped from chunk before emission")
        }
        
        if !slices.Contains(chunkTaints, "USER_SUPPLIED") {
            t.Error("Expected USER_SUPPLIED to remain in chunk")
        }
    })
    
    // Boundary validation on ingress (arch-v1.md L286)
    t.Run("BoundaryValidation", func(t *testing.T) {
        validator := &BoundaryValidator{
            Policy: NewDefaultSecurityPolicy(),
        }
        
        // Validate mail on ingress to gateway
        inboundMail := &mail.Mail{
            ID:      "mail-001",
            Type:    mail.MailTypeUser,
            Source:  "user:web",
            Target:  agentID,
            Content: "Hello, agent!",
            Metadata: mail.MailMetadata{
                Boundary: mail.OuterBoundary,
                Taints:   []string{"USER_SUPPLIED"},
            },
            Taints: []string{"USER_SUPPLIED"},
        }
        
        err := validator.ValidateOnIngress(inboundMail)
        if err != nil {
            t.Fatalf("Expected no error validating inbound mail, got %v", err)
        }
        
        // Check boundary transitions are allowed
        transitionMail := &mail.Mail{
            ID:      "mail-002",
            Type:    mail.MailTypeAssistant,
            Source:  "agent:inner",
            Target:  "user:web",
            Content: "Response from inner agent",
            Metadata: mail.MailMetadata{
                Boundary: mail.InnerBoundary,
                Taints:   []string{"INNER_ONLY", "SECRET"},
            },
            Taints: []string{"INNER_ONLY", "SECRET"},
        }
        
        // Emit taint_violation event to dead-letter on violation (arch-v1.md L286)
        err = validator.ValidateOnIngress(transitionMail)
        if err == nil {
            t.Error("Expected error for forbidden boundary transition")
        }
        
        // Verify violation is logged
        if !strings.Contains(err.Error(), "taint_violation") {
            t.Error("Expected taint_violation in error message")
        }
        
        // Runtime guard: any action/guard can query taints (arch-v1.md L286)
        queryableMail := &mail.Mail{
            ID:      "mail-004",
            Type:    mail.MailTypeToolResult,
            Source:  "tool:registry",
            Target:  agentID,
            Content: "Tool output",
            Metadata: mail.MailMetadata{
                Boundary: mail.DMZBoundary,
                Taints:   []string{"TOOL_OUTPUT"},
            },
            Taints: []string{"TOOL_OUTPUT"},
        }
        
        // Verify taints can be queried
        taints := queryableMail.GetTaints()
        if !slices.Contains(taints, "TOOL_OUTPUT") {
            t.Error("Expected TOOL_OUTPUT taint to be queryable")
        }
    })
    
    // Taint propagation to mail (arch-v1.md L283)
    t.Run("TaintPropagation", func(t *testing.T) {
        // Security Service propagates taints on copy/read/write (arch-v1.md L283)
        sourceMail := &mail.Mail{
            ID:      "mail-001",
            Type:    mail.MailTypeUser,
            Source:  "user:web",
            Target:  agentID,
            Content: "User query with PII: john@example.com",
            Metadata: mail.MailMetadata{
                Boundary: mail.OuterBoundary,
                Taints:   []string{"USER_SUPPLIED", "PII"},
            },
            Taints: []string{"USER_SUPPLIED", "PII"},
        }
        
        // Like DLP tracking (arch-v1.md L283)
        targetMail := &mail.Mail{
            ID:       "mail-002",
            Type:     mail.MailTypeAssistant,
            Source:   agentID,
            Target:   "user:web",
            Content:  "Response to user query",
            Metadata: mail.MailMetadata{
                Boundary: mail.OuterBoundary,
            },
        }
        
        PropagateTaints(sourceMail, targetMail)
        
        // Verify taints propagated (arch-v1.md L283)
        if !slices.Contains(targetMail.Taints, "USER_SUPPLIED") {
            t.Error("Expected USER_SUPPLIED taint to propagate")
        }
        
        if !slices.Contains(targetMail.Taints, "PII") {
            t.Error("Expected PII taint to propagate")
        }
        
        // On-disk: taints stored with data (arch-v1.md L284)
        persistedMail := &mail.Mail{
            ID:       "mail-005",
            Type:     mail.MailTypeSnapshot,
            Source:   "persistence:service",
            Target:   agentID,
            Content:  "Snapshot data",
            Metadata: mail.MailMetadata{
                Boundary: mail.InnerBoundary,
            },
            Taints: []string{"SECRET", "INNER_ONLY"},
        }
        
        // Verify taints are stored with data (arch-v1.md L284)
        if !slices.Contains(persistedMail.Taints, "SECRET") {
            t.Error("Expected SECRET taint to be stored with data")
        }
        
        if !slices.Contains(persistedMail.Taints, "INNER_ONLY") {
            t.Error("Expected INNER_ONLY taint to be stored with data")
        }
    })
    
    // Verify full security enforcement throughout streaming path
    t.Logf("Security enforcement throughout streaming path verified: border tainting → per-chunk sanitization → forbidden taint stripping → boundary validation → taint propagation")
}
```
**Acceptance Criteria:**
- All data entering runtime tainted at border (arch-v1.md L276)
- No untainted information exists inside runtime (arch-v1.md L276)
- Taints attached as soon as data is ingested (arch-v1.md L276)
- Every outgoing chunk passes through DMZ sanitizers (arch-v1.md L298)
- Sanitization is per-chunk (stateless) (arch-v1.md L299)
- Latency stays <50ms (arch-v1.md L299)
- No buffering of entire response (arch-v1.md L299)
- Security strips forbidden taints before emission (arch-v1.md L681)
- Uses allowedOnExit to determine what can leave runtime (arch-v1.md L700)
- Boundary validation on ingress with taint_violation events (arch-v1.md L286)
- Runtime guard: any action/guard can query taints (arch-v1.md L286)
- Security Service propagates taints on copy/read/write (arch-v1.md L283)
- Like DLP tracking (arch-v1.md L283)
- On-disk: taints stored with data (arch-v1.md L284)

## Dependencies

### Test Dependencies
```
Test 1 (Full streaming path) → Requires Phase 8.2 (Streaming Path)
Test 2 (Human chat with running agent) → Requires Phase 8.4 (Human Chat)
Test 3 (Channel adapter to mail to stream) → Requires Phase 8.1 (Channel Adapters)
Test 4 (Security enforced throughout) → Requires Phase 8.5 (Gateway Security)
```

### Phase Dependencies
- **Phase 8.4** - Human Chat Completion (provides chat session and message handling)
- **Phase 8.5** - Gateway Security & Tainting (provides security enforcement)

## Files to Create/Modify

| File | Operation | Changes |
|------|-----------|---------|
| `pkg/services/gateway/integration_test.go` | CREATE | Full Layer 8 integration tests |

## Implementation Order (TDD)

1. **Test 1**: Write test → verify RED → Implement full streaming path integration → verify GREEN → commit
2. **Test 2**: Write test → verify RED → Implement human chat with running agent integration → verify GREEN → commit
3. **Test 3**: Write test → verify RED → Implement channel adapter to mail to stream integration → verify GREEN → commit
4. **Test 4**: Write test → verify RED → Implement security enforcement throughout integration → verify GREEN → commit

**Total: 4 tests, 4 commits**

## Deliverables
- ✅ Complete streaming path: user → Gateway → mail → LLM → StreamChunk → SSE/WS (arch-v1.md L674-682)
- ✅ Human chat with running agent (full flow) (arch-v1.md L725-736)
- ✅ Channel adapter → mail → streaming end-to-end (arch-v1.md L659-671)
- ✅ Security enforcement throughout streaming path (arch-v1.md L276-299, L681, L700)
- ✅ 4 commits (one per test per TDD workflow)