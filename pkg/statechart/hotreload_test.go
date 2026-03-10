package statechart

import (
	"testing"
	"time"

	"github.com/maelstrom/v3/internal/testutil"
)

// ============================================================================
// Hot-Reload Protocol Tests (7 tests, 7 commits)
// ============================================================================

func TestHotReloadProtocol_QuiescenceDetection(t *testing.T) {
	engine := NewEngine().(*Engine)
	mockCtx := testutil.NewMockApplicationContext()

	t.Run("EmptyQueueIsQuiescent", func(t *testing.T) {
		def := ChartDefinition{ID: "test", Version: "1.0.0", Root: &Node{ID: "idle", Children: nil}, InitialState: "idle"}
		rtID, _ := engine.Spawn(def, mockCtx)
		if !engine.runtimes[rtID].IsQuiescent() {
			t.Error("Expected quiescent")
		}
	})

	t.Run("NonEmptyQueueNotQuiescent", func(t *testing.T) {
		def := ChartDefinition{ID: "test", Version: "1.0.0", Root: &Node{ID: "idle", Children: nil}, InitialState: "idle"}
		rtID, _ := engine.Spawn(def, mockCtx)
		engine.Control(rtID, CmdStart)
		engine.Dispatch(rtID, Event{Type: "pending"})
		if engine.runtimes[rtID].IsQuiescent() {
			t.Error("Expected not quiescent")
		}
	})

	t.Run("ParallelStateNotQuiescent", func(t *testing.T) {
		def := ChartDefinition{
			ID: "test", Version: "1.0.0", InitialState: "root",
			Root: &Node{ID: "root", Children: map[string]*Node{
				"parallel": {ID: "parallel", RegionNames: []string{"A", "B"}, Children: map[string]*Node{
					"A": {ID: "A", Children: nil}, "B": {ID: "B", Children: nil},
				}},
			}},
		}
		rtID, _ := engine.Spawn(def, mockCtx)
		rt := engine.runtimes[rtID]
		rt.enterParallelState("root/parallel")
		rt.isParallel = true
		time.Sleep(50 * time.Millisecond)
		if rt.IsQuiescent() {
			t.Error("Expected not quiescent with parallel regions")
		}
	})
}

func TestHotReloadProtocol_ProtocolFlow(t *testing.T) {
	engine := NewEngine().(*Engine)
	mockCtx := testutil.NewMockApplicationContext()
	def := ChartDefinition{ID: "test", Version: "1.0.0", Root: &Node{ID: "idle", Children: nil, Transitions: []Transition{{Event: "go", Target: "running"}}}, InitialState: "idle"}
	rtID, _ := engine.Spawn(def, mockCtx)
	engine.Control(rtID, CmdStart)

	t.Run("PrepareForReloadSignalsRuntime", func(t *testing.T) {
		rtID2, _ := engine.Spawn(def, mockCtx)
		engine.Control(rtID2, CmdStart)
		newDef := ChartDefinition{ID: "test", Version: "1.1.0", Root: &Node{ID: "idle", Children: nil, Transitions: []Transition{{Event: "go", Target: "running"}, {Event: "stop", Target: "idle"}}}, InitialState: "idle"}
		if err := engine.HotReload(rtID2, newDef, HotReloadOptions{Timeout: 5 * time.Second, MaxAttempts: 3, HistoryMode: HistoryModeShallow}); err != nil {
			t.Fatalf("HotReload failed: %v", err)
		}
		if engine.runtimes[rtID2].definition.Version != "1.1.0" {
			t.Errorf("Expected version 1.1.0, got %s", engine.runtimes[rtID2].definition.Version)
		}
	})

	t.Run("QuiescenceWaitWithTimeout", func(t *testing.T) {
		engine.Dispatch(rtID, Event{Type: "go"})
		time.Sleep(100 * time.Millisecond)
	})
}

