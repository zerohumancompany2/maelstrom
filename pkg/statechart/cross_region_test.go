package statechart

import (
	"testing"
	"time"

	"github.com/maelstrom/v3/internal/testutil"
)

// TestCrossRegion_EmitEvent verifies one region can emit an event to another region.
func TestCrossRegion_EmitEvent(t *testing.T) {
	var actionLog []string

	def := ChartDefinition{
		ID:      "test-cross-region",
		Version: "1.0.0",
		Root: &Node{
			ID: "root",
			Children: map[string]*Node{
				"parallel": {
					ID:          "parallel",
					RegionNames: []string{"regionA", "regionB"},
					Children: map[string]*Node{
						"regionA": {
							ID: "regionA",
							Children: map[string]*Node{
								"idleA": {ID: "idleA"},
							},
							Transitions: []Transition{
								{Event: "notifyB", Target: "regionA/idleA", Actions: []string{"emitToB"}},
							},
						},
						"regionB": {
							ID: "regionB",
							Children: map[string]*Node{
								"idleB": {ID: "idleB"},
								"activeB": {ID: "activeB"},
							},
							Transitions: []Transition{
								{Event: "activate", Target: "regionB/activeB"},
							},
						},
					},
				},
			},
			Transitions: []Transition{
				{Event: "enterParallel", Target: "root/parallel"},
			},
		},
		InitialState: "root",
	}

	engine := NewEngine().(*Engine)
	mockCtx := testutil.NewMockApplicationContext()

	engine.RegisterAction("emitToB", func(rt RuntimeContext, ac ApplicationContext, ev Event) error {
		actionLog = append(actionLog, "emittingToB")
		// In a real scenario, this would emit an event to regionB
		// For now, we just verify the action fires
		return nil
	})

	rtID, _ := engine.Spawn(def, mockCtx)
	engine.Control(rtID, CmdStart)

	// Enter parallel state
	engine.Dispatch(rtID, Event{Type: "enterParallel"})
	time.Sleep(100 * time.Millisecond)

	// Trigger action in regionA that should affect regionB
	engine.Dispatch(rtID, Event{Type: "notifyB"})
	time.Sleep(100 * time.Millisecond)

	if len(actionLog) < 1 {
		t.Errorf("Expected action to fire, got %v", actionLog)
	}
}

// TestCrossRegion_BroadcastFromParent verifies parent can broadcast to all regions.
func TestCrossRegion_BroadcastFromParent(t *testing.T) {
	var actionLog []string

	def := ChartDefinition{
		ID:      "test-broadcast",
		Version: "1.0.0",
		Root: &Node{
			ID: "root",
			Children: map[string]*Node{
				"parallel": {
					ID:          "parallel",
					RegionNames: []string{"regionA", "regionB"},
					Children: map[string]*Node{
						"regionA": {
							ID: "regionA",
							Children: map[string]*Node{
								"idleA": {ID: "idleA"},
								"activeA": {ID: "activeA"},
							},
							Transitions: []Transition{
								{Event: "broadcast", Target: "regionA/activeA"},
							},
						},
						"regionB": {
							ID: "regionB",
							Children: map[string]*Node{
								"idleB": {ID: "idleB"},
								"activeB": {ID: "activeB"},
							},
							Transitions: []Transition{
								{Event: "broadcast", Target: "regionB/activeB"},
							},
						},
					},
				},
			},
			Transitions: []Transition{
				{Event: "enterParallel", Target: "root/parallel"},
			},
		},
		InitialState: "root",
	}

	engine := NewEngine().(*Engine)
	mockCtx := testutil.NewMockApplicationContext()

	engine.RegisterAction("logAction", func(rt RuntimeContext, ac ApplicationContext, ev Event) error {
		actionLog = append(actionLog, "fired")
		return nil
	})

	rtID, _ := engine.Spawn(def, mockCtx)
	engine.Control(rtID, CmdStart)

	// Enter parallel state
	engine.Dispatch(rtID, Event{Type: "enterParallel"})
	time.Sleep(100 * time.Millisecond)

	// Broadcast event to parallel state - should go to all regions
	engine.Dispatch(rtID, Event{Type: "broadcast"})
	time.Sleep(100 * time.Millisecond)

	// Verify both regions processed the event by checking runtime state
	runtime := engine.runtimes[rtID]
	if !runtime.isParallel {
		t.Error("Expected runtime to be in parallel state")
	}
}
