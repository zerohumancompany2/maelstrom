package runtime

import (
	"testing"
	"time"

	"github.com/maelstrom/v3/pkg/chart"
	"github.com/maelstrom/v3/pkg/registry"
	"github.com/maelstrom/v3/pkg/statechart"
)

// TestFactory_CreateRuntime verifies single runtime creation.
func TestFactory_CreateRuntime(t *testing.T) {
	reg := registry.New()

	// Store a chart definition
	reg.Set("test.yaml", chart.ChartDefinition{
		ID:      "test-chart",
		Version: "1.0.0",
	})

	factory := NewFactory(reg)
	rt, err := factory.Create("test.yaml")
	if err != nil {
		t.Fatalf("failed to create runtime: %v", err)
	}

	if rt == nil {
		t.Fatal("runtime should not be nil")
	}

	if rt.Context.ChartID != "test-chart" {
		t.Errorf("expected ChartID 'test-chart', got %q", rt.Context.ChartID)
	}
}

// TestFactory_CreateAllFromRegistry verifies bulk creation.
func TestFactory_CreateAllFromRegistry(t *testing.T) {
	reg := registry.New()

	// Store multiple charts
	reg.Set("chart1.yaml", chart.ChartDefinition{ID: "chart-1", Version: "1.0.0"})
	reg.Set("chart2.yaml", chart.ChartDefinition{ID: "chart-2", Version: "1.0.0"})

	factory := NewFactory(reg)
	runtimes, err := factory.CreateAll()
	if err != nil {
		t.Fatalf("failed to create runtimes: %v", err)
	}

	if len(runtimes) != 2 {
		t.Errorf("expected 2 runtimes, got %d", len(runtimes))
	}
}

// TestFactory_WithEngineSpawnsInStatechart verifies Factory connects runtime to engine.
func TestFactory_WithEngineSpawnsInStatechart(t *testing.T) {
	reg := registry.New()
	engine := statechart.NewEngine()

	// Store a chart definition
	reg.Set("test.yaml", statechart.ChartDefinition{
		ID:      "test-chart",
		Version: "1.0.0",
		Root: &statechart.Node{
			ID: "root",
			Children: map[string]*statechart.Node{
				"idle": {ID: "idle"},
			},
			IsInitial: true,
		},
		InitialState: "root/idle",
	})

	factory := NewFactory(reg).WithEngine(engine)
	rt, err := factory.Create("test.yaml")
	if err != nil {
		t.Fatalf("failed to create runtime: %v", err)
	}

	// Verify runtime was connected to engine
	if rt.runtimeID == "" {
		t.Error("runtime should have runtimeID from engine spawn")
	}

	if rt.engine == nil {
		t.Error("runtime should have engine reference")
	}
}

// TestFactory_RuntimeStatechartIntegration verifies full runtime→engine flow.
func TestFactory_RuntimeStatechartIntegration(t *testing.T) {
	reg := registry.New()
	engine := statechart.NewEngine()

	// Track that our action was called
	actionCalled := false
	engine.RegisterAction("transitionAction", func(ctx statechart.RuntimeContext, appCtx statechart.ApplicationContext, evt statechart.Event) error {
		actionCalled = true
		return nil
	})

	// Store a chart with a transition
	reg.Set("workflow.yaml", statechart.ChartDefinition{
		ID:      "workflow-chart",
		Version: "1.0.0",
		Root: &statechart.Node{
			ID: "root",
			Children: map[string]*statechart.Node{
				"idle": {
					ID:           "idle",
					Transitions:  []statechart.Transition{{Event: "start", Target: "root/running"}},
					ExitActions:  []string{"transitionAction"},
				},
				"running": {ID: "running"},
			},
			IsInitial: true,
		},
		InitialState: "root/idle",
	})

	// Create factory with engine connection
	factory := NewFactory(reg).WithEngine(engine)

	// Create runtime (spawns in engine)
	rt, err := factory.Create("workflow.yaml")
	if err != nil {
		t.Fatalf("failed to create runtime: %v", err)
	}

	// Start the engine runtime
	runtimeID := statechart.RuntimeID(rt.runtimeID)
	if err := engine.Control(runtimeID, statechart.CmdStart); err != nil {
		t.Fatalf("failed to start runtime: %v", err)
	}

	// Send event via runtime (should forward to engine)
	if err := rt.SendEvent(Event{Type: "start"}); err != nil {
		t.Fatalf("SendEvent failed: %v", err)
	}

	// Wait for async processing
	time.Sleep(50 * time.Millisecond)

	// Verify action was called (transition happened via statechart engine)
	if !actionCalled {
		t.Error("statechart action should have been called - event was not forwarded correctly")
	}

	// Verify state transition via snapshot
	snap, err := engine.Snapshot(runtimeID)
	if err != nil {
		t.Fatalf("snapshot failed: %v", err)
	}

	// Should be in running state now
	found := false
	for _, state := range snap.ActiveStates {
		if state == "root/running" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected to be in root/running state, got %v", snap.ActiveStates)
	}
}
