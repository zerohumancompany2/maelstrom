package gateway

import (
	"testing"
)

func TestGateway_AdapterNotFound(t *testing.T) {
	// Test: Unknown adapter returns error
	svc := NewGatewayService()

	// Register a valid adapter first
	validAdapter := &WebhookAdapter{}
	err := svc.RegisterAdapter("valid", validAdapter)
	if err != nil {
		t.Fatalf("Failed to register valid adapter: %v", err)
	}

	// Verify the adapter is registered
	_, ok := svc.GetAdapter("valid")
	if !ok {
		t.Fatal("Valid adapter should be registered")
	}

	// Try to get a non-existent adapter
	_, ok = svc.GetAdapter("nonexistent")
	if ok {
		t.Error("Non-existent adapter should not be found")
	}
}
