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
	OutEdges   []GraphEdge // Outgoing edges (this node is From)
	InEdges    []GraphEdge // Incoming edges (this node is To)
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
			OutEdges:   []GraphEdge{},
			InEdges:    []GraphEdge{},
		}
	}
	if _, exists := gs.nodes[to]; !exists {
		gs.nodes[to] = GraphNode{
			ID:         to,
			Properties: make(map[string]any),
			OutEdges:   []GraphEdge{},
			InEdges:    []GraphEdge{},
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

	// Update from node outgoing edges
	fromNode := gs.nodes[from]
	fromNode.OutEdges = append(fromNode.OutEdges, edge)
	gs.nodes[from] = fromNode

	// Update to node incoming edges
	toNode := gs.nodes[to]
	toNode.InEdges = append(toNode.InEdges, edge)
	gs.nodes[to] = toNode

	return nil
}

// Query finds nodes matching a pattern
// arch-v1.md L470: GraphStore must query patterns and return matching nodes
func (gs *graphStore) Query(pattern GraphPattern) ([]GraphNode, error) {
	var results []GraphNode

	for id, node := range gs.nodes {
		matched := false

		// If From specified, only check outgoing edges and return From node
		if pattern.From != "" {
			if id != pattern.From {
				continue
			}
			for _, edge := range node.OutEdges {
				if pattern.Relationship != "" && edge.Type != pattern.Relationship {
					continue
				}
				if pattern.To != "" && edge.To != pattern.To {
					continue
				}
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
				matched = true
				break
			}
		} else if pattern.To != "" {
			// If To specified (without From), return nodes that have edges TO the To node
			for _, edge := range node.OutEdges {
				if edge.To != pattern.To {
					continue
				}
				if pattern.Relationship != "" && edge.Type != pattern.Relationship {
					continue
				}
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
				matched = true
				break
			}
		} else {
			// Only relationship or properties specified - return From nodes
			for _, edge := range node.OutEdges {
				if pattern.Relationship != "" && edge.Type != pattern.Relationship {
					continue
				}
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
				matched = true
				break
			}
		}

		if matched {
			results = append(results, node)
		}
	}

	return results, nil
}

// Traverse traverses relationships from a start node
// arch-v1.md L470: GraphStore must traverse relationships from a node
func (gs *graphStore) Traverse(startNode string, maxDepth int) ([]GraphEdge, error) {
	return nil, NotImplementedError
}