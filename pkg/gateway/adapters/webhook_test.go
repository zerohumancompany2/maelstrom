package adapters

import (
	"testing"

	"github.com/maelstrom/v3/pkg/mail"
)

func TestWebhookAdapter_NormalizeInbound(t *testing.T) {
	adapter := NewWebhookAdapter()

	// Test HTTP POST data normalization
	inputData := []byte(`{"text": "hello world", "user": "alice"}`)

	msg, err := adapter.NormalizeInbound(inputData)
	if err != nil {
		t.Errorf("Expected nil error, got %v", err)
	}

	// Verify mail structure
	if msg.Type != mail.MailTypeMailReceived {
		t.Errorf("Expected MailTypeMailReceived, got %s", msg.Type)
	}

	if msg.Source != "gateway:webhook" {
		t.Errorf("Expected source 'gateway:webhook', got '%s'", msg.Source)
	}

	if msg.Metadata.Boundary != mail.OuterBoundary {
		t.Errorf("Expected OuterBoundary, got %s", msg.Metadata.Boundary)
	}

	if len(msg.Metadata.Taints) == 0 || msg.Metadata.Taints[0] != "USER_SUPPLIED" {
		t.Error("Expected USER_SUPPLIED taint")
	}

	// Verify content extracted
	content, ok := msg.Content.(map[string]any)
	if !ok {
		t.Error("Expected content to be map[string]any")
	}
	if content["text"] != "hello world" {
		t.Error("Expected text to be 'hello world'")
	}
}
