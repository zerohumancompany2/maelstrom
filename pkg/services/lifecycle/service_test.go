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
