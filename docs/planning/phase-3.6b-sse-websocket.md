# Phase 3.6B: SSE & WebSocket Adapters

## Goal
Implement SSE (Server-Sent Events) and WebSocket adapters for bidirectional streaming.

## Scope
- Create `pkg/gateway/adapters/sse.go` with SSE adapter
- Create `pkg/gateway/adapters/websocket.go` with WebSocket adapter
- Implement bidirectional normalization
- Handle streaming chunks
- 2 tests, 2 commits

## Current Status

### Implementation Status
| Component | Status | Notes |
|-----------|--------|-------|
| `SSEAdapter` | ❌ Missing | Server-Sent Events |
| `WebSocketAdapter` | ❌ Missing | WebSocket bidirectional |

### Files Status
| File | Status |
|------|-------|
| `pkg/gateway/adapters/sse.go` | ❌ MISSING - create |
| `pkg/gateway/adapters/websocket.go` | ❌ MISSING - create |

## Required Implementation

### SSE Adapter
```go
// pkg/gateway/adapters/sse.go
package adapters

import (
    "fmt"
    "maelstrom/pkg/mail"
)

type SSEAdapter struct {
    name string
}

func NewSSEAdapter() *SSEAdapter {
    return &SSEAdapter{name: "sse"}
}

func (a *SSEAdapter) Name() string {
    return a.name
}

func (a *SSEAdapter) NormalizeInbound(data []byte) (mail.Mail, error) {
    // SSE inbound is typically client events (rare)
    return mail.Mail{
        ID:     generateID(),
        Type:   mail.MailTypeMailReceived,
        Source: "gateway:sse",
        Content: string(data),
        Metadata: mail.MailMetadata{
            Boundary: mail.OuterBoundary,
            Taints:   []string{"USER_SUPPLIED"},
        },
    }, nil
}

func (a *SSEAdapter) NormalizeOutbound(mail mail.Mail) ([]byte, error) {
    // Format as SSE: event: type\ndata: json\n\n
    var sb strings.Builder
    
    sb.WriteString(fmt.Sprintf("event: %s\n", mail.Type))
    
    dataJSON, err := json.Marshal(mail.Content)
    if err != nil {
        return nil, err
    }
    sb.WriteString(fmt.Sprintf("data: %s\n", string(dataJSON)))
    
    if mail.Metadata.Stream {
        chunk := mail.StreamChunk
        sb.WriteString(fmt.Sprintf("id: %d\n", chunk.Sequence))
        if chunk.IsFinal {
            sb.WriteString("event: end\n")
        }
    }
    
    sb.WriteString("\n")
    return []byte(sb.String()), nil
}
```

### WebSocket Adapter
```go
// pkg/gateway/adapters/websocket.go
package adapters

import (
    "encoding/json"
    "maelstrom/pkg/mail"
)

type WebSocketAdapter struct {
    name string
}

func NewWebSocketAdapter() *WebSocketAdapter {
    return &WebSocketAdapter{name: "websocket"}
}

func (a *WebSocketAdapter) Name() string {
    return a.name
}

func (a *WebSocketAdapter) NormalizeInbound(data []byte) (mail.Mail, error) {
    var payload map[string]any
    if err := json.Unmarshal(data, &payload); err != nil {
        return mail.Mail{}, err
    }
    
    return mail.Mail{
        ID:     generateID(),
        Type:   mail.MailTypeMailReceived,
        Source: "gateway:websocket",
        Content: payload,
        Metadata: mail.MailMetadata{
            Boundary: mail.OuterBoundary,
            Taints:   []string{"USER_SUPPLIED"},
        },
    }, nil
}

func (a *WebSocketAdapter) NormalizeOutbound(mail mail.Mail) ([]byte, error) {
    return json.Marshal(map[string]any{
        "type":    mail.Type,
        "id":      mail.ID,
        "content": mail.Content,
        "source":  mail.Source,
        "stream":  mail.Metadata.Stream,
    })
}
```

## Tests to Write (2 tests, 2 commits)

