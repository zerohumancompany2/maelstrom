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

func TestServiceDiscovery_registryReportsHealthStatus(t *testing.T) {
	sr := NewServiceRegistry()

	// Register services with capabilities
	sr.RegisterWithCapabilities("sys:security", &mockService{id: "sys:security"}, []string{"auth"})
	sr.RegisterWithCapabilities("sys:communication", &mockService{id: "sys:communication"}, []string{"messaging"})

	// Initially all services should be unknown health
	health := sr.GetHealthStatus("sys:security")
	if health != "unknown" {
		t.Fatalf("GetHealthStatus() returned %q for new service, want %q", health, "unknown")
	}

	// Update health to healthy
	sr.UpdateHealthStatus("sys:security", "healthy")
	health = sr.GetHealthStatus("sys:security")
	if health != "healthy" {
		t.Fatalf("GetHealthStatus() returned %q, want %q", health, "healthy")
	}

	// Update health to unhealthy
	sr.UpdateHealthStatus("sys:security", "unhealthy")
	health = sr.GetHealthStatus("sys:security")
	if health != "unhealthy" {
		t.Fatalf("GetHealthStatus() returned %q, want %q", health, "unhealthy")
	}

	// Get health of non-existent service
	health = sr.GetHealthStatus("sys:nonexistent")
	if health != "unknown" {
		t.Fatalf("GetHealthStatus() returned %q for non-existent service, want %q", health, "unknown")
	}

	// Get all unhealthy services
	sr.UpdateHealthStatus("sys:communication", "unhealthy")
	unhealthyServices := sr.GetUnhealthyServices()
	if len(unhealthyServices) != 2 {
		t.Fatalf("GetUnhealthyServices() returned %d services, want 2", len(unhealthyServices))
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

func TestServiceDiscovery_registryReportsServiceDependencies(t *testing.T) {
	sr := NewServiceRegistry()

	// Register services with dependencies
	sr.RegisterWithDependencies("sys:security", &mockService{id: "sys:security"}, []string{})
	sr.RegisterWithDependencies("sys:communication", &mockService{id: "sys:communication"}, []string{"sys:security"})
	sr.RegisterWithDependencies("sys:observability", &mockService{id: "sys:observability"}, []string{"sys:security", "sys:communication"})

	// Get dependencies for a service
	deps := sr.GetDependencies("sys:communication")
	if len(deps) != 1 {
		t.Fatalf("GetDependencies('sys:communication') returned %d deps, want 1", len(deps))
	}
	if deps[0] != "sys:security" {
		t.Fatalf("GetDependencies('sys:communication') returned %q, want %q", deps[0], "sys:security")
	}

	// Get dependencies for service with multiple deps
	deps = sr.GetDependencies("sys:observability")
	if len(deps) != 2 {
		t.Fatalf("GetDependencies('sys:observability') returned %d deps, want 2", len(deps))
	}

	// Get dependents (services that depend on this service)
	dependents := sr.GetDependents("sys:security")
	if len(dependents) != 2 {
		t.Fatalf("GetDependents('sys:security') returned %d dependents, want 2", len(dependents))
	}

	// Get dependents for leaf service
	dependents = sr.GetDependents("sys:observability")
	if len(dependents) != 0 {
		t.Fatalf("GetDependents('sys:observability') returned %d dependents, want 0", len(dependents))
	}

	// Get dependencies for non-existent service
	deps = sr.GetDependencies("sys:nonexistent")
	if len(deps) != 0 {
		t.Fatalf("GetDependencies('sys:nonexistent') returned %d deps, want 0", len(deps))
	}
}
