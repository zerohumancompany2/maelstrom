package statechart

import (
	"testing"
	"time"

	"github.com/maelstrom/v3/internal/testutil"
)

// TestDynamicReclassification_AtomicToCompound changes state from atomic to compound at runtime.
func TestDynamicReclassification_AtomicToCompound(t *testing.T) {
	// Initial: atomic state "idle"
	def := ChartDefinition{
		ID:      "test-reclass",
		Version: "1.0.0",
		Root: &Node{
			ID:       "idle",
			Children: nil,
			Transitions: []Transition{
				{Event: "expand", Target: "idle"}, // Self-transition triggers reclassification
			},
		},
		InitialState: "idle",
	}

	mockCtx := testutil.NewMockApplicationContext()
	rt := &ChartRuntime{
		id:           "test-rt",
		definition:   def,
		state:        RuntimeStateRunning,
		activeState:  "idle",
		regionStates: make(map[string]string),
		eventQueue:   make([]Event, 0),
		appCtx:       mockCtx,
		actions:      make(map[string]ActionFn),
		guards:       make(map[string]GuardFn),
	}

	// Verify initial state is atomic
	node := rt.findNode(rt.definition.Root, "idle")
	if node.NodeType() != NodeTypeAtomic {
		t.Fatal("Initial state should be atomic")
	}

	// Simulate hot-reload: replace definition with compound version
	newDef := ChartDefinition{
		ID:      "test-reclass",
		Version: "1.0.1",
		Root: &Node{
			ID: "idle",
			Children: map[string]*Node{
				"child1": {
					ID:           "child1",
					Children:     nil,
					IsInitial:    true,
				},
			},
			Transitions: []Transition{
				{Event: "expand", Target: "idle"},
			},
		},
		InitialState: "idle",
	}

	// Apply reclassification
	rt.ReclassifyState("idle", newDef.Root.Children["child1"])

	// Verify state is now compound
	node = rt.findNode(rt.definition.Root, "idle")
	if node.NodeType() != NodeTypeCompound {
		t.Errorf("State should be compound after reclassification, got %v", node.NodeType())
	}

	// Active state should update to include default child
	if rt.activeState != "idle/child1" {
		t.Errorf("Active state should be 'idle/child1' after reclassification, got %s", rt.activeState)
	}
}

// TestDynamicReclassification_CompoundToAtomic changes state from compound to atomic at runtime.
func TestDynamicReclassification_CompoundToAtomic(t *testing.T) {
	// Initial: compound state with child
	def := ChartDefinition{
		ID:      "test-reclass",
		Version: "1.0.0",
		Root: &Node{
			ID: "idle",
			Children: map[string]*Node{
				"child1": {ID: "child1"},
			},
			Transitions: []Transition{
				{Event: "collapse", Target: "idle"},
			},
		},
		InitialState: "idle",
	}

	mockCtx := testutil.NewMockApplicationContext()
	rt := &ChartRuntime{
		id:           "test-rt",
		definition:   def,
		state:        RuntimeStateRunning,
		activeState:  "idle/child1",
		regionStates: make(map[string]string),
		eventQueue:   make([]Event, 0),
		appCtx:       mockCtx,
		actions:      make(map[string]ActionFn),
		guards:       make(map[string]GuardFn),
	}

	// Verify initial state is compound
	node := rt.findNode(rt.definition.Root, "idle")
	if node.NodeType() != NodeTypeCompound {
		t.Fatal("Initial state should be compound")
	}

	// Apply reclassification (remove children to make atomic)
	rt.ReclassifyState("idle", nil)

	// Verify state is now atomic
	node = rt.findNode(rt.definition.Root, "idle")
	if node.NodeType() != NodeTypeAtomic {
		t.Errorf("State should be atomic after reclassification, got %v", node.NodeType())
	}

	// Active state should be just "idle"
	if rt.activeState != "idle" {
		t.Errorf("Active state should be 'idle' after reclassification, got %s", rt.activeState)
	}
}

// TestDynamicReclassification_AtomicToParallel changes state from atomic to parallel at runtime.
func TestDynamicReclassification_AtomicToParallel(t *testing.T) {
	// Initial: atomic state
	def := ChartDefinition{
		ID:      "test-reclass",
		Version: "1.0.0",
		Root: &Node{
			ID:       "processing",
			Children: nil,
			Transitions: []Transition{
				{Event: "parallelize", Target: "processing"},
			},
		},
		InitialState: "processing",
	}

	mockCtx := testutil.NewMockApplicationContext()
	rt := &ChartRuntime{
		id:           "test-rt",
		definition:   def,
		state:        RuntimeStateRunning,
		activeState:  "processing",
		regionStates: make(map[string]string),
		eventQueue:   make([]Event, 0),
		appCtx:       mockCtx,
		actions:      make(map[string]ActionFn),
		guards:       make(map[string]GuardFn),
	}

	// Verify initial state is atomic
	node := rt.findNode(rt.definition.Root, "processing")
	if node.NodeType() != NodeTypeAtomic {
		t.Fatal("Initial state should be atomic")
	}

	// Simulate hot-reload: replace with parallel version
	newChildren := map[string]*Node{
		"regionA": {
			ID:       "regionA",
			Children: nil,
		},
		"regionB": {
			ID:       "regionB",
			Children: nil,
		},
	}

	// Apply reclassification
	rt.ReclassifyState("processing", newChildren)

	// Verify state is now parallel
	node = rt.findNode(rt.definition.Root, "processing")
	if node.NodeType() != NodeTypeParallel {
		t.Errorf("State should be parallel after reclassification, got %v", node.NodeType())
	}

	// Should have event router initialized
	time.Sleep(50 * time.Millisecond)
	if !rt.isParallel {
		t.Error("isParallel should be true after reclassification to parallel")
	}
	if rt.eventRouter == nil {
		t.Error("eventRouter should be initialized for parallel state")
	}
}
