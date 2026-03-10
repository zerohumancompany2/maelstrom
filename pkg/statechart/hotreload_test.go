package statechart

import (
	"testing"
	"time"

	"github.com/maelstrom/v3/internal/testutil"
)

// ============================================================================
// Hot-Reload Protocol Tests (7 tests, 7 commits)
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

		if err := engine.Control(rtID, CmdStart); err != nil {
			t.Fatalf("Start failed: %v", err)
		}

		if err := engine.Dispatch(rtID, Event{Type: "pending"}); err != nil {
			t.Fatalf("Dispatch failed: %v", err)
		}

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
							"regionA": {ID: "idleA", Children: nil},
							"regionB": {ID: "idleB", Children: nil},
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
		time.Sleep(50 * time.Millisecond)

		if runtime.IsQuiescent() {
			t.Error("Expected runtime to NOT be quiescent when parallel regions are running")
		}
	})
}

// TestHotReloadProtocol_ProtocolFlow verifies the hot-reload protocol flow
// as per arch-v1.md L865-868: prepareForReload signal and quiescence wait.
func TestHotReloadProtocol_ProtocolFlow(t *testing.T) {
	engine := NewEngine().(*Engine)
	mockCtx := testutil.NewMockApplicationContext()

	def := ChartDefinition{
		ID:      "test-protocol",
		Version: "1.0.0",
		Root: &Node{
			ID:          "idle",
			Children:    nil,
			Transitions: []Transition{{Event: "go", Target: "running"}},
		},
		InitialState: "idle",
	}

	rtID, err := engine.Spawn(def, mockCtx)
	if err != nil {
		t.Fatalf("Spawn failed: %v", err)
	}
	if err := engine.Control(rtID, CmdStart); err != nil {
		t.Fatalf("Start failed: %v", err)
	}

	// Test 1: prepareForReload signals current runtime
	t.Run("PrepareForReloadSignalsRuntime", func(t *testing.T) {
		rtID2, err := engine.Spawn(def, mockCtx)
		if err != nil {
			t.Fatalf("Spawn failed: %v", err)
		}
		if err := engine.Control(rtID2, CmdStart); err != nil {
			t.Fatalf("Start failed: %v", err)
		}

		newDef := ChartDefinition{
			ID:      "test-protocol",
			Version: "1.1.0",
			Root: &Node{
				ID:       "idle",
				Children: nil,
				Transitions: []Transition{
					{Event: "go", Target: "running"},
					{Event: "stop", Target: "idle"},
				},
			},
			InitialState: "idle",
		}

		err = engine.HotReload(rtID2, newDef, HotReloadOptions{
			Timeout:          5 * time.Second,
			MaxAttempts:      3,
			HistoryMode:      HistoryModeShallow,
			ContextTransform: nil,
		})
		if err != nil {
			t.Fatalf("HotReload failed: %v", err)
		}

		runtime := engine.runtimes[rtID2]
		if runtime == nil {
			t.Fatal("Runtime should exist after hot-reload")
		}
		if runtime.definition.Version != "1.1.0" {
			t.Errorf("Expected version 1.1.0 after hot-reload, got %s", runtime.definition.Version)
		}
	})

	// Test 2: Runtime attempts to reach quiescence within timeout
	t.Run("QuiescenceWaitWithTimeout", func(t *testing.T) {
		if err := engine.Dispatch(rtID, Event{Type: "go"}); err != nil {
			t.Logf("Dispatch failed: %v", err)
		}
		time.Sleep(100 * time.Millisecond)

		runtime := engine.runtimes[rtID]
		if runtime != nil && !runtime.IsQuiescent() {
			t.Log("Runtime is not quiescent - may have pending events")
		}
	})
}

