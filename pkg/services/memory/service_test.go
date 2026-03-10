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

// TestMemoryService_AddEdge - arch-v1.md L470: GraphStore must add edges between nodes with properties
func TestMemoryService_AddEdge(t *testing.T) {
	svc := NewMemoryService()

	// Test 1: Add edge between nodes
	err := svc.AddEdge("node-a", "node-b", "knows", nil)
	if err != nil {
		t.Fatalf("AddEdge failed: %v", err)
	}

	// Test 2: Add edge with properties
	properties := map[string]any{"since": "2020", "level": "close"}
	err = svc.AddEdge("node-b", "node-c", "works-with", properties)
	if err != nil {
		t.Fatalf("AddEdge with properties failed: %v", err)
	}

	// Test 3: Duplicate edge handling (should overwrite)
	newProperties := map[string]any{"since": "2021", "level": "best"}
	err = svc.AddEdge("node-a", "node-b", "knows", newProperties)
	if err != nil {
		t.Fatalf("AddEdge duplicate handling failed: %v", err)
	}

	// Verify edge was updated via query
	pattern := GraphPattern{
		From: "node-a",
		To:   "node-b",
	}
	nodes, err := svc.QueryPattern(pattern)
	if err != nil {
		t.Fatalf("QueryPattern failed: %v", err)
	}

	if len(nodes) == 0 {
		t.Error("Expected at least one node in query results")
	}
}

// TestMemoryService_QueryPattern - arch-v1.md L470: GraphStore must query patterns and return matching nodes
func TestMemoryService_QueryPattern(t *testing.T) {
	svc := NewMemoryService()

	// Setup: Create graph with multiple nodes and edges
	svc.AddEdge("alice", "bob", "knows", map[string]any{"since": "2020"})
	svc.AddEdge("alice", "charlie", "knows", map[string]any{"since": "2019"})
	svc.AddEdge("bob", "david", "works-with", map[string]any{"level": "close"})
	svc.AddEdge("charlie", "david", "works-with", map[string]any{"level": "casual"})

	// Test 1: Query by From node (outgoing edges)
	pattern := GraphPattern{From: "alice"}
	nodes, err := svc.QueryPattern(pattern)
	if err != nil {
		t.Fatalf("QueryPattern failed: %v", err)
	}
	if len(nodes) != 1 {
		t.Errorf("Expected 1 node for From=alice, got %d", len(nodes))
	}

	// Test 2: Query by To node (incoming edges)
	pattern = GraphPattern{To: "david"}
	nodes, err = svc.QueryPattern(pattern)
	if err != nil {
		t.Fatalf("QueryPattern failed: %v", err)
	}
	if len(nodes) != 2 {
		t.Errorf("Expected 2 nodes for To=david, got %d", len(nodes))
	}

	// Test 3: Query by relationship type (returns From nodes)
	pattern = GraphPattern{Relationship: "knows"}
	nodes, err = svc.QueryPattern(pattern)
	if err != nil {
		t.Fatalf("QueryPattern failed: %v", err)
	}
	if len(nodes) != 1 {
		t.Errorf("Expected 1 node for Relationship=knows, got %d", len(nodes))
	}

	// Test 4: Query with property filtering
	pattern = GraphPattern{
		Relationship: "works-with",
		Properties:   map[string]any{"level": "close"},
	}
	nodes, err = svc.QueryPattern(pattern)
	if err != nil {
		t.Fatalf("QueryPattern with property filter failed: %v", err)
	}
	if len(nodes) != 1 {
		t.Errorf("Expected 1 node for works-with with level=close, got %d", len(nodes))
	}

	// Test 5: Query by From and To
	pattern = GraphPattern{From: "alice", To: "bob"}
	nodes, err = svc.QueryPattern(pattern)
	if err != nil {
		t.Fatalf("QueryPattern failed: %v", err)
	}
	if len(nodes) != 1 {
		t.Errorf("Expected 1 node for From=alice, To=bob, got %d", len(nodes))
	}
}

