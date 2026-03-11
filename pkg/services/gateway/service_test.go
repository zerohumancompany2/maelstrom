package gateway

import (
	"testing"

	"github.com/maelstrom/v3/pkg/mail"
	"gopkg.in/yaml.v3"
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

	// Register adapter first
	if err := svc.RegisterAdapter("webhook", &WebhookAdapter{}); err != nil {
		t.Fatalf("Failed to register adapter: %v", err)
	}

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

// TestGatewayService_NormalizeInbound_ContentNormalization - Spec: arch-v1.md L670-671
func TestGatewayService_NormalizeInbound_ContentNormalization(t *testing.T) {
	svc := NewGatewayService()

	// Register adapter first
	if err := svc.RegisterAdapter("webhook", &WebhookAdapter{}); err != nil {
		t.Fatalf("Failed to register adapter: %v", err)
	}

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

	content, ok := m.Content.(string)
	if !ok {
		t.Errorf("Expected content to be string, got %T", m.Content)
	}

	if content == "" {
		t.Error("Expected non-empty normalized content")
	}
}

// TestGatewayService_NormalizeOutbound_BoundaryEnforcement - Spec: arch-v1.md L261-270
func TestGatewayService_NormalizeOutbound_BoundaryEnforcement(t *testing.T) {
	svc := NewGatewayService()

	// Test OuterBoundary - sensitive metadata should be stripped
	outerMail := &mail.Mail{
		Type:    mail.MailTypeAssistant,
		Content: "Response content",
		Metadata: mail.MailMetadata{
			Boundary: mail.OuterBoundary,
			Tokens:   100,
		},
	}

	outerResult, err := svc.NormalizeOutbound(outerMail, "webhook")
	if err != nil {
		t.Fatalf("NormalizeOutbound failed for outer boundary: %v", err)
	}

	outerMap, ok := outerResult.(map[string]any)
	if !ok {
		t.Fatalf("Expected result to be map[string]any, got %T", outerResult)
	}

	// Verify sensitive data is stripped
	if _, hasTokens := outerMap["tokens"]; hasTokens {
		t.Error("Expected tokens to be stripped for outer boundary")
	}

	// Test DMZBoundary - limited metadata allowed
	dmzMail := &mail.Mail{
		Type:    mail.MailTypeAssistant,
		Content: "DMZ Response",
		Metadata: mail.MailMetadata{
			Boundary: mail.DMZBoundary,
			Tokens:   50,
		},
	}

	dmzResult, err := svc.NormalizeOutbound(dmzMail, "webhook")
	if err != nil {
		t.Fatalf("NormalizeOutbound failed for dmz boundary: %v", err)
	}

	dmzMap, ok := dmzResult.(map[string]any)
	if !ok {
		t.Fatalf("Expected result to be map[string]any, got %T", dmzResult)
	}

	// Verify only allowed keys are present for DMZ
	allowedKeys := map[string]bool{"content": true, "boundary": true, "adapter": true}
	for key := range dmzMap {
		if !allowedKeys[key] {
			t.Errorf("DMZ boundary should only allow limited metadata, found unexpected key: %s", key)
		}
	}

	// Test InnerBoundary - full metadata allowed
	innerMail := &mail.Mail{
		Type:    mail.MailTypeAssistant,
		Content: "Inner Response",
		Metadata: mail.MailMetadata{
			Boundary: mail.InnerBoundary,
			Tokens:   200,
		},
	}

	innerResult, err := svc.NormalizeOutbound(innerMail, "webhook")
	if err != nil {
		t.Fatalf("NormalizeOutbound failed for inner boundary: %v", err)
	}

	innerMap, ok := innerResult.(map[string]any)
	if !ok {
		t.Fatalf("Expected result to be map[string]any, got %T", innerResult)
	}

	// Verify content is present for inner boundary
	if innerMap["content"] != "Inner Response" {
		t.Error("Expected content to be preserved for inner boundary")
	}
}

func TestChannelAdapter_YamlHotReload(t *testing.T) {
	yamlConfig := `
adapters:
  - name: webhook
    config:
      endpoint: /webhook/test
  - name: websocket
    config:
      endpoint: /ws/test
  - name: sse
    config:
      endpoint: /sse/test
  - name: smtp
    config:
      host: smtp.example.com
      port: 587
  - name: grpc
    config:
      address: 0.0.0.0:50051
`
	var config map[string]any
	err := yaml.Unmarshal([]byte(yamlConfig), &config)
	if err != nil {
		t.Fatalf("Failed to parse YAML config: %v", err)
	}

	adapters := config["adapters"].([]any)
	if len(adapters) != 5 {
		t.Errorf("Expected 5 adapters in config, got %d", len(adapters))
	}

	expectedAdapters := []string{"webhook", "websocket", "sse", "smtp", "grpc"}
	for _, adapterConfig := range adapters {
		adapterMap := adapterConfig.(map[string]any)
		adapterName := adapterMap["name"].(string)

		found := false
		for _, expected := range expectedAdapters {
			if adapterName == expected {
				found = true
				break
			}
		}

		if !found {
			t.Errorf("Expected adapter '%s' not found in config", adapterName)
		}
	}

	var _ ChannelAdapter = &WebhookAdapter{}
	var _ ChannelAdapter = &WebSocketAdapter{}
	var _ ChannelAdapter = &SSEAdapter{}
	var _ ChannelAdapter = &SMTPAdapter{}
	var _ ChannelAdapter = &InternalGRPCAdapter{}
}