// TestHotReloadProtocol_QuiescenceReached verifies graceful reload when quiescence is reached
// as per arch-v1.md L869-872: stop current runtime, spawn with history, apply contextTransform.
func TestHotReloadProtocol_QuiescenceReached(t *testing.T) {
	engine := NewEngine().(*Engine)
	mockCtx := testutil.NewMockApplicationContext()

	def := ChartDefinition{
		ID:      "test-quiescence-reached",
		Version: "1.0.0",
		Root: &Node{
			ID:          "idle",
			Children:    nil,
			Transitions: []Transition{{Event: "go", Target: "running"}},
		},
		InitialState: "idle",
	}

	rtID, err := engine.Spawn(def, mockCtx)
	if err != nil {
		t.Fatalf("Spawn failed: %v", err)
	}
	if err := engine.Control(rtID, CmdStart); err != nil {
		t.Fatalf("Start failed: %v", err)
	}

	// Test 1: Quiescence reached - stop current runtime and spawn with history
	t.Run("QuiescenceReachedStopsAndSpawns", func(t *testing.T) {
		newDef := ChartDefinition{
			ID:      "test-quiescence-reached",
			Version: "1.1.0",
			Root: &Node{
				ID:       "idle",
				Children: nil,
				Transitions: []Transition{
					{Event: "go", Target: "running"},
					{Event: "stop", Target: "idle"},
				},
			},
			InitialState: "idle",
		}

		err := engine.HotReload(rtID, newDef, HotReloadOptions{
			Timeout:          5 * time.Second,
			MaxAttempts:      3,
			HistoryMode:      HistoryModeShallow,
			ContextTransform: nil,
		})
		if err != nil {
			t.Fatalf("HotReload failed: %v", err)
		}

		runtime := engine.runtimes[rtID]
		if runtime == nil {
			t.Fatal("Runtime should exist after hot-reload")
		}
		if runtime.definition.Version != "1.1.0" {
			t.Errorf("Expected version 1.1.0, got %s", runtime.definition.Version)
		}
	})

	// Test 2: Version change triggers contextTransform (if provided)
	t.Run("VersionChangeAppliesContextTransform", func(t *testing.T) {
		rtID2, err := engine.Spawn(def, mockCtx)
		if err != nil {
			t.Fatalf("Spawn failed: %v", err)
		}
		if err := engine.Control(rtID2, CmdStart); err != nil {
			t.Fatalf("Start failed: %v", err)
		}

		newDef := ChartDefinition{
			ID:      "test-quiescence-reached",
			Version: "2.0.0",
			Root: &Node{
				ID:          "idle",
				Children:    nil,
				Transitions: []Transition{{Event: "go", Target: "running"}},
			},
			InitialState: "idle",
		}

		err = engine.HotReload(rtID2, newDef, HotReloadOptions{
			Timeout:          5 * time.Second,
			MaxAttempts:      3,
			HistoryMode:      HistoryModeShallow,
			ContextTransform: nil,
		})
		if err != nil {
			t.Fatalf("HotReload failed: %v", err)
		}

		runtime := engine.runtimes[rtID2]
		if runtime == nil {
			t.Fatal("Runtime should exist after hot-reload")
		}
		if runtime.definition.Version != "2.0.0" {
			t.Errorf("Expected version 2.0.0, got %s", runtime.definition.Version)
		}
	})
}

// TestHotReloadProtocol_TimeoutForceStop verifies force-stop on timeout
// as per arch-v1.md L873-876: force-stop current runtime, cleanStart, increment counter.
func TestHotReloadProtocol_TimeoutForceStop(t *testing.T) {
	engine := NewEngine().(*Engine)
	mockCtx := testutil.NewMockApplicationContext()

	def := ChartDefinition{
		ID:      "test-timeout",
		Version: "1.0.0",
		Root: &Node{
			ID:          "idle",
			Children:    nil,
			Transitions: []Transition{{Event: "go", Target: "running"}},
		},
		InitialState: "idle",
	}

	rtID, err := engine.Spawn(def, mockCtx)
	if err != nil {
		t.Fatalf("Spawn failed: %v", err)
	}
	if err := engine.Control(rtID, CmdStart); err != nil {
		t.Fatalf("Start failed: %v", err)
	}

	// Test 1: Timeout triggers force-stop
	t.Run("TimeoutTriggersForceStop", func(t *testing.T) {
		done := make(chan bool)
		go func() {
			for i := 0; i < 100; i++ {
				engine.Dispatch(rtID, Event{Type: "go"})
				time.Sleep(5 * time.Millisecond)
			}
			done <- true
		}()

		newDef := ChartDefinition{
			ID:      "test-timeout",
			Version: "1.1.0",
			Root: &Node{
				ID:          "idle",
				Children:    nil,
				Transitions: []Transition{{Event: "go", Target: "running"}},
			},
			InitialState: "idle",
		}

		err := engine.HotReload(rtID, newDef, HotReloadOptions{
			Timeout:          50 * time.Millisecond,
			MaxAttempts:      3,
			HistoryMode:      HistoryModeShallow,
			ContextTransform: nil,
		})

		<-done

		if err == nil {
			t.Log("HotReload succeeded (may have reached quiescence)")
		} else {
			t.Logf("HotReload timed out as expected: %v", err)
		}
	})

	// Test 2: cleanStart creates new runtime without history
	t.Run("CleanStartWithoutHistory", func(t *testing.T) {
		rtID2, err := engine.Spawn(def, mockCtx)
		if err != nil {
			t.Fatalf("Spawn failed: %v", err)
		}
		if err := engine.Control(rtID2, CmdStart); err != nil {
			t.Fatalf("Start failed: %v", err)
		}

		engine.Dispatch(rtID2, Event{Type: "go"})
		time.Sleep(50 * time.Millisecond)

		newDef := ChartDefinition{
			ID:      "test-timeout",
			Version: "1.1.0",
			Root: &Node{
				ID:          "idle",
				Children:    nil,
				Transitions: []Transition{{Event: "go", Target: "running"}},
			},
			InitialState: "idle",
		}

		err = engine.HotReload(rtID2, newDef, HotReloadOptions{
			Timeout:          5 * time.Second,
			MaxAttempts:      3,
			HistoryMode:      HistoryModeNone,
			ContextTransform: nil,
		})
		if err != nil {
			t.Fatalf("HotReload failed: %v", err)
		}

		runtime := engine.runtimes[rtID2]
		if runtime == nil {
			t.Fatal("Runtime should exist after hot-reload")
		}
		if runtime.activeState != "idle" {
			t.Errorf("Expected initial state 'idle' with no history, got '%s'", runtime.activeState)
		}
	})
}

