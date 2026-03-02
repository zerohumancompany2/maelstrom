package memory

import (
	"testing"
)

func TestMemory_Store(t *testing.T) {
	svc := NewMemoryService()
	_, err := svc.Store("runtime-1", "test content", map[string]any{"key": "value"})
	if err != nil {
		t.Fatalf("Store failed: %v", err)
	}
}

func TestMemory_QueryVector(t *testing.T) {
	svc := NewMemoryService()
	vector := []float32{0.1, 0.2, 0.3, 0.4}
	results, err := svc.Query(vector, 5, "")
	if err != nil {
		t.Fatalf("Query failed: %v", err)
	}
	if len(results) != 0 {
		t.Errorf("Expected 0 results, got %d", len(results))
	}
}

func TestMemory_QueryText(t *testing.T) {
	svc := NewMemoryService()
	results, err := svc.QueryByQuery("test query", 5, "")
	if err != nil {
		t.Fatalf("QueryByQuery failed: %v", err)
	}
	if len(results) != 0 {
		t.Errorf("Expected 0 results, got %d", len(results))
	}
}

func TestMemory_BoundaryFilter(t *testing.T) {
	svc := NewMemoryService()
	vector := []float32{0.1, 0.2, 0.3}
	results, err := svc.Query(vector, 5, "system")
	if err != nil {
		t.Fatalf("Query with boundary filter failed: %v", err)
	}
	if len(results) != 0 {
		t.Errorf("Expected 0 results with boundary filter, got %d", len(results))
	}
}

func TestMemory_Delete(t *testing.T) {
	svc := NewMemoryService()
	err := svc.Delete("memory-123")
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}
}

func TestMemory_List(t *testing.T) {
	svc := NewMemoryService()
	results, err := svc.List("runtime-1")
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(results) != 0 {
		t.Errorf("Expected 0 results, got %d", len(results))
	}
}
