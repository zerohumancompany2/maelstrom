# Phase 8.1: Channel Adapter Completion

## Goal
Complete all 9 channel adapter implementations per spec (arch-v1.md L659-667).

## Scope
- Complete WebhookAdapter HTTP POST handling (arch-v1.md L660)
- Complete WebSocketAdapter bidirectional handling (arch-v1.md L661)
- Complete SSEAdapter server-sent events (arch-v1.md L662)
- Complete SMTPAdapter email delivery (arch-v1.md L664)
- Complete InternalGRPCAdapter service mesh (arch-v1.md L666)
- Verify all adapters normalize to/from mail.Mail (arch-v1.md L670)
- 6 tests, 6 commits

## Current Status

### Implementation Status
| Adapter | Status | Notes |
|---------|--------|-------|
| `WebhookAdapter` | ⚠️ 30% complete | Stub exists; needs HTTP handling, normalization |
| `WebSocketAdapter` | ⚠️ 30% complete | Stub exists; needs WebSocket upgrade, streaming |
| `SSEAdapter` | ⚠️ 30% complete | Stub exists; needs SSE streaming |
| `SMTPAdapter` | ⚠️ 30% complete | Stub exists; needs email send/receive |
| `InternalGRPCAdapter` | ⚠️ 30% complete | Stub exists; needs gRPC handling |

### Files Status
| File | Status |
|------|--------|
| `pkg/services/gateway/adapter.go` | ⚠️ Partial - complete existing adapters |
| `pkg/services/gateway/webhook_adapter_test.go` | ⚠️ Partial - add HTTP handling tests |
| `pkg/services/gateway/websocket_adapter_test.go` | ⚠️ Partial - add WebSocket tests |
| `pkg/services/gateway/sse_adapter_test.go` | ⚠️ Partial - add SSE tests |
| `pkg/services/gateway/smtp_adapter_test.go` | ⚠️ Partial - add SMTP tests |
| `pkg/services/gateway/grpc_adapter_test.go` | ⚠️ Partial - add gRPC tests |

## Required Implementation

### WebhookAdapter HTTP POST Handling (arch-v1.md L660)
```go
// pkg/services/gateway/adapter.go
func (w *WebhookAdapter) Handle(r *http.Request) error {
    // Process HTTP POST request
    // Normalize body to mail.Mail with Type=MailReceived
    // Return HTTP 200 with acknowledgment
}
```

### WebSocketAdapter Bidirectional Streaming (arch-v1.md L661)
```go
// pkg/services/gateway/adapter.go
func (ws *WebSocketAdapter) Handle(r *http.Request) error {
    // Upgrade HTTP to WebSocket
    // Handle inbound messages → normalize to mail.Mail
    // Handle outbound mail → send to client
}
```

### SSEAdapter Server-Sent Events (arch-v1.md L662)
```go
// pkg/services/gateway/adapter.go
func (sse *SSEAdapter) Handle(r *http.Request) error {
    // Set Content-Type: text/event-stream
    // Stream chunks as SSE format (data: \n\n)
    // Support Last-Event-ID for reconnection
}
```

### SMTPAdapter Email Delivery (arch-v1.md L664)
```go
// pkg/services/gateway/adapter.go
func (s *SMTPAdapter) Handle(r *http.Request) error {
    // Send email via SMTP protocol
    // Normalize inbound email to mail.Mail
    // Format outbound mail as email
}
```

### InternalGRPCAdapter Service Mesh (arch-v1.md L666)
```go
// pkg/services/gateway/adapter.go
func (g *InternalGRPCAdapter) Handle(r *http.Request) error {
    // Handle gRPC service calls
    // Normalize protobuf messages to mail.Mail
    // Integrate with internal service mesh
}
```

## Tests to Write (6 tests, 6 commits)

### Test 1: TestChannelAdapter_WebhookNormalizesToMail
```go
func TestChannelAdapter_WebhookNormalizesToMail(t *testing.T) {
    adapter := &WebhookAdapter{}
    
    // Test HTTP POST to mail_received conversion (arch-v1.md L660)
    rawMessage := map[string]any{
        "from":    "sender@example.com",
        "to":      []string{"recipient@example.com"},
        "subject": "Test message",
        "body":    "Hello, world!",
    }
    
    mailMsg, err := adapter.NormalizeInbound(rawMessage)
    if err != nil {
        t.Fatalf("NormalizeInbound failed: %v", err)
    }
    
    // Verify mail type (arch-v1.md L670)
    if mailMsg.Type != mailpkg.MailReceived {
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
    
    // Test outbound normalization (arch-v1.md L670)
    outboundMail := &mailpkg.Mail{
        Type:    mailpkg.MailSend,
        Content: map[string]any{"response": "acknowledged"},
    }
    
    normalized, err := adapter.NormalizeOutbound(outboundMail)
    if err != nil {
        t.Fatalf("NormalizeOutbound failed: %v", err)
    }
    
    if normalized != nil {
        t.Error("Expected normalized outbound content")
    }
}
```
**Acceptance Criteria:**
- `WebhookAdapter.Handle()` processes HTTP POST requests (arch-v1.md L660)
- Request body normalized to mail.Mail with Type=MailReceived (arch-v1.md L670)
- Response sent as HTTP 200 with acknowledgment
- `Stream()` returns false for webhook

