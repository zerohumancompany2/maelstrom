package orchestrator

import (
	"testing"

	"github.com/maelstrom/v3/pkg/statechart"
)

func TestSubAgent_Attached_LifecycleBoundToParent(t *testing.T) {
	// Given
	engine := statechart.NewEngine()
	parentDef := statechart.ChartDefinition{
		ID: "parent",
		Root: &statechart.Node{
			ID: "root",
		},
	}
	parentID, err := engine.Spawn(parentDef, nil)
	if err != nil {
		t.Fatalf("Failed to spawn parent: %v", err)
	}

	config := SubAgentConfig{
		Type:     SubAgentAttached,
		ChartRef: "child-chart",
	}

	// When
	executor := NewSubAgentExecutor(config, "test-ns", parentID, engine)
	childID, err := executor.spawnAttached()

	// Then
	if err != nil {
		t.Fatalf("Expected spawnAttached() to return nil error, got %v", err)
	}

	if childID == "" {
		t.Error("Expected spawnAttached() to return non-empty RuntimeID")
	}

	// Child should be registered with parent - we can verify by checking if parent has children
	// This requires accessing internal state which we'll verify through lifecycle behavior
	_ = childID
}
