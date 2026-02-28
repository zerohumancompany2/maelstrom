package runtime

import (
	"testing"
)

// TestChartRuntime_InstantiatesFromDefinition verifies runtime creation.
func TestChartRuntime_InstantiatesFromDefinition(t *testing.T) {
	def := map[string]interface{}{
		"id":      "test-chart",
		"version": "1.0.0",
		"spec": map[string]interface{}{
			"initial": "idle",
		},
	}

	rt, err := NewChartRuntime("test-instance", def)
	if err != nil {
		t.Fatalf("failed to create runtime: %v", err)
	}

	if rt.ID != "test-instance" {
		t.Errorf("expected ID 'test-instance', got %q", rt.ID)
	}

	if rt.Context.ChartID != "test-chart" {
		t.Errorf("expected ChartID 'test-chart', got %q", rt.Context.ChartID)
	}
}

// TestChartRuntime_SendEvent verifies event dispatch.
func TestChartRuntime_SendEvent(t *testing.T) {
	// TODO: implement after InstantiatesFromDefinition passes
}

// TestChartRuntime_StateObservation verifies active state tracking.
func TestChartRuntime_StateObservation(t *testing.T) {
	// TODO: implement after SendEvent passes
}
