# Phase 8.2: Streaming Path Integration

## Goal
Implement complete end-to-end streaming path from user input to LLM response streaming per spec (arch-v1.md L672-703).

## Scope
- Implement streaming path flow (arch-v1.md L674-682)
- Implement StreamChunk format (arch-v1.md L696-701)
- Implement contextMap streaming configuration (arch-v1.md L685-691)
- Integrate with CommunicationService for mail routing
- Integrate with SecurityService for taint stripping
- 6 tests, 6 commits

## Current Status

### Implementation Status
| Component | Status | Notes |
|-----------|--------|-------|
| `StreamSession` | ⚠️ 60% complete | Exists in pkg/mail; needs full integration |
| `StreamChunk` | ⚠️ 60% complete | Exists in pkg/mail; needs boundary enforcement |
| `Streaming path` | ❌ Missing | Need to implement end-to-end flow |
| `ContextMap streaming` | ❌ Missing | Need to implement lastN strategy |

### Files Status
| File | Status |
|------|--------|
| `pkg/mail/stream.go` | ⚠️ Partial - add streaming path integration |
| `pkg/mail/stream_test.go` | ⚠️ Partial - add streaming tests |
| `pkg/services/gateway/service.go` | ⚠️ Partial - add streaming integration |

## Required Implementation

### Streaming Path User Input to Mail (arch-v1.md L675)
```go
// pkg/services/gateway/service.go
func (s *gatewayService) HandleUserInput(input string, sessionID string) (*Mail, error) {
    // Create mail_received with taints=["USER_SUPPLIED"], boundary=outer
    // Publish via CommunicationService to DMZ Agent inbox
}
```

### Streaming Path Agent Processing (arch-v1.md L677-678)
```go
// pkg/services/gateway/service.go
func (s *gatewayService) AssembleContextMap(agentID string) (*ContextMap, error) {
    // Pull session + Memory blocks
    // Propagate taints through context assembly
}
```

### Streaming Path Tool Execution (arch-v1.md L679-680)
```go
// pkg/services/gateway/service.go
func (s *gatewayService) ExecuteTool(toolCall ToolCall, namespace string) (*Mail, error) {
    // Route to Orchestrator (isolated sub-chart)
    // Tool Registry.resolve → executeIsolated (new namespace)
    // Return tool_result Mail with taints=["TOOL_OUTPUT"]
}
```

### Streaming Path Partial Assistant Messages (arch-v1.md L681)
```go
// pkg/services/gateway/service.go
func (s *gatewayService) EmitPartialAssistant(content string, sequence int) (*StreamChunk, error) {
    // Emit partial_assistant Messages (stream: true)
    // Security.stripForbiddenTaints before emission
}
```

### StreamChunk Format (arch-v1.md L696-701)
```go
// pkg/mail/stream.go
type StreamChunk struct {
    Chunk   string   `json:"chunk"`
    Sequence int     `json:"sequence"`
    IsFinal bool     `json:"isFinal"`
    Taints  []string `json:"taints"` // Security strips before emission
}
```

### ContextMap Streaming Configuration (arch-v1.md L685-691)
```go
// pkg/mail/stream.go
type ContextMapConfig struct {
    Name          string `yaml:"name"`
    Source        string `yaml:"source"`        // session
    Strategy      string `yaml:"strategy"`      // lastN
    N             int    `yaml:"n"`             // 30
    StreamCommit  bool   `yaml:"streamCommit"`  // false = partials never enter next prompt
}
```

## Tests to Write (6 tests, 6 commits)

