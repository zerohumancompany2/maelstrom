package lifecycle

import (
	"testing"

	"github.com/maelstrom/v3/pkg/mail"
	"github.com/maelstrom/v3/pkg/statechart"
)

func TestLifecycleService_NewLifecycleServiceReturnsNonNil(t *testing.T) {
	svc := NewLifecycleServiceWithoutEngine()

	if svc == nil {
		t.Error("Expected NewLifecycleServiceWithoutEngine to return non-nil")
	}
}

func TestLifecycleService_IDReturnsCorrectString(t *testing.T) {
	svc := NewLifecycleServiceWithoutEngine()

	id := svc.ID()

	if id != "sys:lifecycle" {
		t.Errorf("Expected ID sys:lifecycle, got %s", id)
	}
}

func TestLifecycleService_HandleMailReturnsNil(t *testing.T) {
	svc := NewLifecycleServiceWithoutEngine()

	err := svc.HandleMail(mail.Mail{})

	if err != nil {
		t.Errorf("Expected HandleMail to return nil, got %v", err)
	}
}

func TestLifecycleService_SpawnReturnsNonEmptyRuntimeID(t *testing.T) {
	svc := NewLifecycleServiceWithoutEngine()

	id, err := svc.Spawn(statechart.ChartDefinition{})

	if err != nil {
		t.Errorf("Expected Spawn to return nil error, got %v", err)
	}

	if id == "" {
		t.Error("Expected Spawn to return non-empty RuntimeID")
	}
}

func TestLifecycleService_StopReturnsNil(t *testing.T) {
	svc := NewLifecycleServiceWithoutEngine()

	err := svc.Stop(statechart.RuntimeID("test-123"))

	if err != nil {
		t.Errorf("Expected Stop to return nil, got %v", err)
	}
}

func TestLifecycleService_ListReturnsRuntimeInfo(t *testing.T) {
	svc := NewLifecycleServiceWithoutEngine()

	list, err := svc.List()

	if err != nil {
		t.Errorf("Expected List to return nil error, got %v", err)
	}

	// Verify it's []RuntimeInfo type by checking we can access RuntimeInfo fields
	if len(list) != 0 {
		t.Errorf("Expected empty slice, got %d items", len(list))
	}

	// Type assertion to verify return type
	_, ok := interface{}(list).([]RuntimeInfo)
	if !ok {
		t.Error("Expected List to return []RuntimeInfo type")
	}
}

func TestLifecycleService_ListEmptyWhenNoRuntimes(t *testing.T) {
	svc := NewLifecycleServiceWithoutEngine()

	list, err := svc.List()

	if err != nil {
		t.Errorf("Expected List to return nil error, got %v", err)
	}

	if len(list) != 0 {
		t.Errorf("Expected empty slice, got %d items", len(list))
	}

	if list == nil {
		t.Error("Expected non-nil empty slice")
	}
}

func TestLifecycleService_SpawnTracksRuntime(t *testing.T) {
	engine := statechart.NewEngine()
	svc := NewLifecycleService(engine)

	def := statechart.ChartDefinition{
		ID:           "test-chart",
		Version:      "1.0.0",
		InitialState: "idle",
	}

	rtID, err := svc.Spawn(def)
	if err != nil {
		t.Fatalf("Spawn failed: %v", err)
	}

	list, err := svc.List()
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}

	if len(list) != 1 {
		t.Errorf("Expected 1 runtime in list, got %d", len(list))
	}

	if list[0].ID != string(rtID) {
		t.Errorf("Expected runtime ID %s, got %s", rtID, list[0].ID)
	}

	if list[0].DefinitionID != "test-chart" {
		t.Errorf("Expected DefinitionID test-chart, got %s", list[0].DefinitionID)
	}

	if list[0].Boundary != mail.InnerBoundary {
		t.Errorf("Expected Boundary inner, got %v", list[0].Boundary)
	}
}

func TestLifecycleService_ControlStart(t *testing.T) {
	engine := statechart.NewEngine()
	svc := NewLifecycleService(engine)

	def := statechart.ChartDefinition{
		ID:           "test-chart",
		Version:      "1.0.0",
		InitialState: "idle",
		Root: &statechart.Node{
			ID: "root",
			Children: map[string]*statechart.Node{
				"idle": {ID: "idle"},
			},
		},
	}

	rtID, err := svc.Spawn(def)
	if err != nil {
		t.Fatalf("Spawn failed: %v", err)
	}

	err = svc.Control(rtID, statechart.CmdStart)
	if err != nil {
		t.Errorf("Expected Control(CmdStart) to return nil, got %v", err)
	}
}

