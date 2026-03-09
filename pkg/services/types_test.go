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
