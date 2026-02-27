package statechart

import (
	"testing"
	"time"

	"github.com/maelstrom/v3/internal/testutil"
)

// TestEventRouter_RoutesToSpecificRegion verifies TargetPath routing.
func TestEventRouter_RoutesToSpecificRegion(t *testing.T) {
	// Arrange: two regions with different states
	regionDefs := map[string]ChartDefinition{
		"regionA": {
			ID: "regionA",
			Root: &Node{
				ID: "idle",
				Children: map[string]*Node{
					"active": {ID: "active"},
				},
				Transitions: []Transition{
					{Event: "activate", Target: "idle/active"},
				},
			},
			InitialState: "idle",
		},
		"regionB": {
			ID: "regionB",
			Root: &Node{
				ID: "waiting",
				Children: map[string]*Node{
					"done": {ID: "done"},
				},
				Transitions: []Transition{
					{Event: "finish", Target: "waiting/done"},
				},
			},
			InitialState: "waiting",
		},
	}

	mockCtx := testutil.NewMockApplicationContext()
	router := NewEventRouter(regionDefs, nil, nil, mockCtx)

	go router.Run()

	// Act: send event targeted to regionA
	router.Send(Event{Type: "activate", TargetPath: "region:regionA"})

	// Allow processing
	time.Sleep(50 * time.Millisecond)

	// Assert: regionA should have transitioned, regionB should not
	if router.regions["regionA"].stateMachine.activeState != "idle/active" {
		t.Errorf("regionA should be 'idle/active', got '%s'", router.regions["regionA"].stateMachine.activeState)
	}
	if router.regions["regionB"].stateMachine.activeState != "waiting" {
		t.Errorf("regionB should still be 'waiting', got '%s'", router.regions["regionB"].stateMachine.activeState)
	}
}

// TestEventRouter_BroadcastsToAllRegions verifies broadcast routing (no TargetPath).
func TestEventRouter_BroadcastsToAllRegions(t *testing.T) {
	// Arrange: two regions with same event type
	regionDefs := map[string]ChartDefinition{
		"regionA": {
			ID: "regionA",
			Root: &Node{
				ID: "idle",
				Children: map[string]*Node{
					"active": {ID: "active"},
				},
				Transitions: []Transition{
					{Event: "go", Target: "idle/active"},
				},
			},
			InitialState: "idle",
		},
		"regionB": {
			ID: "regionB",
			Root: &Node{
				ID: "idle",
				Children: map[string]*Node{
					"active": {ID: "active"},
				},
				Transitions: []Transition{
					{Event: "go", Target: "idle/active"},
				},
			},
			InitialState: "idle",
		},
	}

	mockCtx := testutil.NewMockApplicationContext()
	router := NewEventRouter(regionDefs, nil, nil, mockCtx)

	go router.Run()

	// Act: send broadcast event (no TargetPath)
	router.Send(Event{Type: "go"})

	time.Sleep(50 * time.Millisecond)

	// Assert: both regions should have transitioned
	if router.regions["regionA"].stateMachine.activeState != "idle/active" {
		t.Errorf("regionA should be 'idle/active', got '%s'", router.regions["regionA"].stateMachine.activeState)
	}
	if router.regions["regionB"].stateMachine.activeState != "idle/active" {
		t.Errorf("regionB should be 'idle/active', got '%s'", router.regions["regionB"].stateMachine.activeState)
	}
}

// TestEventRouter_ReceivesSysTransition verifies router receives region transitions.
func TestEventRouter_ReceivesSysTransition(t *testing.T) {
	regionDefs := map[string]ChartDefinition{
		"regionA": {
			ID: "regionA",
			Root: &Node{
				ID: "idle",
				Children: map[string]*Node{
					"active": {ID: "active"},
				},
				Transitions: []Transition{
					{Event: "go", Target: "idle/active"},
				},
			},
			InitialState: "idle",
		},
	}

	mockCtx := testutil.NewMockApplicationContext()
	router := NewEventRouter(regionDefs, nil, nil, mockCtx)

	// Use buffered output to prevent blocking
	outputChan := make(chan Event, 10)
	router.outputChan = outputChan

	go router.Run()

	// Act: trigger a transition
	router.Send(Event{Type: "go", TargetPath: "region:regionA"})

	// Assert: should receive SysTransition
	select {
	case ev := <-outputChan:
		if ev.Type != SysTransition {
			t.Errorf("Expected SysTransition, got '%s'", ev.Type)
		}
		payload, ok := ev.Payload.(TransitionPayload)
		if !ok {
			t.Error("Expected TransitionPayload")
			return
		}
		if payload.From != "idle" {
			t.Errorf("Expected From='idle', got '%s'", payload.From)
		}
		if payload.To != "idle/active" {
			t.Errorf("Expected To='idle/active', got '%s'", payload.To)
		}
	case <-time.After(100 * time.Millisecond):
		t.Error("Timeout waiting for SysTransition")
	}
}