func TestLifecycleService_ControlStop(t *testing.T) {
	engine := statechart.NewEngine()
	svc := NewLifecycleService(engine)

	def := statechart.ChartDefinition{
		ID:           "test-chart",
		Version:      "1.0.0",
		InitialState: "idle",
		Root: &statechart.Node{
			ID: "root",
			Children: map[string]*statechart.Node{
				"idle": {ID: "idle"},
			},
		},
	}

	rtID, err := svc.Spawn(def)
	if err != nil {
		t.Fatalf("Spawn failed: %v", err)
	}

	err = svc.Control(rtID, statechart.CmdStop)
	if err != nil {
		t.Errorf("Expected Control(CmdStop) to return nil, got %v", err)
	}
}

func TestLifecycleService_ControlNotFoundReturnsError(t *testing.T) {
	engine := statechart.NewEngine()
	svc := NewLifecycleService(engine)

	err := svc.Control(statechart.RuntimeID("non-existent"), statechart.CmdStart)
	if err == nil {
		t.Error("Expected Control with non-existent ID to return error")
	}

	svcNoEngine := NewLifecycleServiceWithoutEngine()
	err = svcNoEngine.Control(statechart.RuntimeID("any"), statechart.CmdStart)
	if err == nil {
		t.Error("Expected Control without engine to return error")
	}
}

func TestLifecycleService_StartReturnsNil(t *testing.T) {
	svc := NewLifecycleServiceWithoutEngine()

	err := svc.Start()

	if err != nil {
		t.Errorf("Expected Start to return nil, got %v", err)
	}
}

func TestLifecycleService_BootstrapChart(t *testing.T) {
	chart := BootstrapChart()

	if chart.ID != "sys:lifecycle" {
		t.Errorf("Expected ID sys:lifecycle, got %s", chart.ID)
	}

	if chart.Version != "1.0.0" {
		t.Errorf("Expected version 1.0.0, got %s", chart.Version)
	}
}

func TestLifecycleService_SpawnChart(t *testing.T) {
	svc := NewLifecycleServiceWithoutEngine()
	def := statechart.ChartDefinition{
		ID:      "test-chart",
		Version: "1.0.0",
	}
	rtID, err := svc.Spawn(def)
	if err != nil {
		t.Errorf("Spawn should return nil error, got: %v", err)
	}
	if rtID == "" {
		t.Error("Spawn should return non-empty runtime ID")
	}
}

func TestLifecycleService_BoundaryInner(t *testing.T) {
	svc := NewLifecycleServiceWithoutEngine()
	if svc.Boundary() != mail.InnerBoundary {
		t.Errorf("Expected boundary 'inner', got: %v", svc.Boundary())
	}
}

func TestLifecycleService_ID(t *testing.T) {
	chart := BootstrapChart()

	if chart.ID != "sys:lifecycle" {
		t.Errorf("Expected ID sys:lifecycle, got %s", chart.ID)
	}
}

func TestLifecycleService_NewWithEngineReturnsNonNil(t *testing.T) {
	engine := statechart.NewEngine()
	svc := NewLifecycleService(engine)

	if svc == nil {
		t.Error("Expected NewLifecycleService(engine) to return non-nil")
	}
}

func TestLifecycleService_RuntimeStateUpdate(t *testing.T) {
	svc := NewLifecycleServiceWithoutEngine()

	def := statechart.ChartDefinition{
		ID:           "test-chart",
		Version:      "1.0.0",
		InitialState: "idle",
	}
	rtID, _ := svc.Spawn(def)

	err := svc.updateRuntimeState(string(rtID), "running")
	if err != nil {
		t.Errorf("Expected nil error, got %v", err)
	}

	list, _ := svc.List()
	if len(list) != 1 {
		t.Errorf("Expected 1 runtime, got %d", len(list))
	}
	if list[0].ActiveStates[0] != "running" {
		t.Errorf("Expected state 'running', got %v", list[0].ActiveStates)
	}
}

func TestLifecycleService_ListWithStates(t *testing.T) {
	svc := NewLifecycleServiceWithoutEngine()

	def1 := statechart.ChartDefinition{ID: "chart-1", InitialState: "idle"}
	def2 := statechart.ChartDefinition{ID: "chart-2", InitialState: "running"}
	rtID1, _ := svc.Spawn(def1)
	_, _ = svc.Spawn(def2)

	svc.updateRuntimeState(string(rtID1), "active")

	runtimes, _ := svc.List()

	if len(runtimes) != 2 {
		t.Errorf("Expected 2 runtimes, got %d", len(runtimes))
	}
	foundActive := false
	for _, rt := range runtimes {
		if rt.ActiveStates[0] == "active" {
			foundActive = true
			break
		}
	}
	if !foundActive {
		t.Error("Expected to find runtime with 'active' state")
	}
}
