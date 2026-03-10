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

func TestLifecycleService_StateHistory(t *testing.T) {
	svc := NewLifecycleServiceWithoutEngine()

	def := statechart.ChartDefinition{ID: "chart-1", InitialState: "idle"}
	rtID, _ := svc.Spawn(def)

	svc.updateRuntimeState(string(rtID), "running")
	svc.updateRuntimeState(string(rtID), "stopped")

	history := svc.getStateHistory(string(rtID))

	if len(history) != 3 {
		t.Errorf("Expected 3 state transitions, got %d", len(history))
	}
	if history[0].From != "" || history[0].To != "idle" {
		t.Errorf("Expected first transition to idle, got %s -> %s", history[0].From, history[0].To)
	}
	if history[1].From != "idle" || history[1].To != "running" {
		t.Errorf("Expected second transition idle -> running, got %s -> %s", history[1].From, history[1].To)
	}
	if history[2].From != "running" || history[2].To != "stopped" {
		t.Errorf("Expected third transition running -> stopped, got %s -> %s", history[2].From, history[2].To)
	}
}

func TestLifecycleService_HotReload(t *testing.T) {
	svc := NewLifecycleServiceWithoutEngine()

	def := statechart.ChartDefinition{
		ID:           "test-chart",
		Version:      "1.0.0",
		InitialState: "idle",
	}

	rtID, err := svc.Spawn(def)
	if err != nil {
		t.Fatalf("Spawn failed: %v", err)
	}

	err = svc.HotReload(string(rtID))
	if err != nil {
		t.Errorf("Expected HotReload to return nil, got %v", err)
	}
}

func TestLifecycleService_HotReloadStatePreservation(t *testing.T) {
	svc := NewLifecycleServiceWithoutEngine()

	def := statechart.ChartDefinition{
		ID:           "test-chart",
		Version:      "1.0.0",
		InitialState: "idle",
	}

	rtID, err := svc.Spawn(def)
	if err != nil {
		t.Fatalf("Spawn failed: %v", err)
	}

	svc.updateRuntimeState(string(rtID), "running")

	err = svc.preserveState(string(rtID))
	if err != nil {
		t.Errorf("Expected preserveState to return nil, got %v", err)
	}

	savedState := svc.getSavedState(string(rtID))
	if savedState != "running" {
		t.Errorf("Expected saved state 'running', got %s", savedState)
	}
}

func TestLifecycleService_HotReloadFailure(t *testing.T) {
	svc := NewLifecycleServiceWithoutEngine()

	err := svc.HotReload("non-existent-runtime")
	if err == nil {
		t.Error("Expected HotReload to return error for non-existent runtime")
	}
	if err != statechart.ErrRuntimeNotFound {
		t.Errorf("Expected ErrRuntimeNotFound, got %v", err)
	}
}

func TestLifecycleService_HotReloadRollback(t *testing.T) {
	svc := NewLifecycleServiceWithoutEngine()

	def := statechart.ChartDefinition{
		ID:           "test-chart",
		Version:      "1.0.0",
		InitialState: "idle",
	}

	rtID, err := svc.Spawn(def)
	if err != nil {
		t.Fatalf("Spawn failed: %v", err)
	}

	svc.updateRuntimeState(string(rtID), "running")

	err = svc.preserveState(string(rtID))
	if err != nil {
		t.Fatalf("Expected preserveState to return nil, got %v", err)
	}

	svc.updateRuntimeState(string(rtID), "failed")

	err = svc.rollbackReload(string(rtID))
	if err != nil {
		t.Errorf("Expected rollbackReload to return nil, got %v", err)
	}

	list, _ := svc.List()
	if len(list) != 1 {
		t.Errorf("Expected 1 runtime, got %d", len(list))
	}
	if list[0].ActiveStates[0] != "running" {
		t.Errorf("Expected state restored to 'running', got %s", list[0].ActiveStates[0])
	}
}

