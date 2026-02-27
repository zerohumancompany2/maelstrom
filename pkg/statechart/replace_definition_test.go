package statechart

import (
	"testing"
	"time"

	"github.com/maelstrom/v3/internal/testutil"
)

// TestEngine_ReplaceDefinition_Success replaces chart definition at runtime.
func TestEngine_ReplaceDefinition_Success(t *testing.T) {
	engine := NewEngine().(*Engine)
	mockCtx := testutil.NewMockApplicationContext()

	// Initial definition
	def := ChartDefinition{
		ID:      "test-replace",
		Version: "1.0.0",
		Root: &Node{
			ID:       "idle",
			Children: nil,
			Transitions: []Transition{
				{Event: "go", Target: "running"},
			},
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

	// New definition with additional state
	newDef := ChartDefinition{
		ID:      "test-replace",
		Version: "1.1.0",
		Root: &Node{
			ID: "idle",
			Children: map[string]*Node{
				"running": {ID: "running"},
				"paused":  {ID: "paused"},
			},
			Transitions: []Transition{
				{Event: "go", Target: "idle/running"},
				{Event: "pause", Target: "idle/paused"},
			},
		},
		InitialState: "idle",
	}

	// Act: Replace definition
	if err := engine.ReplaceDefinition(rtID, newDef); err != nil {
		t.Fatalf("ReplaceDefinition failed: %v", err)
	}

	// Assert: runtime should have new definition
	runtime := engine.runtimes[rtID]
	if runtime.definition.Version != "1.1.0" {
		t.Errorf("Expected version 1.1.0, got %s", runtime.definition.Version)
	}

	// Should be able to use new transition
	if err := engine.Dispatch(rtID, Event{Type: "pause"}); err != nil {
		t.Fatalf("Dispatch with new event type failed: %v", err)
	}
}

// TestEngine_ReplaceDefinition_InvalidID fails for non-existent runtime.
func TestEngine_ReplaceDefinition_InvalidID(t *testing.T) {
	engine := NewEngine().(*Engine)

	newDef := ChartDefinition{
		ID:      "test",
		Version: "2.0.0",
		Root: &Node{
			ID:       "idle",
			Children: nil,
		},
		InitialState: "idle",
	}

	err := engine.ReplaceDefinition("non-existent", newDef)
	if err == nil {
		t.Error("Expected error for non-existent runtime")
	}
}

// TestEngine_ReplaceDefinition_PreservesState keeps active state if compatible.
func TestEngine_ReplaceDefinition_PreservesState(t *testing.T) {
	engine := NewEngine().(*Engine)
	mockCtx := testutil.NewMockApplicationContext()

	def := ChartDefinition{
		ID:      "test-replace",
		Version: "1.0.0",
		Root: &Node{
			ID: "idle",
			Children: map[string]*Node{
				"running": {ID: "running"},
			},
			Transitions: []Transition{
				{Event: "go", Target: "idle/running"},
			},
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

	// Transition to running
	engine.Dispatch(rtID, Event{Type: "go"})
	time.Sleep(100 * time.Millisecond)

	runtime := engine.runtimes[rtID]
	if runtime.activeState != "idle/running" {
		t.Fatalf("Expected activeState 'idle/running', got '%s'", runtime.activeState)
	}

	// New definition with same state structure but new transition
	newDef := ChartDefinition{
		ID:      "test-replace",
		Version: "1.1.0",
		Root: &Node{
			ID: "idle",
			Children: map[string]*Node{
				"running": {
					ID:       "running",
					Children: nil,
					Transitions: []Transition{
						{Event: "stop", Target: "idle"},
					},
				},
			},
			Transitions: []Transition{
				{Event: "go", Target: "idle/running"},
			},
		},
		InitialState: "idle",
	}

	// Act: Replace definition
	if err := engine.ReplaceDefinition(rtID, newDef); err != nil {
		t.Fatalf("ReplaceDefinition failed: %v", err)
	}

	// Assert: Should still be in running state
	if runtime.activeState != "idle/running" {
		t.Errorf("Active state should be preserved as 'idle/running', got '%s'", runtime.activeState)
	}

	// Should be able to use new transition from running
	if err := engine.Dispatch(rtID, Event{Type: "stop"}); err != nil {
		t.Fatalf("Dispatch with new transition failed: %v", err)
	}
	time.Sleep(100 * time.Millisecond)
}
