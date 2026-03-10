package memory

import (
	"errors"
	"fmt"

	"github.com/maelstrom/v3/pkg/mail"
)

var NotImplementedError = errors.New("not implemented")

type memoryService struct {
	store       map[string]interface{}
	vectorStore VectorStore
	graphStore  GraphStore
}

func NewMemoryService() MemoryService {
	return &memoryService{
		store:       make(map[string]interface{}),
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

func (s *memoryService) Insert(vector []float32, msg MessageSlice) error {
	item := MemoryItem{
		ID:       msg.ID,
		Content:  msg.Content,
		Vector:   vector,
		Boundary: msg.Boundary,
		Metadata: map[string]any{
			"taints": msg.Taints,
		},
	}
	return s.vectorStore.Store(item)
}

func (s *memoryService) Query(vector []float32, topK int, boundaryFilter string) ([]MemoryResult, error) {
	items, err := s.vectorStore.Search(vector, topK)
	if err != nil {
		return nil, err
	}

	results := make([]MemoryResult, 0, len(items))
	for _, item := range items {
		if boundaryFilter != "" {
			switch boundaryFilter {
			case "inner":
				results = append(results, MemoryResult{
					ID:       item.ID,
					Content:  item.Content,
					Score:    0,
					Boundary: item.Boundary,
					Metadata: item.Metadata,
				})
			case "dmz":
				if item.Boundary != "inner" {
					results = append(results, MemoryResult{
						ID:       item.ID,
						Content:  item.Content,
						Score:    0,
						Boundary: item.Boundary,
						Metadata: item.Metadata,
					})
				}
			case "outer":
				if item.Boundary != "inner" && item.Boundary != "dmz" {
					results = append(results, MemoryResult{
						ID:       item.ID,
						Content:  item.Content,
						Score:    0,
						Boundary: item.Boundary,
						Metadata: item.Metadata,
					})
				}
			}
		} else {
			results = append(results, MemoryResult{
				ID:       item.ID,
				Content:  item.Content,
				Score:    0,
				Boundary: item.Boundary,
				Metadata: item.Metadata,
			})
		}
	}

	return results, nil
}

func (s *memoryService) QueryByQuery(query string, topK int, boundaryFilter string) ([]MemoryResult, error) {
	vector, err := s.Embed(query)
	if err != nil {
		return nil, err
	}
	return s.Query(vector, topK, boundaryFilter)
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

func (s *memoryService) HandleMail(m mail.Mail) error {
	return nil
}

func (s *memoryService) Start() error {
	return nil
}

func (s *memoryService) Stop() error {
	return nil
}
