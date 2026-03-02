package persistence

import (
	"testing"

	"github.com/maelstrom/v3/pkg/statechart"
)

func TestPersistence_SnapshotCreate(t *testing.T) {
	svc := NewPersistenceService()

	runtimeID := statechart.RuntimeID("test-runtime-1")
	snap, err := svc.Snapshot(string(runtimeID))

	if err != nil {
		t.Fatalf("Snapshot failed: %v", err)
	}

	if snap.ID == "" {
		t.Error("Expected snapshot ID to be non-empty")
	}

	if snap.RuntimeID != string(runtimeID) {
		t.Errorf("Expected RuntimeID %s, got %s", runtimeID, snap.RuntimeID)
	}

	if snap.Timestamp.IsZero() {
		t.Error("Expected snapshot timestamp to be non-zero")
	}

	if snap.State == nil {
		t.Error("Expected snapshot state to be non-nil")
	}
}

func TestPersistence_SnapshotRestore(t *testing.T) {
	svc := NewPersistenceService()

	originalID := statechart.RuntimeID("restore-test-1")

	snap, err := svc.Snapshot(string(originalID))
	if err != nil {
		t.Fatalf("Snapshot failed: %v", err)
	}

	def := statechart.ChartDefinition{
		ID:      "test-chart",
		Version: "1.0.0",
	}

	restoredID, err := svc.Restore(string(snap.ID), def)
	if err != nil {
		t.Fatalf("Restore failed: %v", err)
	}

	if restoredID == "" {
		t.Error("Expected restored ID to be non-empty")
	}

	if string(restoredID) == string(originalID) {
		t.Error("Expected restored ID to be different from original")
	}
}

func TestPersistence_EventAppend(t *testing.T) {
	svc := NewPersistenceService()

	runtimeID := statechart.RuntimeID("event-test-1")
	event := statechart.Event{
		Type:          "test:event",
		Payload:       map[string]any{"key": "value"},
		CorrelationID: "corr-123",
	}

	err := svc.AppendEvent(string(runtimeID), event)
	if err != nil {
		t.Fatalf("AppendEvent failed: %v", err)
	}

	events, err := svc.GetEvents(string(runtimeID), "")
	if err != nil {
		t.Fatalf("GetEvents failed: %v", err)
	}

	if len(events) != 1 {
		t.Errorf("Expected 1 event, got %d", len(events))
	}

	if events[0].Type != "test:event" {
		t.Errorf("Expected event type 'test:event', got '%s'", events[0].Type)
	}
}

func TestPersistence_EventQuery(t *testing.T) {
	svc := NewPersistenceService()

	runtimeID := statechart.RuntimeID("query-test-1")

	// Append multiple events
	for i := 0; i < 5; i++ {
		event := statechart.Event{
			Type:          "test:event",
			Payload:       map[string]any{"index": i},
			CorrelationID: string(runtimeID),
		}
		if err := svc.AppendEvent(string(runtimeID), event); err != nil {
			t.Fatalf("AppendEvent failed: %v", err)
		}
	}

	// Query all events
	allEvents, err := svc.GetEvents(string(runtimeID), "")
	if err != nil {
		t.Fatalf("GetEvents failed: %v", err)
	}

	if len(allEvents) != 5 {
		t.Errorf("Expected 5 events, got %d", len(allEvents))
	}

	// Query events since last event ID
	sinceID := allEvents[3].ID
	sinceEvents, err := svc.GetEvents(string(runtimeID), sinceID)
	if err != nil {
		t.Fatalf("GetEvents since failed: %v", err)
	}

	if len(sinceEvents) != 1 {
		t.Errorf("Expected 1 event since ID, got %d", len(sinceEvents))
	}
}

func TestPersistence_MigrateShallow(t *testing.T) {
	svc := NewPersistenceService()

	runtimeID := statechart.RuntimeID("migrate-shallow-1")
	event := statechart.Event{Type: "test:event", Payload: map[string]any{"data": "test"}}
	if err := svc.AppendEvent(string(runtimeID), event); err != nil {
		t.Fatalf("AppendEvent failed: %v", err)
	}

	newDef := statechart.ChartDefinition{
		ID:      "new-chart",
		Version: "2.0.0",
	}

	err := svc.Migrate(string(runtimeID), newDef, "shallowHistory")
	if err != nil {
		t.Fatalf("Migrate failed: %v", err)
	}

	events, err := svc.GetEvents(string(runtimeID), "")
	if err != nil {
		t.Fatalf("GetEvents failed: %v", err)
	}

	if len(events) == 0 {
		t.Error("Expected events to be preserved in shallowHistory mode")
	}
}

func TestPersistence_MigrateDeep(t *testing.T) {
	svc := NewPersistenceService()

	runtimeID := statechart.RuntimeID("migrate-deep-1")

	for i := 0; i < 3; i++ {
		event := statechart.Event{Type: "test:event", Payload: map[string]any{"index": i}}
		if err := svc.AppendEvent(string(runtimeID), event); err != nil {
			t.Fatalf("AppendEvent failed: %v", err)
		}
	}

	newDef := statechart.ChartDefinition{
		ID:      "deep-chart",
		Version: "3.0.0",
	}

	err := svc.Migrate(string(runtimeID), newDef, "deepHistory")
	if err != nil {
		t.Fatalf("Migrate failed: %v", err)
	}

	events, err := svc.GetEvents(string(runtimeID), "")
	if err != nil {
		t.Fatalf("GetEvents failed: %v", err)
	}

	if len(events) != 3 {
		t.Errorf("Expected 3 events preserved in deepHistory mode, got %d", len(events))
	}
}

func TestPersistence_MigrateClean(t *testing.T) {
	svc := NewPersistenceService()

	runtimeID := statechart.RuntimeID("migrate-clean-1")

	for i := 0; i < 3; i++ {
		event := statechart.Event{Type: "test:event", Payload: map[string]any{"index": i}}
		if err := svc.AppendEvent(string(runtimeID), event); err != nil {
			t.Fatalf("AppendEvent failed: %v", err)
		}
	}

	newDef := statechart.ChartDefinition{
		ID:      "clean-chart",
		Version: "4.0.0",
	}

	err := svc.Migrate(string(runtimeID), newDef, "cleanStart")
	if err != nil {
		t.Fatalf("Migrate failed: %v", err)
	}

	events, err := svc.GetEvents(string(runtimeID), "")
	if err != nil {
		t.Fatalf("GetEvents failed: %v", err)
	}

	if len(events) != 0 {
		t.Errorf("Expected 0 events in cleanStart mode, got %d", len(events))
	}
}
