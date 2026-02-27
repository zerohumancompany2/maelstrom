package statechart

import (
	"testing"
	"time"

	"github.com/maelstrom/v3/internal/testutil"
)

// TestTransientSubChart_SpawnsWithLifecycleBinding verifies sub-chart auto-starts with parent.
func TestTransientSubChart_SpawnsWithLifecycleBinding(t *testing.T) {
	engine := NewEngine().(*Engine)
	mockCtx := testutil.NewMockApplicationContext()

	// Parent chart that spawns a sub-chart on entry
	parentDef := ChartDefinition{
		ID:      "parent",
		Version: "1.0.0",
		Root: &Node{
			ID:       "idle",
			Children: nil,
			Transitions: []Transition{
				{Event: "spawnChild", Target: "idle"},
			},
		},
		InitialState: "idle",
	}

	childDef := ChartDefinition{
		ID:      "child",
		Version: "1.0.0",
		Root: &Node{
			ID:       "waiting",
			Children: nil,
			Transitions: []Transition{
				{Event: "done", Target: "complete"},
			},
		},
		InitialState: "waiting",
	}

	parentID, err := engine.Spawn(parentDef, mockCtx)
	if err != nil {
		t.Fatalf("Spawn parent failed: %v", err)
	}

	if err := engine.Control(parentID, CmdStart); err != nil {
		t.Fatalf("Start parent failed: %v", err)
	}

	// Spawn transient sub-chart
	childID, err := engine.SpawnTransient(childDef, mockCtx, parentID)
	if err != nil {
		t.Fatalf("SpawnTransient failed: %v", err)
	}

	// Verify child is running
	if _, exists := engine.runtimes[childID]; !exists {
		t.Error("Child runtime should exist")
	}

	// Stop parent - child should also stop
	if err := engine.Control(parentID, CmdStop); err != nil {
		t.Fatalf("Stop parent failed: %v", err)
	}

	time.Sleep(50 * time.Millisecond)

	// Verify child is also stopped
	if _, exists := engine.runtimes[childID]; exists {
		t.Error("Child runtime should be stopped when parent stops")
	}
}

// TestTransientSubChart_ParentStopCleansUpChildren verifies cascade stop.
func TestTransientSubChart_ParentStopCleansUpChildren(t *testing.T) {
	engine := NewEngine().(*Engine)
	mockCtx := testutil.NewMockApplicationContext()

	parentDef := ChartDefinition{
		ID:           "parent",
		Version:      "1.0.0",
		Root:         &Node{ID: "idle"},
		InitialState: "idle",
	}

	childDef := ChartDefinition{
		ID:           "child",
		Version:      "1.0.0",
		Root:         &Node{ID: "waiting"},
		InitialState: "waiting",
	}

	parentID, _ := engine.Spawn(parentDef, mockCtx)
	engine.Control(parentID, CmdStart)

	// Spawn multiple children
	child1, _ := engine.SpawnTransient(childDef, mockCtx, parentID)
	child2, _ := engine.SpawnTransient(childDef, mockCtx, parentID)

	// Stop parent
	engine.Control(parentID, CmdStop)
	time.Sleep(50 * time.Millisecond)

	// All children should be cleaned up
	if _, exists := engine.runtimes[child1]; exists {
		t.Error("Child1 should be stopped")
	}
	if _, exists := engine.runtimes[child2]; exists {
		t.Error("Child2 should be stopped")
	}
}
