package datasource

import (
	"encoding/json"
	"sync"

	"github.com/maelstrom/v3/pkg/security/taint"
)

type objectDataSource struct {
	mu          sync.RWMutex
	taintEngine *taint.TaintEngine
	objects     map[string]map[string]any
	taints      map[string][]string
}

func NewObjectDataSource(cfg *Config) *objectDataSource {
	return &objectDataSource{
		taintEngine: cfg.TaintEngine,
		objects:     make(map[string]map[string]any),
		taints:      make(map[string][]string),
	}
}

func (o *objectDataSource) GetObject(bucket, key string) (any, []string, error) {
	o.mu.RLock()
	defer o.mu.RUnlock()

	bucketMap, ok := o.objects[bucket]
	if !ok {
		return nil, []string{}, nil
	}

	data, ok := bucketMap[key]
	if !ok {
		return nil, []string{}, nil
	}

	taintKey := bucket + "/" + key
	storedTaints, ok := o.taints[taintKey]
	if !ok {
		storedTaints = []string{string(taint.TaintExternal)}
	}

	return data, storedTaints, nil
}

func (o *objectDataSource) PutObject(bucket, key string, data any, taints []string) error {
	o.mu.Lock()
	defer o.mu.Unlock()

	if _, ok := o.objects[bucket]; !ok {
		o.objects[bucket] = make(map[string]any)
	}

	var content any

	switch v := data.(type) {
	case string:
		content = v
	case map[string]any:
		content = v
	default:
		contentJSON, err := json.Marshal(data)
		if err != nil {
			return err
		}
		content = string(contentJSON)
	}

	o.objects[bucket][key] = content

	taintKey := bucket + "/" + key
	if len(taints) == 0 {
		taints = []string{string(taint.TaintExternal)}
	}

	o.taints[taintKey] = append([]string(nil), taints...)
	return nil
}