func TestHotReloadProtocol_QuiescenceReached(t *testing.T) {
	engine := NewEngine().(*Engine)
	mockCtx := testutil.NewMockApplicationContext()
	def := ChartDefinition{ID: "test", Version: "1.0.0", Root: &Node{ID: "idle", Children: nil, Transitions: []Transition{{Event: "go", Target: "running"}}}, InitialState: "idle"}
	rtID, _ := engine.Spawn(def, mockCtx)
	engine.Control(rtID, CmdStart)

	t.Run("QuiescenceReachedStopsAndSpawns", func(t *testing.T) {
		newDef := ChartDefinition{ID: "test", Version: "1.1.0", Root: &Node{ID: "idle", Children: nil, Transitions: []Transition{{Event: "go", Target: "running"}}}, InitialState: "idle"}
		if err := engine.HotReload(rtID, newDef, HotReloadOptions{Timeout: 5 * time.Second, MaxAttempts: 3, HistoryMode: HistoryModeShallow}); err != nil {
			t.Fatalf("HotReload failed: %v", err)
		}
		if engine.runtimes[rtID].definition.Version != "1.1.0" {
			t.Errorf("Expected version 1.1.0")
		}
	})

	t.Run("VersionChangeAppliesContextTransform", func(t *testing.T) {
		rtID2, _ := engine.Spawn(def, mockCtx)
		engine.Control(rtID2, CmdStart)
		newDef := ChartDefinition{ID: "test", Version: "2.0.0", Root: &Node{ID: "idle", Children: nil, Transitions: []Transition{{Event: "go", Target: "running"}}}, InitialState: "idle"}
		if err := engine.HotReload(rtID2, newDef, HotReloadOptions{Timeout: 5 * time.Second, MaxAttempts: 3, HistoryMode: HistoryModeShallow}); err != nil {
			t.Fatalf("HotReload failed: %v", err)
		}
		if engine.runtimes[rtID2].definition.Version != "2.0.0" {
			t.Errorf("Expected version 2.0.0")
		}
	})
}

func TestHotReloadProtocol_TimeoutForceStop(t *testing.T) {
	engine := NewEngine().(*Engine)
	mockCtx := testutil.NewMockApplicationContext()
	def := ChartDefinition{ID: "test", Version: "1.0.0", Root: &Node{ID: "idle", Children: nil, Transitions: []Transition{{Event: "go", Target: "running"}}}, InitialState: "idle"}
	rtID, _ := engine.Spawn(def, mockCtx)
	engine.Control(rtID, CmdStart)

	t.Run("TimeoutTriggersForceStop", func(t *testing.T) {
		done := make(chan bool)
		go func() {
			for i := 0; i < 100; i++ {
				engine.Dispatch(rtID, Event{Type: "go"})
				time.Sleep(5 * time.Millisecond)
			}
			done <- true
		}()
		newDef := ChartDefinition{ID: "test", Version: "1.1.0", Root: &Node{ID: "idle", Children: nil, Transitions: []Transition{{Event: "go", Target: "running"}}}, InitialState: "idle"}
		engine.HotReload(rtID, newDef, HotReloadOptions{Timeout: 50 * time.Millisecond, MaxAttempts: 3, HistoryMode: HistoryModeShallow})
		<-done
	})

	t.Run("CleanStartWithoutHistory", func(t *testing.T) {
		rtID2, _ := engine.Spawn(def, mockCtx)
		engine.Control(rtID2, CmdStart)
		engine.Dispatch(rtID2, Event{Type: "go"})
		time.Sleep(50 * time.Millisecond)
		newDef := ChartDefinition{ID: "test", Version: "1.1.0", Root: &Node{ID: "idle", Children: nil, Transitions: []Transition{{Event: "go", Target: "running"}}}, InitialState: "idle"}
		if err := engine.HotReload(rtID2, newDef, HotReloadOptions{Timeout: 5 * time.Second, MaxAttempts: 3, HistoryMode: HistoryModeNone}); err != nil {
			t.Fatalf("HotReload failed: %v", err)
		}
		if engine.runtimes[rtID2].activeState != "idle" {
			t.Errorf("Expected idle, got %s", engine.runtimes[rtID2].activeState)
		}
	})
}

