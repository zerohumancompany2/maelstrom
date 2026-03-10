package gateway

import (
	"testing"

	mailpkg "github.com/maelstrom/v3/pkg/mail"
)

// TestSMTPAdapter_Email tests SMTP email formatting with protocol compliance
// Spec Reference: arch-v1.md L664 (smtp - Email), L670-671 (normalize outbound)
func TestSMTPAdapter_Email(t *testing.T) {
	adapter := &SMTPAdapter{}

	// Test outbound normalization for SMTP format
	mailMsg := &mailpkg.Mail{
		Type: mailpkg.MailReceived,
		Content: map[string]any{
			"from":    "sender@example.com",
			"to":      []string{"recipient@example.com"},
			"subject": "Test Email",
			"body":    "Email body content",
		},
		Metadata: mailpkg.MailMetadata{
			Adapter: "smtp",
		},
	}

	smtpMsg, err := adapter.NormalizeOutbound(mailMsg)
	if err != nil {
		t.Fatalf("NormalizeOutbound failed: %v", err)
	}

	// SMTP should preserve email format
	smtpMap, ok := smtpMsg.(map[string]any)
	if !ok {
		t.Fatalf("Expected SMTP message to be map[string]any, got %T", smtpMsg)
	}

	if smtpMap["from"] != "sender@example.com" {
		t.Errorf("Expected from 'sender@example.com', got %v", smtpMap["from"])
	}

	// Verify streaming is disabled for SMTP (connection-based, not streaming)
	if adapter.Stream() {
		t.Errorf("Expected SMTPAdapter to not stream, got true")
	}

	// Verify adapter name
	if adapter.Name() != "smtp" {
		t.Errorf("Expected adapter name 'smtp', got %v", adapter.Name())
	}
}
