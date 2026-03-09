# Phase 3.6C: Stub Adapters & Integration

## Goal
Create stub implementations for remaining adapters and verify integration.

## Scope
- Create `pkg/gateway/adapters/pubsub.go` stub
- Create `pkg/gateway/adapters/smtp.go` stub
- Create `pkg/gateway/adapters/platform.go` stub (Slack, WhatsApp, Telegram)
- Integration test for adapter normalization round-trip
- 2 tests, 2 commits

## Current Status

### Implementation Status
| Adapter | Status | Notes |
|---------|--------|-------|
| `pubsub` | ❌ Missing | Message queue |
| `smtp` | ❌ Missing | Email |
| `platform` | ❌ Missing | Slack, WhatsApp, Telegram |

### Files Status
| File | Status |
|------|-------|
| `pkg/gateway/adapters/pubsub.go` | ❌ MISSING - create |
| `pkg/gateway/adapters/smtp.go` | ❌ MISSING - create |
| `pkg/gateway/adapters/platform.go` | ❌ MISSING - create |

## Required Implementation

### PubSub Adapter Stub
```go
// pkg/gateway/adapters/pubsub.go
package adapters

import "maelstrom/pkg/mail"

type PubSubAdapter struct {
    name string
}

func NewPubSubAdapter() *PubSubAdapter {
    return &PubSubAdapter{name: "pubsub"}
}

func (a *PubSubAdapter) Name() string {
    return a.name
}

func (a *PubSubAdapter) NormalizeInbound(data []byte) (mail.Mail, error) {
    return mail.Mail{
        ID:     generateID(),
        Type:   mail.MailTypeMailReceived,
        Source: "gateway:pubsub",
        Content: string(data),
        Metadata: mail.MailMetadata{
            Boundary: mail.OuterBoundary,
            Taints:   []string{"USER_SUPPLIED"},
        },
    }, nil
}

func (a *PubSubAdapter) NormalizeOutbound(mail mail.Mail) ([]byte, error) {
    return []byte(mail.Content.(string)), nil
}
```

### SMTP Adapter Stub
```go
// pkg/gateway/adapters/smtp.go
package adapters

import "maelstrom/pkg/mail"

type SMTPAdapter struct {
    name string
}

func NewSMTPAdapter() *SMTPAdapter {
    return &SMTPAdapter{name: "smtp"}
}

func (a *SMTPAdapter) Name() string {
    return a.name
}

func (a *SMTPAdapter) NormalizeInbound(data []byte) (mail.Mail, error) {
    return mail.Mail{
        ID:     generateID(),
        Type:   mail.MailTypeMailReceived,
        Source: "gateway:smtp",
        Content: string(data),
        Metadata: mail.MailMetadata{
            Boundary: mail.OuterBoundary,
            Taints:   []string{"USER_SUPPLIED"},
        },
    }, nil
}

func (a *SMTPAdapter) NormalizeOutbound(mail mail.Mail) ([]byte, error) {
    return []byte(mail.Content.(string)), nil
}
```

### Platform Adapter Stub
```go
// pkg/gateway/adapters/platform.go
package adapters

import "maelstrom/pkg/mail"

type PlatformAdapter struct {
    name string
}

func NewPlatformAdapter(platform string) *PlatformAdapter {
    return &PlatformAdapter{name: platform}
}

func (a *PlatformAdapter) Name() string {
    return a.name
}

func (a *PlatformAdapter) NormalizeInbound(data []byte) (mail.Mail, error) {
    return mail.Mail{
        ID:     generateID(),
        Type:   mail.MailTypeMailReceived,
        Source: "gateway:" + a.name,
        Content: string(data),
        Metadata: mail.MailMetadata{
            Boundary: mail.OuterBoundary,
            Taints:   []string{"USER_SUPPLIED"},
        },
    }, nil
}

func (a *PlatformAdapter) NormalizeOutbound(mail mail.Mail) ([]byte, error) {
    return []byte(mail.Content.(string)), nil
}

// Convenience constructors
func NewSlackAdapter() *PlatformAdapter {
    return NewPlatformAdapter("slack")
}

func NewWhatsAppAdapter() *PlatformAdapter {
    return NewPlatformAdapter("whatsapp")
}

func NewTelegramAdapter() *PlatformAdapter {
    return NewPlatformAdapter("telegram")
}
```

## Tests to Write (2 tests, 2 commits)

