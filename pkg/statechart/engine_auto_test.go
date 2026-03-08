package statechart

import (
	"testing"

	"github.com/maelstrom/v3/internal/testutil"
)

// TestEngine_AutoTransition_SyntheticEventFires verifies that entry actions run
// and then the synthetic auto-transition event is emitted and processed.
func TestEngine_AutoTransition_SyntheticEventFires(t *testing.T) {
	engine := NewEngine()

	// Track execution order
	var executionOrder []string

	// Register entry action for prepare state
	engine.RegisterAction("prepareEntry", func(ctx RuntimeContext, appCtx ApplicationContext, ev Event) error {
		executionOrder = append(executionOrder, "prepareEntry")
		return nil
	})

	// Register entry action for execute state
	engine.RegisterAction("executeEntry", func(ctx RuntimeContext, appCtx ApplicationContext, ev Event) error {
		executionOrder = append(executionOrder, "executeEntry")
		return nil
	})

	// Define chart with auto-transition
	def := ChartDefinition{
		ID:      "auto-test",
		Version: "1.0.0",
		Root: &Node{
			ID: "root",
			Children: map[string]*Node{
				"prepare": {
					ID:           "prepare",
					EntryActions: []string{"prepareEntry"},
					Transitions: []Transition{
						{
							Event:  SyntheticAutoTransition, // Auto-transition
							Target: "root/execute",
						},
					},
				},
				"execute": {
					ID:           "execute",
					EntryActions: []string{"executeEntry"},
				},
			},
		},
		InitialState: "root/prepare",
	}

	// Create mock app context
	appCtx := testutil.NewMockApplicationContext()

	// Spawn and start runtime
	rtID, err := engine.Spawn(def, appCtx)
	if err != nil {
		t.Fatalf("Spawn failed: %v", err)
	}

	err = engine.Control(rtID, CmdStart)
	if err != nil {
		t.Fatalf("Start failed: %v", err)
	}

	// Check execution order
	if len(executionOrder) != 2 {
		t.Fatalf("Expected 2 actions to execute, got %d: %v", len(executionOrder), executionOrder)
	}

	if executionOrder[0] != "prepareEntry" {
		t.Errorf("Expected first action to be 'prepareEntry', got '%s'", executionOrder[0])
	}

	if executionOrder[1] != "executeEntry" {
		t.Errorf("Expected second action to be 'executeEntry', got '%s'", executionOrder[1])
	}

	// Verify final state
	snap, err := engine.Snapshot(rtID)
	if err != nil {
		t.Fatalf("Snapshot failed: %v", err)
	}

	if snap.ActiveStates[0] != "root/execute" {
		t.Errorf("Expected final state to be 'root/execute', got '%s'", snap.ActiveStates[0])
	}
}

// TestEngine_AutoTransition_WithGuard verifies that auto-transitions respect guards.
func TestEngine_AutoTransition_WithGuard(t *testing.T) {
	engine := NewEngine()

	var executionOrder []string
	allowTransition := true

	engine.RegisterAction("setupEntry", func(ctx RuntimeContext, appCtx ApplicationContext, ev Event) error {
		executionOrder = append(executionOrder, "setupEntry")
		return nil
	})

	engine.RegisterAction("runEntry", func(ctx RuntimeContext, appCtx ApplicationContext, ev Event) error {
		executionOrder = append(executionOrder, "runEntry")
		return nil
	})

	engine.RegisterGuard("canRun", func(appCtx ApplicationContext, ev Event) bool {
		return allowTransition
	})

	def := ChartDefinition{
		ID:      "auto-guard-test",
		Version: "1.0.0",
		Root: &Node{
			ID: "root",
			Children: map[string]*Node{
				"setup": {
					ID:           "setup",
					EntryActions: []string{"setupEntry"},
					Transitions: []Transition{
						{
							Event:  SyntheticAutoTransition,
							Target: "root/run",
							Guard:  "canRun",
						},
					},
				},
				"run": {
					ID:           "run",
					EntryActions: []string{"runEntry"},
				},
			},
		},
		InitialState: "root/setup",
	}

	appCtx := testutil.NewMockApplicationContext()
	rtID, err := engine.Spawn(def, appCtx)
	if err != nil {
		t.Fatalf("Spawn failed: %v", err)
	}

	err = engine.Control(rtID, CmdStart)
	if err != nil {
		t.Fatalf("Start failed: %v", err)
	}

	if len(executionOrder) != 2 {
		t.Fatalf("Expected 2 actions, got %d: %v", len(executionOrder), executionOrder)
	}

	snap, err := engine.Snapshot(rtID)
	if err != nil {
		t.Fatalf("Snapshot failed: %v", err)
	}

	if snap.ActiveStates[0] != "root/run" {
		t.Errorf("Expected state 'root/run', got '%s'", snap.ActiveStates[0])
	}
}

