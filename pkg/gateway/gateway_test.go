package gateway

import (
	"github.com/maelstrom/v3/pkg/mail"
	"testing"
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
