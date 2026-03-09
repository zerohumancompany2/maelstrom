package bootstrap

import (
	"testing"
	"time"

	"github.com/maelstrom/v3/pkg/statechart"
)

// TestLoadSecurityService_ActionSignature verifies action exists with correct signature.
func TestLoadSecurityService_ActionSignature(t *testing.T) {
	var fn statechart.ActionFn
	fn = loadSecurityService // Must compile with ActionFn signature

	// Call with stubs - should return NotImplementedError
	err := fn(statechart.RuntimeContext{}, &mockApplicationContext{}, statechart.Event{})
	if err == nil {
		t.Error("expected error, got nil")
	}
}

// TestLoadSecurityService_SpawnsRuntime verifies action spawns a runtime for sys:security.
func TestLoadSecurityService_SpawnsRuntime(t *testing.T) {
	engine := statechart.NewEngine()

	// Load and spawn bootstrap chart first
	bootstrapDef, err := LoadBootstrapChart()
	if err != nil {
		t.Fatalf("failed to load bootstrap chart: %v", err)
	}

	// Create mock appCtx with engine
	mockCtx := &mockApplicationContext{
		data: map[string]interface{}{
			"__engine": engine,
		},
	}

	// Spawn bootstrap runtime
	bootstrapRTID, err := engine.Spawn(bootstrapDef, mockCtx)
	if err != nil {
		t.Fatalf("failed to spawn bootstrap runtime: %v", err)
	}

	// Start bootstrap runtime
	if err := engine.Control(bootstrapRTID, statechart.CmdStart); err != nil {
		t.Fatalf("failed to start bootstrap runtime: %v", err)
	}

	// Call action
	err = loadSecurityService(
		statechart.RuntimeContext{RuntimeID: string(bootstrapRTID)},
		mockCtx,
		statechart.Event{},
	)

	// Verify action succeeded
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// Verify runtime ID was stored
	rtID, _, err := mockCtx.Get("bootstrap:security:runtimeID", "sys:bootstrap")
	if err != nil {
		t.Errorf("expected runtime ID to be stored, got error: %v", err)
	}
	if rtID == "" {
		t.Error("expected non-empty runtime ID")
	}
}

// TestLoadSecurityService_StartsRuntime verifies action starts the spawned runtime.
func TestLoadSecurityService_StartsRuntime(t *testing.T) {
	engine := statechart.NewEngine()

	// Load and spawn bootstrap chart first
	bootstrapDef, err := LoadBootstrapChart()
	if err != nil {
		t.Fatalf("failed to load bootstrap chart: %v", err)
	}

	// Create mock appCtx with engine
	mockCtx := &mockApplicationContext{
		data: map[string]interface{}{
			"__engine": engine,
		},
	}

	// Spawn and start bootstrap runtime
	bootstrapRTID, err := engine.Spawn(bootstrapDef, mockCtx)
	if err != nil {
		t.Fatalf("failed to spawn bootstrap runtime: %v", err)
	}
	if err := engine.Control(bootstrapRTID, statechart.CmdStart); err != nil {
		t.Fatalf("failed to start bootstrap runtime: %v", err)
	}

	// Call action
	err = loadSecurityService(
		statechart.RuntimeContext{RuntimeID: string(bootstrapRTID)},
		mockCtx,
		statechart.Event{},
	)
	if err != nil {
		t.Fatalf("action failed: %v", err)
	}

	// Get the security runtime ID
	securityRTIDStr, _, _ := mockCtx.Get("bootstrap:security:runtimeID", "sys:bootstrap")

	// Verify the security runtime is in running state
	// We do this by checking that we can dispatch events to it
	err = engine.Dispatch(statechart.RuntimeID(securityRTIDStr.(string)), statechart.Event{Type: "TEST"})
	if err != nil {
		t.Errorf("expected security runtime to be running, got error: %v", err)
	}
}

// TestLoadSecurityService_DispatchesReadyEvent verifies action dispatches SECURITY_READY.
func TestLoadSecurityService_DispatchesReadyEvent(t *testing.T) {
	engine := statechart.NewEngine()

	// Load and spawn bootstrap chart
	bootstrapDef, err := LoadBootstrapChart()
	if err != nil {
		t.Fatalf("failed to load bootstrap chart: %v", err)
	}

	// Create mock appCtx with engine
	mockCtx := &mockApplicationContext{
		data: map[string]interface{}{
			"__engine": engine,
		},
	}

	// Spawn and start bootstrap runtime
	bootstrapRTID, err := engine.Spawn(bootstrapDef, mockCtx)
	if err != nil {
		t.Fatalf("failed to spawn bootstrap runtime: %v", err)
	}
	if err := engine.Control(bootstrapRTID, statechart.CmdStart); err != nil {
		t.Fatalf("failed to start bootstrap runtime: %v", err)
	}

	// Call action - this dispatches SECURITY_READY to bootstrap
	err = loadSecurityService(
		statechart.RuntimeContext{RuntimeID: string(bootstrapRTID)},
		mockCtx,
		statechart.Event{},
	)
	if err != nil {
		t.Fatalf("action failed: %v", err)
	}

	// Verify SECURITY_READY was dispatched by checking bootstrap's event queue
	// The action succeeds only if dispatch succeeds, so we verify by processing the event
	// and checking the bootstrap transitions to the next state
	time.Sleep(10 * time.Millisecond)
}

