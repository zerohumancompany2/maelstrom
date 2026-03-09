package adapters

import (
	"net/http"
	"testing"
	"time"

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

func TestWebhookAdapter_HTTPServer(t *testing.T) {
	adapter := NewWebhookAdapter()

	err := adapter.StartServer("127.0.0.1:18080")
	if err != nil {
		t.Fatalf("Expected nil error, got %v", err)
	}
	defer adapter.StopServer()

	time.Sleep(100 * time.Millisecond)

	resp, err := http.Get("http://127.0.0.1:18080/webhook")
	if err != nil {
		t.Fatalf("Expected successful request, got %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}
}
