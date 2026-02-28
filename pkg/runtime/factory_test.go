package runtime

import (
	"testing"

	"github.com/maelstrom/v3/pkg/chart"
	"github.com/maelstrom/v3/pkg/registry"
)

// TestFactory_CreateRuntime verifies single runtime creation.
func TestFactory_CreateRuntime(t *testing.T) {
	reg := registry.New()

	// Store a chart definition
	reg.Set("test.yaml", chart.ChartDefinition{
		ID:      "test-chart",
		Version: "1.0.0",
	})

	factory := NewFactory(reg)
	rt, err := factory.Create("test.yaml")
	if err != nil {
		t.Fatalf("failed to create runtime: %v", err)
	}

	if rt == nil {
		t.Fatal("runtime should not be nil")
	}

	if rt.Context.ChartID != "test-chart" {
		t.Errorf("expected ChartID 'test-chart', got %q", rt.Context.ChartID)
	}
}

// TestFactory_CreateAllFromRegistry verifies bulk creation.
func TestFactory_CreateAllFromRegistry(t *testing.T) {
	reg := registry.New()

	// Store multiple charts
	reg.Set("chart1.yaml", chart.ChartDefinition{ID: "chart-1", Version: "1.0.0"})
	reg.Set("chart2.yaml", chart.ChartDefinition{ID: "chart-2", Version: "1.0.0"})

	factory := NewFactory(reg)
	runtimes, err := factory.CreateAll()
	if err != nil {
		t.Fatalf("failed to create runtimes: %v", err)
	}

	if len(runtimes) != 2 {
		t.Errorf("expected 2 runtimes, got %d", len(runtimes))
	}
}
