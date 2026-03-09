package gateway

import (
	"net/http"
	"testing"

	"github.com/maelstrom/v3/pkg/gateway/adapters"
	"github.com/maelstrom/v3/pkg/mail"
)

type mockAdapter struct {
	name string
}

func (m *mockAdapter) Name() string {
	return m.name
}

func (m *mockAdapter) NormalizeInbound(data []byte) (mail.Mail, error) {
	return mail.Mail{}, nil
}

func (m *mockAdapter) NormalizeOutbound(mail mail.Mail) ([]byte, error) {
	return []byte{}, nil
}

func TestGateway_RegisterAdapter(t *testing.T) {
	gateway := NewGateway()

	adapter := &mockAdapter{name: "test-adapter"}

	err := gateway.RegisterAdapter(adapter)
	if err != nil {
		t.Errorf("Expected nil error, got %v", err)
	}

	// Verify adapter was registered
	retrieved, err := gateway.GetAdapter("test-adapter")
	if err != nil {
		t.Errorf("Expected nil error, got %v", err)
	}
	if retrieved != adapter {
		t.Error("Expected same adapter instance")
	}

	// Verify in list
	names := gateway.ListAdapters()
	found := false
	for _, n := range names {
		if n == "test-adapter" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected test-adapter in list")
	}
}

func TestGateway_AdapterNotFound(t *testing.T) {
	gateway := NewGateway()

	// Try to get non-registered adapter
	_, err := gateway.GetAdapter("non-existent")
	if err == nil {
		t.Error("Expected error for non-existent adapter")
	}

	if err.Error() != "adapter not found: non-existent" {
		t.Errorf("Expected 'adapter not found: non-existent', got '%s'", err.Error())
	}

	// Verify empty list
	names := gateway.ListAdapters()
	if len(names) != 0 {
		t.Errorf("Expected empty list, got %d adapters", len(names))
	}
}

func TestAdapter_NormalizationRoundTrip(t *testing.T) {
	gateway := NewGateway()

	// Register all adapters
	adapters := []Adapter{
		adapters.NewWebhookAdapter(),
		adapters.NewSSEAdapter(),
		adapters.NewWebSocketAdapter(),
		adapters.NewPubSubAdapter(),
		adapters.NewSMTPAdapter(),
		adapters.NewSlackAdapter(),
		adapters.NewWhatsAppAdapter(),
		adapters.NewTelegramAdapter(),
	}

	for _, adapter := range adapters {
		err := gateway.RegisterAdapter(adapter)
		if err != nil {
			t.Errorf("Failed to register %s: %v", adapter.Name(), err)
		}
	}

	// Verify all adapters registered
	names := gateway.ListAdapters()
	if len(names) != len(adapters) {
		t.Errorf("Expected %d adapters, got %d", len(adapters), len(names))
	}

	// Test round-trip for each adapter
	for _, adapter := range adapters {
		// Create test mail
		originalMail := mail.Mail{
			ID:      "msg-001",
			Type:    mail.MailTypeUser,
			Source:  "test",
			Content: "test content",
			Metadata: mail.MailMetadata{
				Boundary: mail.OuterBoundary,
			},
		}

		// Normalize outbound
		outbound, err := adapter.NormalizeOutbound(originalMail)
		if err != nil {
			t.Errorf("Adapter %s: NormalizeOutbound failed: %v", adapter.Name(), err)
			continue
		}

		// Normalize inbound (simulating response)
		inbound, err := adapter.NormalizeInbound(outbound)
		if err != nil {
			t.Errorf("Adapter %s: NormalizeInbound failed: %v", adapter.Name(), err)
			continue
		}

		// Verify mail type preserved
		if inbound.Type != mail.MailTypeMailReceived {
			t.Errorf("Adapter %s: Expected MailTypeMailReceived, got %s",
				adapter.Name(), inbound.Type)
		}

		// Verify source set correctly
		expectedSource := "gateway:" + adapter.Name()
		if inbound.Source != expectedSource {
			t.Errorf("Adapter %s: Expected source '%s', got '%s'",
				adapter.Name(), expectedSource, inbound.Source)
		}
	}
}

func TestGatewayService_RegisterHTTPEndpoint(t *testing.T) {
	gw := NewGatewayService()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})

	err := gw.RegisterHTTPEndpoint("/test", handler)
	if err != nil {
		t.Errorf("Expected nil error, got %v", err)
	}
}

func TestGatewayService_HTTPEndpointHandler(t *testing.T) {
	gw := NewGatewayService()

	called := false
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
	})

	err := gw.RegisterHTTPEndpoint("/test", handler)
	if err != nil {
		t.Fatalf("Expected nil error, got %v", err)
	}

	rr := &testResponseWriter{}
	req := &http.Request{}

	gw.ServeHTTP(rr, req)

	if !called {
		t.Error("Expected handler to be called")
	}
}

type testResponseWriter struct {
	code int
}

func (t *testResponseWriter) Header() http.Header {
	return make(http.Header)
}

func (t *testResponseWriter) Write([]byte) (int, error) {
	return 0, nil
}

func (t *testResponseWriter) WriteHeader(code int) {
	t.code = code
}
