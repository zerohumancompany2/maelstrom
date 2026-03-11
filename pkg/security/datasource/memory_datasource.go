package datasource

import (
	"sync"

	"github.com/maelstrom/v3/pkg/security/taint"
)

type memoryDataSource struct {
	mu          sync.RWMutex
	taintEngine *taint.TaintEngine
	data        map[string]any
	taints      map[string][]string
}

func NewMemoryDataSource(cfg *Config) *memoryDataSource {
	return &memoryDataSource{
		taintEngine: cfg.TaintEngine,
		data:        make(map[string]any),
		taints:      make(map[string][]string),
	}
}

func (m *memoryDataSource) Get(key string) (any, []string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	data, ok := m.data[key]
	if !ok {
		return nil, []string{}, nil
	}

	storedTaints, ok := m.taints[key]
	if !ok {
		storedTaints = []string{string(taint.TaintExternal)}
	}

	return data, storedTaints, nil
}

func (m *memoryDataSource) Put(key string, data any, taints []string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.data[key] = data

	if len(taints) == 0 {
		taints = []string{string(taint.TaintExternal)}
	}

	m.taints[key] = append([]string(nil), taints...)
	return nil
}
