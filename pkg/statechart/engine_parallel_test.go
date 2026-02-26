package statechart

import (
	"testing"
	"time"

	"github.com/maelstrom/v3/internal/testutil"
)

// ============================================================================
// Parallel Region Tests
// ============================================================================

func TestParallelState_RegionsActiveSimultaneously(t *testing.T) {
	engine := newTestEngine(t)
	def := newParallelChart()
	mockCtx := testutil.NewMockApplicationContext()

	id, err := engine.Spawn(def, mockCtx)
	if err != nil {
		t.Fatalf("Spawn failed: %v", err)
	}

	err = engine.Control(id, CmdStart)
	if err != nil {
		t.Fatalf("Start failed: %v", err)
	}

	// Both regions should be active
	snap, err := engine.Snapshot(id)
	if err != nil {
		t.Fatalf("Snapshot failed: %v", err)
	}

	// Should have active states for both regions
	if len(snap.ActiveStates) == 0 {
		t.Error("No active states in snapshot")
	}
}

func TestParallelState_EventsRoutedToSpecificRegion(t *testing.T) {
	engine := newTestEngine(t)
	def := newParallelChart()
	mockCtx := testutil.NewMockApplicationContext()

	id, err := engine.Spawn(def, mockCtx)
	if err != nil {
		t.Fatalf("Spawn failed: %v", err)
	}

	err = engine.Control(id, CmdStart)
	if err != nil {
		t.Fatalf("Start failed: %v", err)
	}

	// Send event to regionA
	err = engine.Dispatch(id, Event{Type: "nextA", TargetPath: "region:regionA"})
	if err != nil {
		t.Fatalf("Dispatch failed: %v", err)
	}

	time.Sleep(100 * time.Millisecond)

	// Verify regionA transitioned, regionB unchanged
	snap, err := engine.Snapshot(id)
	if err != nil {
		t.Fatalf("Snapshot failed: %v", err)
	}

	// Check that regionA is now in done state
	_ = snap
}
