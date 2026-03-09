package bootstrap

import (
	"testing"

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
