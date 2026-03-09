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

func TestBootstrapChartYAML_HasSuccessTransitions(t *testing.T) {
	def, err := LoadBootstrapChart()
	if err != nil {
		t.Fatalf("failed to load bootstrap chart: %v", err)
	}

	states, ok := def.Spec["states"].(map[string]interface{})
	if !ok {
		t.Fatal("states should be a map")
	}

	checkTransition := func(fromState, event, toState string) {
		t.Helper()
		state, ok := states[fromState]
		if !ok {
			t.Errorf("state %s not found", fromState)
			return
		}
		stateMap, ok := state.(map[string]interface{})
		if !ok {
			t.Errorf("state %s should be a map", fromState)
			return
		}
		transitions, ok := stateMap["transitions"].([]interface{})
		if !ok {
			t.Errorf("state %s should have transitions", fromState)
			return
		}
		found := false
		for _, trans := range transitions {
			transMap, ok := trans.(map[string]interface{})
			if !ok {
				continue
			}
			if transMap["event"] == event && transMap["target"] == toState {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("missing transition: %s --(%s)--> %s", fromState, event, toState)
		}
	}

	checkTransition("sys:bootstrap/security", "SECURITY_READY", "sys:bootstrap/communication")
	checkTransition("sys:bootstrap/communication", "COMMUNICATION_READY", "sys:bootstrap/observability")
	checkTransition("sys:bootstrap/observability", "OBSERVABILITY_READY", "sys:bootstrap/lifecycle")
	checkTransition("sys:bootstrap/lifecycle", "LIFECYCLE_READY", "sys:bootstrap/ready")
}

func TestBootstrapChartYAML_HasErrorTransitions(t *testing.T) {
	def, err := LoadBootstrapChart()
	if err != nil {
		t.Fatalf("failed to load bootstrap chart: %v", err)
	}

	states, ok := def.Spec["states"].(map[string]interface{})
	if !ok {
		t.Fatal("states should be a map")
	}

	checkErrorTransition := func(fromState, event string) {
		t.Helper()
		state, ok := states[fromState]
		if !ok {
			t.Errorf("state %s not found", fromState)
			return
		}
		stateMap, ok := state.(map[string]interface{})
		if !ok {
			t.Errorf("state %s should be a map", fromState)
			return
		}
		transitions, ok := stateMap["transitions"].([]interface{})
		if !ok {
			t.Errorf("state %s should have transitions", fromState)
			return
		}
		found := false
		for _, trans := range transitions {
			transMap, ok := trans.(map[string]interface{})
			if !ok {
				continue
			}
			if transMap["event"] == event && transMap["target"] == "sys:bootstrap/failed" {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("missing error transition: %s --(%s)--> sys:bootstrap/failed", fromState, event)
		}
	}

	checkErrorTransition("sys:bootstrap/security", "securityFailed")
	checkErrorTransition("sys:bootstrap/communication", "communicationFailed")
	checkErrorTransition("sys:bootstrap/observability", "observabilityFailed")
	checkErrorTransition("sys:bootstrap/lifecycle", "lifecycleFailed")
}
