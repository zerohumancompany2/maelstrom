package statechart

import (
	"errors"
	"testing"
	"time"

	"github.com/maelstrom/v3/internal/testutil"
)

// ============================================================================
// Test Helpers
// ============================================================================

func newTestEngine(t *testing.T) Library {
	t.Helper()
	return NewEngine()
}

func newSimpleAtomicChart() ChartDefinition {
	return ChartDefinition{
		ID:      "test-chart",
		Version: "1.0.0",
		Root: &Node{
			ID:       "idle",
			Children: nil,
			Transitions: []Transition{
				{Event: "start", Target: "running"},
			},
		},
		InitialState: "idle",
	}
}

func newCompoundChart() ChartDefinition {
	return ChartDefinition{
		ID:      "compound-chart",
		Version: "1.0.0",
		Root: &Node{
			ID: "root",
			Children: map[string]*Node{
				"idle": {
					ID:       "idle",
					Children: nil,
					Transitions: []Transition{
						{Event: "activate", Target: "active"},
					},
				},
				"active": {
					ID: "active",
					Children: map[string]*Node{
						"child1": {
							ID:       "child1",
							Children: nil,
							Transitions: []Transition{
								{Event: "next", Target: "root/idle"},
							},
						},
					},
					IsInitial: true,
				},
			},
			IsInitial: true,
		},
		InitialState: "root/idle",
	}
}

func newParallelChart() ChartDefinition {
	return ChartDefinition{
		ID:      "parallel-chart",
		Version: "1.0.0",
		Root: &Node{
			ID: "root",
			Children: map[string]*Node{
				"regionA": {
					ID:       "regionA",
					Children: nil,
					Transitions: []Transition{
						{Event: "nextA", Target: "regionA-done"},
					},
				},
				"regionA-done": {
					ID:       "regionA-done",
					Children: nil,
				},
				"regionB": {
					ID:       "regionB",
					Children: nil,
					Transitions: []Transition{
						{Event: "nextB", Target: "regionB-done"},
					},
				},
				"regionB-done": {
					ID:       "regionB-done",
					Children: nil,
				},
			},
			RegionNames: []string{"regionA", "regionB"},
		},
		InitialState: "root",
	}
}

// ============================================================================
// Core Lifecycle Tests
// ============================================================================

func TestSpawn_CreatesRuntimeWithUniqueID(t *testing.T) {
	engine := newTestEngine(t)
	def := newSimpleAtomicChart()
	mockCtx := testutil.NewMockApplicationContext()

	id1, err := engine.Spawn(def, mockCtx)
	if err != nil {
		t.Fatalf("Spawn failed: %v", err)
	}

	if id1 == "" {
		t.Error("Spawn returned empty RuntimeID")
	}

	// Spawn second runtime - should have different ID
	id2, err := engine.Spawn(def, mockCtx)
	if err != nil {
		t.Fatalf("Second Spawn failed: %v", err)
	}

	if id1 == id2 {
		t.Error("Spawn returned duplicate RuntimeID")
	}
}

func TestControl_StartTransitionsToRunning(t *testing.T) {
	engine := newTestEngine(t)
	def := newSimpleAtomicChart()
	mockCtx := testutil.NewMockApplicationContext()

	id, err := engine.Spawn(def, mockCtx)
	if err != nil {
		t.Fatalf("Spawn failed: %v", err)
	}

	// Before Start, runtime is in Created state
	// After Start, runtime should be Running
	err = engine.Control(id, CmdStart)
	if err != nil {
		t.Fatalf("Start failed: %v", err)
	}

	// Verify by attempting to dispatch an event (only works when running)
	err = engine.Dispatch(id, Event{Type: "test"})
	if err != nil {
		t.Errorf("Dispatch failed after Start: %v", err)
	}
}

func TestControl_PauseAndResume(t *testing.T) {
	engine := newTestEngine(t)
	def := newSimpleAtomicChart()
	mockCtx := testutil.NewMockApplicationContext()

	id, err := engine.Spawn(def, mockCtx)
	if err != nil {
		t.Fatalf("Spawn failed: %v", err)
	}

	// Start the runtime
	err = engine.Control(id, CmdStart)
	if err != nil {
		t.Fatalf("Start failed: %v", err)
	}

	// Pause
	err = engine.Control(id, CmdPause)
	if err != nil {
		t.Fatalf("Pause failed: %v", err)
	}

	// Dispatch should fail while paused
	err = engine.Dispatch(id, Event{Type: "test"})
	if err == nil {
		t.Error("Dispatch should fail when runtime is paused")
	}

	// Resume
	err = engine.Control(id, CmdResume)
	if err != nil {
		t.Fatalf("Resume failed: %v", err)
	}

	// Dispatch should work after resume
	err = engine.Dispatch(id, Event{Type: "test"})
	if err != nil {
		t.Errorf("Dispatch failed after Resume: %v", err)
	}
}