### Test 1: TestSSEAdapter_NormalizeOutbound
```go
func TestSSEAdapter_NormalizeOutbound(t *testing.T) {
    adapter := NewSSEAdapter()
    
    // Test regular message
    mail := mail.Mail{
        ID:      "msg-001",
        Type:    mail.MailTypeAssistant,
        Content: "Hello, world!",
        Metadata: mail.MailMetadata{
            Stream: false,
        },
    }
    
    data, err := adapter.NormalizeOutbound(mail)
    if err != nil {
        t.Errorf("Expected nil error, got %v", err)
    }
    
    str := string(data)
    if !strings.Contains(str, "event: assistant") {
        t.Errorf("Expected 'event: assistant' in output, got '%s'", str)
    }
    if !strings.Contains(str, "data:") {
        t.Errorf("Expected 'data:' in output, got '%s'", str)
    }
    
    // Test streaming message
    streamMail := mail.Mail{
        ID:      "msg-002",
        Type:    mail.MailTypePartialAssistant,
        Content: "chunk data",
        Metadata: mail.MailMetadata{
            Stream:  true,
            StreamChunk: mail.StreamChunk{
                Sequence: 1,
                IsFinal:  false,
            },
        },
    }
    
    streamData, _ := adapter.NormalizeOutbound(streamMail)
    streamStr := string(streamData)
    if !strings.Contains(streamStr, "id: 1") {
        t.Errorf("Expected 'id: 1' in streaming output, got '%s'", streamStr)
    }
}
```
**Acceptance Criteria:**
- SSE adapter formats mail for SSE output
- Returns properly formatted SSE data (event:, data:, id:)
- Handles streaming chunks with sequence numbers

### Test 2: TestWebSocketAdapter_Bidirectional
```go
func TestWebSocketAdapter_Bidirectional(t *testing.T) {
    adapter := NewWebSocketAdapter()
    
    // Test inbound normalization
    inputData := []byte(`{"message": "hello", "type": "chat"}`)
    inboundMail, err := adapter.NormalizeInbound(inputData)
    if err != nil {
        t.Errorf("Expected nil error, got %v", err)
    }
    
    if inboundMail.Type != mail.MailTypeMailReceived {
        t.Errorf("Expected MailTypeMailReceived, got %s", inboundMail.Type)
    }
    if inboundMail.Source != "gateway:websocket" {
        t.Errorf("Expected source 'gateway:websocket', got '%s'", inboundMail.Source)
    }
    
    // Test outbound normalization
    outboundMail := mail.Mail{
        ID:      "msg-001",
        Type:    mail.MailTypeAssistant,
        Content: "Response content",
        Source:  "agent:bot",
        Metadata: mail.MailMetadata{
            Stream: false,
        },
    }
    
    outboundData, err := adapter.NormalizeOutbound(outboundMail)
    if err != nil {
        t.Errorf("Expected nil error, got %v", err)
    }
    
    // Verify JSON structure
    var result map[string]any
    if err := json.Unmarshal(outboundData, &result); err != nil {
        t.Errorf("Expected valid JSON, got error: %v", err)
    }
    
    if result["type"] != "assistant" {
        t.Errorf("Expected type 'assistant', got '%v'", result["type"])
    }
    if result["source"] != "agent:bot" {
        t.Errorf("Expected source 'agent:bot', got '%v'", result["source"])
    }
}
```
**Acceptance Criteria:**
- WebSocket adapter handles inbound normalization
- WebSocket adapter handles outbound normalization
- Both directions work correctly
- JSON format preserved

## Dependencies

### Test Dependencies
```
Test 1 → Independent
Test 2 → Independent
```

### Phase Dependencies
- **Phase 3.1** must complete first (Mail types)
- **Phase 3.5** must complete first (Streaming)
- **Phase 3.6a** must complete first (Gateway core)
- **Phase 3.6c** is sibling phase

## Files to Create/Modify

| File | Operation | Changes |
|------|-----------|---------|
| `pkg/gateway/adapters/sse.go` | CREATE | SSE adapter |
| `pkg/gateway/adapters/websocket.go` | CREATE | WebSocket adapter |

## Implementation Order (TDD)

1. **Test 1**: Write test → verify RED → Implement SSE adapter → verify GREEN → commit
2. **Test 2**: Write test → verify RED → Implement WebSocket adapter → verify GREEN → commit

**Total: 2 tests, 2 commits**

## Deliverables
- ✅ SSE adapter in `pkg/gateway/adapters/sse.go`
- ✅ WebSocket adapter in `pkg/gateway/adapters/websocket.go`
- ✅ Bidirectional normalization
- ✅ 2 commits (one per test per TDD workflow)

## Recursive Analysis Decision

**Decision: Split from original Phase 3.6**

**Reasoning:**
- SSE and WebSocket are both streaming adapters
- They share similar patterns (bidirectional, streaming)
- 2 tests is within recommended range (2-5)
- Can be developed independently of other adapters

**Benefits of split:**
- Streaming adapters grouped together
- Consistent pattern for both adapters
- Can be developed in parallel with 3.6a, 3.6c
- Clear separation from stub adapters