package gateway

import (
	"testing"

	mailpkg "github.com/maelstrom/v3/pkg/mail"
)

// TestWebhookAdapter_InboundNormalization tests HTTP POST to mail_received conversion
// Spec Reference: arch-v1.md L659-666 (Channel Adapters), L670-671 (normalize inbound to mail_received)
func TestWebhookAdapter_InboundNormalization(t *testing.T) {
	adapter := &WebhookAdapter{}

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

	content, ok := mailMsg.Content.(map[string]any)
	if !ok {
		t.Fatalf("Expected content to be map[string]any, got %T", mailMsg.Content)
	}

	if content["from"] != "sender@example.com" {
		t.Errorf("Expected from 'sender@example.com', got %v", content["from"])
	}
}

func TestChannelAdapter_WebhookNormalizesToMail(t *testing.T) {
	adapter := &WebhookAdapter{}

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

	if adapter.Stream() {
		t.Error("Expected Stream() to return false for webhook")
	}

	outboundMail := &mailpkg.Mail{
		Type:    mailpkg.MailSend,
		Content: map[string]any{"response": "acknowledged"},
	}

	normalized, err := adapter.NormalizeOutbound(outboundMail)
	if err != nil {
		t.Fatalf("NormalizeOutbound failed: %v", err)
	}

	if normalized == nil {
		t.Error("Expected normalized outbound content")
	}
}
