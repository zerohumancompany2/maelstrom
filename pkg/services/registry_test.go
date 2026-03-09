package services

import (
	"testing"

	"github.com/maelstrom/v3/pkg/mail"
)

func TestServiceRegistry_Register(t *testing.T) {
	sr := NewServiceRegistry()
	svc := &mockService{id: "test:service"}

	err := sr.Register("test:service", svc)
	if err != nil {
		t.Fatalf("Register() returned error: %v", err)
	}

	retrieved, ok := sr.Get("test:service")
	if !ok {
		t.Fatal("Get() returned false for registered service")
	}
	if retrieved.(*mockService) != svc {
		t.Fatal("Get() returned wrong service")
	}
}

type mockService struct {
	id string
}

func (m *mockService) ID() string {
	return m.id
}

func (m *mockService) HandleMail(mail mail.Mail) error {
	return nil
}

func (m *mockService) Start() error {
	return nil
}

func (m *mockService) Stop() error {
	return nil
}