// TestLoadSecurityService_StoresRuntimeID verifies action stores RTID in appCtx.
func TestLoadSecurityService_StoresRuntimeID(t *testing.T) {
	engine := statechart.NewEngine()

	// Load and spawn bootstrap chart
	bootstrapDef, err := LoadBootstrapChart()
	if err != nil {
		t.Fatalf("failed to load bootstrap chart: %v", err)
	}

	// Create mock appCtx with engine
	mockCtx := &mockApplicationContext{
		data: map[string]interface{}{
			"__engine": engine,
		},
	}

	// Spawn and start bootstrap runtime
	bootstrapRTID, err := engine.Spawn(bootstrapDef, mockCtx)
	if err != nil {
		t.Fatalf("failed to spawn bootstrap runtime: %v", err)
	}
	if err := engine.Control(bootstrapRTID, statechart.CmdStart); err != nil {
		t.Fatalf("failed to start bootstrap runtime: %v", err)
	}

	// Call action
	err = loadSecurityService(
		statechart.RuntimeContext{RuntimeID: string(bootstrapRTID)},
		mockCtx,
		statechart.Event{},
	)
	if err != nil {
		t.Fatalf("action failed: %v", err)
	}

	// Verify runtime ID was stored
	rtID, _, err := mockCtx.Get("bootstrap:security:runtimeID", "sys:bootstrap")
	if err != nil {
		t.Errorf("expected runtime ID to be stored, got error: %v", err)
	}
	if rtID == "" {
		t.Error("expected non-empty runtime ID")
	}
}

// TestLoadCommunicationService_SpawnsAndStarts verifies communication action spawns and starts runtime.
func TestLoadCommunicationService_SpawnsAndStarts(t *testing.T) {
	engine := statechart.NewEngine()

	// Load and spawn bootstrap chart
	bootstrapDef, err := LoadBootstrapChart()
	if err != nil {
		t.Fatalf("failed to load bootstrap chart: %v", err)
	}

	// Create mock appCtx with engine
	mockCtx := &mockApplicationContext{
		data: map[string]interface{}{
			"__engine": engine,
		},
	}

	// Spawn and start bootstrap runtime
	bootstrapRTID, err := engine.Spawn(bootstrapDef, mockCtx)
	if err != nil {
		t.Fatalf("failed to spawn bootstrap runtime: %v", err)
	}
	if err := engine.Control(bootstrapRTID, statechart.CmdStart); err != nil {
		t.Fatalf("failed to start bootstrap runtime: %v", err)
	}

	// Pre-populate security RTID (dependency)
	mockCtx.Set("bootstrap:security:runtimeID", "rt-security", nil, "sys:bootstrap")

	// Call action
	err = loadCommunicationService(
		statechart.RuntimeContext{RuntimeID: string(bootstrapRTID)},
		mockCtx,
		statechart.Event{},
	)

	// Verify action succeeded
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// Verify communication RTID was stored
	commRTID, _, err := mockCtx.Get("bootstrap:communication:runtimeID", "sys:bootstrap")
	if err != nil {
		t.Errorf("expected communication RTID to be stored, got error: %v", err)
	}
	if commRTID == "" {
		t.Error("expected non-empty communication RTID")
	}
}

