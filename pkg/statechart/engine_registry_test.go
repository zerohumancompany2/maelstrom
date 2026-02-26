package statechart

import (
	"testing"
	"time"

	"github.com/maelstrom/v3/internal/testutil"
)

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
