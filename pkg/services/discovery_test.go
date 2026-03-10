package services

import (
	"testing"
)

func TestServiceDiscovery_registryFindsByCapability(t *testing.T) {
	sr := NewServiceRegistry()

	// Register services with capabilities
	sr.RegisterWithCapabilities("sys:security", &mockService{id: "sys:security"}, []string{"auth", "authz"})
	sr.RegisterWithCapabilities("sys:communication", &mockService{id: "sys:communication"}, []string{"messaging"})
	sr.RegisterWithCapabilities("sys:observability", &mockService{id: "sys:observability"}, []string{"logging", "metrics"})
	sr.RegisterWithCapabilities("sys:lifecycle", &mockService{id: "sys:lifecycle"}, []string{"spawn", "control"})

	// Find services by capability
	authServices := sr.FindByCapability("auth")
	if len(authServices) != 1 {
		t.Fatalf("FindByCapability('auth') returned %d services, want 1", len(authServices))
	}
	if authServices[0].ID() != "sys:security" {
		t.Fatalf("FindByCapability('auth') returned wrong service: %s", authServices[0].ID())
	}

	// Find services with shared capability
	loggingServices := sr.FindByCapability("logging")
	if len(loggingServices) != 1 {
		t.Fatalf("FindByCapability('logging') returned %d services, want 1", len(loggingServices))
	}
	if loggingServices[0].ID() != "sys:observability" {
		t.Fatalf("FindByCapability('logging') returned wrong service: %s", loggingServices[0].ID())
	}

	// Find non-existent capability
	nonExistent := sr.FindByCapability("nonexistent")
	if len(nonExistent) != 0 {
		t.Fatalf("FindByCapability('nonexistent') returned %d services, want 0", len(nonExistent))
	}
}

func TestServiceDiscovery_registryListsAllServices(t *testing.T) {
	sr := NewServiceRegistry()

	// Register some test services
	sr.Register("sys:security", &mockService{id: "sys:security"})
	sr.Register("sys:communication", &mockService{id: "sys:communication"})
	sr.Register("sys:observability", &mockService{id: "sys:observability"})
	sr.Register("sys:lifecycle", &mockService{id: "sys:lifecycle"})

	// Discover all services
	services := sr.DiscoverServices()

	if len(services) != 4 {
		t.Fatalf("DiscoverServices() returned %d services, want 4", len(services))
	}

	// Verify all services are in the list
	expected := map[string]bool{
		"sys:security":      true,
		"sys:communication": true,
		"sys:observability": true,
		"sys:lifecycle":     true,
	}

	for _, svc := range services {
		if !expected[svc.ID()] {
			t.Fatalf("DiscoverServices() returned unexpected service: %s", svc.ID())
		}
	}
}
