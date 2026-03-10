package memory

import (
	"crypto/sha256"
	"fmt"
)

// VectorStore interface for vector operations
type VectorStore interface {
	Embed(content string) ([]float32, error)
	Search(query []float32, topK int) ([]MemoryItem, error)
	Store(item MemoryItem) error
}

// MemoryItem represents a memory entry with vector embedding
type MemoryItem struct {
	ID       string
	Content  string
	Vector   []float32
	Metadata map[string]any
	Boundary string
}

// vectorStore implementation
type vectorStore struct {
	items map[string]MemoryItem
}

func newVectorStore() VectorStore {
	return &vectorStore{
		items: make(map[string]MemoryItem),
	}
}

// Embed generates a vector representation from content
// arch-v1.md L489: Content embedded to vector representation
func (vs *vectorStore) Embed(content string) ([]float32, error) {
	// Deterministic hash-based embedding (dimension: 384)
	hash := sha256.Sum256([]byte(content))
	
	// Generate 384-dimensional vector from hash
	dimension := 384
	vector := make([]float32, dimension)
	
	// Use hash bytes to generate deterministic float values
	for i := 0; i < dimension; i++ {
		// Cycle through hash bytes and combine with index for variation
		hashIndex := (i * 7 + i/64) % 32
		base := float64(hash[hashIndex]) / 255.0
		
		// Add index-based variation for more dimensions
		variation := float64(i%256) / 256.0
		
		// Normalize to [-1, 1] range
		vector[i] = float32((base + variation*0.1) * 2.0 - 1.0)
	}
	
	return vector, nil
}

// cosineSimilarity computes cosine similarity between two vectors
func cosineSimilarity(a, b []float32) float64 {
	if len(a) != len(b) {
		return 0
	}
	
	var dot, normA, normB float64
	for i := range a {
		dot += float64(a[i]) * float64(b[i])
		normA += float64(a[i]) * float64(a[i])
		normB += float64(b[i]) * float64(b[i])
	}
	
	if normA == 0 || normB == 0 {
		return 0
	}
	
	return dot / (normA * normB)
}

// Search finds topK most similar items to query vector
// arch-v1.md L489: Search returns topK results ranked by similarity
func (vs *vectorStore) Search(query []float32, topK int) ([]MemoryItem, error) {
	type scoredItem struct {
		item  MemoryItem
		score float64
	}
	
	scored := make([]scoredItem, 0, len(vs.items))
	for _, item := range vs.items {
		sim := cosineSimilarity(query, item.Vector)
		scored = append(scored, scoredItem{item: item, score: sim})
	}
	
	// Sort by score descending (simple bubble sort for small datasets)
	for i := 0; i < len(scored); i++ {
		for j := i + 1; j < len(scored); j++ {
			if scored[j].score > scored[i].score {
				scored[i], scored[j] = scored[j], scored[i]
			}
		}
	}
	
	// Return topK results
	if topK > len(scored) {
		topK = len(scored)
	}
	
	results := make([]MemoryItem, topK)
	for i := 0; i < topK; i++ {
		results[i] = scored[i].item
	}
	
	return results, nil
}

// Store adds an item to the vector store
func (vs *vectorStore) Store(item MemoryItem) error {
	vs.items[item.ID] = item
	return nil
}

// generateID creates a deterministic ID from content
func generateID(content string) string {
	hash := sha256.Sum256([]byte(content))
	return "mem-" + string(hash[:8])
}

// generateUniqueID creates a unique ID using binary encoding
func generateUniqueID(content string, metadata map[string]any) string {
	// Combine content and metadata for uniqueness
	var metaStr string
	for k, v := range metadata {
		metaStr += k + ":" + fmt.Sprintf("%v", v)
	}
	
	hash := sha256.Sum256([]byte(content + metaStr))
	
	// Use binary.LittleEndian to convert to string representation
	id := make([]byte, 16)
	for i := 0; i < 16; i++ {
		id[i] = hash[i]
	}
	
	return "mem-" + string(id)
}