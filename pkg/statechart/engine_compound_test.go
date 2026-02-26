package statechart

import (
	"testing"
	"time"

	"github.com/maelstrom/v3/internal/testutil"
)

// ============================================================================
// Compound State Tests
// ============================================================================

func TestCompoundState_EntersDefaultChild(t *testing.T) {
	engine := newTestEngine(t)
	def := newCompoundChart()
	mockCtx := testutil.NewMockApplicationContext()

	id, err := engine.Spawn(def, mockCtx)
	if err != nil {
		t.Fatalf("Spawn failed: %v", err)
	}

	err = engine.Control(id, CmdStart)
	if err != nil {
		t.Fatalf("Start failed: %v", err)
	}

	// Initial state should be root/idle (the default child)
	snap, err := engine.Snapshot(id)
	if err != nil {
		t.Fatalf("Snapshot failed: %v", err)
	}

	if len(snap.ActiveStates) == 0 || snap.ActiveStates[0] != "root/idle" {
		t.Errorf("Expected state 'root/idle', got %v", snap.ActiveStates)
	}
}

func TestCompoundState_TransitionToChild(t *testing.T) {
	engine := newTestEngine(t)
	def := newCompoundChart()
	mockCtx := testutil.NewMockApplicationContext()

	id, err := engine.Spawn(def, mockCtx)
	if err != nil {
		t.Fatalf("Spawn failed: %v", err)
	}

	err = engine.Control(id, CmdStart)
	if err != nil {
		t.Fatalf("Start failed: %v", err)
	}

	// Transition from idle to active (which should enter child1)
	err = engine.Dispatch(id, Event{Type: "activate"})
	if err != nil {
		t.Fatalf("Dispatch failed: %v", err)
	}

	time.Sleep(100 * time.Millisecond)

	snap, err := engine.Snapshot(id)
	if err != nil {
		t.Fatalf("Snapshot failed: %v", err)
	}

	// Should be in active/child1 (default child of active, which now has only 1 child)
	if len(snap.ActiveStates) == 0 || snap.ActiveStates[0] != "active/child1" {
		t.Errorf("Expected state 'active/child1', got %v", snap.ActiveStates)
	}
}
