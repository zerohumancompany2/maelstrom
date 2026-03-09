package kernel

import (
	"context"
	"testing"
	"time"
)

// TestKernel_StartsBootstrapAndWaits verifies kernel starts bootstrap sequence.
func TestKernel_StartsBootstrapAndWaits(t *testing.T) {
	kernel := New()

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
	kernel := New()

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

func TestKernel_SpawnsAllServices(t *testing.T) {
	// Placeholder for future integration test
	// Verifies all 4 services spawn during bootstrap
}

func TestKernel_ServicesReady(t *testing.T) {
	// Placeholder for future integration test
	// Verifies all services emit ready events immediately
}

func TestKernel_KernelReadyEvent(t *testing.T) {
	// Placeholder for future integration test
	// Verifies KERNEL_READY emitted after all services ready
}

func TestKernel_MailSystemRequired(t *testing.T) {
	// Placeholder for future integration test
	// Verifies mail system exists before services spawn
}

func TestKernelConfig_DefaultValues(t *testing.T) {
	cfg := KernelConfig{}
	if cfg.ChartsDir != "" {
		t.Errorf("expected ChartsDir to be empty, got %q", cfg.ChartsDir)
	}
	if cfg.AppVars != nil && len(cfg.AppVars) > 0 {
		t.Errorf("expected AppVars to be nil or empty, got %v", cfg.AppVars)
	}
}
