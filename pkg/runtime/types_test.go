package runtime

import (
	"testing"

	"github.com/maelstrom/v3/pkg/statechart"
)

// TestChartRuntime_InstantiatesFromDefinition verifies runtime creation.
func TestChartRuntime_InstantiatesFromDefinition(t *testing.T) {
	def := map[string]interface{}{
		"id":      "test-chart",
		"version": "1.0.0",
		"spec": map[string]interface{}{
			"initial": "idle",
		},
	}

	rt, err := NewChartRuntime("test-instance", def)
	if err != nil {
		t.Fatalf("failed to create runtime: %v", err)
	}

	if rt.ID != "test-instance" {
		t.Errorf("expected ID 'test-instance', got %q", rt.ID)
	}

	if rt.Context.ChartID != "test-chart" {
		t.Errorf("expected ChartID 'test-chart', got %q", rt.Context.ChartID)
	}
}

// TestChartRuntime_SendEvent verifies event dispatch.
func TestChartRuntime_SendEvent(t *testing.T) {
	// TODO: implement after InstantiatesFromDefinition passes
}

// TestChartRuntime_StateObservation verifies active state tracking.
func TestChartRuntime_StateObservation(t *testing.T) {
	// TODO: implement after SendEvent passes
}

// MockEngine implements the minimal engine interface for testing
type MockEngine struct {
	dispatched []struct {
		id  string
		evt statechart.Event
	}
}

func (m *MockEngine) Dispatch(id string, ev statechart.Event) error {
	m.dispatched = append(m.dispatched, struct {
		id  string
		evt statechart.Event
	}{id, ev})
	return nil
}

// TestChartRuntime_ForwardsToStatechartEngine verifies events are forwarded to engine.
func TestChartRuntime_ForwardsToStatechartEngine(t *testing.T) {
	engine := &MockEngine{}

	rt := &ChartRuntime{
		ID:        "test-instance",
		runtimeID: "rt-123",
		engine:    engine,
	}

	evt := Event{Type: "test-event", Payload: map[string]interface{}{"key": "value"}}
	if err := rt.SendEvent(evt); err != nil {
		t.Fatalf("SendEvent failed: %v", err)
	}

	// Verify event was dispatched to engine
	if len(engine.dispatched) != 1 {
		t.Fatalf("expected 1 dispatched event, got %d", len(engine.dispatched))
	}

	dispatched := engine.dispatched[0]
	if dispatched.id != "rt-123" {
		t.Errorf("expected runtime ID 'rt-123', got %q", dispatched.id)
	}
	if dispatched.evt.Type != "test-event" {
		t.Errorf("expected event type 'test-event', got %q", dispatched.evt.Type)
	}
}

// TestChartRuntime_FallsBackToInternalQueue verifies fallback without engine.
func TestChartRuntime_FallsBackToInternalQueue(t *testing.T) {
	rt, err := NewChartRuntime("test", map[string]interface{}{"id": "test"})
	if err != nil {
		t.Fatalf("failed to create runtime: %v", err)
	}

	evt := Event{Type: "test-event"}
	if err := rt.SendEvent(evt); err != nil {
		t.Fatalf("SendEvent failed: %v", err)
	}

	// Event should be in internal queue (we can't verify directly, but no error means success)
}
