package gateway

import (
	"testing"

	"github.com/maelstrom/v3/pkg/mail"
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

// TestGatewayService_NormalizeInbound - Spec: arch-v1.md L670-671
func TestGatewayService_NormalizeInbound(t *testing.T) {
	svc := NewGatewayService()
	rawMessage := map[string]any{
		"from":    "user@example.com",
		"subject": "Test Message",
		"body":    "Hello, World!",
	}

	m, err := svc.NormalizeInbound("webhook", rawMessage)
	if err != nil {
		t.Fatalf("NormalizeInbound failed: %v", err)
	}

	if m.Type != mail.MailReceived {
		t.Errorf("Expected mail type 'mail_received', got '%s'", m.Type)
	}

	if m.Metadata.Adapter != "webhook" {
		t.Errorf("Expected adapter 'webhook', got '%s'", m.Metadata.Adapter)
	}
}

// TestGatewayService_NormalizeOutbound - Spec: arch-v1.md L671, L261-270
func TestGatewayService_NormalizeOutbound(t *testing.T) {
	svc := NewGatewayService()
	outboundMail := &mail.Mail{
		Type:    mail.MailTypeAssistant,
		Content: "Response from assistant",
		Metadata: mail.MailMetadata{
			Boundary: mail.InnerBoundary,
		},
	}

	result, err := svc.NormalizeOutbound(outboundMail, "webhook")
	if err != nil {
		t.Fatalf("NormalizeOutbound failed: %v", err)
	}

	resultMap, ok := result.(map[string]any)
	if !ok {
		t.Fatalf("Expected result to be map[string]any, got %T", result)
	}

	if resultMap["content"] != outboundMail.Content {
		t.Error("Expected content to be preserved in result")
	}

	if resultMap["boundary"] != string(mail.InnerBoundary) {
		t.Errorf("Expected boundary 'inner', got '%s'", resultMap["boundary"])
	}
}

func TestGatewayService_RegisterAdapter_Success(t *testing.T) {
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

// TestGatewayService_NormalizeInbound_UnregisteredAdapterReturnsError - Spec: arch-v1.md L670-671
func TestGatewayService_NormalizeInbound_UnregisteredAdapterReturnsError(t *testing.T) {
	svc := NewGatewayService()
	rawMessage := map[string]any{
		"from":    "user@example.com",
		"subject": "Test Message",
		"body":    "Hello, World!",
	}

	m, err := svc.NormalizeInbound("nonexistent", rawMessage)
	if err == nil {
		t.Error("Expected error for unregistered adapter, got nil")
	}
	if m != nil {
		t.Error("Expected nil mail for unregistered adapter, got non-nil")
	}
}