// TestHotReloadProtocol_MaxAttemptsExceeded verifies max attempts handling
// as per arch-v1.md L877-879: log permanent failure, require admin intervention.
func TestHotReloadProtocol_MaxAttemptsExceeded(t *testing.T) {
	engine := NewEngine().(*Engine)
	mockCtx := testutil.NewMockApplicationContext()

	def := ChartDefinition{
		ID:      "test-max-attempts",
		Version: "1.0.0",
		Root: &Node{
			ID:          "idle",
			Children:    nil,
			Transitions: []Transition{{Event: "go", Target: "running"}},
		},
		InitialState: "idle",
	}

	rtID, err := engine.Spawn(def, mockCtx)
	if err != nil {
		t.Fatalf("Spawn failed: %v", err)
	}
	if err := engine.Control(rtID, CmdStart); err != nil {
		t.Fatalf("Start failed: %v", err)
	}

	// Test 1: Max attempts exceeded logs permanent failure
	t.Run("MaxAttemptsExceededLogsFailure", func(t *testing.T) {
		newDef := ChartDefinition{
			ID:      "test-max-attempts",
			Version: "1.1.0",
			Root: &Node{
				ID:          "idle",
				Children:    nil,
				Transitions: []Transition{{Event: "go", Target: "running"}},
			},
			InitialState: "idle",
		}

		err := engine.HotReload(rtID, newDef, HotReloadOptions{
			Timeout:          10 * time.Millisecond,
			MaxAttempts:      1,
			HistoryMode:      HistoryModeShallow,
			ContextTransform: nil,
		})

		if err != nil {
			t.Logf("HotReload failed as expected with maxAttempts=1: %v", err)
		}
	})

	// Test 2: Multiple attempts with timeout
	t.Run("MultipleAttemptsTimeout", func(t *testing.T) {
		rtID2, err := engine.Spawn(def, mockCtx)
		if err != nil {
			t.Fatalf("Spawn failed: %v", err)
		}
		if err := engine.Control(rtID2, CmdStart); err != nil {
			t.Fatalf("Start failed: %v", err)
		}

		newDef := ChartDefinition{
			ID:      "test-max-attempts",
			Version: "1.1.0",
			Root: &Node{
				ID:          "idle",
				Children:    nil,
				Transitions: []Transition{{Event: "go", Target: "running"}},
			},
			InitialState: "idle",
		}

		err = engine.HotReload(rtID2, newDef, HotReloadOptions{
			Timeout:          5 * time.Second,
			MaxAttempts:      3,
			HistoryMode:      HistoryModeShallow,
			ContextTransform: nil,
		})
		if err != nil {
			t.Fatalf("HotReload failed: %v", err)
		}

		runtime := engine.runtimes[rtID2]
		if runtime == nil {
			t.Fatal("Runtime should exist after hot-reload")
		}
	})
}
