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
