package datasource

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"

	"github.com/maelstrom/v3/pkg/security/taint"
)

type fileDataSource struct {
	mu          sync.RWMutex
	taintEngine *taint.TaintEngine
	taints      map[string][]string
}

func NewFileDataSource(cfg *Config) *fileDataSource {
	return &fileDataSource{
		taintEngine: cfg.TaintEngine,
		taints:      make(map[string][]string),
	}
}

func (f *fileDataSource) Read(path string) (any, []string, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, nil, err
	}

	var content any
	if err := json.Unmarshal(data, &content); err != nil {
		content = string(data)
	}

	storedTaints, ok := f.taints[path]
	if !ok {
		storedTaints = []string{string(taint.TaintExternal)}
	}

	return content, storedTaints, nil
}

func (f *fileDataSource) Write(path string, data any, taints []string) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	var content []byte
	var err error

	switch v := data.(type) {
	case string:
		content = []byte(v)
	default:
		content, err = json.Marshal(data)
		if err != nil {
			return err
		}
	}

	if err := os.WriteFile(path, content, 0644); err != nil {
		return err
	}

	if len(taints) == 0 {
		taints = []string{string(taint.TaintExternal)}
	}

	f.taints[path] = append([]string(nil), taints...)
	return nil
}
