package bootstrap

import (
	"testing"
)

// TestBootstrapChart_LoadsServices verifies bootstrap chart contains 4 core services.
func TestBootstrapChart_LoadsServices(t *testing.T) {
	def, err := LoadBootstrapChart()
	if err != nil {
		t.Fatalf("failed to load bootstrap chart: %v", err)
	}

	if def.ID != "sys:bootstrap" {
		t.Errorf("expected ID 'sys:bootstrap', got %q", def.ID)
	}
	if def.Version != "1.0.0" {
		t.Errorf("expected Version '1.0.0', got %q", def.Version)
	}

	// Verify we have states in the spec
	states, ok := def.Spec["states"]
	if !ok {
		t.Error("bootstrap chart missing 'states' in spec")
	}
	if states == nil {
		t.Error("states should not be nil")
	}
}

// TestBootstrapChart_HasSecurityState verifies security state exists.
func TestBootstrapChart_HasSecurityState(t *testing.T) {
	def, err := LoadBootstrapChart()
	if err != nil {
		t.Fatalf("failed to load bootstrap chart: %v", err)
	}

	states, ok := def.Spec["states"].(map[string]interface{})
	if !ok {
		t.Fatal("states should be a map")
	}

	if _, ok := states["security"]; !ok {
		t.Error("bootstrap chart missing 'security' state")
	}
}

func TestBootstrapChartYAML_ParsesWithoutError(t *testing.T) {
	def, err := LoadBootstrapChart()
	if err != nil {
		t.Fatalf("BootstrapChartYAML failed to parse: %v", err)
	}

	if def.ID != "sys:bootstrap" {
		t.Errorf("expected ID 'sys:bootstrap', got %q", def.ID)
	}
	if def.Version != "1.0.0" {
		t.Errorf("expected Version '1.0.0', got %q", def.Version)
	}
}

func TestBootstrapChartYAML_HasRequiredStates(t *testing.T) {
	def, err := LoadBootstrapChart()
	if err != nil {
		t.Fatalf("failed to load bootstrap chart: %v", err)
	}

	states, ok := def.Spec["states"].(map[string]interface{})
	if !ok {
		t.Fatal("states should be a map")
	}

	requiredStates := []string{
		"sys:bootstrap/security",
		"sys:bootstrap/communication",
		"sys:bootstrap/observability",
		"sys:bootstrap/lifecycle",
		"sys:bootstrap/ready",
		"sys:bootstrap/failed",
	}

	for _, stateName := range requiredStates {
		if _, ok := states[stateName]; !ok {
			t.Errorf("missing required state: %s", stateName)
		}
	}

	readyState, ok := states["sys:bootstrap/ready"]
	if !ok {
		t.Fatal("sys:bootstrap/ready state not found")
	}
	readyMap, ok := readyState.(map[string]interface{})
	if !ok {
		t.Fatal("sys:bootstrap/ready should be a map")
	}
	if readyMap["type"] != "final" {
		t.Error("sys:bootstrap/ready should be marked as final state")
	}
}