### Test 1: TestGateway_AdapterNotFound
```go
func TestGateway_AdapterNotFound(t *testing.T) {
    gateway := NewGateway()
    
    // Try to get non-registered adapter
    _, err := gateway.GetAdapter("non-existent")
    if err == nil {
        t.Error("Expected error for non-existent adapter")
    }
    
    if err.Error() != "adapter not found: non-existent" {
        t.Errorf("Expected 'adapter not found: non-existent', got '%s'", err.Error())
    }
    
    // Verify empty list
    names := gateway.ListAdapters()
    if len(names) != 0 {
        t.Errorf("Expected empty list, got %d adapters", len(names))
    }
}
```
**Acceptance Criteria:**
- Using non-registered adapter returns error
- Error indicates adapter not found
- ListAdapters returns empty list when no adapters registered

### Test 2: TestAdapter_NormalizationRoundTrip
```go
func TestAdapter_NormalizationRoundTrip(t *testing.T) {
    gateway := NewGateway()
    
    // Register all adapters
    adapters := []Adapter{
        NewWebhookAdapter(),
        NewSSEAdapter(),
        NewWebSocketAdapter(),
        NewPubSubAdapter(),
        NewSMTPAdapter(),
        NewSlackAdapter(),
        NewWhatsAppAdapter(),
        NewTelegramAdapter(),
    }
    
    for _, adapter := range adapters {
        err := gateway.RegisterAdapter(adapter)
        if err != nil {
            t.Errorf("Failed to register %s: %v", adapter.Name(), err)
        }
    }
    
    // Verify all adapters registered
    names := gateway.ListAdapters()
    if len(names) != len(adapters) {
        t.Errorf("Expected %d adapters, got %d", len(adapters), len(names))
    }
    
    // Test round-trip for each adapter
    for _, adapter := range adapters {
        // Create test mail
        originalMail := mail.Mail{
            ID:     "msg-001",
            Type:   mail.MailTypeUser,
            Source: "test",
            Content: "test content",
            Metadata: mail.MailMetadata{
                Boundary: mail.OuterBoundary,
            },
        }
        
        // Normalize outbound
        outbound, err := adapter.NormalizeOutbound(originalMail)
        if err != nil {
            t.Errorf("Adapter %s: NormalizeOutbound failed: %v", adapter.Name(), err)
            continue
        }
        
        // Normalize inbound (simulating response)
        inbound, err := adapter.NormalizeInbound(outbound)
        if err != nil {
            t.Errorf("Adapter %s: NormalizeInbound failed: %v", adapter.Name(), err)
            continue
        }
        
        // Verify mail type preserved
        if inbound.Type != mail.MailTypeMailReceived {
            t.Errorf("Adapter %s: Expected MailTypeMailReceived, got %s", 
                adapter.Name(), inbound.Type)
        }
        
        // Verify source set correctly
        expectedSource := "gateway:" + adapter.Name()
        if inbound.Source != expectedSource {
            t.Errorf("Adapter %s: Expected source '%s', got '%s'", 
                adapter.Name(), expectedSource, inbound.Source)
        }
    }
}
```
**Acceptance Criteria:**
- Inbound then outbound preserves content structure
- Mail type preserved through normalization
- All adapters can be registered and used
- Each adapter has correct source prefix

## Dependencies

### Test Dependencies
```
Test 1 → Independent
Test 2 → Test 1 (All adapters before integration)
```

### Phase Dependencies
- **Phase 3.1** must complete first (Mail types)
- **Phase 3.6a** must complete first (Gateway core, Webhook)
- **Phase 3.6b** must complete first (SSE, WebSocket)

## Files to Create/Modify

| File | Operation | Changes |
|------|-----------|---------|
| `pkg/gateway/adapters/pubsub.go` | CREATE | PubSub adapter stub |
| `pkg/gateway/adapters/smtp.go` | CREATE | SMTP adapter stub |
| `pkg/gateway/adapters/platform.go` | CREATE | Platform adapter stub (Slack, WhatsApp, Telegram) |

## Implementation Order (TDD)

1. **Test 1**: Write test → verify RED → Add error handling for missing adapters → verify GREEN → commit
2. **Test 2**: Write test → verify RED → Implement all stub adapters → verify GREEN → commit

**Total: 2 tests, 2 commits**

## Deliverables
- ✅ PubSub adapter stub in `pkg/gateway/adapters/pubsub.go`
- ✅ SMTP adapter stub in `pkg/gateway/adapters/smtp.go`
- ✅ Platform adapter stub in `pkg/gateway/adapters/platform.go`
- ✅ All 8 adapters registered and working
- ✅ 2 commits (one per test per TDD workflow)

## Recursive Analysis Decision

**Decision: Split from original Phase 3.6**

**Reasoning:**
- Stub adapters are simpler implementations
- Can be created quickly after core adapters
- 2 tests is within recommended range (2-5)
- Integration test verifies all adapters work together

**Benefits of split:**
- Core adapters (3.6a, 3.6b) can be developed first
- Stub adapters follow the same pattern
- Integration test ensures all adapters work
- Clear progression from complex to simple