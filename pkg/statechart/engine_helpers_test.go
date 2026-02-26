package statechart

import (
	"testing"

	"github.com/maelstrom/v3/internal/testutil"
)

// ============================================================================
// Test Helpers
// ============================================================================

func newTestEngine(t *testing.T) Library {
	t.Helper()
	return NewEngine()
}

func newSimpleAtomicChart() ChartDefinition {
	return ChartDefinition{
		ID:      "test-chart",
		Version: "1.0.0",
		Root: &Node{
			ID:       "idle",
			Children: nil,
			Transitions: []Transition{
				{Event: "start", Target: "running"},
			},
		},
		InitialState: "idle",
	}
}

func newCompoundChart() ChartDefinition {
	return ChartDefinition{
		ID:      "compound-chart",
		Version: "1.0.0",
		Root: &Node{
			ID: "root",
			Children: map[string]*Node{
				"idle": {
					ID:       "idle",
					Children: nil,
					Transitions: []Transition{
						{Event: "activate", Target: "active"},
					},
				},
				"active": {
					ID: "active",
					Children: map[string]*Node{
						"child1": {
							ID:       "child1",
							Children: nil,
							Transitions: []Transition{
								{Event: "next", Target: "root/idle"},
							},
						},
					},
					IsInitial: true,
				},
			},
			IsInitial: true,
		},
		InitialState: "root/idle",
	}
}

func newParallelChart() ChartDefinition {
	return ChartDefinition{
		ID:      "parallel-chart",
		Version: "1.0.0",
		Root: &Node{
			ID: "root",
			Children: map[string]*Node{
				"regionA": {
					ID:       "regionA",
					Children: nil,
					Transitions: []Transition{
						{Event: "nextA", Target: "regionA-done"},
					},
				},
				"regionA-done": {
					ID:       "regionA-done",
					Children: nil,
				},
				"regionB": {
					ID:       "regionB",
					Children: nil,
					Transitions: []Transition{
						{Event: "nextB", Target: "regionB-done"},
					},
				},
				"regionB-done": {
					ID:       "regionB-done",
					Children: nil,
				},
			},
			RegionNames: []string{"regionA", "regionB"},
		},
		InitialState: "root",
	}
}

// newMockContext is a helper to create a mock application context
func newMockContext() ApplicationContext {
	return testutil.NewMockApplicationContext()
}
