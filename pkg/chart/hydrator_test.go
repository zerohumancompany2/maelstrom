package chart

import (
	"os"
	"testing"
)

// TestHydrateChart_SimpleYAML verifies basic YAML hydration without transforms.
func TestHydrateChart_SimpleYAML(t *testing.T) {
	hydrator := DefaultHydrator()

	content := []byte(`
id: test-chart
version: 1.0.0
spec:
  name: test
`)

	def, err := hydrator(content)
	if err != nil {
		t.Fatalf("hydration failed: %v", err)
	}

	if def.ID != "test-chart" {
		t.Errorf("expected ID 'test-chart', got %q", def.ID)
	}
	if def.Version != "1.0.0" {
		t.Errorf("expected Version '1.0.0', got %q", def.Version)
	}
}

// TestHydrateChart_EnvSubstitution verifies ${VAR} patterns are replaced.
func TestHydrateChart_EnvSubstitution(t *testing.T) {
	os.Setenv("TEST_VAR", "test-value")
	defer os.Unsetenv("TEST_VAR")

	hydrator := DefaultHydrator()

	content := []byte(`
id: test-chart
version: 1.0.0
spec:
  value: ${TEST_VAR}
`)

	def, err := hydrator(content)
	if err != nil {
		t.Fatalf("hydration failed: %v", err)
	}

	// After hydration, the env var should be substituted
	// (Actual check depends on how we store the hydrated spec)
	spec := def.Spec["value"]
	if spec != "test-value" {
		t.Errorf("expected env substitution 'test-value', got %v", spec)
	}
}

// TestHydrateChart_MissingEnvVar verifies error on missing required env var.
func TestHydrateChart_MissingEnvVar(t *testing.T) {
	os.Unsetenv("MISSING_VAR")

	hydrator := DefaultHydrator()

	content := []byte(`
id: test-chart
version: 1.0.0
spec:
  value: ${MISSING_VAR}
`)

	_, err := hydrator(content)
	if err == nil {
		t.Error("expected error for missing env var, got nil")
	}
}

// TestHydrateChart_TemplateExecution verifies {{template}} directives work.
func TestHydrateChart_TemplateExecution(t *testing.T) {
	// TODO: implement after SimpleYAML passes
	t.Skip("templates not yet implemented")
}

// TestHydrateChart_TemplateSyntaxError verifies error on invalid template syntax.
func TestHydrateChart_TemplateSyntaxError(t *testing.T) {
	// TODO: implement after TemplateExecution passes
	t.Skip("templates not yet implemented")
}

// TestHydrateChart_InvalidYAML verifies error on malformed YAML.
func TestHydrateChart_InvalidYAML(t *testing.T) {
	hydrator := DefaultHydrator()

	content := []byte(`
id: test-chart
version: 1.0.0
spec:
  broken: [unclosed
`)

	_, err := hydrator(content)
	if err == nil {
		t.Error("expected error for invalid YAML, got nil")
	}
}

// TestHydrateChart_ValidationError verifies validation catches invalid charts.
func TestHydrateChart_ValidationError(t *testing.T) {
	// TODO: implement validation after basic hydration works
	t.Skip("validation not yet implemented")
}
