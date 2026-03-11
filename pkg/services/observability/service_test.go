package observability

import (
	"fmt"
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

	outcome := svc.HandleMail(mail.Mail{})

	if outcome.Status != "success" {
		t.Errorf("Expected HandleMail to return success, got %s", outcome.Status)
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

func TestObservabilityService_GetMetrics(t *testing.T) {
	svc := NewObservabilityService()

	metrics := svc.GetMetrics()

	if metrics.StateCounts == nil {
		t.Error("Expected StateCounts to be initialized, got nil")
	}
	if metrics.LastUpdate.IsZero() {
		t.Error("Expected LastUpdate to be set, got zero time")
	}
}

func TestObservabilityService_MetricsUpdateOnTrace(t *testing.T) {
	svc := NewObservabilityService()

	trace1 := services.Trace{
		ID:        "trace-1",
		RuntimeID: "runtime-1",
		EventType: "transition",
		StatePath: "root/state-a",
	}
	trace2 := services.Trace{
		ID:        "trace-2",
		RuntimeID: "runtime-1",
		EventType: "entry",
		StatePath: "root/state-a",
	}
	trace3 := services.Trace{
		ID:        "trace-3",
		RuntimeID: "runtime-1",
		EventType: "exit",
		StatePath: "root/state-b",
	}

	svc.EmitTrace(trace1)
	metrics := svc.GetMetrics()
	if metrics.StateCounts["root/state-a"] != 1 {
		t.Errorf("Expected StateCounts['root/state-a'] = 1, got %d", metrics.StateCounts["root/state-a"])
	}

	svc.EmitTrace(trace2)
	metrics = svc.GetMetrics()
	if metrics.StateCounts["root/state-a"] != 2 {
		t.Errorf("Expected StateCounts['root/state-a'] = 2, got %d", metrics.StateCounts["root/state-a"])
	}

	svc.EmitTrace(trace3)
	metrics = svc.GetMetrics()
	if metrics.StateCounts["root/state-b"] != 1 {
		t.Errorf("Expected StateCounts['root/state-b'] = 1, got %d", metrics.StateCounts["root/state-b"])
	}
}

func TestObservabilityService_QueryDeadLettersWithFilters(t *testing.T) {
	svc := NewObservabilityService()

	mail1 := mail.Mail{ID: "mail-1", Source: "src-1", Target: "tgt-1", Type: mail.Error}
	mail2 := mail.Mail{ID: "mail-2", Source: "src-2", Target: "tgt-2", Type: mail.Error}
	mail3 := mail.Mail{ID: "mail-3", Source: "src-3", Target: "tgt-3", Type: mail.Error}
	svc.LogDeadLetter(mail1, "reason-a")
	svc.LogDeadLetter(mail2, "reason-b")
	svc.LogDeadLetter(mail3, "reason-a")

	filters := &DeadLetterFilters{Reason: "reason-a"}
	entries := svc.QueryDeadLettersWithFilters(filters)

	if len(entries) != 2 {
		t.Errorf("Expected 2 dead letter entries with reason-a, got %d", len(entries))
	}
	if entries[0].Reason != "reason-a" {
		t.Errorf("Expected reason 'reason-a', got %s", entries[0].Reason)
	}
	if entries[1].Reason != "reason-a" {
		t.Errorf("Expected reason 'reason-a', got %s", entries[1].Reason)
	}
}

func TestObservabilityService_TransitionRate(t *testing.T) {
	svc := NewObservabilityService()

	svc.trackTransition("idle", "running")
	svc.trackTransition("running", "stopped")
	svc.trackTransition("idle", "running")

	metrics := svc.GetMetrics()

	if metrics.TransitionRate != 3 {
		t.Errorf("Expected TransitionRate 3, got %f", metrics.TransitionRate)
	}
}

func TestObservabilityService_EventRate(t *testing.T) {
	svc := NewObservabilityService()

	svc.trackEvent("transition")
	svc.trackEvent("entry")
	svc.trackEvent("exit")

	metrics := svc.GetMetrics()

	if metrics.EventRate != 3 {
		t.Errorf("Expected EventRate 3, got %f", metrics.EventRate)
	}
}

func TestObservabilityService_MetricsCollector(t *testing.T) {
	svc := NewObservabilityService()

	metrics := svc.GetMetrics()

	if metrics.StateCounts == nil {
		t.Error("Expected StateCounts to be non-nil")
	}
	if metrics.LastUpdate.IsZero() {
		t.Error("Expected LastUpdate to be set")
	}
	if metrics.TransitionRate != 0 {
		t.Errorf("Expected TransitionRate 0, got %f", metrics.TransitionRate)
	}
	if metrics.EventRate != 0 {
		t.Errorf("Expected EventRate 0, got %f", metrics.EventRate)
	}
}

func TestObservabilityService_MetricsAggregation(t *testing.T) {
	svc := NewObservabilityService()

	svc.trackTransition("idle", "running")
	svc.trackEvent("transition")

	aggregated := svc.aggregateMetrics(5 * time.Minute)

	if aggregated.TransitionRate != 1 {
		t.Errorf("Expected TransitionRate 1, got %f", aggregated.TransitionRate)
	}
	if aggregated.EventRate != 1 {
		t.Errorf("Expected EventRate 1, got %f", aggregated.EventRate)
	}
}

func TestObservabilityService_QueryDeadLettersNoCopy(t *testing.T) {
	svc := NewObservabilityService()

	mail1 := mail.Mail{ID: "mail-1", Source: "src-1", Target: "tgt-1", Type: mail.Error}
	svc.LogDeadLetter(mail1, "reason-1")

	filters := &DeadLetterFilters{Reason: "reason-1"}
	entries := svc.QueryDeadLettersNoCopy(filters)

	if len(entries) != 1 {
		t.Errorf("Expected 1 entry, got %d", len(entries))
	}
	if entries[0].Mail.ID != "mail-1" {
		t.Errorf("Expected mail ID mail-1, got %s", entries[0].Mail.ID)
	}
}

func TestObservabilityService_QueryDeadLettersMemory(t *testing.T) {
	svc := NewObservabilityService()

	for i := 0; i < 100; i++ {
		mail := mail.Mail{ID: fmt.Sprintf("mail-%d", i), Source: "src", Target: "tgt", Type: mail.Error}
		svc.LogDeadLetter(mail, "reason")
	}

	memoryUsage := svc.getMemoryUsage()

	if memoryUsage == 0 {
		t.Error("Expected memory usage to be greater than 0")
	}
}

func TestObservabilityService_QueryDeadLettersLargeSet(t *testing.T) {
	svc := NewObservabilityService()

	for i := 0; i < 10000; i++ {
		mail := mail.Mail{ID: fmt.Sprintf("mail-%d", i), Source: "src", Target: "tgt", Type: mail.Error}
		svc.LogDeadLetter(mail, "reason")
	}

	start := time.Now()
	filters := &DeadLetterFilters{Reason: "reason"}
	entries := svc.QueryDeadLettersNoCopy(filters)
	duration := time.Since(start)

	if len(entries) != 10000 {
		t.Errorf("Expected 10000 entries, got %d", len(entries))
	}
	if duration > 1*time.Second {
		t.Errorf("Query took too long: %v", duration)
	}
}

func TestHardcodedServices_ObservabilityDeadLetter(t *testing.T) {
	svc := NewObservabilityService()

	failedMail := mail.Mail{
		ID:        "failed-001",
		Type:      mail.MailTypeError,
		Source:    "agent:failed",
		Target:    "agent:target",
		Content:   map[string]any{"error": "processing failed"},
		CreatedAt: time.Now(),
	}

	svc.LogDeadLetter(failedMail, "Max retries exceeded")

	entries, _ := svc.QueryDeadLetters()
	if len(entries) != 1 {
		t.Fatalf("Expected 1 dead-letter entry, got %d", len(entries))
	}

	entry := entries[0]
	if entry.Mail.ID != "failed-001" {
		t.Errorf("Expected MailID 'failed-001', got '%s'", entry.Mail.ID)
	}
	if entry.Reason != "Max retries exceeded" {
		t.Errorf("Expected Reason 'Max retries exceeded', got '%s'", entry.Reason)
	}
	if entry.Logged.IsZero() {
		t.Error("Expected Logged timestamp to be set")
	}
}

func TestObservabilityService_MetricsCollectionStateCounts(t *testing.T) {
	svc := NewObservabilityService()

	trace1 := services.Trace{
		ID:        "trace-1",
		RuntimeID: "runtime-1",
		EventType: "transition",
		StatePath: "root/observe",
	}
	trace2 := services.Trace{
		ID:        "trace-2",
		RuntimeID: "runtime-1",
		EventType: "entry",
		StatePath: "root/observe",
	}
	trace3 := services.Trace{
		ID:        "trace-3",
		RuntimeID: "runtime-2",
		EventType: "transition",
		StatePath: "root/orient",
	}

	svc.EmitTrace(trace1)
	svc.EmitTrace(trace2)
	svc.EmitTrace(trace3)

	metrics := svc.GetMetrics()

	if metrics.StateCounts["root/observe"] != 2 {
		t.Errorf("Expected StateCounts['root/observe'] = 2, got %d", metrics.StateCounts["root/observe"])
	}
	if metrics.StateCounts["root/orient"] != 1 {
		t.Errorf("Expected StateCounts['root/orient'] = 1, got %d", metrics.StateCounts["root/orient"])
	}
}

func TestObservabilityService_MetricsCollectionTransitionRates(t *testing.T) {
	svc := NewObservabilityService()

	time.Sleep(10 * time.Millisecond)

	svc.trackTransition("idle", "running")
	svc.trackTransition("running", "processing")
	svc.trackTransition("processing", "idle")

	time.Sleep(10 * time.Millisecond)

	metrics := svc.GetMetrics()

	if metrics.TransitionRate != 3 {
		t.Errorf("Expected TransitionRate 3, got %f", metrics.TransitionRate)
	}
}

func TestObservabilityService_MetricsCollectionEventRates(t *testing.T) {
	svc := NewObservabilityService()

	svc.trackEvent("user_query")
	svc.trackEvent("heartbeat")
	svc.trackEvent("llm_decision")
	svc.trackEvent("orchestration_complete")

	metrics := svc.GetMetrics()

	if metrics.EventRate != 4 {
		t.Errorf("Expected EventRate 4, got %f", metrics.EventRate)
	}
}

func TestObservabilityService_MetricsCollectionMailDelivery(t *testing.T) {
	svc := NewObservabilityService()

	mail1 := mail.Mail{ID: "mail-1", Source: "agent-1", Target: "agent-2", Type: mail.Assistant}
	mail2 := mail.Mail{ID: "mail-2", Source: "agent-2", Target: "agent-3", Type: mail.User}

	svc.HandleMail(mail1)
	svc.TrackMailDelivered()
	svc.TrackMailFailed()
	svc.TrackMailRetried()
	svc.HandleMail(mail2)

	metrics := svc.GetMetrics()

	if metrics.MailDelivered != 3 {
		t.Errorf("Expected MailDelivered 3, got %d", metrics.MailDelivered)
	}
	if metrics.MailFailed != 1 {
		t.Errorf("Expected MailFailed 1, got %d", metrics.MailFailed)
	}
	if metrics.MailRetried != 1 {
		t.Errorf("Expected MailRetried 1, got %d", metrics.MailRetried)
	}
}

func TestMetricsReporter_CollectReturnsMetrics(t *testing.T) {
	reporter := NewMetricsReporter()

	reporter.mu.Lock()
	reporter.collected.TransitionRate = 10
	reporter.collected.EventRate = 20
	reporter.collected.MailDelivered = 100
	reporter.collected.MailFailed = 5
	reporter.collected.MailRetried = 2
	reporter.mu.Unlock()

	metrics := reporter.Collect()

	if metrics.MailDelivered != 100 {
		t.Errorf("Expected MailDelivered 100, got %d", metrics.MailDelivered)
	}
	if metrics.MailFailed != 5 {
		t.Errorf("Expected MailFailed 5, got %d", metrics.MailFailed)
	}
	if metrics.MailRetried != 2 {
		t.Errorf("Expected MailRetried 2, got %d", metrics.MailRetried)
	}
	if metrics.TransitionRate < 10 {
		t.Errorf("Expected TransitionRate >= 10, got %f", metrics.TransitionRate)
	}
	if metrics.EventRate < 20 {
		t.Errorf("Expected EventRate >= 20, got %f", metrics.EventRate)
	}
}

func TestMetricsReporter_ReportJSON(t *testing.T) {
	reporter := NewMetricsReporter()

	reporter.mu.Lock()
	reporter.collected.StateCounts["test/state"] = 5
	reporter.collected.TransitionRate = 10
	reporter.collected.EventRate = 20
	reporter.collected.MailDelivered = 100
	reporter.collected.MailFailed = 5
	reporter.collected.MailRetried = 2
	reporter.mu.Unlock()

	err := reporter.Report("json")

	if err != nil {
		t.Errorf("Expected nil error, got %v", err)
	}
}

func TestMetricsReporter_ReportText(t *testing.T) {
	reporter := NewMetricsReporter()

	reporter.mu.Lock()
	reporter.collected.StateCounts["test/state"] = 5
	reporter.collected.TransitionRate = 10
	reporter.collected.EventRate = 20
	reporter.collected.MailDelivered = 100
	reporter.collected.MailFailed = 5
	reporter.collected.MailRetried = 2
	reporter.mu.Unlock()

	err := reporter.Report("text")

	if err != nil {
		t.Errorf("Expected nil error, got %v", err)
	}
}

func TestMetricsReporter_ReportInvalidFormat(t *testing.T) {
	reporter := NewMetricsReporter()

	err := reporter.Report("invalid")

	if err == nil {
		t.Error("Expected error for invalid format, got nil")
	}
}

func TestMetricsReporter_Reset(t *testing.T) {
	reporter := NewMetricsReporter()

	reporter.mu.Lock()
	reporter.collected.TransitionRate = 100
	reporter.collected.EventRate = 200
	reporter.collected.MailDelivered = 1000
	reporter.collected.MailFailed = 50
	reporter.collected.MailRetried = 10
	reporter.mu.Unlock()

	reporter.Reset()

	metrics := reporter.Collect()

	if metrics.TransitionRate != 0 {
		t.Errorf("Expected TransitionRate 0 after reset, got %f", metrics.TransitionRate)
	}
	if metrics.EventRate != 0 {
		t.Errorf("Expected EventRate 0 after reset, got %f", metrics.EventRate)
	}
	if metrics.MailDelivered != 0 {
		t.Errorf("Expected MailDelivered 0 after reset, got %d", metrics.MailDelivered)
	}
	if metrics.MailFailed != 0 {
		t.Errorf("Expected MailFailed 0 after reset, got %d", metrics.MailFailed)
	}
	if metrics.MailRetried != 0 {
		t.Errorf("Expected MailRetried 0 after reset, got %d", metrics.MailRetried)
	}
}