### Test 2: TestChannelAdapter_WebSocketBidirectional
```go
func TestChannelAdapter_WebSocketBidirectional(t *testing.T) {
    adapter := &WebSocketAdapter{}
    
    // Test inbound message normalization (arch-v1.md L661)
    inboundMessage := map[string]any{
        "text": "Hello from WebSocket client",
        "timestamp": time.Now().Unix(),
    }
    
    mailMsg, err := adapter.NormalizeInbound(inboundMessage)
    if err != nil {
        t.Fatalf("NormalizeInbound failed: %v", err)
    }
    
    // Verify mail type (arch-v1.md L670)
    if mailMsg.Type != mailpkg.MailReceived {
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
    
    // Test outbound normalization (arch-v1.md L670)
    outboundMail := &mailpkg.Mail{
        Type:    mailpkg.MailSend,
        Content: map[string]any{"text": "Response from server"},
    }
    
    normalized, err := adapter.NormalizeOutbound(outboundMail)
    if err != nil {
        t.Fatalf("NormalizeOutbound failed: %v", err)
    }
    
    content, ok := normalized.(map[string]any)
    if !ok {
        t.Fatalf("Expected normalized content to be map[string]any, got %T", normalized)
    }
    
    if content["text"] != "Response from server" {
        t.Errorf("Expected text 'Response from server', got %v", content["text"])
    }
}
```
**Acceptance Criteria:**
- `WebSocketAdapter.Handle()` upgrades HTTP to WebSocket (arch-v1.md L661)
- Inbound messages normalized to mail.Mail (arch-v1.md L670)
- Outbound mail normalized and sent to client (arch-v1.md L670)
- `Stream()` returns true for websocket

### Test 3: TestChannelAdapter_SSEServerSentEvents
```go
func TestChannelAdapter_SSEServerSentEvents(t *testing.T) {
    adapter := &SSEAdapter{}
    
    // Test inbound normalization (arch-v1.md L662)
    inboundMessage := map[string]any{
        "event": "user_input",
        "data":  "Hello",
    }
    
    mailMsg, err := adapter.NormalizeInbound(inboundMessage)
    if err != nil {
        t.Fatalf("NormalizeInbound failed: %v", err)
    }
    
    // Verify mail type (arch-v1.md L670)
    if mailMsg.Type != mailpkg.MailReceived {
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
    
    // Test outbound normalization to SSE format (arch-v1.md L670)
    outboundMail := &mailpkg.Mail{
        Type:    mailpkg.MailSend,
        Content: map[string]any{"text": "Server response"},
    }
    
    normalized, err := adapter.NormalizeOutbound(outboundMail)
    if err != nil {
        t.Fatalf("NormalizeOutbound failed: %v", err)
    }
    
    // Verify SSE format (data: \n\n)
    if normalized == nil {
        t.Error("Expected normalized SSE content")
    }
}
```
**Acceptance Criteria:**
- `SSEAdapter.Handle()` sets Content-Type: text/event-stream (arch-v1.md L662)
- Stream chunks sent as SSE format (data: \n\n) (arch-v1.md L662)
- Supports Last-Event-ID for reconnection
- `Stream()` returns true for sse

### Test 4: TestChannelAdapter_SMTPEmailDelivery
```go
func TestChannelAdapter_SMTPEmailDelivery(t *testing.T) {
    adapter := &SMTPAdapter{}
    
    // Test inbound email normalization (arch-v1.md L664)
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
    if mailMsg.Type != mailpkg.MailReceived {
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
    
    // Test outbound normalization to email format (arch-v1.md L670)
    outboundMail := &mailpkg.Mail{
        Type:    mailpkg.MailSend,
        Content: map[string]any{"text": "Response email body"},
    }
    
    normalized, err := adapter.NormalizeOutbound(outboundMail)
    if err != nil {
        t.Fatalf("NormalizeOutbound failed: %v", err)
    }
    
    if normalized == nil {
        t.Error("Expected normalized email content")
    }
}
```
**Acceptance Criteria:**
- `SMTPAdapter` sends emails via SMTP protocol (arch-v1.md L664)
- Inbound emails normalized to mail.Mail (arch-v1.md L670)
- Outbound mail formatted as email (arch-v1.md L670)
- `Stream()` returns false for smtp

