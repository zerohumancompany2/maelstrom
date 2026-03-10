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

func TestMemoryService_Store(t *testing.T) {
	svc := NewMemoryService()

	err := svc.StoreKey("test-key", "test-value")
	if err != nil {
		t.Errorf("Expected nil error, got %v", err)
	}
}

func TestMemoryService_Query(t *testing.T) {
	svc := NewMemoryService()

	svc.StoreKey("test-key", "test-value")

	val, err := svc.QueryKey("test-key")
	if err != nil {
		t.Errorf("Expected nil error, got %v", err)
	}

	if val != "test-value" {
		t.Errorf("Expected 'test-value', got '%v'", val)
	}
}

// TestMemoryService_ID - arch-v1.md L470: MemoryService must return ID "sys:memory"
func TestMemoryService_ID(t *testing.T) {
	svc := NewMemoryService()

	id := svc.ID()
	if id != "sys:memory" {
		t.Errorf("Expected ID 'sys:memory', got '%s'", id)
	}
}
