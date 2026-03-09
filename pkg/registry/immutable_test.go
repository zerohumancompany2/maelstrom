package registry

import (
	"testing"

	"github.com/maelstrom/v3/pkg/mail"
)

// TestBoundaryImmutability_HotReload verifies boundary remains unchanged after hot-reload operation.
func TestBoundaryImmutability_HotReload(t *testing.T) {
	// Given: Chart with boundary set to "inner" loaded in ChartRegistry
	currentBoundary := mail.InnerBoundary
	enforcer := NewImmutableBoundaryEnforcer()

	// When: Hot-reload is triggered with new ChartDefinition that attempts to change boundary to "dmz"
	newBoundary := mail.DMZBoundary
	preservedBoundary := enforcer.PreserveBoundaryDuringHotReload(currentBoundary)

	// Then: Boundary remains "inner" and reload succeeds with boundary preserved
	if preservedBoundary != currentBoundary {
		t.Errorf("boundary changed during hot-reload: expected %q, got %q", currentBoundary, preservedBoundary)
	}

	// Verify that enforcing immutability returns error when boundary changes
	err := enforcer.EnforceBoundaryImmutability(currentBoundary, newBoundary)
	if err == nil {
		t.Error("expected error when boundary changes, got nil")
	}
	if err != ErrBoundaryChangeNotAllowed {
		t.Errorf("expected ErrBoundaryChangeNotAllowed, got: %v", err)
	}
}
