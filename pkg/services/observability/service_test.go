package observability

import (
	"testing"
	"time"

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

func TestObservabilityService_QueryTracesByEventType(t *testing.T) {
	svc := NewObservabilityService()

	trace1 := services.Trace{
		ID:        "trace-1",
		RuntimeID: "runtime-1",
		EventType: "transition",
		StatePath: "state/1",
	}
	trace2 := services.Trace{
		ID:        "trace-2",
		RuntimeID: "runtime-1",
		EventType: "entry",
		StatePath: "state/2",
	}
	trace3 := services.Trace{
		ID:        "trace-3",
		RuntimeID: "runtime-1",
		EventType: "transition",
		StatePath: "state/3",
	}
	svc.EmitTrace(trace1)
	svc.EmitTrace(trace2)
	svc.EmitTrace(trace3)

	filters := services.TraceFilters{EventType: "transition"}
	traces, err := svc.QueryTraces(filters)

	if err != nil {
		t.Errorf("Expected nil error, got %v", err)
	}
	if len(traces) != 2 {
		t.Errorf("Expected 2 transition events, got %d", len(traces))
	}
}

func TestObservabilityService_QueryTracesByTimeRange(t *testing.T) {
	svc := NewObservabilityService()

	now := time.Now()
	earlier := now.Add(-2 * time.Hour)
	within := now.Add(-1 * time.Hour)
	later := now.Add(1 * time.Hour)

	trace1 := services.Trace{
		ID:        "trace-1",
		RuntimeID: "runtime-1",
		EventType: "transition",
		StatePath: "state/1",
		Timestamp: earlier,
	}
	trace2 := services.Trace{
		ID:        "trace-2",
		RuntimeID: "runtime-1",
		EventType: "entry",
		StatePath: "state/2",
		Timestamp: within,
	}
	trace3 := services.Trace{
		ID:        "trace-3",
		RuntimeID: "runtime-1",
		EventType: "exit",
		StatePath: "state/3",
		Timestamp: later,
	}
	svc.EmitTrace(trace1)
	svc.EmitTrace(trace2)
	svc.EmitTrace(trace3)

	filters := services.TraceFilters{
		FromTime: earlier,
		ToTime:   within,
	}
	traces, err := svc.QueryTraces(filters)

	if err != nil {
		t.Errorf("Expected nil error, got %v", err)
	}
	if len(traces) != 2 {
		t.Errorf("Expected 2 traces in time range, got %d", len(traces))
	}
}

func TestObservabilityService_LogDeadLetter(t *testing.T) {
	svc := NewObservabilityService()

	testMail := mail.Mail{
		ID:     "mail-123",
		Source: "test-source",
		Target: "test-target",
		Type:   mail.Error,
	}

	err := svc.LogDeadLetter(testMail, "max retries exceeded")

	if err != nil {
		t.Errorf("Expected nil error, got %v", err)
	}

	entries, err := svc.QueryDeadLetters()
	if err != nil {
		t.Errorf("Expected nil error from QueryDeadLetters, got %v", err)
	}
	if len(entries) != 1 {
		t.Errorf("Expected 1 dead letter entry, got %d", len(entries))
	}
	if entries[0].Reason != "max retries exceeded" {
		t.Errorf("Expected reason 'max retries exceeded', got %s", entries[0].Reason)
	}
	if entries[0].Logged.IsZero() {
		t.Error("Expected Logged timestamp to be set")
	}
}

func TestObservabilityService_QueryDeadLetters(t *testing.T) {
	svc := NewObservabilityService()

	entries, err := svc.QueryDeadLetters()
	if err != nil {
		t.Errorf("Expected nil error, got %v", err)
	}
	if len(entries) != 0 {
		t.Errorf("Expected 0 dead letter entries, got %d", len(entries))
	}

	mail1 := mail.Mail{ID: "mail-1", Source: "src-1", Target: "tgt-1", Type: mail.Error}
	mail2 := mail.Mail{ID: "mail-2", Source: "src-2", Target: "tgt-2", Type: mail.Error}
	svc.LogDeadLetter(mail1, "reason-1")
	svc.LogDeadLetter(mail2, "reason-2")

	entries, err = svc.QueryDeadLetters()
	if err != nil {
		t.Errorf("Expected nil error, got %v", err)
	}
	if len(entries) != 2 {
		t.Errorf("Expected 2 dead letter entries, got %d", len(entries))
	}
	if entries[0].Mail.ID != "mail-1" {
		t.Errorf("Expected first mail ID 'mail-1', got %s", entries[0].Mail.ID)
	}
	if entries[1].Reason != "reason-2" {
		t.Errorf("Expected second reason 'reason-2', got %s", entries[1].Reason)
	}
}
