package statechart

import (
	"errors"
	"testing"
	"time"

	"github.com/maelstrom/v3/internal/testutil"
)

// ============================================================================
// Edge Case Tests
// ============================================================================

func TestDispatch_NonExistentRuntime(t *testing.T) {
	engine := newTestEngine(t)

	err := engine.Dispatch("non-existent-id", Event{Type: "test"})
	if err == nil {
		t.Error("Dispatch to non-existent runtime should fail")
	}

	if !errors.Is(err, ErrRuntimeNotFound) {
		t.Errorf("Expected ErrRuntimeNotFound, got: %v", err)
	}
}

func TestDispatch_RuntimeNotRunning(t *testing.T) {
	engine := newTestEngine(t)
	def := newSimpleAtomicChart()
	mockCtx := testutil.NewMockApplicationContext()

	id, err := engine.Spawn(def, mockCtx)
	if err != nil {
		t.Fatalf("Spawn failed: %v", err)
	}

	// Don't start - runtime is in Created state
	err = engine.Dispatch(id, Event{Type: "test"})
	if err == nil {
		t.Error("Dispatch to non-running runtime should fail")
	}
}

// ============================================================================
// Action Error Handling Tests
// ============================================================================

func TestAction_ErrorDoesNotPreventTransition(t *testing.T) {
	engine := newTestEngine(t)

	engine.RegisterAction("failingAction", func(runtimeCtx RuntimeContext, appCtx ApplicationContext, event Event) error {
		return errors.New("action failed")
	})

	def := ChartDefinition{
		ID:      "error-chart",
		Version: "1.0.0",
		Root: &Node{
			ID:       "idle",
			Children: nil,
			Transitions: []Transition{
				{Event: "go", Target: "done", Actions: []string{"failingAction"}},
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

	err = engine.Dispatch(id, Event{Type: "go"})
	if err != nil {
		t.Fatalf("Dispatch failed: %v", err)
	}

	time.Sleep(100 * time.Millisecond)

	// Transition should have completed despite action error
	snap, err := engine.Snapshot(id)
	if err != nil {
		t.Fatalf("Snapshot failed: %v", err)
	}

	if len(snap.ActiveStates) == 0 || snap.ActiveStates[0] != "done" {
		t.Errorf("Expected state 'done', got %v", snap.ActiveStates)
	}
}

// ============================================================================
// Error Handling Tests
// ============================================================================

func TestDispatch_UnregisteredAction(t *testing.T) {
	engine := newTestEngine(t)

	def := ChartDefinition{
		ID:      "unreg-action-chart",
		Version: "1.0.0",
		Root: &Node{
			ID:       "idle",
			Children: nil,
			Transitions: []Transition{
				{Event: "go", Target: "done", Actions: []string{"unregisteredAction"}},
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

	err = engine.Dispatch(id, Event{Type: "go"})
	if err != nil {
		t.Fatalf("Dispatch failed: %v", err)
	}

	time.Sleep(100 * time.Millisecond)

	// Transition completes even with unregistered actions (just skips them)
	snap, err := engine.Snapshot(id)
	if err != nil {
		t.Fatalf("Snapshot failed: %v", err)
	}

	// State should be 'done' - unregistered actions don't block transitions
	if len(snap.ActiveStates) == 0 || snap.ActiveStates[0] != "done" {
		t.Errorf("Expected state 'done', got %v", snap.ActiveStates)
	}
}

// ============================================================================
// Implied Behavior Tests
// ============================================================================

func TestImplied_FIFOOrderingWithinRuntime(t *testing.T) {
	// Covered by TestDispatch_EventsProcessedFIFO
	t.Skip("Covered by TestDispatch_EventsProcessedFIFO")
}

func TestImplied_RunToCompletion(t *testing.T) {
	// Covered by TestRunToCompletion_EventFullyProcessedBeforeNext
	t.Skip("Covered by TestRunToCompletion_EventFullyProcessedBeforeNext")
}

func TestImplied_RegionIndependence(t *testing.T) {
	// Region independence requires parallel state support with event routing.
	// This is deferred to a follow-up implementation.
	t.Skip("Region independence requires parallel event routing - deferred to v2")
}
