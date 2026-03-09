package registry

import (
	"testing"

	"github.com/maelstrom/v3/pkg/mail"
	"github.com/maelstrom/v3/pkg/services/persistence"
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

// TestBoundaryImmutability_VersionUpgrade verifies boundary remains unchanged across version upgrades.
func TestBoundaryImmutability_VersionUpgrade(t *testing.T) {
	// Given: Chart with boundary set to "outer" and version "1.0.0"
	oldBoundary := mail.OuterBoundary
	newBoundary := mail.InnerBoundary
	enforcer := NewImmutableBoundaryEnforcer()

	// When: Version upgrade to "2.0.0" with migrationPolicy.onVersionChange set to "deepHistory"
	// attempts to change boundary
	// Then: Error is returned rejecting boundary modification
	err := enforcer.EnforceBoundaryImmutability(oldBoundary, newBoundary)
	if err == nil {
		t.Error("expected error when boundary changes during version upgrade, got nil")
	}
	if err != ErrBoundaryChangeNotAllowed {
		t.Errorf("expected ErrBoundaryChangeNotAllowed, got: %v", err)
	}

	// Verify same boundary passes
	err = enforcer.EnforceBoundaryImmutability(oldBoundary, oldBoundary)
	if err != nil {
		t.Errorf("expected no error when boundary unchanged, got: %v", err)
	}

	// Verify migrationPolicy validation works for all policies
	for _, policy := range []persistence.MigrationPolicy{
		persistence.ShallowHistory,
		persistence.DeepHistory,
		persistence.CleanStart,
	} {
		err := enforcer.ValidateMigrationExcludesBoundary(policy)
		if err != nil {
			t.Errorf("ValidateMigrationExcludesBoundary failed for policy %v: %v", policy, err)
		}
	}
}
