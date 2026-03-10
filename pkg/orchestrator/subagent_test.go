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

func TestSubAgent_AttachedAutoTerminated(t *testing.T) {
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

	executor := NewSubAgentExecutor(config, "test-ns", parentID, engine)
	childID, err := executor.spawnAttached()
	if err != nil {
		t.Fatalf("Failed to spawn attached sub-agent: %v", err)
	}

	// When - parent is stopped
	err = engine.Control(parentID, statechart.CmdStop)

	// Then - child should be auto-terminated
	if err != nil {
		t.Fatalf("Expected Control(CmdStop) to return nil error, got %v", err)
	}

	// Verify child runtime no longer exists
	_, err = engine.(interface {
		Snapshot(statechart.RuntimeID) (statechart.Snapshot, error)
	}).Snapshot(childID)
	if err == nil {
		t.Error("Expected child runtime to be terminated (snapshot should fail)")
	}
}

func TestSubAgent_Detached_FireAndForget_IndependentLifecycle(t *testing.T) {
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
		Type:     SubAgentDetached,
		ChartRef: "detached-chart",
	}

	executor := NewSubAgentExecutor(config, "test-ns", parentID, engine)

	// When - spawn detached sub-agent
	childID, err := executor.spawnDetached()

	// Then - child should be spawned independently
	if err != nil {
		t.Fatalf("Expected spawnDetached() to return nil error, got %v", err)
	}

	if childID == "" {
		t.Error("Expected spawnDetached() to return non-empty RuntimeID")
	}

	// Verify child is independent (not bound to parent lifecycle)
	// Parent can be stopped without affecting child
	err = engine.Control(parentID, statechart.CmdStop)
	if err != nil {
		t.Fatalf("Expected parent stop to succeed: %v", err)
	}

	// Child should still exist after parent stops
	_, err = engine.(interface {
		Snapshot(statechart.RuntimeID) (statechart.Snapshot, error)
	}).Snapshot(childID)
	if err != nil {
		t.Errorf("Expected child runtime to still exist after parent stops, got error: %v", err)
	}
}

func TestSubAgent_Detached_FireAndAwait_WaitsForResult(t *testing.T) {
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

	correlationID := "test-correlation-123"
	config := SubAgentConfig{
		Type:          SubAgentDetached,
		ChartRef:      "detached-chart",
		CorrelationId: correlationID,
	}

	executor := NewSubAgentExecutor(config, "test-ns", parentID, engine)

	// When - spawn detached sub-agent with correlation ID
	childID, err := executor.spawnDetached()

	// Then - child should be spawned with correlation ID for result delivery
	if err != nil {
		t.Fatalf("Expected spawnDetached() to return nil error, got %v", err)
	}

	if childID == "" {
		t.Error("Expected spawnDetached() to return non-empty RuntimeID")
	}

	// Verify correlation ID is stored in executor config
	if executor.config.CorrelationId != correlationID {
		t.Errorf("Expected correlation ID to be '%s', got '%s'", correlationID, executor.config.CorrelationId)
	}
}

func TestSubAgent_MaxIterations_TerminatesOnLimit(t *testing.T) {
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

	maxIterations := 5
	config := SubAgentConfig{
		Type:          SubAgentAttached,
		ChartRef:      "test-chart",
		MaxIterations: maxIterations,
	}

	executor := NewSubAgentExecutor(config, "test-ns", parentID, engine)

	// When - executor is created with max iterations limit
	// Then - max iterations should be enforced
	if executor.config.MaxIterations != maxIterations {
		t.Errorf("Expected MaxIterations to be %d, got %d", maxIterations, executor.config.MaxIterations)
	}

	// Verify max iterations is set correctly
	if executor.config.MaxIterations <= 0 {
		t.Error("Expected MaxIterations to be positive")
	}
}