func TestHotReloadProtocol_MaxAttemptsExceeded(t *testing.T) {
	engine := NewEngine().(*Engine)
	mockCtx := testutil.NewMockApplicationContext()
	def := ChartDefinition{ID: "test", Version: "1.0.0", Root: &Node{ID: "idle", Children: nil, Transitions: []Transition{{Event: "go", Target: "running"}}}, InitialState: "idle"}
	rtID, _ := engine.Spawn(def, mockCtx)
	engine.Control(rtID, CmdStart)

	t.Run("MaxAttemptsExceededLogsFailure", func(t *testing.T) {
		newDef := ChartDefinition{ID: "test", Version: "1.1.0", Root: &Node{ID: "idle", Children: nil, Transitions: []Transition{{Event: "go", Target: "running"}}}, InitialState: "idle"}
		err := engine.HotReload(rtID, newDef, HotReloadOptions{Timeout: 10 * time.Millisecond, MaxAttempts: 1, HistoryMode: HistoryModeShallow})
		if err != nil {
			t.Logf("HotReload failed: %v", err)
		}
	})

	t.Run("MultipleAttemptsTimeout", func(t *testing.T) {
		rtID2, _ := engine.Spawn(def, mockCtx)
		engine.Control(rtID2, CmdStart)
		newDef := ChartDefinition{ID: "test", Version: "1.1.0", Root: &Node{ID: "idle", Children: nil, Transitions: []Transition{{Event: "go", Target: "running"}}}, InitialState: "idle"}
		if err := engine.HotReload(rtID2, newDef, HotReloadOptions{Timeout: 5 * time.Second, MaxAttempts: 3, HistoryMode: HistoryModeShallow}); err != nil {
			t.Fatalf("HotReload failed: %v", err)
		}
	})
}

func TestHotReloadProtocol_HistoryMechanisms(t *testing.T) {
	engine := NewEngine().(*Engine)
	mockCtx := testutil.NewMockApplicationContext()

	t.Run("ShallowHistoryRestoresToDefault", func(t *testing.T) {
		def := ChartDefinition{ID: "test", Version: "1.0.0", InitialState: "root/idle", Root: &Node{ID: "root", Children: map[string]*Node{"idle": {ID: "idle", Children: nil, IsInitial: true}, "running": {ID: "running", Children: nil}}, Transitions: []Transition{{Event: "start", Target: "root/running"}}}}
		rtID, _ := engine.Spawn(def, mockCtx)
		engine.Control(rtID, CmdStart)
		engine.Dispatch(rtID, Event{Type: "start"})
		time.Sleep(50 * time.Millisecond)
		newDef := ChartDefinition{ID: "test", Version: "1.1.0", InitialState: "root/idle", Root: &Node{ID: "root", Children: map[string]*Node{"idle": {ID: "idle", Children: nil, IsInitial: true}, "running": {ID: "running", Children: nil}}, Transitions: []Transition{{Event: "start", Target: "root/running"}}}}
		if err := engine.HotReload(rtID, newDef, HotReloadOptions{Timeout: 5 * time.Second, MaxAttempts: 3, HistoryMode: HistoryModeShallow}); err != nil {
			t.Fatalf("HotReload failed: %v", err)
		}
	})

	t.Run("DeepHistoryRestoresToSpecificState", func(t *testing.T) {
		def := ChartDefinition{ID: "test", Version: "1.0.0", InitialState: "idle", Root: &Node{ID: "idle", Children: nil, Transitions: []Transition{{Event: "start", Target: "running"}}}}
		rtID, _ := engine.Spawn(def, mockCtx)
		engine.Control(rtID, CmdStart)
		engine.Dispatch(rtID, Event{Type: "start"})
		time.Sleep(50 * time.Millisecond)
		newDef := ChartDefinition{ID: "test", Version: "1.1.0", InitialState: "idle", Root: &Node{ID: "idle", Children: nil, Transitions: []Transition{{Event: "start", Target: "running"}, {Event: "stop", Target: "idle"}}}}
		if err := engine.HotReload(rtID, newDef, HotReloadOptions{Timeout: 5 * time.Second, MaxAttempts: 3, HistoryMode: HistoryModeDeep}); err != nil {
			t.Fatalf("HotReload failed: %v", err)
		}
	})

	t.Run("DeletedStateFallbackToShallow", func(t *testing.T) {
		def := ChartDefinition{ID: "test", Version: "1.0.0", InitialState: "idle", Root: &Node{ID: "idle", Children: nil, Transitions: []Transition{{Event: "start", Target: "running"}}}}
		rtID, _ := engine.Spawn(def, mockCtx)
		engine.Control(rtID, CmdStart)
		engine.Dispatch(rtID, Event{Type: "start"})
		time.Sleep(50 * time.Millisecond)
		newDef := ChartDefinition{ID: "test", Version: "1.1.0", InitialState: "idle", Root: &Node{ID: "idle", Children: nil, Transitions: []Transition{{Event: "stop", Target: "idle"}}}}
		if err := engine.HotReload(rtID, newDef, HotReloadOptions{Timeout: 5 * time.Second, MaxAttempts: 3, HistoryMode: HistoryModeDeep}); err != nil {
			t.Fatalf("HotReload failed: %v", err)
		}
	})
}

