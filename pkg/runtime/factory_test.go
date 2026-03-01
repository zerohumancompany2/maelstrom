package runtime

import (
	"testing"
	"time"

	"github.com/maelstrom/v3/pkg/registry"
	"github.com/maelstrom/v3/pkg/statechart"
)

// TestFactory_CreateRuntime verifies single runtime creation.
func TestFactory_CreateRuntime(t *testing.T) {
	reg := registry.New()

	factory := NewFactory(reg)
	_, err := factory.Create("nonexistent.yaml")
	if err == nil {
		t.Fatal("expected error for nonexistent chart")
	}
}

// TestFactory_CreateAllFromRegistry verifies bulk creation without engine.
func TestFactory_CreateAllFromRegistry(t *testing.T) {
	reg := registry.New()

	// Store multiple chart definitions
	reg.Set("chart1.yaml", statechart.ChartDefinition{ID: "chart-1", Version: "1.0.0"})
	reg.Set("chart2.yaml", statechart.ChartDefinition{ID: "chart-2", Version: "1.0.0"})

	factory := NewFactory(reg)
	runtimes, err := factory.CreateAll()
	if err != nil {
		t.Fatalf("failed to create runtimes: %v", err)
	}

	if len(runtimes) != 2 {
		t.Errorf("expected 2 runtimes, got %d", len(runtimes))
	}
}

// TestFactory_WithEngine_ConfiguresEngine verifies fluent engine configuration.
func TestFactory_WithEngine_ConfiguresEngine(t *testing.T) {
	reg := registry.New()
	engine := statechart.NewEngine()

	factory := NewFactory(reg)

	// WithEngine should return the same factory (fluent)
	result := factory.WithEngine(engine)
	if result != factory {
		t.Error("WithEngine should return the same factory for fluent configuration")
	}
}

// TestFactory_Create_SpawnsInEngine verifies Create spawns runtime in engine.
func TestFactory_Create_SpawnsInEngine(t *testing.T) {
	reg := registry.New()
	engine := statechart.NewEngine()

	// Store a statechart definition
	def := statechart.ChartDefinition{
		ID:      "test-chart",
		Version: "1.0.0",
		Root: &statechart.Node{
			ID: "idle",
			Transitions: []statechart.Transition{
				{Event: "start", Target: "running"},
			},
		},
		InitialState: "idle",
	}
	reg.Set("test.yaml", def)

	factory := NewFactory(reg).WithEngine(engine)
	rt, err := factory.Create("test.yaml")
	if err != nil {
		t.Fatalf("failed to create runtime: %v", err)
	}

	if rt == nil {
		t.Fatal("runtime should not be nil")
	}

	// Verify mapping was stored
	id, ok := factory.GetStatechartID("test.yaml")
	if !ok {
		t.Fatal("expected mapping to exist for test.yaml")
	}

	if id == "" {
		t.Error("expected non-empty RuntimeID")
	}
}

// TestFactory_DispatchEvent_ForwardsToEngine verifies events flow through engine.
func TestFactory_DispatchEvent_ForwardsToEngine(t *testing.T) {
	reg := registry.New()
	engine := statechart.NewEngine()

	def := statechart.ChartDefinition{
		ID:      "test-chart",
		Version: "1.0.0",
		Root: &statechart.Node{
			ID: "idle",
			Transitions: []statechart.Transition{
				{Event: "go", Target: "active"},
			},
		},
		InitialState: "idle",
	}
	reg.Set("chart.yaml", def)

	factory := NewFactory(reg).WithEngine(engine)
	_, err := factory.Create("chart.yaml")
	if err != nil {
		t.Fatalf("failed to create runtime: %v", err)
	}

	chartID := "chart.yaml"
	id, ok := factory.GetStatechartID(chartID)
	if !ok {
		t.Fatal("expected mapping to exist")
	}

	err = engine.Control(id, statechart.CmdStart)
	if err != nil {
		t.Fatalf("Start failed: %v", err)
	}

	event := Event{Type: "go"}
	err = factory.DispatchEvent(chartID, event)
	if err != nil {
		t.Fatalf("DispatchEvent failed: %v", err)
	}

	// Allow async processing
	time.Sleep(100 * time.Millisecond)

	// Verify engine received the event via snapshot
	snap, err := engine.Snapshot(id)
	if err != nil {
		t.Fatalf("Snapshot failed: %v", err)
	}

	if len(snap.ActiveStates) == 0 || snap.ActiveStates[0] != "active" {
		t.Errorf("expected state 'active', got %v", snap.ActiveStates)
	}
}

// TestFactory_Create_WithNonChartDefinition_ReturnsError verifies type validation.
func TestFactory_Create_WithNonChartDefinition_ReturnsError(t *testing.T) {
	reg := registry.New()
	engine := statechart.NewEngine()

	// Store a non-chart definition
	reg.Set("bad.yaml", map[string]interface{}{
		"key": "value",
	})

	factory := NewFactory(reg).WithEngine(engine)
	_, err := factory.Create("bad.yaml")
	if err == nil {
		t.Fatal("expected error for non-chart definition")
	}
}

// TestFactory_DispatchEvent_UnknownChart_ReturnsError verifies error for unknown chart.
func TestFactory_DispatchEvent_UnknownChart_ReturnsError(t *testing.T) {
	reg := registry.New()
	engine := statechart.NewEngine()

	factory := NewFactory(reg).WithEngine(engine)

	err := factory.DispatchEvent("unknown-chart", Event{Type: "test"})
	if err == nil {
		t.Fatal("expected error for unknown chart")
	}

	expectedErr := "chart not found: unknown-chart"
	if err.Error() != expectedErr {
		t.Errorf("expected error %q, got %q", expectedErr, err.Error())
	}
}