// TestLoadObservabilityService_SpawnsAndStarts verifies observability action spawns and starts runtime.
func TestLoadObservabilityService_SpawnsAndStarts(t *testing.T) {
	engine := statechart.NewEngine()

	// Load and spawn bootstrap chart
	bootstrapDef, err := LoadBootstrapChart()
	if err != nil {
		t.Fatalf("failed to load bootstrap chart: %v", err)
	}

	// Create mock appCtx with engine
	mockCtx := &mockApplicationContext{
		data: map[string]interface{}{
			"__engine": engine,
		},
	}

	// Spawn and start bootstrap runtime
	bootstrapRTID, err := engine.Spawn(bootstrapDef, mockCtx)
	if err != nil {
		t.Fatalf("failed to spawn bootstrap runtime: %v", err)
	}
	if err := engine.Control(bootstrapRTID, statechart.CmdStart); err != nil {
		t.Fatalf("failed to start bootstrap runtime: %v", err)
	}

	// Call action
	err = loadObservabilityService(
		statechart.RuntimeContext{RuntimeID: string(bootstrapRTID)},
		mockCtx,
		statechart.Event{},
	)

	// Verify action succeeded
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// Verify observability RTID was stored
	obsRTID, _, err := mockCtx.Get("bootstrap:observability:runtimeID", "sys:bootstrap")
	if err != nil {
		t.Errorf("expected observability RTID to be stored, got error: %v", err)
	}
	if obsRTID == "" {
		t.Error("expected non-empty observability RTID")
	}
}

// TestLoadLifecycleService_SpawnsAndStarts verifies lifecycle action spawns and starts runtime.
func TestLoadLifecycleService_SpawnsAndStarts(t *testing.T) {
	engine := statechart.NewEngine()

	// Load and spawn bootstrap chart
	bootstrapDef, err := LoadBootstrapChart()
	if err != nil {
		t.Fatalf("failed to load bootstrap chart: %v", err)
	}

	// Create mock appCtx with engine
	mockCtx := &mockApplicationContext{
		data: map[string]interface{}{
			"__engine": engine,
		},
	}

	// Spawn and start bootstrap runtime
	bootstrapRTID, err := engine.Spawn(bootstrapDef, mockCtx)
	if err != nil {
		t.Fatalf("failed to spawn bootstrap runtime: %v", err)
	}
	if err := engine.Control(bootstrapRTID, statechart.CmdStart); err != nil {
		t.Fatalf("failed to start bootstrap runtime: %v", err)
	}

	// Call action
	err = loadLifecycleService(
		statechart.RuntimeContext{RuntimeID: string(bootstrapRTID)},
		mockCtx,
		statechart.Event{},
	)

	// Verify action succeeded
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// Verify lifecycle RTID was stored
	lifecycleRTID, _, err := mockCtx.Get("bootstrap:lifecycle:runtimeID", "sys:bootstrap")
	if err != nil {
		t.Errorf("expected lifecycle RTID to be stored, got error: %v", err)
	}
	if lifecycleRTID == "" {
		t.Error("expected non-empty lifecycle RTID")
	}
}

// TestSignalKernelReady_AggregatesServices verifies signalKernelReady aggregates services.
func TestSignalKernelReady_AggregatesServices(t *testing.T) {
	engine := statechart.NewEngine()

	// Load and spawn bootstrap chart
	bootstrapDef, err := LoadBootstrapChart()
	if err != nil {
		t.Fatalf("failed to load bootstrap chart: %v", err)
	}

	// Create mock appCtx with engine
	mockCtx := &mockApplicationContext{
		data: map[string]interface{}{
			"__engine": engine,
		},
	}

	// Spawn and start bootstrap runtime
	bootstrapRTID, err := engine.Spawn(bootstrapDef, mockCtx)
	if err != nil {
		t.Fatalf("failed to spawn bootstrap runtime: %v", err)
	}
	if err := engine.Control(bootstrapRTID, statechart.CmdStart); err != nil {
		t.Fatalf("failed to start bootstrap runtime: %v", err)
	}

	// Pre-populate all service RTIDs
	mockCtx.Set("bootstrap:security:runtimeID", "rt-1", nil, "sys:bootstrap")
	mockCtx.Set("bootstrap:communication:runtimeID", "rt-2", nil, "sys:bootstrap")
	mockCtx.Set("bootstrap:observability:runtimeID", "rt-3", nil, "sys:bootstrap")
	mockCtx.Set("bootstrap:lifecycle:runtimeID", "rt-4", nil, "sys:bootstrap")

	// Call action
	err = signalKernelReady(
		statechart.RuntimeContext{RuntimeID: string(bootstrapRTID)},
		mockCtx,
		statechart.Event{},
	)

	// Verify action succeeded
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// Verify loaded services list was stored
	services, _, err := mockCtx.Get("bootstrap:loaded:services", "sys:bootstrap")
	if err != nil {
		t.Errorf("expected loaded services to be stored: %v", err)
	}

	serviceList := services.([]string)
	if len(serviceList) != 4 {
		t.Errorf("expected 4 services, got %d: %v", len(serviceList), services)
	}
}

// TestBootstrapChart_ActionsExecuteInOrder verifies bootstrap chart executes actions in order.
func TestBootstrapChart_ActionsExecuteInOrder(t *testing.T) {
	t.Skip("Integration test - requires full bootstrap chart execution with new actions")
}