### Test 1: TestStreamingPath_UserInputToMail
```go
func TestStreamingPath_UserInputToMail(t *testing.T) {
    svc := NewGatewayService()
    sessionID := "test-session-001"
    userInput := "Hello, agent!"
    
    // User input creates mail_received Mail (arch-v1.md L675)
    mail, err := svc.HandleUserInput(userInput, sessionID)
    if err != nil {
        t.Fatalf("Expected no error handling user input, got %v", err)
    }
    
    // Verify mail type (arch-v1.md L675)
    if mail.Type != "mail_received" {
        t.Errorf("Expected mail type 'mail_received', got '%s'", mail.Type)
    }
    
    // Verify taints=["USER_SUPPLIED"] (arch-v1.md L675)
    if len(mail.Taints) != 1 {
        t.Errorf("Expected 1 taint, got %d", len(mail.Taints))
    }
    if mail.Taints[0] != "USER_SUPPLIED" {
        t.Errorf("Expected taint 'USER_SUPPLIED', got '%s'", mail.Taints[0])
    }
    
    // Verify boundary=outer (arch-v1.md L675)
    if mail.Boundary != "outer" {
        t.Errorf("Expected boundary 'outer', got '%s'", mail.Boundary)
    }
    
    // Verify delivered to DMZ Agent inbox via Communication.publish
    if mail.Target != "agent:dmz" {
        t.Errorf("Expected target 'agent:dmz', got '%s'", mail.Target)
    }
    
    // Verify content matches user input
    if mail.Content != userInput {
        t.Errorf("Expected content '%s', got '%s'", userInput, mail.Content)
    }
}
```
**Acceptance Criteria:**
- `HandleUserInput()` creates mail_received Mail (arch-v1.md L675)
- Mail has taints=["USER_SUPPLIED"], boundary=outer
- Communication.publish() delivers to DMZ Agent inbox

### Test 2: TestStreamingPath_MailToLLMStream
```go
func TestStreamingPath_MailToLLMStream(t *testing.T) {
    svc := NewGatewayService()
    agentID := "test-agent-001"
    
    // Setup: Create mail_received in agent inbox
    mail := Mail{
        ID:       "mail-001",
        Type:     "mail_received",
        Source:   "user",
        Target:   "agent:" + agentID,
        Content:  "Process this request",
        Taints:   []string{"USER_SUPPLIED"},
        Boundary: "outer",
    }
    
    // Engine.dispatchEvent enters LLMReason state (arch-v1.md L677)
    state, err := svc.DispatchEvent(agentID, mail)
    if err != nil {
        t.Fatalf("Expected no error dispatching event, got %v", err)
    }
    
    // Verify LLMReason state entered (arch-v1.md L677)
    if state.Name != "LLMReason" {
        t.Errorf("Expected state 'LLMReason', got '%s'", state.Name)
    }
    
    // AgentExtensions.assembleContextMap pulls session + Memory blocks (arch-v1.md L678)
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
    
    // Verify streaming enabled for LLM call
    if !contextMap.StreamingEnabled {
        t.Error("Expected streaming enabled for LLM call")
    }
}
```
**Acceptance Criteria:**
- `DispatchEvent()` enters LLMReason state (arch-v1.md L677)
- `AssembleContextMap()` pulls session + Memory blocks (arch-v1.md L678)
- Taints propagated through context assembly

### Test 3: TestStreamingPath_LLMAssistantToStreamChunk
```go
func TestStreamingPath_LLMAssistantToStreamChunk(t *testing.T) {
    svc := NewGatewayService()
    agentID := "test-agent-001"
    
    // LLM returns assistant response
    llmResponse := "This is the assistant response to your query."
    
    // Engine emits partial_assistant Messages (stream: true) (arch-v1.md L681)
    chunk, err := svc.EmitPartialAssistant(llmResponse, 1)
    if err != nil {
        t.Fatalf("Expected no error emitting partial assistant, got %v", err)
    }
    
    // Verify StreamChunk format (arch-v1.md L696-701)
    if chunk.Chunk != llmResponse {
        t.Errorf("Expected chunk content '%s', got '%s'", llmResponse, chunk.Chunk)
    }
    
    // Verify sequence number
    if chunk.Sequence != 1 {
        t.Errorf("Expected sequence 1, got %d", chunk.Sequence)
    }
    
    // Verify isFinal flag for single chunk
    if !chunk.IsFinal {
        t.Error("Expected isFinal to be true for single chunk")
    }
    
    // Verify taints array exists (arch-v1.md L696-701)
    if chunk.Taints == nil {
        t.Error("Expected taints array to exist")
    }
    
    // Verify partial_assistant message type
    if chunk.MessageType != "partial_assistant" {
        t.Errorf("Expected message type 'partial_assistant', got '%s'", chunk.MessageType)
    }
}
```
**Acceptance Criteria:**
- `EmitPartialAssistant()` emits partial_assistant Messages with stream: true (arch-v1.md L681)
- StreamChunk has chunk, sequence, isFinal, taints fields (arch-v1.md L696-701)
- Communication routes to Gateway

