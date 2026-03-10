package gateway

import (
	"testing"
)

// TestGatewayService_ID - Spec: arch-v1.md L466, L477-480
func TestGatewayService_ID(t *testing.T) {
	svc := NewGatewayService()
	id := svc.ID()
	if id != "sys:gateway" {
		t.Errorf("Expected ID 'sys:gateway', got '%s'", id)
	}
}

// TestGatewayService_RegisterAdapter_DuplicateReturnsError - Spec: arch-v1.md L659-666
func TestGatewayService_RegisterAdapter_DuplicateReturnsError(t *testing.T) {
	svc := NewGatewayService()
	adapter := &WebhookAdapter{}

	// First registration should succeed
	if err := svc.RegisterAdapter("webhook", adapter); err != nil {
		t.Fatalf("First registration should succeed: %v", err)
	}

	// Duplicate registration should return error
	if err := svc.RegisterAdapter("webhook", adapter); err == nil {
		t.Error("Expected error on duplicate registration, got nil")
	}
}

func TestGateway_RegisterAdapter(t *testing.T) {
	// Test: Register webhook, websocket, sse, smtp adapters
	svc := NewGatewayService()

	// Register webhook adapter
	webhook := &WebhookAdapter{}
	if err := svc.RegisterAdapter("webhook", webhook); err != nil {
		t.Fatalf("Failed to register webhook adapter: %v", err)
	}

	// Register websocket adapter
	ws := &WebSocketAdapter{}
	if err := svc.RegisterAdapter("websocket", ws); err != nil {
		t.Fatalf("Failed to register websocket adapter: %v", err)
	}

	// Register sse adapter
	sse := &SSEAdapter{}
	if err := svc.RegisterAdapter("sse", sse); err != nil {
		t.Fatalf("Failed to register sse adapter: %v", err)
	}

	// Register smtp adapter
	smtp := &SMTPAdapter{}
	if err := svc.RegisterAdapter("smtp", smtp); err != nil {
		t.Fatalf("Failed to register smtp adapter: %v", err)
	}

	// Verify all adapters are registered
	adapt, ok := svc.GetAdapter("webhook")
	if !ok {
		t.Fatal("webhook adapter not registered")
	}
	if adapt.Name() != "webhook" {
		t.Errorf("Expected webhook adapter name 'webhook', got '%s'", adapt.Name())
	}

	adapt, ok = svc.GetAdapter("websocket")
	if !ok {
		t.Fatal("websocket adapter not registered")
	}
	if adapt.Name() != "websocket" {
		t.Errorf("Expected websocket adapter name 'websocket', got '%s'", adapt.Name())
	}

	adapt, ok = svc.GetAdapter("sse")
	if !ok {
		t.Fatal("sse adapter not registered")
	}
	if adapt.Name() != "sse" {
		t.Errorf("Expected sse adapter name 'sse', got '%s'", adapt.Name())
	}

	adapt, ok = svc.GetAdapter("smtp")
	if !ok {
		t.Fatal("smtp adapter not registered")
	}
	if adapt.Name() != "smtp" {
		t.Errorf("Expected smtp adapter name 'smtp', got '%s'", adapt.Name())
	}
}
