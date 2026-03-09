package services

import (
	"testing"
	"time"
)

func TestServices_TraceFiltersHasRequiredFields(t *testing.T) {
	now := time.Now()
	filters := TraceFilters{
		RuntimeID: "test",
		EventType: "transition",
		FromTime:  now,
		ToTime:    now,
	}

	if filters.RuntimeID != "test" {
		t.Errorf("expected RuntimeID to be 'test', got '%s'", filters.RuntimeID)
	}
	if filters.EventType != "transition" {
		t.Errorf("expected EventType to be 'transition', got '%s'", filters.EventType)
	}
	if !filters.FromTime.Equal(now) {
		t.Error("expected FromTime to be now")
	}
	if !filters.ToTime.Equal(now) {
		t.Error("expected ToTime to be now")
	}
}

func TestServices_MetricsCollectorHasRequiredFields(t *testing.T) {
	collector := MetricsCollector{
		StateCounts:    map[string]int{"state1": 1},
		TransitionRate: 1.5,
		EventRate:      2.5,
		LastUpdate:     time.Now(),
	}

	if len(collector.StateCounts) != 1 {
		t.Errorf("expected StateCounts to have 1 entry, got %d", len(collector.StateCounts))
	}
	if collector.StateCounts["state1"] != 1 {
		t.Errorf("expected StateCounts['state1'] to be 1, got %d", collector.StateCounts["state1"])
	}
	if collector.TransitionRate != 1.5 {
		t.Errorf("expected TransitionRate to be 1.5, got %f", collector.TransitionRate)
	}
	if collector.EventRate != 2.5 {
		t.Errorf("expected EventRate to be 2.5, got %f", collector.EventRate)
	}
}
