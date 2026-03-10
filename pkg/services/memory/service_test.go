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

// TestMemoryService_Embed - arch-v1.md L489: Content embedded to vector representation
func TestMemoryService_Embed(t *testing.T) {
	svc := NewMemoryService()

	content := "test content for embedding"
	
	// Embed content to vector
	vector, err := svc.Embed(content)
	if err != nil {
		t.Fatalf("Embed failed: %v", err)
	}
	
	// Check vector dimension is consistent (384)
	if len(vector) != 384 {
		t.Errorf("Expected vector dimension 384, got %d", len(vector))
	}
	
	// Check deterministic: same content produces same vector
	vector2, err := svc.Embed(content)
	if err != nil {
		t.Fatalf("Embed failed on second call: %v", err)
	}
	
	for i := range vector {
		if vector[i] != vector2[i] {
			t.Errorf("Embedding not deterministic at index %d: %f != %f", i, vector[i], vector2[i])
		}
	}
}

// TestMemoryService_VectorSearch - arch-v1.md L489: Search returns topK results ranked by similarity
func TestMemoryService_VectorSearch(t *testing.T) {
	svc := NewMemoryService()

	// Empty store returns empty results
	results, err := svc.VectorSearch([]float32{0.1, 0.2, 0.3}, 5)
	if err != nil {
		t.Fatalf("VectorSearch on empty store failed: %v", err)
	}
	if len(results) != 0 {
		t.Errorf("Expected 0 results from empty store, got %d", len(results))
	}
	
	// Store some items
	item1 := MemoryItem{
		ID:      "item-1",
		Content: "test content 1",
		Vector:  []float32{1.0, 0.0, 0.0},
	}
	item2 := MemoryItem{
		ID:      "item-2",
		Content: "test content 2",
		Vector:  []float32{0.0, 1.0, 0.0},
	}
	
	svc.StoreVectorItem(item1)
	svc.StoreVectorItem(item2)
	
	// Search with query similar to item1
	query := []float32{1.0, 0.1, 0.0}
	results, err = svc.VectorSearch(query, 2)
	if err != nil {
		t.Fatalf("VectorSearch failed: %v", err)
	}
	
	if len(results) != 2 {
		t.Errorf("Expected 2 results, got %d", len(results))
	}
	
	// Check topK: requesting 1 should return 1
	results, err = svc.VectorSearch(query, 1)
	if err != nil {
		t.Fatalf("VectorSearch with topK=1 failed: %v", err)
	}
	if len(results) != 1 {
		t.Errorf("Expected 1 result with topK=1, got %d", len(results))
	}
	
	// Check cosine similarity ranking: item1 should be first (more similar to query)
	if results[0].ID != "item-1" {
		t.Errorf("Expected item-1 to be first (most similar), got %s", results[0].ID)
	}
}

// TestMemoryService_StoreItem - arch-v1.md L489: Item stored with metadata, vector auto-computed
func TestMemoryService_StoreItem(t *testing.T) {
	svc := NewMemoryService()

	metadata := map[string]any{"key": "value", "boundary": "system"}
	
	// Store item with content and metadata
	memoryId, err := svc.Store("runtime-1", "test content", metadata)
	if err != nil {
		t.Fatalf("Store failed: %v", err)
	}
	
	if memoryId == "" {
		t.Error("Expected non-empty memory ID")
	}
	
	// Item should be retrievable via VectorSearch
	queryVector, err := svc.Embed("test content")
	if err != nil {
		t.Fatalf("Embed failed: %v", err)
	}
	
	results, err := svc.VectorSearch(queryVector, 5)
	if err != nil {
		t.Fatalf("VectorSearch failed: %v", err)
	}
	
	if len(results) != 1 {
		t.Errorf("Expected 1 result, got %d", len(results))
	}
	
	// Check that the stored item has the correct content
	if results[0].Content != "test content" {
		t.Errorf("Expected content 'test content', got '%s'", results[0].Content)
	}
	
	// Check that vector was auto-computed (non-empty)
	if len(results[0].Vector) == 0 {
		t.Error("Expected auto-computed vector to be non-empty")
	}
}
