package gateway

import (
	"testing"

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