// TestFactory_CreateAll_MultipleRuntimes verifies CreateAll spawns all in engine.
func TestFactory_CreateAll_MultipleRuntimes(t *testing.T) {
	reg := registry.New()
	engine := statechart.NewEngine()

	reg.Set("chart1.yaml", statechart.ChartDefinition{
		ID:      "chart-1",
		Version: "1.0.0",
		Root: &statechart.Node{
			ID: "idle",
			Transitions: []statechart.Transition{
				{Event: "go", Target: "running"},
			},
		},
		InitialState: "idle",
	})
	reg.Set("chart2.yaml", statechart.ChartDefinition{
		ID:      "chart-2",
		Version: "1.0.0",
		Root: &statechart.Node{
			ID: "idle",
			Transitions: []statechart.Transition{
				{Event: "go", Target: "running"},
			},
		},
		InitialState: "idle",
	})

	factory := NewFactory(reg).WithEngine(engine)
	runtimes, err := factory.CreateAll()
	if err != nil {
		t.Fatalf("CreateAll failed: %v", err)
	}

	if len(runtimes) != 2 {
		t.Errorf("expected 2 runtimes, got %d", len(runtimes))
	}

	// Verify both were spawned in engine
	if _, ok := factory.GetStatechartID("chart1.yaml"); !ok {
		t.Error("expected mapping for chart1.yaml")
	}
	if _, ok := factory.GetStatechartID("chart2.yaml"); !ok {
		t.Error("expected mapping for chart2.yaml")
	}

	// Start runtimes before dispatching events
	id1, _ := factory.GetStatechartID("chart1.yaml")
	id2, _ := factory.GetStatechartID("chart2.yaml")
	err = engine.Control(id1, statechart.CmdStart)
	if err != nil {
		t.Fatalf("Start chart1 failed: %v", err)
	}
	err = engine.Control(id2, statechart.CmdStart)
	if err != nil {
		t.Fatalf("Start chart2 failed: %v", err)
	}

	// Dispatch events to verify engine integration
	err = factory.DispatchEvent("chart1.yaml", Event{Type: "go"})
	if err != nil {
		t.Errorf("DispatchEvent for chart1.yaml failed: %v", err)
	}
	err = factory.DispatchEvent("chart2.yaml", Event{Type: "go"})
	if err != nil {
		t.Errorf("DispatchEvent for chart2.yaml failed: %v", err)
	}

	// Allow async processing
	time.Sleep(100 * time.Millisecond)

	// Verify states changed in engine
	snap1, err := engine.Snapshot(id1)
	if err != nil {
		t.Fatalf("Snapshot for chart1 failed: %v", err)
	}
	if len(snap1.ActiveStates) == 0 || snap1.ActiveStates[0] != "running" {
		t.Errorf("expected chart1 state 'running', got %v", snap1.ActiveStates)
	}

	snap2, err := engine.Snapshot(id2)
	if err != nil {
		t.Fatalf("Snapshot for chart2 failed: %v", err)
	}
	if len(snap2.ActiveStates) == 0 || snap2.ActiveStates[0] != "running" {
		t.Errorf("expected chart2 state 'running', got %v", snap2.ActiveStates)
	}
}

// TestFactory_FullIntegration_EventCausesTransition verifies full integration flow.
func TestFactory_FullIntegration_EventCausesTransition(t *testing.T) {
	reg := registry.New()
	engine := statechart.NewEngine()

	// Register a custom action
	actionCalled := false
	engine.RegisterAction("setFlag", func(runtimeCtx statechart.RuntimeContext, appCtx statechart.ApplicationContext, event statechart.Event) error {
		actionCalled = true
		return nil
	})

	def := statechart.ChartDefinition{
		ID:      "integration-chart",
		Version: "1.0.0",
		Root: &statechart.Node{
			ID: "idle",
			Transitions: []statechart.Transition{
				{Event: "go", Actions: []string{"setFlag"}, Target: "active"},
			},
		},
		InitialState: "idle",
	}
	reg.Set("integration.yaml", def)

	factory := NewFactory(reg).WithEngine(engine)
	_, err := factory.Create("integration.yaml")
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	// Start runtime before dispatching events
	id, ok := factory.GetStatechartID("integration.yaml")
	if !ok {
		t.Fatal("expected mapping to exist")
	}
	err = engine.Control(id, statechart.CmdStart)
	if err != nil {
		t.Fatalf("Start failed: %v", err)
	}

	// Dispatch event that triggers transition and action
	err = factory.DispatchEvent("integration.yaml", Event{Type: "go"})
	if err != nil {
		t.Fatalf("DispatchEvent failed: %v", err)
	}

	// Allow async processing
	time.Sleep(100 * time.Millisecond)

	// Verify state changed
	snap, err := engine.Snapshot(id)
	if err != nil {
		t.Fatalf("Snapshot failed: %v", err)
	}

	if len(snap.ActiveStates) == 0 || snap.ActiveStates[0] != "active" {
		t.Errorf("expected state 'active', got %v", snap.ActiveStates)
	}

	if !actionCalled {
		t.Error("expected action to be called")
	}
}