func TestHotReload_QuiescenceEmptyQueue(t *testing.T) {
	svc := NewLifecycleServiceWithoutEngine()

	def := statechart.ChartDefinition{
		ID:           "test-chart",
		Version:      "1.0.0",
		InitialState: "idle",
	}

	rtID, err := svc.Spawn(def)
	if err != nil {
		t.Fatalf("Spawn failed: %v", err)
	}

	isQuiescent, err := svc.checkQuiescence(string(rtID))
	if err != nil {
		t.Fatalf("checkQuiescence should return nil error for empty queue, got: %v", err)
	}
	if !isQuiescent {
		t.Error("Expected runtime to be quiescent with empty event queue")
	}
}

func TestHotReload_QuiescenceNoActiveRegions(t *testing.T) {
	svc := NewLifecycleServiceWithoutEngine()

	def := statechart.ChartDefinition{
		ID:           "test-chart",
		Version:      "1.0.0",
		InitialState: "idle",
	}

	rtID, err := svc.Spawn(def)
	if err != nil {
		t.Fatalf("Spawn failed: %v", err)
	}

	isQuiescent, err := svc.checkQuiescence(string(rtID))
	if err != nil {
		t.Fatalf("checkQuiescence should return nil error, got: %v", err)
	}
	if !isQuiescent {
		t.Error("Expected runtime to be quiescent with no active parallel regions")
	}
}

func TestHotReload_QuiescenceNoInflightTools(t *testing.T) {
	svc := NewLifecycleServiceWithoutEngine()

	def := statechart.ChartDefinition{
		ID:           "test-chart",
		Version:      "1.0.0",
		InitialState: "idle",
	}

	rtID, err := svc.Spawn(def)
	if err != nil {
		t.Fatalf("Spawn failed: %v", err)
	}

	isQuiescent, err := svc.checkQuiescence(string(rtID))
	if err != nil {
		t.Fatalf("checkQuiescence should return nil error, got: %v", err)
	}
	if !isQuiescent {
		t.Error("Expected runtime to be quiescent with no inflight tool calls")
	}
}

func TestHotReload_PrepareForReload(t *testing.T) {
	svc := NewLifecycleServiceWithoutEngine()

	def := statechart.ChartDefinition{
		ID:           "test-chart",
		Version:      "1.0.0",
		InitialState: "idle",
	}

	rtID, err := svc.Spawn(def)
	if err != nil {
		t.Fatalf("Spawn failed: %v", err)
	}

	err = svc.prepareForReload(string(rtID), 1000)
	if err != nil {
		t.Fatalf("prepareForReload should return nil error for quiescent runtime, got: %v", err)
	}
}

func TestHotReload_ShallowHistory(t *testing.T) {
	svc := NewLifecycleServiceWithoutEngine()

	snapshot := statechart.Snapshot{
		RuntimeID:      statechart.RuntimeID("original-runtime"),
		DefinitionID:   "test-chart",
		ActiveStates:   []string{"idle"},
		RuntimeContext: statechart.RuntimeContext{ChartID: "test-chart", RuntimeID: "original-runtime"},
	}

	rtID, err := svc.restoreWithShallowHistory(snapshot)
	if err != nil {
		t.Fatalf("restoreWithShallowHistory should return nil error, got: %v", err)
	}

	if rtID == "" {
		t.Error("Expected non-empty RuntimeID")
	}

	list, err := svc.List()
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}

	if len(list) != 1 {
		t.Errorf("Expected 1 runtime, got %d", len(list))
	}

	if list[0].ActiveStates[0] != "idle" {
		t.Errorf("Expected state 'idle', got %s", list[0].ActiveStates[0])
	}
}

