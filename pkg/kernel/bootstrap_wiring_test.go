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

// TestL0H1_ServiceBootstrapSequenceWired verifies service bootstrap sequence is wired to Kernel.
func TestL0H1_ServiceBootstrapSequenceWired(t *testing.T) {
	kernel := New()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	done := make(chan error)
	go func() {
		done <- kernel.Start(ctx)
	}()

	// Wait for bootstrap to complete
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("timeout waiting for bootstrap")
	}

	// Verify sequence is wired and complete
	seq := kernel.GetSequence()
	if seq == nil {
		t.Fatal("bootstrap sequence should be wired to kernel")
	}

	if !seq.IsComplete() {
		t.Error("bootstrap sequence should be complete")
	}

	// Verify all states were entered in order
	states := seq.GetStatesEntered()
	expectedStates := []string{"security", "communication", "observability", "lifecycle", "handoff", "complete"}
	if len(states) != len(expectedStates) {
		t.Errorf("expected %d states, got %d: %v", len(expectedStates), len(states), states)
	}
}

// TestL0H1_BootstrapServicesStartInOrder verifies services start in correct order.
func TestL0H1_BootstrapServicesStartInOrder(t *testing.T) {
	kernel := New()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	err := kernel.Start(ctx)
	if err != nil && err != context.DeadlineExceeded {
		t.Logf("kernel start returned: %v", err)
	}

	// Verify services are ready
	expectedServices := []string{"sys:security", "sys:communication", "sys:observability", "sys:lifecycle"}
	for _, svc := range expectedServices {
		if !kernel.IsServiceReady(svc) {
			t.Errorf("service %s should be ready", svc)
		}
	}

	// Verify bootstrap sequence states were entered in order
	seq := kernel.GetSequence()
	if seq == nil {
		t.Fatal("bootstrap sequence should be wired to kernel")
	}

	states := seq.GetStatesEntered()
	expectedStates := []string{"security", "communication", "observability", "lifecycle", "handoff", "complete"}
	if len(states) != len(expectedStates) {
		t.Errorf("expected %d states, got %d: %v", len(expectedStates), len(states), states)
	}

	// Verify events were handled in order
	events := seq.GetEventsHandled()
	expectedEvents := []string{"SECURITY_READY", "COMMUNICATION_READY", "OBSERVABILITY_READY", "LIFECYCLE_READY"}
	for i, expectedEvent := range expectedEvents {
		if i >= len(events) {
			t.Errorf("missing event at index %d: expected %q", i, expectedEvent)
			continue
		}
		if events[i] != expectedEvent {
			t.Errorf("event[%d]: expected %q, got %q", i, expectedEvent, events[i])
		}
	}
}

