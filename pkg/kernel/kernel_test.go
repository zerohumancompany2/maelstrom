package kernel

import (
	"context"
	"testing"
	"time"

	"github.com/maelstrom/v3/pkg/chart"
)

// TestKernel_StartsBootstrapAndWaits verifies kernel starts bootstrap sequence.
func TestKernel_StartsBootstrapAndWaits(t *testing.T) {
	// Create a temporary directory for chart registry
	tmpDir := t.TempDir()

	chartReg, err := chart.NewChartRegistry(tmpDir, chart.DefaultHydrator())
	if err != nil {
		t.Fatalf("failed to create chart registry: %v", err)
	}

	kernel := New(chartReg)

	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()

	// Start kernel in goroutine
	done := make(chan error)
	go func() {
		done <- kernel.Start(ctx)
	}()

	// Wait a bit for bootstrap to progress
	time.Sleep(100 * time.Millisecond)

	// Cancel context to stop kernel
	cancel()

	select {
	case err := <-done:
		if err != context.Canceled && err != context.DeadlineExceeded {
			t.Errorf("unexpected error: %v", err)
		}
	case <-time.After(1 * time.Second):
		t.Fatal("timeout waiting for kernel to stop")
	}
}

// TestKernel_BootstrapCompletes verifies full bootstrap sequence.
func TestKernel_BootstrapCompletes(t *testing.T) {
	tmpDir := t.TempDir()

	chartReg, err := chart.NewChartRegistry(tmpDir, chart.DefaultHydrator())
	if err != nil {
		t.Fatalf("failed to create chart registry: %v", err)
	}

	kernel := New(chartReg)

	// Use a longer timeout to allow bootstrap to complete
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	// Start kernel
	done := make(chan error)
	go func() {
		done <- kernel.Start(ctx)
	}()

	// Wait for bootstrap to complete (need more time for async events)
	time.Sleep(800 * time.Millisecond)

	if !kernel.IsBootstrapComplete() {
		t.Error("bootstrap should be complete")
	}

	cancel()

	select {
	case <-done:
		// OK
	case <-time.After(1 * time.Second):
		t.Fatal("timeout waiting for kernel")
	}
}