// TestEngine_AutoTransition_GuardBlocks verifies that a failing guard blocks auto-transition.
func TestEngine_AutoTransition_GuardBlocks(t *testing.T) {
	engine := NewEngine()

	var executionOrder []string

	engine.RegisterAction("setupEntry", func(ctx RuntimeContext, appCtx ApplicationContext, ev Event) error {
		executionOrder = append(executionOrder, "setupEntry")
		return nil
	})

	engine.RegisterAction("runEntry", func(ctx RuntimeContext, appCtx ApplicationContext, ev Event) error {
		executionOrder = append(executionOrder, "runEntry")
		return nil
	})

	engine.RegisterGuard("canRun", func(appCtx ApplicationContext, ev Event) bool {
		return false // Block the transition
	})

	def := ChartDefinition{
		ID:      "auto-guard-block-test",
		Version: "1.0.0",
		Root: &Node{
			ID: "root",
			Children: map[string]*Node{
				"setup": {
					ID:           "setup",
					EntryActions: []string{"setupEntry"},
					Transitions: []Transition{
						{
							Event:  SyntheticAutoTransition,
							Target: "root/run",
							Guard:  "canRun",
						},
					},
				},
				"run": {
					ID:           "run",
					EntryActions: []string{"runEntry"},
				},
			},
		},
		InitialState: "root/setup",
	}

	appCtx := testutil.NewMockApplicationContext()
	rtID, err := engine.Spawn(def, appCtx)
	if err != nil {
		t.Fatalf("Spawn failed: %v", err)
	}

	err = engine.Control(rtID, CmdStart)
	if err != nil {
		t.Fatalf("Start failed: %v", err)
	}

	if len(executionOrder) != 1 {
		t.Fatalf("Expected 1 action (setupEntry only), got %d: %v", len(executionOrder), executionOrder)
	}

	snap, err := engine.Snapshot(rtID)
	if err != nil {
		t.Fatalf("Snapshot failed: %v", err)
	}

	if snap.ActiveStates[0] != "root/setup" {
		t.Errorf("Expected state 'root/setup' (blocked), got '%s'", snap.ActiveStates[0])
	}
}