func TestHotReload_DeepHistory(t *testing.T) {
	svc := NewLifecycleServiceWithoutEngine()

	def := statechart.ChartDefinition{
		ID:           "test-chart",
		Version:      "1.0.0",
		InitialState: "idle",
		Root: &statechart.Node{
			ID: "root",
			Children: map[string]*statechart.Node{
				"idle": {
					ID: "idle",
					Children: map[string]*statechart.Node{
						"sub-idle": {ID: "sub-idle"},
					},
				},
			},
		},
	}

	targetState := "idle/sub-idle"

	snapshot := statechart.Snapshot{
		RuntimeID:      statechart.RuntimeID("original-runtime"),
		DefinitionID:   "test-chart",
		ActiveStates:   []string{"idle"},
		RuntimeContext: statechart.RuntimeContext{ChartID: "test-chart", RuntimeID: "original-runtime"},
		RegionStates:   map[string]string{"idle": "sub-idle"},
	}

	rtID, err := svc.restoreWithDeepHistory(snapshot, targetState, def)
	if err != nil {
		t.Fatalf("restoreWithDeepHistory should return nil error, got: %v", err)
	}

	if rtID == "" {
		t.Error("Expected non-empty RuntimeID")
	}

	list, err := svc.List()
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}

	if len(list) != 1 {
		t.Errorf("Expected 1 runtime, got %d", len(list))
	}

	if list[0].ActiveStates[0] != "sub-idle" {
		t.Errorf("Expected state 'sub-idle', got %s", list[0].ActiveStates[0])
	}
}

func TestHotReload_DeletedStateFallback(t *testing.T) {
	svc := NewLifecycleServiceWithoutEngine()

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

	targetState := "deleted-state/path"

	snapshot := statechart.Snapshot{
		RuntimeID:      statechart.RuntimeID("original-runtime"),
		DefinitionID:   "test-chart",
		ActiveStates:   []string{"idle"},
		RuntimeContext: statechart.RuntimeContext{ChartID: "test-chart", RuntimeID: "original-runtime"},
		RegionStates:   map[string]string{"idle": "deleted-state"},
	}

	rtID, err := svc.restoreWithDeepHistory(snapshot, targetState, def)
	if err != nil {
		t.Fatalf("restoreWithDeepHistory should return nil error, got: %v", err)
	}

	if rtID == "" {
		t.Error("Expected non-empty RuntimeID")
	}

	list, err := svc.List()
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}

	if len(list) != 1 {
		t.Errorf("Expected 1 runtime, got %d", len(list))
	}

	if list[0].ActiveStates[0] != "idle" {
		t.Errorf("Expected fallback to state 'idle', got %s", list[0].ActiveStates[0])
	}
}

func TestHotReload_HistoryPreservation(t *testing.T) {
	svc := NewLifecycleServiceWithoutEngine()

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

	originalRtID, err := svc.Spawn(def)
	if err != nil {
		t.Fatalf("Spawn failed: %v", err)
	}

	svc.updateRuntimeState(string(originalRtID), "running")
	svc.updateRuntimeState(string(originalRtID), "stopped")

	originalHistory := svc.getStateHistory(string(originalRtID))
	if len(originalHistory) != 3 {
		t.Fatalf("Expected 3 state transitions, got %d", len(originalHistory))
	}

	snapshot := statechart.Snapshot{
		RuntimeID:      originalRtID,
		DefinitionID:   "test-chart",
		ActiveStates:   []string{"stopped"},
		RuntimeContext: statechart.RuntimeContext{ChartID: "test-chart", RuntimeID: string(originalRtID)},
	}

	newRtID, err := svc.restoreWithShallowHistory(snapshot)
	if err != nil {
		t.Fatalf("restoreWithShallowHistory should return nil error, got: %v", err)
	}

	if newRtID == "" {
		t.Error("Expected non-empty RuntimeID")
	}

	newHistory := svc.getStateHistory(string(newRtID))
	if len(newHistory) != 1 {
		t.Errorf("Expected 1 state transition in new runtime, got %d", len(newHistory))
	}

	if newHistory[0].To != "stopped" {
		t.Errorf("Expected state 'stopped', got %s", newHistory[0].To)
	}

	list, err := svc.List()
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}

	if len(list) != 2 {
		t.Errorf("Expected 2 runtimes, got %d", len(list))
	}
}

