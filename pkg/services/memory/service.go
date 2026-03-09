package memory

import "errors"

var NotImplementedError = errors.New("not implemented")

type memoryService struct {
	store map[string]interface{}
}

func NewMemoryService() MemoryService {
	return &memoryService{
		store: make(map[string]interface{}),
	}
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
	return "memory-1", nil
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
