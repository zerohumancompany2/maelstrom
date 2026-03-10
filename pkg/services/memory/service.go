package memory

import (
	"errors"
	"fmt"
)

var NotImplementedError = errors.New("not implemented")

type memoryService struct {
	store      map[string]interface{}
	vectorStore VectorStore
	graphStore GraphStore
}

func NewMemoryService() MemoryService {
	return &memoryService{
		store:      make(map[string]interface{}),
		vectorStore: newVectorStore(),
		graphStore:  newGraphStore(),
	}
}

func (s *memoryService) ID() string {
	return "sys:memory"
}

func (s *memoryService) Embed(content string) ([]float32, error) {
	return s.vectorStore.Embed(content)
}

func (s *memoryService) VectorSearch(query []float32, topK int) ([]MemoryItem, error) {
	return s.vectorStore.Search(query, topK)
}

func (s *memoryService) StoreVectorItem(item MemoryItem) error {
	return s.vectorStore.Store(item)
}

func (s *memoryService) StoreKey(key string, value interface{}) error {
	s.store[key] = value
	return nil
}

func (s *memoryService) QueryKey(key string) (interface{}, error) {
	val, ok := s.store[key]
	if !ok {
		return nil, errors.New("key not found")
	}
	return val, nil
}

func (s *memoryService) Store(runtimeId string, content string, metadata map[string]any) (string, error) {
	// Auto-compute vector from content
	vector, err := s.vectorStore.Embed(content)
	if err != nil {
		return "", fmt.Errorf("failed to embed content: %w", err)
	}
	
	// Generate unique ID
	id := generateUniqueID(content, metadata)
	
	// Extract boundary safely
	boundary := ""
	if metadata != nil {
		if b, ok := metadata["boundary"].(string); ok {
			boundary = b
		}
	}
	
	// Create and store memory item
	item := MemoryItem{
		ID:       id,
		Content:  content,
		Vector:   vector,
		Metadata: metadata,
		Boundary: boundary,
	}
	
	err = s.vectorStore.Store(item)
	if err != nil {
		return "", fmt.Errorf("failed to store item: %w", err)
	}
	
	return id, nil
}

func (s *memoryService) Query(vector []float32, topK int, boundaryFilter string) ([]MemoryResult, error) {
	return []MemoryResult{}, nil
}

func (s *memoryService) QueryByQuery(query string, topK int, boundaryFilter string) ([]MemoryResult, error) {
	return []MemoryResult{}, nil
}

func (s *memoryService) Delete(memoryId string) error {
	return nil
}

func (s *memoryService) List(runtimeId string) ([]MemoryResult, error) {
	return []MemoryResult{}, nil
}

func (s *memoryService) AddEdge(from, to, relationship string, properties map[string]any) error {
	return s.graphStore.AddEdge(from, to, relationship, properties)
}

func (s *memoryService) QueryPattern(pattern GraphPattern) ([]GraphNode, error) {
	return s.graphStore.Query(pattern)
}

func (s *memoryService) TraverseRelationships(startNode string, maxDepth int) ([]GraphEdge, error) {
	return s.graphStore.Traverse(startNode, maxDepth)
}
