package gateway

import (
	"testing"
	"time"

	mailpkg "github.com/maelstrom/v3/pkg/mail"
)

// TestWebSocketAdapter_NormalizeInboundWithTaints tests that inbound messages are properly tainted
// Spec Reference: arch-v1.md L279-283 (Data Tainting), L661 (websocket - Full bidirectional)
func TestWebSocketAdapter_NormalizeInboundWithTaints(t *testing.T) {
	adapter := NewWebSocketAdapter(8081)

	rawMessage := map[string]any{
		"type": "message",
		"data": "Hello from client",
	}

	mailMsg, err := adapter.NormalizeInbound(rawMessage)
	if err != nil {
		t.Fatalf("NormalizeInbound failed: %v", err)
	}

	expectedTaints := []string{TaintUserSupplied, TaintExternal}
	if len(mailMsg.Taints) != len(expectedTaints) {
		t.Errorf("Expected %d taints, got %d: %v", len(expectedTaints), len(mailMsg.Taints), mailMsg.Taints)
	}

	for _, expectedTaint := range expectedTaints {
		found := false
		for _, taint := range mailMsg.Taints {
			if taint == expectedTaint {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected taint %s not found", expectedTaint)
		}
	}
}

// TestWebSocketAdapter_NormalizeInboundNilError tests nil input handling
func TestWebSocketAdapter_NormalizeInboundNilError(t *testing.T) {
	adapter := NewWebSocketAdapter(8081)

	_, err := adapter.NormalizeInbound(nil)
	if err == nil {
		t.Error("Expected error for nil input")
	}
}

// TestWebSocketAdapter_NormalizeOutboundNilError tests nil mail handling
func TestWebSocketAdapter_NormalizeOutboundNilError(t *testing.T) {
	adapter := NewWebSocketAdapter(8081)

	_, err := adapter.NormalizeOutbound(nil)
	if err == nil {
		t.Error("Expected error for nil mail")
	}
}

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

// TestWebSocketAdapter_StreamEnabled tests that streaming is enabled
func TestWebSocketAdapter_StreamEnabled(t *testing.T) {
	adapter := NewWebSocketAdapter(8081)

	if !adapter.Stream() {
		t.Error("Expected Stream() to return true for websocket")
	}
}

// TestWebSocketAdapter_MailStructure tests mail structure completeness
// Spec Reference: arch-v1.md L170-186 (Message format)
func TestWebSocketAdapter_MailStructure(t *testing.T) {
	adapter := NewWebSocketAdapter(8081)

	rawMessage := map[string]any{
		"type": "message",
		"data": "test",
	}

	mailMsg, err := adapter.NormalizeInbound(rawMessage)
	if err != nil {
		t.Fatalf("NormalizeInbound failed: %v", err)
	}

	if mailMsg.ID == "" {
		t.Error("Expected non-empty ID")
	}

	if mailMsg.CorrelationID == "" {
		t.Error("Expected non-empty CorrelationID")
	}

	if mailMsg.Type != mailpkg.MailReceived {
		t.Errorf("Expected type mail_received, got %v", mailMsg.Type)
	}

	if mailMsg.CreatedAt.IsZero() {
		t.Error("Expected non-zero CreatedAt")
	}

	if mailMsg.Source != "gateway" {
		t.Errorf("Expected source 'gateway', got %v", mailMsg.Source)
	}

	if mailMsg.Metadata.Adapter != "websocket" {
		t.Errorf("Expected adapter 'websocket', got %v", mailMsg.Metadata.Adapter)
	}

	if mailMsg.Metadata.Boundary != mailpkg.OuterBoundary {
		t.Errorf("Expected boundary outer, got %v", mailMsg.Metadata.Boundary)
	}

	if !mailMsg.Metadata.Stream {
		t.Error("Expected Stream to be true")
	}
}

// TestChannelAdapter_WebSocketBidirectional tests websocket adapter normalization
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

// TestWebSocketAdapter_PortConfiguration tests port configuration
func TestWebSocketAdapter_PortConfiguration(t *testing.T) {
	port := 9091
	adapter := NewWebSocketAdapter(port)

	if adapter.Name() != "websocket" {
		t.Errorf("Expected name 'websocket', got %v", adapter.Name())
	}

	if !adapter.Stream() {
		t.Error("Expected Stream() to return true")
	}
}

// TestWebSocketAdapter_BoundaryValidation tests boundary is set correctly
func TestWebSocketAdapter_BoundaryValidation(t *testing.T) {
	adapter := NewWebSocketAdapter(8081)

	rawMessage := map[string]any{
		"body": "test",
	}

	mailMsg, err := adapter.NormalizeInbound(rawMessage)
	if err != nil {
		t.Fatalf("NormalizeInbound failed: %v", err)
	}

	if mailMsg.Metadata.Boundary != mailpkg.OuterBoundary {
		t.Errorf("Expected boundary outer, got %v", mailMsg.Metadata.Boundary)
	}
}

// TestWebSocketAdapter_CreatedAtValidation tests CreatedAt is set
func TestWebSocketAdapter_CreatedAtValidation(t *testing.T) {
	adapter := NewWebSocketAdapter(8081)

	before := time.Now()
	rawMessage := map[string]any{"body": "test"}

	mailMsg, err := adapter.NormalizeInbound(rawMessage)
	if err != nil {
		t.Fatalf("NormalizeInbound failed: %v", err)
	}

	after := time.Now()

	if mailMsg.CreatedAt.Before(before) || mailMsg.CreatedAt.After(after) {
		t.Errorf("CreatedAt %v not within expected range [%v, %v]", mailMsg.CreatedAt, before, after)
	}
}