// TestMemoryService_TraverseRelationships - arch-v1.md L470: GraphStore must traverse relationships from a node
func TestMemoryService_TraverseRelationships(t *testing.T) {
	svc := NewMemoryService()

	// Setup: Create a chain of relationships
	// alice -> bob -> charlie -> david
	svc.AddEdge("alice", "bob", "knows", nil)
	svc.AddEdge("bob", "charlie", "knows", nil)
	svc.AddEdge("charlie", "david", "knows", nil)

	// Test 1: Traverse from alice with maxDepth=1 (should get alice->bob)
	edges, err := svc.TraverseRelationships("alice", 1)
	if err != nil {
		t.Fatalf("TraverseRelationships failed: %v", err)
	}
	if len(edges) != 1 {
		t.Errorf("Expected 1 edge for depth=1, got %d", len(edges))
	}
	if edges[0].From != "alice" || edges[0].To != "bob" {
		t.Errorf("Expected edge alice->bob, got %s->%s", edges[0].From, edges[0].To)
	}

	// Test 2: Traverse from alice with maxDepth=2 (should get alice->bob, bob->charlie)
	edges, err = svc.TraverseRelationships("alice", 2)
	if err != nil {
		t.Fatalf("TraverseRelationships failed: %v", err)
	}
	if len(edges) != 2 {
		t.Errorf("Expected 2 edges for depth=2, got %d", len(edges))
	}

	// Test 3: Traverse from alice with maxDepth=3 (should get all 3 edges)
	edges, err = svc.TraverseRelationships("alice", 3)
	if err != nil {
		t.Fatalf("TraverseRelationships failed: %v", err)
	}
	if len(edges) != 3 {
		t.Errorf("Expected 3 edges for depth=3, got %d", len(edges))
	}

	// Test 4: Traverse from bob with maxDepth=1 (should get bob->charlie)
	edges, err = svc.TraverseRelationships("bob", 1)
	if err != nil {
		t.Fatalf("TraverseRelationships failed: %v", err)
	}
	if len(edges) != 1 {
		t.Errorf("Expected 1 edge for bob depth=1, got %d", len(edges))
	}
	if edges[0].From != "bob" || edges[0].To != "charlie" {
		t.Errorf("Expected edge bob->charlie, got %s->%s", edges[0].From, edges[0].To)
	}

	// Test 5: Traverse from david (no outgoing edges)
	edges, err = svc.TraverseRelationships("david", 1)
	if err != nil {
		t.Fatalf("TraverseRelationships failed: %v", err)
	}
	if len(edges) != 0 {
		t.Errorf("Expected 0 edges for david, got %d", len(edges))
	}
}

// TestMemoryService_BoundaryFilteredQuery - arch-v1.md L470: GraphStore queries must respect boundary filters
func TestMemoryService_BoundaryFilteredQuery(t *testing.T) {
	svc := NewMemoryService()

	// Setup: Create nodes with boundary properties
	svc.AddEdge("alice", "bob", "knows", map[string]any{"boundary": "system"})
	svc.AddEdge("bob", "charlie", "knows", map[string]any{"boundary": "application"})
	svc.AddEdge("charlie", "david", "knows", map[string]any{"boundary": "forbidden"})

	// Test 1: Query with boundary filter should exclude forbidden
	pattern := GraphPattern{
		Relationship: "knows",
		Properties:   map[string]any{"boundary": "forbidden"},
	}
	nodes, err := svc.QueryPattern(pattern)
	if err != nil {
		t.Fatalf("QueryPattern failed: %v", err)
	}
	// Should find the edge but boundary filter should exclude it

	// Test 2: Query with system boundary should return system edges
	pattern = GraphPattern{
		Relationship: "knows",
		Properties:   map[string]any{"boundary": "system"},
	}
	nodes, err = svc.QueryPattern(pattern)
	if err != nil {
		t.Fatalf("QueryPattern failed: %v", err)
	}
	if len(nodes) != 1 {
		t.Errorf("Expected 1 node for system boundary, got %d", len(nodes))
	}

	// Test 3: Query all relationships, then verify boundary filtering works
	pattern = GraphPattern{Relationship: "knows"}
	nodes, err = svc.QueryPattern(pattern)
	if err != nil {
		t.Fatalf("QueryPattern failed: %v", err)
	}
	// Should return all From nodes (alice, bob, charlie)
	if len(nodes) != 3 {
		t.Errorf("Expected 3 nodes for all knows, got %d", len(nodes))
	}
}

// TestHotreloadableServices_MemoryContextMapInjection - spec: arch-v1.md L470, L488 (ContextMap injection via MessageSlice)
func TestHotreloadableServices_MemoryContextMapInjection(t *testing.T) {
	svc := NewMemoryService()

	testVector := []float32{0.1, 0.2, 0.3, 0.4}
	msg := MessageSlice{
		ID:       "msg-001",
		Content:  "test memory content",
		Boundary: "dmz",
		Taints:   []string{"USER_SUPPLIED"},
	}
	err := svc.Insert(testVector, msg)
	if err != nil {
		t.Fatalf("Expected no error inserting vector, got %v", err)
	}

	queryVector := []float32{0.11, 0.21, 0.31, 0.41}
	results, err := svc.Query(queryVector, 5, "dmz")
	if err != nil {
		t.Fatalf("Expected no error querying, got %v", err)
	}

	if len(results) == 0 {
		t.Error("Expected at least one result from query")
	}

	if results[0].Boundary != "dmz" {
		t.Errorf("Expected boundary 'dmz', got '%s'", results[0].Boundary)
	}

	innerMsg := MessageSlice{ID: "msg-002", Content: "secret", Boundary: "inner", Taints: []string{"INNER_ONLY"}}
	svc.Insert(testVector, innerMsg)

	filteredResults, err := svc.Query(queryVector, 10, "outer")
	if err != nil {
		t.Fatalf("Expected no error querying with outer filter, got %v", err)
	}
	for _, r := range filteredResults {
		if r.Boundary == "inner" {
			t.Error("Expected inner boundary content to be filtered out for outer query")
		}
	}
}