### Test 4: TestStreamingPath_StreamChunkTaintStripping
```go
func TestStreamingPath_StreamChunkTaintStripping(t *testing.T) {
    svc := NewGatewayService()
    
    // Create StreamChunk with forbidden taints
    chunk := &StreamChunk{
        Chunk:    "This contains SECRET and PII data",
        Sequence: 1,
        IsFinal:  false,
        Taints:   []string{"USER_SUPPLIED", "SECRET", "PII", "TOOL_OUTPUT"},
    }
    
    // Security.stripForbiddenTaints applied before emission (arch-v1.md L681)
    strippedChunk, err := svc.StripForbiddenTaints(chunk)
    if err != nil {
        t.Fatalf("Expected no error stripping taints, got %v", err)
    }
    
    // Verify forbidden taints removed (arch-v1.md L681)
    if slices.Contains(strippedChunk.Taints, "SECRET") {
        t.Error("Expected SECRET taint to be stripped")
    }
    if slices.Contains(strippedChunk.Taints, "PII") {
        t.Error("Expected PII taint to be stripped")
    }
    
    // Verify allowed taints preserved
    if !slices.Contains(strippedChunk.Taints, "USER_SUPPLIED") {
        t.Error("Expected USER_SUPPLIED taint to be preserved")
    }
    if !slices.Contains(strippedChunk.Taints, "TOOL_OUTPUT") {
        t.Error("Expected TOOL_OUTPUT taint to be preserved")
    }
    
    // Verify chunk content unchanged (only taints stripped)
    if strippedChunk.Chunk != chunk.Chunk {
        t.Errorf("Expected chunk content unchanged, got '%s'", strippedChunk.Chunk)
    }
    
    // Verify sequence and isFinal preserved
    if strippedChunk.Sequence != chunk.Sequence {
        t.Errorf("Expected sequence %d, got %d", chunk.Sequence, strippedChunk.Sequence)
    }
    if strippedChunk.IsFinal != chunk.IsFinal {
        t.Errorf("Expected isFinal %v, got %v", chunk.IsFinal, strippedChunk.IsFinal)
    }
}
```
**Acceptance Criteria:**
- `StripForbiddenTaints()` removes SECRET/PII taints (arch-v1.md L681)
- Allowed taints (USER_SUPPLIED, TOOL_OUTPUT) preserved
- Chunk content and metadata unchanged

### Test 5: TestStreamingPath_SSEChunkFormat
```go
func TestStreamingPath_SSEChunkFormat(t *testing.T) {
    svc := NewGatewayService()
    
    // Create StreamChunk for SSE output
    chunk := &StreamChunk{
        Chunk:    "Assistant response part 1",
        Sequence: 1,
        IsFinal:  false,
        Taints:   []string{"USER_SUPPLIED"},
    }
    
    // Format as SSE chunk (arch-v1.md L696-701)
    sseData, err := svc.FormatSSEChunk(chunk)
    if err != nil {
        t.Fatalf("Expected no error formatting SSE chunk, got %v", err)
    }
    
    // Verify SSE format: data: <json>\n\n
    expectedPrefix := "data: "
    if !strings.HasPrefix(sseData, expectedPrefix) {
        t.Errorf("Expected SSE format starting with '%s', got '%s'", expectedPrefix, sseData[:min(20, len(sseData))])
    }
    
    // Verify JSON payload can be parsed
    jsonStr := strings.TrimPrefix(sseData, "data: ")
    var parsed map[string]any
    err = json.Unmarshal([]byte(jsonStr), &parsed)
    if err != nil {
        t.Fatalf("Expected valid JSON in SSE data, got %v", err)
    }
    
    // Verify chunk field in JSON
    if parsed["chunk"] != "Assistant response part 1" {
        t.Error("Expected chunk field in SSE JSON")
    }
    
    // Verify sequence field in JSON
    if parsed["sequence"].(float64) != 1 {
        t.Error("Expected sequence field in SSE JSON")
    }
    
    // Verify isFinal field in JSON
    if parsed["isFinal"] != false {
        t.Error("Expected isFinal field in SSE JSON")
    }
    
    // Verify taints stripped from SSE output (arch-v1.md L696-701)
    if _, exists := parsed["taints"]; exists {
        t.Error("Expected taints to be stripped from SSE output")
    }
}
```
**Acceptance Criteria:**
- SSE chunk format: `data: <json>\n\n` (arch-v1.md L696-701)
- JSON payload contains chunk, sequence, isFinal fields
- Taints stripped from SSE output before emission