func TestHotReload_ContextTransform(t *testing.T) {
	svc := NewLifecycleServiceWithoutEngine()

	oldContext := map[string]any{
		"userId":   "user-123",
		"state":    "idle",
		"version":  "1.0.0",
		"settings": map[string]any{"theme": "dark"},
	}

	newVersion := "2.0.0"

	template := `{"userId":"{{GetMapValue .OldContext "userId"}}","state":"{{GetMapValue .OldContext "state"}}","newVersion":"{{.NewVersion}}","contextVersion":"{{.ContextVersion}}"}`

	newContext, err := svc.applyContextTransform(oldContext, newVersion, template)
	if err != nil {
		t.Fatalf("applyContextTransform should return nil error, got: %v", err)
	}

	if newContext == nil {
		t.Error("Expected non-nil new context")
	}

	newContextMap, ok := newContext.(map[string]any)
	if !ok {
		t.Error("Expected new context to be map[string]any")
	}

	if newContextMap["userId"] != "user-123" {
		t.Errorf("Expected userId 'user-123', got %v", newContextMap["userId"])
	}

	if newContextMap["state"] != "idle" {
		t.Errorf("Expected state 'idle', got %v", newContextMap["state"])
	}

	if newContextMap["newVersion"] != "2.0.0" {
		t.Errorf("Expected newVersion '2.0.0', got %v", newContextMap["newVersion"])
	}

	if newContextMap["contextVersion"] != "2.0.0" {
		t.Errorf("Expected contextVersion '2.0.0', got %v", newContextMap["contextVersion"])
	}
}

func TestHotReload_TransformFailureFallback(t *testing.T) {
	svc := NewLifecycleServiceWithoutEngine()

	oldContext := map[string]any{
		"userId": "user-123",
		"state":  "idle",
	}

	newVersion := "2.0.0"

	invalidTemplate := `{{invalid syntax that will fail`

	cleanStartCalled := false
	cleanContext := map[string]any{"version": newVersion}

	_, err := svc.applyContextTransformWithFallback(oldContext, newVersion, invalidTemplate, func() (any, error) {
		cleanStartCalled = true
		return cleanContext, nil
	})

	if err != nil {
		t.Fatalf("applyContextTransformWithFallback should return nil error after fallback, got: %v", err)
	}

	if !cleanStartCalled {
		t.Error("Expected cleanStart to be called on transform failure")
	}
}

func TestHotReload_TemplateValidation(t *testing.T) {
	svc := NewLifecycleServiceWithoutEngine()

	invalidTemplate := `{{invalid syntax that will fail`

	err := svc.validateTransformTemplate(invalidTemplate)
	if err == nil {
		t.Error("Expected validateTransformTemplate to return error for invalid syntax")
	}

	validTemplate := `{"key":"{{GetMapValue .OldContext "key"}}","version":"{{.NewVersion}}"}`

	err = svc.validateTransformTemplate(validTemplate)
	if err != nil {
		t.Errorf("Expected validateTransformTemplate to return nil for valid template, got: %v", err)
	}
}

func TestHardcodedServices_LifecycleSpawnTracking(t *testing.T) {
	svc := NewLifecycleServiceWithoutEngine()

	def := statechart.ChartDefinition{
		ID:           "test-agent",
		Version:      "1.0.0",
		InitialState: "idle",
	}

	runtimeID, err := svc.Spawn(def)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	runtimes, err := svc.List()
	if err != nil {
		t.Fatalf("Expected no error from List, got %v", err)
	}
	if len(runtimes) != 1 {
		t.Fatalf("Expected 1 runtime, got %d", len(runtimes))
	}

	info := runtimes[0]
	if info.ID != string(runtimeID) {
		t.Errorf("Expected runtime ID '%s', got '%s'", runtimeID, info.ID)
	}
	if info.DefinitionID != "test-agent" {
		t.Errorf("Expected DefinitionID 'test-agent', got '%s'", info.DefinitionID)
	}
	if info.Boundary != mail.InnerBoundary {
		t.Errorf("Expected Boundary InnerBoundary, got %s", info.Boundary)
	}
}