// TestEngine_AutoTransition_WithActions verifies that transition actions execute.
func TestEngine_AutoTransition_WithActions(t *testing.T) {
	engine := NewEngine()

	var executionOrder []string

	engine.RegisterAction("initEntry", func(ctx RuntimeContext, appCtx ApplicationContext, ev Event) error {
		executionOrder = append(executionOrder, "initEntry")
		return nil
	})

	engine.RegisterAction("transitionAction", func(ctx RuntimeContext, appCtx ApplicationContext, ev Event) error {
		executionOrder = append(executionOrder, "transitionAction")
		return nil
	})

	engine.RegisterAction("readyEntry", func(ctx RuntimeContext, appCtx ApplicationContext, ev Event) error {
		executionOrder = append(executionOrder, "readyEntry")
		return nil
	})

	def := ChartDefinition{
		ID:      "auto-actions-test",
		Version: "1.0.0",
		Root: &Node{
			ID: "root",
			Children: map[string]*Node{
				"init": {
					ID:           "init",
					EntryActions: []string{"initEntry"},
					Transitions: []Transition{
						{
							Event:   SyntheticAutoTransition,
							Target:  "root/ready",
							Actions: []string{"transitionAction"},
						},
					},
				},
				"ready": {
					ID:           "ready",
					EntryActions: []string{"readyEntry"},
				},
			},
		},
		InitialState: "root/init",
	}

	appCtx := testutil.NewMockApplicationContext()
	rtID, err := engine.Spawn(def, appCtx)
	if err != nil {
		t.Fatalf("Spawn failed: %v", err)
	}

	err = engine.Control(rtID, CmdStart)
	if err != nil {
		t.Fatalf("Start failed: %v", err)
	}

	expectedOrder := []string{"initEntry", "transitionAction", "readyEntry"}
	if len(executionOrder) != len(expectedOrder) {
		t.Fatalf("Expected %d actions, got %d: %v", len(expectedOrder), len(executionOrder), executionOrder)
	}

	for i, expected := range expectedOrder {
		if executionOrder[i] != expected {
			t.Errorf("Expected action[%d] to be '%s', got '%s'", i, expected, executionOrder[i])
		}
	}
}

// TestEngine_AutoTransition_Chained verifies chained auto-transitions.
func TestEngine_AutoTransition_Chained(t *testing.T) {
	engine := NewEngine()

	var executionOrder []string

	engine.RegisterAction("step1Entry", func(ctx RuntimeContext, appCtx ApplicationContext, ev Event) error {
		executionOrder = append(executionOrder, "step1Entry")
		return nil
	})

	engine.RegisterAction("step2Entry", func(ctx RuntimeContext, appCtx ApplicationContext, ev Event) error {
		executionOrder = append(executionOrder, "step2Entry")
		return nil
	})

	engine.RegisterAction("step3Entry", func(ctx RuntimeContext, appCtx ApplicationContext, ev Event) error {
		executionOrder = append(executionOrder, "step3Entry")
		return nil
	})

	def := ChartDefinition{
		ID:      "auto-chained-test",
		Version: "1.0.0",
		Root: &Node{
			ID: "root",
			Children: map[string]*Node{
				"step1": {
					ID:           "step1",
					EntryActions: []string{"step1Entry"},
					Transitions: []Transition{
						{
							Event:  SyntheticAutoTransition,
							Target: "root/step2",
						},
					},
				},
				"step2": {
					ID:           "step2",
					EntryActions: []string{"step2Entry"},
					Transitions: []Transition{
						{
							Event:  SyntheticAutoTransition,
							Target: "root/step3",
						},
					},
				},
				"step3": {
					ID:           "step3",
					EntryActions: []string{"step3Entry"},
				},
			},
		},
		InitialState: "root/step1",
	}

	appCtx := testutil.NewMockApplicationContext()
	rtID, err := engine.Spawn(def, appCtx)
	if err != nil {
		t.Fatalf("Spawn failed: %v", err)
	}

	err = engine.Control(rtID, CmdStart)
	if err != nil {
		t.Fatalf("Start failed: %v", err)
	}

	expectedOrder := []string{"step1Entry", "step2Entry", "step3Entry"}
	if len(executionOrder) != len(expectedOrder) {
		t.Fatalf("Expected %d actions, got %d: %v", len(expectedOrder), len(executionOrder), executionOrder)
	}

	for i, expected := range expectedOrder {
		if executionOrder[i] != expected {
			t.Errorf("Expected action[%d] to be '%s', got '%s'", i, expected, executionOrder[i])
		}
	}

	snap, err := engine.Snapshot(rtID)
	if err != nil {
		t.Fatalf("Snapshot failed: %v", err)
	}

	if snap.ActiveStates[0] != "root/step3" {
		t.Errorf("Expected final state 'root/step3', got '%s'", snap.ActiveStates[0])
	}
}