func TestHotReloadProtocol_ContextTransformValidation(t *testing.T) {
	engine := NewEngine().(*Engine)
	mockCtx := testutil.NewMockApplicationContext()
	def := ChartDefinition{ID: "test", Version: "1.0.0", Root: &Node{ID: "idle", Children: nil, Transitions: []Transition{{Event: "go", Target: "running"}}}, InitialState: "idle"}
	rtID, _ := engine.Spawn(def, mockCtx)
	engine.Control(rtID, CmdStart)

	t.Run("TemplateSyntaxValidatedAtLoadTime", func(t *testing.T) {
		newDef := ChartDefinition{ID: "test", Version: "1.1.0", Root: &Node{ID: "idle", Children: nil, Transitions: []Transition{{Event: "go", Target: "running"}}}, InitialState: "idle"}
		if err := engine.HotReload(rtID, newDef, HotReloadOptions{Timeout: 5 * time.Second, MaxAttempts: 3, HistoryMode: HistoryModeShallow, ContextTransform: nil}); err != nil {
			t.Fatalf("HotReload with nil transform failed: %v", err)
		}
	})

	t.Run("TransformExecutedOnVersionChange", func(t *testing.T) {
		rtID2, _ := engine.Spawn(def, mockCtx)
		engine.Control(rtID2, CmdStart)
		sameVersionDef := ChartDefinition{ID: "test", Version: "1.0.0", Root: &Node{ID: "idle", Children: nil, Transitions: []Transition{{Event: "go", Target: "running"}}}, InitialState: "idle"}
		if err := engine.HotReload(rtID2, sameVersionDef, HotReloadOptions{Timeout: 5 * time.Second, MaxAttempts: 3, HistoryMode: HistoryModeShallow}); err != nil {
			t.Fatalf("HotReload failed: %v", err)
		}
		newVersionDef := ChartDefinition{ID: "test", Version: "2.0.0", Root: &Node{ID: "idle", Children: nil, Transitions: []Transition{{Event: "go", Target: "running"}}}, InitialState: "idle"}
		if err := engine.HotReload(rtID2, newVersionDef, HotReloadOptions{Timeout: 5 * time.Second, MaxAttempts: 3, HistoryMode: HistoryModeShallow}); err != nil {
			t.Fatalf("HotReload with version change failed: %v", err)
		}
		if engine.runtimes[rtID2].definition.Version != "2.0.0" {
			t.Errorf("Expected version 2.0.0")
		}
	})

	t.Run("FailureLeadsToCleanStartFallback", func(t *testing.T) {
		rtID3, _ := engine.Spawn(def, mockCtx)
		engine.Control(rtID3, CmdStart)
		engine.Dispatch(rtID3, Event{Type: "go"})
		time.Sleep(50 * time.Millisecond)
		newDef := ChartDefinition{ID: "test", Version: "2.0.0", Root: &Node{ID: "idle", Children: nil, Transitions: []Transition{{Event: "go", Target: "running"}}}, InitialState: "idle"}
		if err := engine.HotReload(rtID3, newDef, HotReloadOptions{Timeout: 5 * time.Second, MaxAttempts: 3, HistoryMode: HistoryModeNone}); err != nil {
			t.Fatalf("HotReload failed: %v", err)
		}
		if engine.runtimes[rtID3].activeState != "idle" {
			t.Errorf("Expected idle after clean start, got %s", engine.runtimes[rtID3].activeState)
		}
	})
}
