package kernel

import (
	"testing"

	"github.com/maelstrom/v3/pkg/statechart"
)

// TestKernel_UsesStatechartChartDefinition verifies kernel can use statechart.ChartDefinition.
func TestKernel_UsesStatechartChartDefinition(t *testing.T) {
	def := statechart.ChartDefinition{
		ID:      "test-chart",
		Version: "1.0.0",
	}

	if def.GetID() != "test-chart" {
		t.Errorf("GetID() = %q, want %q", def.GetID(), "test-chart")
	}
}