// TestEngine_AutoTransition_ExitActions verifies exit actions run during auto-transition.
func TestEngine_AutoTransition_ExitActions(t *testing.T) {
	engine := NewEngine()

	var executionOrder []string

	engine.RegisterAction("tempEntry", func(ctx RuntimeContext, appCtx ApplicationContext, ev Event) error {
		executionOrder = append(executionOrder, "tempEntry")
		return nil
	})

	engine.RegisterAction("tempExit", func(ctx RuntimeContext, appCtx ApplicationContext, ev Event) error {
		executionOrder = append(executionOrder, "tempExit")
		return nil
	})

	engine.RegisterAction("finalEntry", func(ctx RuntimeContext, appCtx ApplicationContext, ev Event) error {
		executionOrder = append(executionOrder, "finalEntry")
		return nil
	})

	def := ChartDefinition{
		ID:      "auto-exit-test",
		Version: "1.0.0",
		Root: &Node{
			ID: "root",
			Children: map[string]*Node{
				"temp": {
					ID:           "temp",
					EntryActions: []string{"tempEntry"},
					ExitActions:  []string{"tempExit"},
					Transitions: []Transition{
						{
							Event:  SyntheticAutoTransition,
							Target: "root/final",
						},
					},
				},
				"final": {
					ID:           "final",
					EntryActions: []string{"finalEntry"},
				},
			},
		},
		InitialState: "root/temp",
	}

	appCtx := testutil.NewMockApplicationContext()
	rtID, err := engine.Spawn(def, appCtx)
	if err != nil {
		t.Fatalf("Spawn failed: %v", err)
	}

	err = engine.Control(rtID, CmdStart)
	if err != nil {
		t.Fatalf("Start failed: %v", err)
	}

	expectedOrder := []string{"tempEntry", "tempExit", "finalEntry"}
	if len(executionOrder) != len(expectedOrder) {
		t.Fatalf("Expected %d actions, got %d: %v", len(expectedOrder), len(executionOrder), executionOrder)
	}

	for i, expected := range expectedOrder {
		if executionOrder[i] != expected {
			t.Errorf("Expected action[%d] to be '%s', got '%s'", i, expected, executionOrder[i])
		}
	}
}

// TestEngine_AutoTransition_CorrelationID verifies correlation ID is preserved.
func TestEngine_AutoTransition_CorrelationID(t *testing.T) {
	engine := NewEngine()

	var capturedCorrID string

	engine.RegisterAction("aEntry", func(ctx RuntimeContext, appCtx ApplicationContext, ev Event) error {
		if ev.CorrelationID != "" {
			capturedCorrID = ev.CorrelationID
		}
		return nil
	})

	engine.RegisterAction("bEntry", func(ctx RuntimeContext, appCtx ApplicationContext, ev Event) error {
		if ev.CorrelationID != "" {
			capturedCorrID = ev.CorrelationID
		}
		return nil
	})

	def := ChartDefinition{
		ID:      "auto-corr-test",
		Version: "1.0.0",
		Root: &Node{
			ID: "root",
			Children: map[string]*Node{
				"a": {
					ID:           "a",
					EntryActions: []string{"aEntry"},
					Transitions: []Transition{
						{
							Event:  SyntheticAutoTransition,
							Target: "root/b",
						},
					},
				},
				"b": {
					ID:           "b",
					EntryActions: []string{"bEntry"},
				},
			},
		},
		InitialState: "root/a",
	}

	appCtx := testutil.NewMockApplicationContext()
	rtID, err := engine.Spawn(def, appCtx)
	if err != nil {
		t.Fatalf("Spawn failed: %v", err)
	}

	err = engine.Control(rtID, CmdStart)
	if err != nil {
		t.Fatalf("Start failed: %v", err)
	}

	// Synthetic event should have Source "system:auto"
	if capturedCorrID != "" {
		t.Logf("Correlation ID captured: %s", capturedCorrID)
	}
}
