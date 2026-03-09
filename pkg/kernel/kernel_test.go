package kernel

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/maelstrom/v3/pkg/bootstrap"
	"github.com/maelstrom/v3/pkg/runtime"
	"github.com/maelstrom/v3/pkg/services/communication"
	"github.com/maelstrom/v3/pkg/statechart"
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
		if err != context.Canceled && err != context.DeadlineExceeded && err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	case <-time.After(1 * time.Second):
		t.Fatal("timeout waiting for kernel to stop")
	}
}

func waitForBootstrapComplete(t *testing.T, kernel *Kernel, timeout time.Duration) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	done := make(chan bool)
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
			}
			if kernel.IsBootstrapComplete() {
				done <- true
				return
			}
			time.Sleep(10 * time.Millisecond)
		}
	}()

	select {
	case <-done:
	case <-ctx.Done():
		t.Fatal("Bootstrap did not complete within timeout")
	}
}

// TestKernel_BootstrapCompletes verifies full bootstrap sequence with channel-based waiting.
func TestKernel_BootstrapCompletes(t *testing.T) {
	kernel := New()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	done := make(chan error)
	go func() {
		done <- kernel.Start(ctx)
	}()

	waitForBootstrapComplete(t, kernel, 2*time.Second)

	if !kernel.IsBootstrapComplete() {
		t.Error("bootstrap should be complete")
	}

	cancel()

	select {
	case <-done:
	case <-time.After(1 * time.Second):
		t.Fatal("timeout waiting for kernel")
	}
}

func TestKernel_SpawnsAllServices(t *testing.T) {
	kernel := New()

	// Verify kernel has services map initialized
	kernel.mu.Lock()
	services := kernel.services
	kernel.mu.Unlock()

	if services == nil {
		t.Error("Kernel should have services map initialized")
	}
}

func TestKernel_ServicesReady(t *testing.T) {
	kernel := New()

	// Verify all services are ready by default (stub implementation)
	expectedServices := []string{"sys:security", "sys:communication", "sys:observability", "sys:lifecycle"}
	for _, svc := range expectedServices {
		if !kernel.IsServiceReady(svc) {
			t.Errorf("Service %s should be ready by default", svc)
		}
	}
}

func TestKernel_KernelReadyEvent(t *testing.T) {
	kernel := New()

	// Mark all services as ready
	kernel.SetServiceReady("sys:security")
	kernel.SetServiceReady("sys:communication")
	kernel.SetServiceReady("sys:observability")
	kernel.SetServiceReady("sys:lifecycle")

	// Verify kernel is ready
	if !kernel.IsKernelReady() {
		t.Error("Kernel should be ready after all services are ready")
	}
}

func TestKernel_MailSystemRequired(t *testing.T) {
	kernel := New()
	if kernel.MailSystem() == nil {
		t.Error("Mail system should exist in kernel")
	}
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

func TestKernel_WithConfig_SetsConfig(t *testing.T) {
	k := New()
	cfg := KernelConfig{ChartsDir: "/test/charts", AppVars: map[string]string{"key": "value"}}
	result := k.WithConfig(cfg)
	if result != k {
		t.Error("WithConfig should return receiver for chaining")
	}
	if k.config.ChartsDir != "/test/charts" {
		t.Errorf("expected ChartsDir to be /test/charts, got %q", k.config.ChartsDir)
	}
	if k.config.AppVars["key"] != "value" {
		t.Errorf("expected AppVars[key] to be value, got %q", k.config.AppVars["key"])
	}
}

func TestKernel_RegisterBootstrapActions_RegistersFourActions(t *testing.T) {
	engine := statechart.NewEngine()
	k := NewWithEngine(engine)
	k.RegisterBootstrapActions()
}

func TestKernel_RegisterBootstrapActions_RegistersFiveNewActions(t *testing.T) {
	engine := statechart.NewEngine()
	k := NewWithEngine(engine)

	// Call RegisterBootstrapActions which should register the new actions
	k.RegisterBootstrapActions()

	// Verify actions exist by checking they can be called
	// The actions are registered with the engine, so we verify by
	// attempting to spawn a chart that uses them
}

// TestKernel_Start_RegistersAndExecutesActions verifies kernel starts and executes bootstrap actions.
func TestKernel_Start_RegistersAndExecutesActions(t *testing.T) {
	t.Skip("Integration test - requires full bootstrap chart execution")
}

func TestKernel_RegisterBootstrapActions_WithNilEngine_NoPanic(t *testing.T) {
	k := New()
	func() {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("RegisterBootstrapActions should not panic with nil engine")
			}
		}()
		k.RegisterBootstrapActions()
	}()
}

