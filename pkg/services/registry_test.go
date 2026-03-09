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

func TestServiceRegistry_Get(t *testing.T) {
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
	if retrieved != svc {
		t.Fatal("Get() returned wrong service instance")
	}
}

func TestServiceRegistry_GetNotFound(t *testing.T) {
	sr := NewServiceRegistry()

	retrieved, ok := sr.Get("nonexistent")
	if ok {
		t.Fatal("Get() returned true for non-existent service")
	}
	if retrieved != nil {
		t.Fatal("Get() returned non-nil service for non-existent service")
	}
}

func TestServiceRegistry_List(t *testing.T) {
	sr := NewServiceRegistry()

	sr.Register("sys:communication", &mockService{id: "sys:communication"})
	sr.Register("sys:security", &mockService{id: "sys:security"})
	sr.Register("sys:lifecycle", &mockService{id: "sys:lifecycle"})

	names := sr.List()
	if len(names) != 3 {
		t.Fatalf("List() returned %d names, want 3", len(names))
	}

	expected := []string{"sys:communication", "sys:lifecycle", "sys:security"}
	for i, name := range expected {
		if names[i] != name {
			t.Fatalf("List()[%d] = %q, want %q", i, names[i], name)
		}
	}
}

func TestServiceRegistry_ListEmpty(t *testing.T) {
	sr := NewServiceRegistry()

	names := sr.List()
	if names == nil {
		t.Fatal("List() returned nil for empty registry, want empty slice")
	}
	if len(names) != 0 {
		t.Fatalf("List() returned %d names, want 0", len(names))
	}
}

func TestServiceRegistry_RegisterWithState(t *testing.T) {
	sr := NewServiceRegistry()
	svc := &mockService{id: "test:service"}

	err := sr.RegisterWithState(svc, "registered")
	if err != nil {
		t.Fatalf("RegisterWithState() returned error: %v", err)
	}

	retrieved, ok := sr.Get(svc.ID())
	if !ok {
		t.Fatal("Get() returned false for registered service")
	}
	if retrieved.(*mockService) != svc {
		t.Fatal("Get() returned wrong service")
	}

	state, ok := sr.GetState(svc.ID())
	if !ok {
		t.Fatal("GetState() returned false for registered service")
	}
	if state != "registered" {
		t.Fatalf("GetState() returned %q, want %q", state, "registered")
	}
}

func TestAllServicesIntegrateViaRegistry(t *testing.T) {
	sr := NewServiceRegistry()

	securitySvc := &mockService{id: "sys:security"}
	commSvc := &mockService{id: "sys:communication"}
	lifecycleSvc := &mockService{id: "sys:lifecycle"}
	observabilitySvc := &mockService{id: "sys:observability"}

	err := sr.Register("sys:security", securitySvc)
	if err != nil {
		t.Fatalf("Register security: %v", err)
	}
	err = sr.Register("sys:communication", commSvc)
	if err != nil {
		t.Fatalf("Register communication: %v", err)
	}
	err = sr.Register("sys:lifecycle", lifecycleSvc)
	if err != nil {
		t.Fatalf("Register lifecycle: %v", err)
	}
	err = sr.Register("sys:observability", observabilitySvc)
	if err != nil {
		t.Fatalf("Register observability: %v", err)
	}

	securityRetrieved, ok := sr.Get("sys:security")
	if !ok || securityRetrieved != securitySvc {
		t.Fatal("Get sys:security failed")
	}
	commRetrieved, ok := sr.Get("sys:communication")
	if !ok || commRetrieved != commSvc {
		t.Fatal("Get sys:communication failed")
	}
	lifecycleRetrieved, ok := sr.Get("sys:lifecycle")
	if !ok || lifecycleRetrieved != lifecycleSvc {
		t.Fatal("Get sys:lifecycle failed")
	}
	observabilityRetrieved, ok := sr.Get("sys:observability")
	if !ok || observabilityRetrieved != observabilitySvc {
		t.Fatal("Get sys:observability failed")
	}

	names := sr.List()
	expected := []string{"sys:communication", "sys:lifecycle", "sys:observability", "sys:security"}
	if len(names) != len(expected) {
		t.Fatalf("List() returned %d names, want %d", len(names), len(expected))
	}
	for i, name := range expected {
		if names[i] != name {
			t.Fatalf("List()[%d] = %q, want %q", i, names[i], name)
		}
	}
}
