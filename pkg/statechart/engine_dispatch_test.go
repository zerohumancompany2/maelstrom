package statechart

import (
	"testing"
	"time"

	"github.com/maelstrom/v3/internal/testutil"
)

// ============================================================================
// Event Dispatch & Transitions Tests
// ============================================================================

func TestDispatch_SingleAtomicStateTransition(t *testing.T) {
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

	// Dispatch event that triggers transition
	err = engine.Dispatch(id, Event{Type: "start"})
	if err != nil {
		t.Fatalf("Dispatch failed: %v", err)
	}

	// Allow time for async processing
	time.Sleep(100 * time.Millisecond)

	// Verify state changed (would need snapshot or inspection)
	snap, err := engine.Snapshot(id)
	if err != nil {
		t.Fatalf("Snapshot failed: %v", err)
	}

	if len(snap.ActiveStates) == 0 || snap.ActiveStates[0] != "running" {
		t.Errorf("Expected state 'running', got %v", snap.ActiveStates)
	}
}

func TestDispatch_GuardBlocksTransition(t *testing.T) {
	engine := newTestEngine(t)

	// Register a guard that blocks
	engine.RegisterGuard("alwaysBlock", func(appCtx ApplicationContext, event Event) bool {
		return false
	})

	def := ChartDefinition{
		ID:      "guarded-chart",
		Version: "1.0.0",
		Root: &Node{
			ID:       "idle",
			Children: nil,
			Transitions: []Transition{
				{Event: "try", Guard: "alwaysBlock", Target: "done"},
			},
		},
		InitialState: "idle",
	}

	mockCtx := testutil.NewMockApplicationContext()
	id, err := engine.Spawn(def, mockCtx)
	if err != nil {
		t.Fatalf("Spawn failed: %v", err)
	}

	err = engine.Control(id, CmdStart)
	if err != nil {
		t.Fatalf("Start failed: %v", err)
	}

	// Dispatch event that should be blocked
	err = engine.Dispatch(id, Event{Type: "try"})
	if err != nil {
		t.Fatalf("Dispatch failed: %v", err)
	}

	time.Sleep(100 * time.Millisecond)

	// Verify state is still idle (transition was blocked)
	snap, err := engine.Snapshot(id)
	if err != nil {
		t.Fatalf("Snapshot failed: %v", err)
	}

	if len(snap.ActiveStates) == 0 || snap.ActiveStates[0] != "idle" {
		t.Errorf("Expected state 'idle' (blocked), got %v", snap.ActiveStates)
	}
}

func TestDispatch_GuardAllowsTransition(t *testing.T) {
	engine := newTestEngine(t)

	// Register a guard that allows
	engine.RegisterGuard("alwaysAllow", func(appCtx ApplicationContext, event Event) bool {
		return true
	})

	def := ChartDefinition{
		ID:      "guarded-chart",
		Version: "1.0.0",
		Root: &Node{
			ID:       "idle",
			Children: nil,
			Transitions: []Transition{
				{Event: "try", Guard: "alwaysAllow", Target: "done"},
			},
		},
		InitialState: "idle",
	}

	mockCtx := testutil.NewMockApplicationContext()
	id, err := engine.Spawn(def, mockCtx)
	if err != nil {
		t.Fatalf("Spawn failed: %v", err)
	}

	err = engine.Control(id, CmdStart)
	if err != nil {
		t.Fatalf("Start failed: %v", err)
	}

	// Dispatch event that should be allowed
	err = engine.Dispatch(id, Event{Type: "try"})
	if err != nil {
		t.Fatalf("Dispatch failed: %v", err)
	}

	time.Sleep(100 * time.Millisecond)

	// Verify state changed to done
	snap, err := engine.Snapshot(id)
	if err != nil {
		t.Fatalf("Snapshot failed: %v", err)
	}

	if len(snap.ActiveStates) == 0 || snap.ActiveStates[0] != "done" {
		t.Errorf("Expected state 'done', got %v", snap.ActiveStates)
	}
}

func TestTransition_NoMatchingEvent(t *testing.T) {
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

	// Dispatch event with no matching transition
	err = engine.Dispatch(id, Event{Type: "unknown-event"})
	if err != nil {
		t.Fatalf("Dispatch failed: %v", err)
	}

	time.Sleep(50 * time.Millisecond)

	// State should remain unchanged
	snap, err := engine.Snapshot(id)
	if err != nil {
		t.Fatalf("Snapshot failed: %v", err)
	}

	if len(snap.ActiveStates) == 0 || snap.ActiveStates[0] != "idle" {
		t.Errorf("Expected state 'idle' (no transition), got %v", snap.ActiveStates)
	}
}

func TestMultipleTransitions_SameEventDifferentGuards(t *testing.T) {
	engine := newTestEngine(t)

	engine.RegisterGuard("isHighPriority", func(appCtx ApplicationContext, event Event) bool {
		priority, ok := event.Payload.(string)
		return ok && priority == "high"
	})

	engine.RegisterGuard("isLowPriority", func(appCtx ApplicationContext, event Event) bool {
		priority, ok := event.Payload.(string)
		return ok && priority == "low"
	})

	def := ChartDefinition{
		ID:      "priority-chart",
		Version: "1.0.0",
		Root: &Node{
			ID:       "idle",
			Children: nil,
			Transitions: []Transition{
				{Event: "process", Guard: "isHighPriority", Target: "high-priority"},
				{Event: "process", Guard: "isLowPriority", Target: "low-priority"},
			},
		},
		InitialState: "idle",
	}

	mockCtx := testutil.NewMockApplicationContext()
	id, err := engine.Spawn(def, mockCtx)
	if err != nil {
		t.Fatalf("Spawn failed: %v", err)
	}

	err = engine.Control(id, CmdStart)
	if err != nil {
		t.Fatalf("Start failed: %v", err)
	}

	// Test high priority path
	err = engine.Dispatch(id, Event{Type: "process", Payload: "high"})
	if err != nil {
		t.Fatalf("Dispatch high failed: %v", err)
	}

	time.Sleep(100 * time.Millisecond)

	snap, err := engine.Snapshot(id)
	if err != nil {
		t.Fatalf("Snapshot failed: %v", err)
	}

	if len(snap.ActiveStates) == 0 || snap.ActiveStates[0] != "high-priority" {
		t.Errorf("Expected state 'high-priority', got %v", snap.ActiveStates)
	}
}
