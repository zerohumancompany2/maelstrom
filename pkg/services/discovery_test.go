package services

import (
	"testing"
)

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
