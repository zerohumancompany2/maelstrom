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
