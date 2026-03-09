package observability

import (
	"testing"

	"github.com/maelstrom/v3/pkg/mail"
	"github.com/maelstrom/v3/pkg/services"
)

func TestObservabilityService_NewObservabilityServiceReturnsNonNil(t *testing.T) {
	svc := NewObservabilityService()

	if svc == nil {
		t.Error("Expected NewObservabilityService to return non-nil")
	}
}

func TestObservabilityService_IDReturnsCorrectString(t *testing.T) {
	svc := NewObservabilityService()

	id := svc.ID()

	if id != "sys:observability" {
		t.Errorf("Expected ID sys:observability, got %s", id)
	}
}

func TestObservabilityService_HandleMailReturnsNil(t *testing.T) {
	svc := NewObservabilityService()

	err := svc.HandleMail(mail.Mail{})

	if err != nil {
		t.Errorf("Expected HandleMail to return nil, got %v", err)
	}
}

func TestObservabilityService_EmitTraceReturnsNil(t *testing.T) {
	svc := NewObservabilityService()

	err := svc.EmitTrace(services.Trace{})

	if err != nil {
		t.Errorf("Expected EmitTrace to return nil, got %v", err)
	}
}

func TestObservabilityService_QueryTracesReturnsNilNil(t *testing.T) {
	svc := NewObservabilityService()

	filters := services.TraceFilters{RuntimeID: "runtime-123"}
	traces, err := svc.QueryTraces(filters)

	if err != nil {
		t.Errorf("Expected QueryTraces to return nil error, got %v", err)
	}

	if traces != nil {
		t.Error("Expected QueryTraces to return nil slice")
	}
}

func TestObservabilityService_StartReturnsNil(t *testing.T) {
	svc := NewObservabilityService()

	err := svc.Start()

	if err != nil {
		t.Errorf("Expected Start to return nil, got %v", err)
	}
}

func TestObservabilityService_StopReturnsNil(t *testing.T) {
	svc := NewObservabilityService()

	err := svc.Stop()

	if err != nil {
		t.Errorf("Expected Stop to return nil, got %v", err)
	}
}

func TestObservabilityService_BootstrapChart(t *testing.T) {
	chart := BootstrapChart()

	if chart.ID != "sys:observability" {
		t.Errorf("Expected ID sys:observability, got %s", chart.ID)
	}

	if chart.Version != "1.0.0" {
		t.Errorf("Expected version 1.0.0, got %s", chart.Version)
	}
}

func TestObservabilityService_EmitTrace(t *testing.T) {
	svc := NewObservabilityService()
	trace := services.Trace{
		ID:        "test-trace-1",
		RuntimeID: "test-runtime",
		EventType: "transition",
		StatePath: "root/child",
	}
	err := svc.EmitTrace(trace)
	if err != nil {
		t.Errorf("EmitTrace should return nil, got: %v", err)
	}
	filters := services.TraceFilters{RuntimeID: "test-runtime"}
	traces, err := svc.QueryTraces(filters)
	if err != nil {
		t.Errorf("QueryTraces should return nil error, got: %v", err)
	}
	if len(traces) != 1 {
		t.Errorf("Expected 1 trace, got %d", len(traces))
	}
	if traces[0].ID != "test-trace-1" {
		t.Errorf("Expected trace ID 'test-trace-1', got: %s", traces[0].ID)
	}
}

func TestObservabilityService_BoundaryInner(t *testing.T) {
	svc := NewObservabilityService()
	if svc.Boundary() != mail.InnerBoundary {
		t.Errorf("Expected boundary 'inner', got: %v", svc.Boundary())
	}
}

func TestObservabilityService_ID(t *testing.T) {
	chart := BootstrapChart()

	if chart.ID != "sys:observability" {
		t.Errorf("Expected ID sys:observability, got %s", chart.ID)
	}
}

func TestObservabilityService_EmptyFiltersReturnsAll(t *testing.T) {
	svc := NewObservabilityService()

	trace1 := services.Trace{
		ID:        "trace-1",
		RuntimeID: "runtime-1",
		EventType: "transition",
		StatePath: "state/1",
	}
	trace2 := services.Trace{
		ID:        "trace-2",
		RuntimeID: "runtime-2",
		EventType: "entry",
		StatePath: "state/2",
	}
	svc.EmitTrace(trace1)
	svc.EmitTrace(trace2)

	filters := services.TraceFilters{}
	traces, err := svc.QueryTraces(filters)

	if err != nil {
		t.Errorf("Expected nil error, got %v", err)
	}
	if len(traces) != 2 {
		t.Errorf("Expected 2 traces, got %d", len(traces))
	}
}

func TestObservabilityService_QueryTracesWithFilters(t *testing.T) {
	svc := NewObservabilityService()

	trace1 := services.Trace{
		ID:        "trace-1",
		RuntimeID: "runtime-a",
		EventType: "transition",
		StatePath: "state/1",
	}
	trace2 := services.Trace{
		ID:        "trace-2",
		RuntimeID: "runtime-a",
		EventType: "entry",
		StatePath: "state/2",
	}
	trace3 := services.Trace{
		ID:        "trace-3",
		RuntimeID: "runtime-b",
		EventType: "transition",
		StatePath: "state/3",
	}
	svc.EmitTrace(trace1)
	svc.EmitTrace(trace2)
	svc.EmitTrace(trace3)

	filters := services.TraceFilters{RuntimeID: "runtime-a"}
	traces, err := svc.QueryTraces(filters)

	if err != nil {
		t.Errorf("Expected nil error, got %v", err)
	}
	if len(traces) != 2 {
		t.Errorf("Expected 2 traces for runtime-a, got %d", len(traces))
	}
}
