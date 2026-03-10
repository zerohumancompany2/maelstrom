package statechart

import (
	"testing"
	"time"

	"github.com/maelstrom/v3/internal/testutil"
)

// ============================================================================
// Hot-Reload Protocol Tests
// ============================================================================

// TestHotReloadProtocol_QuiescenceDetection verifies quiescence detection
// as per arch-v1.md L858-861: empty queue, no active regions, no inflight tools.
func TestHotReloadProtocol_QuiescenceDetection(t *testing.T) {
	engine := NewEngine().(*Engine)
	mockCtx := testutil.NewMockApplicationContext()

	// Test 1: Quiescent when event queue empty
	t.Run("EmptyQueueIsQuiescent", func(t *testing.T) {
		def := ChartDefinition{
			ID:      "test-quiescent",
			Version: "1.0.0",
			Root: &Node{
				ID:       "idle",
				Children: nil,
			},
			InitialState: "idle",
		}

		rtID, err := engine.Spawn(def, mockCtx)
		if err != nil {
			t.Fatalf("Spawn failed: %v", err)
		}

		// Runtime in Created state should be quiescent (no events, no parallel)
		runtime := engine.runtimes[rtID]
		if !runtime.IsQuiescent() {
			t.Error("Expected runtime to be quiescent with empty queue and no parallel regions")
		}
	})

	// Test 2: Not quiescent when event queue has pending events
	t.Run("NonEmptyQueueNotQuiescent", func(t *testing.T) {
		def := ChartDefinition{
			ID:      "test-quiescent",
			Version: "1.0.0",
			Root: &Node{
				ID:       "idle",
				Children: nil,
			},
			InitialState: "idle",
		}

		rtID, err := engine.Spawn(def, mockCtx)
		if err != nil {
			t.Fatalf("Spawn failed: %v", err)
		}

		// Start runtime and dispatch event
		if err := engine.Control(rtID, CmdStart); err != nil {
			t.Fatalf("Start failed: %v", err)
		}

		if err := engine.Dispatch(rtID, Event{Type: "pending"}); err != nil {
			t.Fatalf("Dispatch failed: %v", err)
		}

		// Check immediately - should not be quiescent with event in queue
		runtime := engine.runtimes[rtID]
		if runtime.IsQuiescent() {
			t.Error("Expected runtime to NOT be quiescent with pending events in queue")
		}
	})

	// Test 3: Not quiescent when in parallel state with active regions
	t.Run("ParallelStateNotQuiescent", func(t *testing.T) {
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

		rtID, err := engine.Spawn(def, mockCtx)
		if err != nil {
			t.Fatalf("Spawn failed: %v", err)
		}

		runtime := engine.runtimes[rtID]
		runtime.enterParallelState("root/parallel")
		runtime.isParallel = true

		// Give router time to start
		time.Sleep(50 * time.Millisecond)

		if runtime.IsQuiescent() {
			t.Error("Expected runtime to NOT be quiescent when parallel regions are running")
		}
	})
}
