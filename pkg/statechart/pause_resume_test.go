package statechart

import (
	"testing"
	"time"

	"github.com/maelstrom/v3/internal/testutil"
)

// TestEngine_PauseResumeParallelRegions verifies pause/resume broadcasts to all regions.
func TestEngine_PauseResumeParallelRegions(t *testing.T) {
	def := ChartDefinition{
		ID:      "test-pause-resume",
		Version: "1.0.0",
		Root: &Node{
			ID: "root",
			Children: map[string]*Node{
				"parallel": {
					ID:          "parallel",
					RegionNames: []string{"regionA", "regionB"},
					Children: map[string]*Node{
						"regionA": {
							ID: "idleA",
							Children: map[string]*Node{
								"active": {ID: "active"},
							},
							Transitions: []Transition{
								{Event: "go", Target: "idleA/active"},
							},
						},
						"regionB": {
							ID: "idleB",
							Children: map[string]*Node{
								"running": {ID: "running"},
							},
							Transitions: []Transition{
								{Event: "run", Target: "idleB/running"},
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

	rtID, err := engine.Spawn(def, mockCtx)
	if err != nil {
		t.Fatalf("Spawn failed: %v", err)
	}

	if err := engine.Control(rtID, CmdStart); err != nil {
		t.Fatalf("Start failed: %v", err)
	}

	runtime := engine.runtimes[rtID]

	// Enter parallel state
	engine.Dispatch(rtID, Event{Type: "enterParallel"})
	time.Sleep(100 * time.Millisecond)

	if !runtime.isParallel {
		t.Fatal("Expected to be in parallel state")
	}

	// Act: Pause
	if err := engine.Control(rtID, CmdPause); err != nil {
		t.Fatalf("Pause failed: %v", err)
	}
	time.Sleep(50 * time.Millisecond)

	// Assert: Dispatch should fail while paused (events rejected)
	err = engine.Dispatch(rtID, Event{Type: "go", TargetPath: "region:regionA"})
	if err == nil {
		t.Error("Expected Dispatch to fail while paused")
	}

	// State should still be idleA
	if runtime.eventRouter.regions["regionA"].stateMachine.activeState != "idleA" {
		t.Error("RegionA should still be in idleA (paused)")
	}

	// Act: Resume
	if err := engine.Control(rtID, CmdResume); err != nil {
		t.Fatalf("Resume failed: %v", err)
	}
	time.Sleep(50 * time.Millisecond)

	// Now events should process
	engine.Dispatch(rtID, Event{Type: "go", TargetPath: "region:regionA"})
	time.Sleep(50 * time.Millisecond)

	// State SHOULD have changed now that we're resumed
	if runtime.eventRouter.regions["regionA"].stateMachine.activeState != "idleA/active" {
		t.Errorf("RegionA should be in idleA/active after resume, got %s",
			runtime.eventRouter.regions["regionA"].stateMachine.activeState)
	}
}
