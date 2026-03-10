package memory

import "fmt"

// GraphStore interface for graph operations
// arch-v1.md L470: sys:memory — Long-term memory (vector/graph stores)
type GraphStore interface {
	AddEdge(from, to, relationship string, properties map[string]any) error
	Query(pattern GraphPattern) ([]GraphNode, error)
	Traverse(startNode string, maxDepth int) ([]GraphEdge, error)
}

// GraphNode represents a node in the graph
type GraphNode struct {
	ID         string
	Properties map[string]any
	Edges      []GraphEdge
}

// GraphEdge represents an edge between nodes
type GraphEdge struct {
	From       string
	To         string
	Type       string
	Properties map[string]any
}

// GraphPattern represents a query pattern for graph queries
type GraphPattern struct {
	From         string
	To           string
	Relationship string
	Properties   map[string]any
}

// graphStore implementation
type graphStore struct {
	nodes map[string]GraphNode
	edges map[string]GraphEdge
}

func newGraphStore() GraphStore {
	return &graphStore{
		nodes: make(map[string]GraphNode),
		edges: make(map[string]GraphEdge),
	}
}

// AddEdge adds an edge between two nodes with properties
// arch-v1.md L470: GraphStore must add edges between nodes with properties
func (gs *graphStore) AddEdge(from, to, relationship string, properties map[string]any) error {
	edgeKey := fmt.Sprintf("%s:%s:%s", from, to, relationship)
	
	// Ensure nodes exist
	if _, exists := gs.nodes[from]; !exists {
		gs.nodes[from] = GraphNode{
			ID:         from,
			Properties: make(map[string]any),
			Edges:      []GraphEdge{},
		}
	}
	if _, exists := gs.nodes[to]; !exists {
		gs.nodes[to] = GraphNode{
			ID:         to,
			Properties: make(map[string]any),
			Edges:      []GraphEdge{},
		}
	}
	
	// Create or update edge
	edge := GraphEdge{
		From:       from,
		To:         to,
		Type:       relationship,
		Properties: properties,
	}
	gs.edges[edgeKey] = edge
	
	// Update node edges (need to get node first, modify, then set back)
	node := gs.nodes[from]
	node.Edges = append(node.Edges, edge)
	gs.nodes[from] = node
	
	return nil
}

// Query finds nodes matching a pattern
// arch-v1.md L470: GraphStore must query patterns and return matching nodes
func (gs *graphStore) Query(pattern GraphPattern) ([]GraphNode, error) {
	var results []GraphNode
	
	for id, node := range gs.nodes {
		// If From specified without To, only return the From node
		if pattern.From != "" && pattern.To == "" && id != pattern.From {
			continue
		}
		// If To specified without From, only return the To node
		if pattern.To != "" && pattern.From == "" && id != pattern.To {
			continue
		}
		
		// Check edges match pattern
		for _, edge := range node.Edges {
			// Check relationship type
			if pattern.Relationship != "" && edge.Type != pattern.Relationship {
				continue
			}
			// Check if edge connects From to To
			if pattern.From != "" && pattern.To != "" {
				if edge.From != pattern.From || edge.To != pattern.To {
					continue
				}
			}
			// Check if edge matches From (outgoing)
			if pattern.From != "" && pattern.To == "" {
				if edge.From != pattern.From {
					continue
				}
			}
			// Check if edge matches To (incoming)
			if pattern.To != "" && pattern.From == "" {
				if edge.To != pattern.To {
					continue
				}
			}
			
			// Check property filtering (only if Properties specified)
			if pattern.Properties != nil {
				match := true
				for key, val := range pattern.Properties {
					if edge.Properties == nil || edge.Properties[key] != val {
						match = false
						break
					}
				}
				if !match {
					continue
				}
			}
			
			results = append(results, node)
			break
		}
	}
	
	return results, nil
}

// Traverse traverses relationships from a start node
// arch-v1.md L470: GraphStore must traverse relationships from a node
func (gs *graphStore) Traverse(startNode string, maxDepth int) ([]GraphEdge, error) {
	return nil, NotImplementedError
}