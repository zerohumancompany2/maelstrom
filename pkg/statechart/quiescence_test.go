package statechart

import (
	"testing"
	"time"

	"github.com/maelstrom/v3/internal/testutil"
)

// TestChartRuntime_IsQuiescent_EmptyQueue verifies quiescence when no events pending.
func TestChartRuntime_IsQuiescent_EmptyQueue(t *testing.T) {
	def := ChartDefinition{
		ID:      "test-quiescent",
		Version: "1.0.0",
		Root: &Node{
			ID:       "idle",
			Children: nil,
		},
		InitialState: "idle",
	}

	mockCtx := testutil.NewMockApplicationContext()
	rt := &ChartRuntime{
		id:           "test-rt",
		definition:   def,
		state:        RuntimeStateRunning,
		activeState:  "idle",
		regionStates: make(map[string]string),
		eventQueue:   make([]Event, 0),
		appCtx:       mockCtx,
		actions:      make(map[string]ActionFn),
		guards:       make(map[string]GuardFn),
		isParallel:   false,
	}

	if !rt.IsQuiescent() {
		t.Error("Expected runtime to be quiescent with empty queue and no parallel regions")
	}
}

// TestChartRuntime_IsQuiescent_NonEmptyQueue verifies non-quiescence with pending events.
func TestChartRuntime_IsQuiescent_NonEmptyQueue(t *testing.T) {
	def := ChartDefinition{
		ID:      "test-quiescent",
		Version: "1.0.0",
		Root: &Node{
			ID:       "idle",
			Children: nil,
		},
		InitialState: "idle",
	}

	mockCtx := testutil.NewMockApplicationContext()
	rt := &ChartRuntime{
		id:           "test-rt",
		definition:   def,
		state:        RuntimeStateRunning,
		activeState:  "idle",
		regionStates: make(map[string]string),
		eventQueue:   []Event{{Type: "pending"}},
		appCtx:       mockCtx,
		actions:      make(map[string]ActionFn),
		guards:       make(map[string]GuardFn),
		isParallel:   false,
	}

	if rt.IsQuiescent() {
		t.Error("Expected runtime to NOT be quiescent with pending events in queue")
	}
}

// TestChartRuntime_IsQuiescent_ParallelRegionsProcessing verifies non-quiescence when parallel regions active.
func TestChartRuntime_IsQuiescent_ParallelRegionsProcessing(t *testing.T) {
	def := ChartDefinition{
		ID:      "test-parallel-quiescent",
		Version: "1.0.0",
		Root: &Node{
			ID: "root",
			Children: map[string]*Node{
				"parallel": {
					ID:          "parallel",
					RegionNames: []string{"regionA", "regionB"},
					Children: map[string]*Node{
						"regionA": {
							ID:       "idleA",
							Children: nil,
						},
						"regionB": {
							ID:       "idleB",
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

	// Start parallel state
	rt.enterParallelState("root/parallel")
	time.Sleep(50 * time.Millisecond)

	if rt.IsQuiescent() {
		t.Error("Expected runtime to NOT be quiescent when parallel regions are running")
	}
}

// TestChartRuntime_AwaitQuiescence waits for quiescence with timeout.
func TestChartRuntime_AwaitQuiescence(t *testing.T) {
	def := ChartDefinition{
		ID:      "test-await-quiescent",
		Version: "1.0.0",
		Root: &Node{
			ID:       "idle",
			Children: nil,
		},
		InitialState: "idle",
	}

	mockCtx := testutil.NewMockApplicationContext()
	rt := &ChartRuntime{
		id:           "test-rt",
		definition:   def,
		state:        RuntimeStateRunning,
		activeState:  "idle",
		regionStates: make(map[string]string),
		eventQueue:   make([]Event, 0),
		appCtx:       mockCtx,
		actions:      make(map[string]ActionFn),
		guards:       make(map[string]GuardFn),
		isParallel:   false,
	}

	// Should return immediately when already quiescent
	done := make(chan bool)
	go func() {
		err := rt.AwaitQuiescence(100 * time.Millisecond)
		done <- err == nil
	}()

	select {
	case success := <-done:
		if !success {
			t.Error("AwaitQuiescence should succeed when already quiescent")
		}
	case <-time.After(200 * time.Millisecond):
		t.Error("AwaitQuiescence should return immediately when quiescent")
	}
}
