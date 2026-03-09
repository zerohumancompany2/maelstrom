# Phase 3.6A: Gateway Core & Webhook Adapter

## Goal
Implement Gateway service and Webhook adapter for HTTP POST endpoints.

## Scope
- Create `pkg/gateway/gateway.go` with Gateway service
- Create `pkg/gateway/adapters/webhook.go` with Webhook adapter
- Implement Adapter interface
- Normalize inbound to mail_received
- 2 tests, 2 commits

## Current Status

### Implementation Status
| Component | Status | Notes |
|-----------|--------|-------|
| `Gateway` | ❌ Missing | Gateway service |
| `Adapter` interface | ❌ Missing | Adapter interface |
| `WebhookAdapter` | ❌ Missing | HTTP POST adapter |

### Files Status
| File | Status |
|------|-------|
| `pkg/gateway/gateway.go` | ❌ MISSING - create |
| `pkg/gateway/adapters/webhook.go` | ❌ MISSING - create |

## Required Implementation

### Gateway Service
```go
// pkg/gateway/gateway.go
package gateway

import (
    "sync"
    "maelstrom/pkg/mail"
)

type Adapter interface {
    Name() string
    NormalizeInbound(data []byte) (mail.Mail, error)
    NormalizeOutbound(mail mail.Mail) ([]byte, error)
}

type Gateway struct {
    adapters map[string]Adapter
    mu       sync.RWMutex
}

func NewGateway() *Gateway {
    return &Gateway{
        adapters: make(map[string]Adapter),
    }
}

func (g *Gateway) RegisterAdapter(adapter Adapter) error {
    g.mu.Lock()
    defer g.mu.Unlock()
    g.adapters[adapter.Name()] = adapter
    return nil
}

func (g *Gateway) GetAdapter(name string) (Adapter, error) {
    g.mu.RLock()
    defer g.mu.RUnlock()
    adapter, exists := g.adapters[name]
    if !exists {
        return nil, fmt.Errorf("adapter not found: %s", name)
    }
    return adapter, nil
}

func (g *Gateway) ListAdapters() []string {
    g.mu.RLock()
    defer g.mu.RUnlock()
    names := make([]string, 0, len(g.adapters))
    for name := range g.adapters {
        names = append(names, name)
    }
    return names
}
```

### Webhook Adapter
```go
// pkg/gateway/adapters/webhook.go
package adapters

import (
    "encoding/json"
    "maelstrom/pkg/mail"
)

type WebhookAdapter struct {
    name string
}

func NewWebhookAdapter() *WebhookAdapter {
    return &WebhookAdapter{name: "webhook"}
}

func (a *WebhookAdapter) Name() string {
    return a.name
}

func (a *WebhookAdapter) NormalizeInbound(data []byte) (mail.Mail, error) {
    var payload map[string]any
    if err := json.Unmarshal(data, &payload); err != nil {
        return mail.Mail{}, err
    }
    
    return mail.Mail{
        ID:     generateID(),
        Type:   mail.MailTypeMailReceived,
        Source: "gateway:webhook",
        Target: "agent:default",
        Content: payload,
        Metadata: mail.MailMetadata{
            Boundary: mail.OuterBoundary,
            Taints:   []string{"USER_SUPPLIED"},
        },
    }, nil
}

func (a *WebhookAdapter) NormalizeOutbound(mail mail.Mail) ([]byte, error) {
    return json.Marshal(map[string]any{
        "type":    mail.Type,
        "content": mail.Content,
        "source":  mail.Source,
    })
}

func generateID() string {
    // In real implementation, use proper UUID generation
    return "webhook-" + randomString(8)
}

func randomString(n int) string {
    // Simplified for documentation
    return "abc123"
}
```

## Tests to Write (2 tests, 2 commits)

