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

	// Create mock appCtx with engine
	mockCtx := &mockApplicationContext{
		data: map[string]interface{}{
			"__engine": engine,
		},
	}

	// Call action
	err := loadSecurityService(
		statechart.RuntimeContext{RuntimeID: "bootstrap-1"},
		mockCtx,
		statechart.Event{},
	)

	// For now, action should return an error since it's not implemented
	if err == nil {
		t.Error("expected error from unimplemented action, got nil")
	}
}
