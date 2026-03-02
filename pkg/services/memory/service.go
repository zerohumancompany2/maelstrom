package memory

import "errors"

var NotImplementedError = errors.New("not implemented")

// memoryService implements MemoryService
type memoryService struct{}

// NewMemoryService creates a new MemoryService instance
func NewMemoryService() MemoryService {
	return &memoryService{}
}

// Store stores a memory entry with content and metadata
func (s *memoryService) Store(runtimeId string, content string, metadata map[string]any) (string, error) {
	return "memory-1", nil
}

// Query performs vector similarity search
func (s *memoryService) Query(vector []float32, topK int, boundaryFilter string) ([]MemoryResult, error) {
	return []MemoryResult{}, nil
}

// QueryByQuery performs text similarity search
func (s *memoryService) QueryByQuery(query string, topK int, boundaryFilter string) ([]MemoryResult, error) {
	return nil, NotImplementedError
}

// Delete removes a memory entry by ID
func (s *memoryService) Delete(memoryId string) error {
	return NotImplementedError
}

// List returns all memories for a given runtime
func (s *memoryService) List(runtimeId string) ([]MemoryResult, error) {
	return nil, NotImplementedError
}
