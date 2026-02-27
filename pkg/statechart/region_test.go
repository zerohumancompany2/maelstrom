package statechart

import (
	"testing"
	"time"

	"github.com/maelstrom/v3/internal/testutil"
)

// TestRegionRuntime_UserEventProcessing verifies regions process user events.
func TestRegionRuntime_UserEventProcessing(t *testing.T) {
	// Arrange: simple state machine in region
	def := ChartDefinition{
		ID:      "test-region",
		Version: "1.0.0",
		Root: &Node{
			ID: "idle",
			Children: map[string]*Node{
				"running": {ID: "running"},
			},
			Transitions: []Transition{
				{Event: "start", Target: "idle/running"},
			},
		},
		InitialState: "idle",
	}

	mockCtx := testutil.NewMockApplicationContext()
	sm := &StateMachine{
		definition:  def,
		activeState: def.InitialState,
		actions:     make(map[string]ActionFn),
		guards:      make(map[string]GuardFn),
		appCtx:      mockCtx,
	}

	inputChan := make(chan Event, 10)
	outputChan := make(chan Event, 10)

	region := &RegionRuntime{
		name:         "testRegion",
		stateMachine: sm,
		inputChan:    inputChan,
		outputChan:   outputChan,
		state:        RegionStateRunning,
	}

	// Start region goroutine
	go region.Run()

	// Send sys:enter to trigger initial entry
	inputChan <- Event{Type: SysEnter}
	time.Sleep(10 * time.Millisecond)

	// Act: send user event
	inputChan <- Event{Type: "start"}

	// Wait for processing
	time.Sleep(50 * time.Millisecond)

	// Assert: state machine should have transitioned
	if region.stateMachine.activeState != "idle/running" {
		t.Errorf("Expected state 'running', got '%s'", region.stateMachine.activeState)
	}
}

// TestRegionRuntime_EmitsSysDone verifies region emits SysDone on final state.
func TestRegionRuntime_EmitsSysDone(t *testing.T) {
	def := ChartDefinition{
		ID:      "test-final",
		Version: "1.0.0",
		Root: &Node{
			ID: "processing",
			Children: map[string]*Node{
				"done": {ID: "done"}, // No children = final state (NodeTypeAtomic)
			},
			Transitions: []Transition{
				{Event: "complete", Target: "processing/done"},
			},
		},
		InitialState: "processing",
	}

	mockCtx := testutil.NewMockApplicationContext()
	sm := &StateMachine{
		definition:  def,
		activeState: def.InitialState,
		actions:     make(map[string]ActionFn),
		guards:      make(map[string]GuardFn),
		appCtx:      mockCtx,
	}

	inputChan := make(chan Event, 10)
	outputChan := make(chan Event, 10)

	region := &RegionRuntime{
		name:         "testRegion",
		stateMachine: sm,
		inputChan:    inputChan,
		outputChan:   outputChan,
		state:        RegionStateRunning,
	}

	go region.Run()

	// Send event that triggers transition to final state
	inputChan <- Event{Type: "complete"}

	// First, we should get SysTransition
	select {
	case ev := <-outputChan:
		if ev.Type != SysTransition {
			t.Errorf("Expected SysTransition first, got '%s'", ev.Type)
		}
	case <-time.After(100 * time.Millisecond):
		t.Error("Timeout waiting for SysTransition")
		return
	}

	// Then, we should get SysDone
	select {
	case ev := <-outputChan:
		if ev.Type != SysDone {
			t.Errorf("Expected SysDone, got '%s'", ev.Type)
		}
		if ev.Source != "region:testRegion" {
			t.Errorf("Expected Source 'region:testRegion', got '%s'", ev.Source)
		}
	case <-time.After(100 * time.Millisecond):
		t.Error("Timeout waiting for SysDone")
	}
}
