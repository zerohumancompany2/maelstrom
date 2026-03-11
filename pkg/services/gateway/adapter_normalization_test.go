package gateway

import (
	"testing"

	mailpkg "github.com/maelstrom/v3/pkg/mail"
)

// TestAdapterNormalization_Webhook tests adapter normalization for webhook
// Spec Reference: arch-v1.md L659-666 (Channel Adapters), L670-671 (normalize inbound to mail_received)
func TestAdapterNormalization_Webhook(t *testing.T) {
	adapter := NewWebhookAdapter(8080)

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

	if mailMsg.Type != mailpkg.MailReceived {
		t.Errorf("Expected type mail_received, got %v", mailMsg.Type)
	}

	if mailMsg.Metadata.Adapter != "webhook" {
		t.Errorf("Expected adapter 'webhook', got %v", mailMsg.Metadata.Adapter)
	}
}

// TestAdapterNormalization_WebSocket tests adapter normalization for websocket
// Spec Reference: arch-v1.md L661 (websocket - Full bidirectional), L670-671 (normalize inbound)
func TestAdapterNormalization_WebSocket(t *testing.T) {
	adapter := NewWebSocketAdapter(8081)

	rawMessage := map[string]any{
		"type": "message",
		"data": "Hello from client",
	}

	mailMsg, err := adapter.NormalizeInbound(rawMessage)
	if err != nil {
		t.Fatalf("NormalizeInbound failed: %v", err)
	}

	if mailMsg.Type != mailpkg.MailReceived {
		t.Errorf("Expected type mail_received, got %v", mailMsg.Type)
	}

	if mailMsg.Metadata.Adapter != "websocket" {
		t.Errorf("Expected adapter 'websocket', got %v", mailMsg.Metadata.Adapter)
	}
}

// TestAdapterNormalization_SSE tests adapter normalization for SSE
// Spec Reference: arch-v1.md L662 (sse - Server-Sent Events), L670-671 (normalize inbound)
func TestAdapterNormalization_SSE(t *testing.T) {
	adapter := NewSSEAdapter(8082)

	rawMessage := map[string]any{
		"event": "user_input",
		"data":  "Hello",
	}

	mailMsg, err := adapter.NormalizeInbound(rawMessage)
	if err != nil {
		t.Fatalf("NormalizeInbound failed: %v", err)
	}

	if mailMsg.Type != mailpkg.MailReceived {
		t.Errorf("Expected type mail_received, got %v", mailMsg.Type)
	}

	if mailMsg.Metadata.Adapter != "sse" {
		t.Errorf("Expected adapter 'sse', got %v", mailMsg.Metadata.Adapter)
	}
}

// TestAdapterNormalization_TaintsApplied tests that taints are applied correctly
// Spec Reference: arch-v1.md L279-283 (Data Tainting)
func TestAdapterNormalization_TaintsApplied(t *testing.T) {
	webhookAdapter := NewWebhookAdapter(8080)
	wsAdapter := NewWebSocketAdapter(8081)
	sseAdapter := NewSSEAdapter(8082)

	adapters := []ChannelAdapter{webhookAdapter, wsAdapter, sseAdapter}

	for _, adapter := range adapters {
		rawMessage := map[string]any{"test": "data"}

		mailMsg, err := adapter.NormalizeInbound(rawMessage)
		if err != nil {
			t.Fatalf("%s NormalizeInbound failed: %v", adapter.Name(), err)
		}

		if len(mailMsg.Taints) == 0 {
			t.Errorf("%s: Expected taints to be applied", adapter.Name())
		}

		if len(mailMsg.Metadata.Taints) == 0 {
			t.Errorf("%s: Expected metadata taints to be applied", adapter.Name())
		}
	}
}

// TestAdapterNormalization_BoundarySet tests that boundary is set correctly
// Spec Reference: arch-v1.md L267-274 (Boundary Model)
func TestAdapterNormalization_BoundarySet(t *testing.T) {
	webhookAdapter := NewWebhookAdapter(8080)
	wsAdapter := NewWebSocketAdapter(8081)
	sseAdapter := NewSSEAdapter(8082)

	adapters := []ChannelAdapter{webhookAdapter, wsAdapter, sseAdapter}

	for _, adapter := range adapters {
		rawMessage := map[string]any{"test": "data"}

		mailMsg, err := adapter.NormalizeInbound(rawMessage)
		if err != nil {
			t.Fatalf("%s NormalizeInbound failed: %v", adapter.Name(), err)
		}

		if mailMsg.Metadata.Boundary != mailpkg.OuterBoundary {
			t.Errorf("%s: Expected boundary outer, got %v", adapter.Name(), mailMsg.Metadata.Boundary)
		}
	}
}

// TestAdapterNormalization_RequiredFields tests that all required fields are present
// Spec Reference: arch-v1.md L170-186 (Message format)
func TestAdapterNormalization_RequiredFields(t *testing.T) {
	webhookAdapter := NewWebhookAdapter(8080)
	wsAdapter := NewWebSocketAdapter(8081)
	sseAdapter := NewSSEAdapter(8082)

	adapters := []ChannelAdapter{webhookAdapter, wsAdapter, sseAdapter}

	for _, adapter := range adapters {
		rawMessage := map[string]any{"test": "data"}

		mailMsg, err := adapter.NormalizeInbound(rawMessage)
		if err != nil {
			t.Fatalf("%s NormalizeInbound failed: %v", adapter.Name(), err)
		}

		if mailMsg.ID == "" {
			t.Errorf("%s: Expected non-empty ID", adapter.Name())
		}

		if mailMsg.CorrelationID == "" {
			t.Errorf("%s: Expected non-empty CorrelationID", adapter.Name())
		}

		if mailMsg.Source == "" {
			t.Errorf("%s: Expected non-empty Source", adapter.Name())
		}
	}
}
