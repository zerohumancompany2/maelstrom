package memory

import (
	"github.com/maelstrom/v3/pkg/mail"
	"github.com/maelstrom/v3/pkg/statechart"
)

// MemoryService interface for memory operations
type MemoryService interface {
	ID() string
	Embed(content string) ([]float32, error)
	VectorSearch(query []float32, topK int) ([]MemoryItem, error)
	StoreVectorItem(item MemoryItem) error
	Store(runtimeId string, content string, metadata map[string]any) (string, error)
	Query(vector []float32, topK int, boundaryFilter string) ([]MemoryResult, error)
	QueryByQuery(query string, topK int, boundaryFilter string) ([]MemoryResult, error)
	Delete(memoryId string) error
	List(runtimeId string) ([]MemoryResult, error)
	StoreKey(key string, value interface{}) error
	QueryKey(key string) (interface{}, error)
	AddEdge(from, to, relationship string, properties map[string]any) error
	QueryPattern(pattern GraphPattern) ([]GraphNode, error)
	TraverseRelationships(startNode string, maxDepth int) ([]GraphEdge, error)
	HandleMail(mail mail.Mail) error
	Start() error
	Stop() error
}

// MemoryResult represents a memory entry
type MemoryResult struct {
	ID       string
	Content  string
	Score    float64
	Boundary string
	Metadata map[string]any
}

// BootstrapChart returns the chart definition for sys:memory
func BootstrapChart() statechart.ChartDefinition {
	return statechart.ChartDefinition{
		ID:      "sys:memory",
		Version: "1.0.0",
	}
}
