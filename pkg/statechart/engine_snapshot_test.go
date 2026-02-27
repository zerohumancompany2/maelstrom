package statechart

import (
	"testing"

	"github.com/maelstrom/v3/internal/testutil"
)

// ============================================================================
// Snapshot/Restore Tests
// ============================================================================

func TestSnapshot_CapturesRuntimeState(t *testing.T) {
	engine := newTestEngine(t)
	def := newSimpleAtomicChart()
	mockCtx := testutil.NewMockApplicationContext()

	id, err := engine.Spawn(def, mockCtx)
	if err != nil {
		t.Fatalf("Spawn failed: %v", err)
	}

	err = engine.Control(id, CmdStart)
	if err != nil {
		t.Fatalf("Start failed: %v", err)
	}

	snap, err := engine.Snapshot(id)
	if err != nil {
		t.Fatalf("Snapshot failed: %v", err)
	}

	if snap.RuntimeID != id {
		t.Errorf("Snapshot.RuntimeID = %s, want %s", snap.RuntimeID, id)
	}

	if snap.DefinitionID != def.ID {
		t.Errorf("Snapshot.DefinitionID = %s, want %s", snap.DefinitionID, def.ID)
	}

	if len(snap.ActiveStates) == 0 {
		t.Error("Snapshot.ActiveStates is empty")
	}
}

func TestRestore_CreatesEquivalentRuntime(t *testing.T) {
	engine := newTestEngine(t)
	def := newSimpleAtomicChart()
	mockCtx := testutil.NewMockApplicationContext()

	id, err := engine.Spawn(def, mockCtx)
	if err != nil {
		t.Fatalf("Spawn failed: %v", err)
	}

	err = engine.Control(id, CmdStart)
	if err != nil {
		t.Fatalf("Start failed: %v", err)
	}

	// Take snapshot
	snap, err := engine.Snapshot(id)
	if err != nil {
		t.Fatalf("Snapshot failed: %v", err)
	}

	// Restore with definition and context
	newID, err := engine.Restore(snap, def, mockCtx)
	if err != nil {
		t.Fatalf("Restore failed: %v", err)
	}

	if newID == "" {
		t.Error("Restore returned empty RuntimeID")
	}

	if newID == id {
		t.Error("Restore should return a new RuntimeID, not the original")
	}

	// Verify restored runtime has same configuration
	newSnap, err := engine.Snapshot(newID)
	if err != nil {
		t.Fatalf("Snapshot of restored runtime failed: %v", err)
	}

	if len(newSnap.ActiveStates) != len(snap.ActiveStates) {
		t.Errorf("Restored runtime has different active states: %v vs %v", newSnap.ActiveStates, snap.ActiveStates)
	}
}

func TestSnapshot_IncludesEventQueue(t *testing.T) {
	engine := newTestEngine(t)
	def := newSimpleAtomicChart()
	mockCtx := testutil.NewMockApplicationContext()

	id, err := engine.Spawn(def, mockCtx)
	if err != nil {
		t.Fatalf("Spawn failed: %v", err)
	}

	err = engine.Control(id, CmdStart)
	if err != nil {
		t.Fatalf("Start failed: %v", err)
	}

	// Pause to queue events without processing
	err = engine.Control(id, CmdPause)
	if err != nil {
		t.Fatalf("Pause failed: %v", err)
	}

	// Queue some events (these won't be processed while paused)
	// Note: Dispatch may fail while paused, so we need to test differently
	// For now, just verify snapshot structure

	snap, err := engine.Snapshot(id)
	if err != nil {
		t.Fatalf("Snapshot failed: %v", err)
	}

	// EventQueue should exist (may be empty)
	if snap.EventQueue == nil {
		t.Error("Snapshot.EventQueue is nil")
	}
}

func TestRestore_NewIDDifferentFromOriginal(t *testing.T) {
	engine := newTestEngine(t)
	def := newSimpleAtomicChart()
	mockCtx := testutil.NewMockApplicationContext()

	id, err := engine.Spawn(def, mockCtx)
	if err != nil {
		t.Fatalf("Spawn failed: %v", err)
	}

	err = engine.Control(id, CmdStart)
	if err != nil {
		t.Fatalf("Start failed: %v", err)
	}

	snap, err := engine.Snapshot(id)
	if err != nil {
		t.Fatalf("Snapshot failed: %v", err)
	}

	newID, err := engine.Restore(snap, def, mockCtx)
	if err != nil {
		t.Fatalf("Restore failed: %v", err)
	}

	if newID == id {
		t.Error("Restored runtime should have a new ID")
	}

	if newID == "" {
		t.Error("Restored runtime ID is empty")
	}
}
