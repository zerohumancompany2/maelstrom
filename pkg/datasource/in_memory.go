package datasource

import (
	"sync"

	"github.com/maelstrom/v3/pkg/security"
)

type inMemoryDataSource struct {
	mu                 sync.RWMutex
	taints             map[string][]string
	allowedForBoundary []security.BoundaryType
}

func NewInMemoryDataSource() DataSource {
	return &inMemoryDataSource{
		taints: make(map[string][]string),
	}
}

func (im *inMemoryDataSource) TagOnWrite(path string, taints []string) error {
	im.mu.Lock()
	defer im.mu.Unlock()
	im.taints[path] = append([]string(nil), taints...)
	return nil
}

func (im *inMemoryDataSource) GetTaints(path string) ([]string, error) {
	// TODO: implement
	return nil, nil
}

func (im *inMemoryDataSource) ValidateAccess(boundary security.BoundaryType) error {
	// TODO: implement
	return nil
}
