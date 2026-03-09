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

func TestServiceRegistry_RegisterDuplicate(t *testing.T) {
	sr := NewServiceRegistry()
	svc1 := &mockService{id: "test:service"}
	svc2 := &mockService{id: "test:service2"}

	err := sr.Register("test:service", svc1)
	if err != nil {
		t.Fatalf("First Register() returned error: %v", err)
	}

	err = sr.Register("test:service", svc2)
	if err != ErrAlreadyRegistered {
		t.Fatalf("Register() returned wrong error: got %v, want %v", err, ErrAlreadyRegistered)
	}

	retrieved, ok := sr.Get("test:service")
	if !ok {
		t.Fatal("Get() returned false for original service")
	}
	if retrieved.(*mockService) != svc1 {
		t.Fatal("Get() returned wrong service - original was overwritten")
	}
}