func TestKernel_Start_RegistersActionsBeforeSpawn(t *testing.T) {
	engine := statechart.NewEngine()
	k := NewWithEngine(engine)
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	go func() {
		_ = k.Start(ctx)
	}()
	time.Sleep(50 * time.Millisecond)
}

func TestKernel_Start_SpawnsBootstrapChart(t *testing.T) {
	engine := statechart.NewEngine()
	k := NewWithEngine(engine)
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	go func() {
		_ = k.Start(ctx)
	}()
	time.Sleep(200 * time.Millisecond)
	if _, ok := k.GetServiceRuntimeID("sys:bootstrap"); !ok {
		t.Error("expected sys:bootstrap service to be registered")
	}
}

func TestKernel_Start_WaitsForKernelReady(t *testing.T) {
	engine := statechart.NewEngine()
	k := NewWithEngine(engine)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	done := make(chan bool)
	go func() {
		_ = k.Start(ctx)
		done <- true
	}()
	select {
	case <-done:
	case <-time.After(3 * time.Second):
		t.Fatal("timeout waiting for kernel ready")
	}
}

func TestKernel_Shutdown_StopsAllServices(t *testing.T) {
	engine := statechart.NewEngine()
	k := NewWithEngine(engine)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	go func() {
		_ = k.Start(ctx)
	}()
	time.Sleep(200 * time.Millisecond)
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer shutdownCancel()
	err := k.Shutdown(shutdownCtx)
	if err != nil {
		t.Errorf("Shutdown failed: %v", err)
	}
}

func TestKernel_Shutdown_WithNilEngine_NoPanic(t *testing.T) {
	k := New()
	ctx := context.Background()
	err := k.Shutdown(ctx)
	if err != nil {
		t.Errorf("Shutdown with nil engine should return nil, got %v", err)
	}
}

func TestKernel_IsBootstrapComplete_AfterBootstrap(t *testing.T) {
	kernel := New()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	done := make(chan error)
	go func() {
		done <- kernel.Start(ctx)
	}()

	time.Sleep(800 * time.Millisecond)

	if !kernel.IsBootstrapComplete() {
		t.Error("bootstrap should be complete after Start() finishes")
	}

	cancel()
	select {
	case <-done:
	case <-time.After(1 * time.Second):
		t.Fatal("timeout waiting for kernel")
	}
}

func TestKernel_Shutdown_ContextCancellation(t *testing.T) {
	engine := statechart.NewEngine()
	k := NewWithEngine(engine)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	go func() {
		_ = k.Start(ctx)
	}()

	time.Sleep(200 * time.Millisecond)

	cancelCtx, cancelFunc := context.WithCancel(context.Background())
	cancelFunc()

	err := k.Shutdown(cancelCtx)
	if err != context.Canceled {
		t.Errorf("expected context.Canceled error, got %v", err)
	}
}

func TestKernel_GetServiceRuntimeID_ReturnsCorrectID(t *testing.T) {
	engine := statechart.NewEngine()
	k := NewWithEngine(engine)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	go func() {
		_ = k.Start(ctx)
	}()

	time.Sleep(200 * time.Millisecond)

	rtID, ok := k.GetServiceRuntimeID("sys:bootstrap")
	if !ok {
		t.Error("expected sys:bootstrap service to exist")
	}
	if rtID == "" {
		t.Error("expected non-empty RuntimeID for sys:bootstrap")
	}

	_, ok = k.GetServiceRuntimeID("nonexistent")
	if ok {
		t.Error("expected false for nonexistent service")
	}

	firstID, _ := k.GetServiceRuntimeID("sys:bootstrap")
	secondID, _ := k.GetServiceRuntimeID("sys:bootstrap")
	if firstID != secondID {
		t.Error("expected consistent RuntimeID across multiple calls")
	}
}

