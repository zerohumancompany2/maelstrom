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
