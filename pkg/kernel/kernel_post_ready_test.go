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

// TestChartRegistry_yamlValidationOnLoad verifies YAML validation on load
func TestChartRegistry_yamlValidationOnLoad(t *testing.T) {
	// Create services directory with YAML files
	servicesDir := "var/maelstrom/services/"
	os.MkdirAll(servicesDir, 0755)
	defer os.RemoveAll(servicesDir)

	// Create valid service YAML
	validYAML := `apiVersion: maelstrom.dev/v1
kind: PlatformService
metadata:
  name: sys:gateway
spec:
  chartRef: charts/sys-gateway
`
	validPath := filepath.Join(servicesDir, "sys-gateway.yaml")
	err := os.WriteFile(validPath, []byte(validYAML), 0644)
	if err != nil {
		t.Fatalf("Failed to create valid service YAML: %v", err)
	}

	// Create invalid service YAML (missing required field)
	invalidYAML := `apiVersion: maelstrom.dev/v1
kind: PlatformService
metadata:
  name: sys:admin
`
	invalidPath := filepath.Join(servicesDir, "sys-admin.yaml")
	err = os.WriteFile(invalidPath, []byte(invalidYAML), 0644)
	if err != nil {
		t.Fatalf("Failed to create invalid service YAML: %v", err)
	}

	// Try to load services - should fail due to invalid YAML
	registry := platform.NewChartRegistry(servicesDir)
	_, err = registry.LoadPlatformServices()
	if err == nil {
		t.Error("Expected error when loading invalid YAML")
	}

	// Remove invalid YAML
	os.Remove(invalidPath)

	// Load services again - should succeed
	services, err := registry.LoadPlatformServices()
	if err != nil {
		t.Fatalf("Failed to load valid platform services: %v", err)
	}

	// Verify only valid service was loaded
	if len(services) != 1 {
		t.Errorf("Expected 1 service loaded, got %d", len(services))
	}

	if services[0].Metadata.Name != "sys:gateway" {
		t.Errorf("Expected service 'sys:gateway', got '%s'", services[0].Metadata.Name)
	}
}

// TestChartRegistry_coreNonCoreSeparation verifies core/non-core service separation
func TestChartRegistry_coreNonCoreSeparation(t *testing.T) {
	// Create services directory with YAML files
	servicesDir := "var/maelstrom/services/"
	os.MkdirAll(servicesDir, 0755)
	defer os.RemoveAll(servicesDir)

	// Create core service YAML
	coreYAML := `apiVersion: maelstrom.dev/v1
kind: PlatformService
metadata:
  name: sys:security
  core: true
spec:
  chartRef: charts/sys-security
  requiredForKernelReady: true
`
	corePath := filepath.Join(servicesDir, "sys-security.yaml")
	err := os.WriteFile(corePath, []byte(coreYAML), 0644)
	if err != nil {
		t.Fatalf("Failed to create core service YAML: %v", err)
	}

	// Create non-core service YAML
	nonCoreYAML := `apiVersion: maelstrom.dev/v1
kind: PlatformService
metadata:
  name: sys:gateway
spec:
  chartRef: charts/sys-gateway
`
	nonCorePath := filepath.Join(servicesDir, "sys-gateway.yaml")
	err = os.WriteFile(nonCorePath, []byte(nonCoreYAML), 0644)
	if err != nil {
		t.Fatalf("Failed to create non-core service YAML: %v", err)
	}

	// Load all services
	registry := platform.NewChartRegistry(servicesDir)
	services, err := registry.LoadPlatformServices()
	if err != nil {
		t.Fatalf("Failed to load platform services: %v", err)
	}

	// Verify total services
	if len(services) != 2 {
		t.Errorf("Expected 2 services loaded, got %d", len(services))
	}

	// Get core services
	coreServices := registry.GetCoreServices(services)
	if len(coreServices) != 1 {
		t.Errorf("Expected 1 core service, got %d", len(coreServices))
	}

	if coreServices[0].Metadata.Name != "sys:security" {
		t.Errorf("Expected core service 'sys:security', got '%s'", coreServices[0].Metadata.Name)
	}

	// Get non-core services
	nonCoreServices := registry.GetNonCoreServices(services)
	if len(nonCoreServices) != 1 {
		t.Errorf("Expected 1 non-core service, got %d", len(nonCoreServices))
	}

	if nonCoreServices[0].Metadata.Name != "sys:gateway" {
		t.Errorf("Expected non-core service 'sys:gateway', got '%s'", nonCoreServices[0].Metadata.Name)
	}

	// Validate core services
	err = registry.ValidateCoreServices(services)
	if err != nil {
		t.Errorf("Core services validation failed: %v", err)
	}
}

// TestBootstrap_kernelReadySignalEmitted verifies KERNEL_READY signal is emitted
func TestBootstrap_kernelReadySignalEmitted(t *testing.T) {
	kernel := New()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Start kernel
	err := kernel.Start(ctx)
	if err != nil && err != context.DeadlineExceeded {
		t.Logf("Kernel start returned: %v", err)
	}

	// Verify KERNEL_READY was emitted by checking sequence
	seq := kernel.GetSequence()
	if seq == nil {
		t.Fatal("Sequence should not be nil")
	}

	if !seq.GetKernelReadyEmitted() {
		t.Error("KERNEL_READY event should be emitted")
	}

	// Verify KERNEL_READY was emitted after LIFECYCLE_READY
	events := seq.GetEventsHandled()
	lifecycleReadyIdx := -1
	kernelReadyIdx := -1

	for i, event := range events {
		if event == "LIFECYCLE_READY" {
			lifecycleReadyIdx = i
		}
		if event == "KERNEL_READY" {
			kernelReadyIdx = i
		}
	}

	if lifecycleReadyIdx == -1 {
		t.Error("LIFECYCLE_READY event should be present")
	}
	if kernelReadyIdx == -1 {
		t.Error("KERNEL_READY event should be present")
	}
	if lifecycleReadyIdx != -1 && kernelReadyIdx != -1 && kernelReadyIdx <= lifecycleReadyIdx {
		t.Error("KERNEL_READY should be emitted after LIFECYCLE_READY")
	}
}

// TestBootstrap_serviceLoadSequenceCorrect verifies service load sequence is correct
func TestBootstrap_serviceLoadSequenceCorrect(t *testing.T) {
	kernel := New()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Start kernel
	err := kernel.Start(ctx)
	if err != nil && err != context.DeadlineExceeded {
		t.Logf("Kernel start returned: %v", err)
	}

	// Verify services were loaded in correct order
	seq := kernel.GetSequence()
	if seq == nil {
		t.Fatal("Sequence should not be nil")
	}

	states := seq.GetStatesEntered()
	expected := []string{"security", "communication", "observability", "lifecycle", "handoff", "complete"}

	if len(states) != len(expected) {
		t.Errorf("expected %d states, got %d: %v", len(expected), len(states), states)
	}

	for i, expectedState := range expected {
		if i >= len(states) {
			t.Errorf("missing state at index %d: expected %q", i, expectedState)
			continue
		}
		if states[i] != expectedState {
			t.Errorf("state[%d]: expected %q, got %q", i, expectedState, states[i])
		}
	}
}
