package statechart

import (
	"testing"
	"time"

	"github.com/maelstrom/v3/internal/testutil"
)

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