### Test 6: TestStreamingPath_WebSocketChunkFormat
```go
func TestStreamingPath_WebSocketChunkFormat(t *testing.T) {
    svc := NewGatewayService()
    
    // Create StreamChunk for WebSocket output
    chunk := &StreamChunk{
        Chunk:    "Assistant response part 1",
        Sequence: 1,
        IsFinal:  false,
        Taints:   []string{"USER_SUPPLIED"},
    }
    
    // Format as WebSocket chunk (arch-v1.md L696-701)
    wsData, err := svc.FormatWebSocketChunk(chunk)
    if err != nil {
        t.Fatalf("Expected no error formatting WebSocket chunk, got %v", err)
    }
    
    // Verify WebSocket format: raw JSON
    var parsed map[string]any
    err = json.Unmarshal(wsData, &parsed)
    if err != nil {
        t.Fatalf("Expected valid JSON in WebSocket data, got %v", err)
    }
    
    // Verify chunk field in JSON
    if parsed["chunk"] != "Assistant response part 1" {
        t.Error("Expected chunk field in WebSocket JSON")
    }
    
    // Verify sequence field in JSON
    if parsed["sequence"].(float64) != 1 {
        t.Error("Expected sequence field in WebSocket JSON")
    }
    
    // Verify isFinal field in JSON
    if parsed["isFinal"] != false {
        t.Error("Expected isFinal field in WebSocket JSON")
    }
    
    // Verify taints stripped from WebSocket output (arch-v1.md L696-701)
    if _, exists := parsed["taints"]; exists {
        t.Error("Expected taints to be stripped from WebSocket output")
    }
    
    // Test final chunk format
    finalChunk := &StreamChunk{
        Chunk:    "Assistant response part 2 (final)",
        Sequence: 2,
        IsFinal:  true,
        Taints:   []string{},
    }
    
    finalData, err := svc.FormatWebSocketChunk(finalChunk)
    if err != nil {
        t.Fatalf("Expected no error formatting final WebSocket chunk, got %v", err)
    }
    
    var finalParsed map[string]any
    json.Unmarshal(finalData, &finalParsed)
    if finalParsed["isFinal"] != true {
        t.Error("Expected isFinal to be true for final chunk")
    }
}
```
**Acceptance Criteria:**
- WebSocket chunk format: raw JSON (arch-v1.md L696-701)
- JSON payload contains chunk, sequence, isFinal fields
- Taints stripped from WebSocket output before emission

## Dependencies

### Test Dependencies
```
Test 1 → Test 2 (User input to agent processing)
Test 3 (LLM assistant to stream chunk)
Test 4 (Stream chunk taint stripping)
Test 5 (SSE chunk format)
Test 6 (WebSocket chunk format)
```

### Phase Dependencies
- **Phase 8.1** - Channel Adapter Completion (provides adapter infrastructure)

## Files to Create/Modify

| File | Operation | Changes |
|------|-----------|---------|
| `pkg/mail/stream.go` | MODIFY | Add streaming path integration, StreamChunk format |
| `pkg/mail/stream_test.go` | MODIFY | Add streaming path tests |
| `pkg/services/gateway/service.go` | MODIFY | Add streaming integration, contextMap assembly |

## Implementation Order (TDD)

1. **Test 1**: Write test → verify RED → Implement user input to mail → verify GREEN → commit
2. **Test 2**: Write test → verify RED → Implement agent processing → verify GREEN → commit
3. **Test 3**: Write test → verify RED → Implement LLM assistant to stream chunk → verify GREEN → commit
4. **Test 4**: Write test → verify RED → Implement stream chunk taint stripping → verify GREEN → commit
5. **Test 5**: Write test → verify RED → Implement SSE chunk format → verify GREEN → commit
6. **Test 6**: Write test → verify RED → Implement WebSocket chunk format → verify GREEN → commit

**Total: 6 tests, 6 commits**

## Deliverables
- ✅ End-to-end streaming path complete (arch-v1.md L672-703)
- ✅ StreamChunk format with chunk, sequence, isFinal, taints fields
- ✅ SSE chunk format: `data: <json>\n\n` with taints stripped
- ✅ WebSocket chunk format: raw JSON with taints stripped
- ✅ ContextMap streaming with streamCommit: false configuration
- ✅ 6 commits (one per test per TDD workflow)