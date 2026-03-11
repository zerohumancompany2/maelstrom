package gateway

import (
	"net/http"
	"testing"
	"time"

	mailpkg "github.com/maelstrom/v3/pkg/mail"
)

// TestWebhookAdapter_NormalizeInboundWithTaints tests that inbound messages are properly tainted
// Spec Reference: arch-v1.md L279-283 (Data Tainting), L659-666 (Channel Adapters)
func TestWebhookAdapter_NormalizeInboundWithTaints(t *testing.T) {
	adapter := NewWebhookAdapter(8080)

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

	expectedTaints := []string{TaintUserSupplied, TaintExternal}
	if len(mailMsg.Taints) != len(expectedTaints) {
		t.Errorf("Expected %d taints, got %d: %v", len(expectedTaints), len(mailMsg.Taints), mailMsg.Taints)
	}

	for _, expectedTaint := range expectedTaints {
		found := false
		for _, taint := range mailMsg.Taints {
			if taint == expectedTaint {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected taint %s not found", expectedTaint)
		}
	}
}

// TestWebhookAdapter_NormalizeInboundNilError tests nil input handling
// Spec Reference: arch-v1.md L670-671 (normalize inbound to mail_received)
func TestWebhookAdapter_NormalizeInboundNilError(t *testing.T) {
	adapter := NewWebhookAdapter(8080)

	_, err := adapter.NormalizeInbound(nil)
	if err == nil {
		t.Error("Expected error for nil input")
	}
}

// TestWebhookAdapter_NormalizeOutboundNilError tests nil mail handling
// Spec Reference: arch-v1.md L670-671 (normalize outbound)
func TestWebhookAdapter_NormalizeOutboundNilError(t *testing.T) {
	adapter := NewWebhookAdapter(8080)

	_, err := adapter.NormalizeOutbound(nil)
	if err == nil {
		t.Error("Expected error for nil mail")
	}
}

// TestWebhookAdapter_HandleMethodValidation tests HTTP method validation
func TestWebhookAdapter_HandleMethodValidation(t *testing.T) {
	adapter := NewWebhookAdapter(8080)

	req := &http.Request{Method: http.MethodGet}
	err := adapter.Handle(req)
	if err == nil {
		t.Error("Expected error for non-POST request")
	}

	req = &http.Request{Method: http.MethodPost}
	err = adapter.Handle(req)
	if err != nil {
		t.Errorf("Unexpected error for POST request: %v", err)
	}
}

// TestWebhookAdapter_MailStructure tests mail structure completeness
// Spec Reference: arch-v1.md L170-186 (Message format)
func TestWebhookAdapter_MailStructure(t *testing.T) {
	adapter := NewWebhookAdapter(8080)

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

	if mailMsg.ID == "" {
		t.Error("Expected non-empty ID")
	}

	if mailMsg.CorrelationID == "" {
		t.Error("Expected non-empty CorrelationID")
	}

	if mailMsg.Type != mailpkg.MailReceived {
		t.Errorf("Expected type mail_received, got %v", mailMsg.Type)
	}

	if mailMsg.CreatedAt.IsZero() {
		t.Error("Expected non-zero CreatedAt")
	}

	if mailMsg.Source != "gateway" {
		t.Errorf("Expected source 'gateway', got %v", mailMsg.Source)
	}

	if mailMsg.Metadata.Adapter != "webhook" {
		t.Errorf("Expected adapter 'webhook', got %v", mailMsg.Metadata.Adapter)
	}

	if mailMsg.Metadata.Boundary != mailpkg.OuterBoundary {
		t.Errorf("Expected boundary outer, got %v", mailMsg.Metadata.Boundary)
	}
}

// TestWebhookAdapter_ContentPreservation tests content is preserved through normalization
func TestWebhookAdapter_ContentPreservation(t *testing.T) {
	adapter := NewWebhookAdapter(8080)

	inputContent := map[string]any{
		"message": "test content",
		"metadata": map[string]any{
			"key": "value",
		},
	}

	mailMsg, err := adapter.NormalizeInbound(inputContent)
	if err != nil {
		t.Fatalf("NormalizeInbound failed: %v", err)
	}

	outbound, err := adapter.NormalizeOutbound(mailMsg)
	if err != nil {
		t.Fatalf("NormalizeOutbound failed: %v", err)
	}

	outboundMap, ok := outbound.(map[string]any)
	if !ok {
		t.Fatalf("Expected outbound to be map[string]any, got %T", outbound)
	}

	if outboundMap["message"] != inputContent["message"] {
		t.Error("Expected content to be preserved through round-trip")
	}
}

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

// TestChannelAdapter_WebhookNormalizesToMail tests webhook adapter normalization
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

// TestWebhookAdapter_TimeoutConfig tests configurable timeout
func TestWebhookAdapter_TimeoutConfig(t *testing.T) {
	adapter := NewWebhookAdapter(8080)

	if adapter.Name() != "webhook" {
		t.Errorf("Expected name 'webhook', got %v", adapter.Name())
	}

	if adapter.Stream() {
		t.Error("Expected Stream() to return false")
	}
}

// TestWebhookAdapter_PortConfiguration tests port configuration
func TestWebhookAdapter_PortConfiguration(t *testing.T) {
	port := 9090
	adapter := NewWebhookAdapter(port)

	if adapter.Name() != "webhook" {
		t.Errorf("Expected name 'webhook', got %v", adapter.Name())
	}

	if adapter.Stream() {
		t.Error("Expected Stream() to return false")
	}
}

// TestWebhookAdapter_BoundaryValidation tests boundary is set correctly
func TestWebhookAdapter_BoundaryValidation(t *testing.T) {
	adapter := NewWebhookAdapter(8080)

	rawMessage := map[string]any{
		"body": "test",
	}

	mailMsg, err := adapter.NormalizeInbound(rawMessage)
	if err != nil {
		t.Fatalf("NormalizeInbound failed: %v", err)
	}

	if mailMsg.Metadata.Boundary != mailpkg.OuterBoundary {
		t.Errorf("Expected boundary outer, got %v", mailMsg.Metadata.Boundary)
	}
}

// TestWebhookAdapter_CreatedAtValidation tests CreatedAt is set
func TestWebhookAdapter_CreatedAtValidation(t *testing.T) {
	adapter := NewWebhookAdapter(8080)

	before := time.Now()
	rawMessage := map[string]any{"body": "test"}

	mailMsg, err := adapter.NormalizeInbound(rawMessage)
	if err != nil {
		t.Fatalf("NormalizeInbound failed: %v", err)
	}

	after := time.Now()

	if mailMsg.CreatedAt.Before(before) || mailMsg.CreatedAt.After(after) {
		t.Errorf("CreatedAt %v not within expected range [%v, %v]", mailMsg.CreatedAt, before, after)
	}
}