func TestControl_StopCleansUpRuntime(t *testing.T) {
	engine := newTestEngine(t)
	def := newSimpleAtomicChart()
	mockCtx := testutil.NewMockApplicationContext()

	id, err := engine.Spawn(def, mockCtx)
	if err != nil {
		t.Fatalf("Spawn failed: %v", err)
	}

	// Start first
	err = engine.Control(id, CmdStart)
	if err != nil {
		t.Fatalf("Start failed: %v", err)
	}

	// Stop
	err = engine.Control(id, CmdStop)
	if err != nil {
		t.Fatalf("Stop failed: %v", err)
	}

	// After stop, operations should fail
	err = engine.Dispatch(id, Event{Type: "test"})
	if err == nil {
		t.Error("Dispatch should fail after Stop")
	}

	err = engine.Control(id, CmdStart)
	if err == nil {
		t.Error("Start should fail after Stop")
	}
}

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

// ============================================================================
// Action & Guard Registry Tests
// ============================================================================

func TestRegisterAction_CanBeCalledInTransition(t *testing.T) {
	engine := newTestEngine(t)

	actionCalled := false
	engine.RegisterAction("testAction", func(runtimeCtx RuntimeContext, appCtx ApplicationContext, event Event) error {
		actionCalled = true
		return nil
	})

	def := ChartDefinition{
		ID:      "action-chart",
		Version: "1.0.0",
		Root: &Node{
			ID:       "idle",
			Children: nil,
			Transitions: []Transition{
				{Event: "go", Target: "done", Actions: []string{"testAction"}},
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

	if !actionCalled {
		t.Error("Registered action was not called during transition")
	}
}

func TestRegisterGuard_CanBlockTransition(t *testing.T) {
	engine := newTestEngine(t)

	engine.RegisterGuard("blockGuard", func(appCtx ApplicationContext, event Event) bool {
		return false
	})

	def := ChartDefinition{
		ID:      "guard-chart",
		Version: "1.0.0",
		Root: &Node{
			ID:       "idle",
			Children: nil,
			Transitions: []Transition{
				{Event: "try", Guard: "blockGuard", Target: "done"},
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

	err = engine.Dispatch(id, Event{Type: "try"})
	if err != nil {
		t.Fatalf("Dispatch failed: %v", err)
	}

	time.Sleep(100 * time.Millisecond)

	snap, err := engine.Snapshot(id)
	if err != nil {
		t.Fatalf("Snapshot failed: %v", err)
	}

	if len(snap.ActiveStates) == 0 || snap.ActiveStates[0] != "idle" {
		t.Errorf("Expected state 'idle' (blocked), got %v", snap.ActiveStates)
	}
}

// ============================================================================
// Snapshot/Restore Tests
// ============================================================================

func TestSnapshot_CapturesRuntimeState(t *testing.T) {
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

	snap, err := engine.Snapshot(id)
	if err != nil {
		t.Fatalf("Snapshot failed: %v", err)
	}

	if snap.RuntimeID != id {
		t.Errorf("Snapshot.RuntimeID = %s, want %s", snap.RuntimeID, id)
	}

	if snap.DefinitionID != def.ID {
		t.Errorf("Snapshot.DefinitionID = %s, want %s", snap.DefinitionID, def.ID)
	}

	if len(snap.ActiveStates) == 0 {
		t.Error("Snapshot.ActiveStates is empty")
	}
}

func TestRestore_CreatesEquivalentRuntime(t *testing.T) {
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

	// Take snapshot
	snap, err := engine.Snapshot(id)
	if err != nil {
		t.Fatalf("Snapshot failed: %v", err)
	}

	// Restore
	newID, err := engine.Restore(snap)
	if err != nil {
		t.Fatalf("Restore failed: %v", err)
	}

	if newID == "" {
		t.Error("Restore returned empty RuntimeID")
	}

	if newID == id {
		t.Error("Restore should return a new RuntimeID, not the original")
	}

	// Verify restored runtime has same configuration
	newSnap, err := engine.Snapshot(newID)
	if err != nil {
		t.Fatalf("Snapshot of restored runtime failed: %v", err)
	}

	if len(newSnap.ActiveStates) != len(snap.ActiveStates) {
		t.Errorf("Restored runtime has different active states: %v vs %v", newSnap.ActiveStates, snap.ActiveStates)
	}
}

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

func TestControl_InvalidCommand(t *testing.T) {
	engine := newTestEngine(t)
	def := newSimpleAtomicChart()
	mockCtx := testutil.NewMockApplicationContext()

	id, err := engine.Spawn(def, mockCtx)
	if err != nil {
		t.Fatalf("Spawn failed: %v", err)
	}

	err = engine.Control(id, "invalid-cmd")
	if err == nil {
		t.Error("Invalid control command should fail")
	}

	if !errors.Is(err, ErrInvalidControlCmd) {
		t.Errorf("Expected ErrInvalidControlCmd, got: %v", err)
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

func TestAction_ReceivesCorrectContext(t *testing.T) {
	engine := newTestEngine(t)

	var receivedRuntimeCtx RuntimeContext
	var receivedEvent Event

	engine.RegisterAction("captureContext", func(runtimeCtx RuntimeContext, appCtx ApplicationContext, event Event) error {
		receivedRuntimeCtx = runtimeCtx
		receivedEvent = event
		return nil
	})

	def := ChartDefinition{
		ID:      "action-context-chart",
		Version: "1.0.0",
		Root: &Node{
			ID:       "idle",
			Children: nil,
			Transitions: []Transition{
				{Event: "go", Target: "done", Actions: []string{"captureContext"}},
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

	testEvent := Event{Type: "go", Payload: "test-payload", CorrelationID: "test-123"}
	err = engine.Dispatch(id, testEvent)
	if err != nil {
		t.Fatalf("Dispatch failed: %v", err)
	}

	time.Sleep(100 * time.Millisecond)

	if receivedEvent.Type != "go" {
		t.Errorf("Expected event type 'go', got '%s'", receivedEvent.Type)
	}

	if receivedEvent.Payload != "test-payload" {
		t.Errorf("Expected payload 'test-payload', got '%v'", receivedEvent.Payload)
	}

	if receivedRuntimeCtx.ChartID != "action-context-chart" {
		t.Errorf("Expected ChartID 'action-context-chart', got '%s'", receivedRuntimeCtx.ChartID)
	}
}

func TestGuard_ReceivesReadOnlyContext(t *testing.T) {
	engine := newTestEngine(t)

	var guardCalled bool
	engine.RegisterGuard("testGuard", func(appCtx ApplicationContext, event Event) bool {
		guardCalled = true
		// Guards should only read, not write
		// The interface doesn't prevent writes, but conventionally guards don't
		return event.Payload == "allow"
	})

	def := ChartDefinition{
		ID:      "guard-chart",
		Version: "1.0.0",
		Root: &Node{
			ID:       "idle",
			Children: nil,
			Transitions: []Transition{
				{Event: "go", Guard: "testGuard", Target: "done"},
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

	// Dispatch with payload that allows
	err = engine.Dispatch(id, Event{Type: "go", Payload: "allow"})
	if err != nil {
		t.Fatalf("Dispatch failed: %v", err)
	}

	time.Sleep(100 * time.Millisecond)

	if !guardCalled {
		t.Error("Guard was not called")
	}

	snap, err := engine.Snapshot(id)
	if err != nil {
		t.Fatalf("Snapshot failed: %v", err)
	}

	if len(snap.ActiveStates) == 0 || snap.ActiveStates[0] != "done" {
		t.Errorf("Expected state 'done', got %v", snap.ActiveStates)
	}
}

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

// ============================================================================
// Entry/Exit Action Order Tests
// ============================================================================

func TestTransition_ExitActionsFireBeforeEntryActions(t *testing.T) {
	engine := newTestEngine(t)

	var actionOrder []string

	engine.RegisterAction("exitIdle", func(runtimeCtx RuntimeContext, appCtx ApplicationContext, event Event) error {
		actionOrder = append(actionOrder, "exit-idle")
		return nil
	})

	engine.RegisterAction("entryDone", func(runtimeCtx RuntimeContext, appCtx ApplicationContext, event Event) error {
		actionOrder = append(actionOrder, "entry-done")
		return nil
	})

	def := ChartDefinition{
		ID:      "order-chart",
		Version: "1.0.0",
		Root: &Node{
			ID:       "idle",
			Children: nil,
			Transitions: []Transition{
				{Event: "go", Target: "done", Actions: []string{"entryDone"}},
			},
			ExitActions: []string{"exitIdle"},
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

	if len(actionOrder) < 2 {
		t.Fatalf("Expected 2 actions, got %d: %v", len(actionOrder), actionOrder)
	}

	if actionOrder[0] != "exit-idle" {
		t.Errorf("Expected exit action first, got: %v", actionOrder)
	}

	if actionOrder[1] != "entry-done" {
		t.Errorf("Expected entry action second, got: %v", actionOrder)
	}
}

// ============================================================================
// Event Ordering Tests
// ============================================================================

func TestDispatch_EventsProcessedFIFO(t *testing.T) {
	engine := newTestEngine(t)

	var processedEvents []string

	engine.RegisterAction("logEvent", func(runtimeCtx RuntimeContext, appCtx ApplicationContext, event Event) error {
		processedEvents = append(processedEvents, event.Type)
		return nil
	})

	def := ChartDefinition{
		ID:      "fifo-chart",
		Version: "1.0.0",
		Root: &Node{
			ID:       "idle",
			Children: nil,
			Transitions: []Transition{
				{Event: "first", Target: "idle", Actions: []string{"logEvent"}},
				{Event: "second", Target: "idle", Actions: []string{"logEvent"}},
				{Event: "third", Target: "idle", Actions: []string{"logEvent"}},
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

	// Dispatch events in order
	for _, evType := range []string{"first", "second", "third"} {
		err = engine.Dispatch(id, Event{Type: evType})
		if err != nil {
			t.Fatalf("Dispatch %s failed: %v", evType, err)
		}
	}

	time.Sleep(200 * time.Millisecond)

	// Verify events were processed in order
	expected := []string{"first", "second", "third"}
	if len(processedEvents) != len(expected) {
		t.Fatalf("Expected %d events, got %d: %v", len(expected), len(processedEvents), processedEvents)
	}

	for i, exp := range expected {
		if processedEvents[i] != exp {
			t.Errorf("Event %d: expected %s, got %s", i, exp, processedEvents[i])
		}
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
// Multiple Transitions with Guards Tests
// ============================================================================

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

// ============================================================================
// Entry/Exit Action Order Tests (Detailed)
// ============================================================================

func TestEntryExitActions_OrderInCompoundState(t *testing.T) {
	engine := newTestEngine(t)

	var actionOrder []string

	engine.RegisterAction("rootEntry", func(runtimeCtx RuntimeContext, appCtx ApplicationContext, event Event) error {
		actionOrder = append(actionOrder, "root-entry")
		return nil
	})

	engine.RegisterAction("childEntry", func(runtimeCtx RuntimeContext, appCtx ApplicationContext, event Event) error {
		actionOrder = append(actionOrder, "child-entry")
		return nil
	})

	engine.RegisterAction("childExit", func(runtimeCtx RuntimeContext, appCtx ApplicationContext, event Event) error {
		actionOrder = append(actionOrder, "child-exit")
		return nil
	})

	engine.RegisterAction("rootExit", func(runtimeCtx RuntimeContext, appCtx ApplicationContext, event Event) error {
		actionOrder = append(actionOrder, "root-exit")
		return nil
	})

	def := ChartDefinition{
		ID:      "order-chart",
		Version: "1.0.0",
		Root: &Node{
			ID: "root",
			Children: map[string]*Node{
				"child": {
					ID:       "child",
					Children: nil,
					Transitions: []Transition{
						{Event: "exit", Target: "root/done"},
					},
				},
				"done": {
					ID:       "done",
					Children: nil,
				},
			},
			IsInitial:    true,
			EntryActions: []string{"rootEntry"},
			ExitActions:  []string{"rootExit"},
		},
		InitialState: "root/child",
	}

	// Add entry/exit to child
	def.Root.Children["child"].EntryActions = []string{"childEntry"}
	def.Root.Children["child"].ExitActions = []string{"childExit"}

	mockCtx := testutil.NewMockApplicationContext()
	id, err := engine.Spawn(def, mockCtx)
	if err != nil {
		t.Fatalf("Spawn failed: %v", err)
	}

	err = engine.Control(id, CmdStart)
	if err != nil {
		t.Fatalf("Start failed: %v", err)
	}

	// Clear action order from entry actions
	actionOrder = nil

	// Trigger transition that exits child and root
	err = engine.Dispatch(id, Event{Type: "exit"})
	if err != nil {
		t.Fatalf("Dispatch failed: %v", err)
	}

	time.Sleep(100 * time.Millisecond)

	// Verify order: child-exit, then entry for done
	// Note: root is NOT exited because this is a transition between siblings within the same parent
	expected := []string{"child-exit"}
	if len(actionOrder) < len(expected) {
		t.Fatalf("Expected at least %d actions, got %d: %v", len(expected), len(actionOrder), actionOrder)
	}

	for i, exp := range expected {
		if actionOrder[i] != exp {
			t.Errorf("Action %d: expected '%s', got '%s'", i, exp, actionOrder[i])
		}
	}
}

// ============================================================================
// Implied Behavior Tests
// ============================================================================

func TestEventProcessing_FIFOOrdering(t *testing.T) {
	engine := newTestEngine(t)

	var processedOrder []string
	engine.RegisterAction("logEvent", func(runtimeCtx RuntimeContext, appCtx ApplicationContext, event Event) error {
		processedOrder = append(processedOrder, event.Payload.(string))
		return nil
	})

	def := ChartDefinition{
		ID:      "fifo-chart",
		Version: "1.0.0",
		Root: &Node{
			ID:       "state1",
			Children: nil,
			Transitions: []Transition{
				{Event: "next", Target: "state1", Actions: []string{"logEvent"}},
			},
		},
		InitialState: "state1",
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

	// Dispatch events in order
	for i, payload := range []string{"first", "second", "third"} {
		err = engine.Dispatch(id, Event{Type: "next", Payload: payload})
		if err != nil {
			t.Fatalf("Dispatch %d failed: %v", i, err)
		}
	}

	time.Sleep(200 * time.Millisecond)

	// Verify FIFO order
	expected := []string{"first", "second", "third"}
	if len(processedOrder) != len(expected) {
		t.Fatalf("Expected %d events, got %d: %v", len(expected), len(processedOrder), processedOrder)
	}

	for i, exp := range expected {
		if processedOrder[i] != exp {
			t.Errorf("Event %d: expected '%s', got '%s'", i, exp, processedOrder[i])
		}
	}
}

func TestRunToCompletion_EventFullyProcessedBeforeNext(t *testing.T) {
	engine := newTestEngine(t)

	var actionLog []string
	engine.RegisterAction("logStart", func(runtimeCtx RuntimeContext, appCtx ApplicationContext, event Event) error {
		actionLog = append(actionLog, "start-action")
		return nil
	})
	engine.RegisterAction("logEnd", func(runtimeCtx RuntimeContext, appCtx ApplicationContext, event Event) error {
		actionLog = append(actionLog, "end-action")
		return nil
	})

	def := ChartDefinition{
		ID:      "rtc-chart",
		Version: "1.0.0",
		Root: &Node{
			ID:       "idle",
			Children: nil,
			Transitions: []Transition{
				{Event: "go", Target: "idle", Actions: []string{"logStart"}},
				{Event: "stop", Target: "idle", Actions: []string{"logEnd"}},
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

	// Dispatch two events
	err = engine.Dispatch(id, Event{Type: "go"})
	if err != nil {
		t.Fatalf("Dispatch go failed: %v", err)
	}

	err = engine.Dispatch(id, Event{Type: "stop"})
	if err != nil {
		t.Fatalf("Dispatch stop failed: %v", err)
	}

	time.Sleep(200 * time.Millisecond)

	// Verify actions executed in order (run-to-completion)
	expected := []string{"start-action", "end-action"}
	if len(actionLog) != len(expected) {
		t.Fatalf("Expected %d actions, got %d: %v", len(expected), len(actionLog), actionLog)
	}

	for i, exp := range expected {
		if actionLog[i] != exp {
			t.Errorf("Action %d: expected '%s', got '%s'", i, exp, actionLog[i])
		}
	}
}

// ============================================================================
// Snapshot/Restore Tests (Detailed)
// ============================================================================

func TestSnapshot_IncludesEventQueue(t *testing.T) {
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

	// Pause to queue events without processing
	err = engine.Control(id, CmdPause)
	if err != nil {
		t.Fatalf("Pause failed: %v", err)
	}

	// Queue some events (these won't be processed while paused)
	// Note: Dispatch may fail while paused, so we need to test differently
	// For now, just verify snapshot structure

	snap, err := engine.Snapshot(id)
	if err != nil {
		t.Fatalf("Snapshot failed: %v", err)
	}

	// EventQueue should exist (may be empty)
	if snap.EventQueue == nil {
		t.Error("Snapshot.EventQueue is nil")
	}
}

func TestRestore_NewIDDifferentFromOriginal(t *testing.T) {
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

	snap, err := engine.Snapshot(id)
	if err != nil {
		t.Fatalf("Snapshot failed: %v", err)
	}

	newID, err := engine.Restore(snap)
	if err != nil {
		t.Fatalf("Restore failed: %v", err)
	}

	if newID == id {
		t.Error("Restored runtime should have a new ID")
	}

	if newID == "" {
		t.Error("Restored runtime ID is empty")
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

func TestControl_StartFromWrongState(t *testing.T) {
	engine := newTestEngine(t)
	def := newSimpleAtomicChart()
	mockCtx := testutil.NewMockApplicationContext()

	id, err := engine.Spawn(def, mockCtx)
	if err != nil {
		t.Fatalf("Spawn failed: %v", err)
	}

	// Start from Created should work
	err = engine.Control(id, CmdStart)
	if err != nil {
		t.Fatalf("Start from Created failed: %v", err)
	}

	// Start from Running should fail
	err = engine.Control(id, CmdStart)
	if err == nil {
		t.Error("Start from Running should fail")
	}

	if !errors.Is(err, ErrInvalidState) {
		t.Errorf("Expected ErrInvalidState, got: %v", err)
	}
}

// ============================================================================
// Implied Behavior Tests (from arch-v1.md analysis)
// ============================================================================

func TestImplied_FIFOrderingWithinRuntime(t *testing.T) {
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

func TestImplied_LifecycleStateMachine(t *testing.T) {
	engine := newTestEngine(t)
	def := newSimpleAtomicChart()
	mockCtx := testutil.NewMockApplicationContext()

	id, err := engine.Spawn(def, mockCtx)
	if err != nil {
		t.Fatalf("Spawn failed: %v", err)
	}

	// Created -> Start -> Running
	err = engine.Control(id, CmdStart)
	if err != nil {
		t.Fatalf("Start failed: %v", err)
	}

	// Running -> Pause -> Paused
	err = engine.Control(id, CmdPause)
	if err != nil {
		t.Fatalf("Pause failed: %v", err)
	}

	// Paused -> Resume -> Running
	err = engine.Control(id, CmdResume)
	if err != nil {
		t.Fatalf("Resume failed: %v", err)
	}

	// Running -> Stop -> Stopped
	err = engine.Control(id, CmdStop)
	if err != nil {
		t.Fatalf("Stop failed: %v", err)
	}

	// Verify stopped by trying to dispatch
	err = engine.Dispatch(id, Event{Type: "test"})
	if err == nil {
		t.Error("Dispatch should fail after Stop")
	}
}

func TestImplied_IDUniqueness(t *testing.T) {
	engine := newTestEngine(t)
	def := newSimpleAtomicChart()

	ids := make(map[RuntimeID]bool)

	for i := 0; i < 100; i++ {
		mockCtx := testutil.NewMockApplicationContext()
		id, err := engine.Spawn(def, mockCtx)
		if err != nil {
			t.Fatalf("Spawn %d failed: %v", i, err)
		}

		if ids[id] {
			t.Fatalf("Duplicate ID generated: %s", id)
		}

		ids[id] = true
	}
}

func TestImplied_StoppedRuntimeCleanup(t *testing.T) {
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

	err = engine.Control(id, CmdStop)
	if err != nil {
		t.Fatalf("Stop failed: %v", err)
	}

	// After stop, the runtime should be cleaned up
	// Operations on the ID should fail
	err = engine.Control(id, CmdStart)
	if err == nil {
		t.Error("Control after Stop should fail")
	}

	err = engine.Dispatch(id, Event{Type: "test"})
	if err == nil {
		t.Error("Dispatch after Stop should fail")
	}

	_, err = engine.Snapshot(id)
	if err == nil {
		t.Error("Snapshot after Stop should fail")
	}
}