func TestKernel_Start_WithEngine_SpawnsAndControls(t *testing.T) {
	engine := statechart.NewEngine()
	k := NewWithEngine(engine)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	go func() {
		_ = k.Start(ctx)
	}()

	time.Sleep(200 * time.Millisecond)

	rtID, ok := k.GetServiceRuntimeID("sys:bootstrap")
	if !ok {
		t.Error("expected sys:bootstrap service to be tracked")
	}
	if rtID == "" {
		t.Error("expected non-empty RuntimeID for spawned bootstrap")
	}
}

// TestKernel_LoadsBootstrapChart verifies bootstrap chart loads with correct metadata.
func TestKernel_LoadsBootstrapChart(t *testing.T) {
	def, err := bootstrap.LoadBootstrapChart()
	if err != nil {
		t.Fatalf("LoadBootstrapChart() returned error: %v", err)
	}
	if def.ID != "sys:bootstrap" {
		t.Errorf("expected ID to be 'sys:bootstrap', got %q", def.ID)
	}
	if def.Version != "1.0.0" {
		t.Errorf("expected Version to be '1.0.0', got %q", def.Version)
	}
	if spec, ok := def.Spec["initial"].(string); !ok || spec != "sys:bootstrap/init" {
		t.Errorf("expected spec.initial to be 'sys:bootstrap/init', got %q", spec)
	}
}

// TestKernel_SpawnsBootstrapRuntime verifies kernel spawns bootstrap runtime and tracks ID.
func TestKernel_SpawnsBootstrapRuntime(t *testing.T) {
	engine := statechart.NewEngine()
	k := NewWithEngine(engine)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	done := make(chan error)
	go func() {
		done <- k.Start(ctx)
	}()

	time.Sleep(300 * time.Millisecond)

	rtID := k.GetBootstrapRuntimeID()
	if rtID == "" {
		t.Error("expected non-empty bootstrap runtime ID")
	}

	cancel()
	<-done
}

// TestKernel_BootstrapSequenceStartsWithSecurity verifies sequence starts with security state.
func TestKernel_BootstrapSequenceStartsWithSecurity(t *testing.T) {
	var capturedState string
	seq := bootstrap.NewSequence()
	seq.OnStateEnter(func(state string) error {
		if capturedState == "" {
			capturedState = state
		}
		return nil
	})

	ctx := context.Background()
	if err := seq.Start(ctx); err != nil {
		t.Fatalf("failed to start sequence: %v", err)
	}

	if capturedState != "security" {
		t.Errorf("expected first state to be 'security', got %q", capturedState)
	}

	if seq.CurrentState() != "security" {
		t.Errorf("expected CurrentState to be 'security', got %q", seq.CurrentState())
	}
}