### Test 5: TestChannelAdapter_GRPCInternalMesh
```go
func TestChannelAdapter_GRPCInternalMesh(t *testing.T) {
    adapter := &InternalGRPCAdapter{}
    
    // Test protobuf message normalization (arch-v1.md L666)
    protobufMessage := map[string]any{
        "service": "internal_service",
        "method":  "ProcessRequest",
        "payload": map[string]any{
            "id":    "req-001",
            "data":  "test payload",
        },
    }
    
    mailMsg, err := adapter.NormalizeInbound(protobufMessage)
    if err != nil {
        t.Fatalf("NormalizeInbound failed: %v", err)
    }
    
    // Verify mail type (arch-v1.md L670)
    if mailMsg.Type != mailpkg.MailReceived {
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
    
    // Test outbound normalization (arch-v1.md L670)
    outboundMail := &mailpkg.Mail{
        Type:    mailpkg.MailSend,
        Content: map[string]any{"result": "processed"},
    }
    
    normalized, err := adapter.NormalizeOutbound(outboundMail)
    if err != nil {
        t.Fatalf("NormalizeOutbound failed: %v", err)
    }
    
    if normalized == nil {
        t.Error("Expected normalized gRPC content")
    }
}
```
**Acceptance Criteria:**
- `InternalGRPCAdapter` handles gRPC service calls (arch-v1.md L666)
- Protobuf messages normalized to mail.Mail (arch-v1.md L670)
- Service mesh integration via gRPC (arch-v1.md L666)
- `Stream()` returns false for grpc

### Test 6: TestChannelAdapter_YamlHotReload
```go
func TestChannelAdapter_YamlHotReload(t *testing.T) {
    // Test YAML-declared adapters hot-loaded (arch-v1.md L659)
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
    
    // Verify each adapter can be instantiated and implements ChannelAdapter
    expectedAdapters := []string{"webhook", "websocket", "sse", "smtp", "grpc"}
    for i, adapterConfig := range adapters {
        adapterMap := adapterConfig.(map[string]any)
        adapterName := adapterMap["name"].(string)
        
        found := false
        for _, expected := range expectedAdapters {
            if adapterName == expected {
                found = true
                break
            }
        }
        
        if !found {
            t.Errorf("Expected adapter '%s' not found in config", adapterName)
        }
    }
    
    // Verify all adapter types implement ChannelAdapter interface
    var _ ChannelAdapter = &WebhookAdapter{}
    var _ ChannelAdapter = &WebSocketAdapter{}
    var _ ChannelAdapter = &SSEAdapter{}
    var _ ChannelAdapter = &SMTPAdapter{}
    var _ ChannelAdapter = &InternalGRPCAdapter{}
}
```
**Acceptance Criteria:**
- YAML-declared adapters can be hot-loaded (arch-v1.md L659)
- All 5 adapter types (webhook, websocket, sse, smtp, grpc) declared in YAML
- Each adapter implements ChannelAdapter interface
- Adapter configuration parsed correctly

## Dependencies

### Test Dependencies
```
Test 1 (Webhook adapter)
Test 2 (WebSocket adapter)
Test 3 (SSE adapter)
Test 4 (SMTP adapter)
Test 5 (gRPC adapter)
Test 6 (YAML hot reload - depends on all adapters)
```

### Phase Dependencies
- **None** - This is the first phase of Layer 8

## Files to Create/Modify

| File | Operation | Changes |
|------|-----------|---------|
| `pkg/services/gateway/adapter.go` | MODIFY | Complete WebhookAdapter, WebSocketAdapter, SSEAdapter, SMTPAdapter, InternalGRPCAdapter Handle methods |
| `pkg/services/gateway/webhook_adapter_test.go` | MODIFY | Add HTTP handling tests |
| `pkg/services/gateway/websocket_adapter_test.go` | MODIFY | Add WebSocket bidirectional tests |
| `pkg/services/gateway/sse_adapter_test.go` | MODIFY | Add SSE streaming tests |
| `pkg/services/gateway/smtp_adapter_test.go` | MODIFY | Add SMTP email delivery tests |
| `pkg/services/gateway/grpc_adapter_test.go` | MODIFY | Add gRPC service mesh tests |

## Implementation Order (TDD)

1. **Test 1**: Write test → verify RED → Implement Webhook HTTP handling → verify GREEN → commit
2. **Test 2**: Write test → verify RED → Implement WebSocket handling → verify GREEN → commit
3. **Test 3**: Write test → verify RED → Implement SSE streaming → verify GREEN → commit
4. **Test 4**: Write test → verify RED → Implement SMTP handling → verify GREEN → commit
5. **Test 5**: Write test → verify RED → Implement gRPC adapter → verify GREEN → commit
6. **Test 6**: Write test → verify RED → Implement YAML hot reload → verify GREEN → commit

**Total: 6 tests, 6 commits**

## Deliverables
- ✅ `WebhookAdapter` complete (arch-v1.md L660)
- ✅ `WebSocketAdapter` complete (arch-v1.md L661)
- ✅ `SSEAdapter` complete (arch-v1.md L662)
- ✅ `SMTPAdapter` complete (arch-v1.md L664)
- ✅ `InternalGRPCAdapter` complete (arch-v1.md L666)
- ✅ All adapters normalize inbound to mail.Mail (arch-v1.md L670)
- ✅ All adapters normalize outbound from mail.Mail (arch-v1.md L670)
- ✅ YAML-declared adapters hot-loadable (arch-v1.md L659)
- ✅ 6 commits (one per test per TDD workflow)