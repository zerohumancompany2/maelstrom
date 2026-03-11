package gateway

import (
	"testing"
	"time"

	mailpkg "github.com/maelstrom/v3/pkg/mail"
)

// TestSSEAdapter_NormalizeInboundWithTaints tests that inbound messages are properly tainted
// Spec Reference: arch-v1.md L279-283 (Data Tainting), L662 (sse - Server-Sent Events)
func TestSSEAdapter_NormalizeInboundWithTaints(t *testing.T) {
	adapter := NewSSEAdapter(8082)

	rawMessage := map[string]any{
		"event": "user_input",
		"data":  "Hello",
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

// TestSSEAdapter_NormalizeInboundNilError tests nil input handling
func TestSSEAdapter_NormalizeInboundNilError(t *testing.T) {
	adapter := NewSSEAdapter(8082)

	_, err := adapter.NormalizeInbound(nil)
	if err == nil {
		t.Error("Expected error for nil input")
	}
}

// TestSSEAdapter_NormalizeOutboundNilError tests nil mail handling
func TestSSEAdapter_NormalizeOutboundNilError(t *testing.T) {
	adapter := NewSSEAdapter(8082)

	_, err := adapter.NormalizeOutbound(nil)
	if err == nil {
		t.Error("Expected error for nil mail")
	}
}

// TestSSEAdapter_FirewallFriendly tests SSE event formatting on single HTTP connection
// Spec Reference: arch-v1.md L662 (sse - Server-Sent Events firewall-friendly), L670-671 (normalize outbound)
func TestSSEAdapter_FirewallFriendly(t *testing.T) {
	adapter := &SSEAdapter{}

	// Test outbound normalization for SSE format
	mailMsg := &mailpkg.Mail{
		Type: mailpkg.Assistant,
		Content: map[string]any{
			"event": "update",
			"data":  "Server event data",
		},
		Metadata: mailpkg.MailMetadata{
			Adapter: "sse",
		},
	}

	sseEvent, err := adapter.NormalizeOutbound(mailMsg)
	if err != nil {
		t.Fatalf("NormalizeOutbound failed: %v", err)
	}

	// SSE should format events correctly
	sseMap, ok := sseEvent.(map[string]any)
	if !ok {
		t.Fatalf("Expected SSE event to be map[string]any, got %T", sseEvent)
	}

	// The event field in the output is the mail type, but the data contains the original content
	if sseMap["data"] == nil {
		t.Error("Expected data field in SSE event")
	}

	// Verify streaming is enabled for SSE
	if !adapter.Stream() {
		t.Errorf("Expected SSEAdapter to stream, got false")
	}

	// Verify adapter name
	if adapter.Name() != "sse" {
		t.Errorf("Expected adapter name 'sse', got %v", adapter.Name())
	}
}

// TestSSEAdapter_StreamEnabled tests that streaming is enabled
func TestSSEAdapter_StreamEnabled(t *testing.T) {
	adapter := NewSSEAdapter(8082)

	if !adapter.Stream() {
		t.Error("Expected Stream() to return true for sse")
	}
}

// TestSSEAdapter_MailStructure tests mail structure completeness
// Spec Reference: arch-v1.md L170-186 (Message format)
func TestSSEAdapter_MailStructure(t *testing.T) {
	adapter := NewSSEAdapter(8082)

	rawMessage := map[string]any{
		"event": "test",
		"data":  "test data",
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

	if mailMsg.Metadata.Adapter != "sse" {
		t.Errorf("Expected adapter 'sse', got %v", mailMsg.Metadata.Adapter)
	}

	if mailMsg.Metadata.Boundary != mailpkg.OuterBoundary {
		t.Errorf("Expected boundary outer, got %v", mailMsg.Metadata.Boundary)
	}

	if !mailMsg.Metadata.Stream {
		t.Error("Expected Stream to be true")
	}
}

// TestChannelAdapter_SSEServerSentEvents tests SSE adapter normalization
func TestChannelAdapter_SSEServerSentEvents(t *testing.T) {
	adapter := &SSEAdapter{}

	inboundMessage := map[string]any{
		"event": "user_input",
		"data":  "Hello",
	}

	mailMsg, err := adapter.NormalizeInbound(inboundMessage)
	if err != nil {
		t.Fatalf("NormalizeInbound failed: %v", err)
	}

	if mailMsg.Type != mailpkg.MailReceived {
		t.Errorf("Expected type mail_received, got %v", mailMsg.Type)
	}

	if mailMsg.Metadata.Adapter != "sse" {
		t.Errorf("Expected adapter 'sse', got %v", mailMsg.Metadata.Adapter)
	}

	if !adapter.Stream() {
		t.Error("Expected Stream() to return true for sse")
	}

	outboundMail := &mailpkg.Mail{
		Type:    mailpkg.MailSend,
		Content: map[string]any{"text": "Server response"},
	}

	normalized, err := adapter.NormalizeOutbound(outboundMail)
	if err != nil {
		t.Fatalf("NormalizeOutbound failed: %v", err)
	}

	if normalized == nil {
		t.Error("Expected normalized SSE content")
	}
}

// TestSSEAdapter_PortConfiguration tests port configuration
func TestSSEAdapter_PortConfiguration(t *testing.T) {
	port := 9092
	adapter := NewSSEAdapter(port)

	if adapter.Name() != "sse" {
		t.Errorf("Expected name 'sse', got %v", adapter.Name())
	}

	if !adapter.Stream() {
		t.Error("Expected Stream() to return true")
	}
}

// TestSSEAdapter_BoundaryValidation tests boundary is set correctly
func TestSSEAdapter_BoundaryValidation(t *testing.T) {
	adapter := NewSSEAdapter(8082)

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

// TestSSEAdapter_CreatedAtValidation tests CreatedAt is set
func TestSSEAdapter_CreatedAtValidation(t *testing.T) {
	adapter := NewSSEAdapter(8082)

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
