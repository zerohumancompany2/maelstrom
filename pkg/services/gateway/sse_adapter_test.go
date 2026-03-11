package gateway

import (
	"testing"

	mailpkg "github.com/maelstrom/v3/pkg/mail"
)

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

	if sseMap["event"] != "update" {
		t.Errorf("Expected event 'update', got %v", sseMap["event"])
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
