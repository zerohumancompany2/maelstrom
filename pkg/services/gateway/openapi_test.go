package gateway

import (
	"testing"
)

func TestGateway_OpenAPI(t *testing.T) {
	// Test: Auto-generate OpenAPI from chart events
	svc := NewGatewayService()

	// Register adapters
	svc.RegisterAdapter("webhook", &WebhookAdapter{})
	svc.RegisterAdapter("websocket", &WebSocketAdapter{})
	svc.RegisterAdapter("sse", &SSEAdapter{})
	svc.RegisterAdapter("smtp", &SMTPAdapter{})

	// Get OpenAPI spec
	spec, err := svc.GetOpenAPI()
	if err != nil {
		t.Fatalf("Failed to get OpenAPI spec: %v", err)
	}

	// Verify spec is generated
	if spec == nil {
		t.Fatal("OpenAPI spec is nil")
	}

	// Verify version is set
	if spec.Version == "" {
		t.Error("OpenAPI spec version is empty")
	}
}