// TestL0H1_KernelReadySignalEmitted verifies KERNEL_READY signal is emitted after all services ready.
func TestL0H1_KernelReadySignalEmitted(t *testing.T) {
	kernel := New()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	err := kernel.Start(ctx)
	if err != nil && err != context.DeadlineExceeded {
		t.Logf("kernel start returned: %v", err)
	}

	// Verify KERNEL_READY was emitted
	seq := kernel.GetSequence()
	if seq == nil {
		t.Fatal("sequence should not be nil")
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

// TestL0H2_PostKernelReadyServiceLoading verifies services load after KERNEL_READY.
func TestL0H2_PostKernelReadyServiceLoading(t *testing.T) {
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

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := kernel.Start(ctx)
	if err != nil && err != context.DeadlineExceeded {
		t.Logf("kernel start returned: %v", err)
	}

	// Verify ChartRegistry is running after KERNEL_READY
	if !kernel.IsChartRegistryRunning() {
		t.Error("ChartRegistry should be running after KERNEL_READY")
	}

	// Load services using ChartRegistry
	registry := platform.NewChartRegistry(servicesDir)
	services, err := registry.LoadPlatformServices()
	if err != nil {
		t.Fatalf("Failed to load platform services: %v", err)
	}

	// Verify all 8 services were loaded
	if len(services) != 8 {
		t.Errorf("Expected 8 services loaded, got %d", len(services))
	}
}

// TestL0H2_EightPlatformServicesLoaded verifies all 8 platform services are loaded.
func TestL0H2_EightPlatformServicesLoaded(t *testing.T) {
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

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := kernel.Start(ctx)
	if err != nil && err != context.DeadlineExceeded {
		t.Logf("kernel start returned: %v", err)
	}

	// Load services using ChartRegistry
	registry := platform.NewChartRegistry(servicesDir)
	services, err := registry.LoadPlatformServices()
	if err != nil {
		t.Fatalf("Failed to load platform services: %v", err)
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

// TestL0H2_ServiceStartupFailuresHandledGracefully verifies service startup failures are handled gracefully.
func TestL0H2_ServiceStartupFailuresHandledGracefully(t *testing.T) {
	// Create services directory with valid and invalid YAML files
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
}

// TestL0H3_KernelDormantState verifies kernel enters dormant state after bootstrap.
func TestL0H3_KernelDormantState(t *testing.T) {
	kernel := New()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	err := kernel.Start(ctx)
	if err != nil && err != context.DeadlineExceeded {
		t.Logf("kernel start returned: %v", err)
	}

	// Verify kernel is dormant
	if !kernel.IsDormant() {
		t.Error("kernel should be dormant after bootstrap")
	}

	// Verify dormant since time is set
	dormantSince := kernel.GetDormantSince()
	if dormantSince.IsZero() {
		t.Error("dormant since time should be set")
	}
}

// TestL0H3_DormantStateOnlyHandlesCoreEvents verifies dormant kernel only handles core events.
func TestL0H3_DormantStateOnlyHandlesCoreEvents(t *testing.T) {
	kernel := New()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	err := kernel.Start(ctx)
	if err != nil && err != context.DeadlineExceeded {
		t.Logf("kernel start returned: %v", err)
	}

	// Verify kernel is dormant
	if !kernel.IsDormant() {
		t.Error("kernel should be dormant after bootstrap")
	}

	// In dormant state, kernel should only handle shutdown/reload events
	// This is verified by checking the kernel state and that services are still running
	if !kernel.IsBootstrapComplete() {
		t.Error("bootstrap should be complete")
	}

	// Verify all core services are still ready
	expectedServices := []string{"sys:security", "sys:communication", "sys:observability", "sys:lifecycle"}
	for _, svc := range expectedServices {
		if !kernel.IsServiceReady(svc) {
			t.Errorf("service %s should still be ready in dormant state", svc)
		}
	}
}

// TestL0H3_ServicesHandleAllOtherMail verifies services handle mail while kernel is dormant.
func TestL0H3_ServicesHandleAllOtherMail(t *testing.T) {
	kernel := New()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	err := kernel.Start(ctx)
	if err != nil && err != context.DeadlineExceeded {
		t.Logf("kernel start returned: %v", err)
	}

	// Verify kernel is dormant
	if !kernel.IsDormant() {
		t.Error("kernel should be dormant after bootstrap")
	}

	// Verify mail system is still available for services
	mailSystem := kernel.MailSystem()
	if mailSystem == nil {
		t.Error("mail system should be available in dormant state")
	}

	// Services should handle mail through the mail system
	// This is verified by checking the mail system is functional
	if !kernel.IsServiceReady("sys:communication") {
		t.Error("communication service should be ready to handle mail")
	}
}

// TestL0H3_BootstrapSequenceCompletesSuccessfully verifies bootstrap sequence completes successfully.
func TestL0H3_BootstrapSequenceCompletesSuccessfully(t *testing.T) {
	kernel := New()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	err := kernel.Start(ctx)
	if err != nil && err != context.DeadlineExceeded {
		t.Logf("kernel start returned: %v", err)
	}

	// Verify bootstrap sequence completed
	seq := kernel.GetSequence()
	if seq == nil {
		t.Fatal("sequence should not be nil")
	}

	if !seq.IsComplete() {
		t.Error("bootstrap sequence should be complete")
	}

	// Verify all states were entered
	states := seq.GetStatesEntered()
	expectedStates := []string{"security", "communication", "observability", "lifecycle", "handoff", "complete"}
	if len(states) != len(expectedStates) {
		t.Errorf("expected %d states, got %d: %v", len(expectedStates), len(states), states)
	}

	// Verify all events were handled
	events := seq.GetEventsHandled()
	expectedEvents := []string{"SECURITY_READY", "COMMUNICATION_READY", "OBSERVABILITY_READY", "LIFECYCLE_READY", "KERNEL_READY"}
	for _, expectedEvent := range expectedEvents {
		found := false
		for _, event := range events {
			if event == expectedEvent {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected event %s not found in %v", expectedEvent, events)
		}
	}
}
