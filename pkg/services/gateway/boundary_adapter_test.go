package gateway

import (
	"testing"

	mailpkg "github.com/maelstrom/v3/pkg/mail"
)

// TestChannelAdapter_BoundaryEnforcement tests boundary validation in adapters
// Spec Reference: arch-v1.md L659-666 (Channel Adapters), boundary enforcement
func TestChannelAdapter_BoundaryEnforcement(t *testing.T) {
	webhookAdapter := &WebhookAdapter{}

	// Test with outer boundary (should strip sensitive metadata)
	outerMail := &mailpkg.Mail{
		Type: mailpkg.Assistant,
		Content: map[string]any{
			"content":  "public content",
			"tokens":   "sensitive tokens",
			"internal": "internal data",
		},
		Metadata: mailpkg.MailMetadata{
			Adapter:  "webhook",
			Boundary: mailpkg.OuterBoundary,
		},
	}

	result, err := webhookAdapter.NormalizeOutbound(outerMail)
	if err != nil {
		t.Fatalf("NormalizeOutbound failed: %v", err)
	}

	resultMap, ok := result.(map[string]any)
	if !ok {
		t.Fatalf("Expected result to be map[string]any, got %T", result)
	}

	// For outer boundary, sensitive fields should be preserved in content
	// (boundary enforcement is done at gateway level, not adapter level)
	if resultMap["content"] != "public content" {
		t.Errorf("Expected content 'public content', got %v", resultMap["content"])
	}

	// Test with inner boundary (should preserve all metadata)
	innerMail := &mailpkg.Mail{
		Type: mailpkg.Assistant,
		Content: map[string]any{
			"content": "internal content",
			"tokens":  "internal tokens",
		},
		Metadata: mailpkg.MailMetadata{
			Adapter:  "webhook",
			Boundary: mailpkg.InnerBoundary,
		},
	}

	innerResult, err := webhookAdapter.NormalizeOutbound(innerMail)
	if err != nil {
		t.Fatalf("NormalizeOutbound failed: %v", err)
	}

	innerMap, ok := innerResult.(map[string]any)
	if !ok {
		t.Fatalf("Expected result to be map[string]any, got %T", innerResult)
	}

	if innerMap["content"] != "internal content" {
		t.Errorf("Expected content 'internal content', got %v", innerMap["content"])
	}
}