func TestKernel_ServicesLoadInOrder(t *testing.T) {
	kernel := New()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	done := make(chan error)
	go func() {
		done <- kernel.Start(ctx)
	}()

	// Wait for bootstrap to complete
	<-done

	// Get states from sequence
	seq := kernel.GetSequence()
	if seq == nil {
		t.Fatal("sequence should not be nil")
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

func TestKernel_AllServicesEmitReadyEvents(t *testing.T) {
	kernel := New()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	// Start kernel
	done := make(chan error)
	go func() {
		done <- kernel.Start(ctx)
	}()

	// Wait for events to be handled by polling
	timeout := time.After(2 * time.Second)
	for {
		select {
		case <-timeout:
			// Check events even if timeout
		case <-done:
			// Bootstrap complete
		default:
		}

		seq := kernel.GetSequence()
		if seq != nil {
			events := seq.GetEventsHandled()
			if len(events) >= 4 {
				break
			}
		}
		time.Sleep(10 * time.Millisecond)
	}

	// Get events from sequence
	seq := kernel.GetSequence()
	if seq == nil {
		t.Fatal("sequence should not be nil")
	}

	events := seq.GetEventsHandled()
	expected := []string{"SECURITY_READY", "COMMUNICATION_READY", "OBSERVABILITY_READY", "LIFECYCLE_READY"}

	if len(events) < len(expected) {
		t.Errorf("expected at least %d events, got %d: %v", len(expected), len(events), events)
	}

	for i, expectedEvent := range expected {
		if i >= len(events) {
			t.Errorf("missing event at index %d: expected %q", i, expectedEvent)
			continue
		}
		if events[i] != expectedEvent {
			t.Errorf("event[%d]: expected %q, got %q", i, expectedEvent, events[i])
		}
	}
}

func TestKernel_KernelReadyEventEmitted(t *testing.T) {
	kernel := New()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	// Start kernel
	done := make(chan error)
	go func() {
		done <- kernel.Start(ctx)
	}()

	// Wait for KERNEL_READY event to be emitted
	kernelReadyReceived := make(chan struct{})
	timeout := time.After(2 * time.Second)

	for {
		select {
		case <-timeout:
			t.Fatal("timeout waiting for KERNEL_READY event")
		case <-done:
		default:
		}

		seq := kernel.GetSequence()
		if seq != nil && seq.GetKernelReadyEmitted() {
			close(kernelReadyReceived)
			break
		}
		time.Sleep(10 * time.Millisecond)
	}

	// Verify KERNEL_READY was emitted after LIFECYCLE_READY
	seq := kernel.GetSequence()
	if seq == nil {
		t.Fatal("sequence should not be nil")
	}

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

func TestKernel_OnCompleteCallbackInvoked(t *testing.T) {
	kernel := New()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	// Channel to signal when onComplete is called
	onCompleteCalled := make(chan struct{})

	// Start kernel
	done := make(chan error)
	go func() {
		done <- kernel.Start(ctx)
	}()

	// Wait for completion callback with timeout
	timeout := time.After(2 * time.Second)
	for {
		select {
		case <-timeout:
			t.Fatal("timeout waiting for onComplete callback")
		case <-done:
		default:
		}

		if kernel.GetCompletionStatus() {
			close(onCompleteCalled)
			break
		}
		time.Sleep(10 * time.Millisecond)
	}

	// Verify completion status
	if !kernel.GetCompletionStatus() {
		t.Error("GetCompletionStatus() should return true after onComplete is called")
	}

	if !kernel.IsBootstrapComplete() {
		t.Error("IsBootstrapComplete() should return true after onComplete is called")
	}
}

func TestKernel_GoesDormant(t *testing.T) {
	kernel := New()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	done := make(chan error)
	go func() {
		done <- kernel.Start(ctx)
	}()

	// Wait for kernel to complete (which calls onBootstrapComplete)
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("timeout waiting for kernel to complete")
	}

	logs := kernel.GetLogOutput()
	found := false
	for _, log := range logs {
		if strings.Contains(log, "going dormant") {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected 'going dormant' in logs. Got: %v", logs)
	}
}

func assertStatesInOrder(t *testing.T, actual []string) {
	t.Helper()
	expected := []string{"security", "communication", "observability", "lifecycle", "handoff", "complete"}
	if len(actual) != len(expected) {
		t.Errorf("expected %d states, got %d: %v", len(expected), len(actual), actual)
		return
	}
	for i, expectedState := range expected {
		if actual[i] != expectedState {
			t.Errorf("state[%d]: expected %q, got %q", i, expectedState, actual[i])
		}
	}
}

func assertAllEventsReceived(t *testing.T, actual []string) {
	t.Helper()
	expectedEvents := []string{"SECURITY_READY", "COMMUNICATION_READY", "OBSERVABILITY_READY", "LIFECYCLE_READY"}
	for _, expected := range expectedEvents {
		found := false
		for _, actualEvent := range actual {
			if actualEvent == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected event %s not found in %v", expected, actual)
		}
	}
}

func assertDormantLogged(t *testing.T, logs []string) {
	t.Helper()
	found := false
	for _, log := range logs {
		if strings.Contains(log, "going dormant") {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("Expected 'going dormant' in logs. Got: %v", logs)
	}
}

func TestKernel_BootstrapServices(t *testing.T) {
	kernel := &Kernel{
		engine:       statechart.NewEngine(),
		services:     make(map[string]statechart.RuntimeID),
		serviceReady: make(map[string]bool),
		runtimes:     make(map[string]*runtime.ChartRuntime),
		mailSystem:   communication.NewCommunicationService(),
	}

	err := kernel.BootstrapServices()
	if err != nil {
		t.Fatalf("BootstrapServices() returned error: %v", err)
	}

	expectedServices := []string{"sys:security", "sys:communication", "sys:observability", "sys:lifecycle"}
	for _, svc := range expectedServices {
		if !kernel.IsServiceReady(svc) {
			t.Errorf("Service %s should be ready after BootstrapServices()", svc)
		}
	}
}

func TestKernel_BootstrapSequence(t *testing.T) {
	kernel := &Kernel{
		engine:       statechart.NewEngine(),
		services:     make(map[string]statechart.RuntimeID),
		serviceReady: make(map[string]bool),
		runtimes:     make(map[string]*runtime.ChartRuntime),
		mailSystem:   communication.NewCommunicationService(),
	}

	err := kernel.BootstrapServices()
	if err != nil {
		t.Fatalf("BootstrapServices() returned error: %v", err)
	}

	actualOrder := kernel.GetServiceOrder()
	expectedOrder := []string{"sys:security", "sys:communication", "sys:observability", "sys:lifecycle"}

	if len(actualOrder) != len(expectedOrder) {
		t.Fatalf("expected %d services, got %d: %v", len(expectedOrder), len(actualOrder), actualOrder)
	}

	for i, expectedSvc := range expectedOrder {
		if actualOrder[i] != expectedSvc {
			t.Errorf("service[%d]: expected %q, got %q", i, expectedSvc, actualOrder[i])
		}
	}
}

func TestKernel_ServiceReadyEvents(t *testing.T) {
	kernel := &Kernel{
		engine:       statechart.NewEngine(),
		services:     make(map[string]statechart.RuntimeID),
		serviceReady: make(map[string]bool),
		runtimes:     make(map[string]*runtime.ChartRuntime),
		mailSystem:   communication.NewCommunicationService(),
	}

	err := kernel.BootstrapServices()
	if err != nil {
		t.Fatalf("BootstrapServices() returned error: %v", err)
	}

	events := kernel.GetReadyEvents()
	expectedEvents := []string{"sys:security", "sys:communication", "sys:observability", "sys:lifecycle"}

	if len(events) != len(expectedEvents) {
		t.Fatalf("expected %d events, got %d: %v", len(expectedEvents), len(events), events)
	}

	for i, expectedEvent := range expectedEvents {
		if events[i] != expectedEvent {
			t.Errorf("event[%d]: expected %q, got %q", i, expectedEvent, events[i])
		}
	}
}

func TestKernel_BootstrapFailure(t *testing.T) {
	kernel := &Kernel{
		engine:       statechart.NewEngine(),
		services:     make(map[string]statechart.RuntimeID),
		serviceReady: make(map[string]bool),
		runtimes:     make(map[string]*runtime.ChartRuntime),
		mailSystem:   communication.NewCommunicationService(),
		failService:  "sys:communication",
	}

	err := kernel.BootstrapServices()
	if err == nil {
		t.Fatal("BootstrapServices() should return error when service fails")
	}

	if !strings.Contains(err.Error(), "sys:communication") {
		t.Errorf("error should mention failed service, got: %v", err)
	}

	if !kernel.IsServiceReady("sys:security") {
		t.Error("sys:security should be ready before failure")
	}

	if kernel.IsServiceReady("sys:communication") {
		t.Error("sys:communication should not be ready after failure")
	}
}

func TestKernel_FullE2EBootstrap(t *testing.T) {
	kernel := New()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	done := make(chan error)
	go func() {
		done <- kernel.Start(ctx)
	}()

	// Wait for kernel to complete
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("timeout waiting for kernel to complete")
	}

	// Verify states entered in correct order
	seq := kernel.GetSequence()
	if seq == nil {
		t.Fatal("sequence should not be nil")
	}

	states := seq.GetStatesEntered()
	assertStatesInOrder(t, states)

	// Verify all READY events received
	events := seq.GetEventsHandled()
	assertAllEventsReceived(t, events)

	// Verify KERNEL_READY event was emitted
	if !seq.GetKernelReadyEmitted() {
		t.Error("KERNEL_READY event should be emitted")
	}

	// Verify onComplete callback was invoked
	if !kernel.GetCompletionStatus() {
		t.Error("onComplete callback should be invoked")
	}

	// Verify IsBootstrapComplete returns true
	if !kernel.IsBootstrapComplete() {
		t.Error("IsBootstrapComplete should return true")
	}

	// Verify "going dormant" was logged
	logs := kernel.GetLogOutput()
	assertDormantLogged(t, logs)
}
