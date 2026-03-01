package statechart

import (
	"testing"
)

func TestDefaultHydrator_ParsesYAML(t *testing.T) {
	hydrator := DefaultHydrator()

	yamlContent := []byte(`
id: test-chart
version: 1.0.0
spec:
  initial: idle
  states:
    idle:
      type: atomic
`)

	def, err := hydrator(yamlContent)
	if err != nil {
		t.Fatalf("hydrator failed: %v", err)
	}

	if def.ID != "test-chart" {
		t.Errorf("ID = %q, want %q", def.ID, "test-chart")
	}

	if def.Version != "1.0.0" {
		t.Errorf("Version = %q, want %q", def.Version, "1.0.0")
	}
}
