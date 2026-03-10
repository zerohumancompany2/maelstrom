package gateway

import (
	"testing"

	mailpkg "github.com/maelstrom/v3/pkg/mail"
)

// TestInternalGRPCAdapter_DirectRouting tests internal gRPC direct routing
// Spec Reference: arch-v1.md L659-666 (Channel Adapters), internal gRPC routing
func TestInternalGRPCAdapter_DirectRouting(t *testing.T) {
	adapter := &InternalGRPCAdapter{}

	// Test inbound normalization for gRPC
	inboundMsg := map[string]any{
		"service": "agent:core",
		"method":  "process",
		"payload": map[string]any{
			"data": "test payload",
		},
	}

	mailMsg, err := adapter.NormalizeInbound(inboundMsg)
	if err != nil {
		t.Fatalf("NormalizeInbound failed: %v", err)
	}

	if mailMsg.Type != mailpkg.MailReceived {
		t.Errorf("Expected type mail_received, got %v", mailMsg.Type)
	}

	if mailMsg.Metadata.Adapter != "grpc" {
		t.Errorf("Expected adapter 'grpc', got %v", mailMsg.Metadata.Adapter)
	}

	// Test outbound normalization
	outboundMail := &mailpkg.Mail{
		Type: mailpkg.ToolResult,
		Content: map[string]any{
			"result": "success",
		},
		Metadata: mailpkg.MailMetadata{
			Adapter: "grpc",
		},
	}

	outboundMsg, err := adapter.NormalizeOutbound(outboundMail)
	if err != nil {
		t.Fatalf("NormalizeOutbound failed: %v", err)
	}

	outboundMap, ok := outboundMsg.(map[string]any)
	if !ok {
		t.Fatalf("Expected outbound to be map[string]any, got %T", outboundMsg)
	}

	if outboundMap["result"] != "success" {
		t.Errorf("Expected result 'success', got %v", outboundMap["result"])
	}

	// Verify streaming is disabled for gRPC (request/response, not streaming)
	if adapter.Stream() {
		t.Errorf("Expected InternalGRPCAdapter to not stream, got true")
	}

	// Verify adapter name
	if adapter.Name() != "grpc" {
		t.Errorf("Expected adapter name 'grpc', got %v", adapter.Name())
	}
}
