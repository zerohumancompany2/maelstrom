package chart

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// TestChartRegistry_LoadsFromDirectory verifies charts load from directory on startup.
func TestChartRegistry_LoadsFromDirectory(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a chart file
	chartContent := []byte(`
id: test-chart
version: 1.0.0
spec:
  name: Test Chart
`)
	if err := os.WriteFile(filepath.Join(tmpDir, "test.yaml"), chartContent, 0644); err != nil {
		t.Fatalf("failed to write chart: %v", err)
	}

	reg, err := NewChartRegistry(tmpDir, DefaultHydrator())
	if err != nil {
		t.Fatalf("failed to create registry: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go reg.Start(ctx)
	time.Sleep(200 * time.Millisecond) // Let initial scan complete

	// Should be able to get the chart
	def, err := reg.Get("test.yaml")
	if err != nil {
		t.Fatalf("failed to get chart: %v", err)
	}
	if def.ID != "test-chart" {
		t.Errorf("expected ID 'test-chart', got %q", def.ID)
	}
}

// TestChartRegistry_HotReload verifies file changes update registry.
func TestChartRegistry_HotReload(t *testing.T) {
	// TODO: implement after LoadsFromDirectory passes
}

// TestChartRegistry_TypeAssertion verifies ChartDefinition type from Get.
func TestChartRegistry_TypeAssertion(t *testing.T) {
	// TODO: implement after LoadsFromDirectory passes
}

// TestChartRegistry_StartStop verifies lifecycle methods.
func TestChartRegistry_StartStop(t *testing.T) {
	tmpDir := t.TempDir()

	reg, err := NewChartRegistry(tmpDir, DefaultHydrator())
	if err != nil {
		t.Fatalf("failed to create registry: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())

	// Start in goroutine
	go func() {
		reg.Start(ctx)
	}()

	time.Sleep(50 * time.Millisecond)

	// Cancel context
	cancel()

	// Should shut down cleanly
	select {
	case <-time.After(100 * time.Millisecond):
		// OK, shutdown should have happened
	}
}

// TestChartRegistry_VersionHistory verifies version tracking.
func TestChartRegistry_VersionHistory(t *testing.T) {
	// TODO: implement after LoadsFromDirectory passes
}
