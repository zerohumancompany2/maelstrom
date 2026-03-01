package statechart

import "testing"

func TestChartDefinition_GetID(t *testing.T) {
	def := ChartDefinition{
		ID:      "test-chart",
		Version: "1.0.0",
	}

	if def.GetID() != "test-chart" {
		t.Errorf("GetID() = %q, want %q", def.GetID(), "test-chart")
	}
}
