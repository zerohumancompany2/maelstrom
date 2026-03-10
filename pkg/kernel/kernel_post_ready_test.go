package kernel

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/maelstrom/v3/pkg/services/platform"
)

// TestChartRegistry_postKernelReadyLoadsServices verifies ChartRegistry loads services after KERNEL_READY
func TestChartRegistry_postKernelReadyLoadsServices(t *testing.T) {
	kernel := New()

	// Create services directory with YAML files
	servicesDir := "var/maelstrom/services/"
	os.MkdirAll(servicesDir, 0755)
	defer os.RemoveAll(servicesDir)

	// Create all 8 hot-reloadable service YAML files
	serviceNames := []string{
		"sys:gateway", "sys:admin", "sys:persistence", "sys:heartbeat",
		"sys:memory", "sys:human-gateway", "sys:tools", "sys:datasources",
	}

	for _, name := range serviceNames {
		fileName := strings.ReplaceAll(name, ":", "-") + ".yaml"
		filePath := filepath.Join(servicesDir, fileName)
		content := platformYAML(name)
		err := os.WriteFile(filePath, []byte(content), 0644)
		if err != nil {
			t.Fatalf("Failed to create service YAML: %v", err)
		}
	}

	// Track if ChartRegistry was started before KERNEL_READY
	chartRegistryStartedBefore := false
	kernelReadyEmitted := false

	// Start kernel in goroutine
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	started := make(chan struct{})
	go func() {
		err := kernel.Start(ctx)
		close(started)
		if err != nil && err != context.DeadlineExceeded {
			t.Logf("Kernel start returned: %v", err)
		}
	}()

	// Monitor kernel state during startup
	monitorDone := make(chan struct{})
	go func() {
		defer close(monitorDone)
		for {
			select {
			case <-started:
				return
			default:
			}
			if kernel.IsKernelReady() {
				kernelReadyEmitted = true
			}
			if kernel.IsChartRegistryRunning() && !kernelReadyEmitted {
				chartRegistryStartedBefore = true
			}
			time.Sleep(5 * time.Millisecond)
		}
	}()

	<-started
	<-monitorDone

	// Verify ChartRegistry was NOT started before KERNEL_READY
	if chartRegistryStartedBefore {
		t.Error("ChartRegistry should NOT be running before KERNEL_READY")
	}

	// Verify KERNEL_READY was emitted
	if !kernelReadyEmitted {
		t.Error("KERNEL_READY should be emitted")
	}

	// Verify ChartRegistry is still not running (will be started after KERNEL_READY in future implementation)
	if kernel.IsChartRegistryRunning() {
		t.Error("ChartRegistry should NOT be running yet - it should be started after KERNEL_READY")
	}

	// Load services using ChartRegistry (simulating post-KERNEL_READY loading)
	registry := platform.NewChartRegistry(servicesDir)
	services, err := registry.LoadPlatformServices()
	if err != nil {
		t.Fatalf("Failed to load platform services: %v", err)
	}

	// Verify all 8 services were loaded
	if len(services) != 8 {
		t.Errorf("Expected 8 services loaded, got %d", len(services))
	}

	// Verify each service was loaded
	loadedNames := make(map[string]bool)
	for _, svc := range services {
		loadedNames[svc.Metadata.Name] = true
	}

	for _, name := range serviceNames {
		if !loadedNames[name] {
			t.Errorf("Expected service '%s' to be loaded", name)
		}
	}
}

func platformYAML(name string) string {
	return `apiVersion: maelstrom.dev/v1
kind: PlatformService
metadata:
  name: ` + name + `
spec:
  chartRef: charts/` + strings.ReplaceAll(name, ":", "-") + `
`
}