### Test 1: TestGateway_RegisterAdapter
```go
func TestGateway_RegisterAdapter(t *testing.T) {
    gateway := NewGateway()
    
    // Create mock adapter
    type mockAdapter struct {
        name string
    }
    func (m *mockAdapter) Name() string { return m.name }
    func (m *mockAdapter) NormalizeInbound(data []byte) (mail.Mail, error) {
        return mail.Mail{}, nil
    }
    func (m *mockAdapter) NormalizeOutbound(mail mail.Mail) ([]byte, error) {
        return []byte{}, nil
    }
    
    adapter := &mockAdapter{name: "test-adapter"}
    
    err := gateway.RegisterAdapter(adapter)
    if err != nil {
        t.Errorf("Expected nil error, got %v", err)
    }
    
    // Verify adapter was registered
    retrieved, err := gateway.GetAdapter("test-adapter")
    if err != nil {
        t.Errorf("Expected nil error, got %v", err)
    }
    if retrieved != adapter {
        t.Error("Expected same adapter instance")
    }
    
    // Verify in list
    names := gateway.ListAdapters()
    found := false
    for _, n := range names {
        if n == "test-adapter" {
            found = true
            break
        }
    }
    if !found {
        t.Error("Expected test-adapter in list")
    }
}
```
**Acceptance Criteria:**
- Gateway can register adapters
- Adapters stored by name
- Returns nil error on success
- GetAdapter retrieves registered adapter
- ListAdapters includes registered adapter

### Test 2: TestWebhookAdapter_NormalizeInbound
```go
func TestWebhookAdapter_NormalizeInbound(t *testing.T) {
    adapter := NewWebhookAdapter()
    
    // Test HTTP POST data normalization
    inputData := []byte(`{"text": "hello world", "user": "alice"}`)
    
    mail, err := adapter.NormalizeInbound(inputData)
    if err != nil {
        t.Errorf("Expected nil error, got %v", err)
    }
    
    // Verify mail structure
    if mail.Type != mail.MailTypeMailReceived {
        t.Errorf("Expected MailTypeMailReceived, got %s", mail.Type)
    }
    
    if mail.Source != "gateway:webhook" {
        t.Errorf("Expected source 'gateway:webhook', got '%s'", mail.Source)
    }
    
    if mail.Metadata.Boundary != mail.OuterBoundary {
        t.Errorf("Expected OuterBoundary, got %s", mail.Metadata.Boundary)
    }
    
    if len(mail.Metadata.Taints) == 0 || mail.Metadata.Taints[0] != "USER_SUPPLIED" {
        t.Error("Expected USER_SUPPLIED taint")
    }
    
    // Verify content extracted
    content, ok := mail.Content.(map[string]any)
    if !ok {
        t.Error("Expected content to be map[string]any")
    }
    if content["text"] != "hello world" {
        t.Error("Expected text to be 'hello world'")
    }
}
```
**Acceptance Criteria:**
- Webhook adapter normalizes HTTP POST to mail_received
- Source set to gateway:webhook
- Content extracted from request body
- Taints set to USER_SUPPLIED
- Boundary set to OuterBoundary

## Dependencies

### Test Dependencies
```
Test 1 → Test 2 (Register before use)
```

### Phase Dependencies
- **Phase 3.1** must complete first (Mail types)
- **Phase 3.5** must complete first (Streaming for outbound)
- **Phase 3.6b, 3.6c** are sibling phases

## Files to Create/Modify

| File | Operation | Changes |
|------|-----------|---------|
| `pkg/gateway/gateway.go` | CREATE | Gateway service and Adapter interface |
| `pkg/gateway/adapters/webhook.go` | CREATE | Webhook adapter |

## Implementation Order (TDD)

1. **Test 1**: Write test → verify RED → Implement Gateway and Adapter interface → verify GREEN → commit
2. **Test 2**: Write test → verify RED → Implement Webhook adapter → verify GREEN → commit

**Total: 2 tests, 2 commits**

## Deliverables
- ✅ Gateway service in `pkg/gateway/gateway.go`
- ✅ Adapter interface
- ✅ Webhook adapter in `pkg/gateway/adapters/webhook.go`
- ✅ 2 commits (one per test per TDD workflow)

## Recursive Analysis Decision

**Decision: Split from original Phase 3.6**

**Reasoning:**
- Original Phase 3.6 had 6 tests across multiple adapters
- Gateway core + Webhook is a logical starting point
- 2 tests is within recommended range (2-5)
- Can be developed independently of other adapters

**Benefits of split:**
- Gateway core can be tested first
- Webhook is simplest adapter (HTTP POST)
- Other adapters can follow the pattern
- Parallel development possible (3.6a, 3.6b, 3.6c)