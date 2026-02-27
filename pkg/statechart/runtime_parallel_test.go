package statechart

import (
	"testing"
	"time"

	"github.com/maelstrom/v3/internal/testutil"
)

// TestChartRuntime_EntersParallelState verifies runtime spawns regions on parallel entry.
func TestChartRuntime_EntersParallelState(t *testing.T) {
	// Arrange: definition with parallel state (2+ children)
	def := ChartDefinition{
		ID:      "test-parallel",
		Version: "1.0.0",
		Root: &Node{
			ID: "root",
			Children: map[string]*Node{
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
							Transitions: []Transition{
								{Event: "ready", Target: "regionBState/waiting"},
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

	mockCtx := testutil.NewMockApplicationContext()
	rt := &ChartRuntime{
		id:           "test-rt",
		definition:   def,
		state:        RuntimeStateRunning,
		activeState:  def.InitialState,
		regionStates: make(map[string]string),
		eventQueue:   make([]Event, 0),
		appCtx:       mockCtx,
		actions:      make(map[string]ActionFn),
		guards:       make(map[string]GuardFn),
	}

	// Act: transition to parallel state
	// This would normally happen via processEvent, but we're testing the hook
	rt.enterParallelState("root/parallel")

	// Assert: should have event router
	if rt.eventRouter == nil {
		t.Fatal("Expected eventRouter to be created")
	}

	// Should have regions
	if len(rt.eventRouter.regions) != 2 {
		t.Errorf("Expected 2 regions, got %d", len(rt.eventRouter.regions))
	}

	// Should be marked as parallel
	if !rt.isParallel {
		t.Error("Expected isParallel to be true")
	}
}

// TestChartRuntime_RoutesEventsInParallelState verifies event routing in parallel mode.
func TestChartRuntime_RoutesEventsInParallelState(t *testing.T) {
	// Arrange: parallel state with two regions
	def := ChartDefinition{
		ID:      "test-parallel",
		Version: "1.0.0",
		Root: &Node{
			ID: "root",
			Children: map[string]*Node{
				"parallel": {
					ID:          "parallel",
					RegionNames: []string{"regionA", "regionB"},
					Children: map[string]*Node{
						"regionA": {
							ID: "idleA",
							Children: map[string]*Node{
								"active": {ID: "active"},
							},
							Transitions: []Transition{
								{Event: "go", Target: "idleA/active"},
							},
							IsInitial: true,
						},
						"regionB": {
							ID: "idleB",
							Children: map[string]*Node{
								"done": {ID: "done"},
							},
							Transitions: []Transition{
								{Event: "finish", Target: "idleB/done"},
							},
							IsInitial: true,
						},
					},
				},
			},
		},
		InitialState: "root",
	}

	mockCtx := testutil.NewMockApplicationContext()
	rt := &ChartRuntime{
		id:           "test-rt",
		definition:   def,
		state:        RuntimeStateRunning,
		activeState:  "root/parallel",
		regionStates: make(map[string]string),
		eventQueue:   make([]Event, 0),
		appCtx:       mockCtx,
		actions:      make(map[string]ActionFn),
		guards:       make(map[string]GuardFn),
		isParallel:   true,
	}

	// Setup router
	rt.enterParallelState("root/parallel")

	// Allow regions to initialize
	time.Sleep(50 * time.Millisecond)

	// Act: send targeted event to regionA
	rt.processParallelEvent(Event{Type: "go", TargetPath: "region:regionA"})
	time.Sleep(50 * time.Millisecond)

	// Assert: regionA should have transitioned
	if rt.eventRouter.regions["regionA"].stateMachine.activeState != "idleA/active" {
		t.Errorf("regionA should be 'idleA/active', got '%s'",
			rt.eventRouter.regions["regionA"].stateMachine.activeState)
	}

	// regionB should still be in initial state
	if rt.eventRouter.regions["regionB"].stateMachine.activeState != "idleB" {
		t.Errorf("regionB should be 'idleB', got '%s'",
			rt.eventRouter.regions["regionB"].stateMachine.activeState)
	}
}

// TestChartRuntime_ExitsParallelState verifies cleanup on parallel exit.
func TestChartRuntime_ExitsParallelState(t *testing.T) {
	// Arrange: runtime in parallel state
	def := ChartDefinition{
		ID:      "test-parallel",
		Version: "1.0.0",
		Root: &Node{
			ID: "root",
			Children: map[string]*Node{
				"parallel": {
					ID:          "parallel",
					RegionNames: []string{"regionA"},
					Children: map[string]*Node{
						"regionA": {
							ID:       "idleA",
							Children: nil,
						},
					},
				},
			},
		},
		InitialState: "root",
	}

	mockCtx := testutil.NewMockApplicationContext()
	rt := &ChartRuntime{
		id:           "test-rt",
		definition:   def,
		state:        RuntimeStateRunning,
		activeState:  "root/parallel",
		regionStates: make(map[string]string),
		eventQueue:   make([]Event, 0),
		appCtx:       mockCtx,
		actions:      make(map[string]ActionFn),
		guards:       make(map[string]GuardFn),
		isParallel:   true,
	}

	rt.enterParallelState("root/parallel")
	time.Sleep(50 * time.Millisecond)

	// Act: exit parallel state
	rt.exitParallelState()

	// Assert: should not be parallel anymore
	if rt.isParallel {
		t.Error("Expected isParallel to be false after exit")
	}

	// Router should be nil or stopped
	if rt.eventRouter != nil {
		t.Error("Expected eventRouter to be nil after exit")
	}
}
