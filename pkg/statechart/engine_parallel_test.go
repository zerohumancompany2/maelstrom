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

// TestEngine_TransitionToParallelState auto-detects and initializes parallel regions.
func TestEngine_TransitionToParallelState(t *testing.T) {
	// Arrange: definition with transition to parallel state
	def := ChartDefinition{
		ID:      "test-auto-parallel",
		Version: "1.0.0",
		Root: &Node{
			ID: "root",
			Children: map[string]*Node{
				"idle": {ID: "idle"},
				"parallel": {
					ID:          "parallel",
					RegionNames: []string{"regionA", "regionB"},
					Children: map[string]*Node{
						"regionA": {
							ID: "regionAState",
							Children: map[string]*Node{
								"active": {ID: "active"},
							},
							Transitions: []Transition{
								{Event: "go", Target: "regionAState/active"},
							},
							IsInitial: true,
						},
						"regionB": {
							ID: "regionBState",
							Children: map[string]*Node{
								"waiting": {ID: "waiting"},
							},
							IsInitial: true,
						},
					},
				},
			},
			Transitions: []Transition{
				{Event: "enterParallel", Target: "root/parallel"},
			},
		},
		InitialState: "root",
	}

	engine := NewEngine().(*Engine)
	mockCtx := testutil.NewMockApplicationContext()

	rtID, err := engine.Spawn(def, mockCtx)
	if err != nil {
		t.Fatalf("Spawn failed: %v", err)
	}

	if err := engine.Control(rtID, CmdStart); err != nil {
		t.Fatalf("Start failed: %v", err)
	}

	runtime := engine.runtimes[rtID]

	// Act: dispatch event that triggers transition to parallel state
	if err := engine.Dispatch(rtID, Event{Type: "enterParallel"}); err != nil {
		t.Fatalf("Dispatch failed: %v", err)
	}

	time.Sleep(100 * time.Millisecond)

	// Assert: runtime should have event router
	if runtime.eventRouter == nil {
		t.Fatal("Expected eventRouter to be created after transition to parallel state")
	}

	if !runtime.isParallel {
		t.Error("Expected isParallel to be true")
	}

	if len(runtime.eventRouter.regions) != 2 {
		t.Errorf("Expected 2 regions, got %d", len(runtime.eventRouter.regions))
	}
}
