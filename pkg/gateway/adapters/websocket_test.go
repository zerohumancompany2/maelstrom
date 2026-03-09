package adapters

import (
	"encoding/json"
	"testing"

	"github.com/maelstrom/v3/pkg/mail"
)

func TestWebSocketAdapter_Bidirectional(t *testing.T) {
	adapter := NewWebSocketAdapter()

	// Test inbound normalization
	inputData := []byte(`{"message": "hello", "type": "chat"}`)
	inboundMail, err := adapter.NormalizeInbound(inputData)
	if err != nil {
		t.Errorf("Expected nil error, got %v", err)
	}

	if inboundMail.Type != mail.MailTypeMailReceived {
		t.Errorf("Expected MailTypeMailReceived, got %s", inboundMail.Type)
	}
	if inboundMail.Source != "gateway:websocket" {
		t.Errorf("Expected source 'gateway:websocket', got '%s'", inboundMail.Source)
	}

	// Test outbound normalization
	outboundMail := mail.Mail{
		ID:      "msg-001",
		Type:    mail.MailTypeAssistant,
		Content: "Response content",
		Source:  "agent:bot",
		Metadata: mail.MailMetadata{
			Stream: nil,
		},
	}

	outboundData, err := adapter.NormalizeOutbound(outboundMail)
	if err != nil {
		t.Errorf("Expected nil error, got %v", err)
	}

	// Verify JSON structure
	var result map[string]any
	if err := json.Unmarshal(outboundData, &result); err != nil {
		t.Errorf("Expected valid JSON, got error: %v", err)
	}

	if result["type"] != "assistant" {
		t.Errorf("Expected type 'assistant', got '%v'", result["type"])
	}
	if result["source"] != "agent:bot" {
		t.Errorf("Expected source 'agent:bot', got '%v'", result["source"])
	}
}
