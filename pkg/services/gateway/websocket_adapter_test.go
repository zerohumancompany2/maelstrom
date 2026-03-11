package gateway

import (
	"testing"

	mailpkg "github.com/maelstrom/v3/pkg/mail"
)

// TestWebSocketAdapter_Bidirectional tests bidirectional message flow with connection state
// Spec Reference: arch-v1.md L661 (websocket - Full bidirectional), L670-671 (normalize inbound/outbound)
func TestWebSocketAdapter_Bidirectional(t *testing.T) {
	adapter := &WebSocketAdapter{}

	// Test inbound normalization
	inboundMsg := map[string]any{
		"type": "message",
		"data": "Hello from client",
	}

	mailMsg, err := adapter.NormalizeInbound(inboundMsg)
	if err != nil {
		t.Fatalf("NormalizeInbound failed: %v", err)
	}

	if mailMsg.Type != mailpkg.MailReceived {
		t.Errorf("Expected type mail_received, got %v", mailMsg.Type)
	}

	if mailMsg.Metadata.Adapter != "websocket" {
		t.Errorf("Expected adapter 'websocket', got %v", mailMsg.Metadata.Adapter)
	}

	// Test outbound normalization
	outboundMail := &mailpkg.Mail{
		Type: mailpkg.Assistant,
		Content: map[string]any{
			"type": "response",
			"data": "Hello from server",
		},
		Metadata: mailpkg.MailMetadata{
			Adapter: "websocket",
		},
	}

	outboundMsg, err := adapter.NormalizeOutbound(outboundMail)
	if err != nil {
		t.Fatalf("NormalizeOutbound failed: %v", err)
	}

	outboundMap, ok := outboundMsg.(map[string]any)
	if !ok {
		t.Fatalf("Expected outbound to be map[string]any, got %T", outboundMsg)
	}

	if outboundMap["type"] != "response" {
		t.Errorf("Expected type 'response', got %v", outboundMap["type"])
	}

	// Verify streaming is enabled for websocket
	if !adapter.Stream() {
		t.Errorf("Expected WebSocketAdapter to stream, got false")
	}
}

func TestChannelAdapter_WebSocketBidirectional(t *testing.T) {
	adapter := &WebSocketAdapter{}

	inboundMessage := map[string]any{
		"text":      "Hello from WebSocket client",
		"timestamp": 1234567890,
	}

	mailMsg, err := adapter.NormalizeInbound(inboundMessage)
	if err != nil {
		t.Fatalf("NormalizeInbound failed: %v", err)
	}

	if mailMsg.Type != mailpkg.MailReceived {
		t.Errorf("Expected type mail_received, got %v", mailMsg.Type)
	}

	if mailMsg.Metadata.Adapter != "websocket" {
		t.Errorf("Expected adapter 'websocket', got %v", mailMsg.Metadata.Adapter)
	}

	if !adapter.Stream() {
		t.Error("Expected Stream() to return true for websocket")
	}

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
