package memory

import (
	"testing"
)

func TestMemory_BoundaryFilter(t *testing.T) {
	// Arrange
	svc := NewMemoryService()
	vector := []float32{0.1, 0.2, 0.3}

	// Act
	results, err := svc.Query(vector, 5, "system")

	// Assert
	if err != nil {
		t.Fatalf("Query with boundary filter failed: %v", err)
	}

	if len(results) != 0 {
		t.Errorf("Expected 0 results with boundary filter, got %d", len(results))
	}
}
