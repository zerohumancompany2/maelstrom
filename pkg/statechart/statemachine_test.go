package statechart

import (
	"testing"

	"github.com/maelstrom/v3/internal/testutil"
)

// TestStateMachine_AtomicTransition verifies basic atomic state transition.
func TestStateMachine_AtomicTransition(t *testing.T) {
	// Arrange: simple atomic chart with idle -> running transition
	def := ChartDefinition{
		ID:      "test-atomic",
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

	mockCtx := testutil.NewMockApplicationContext()
	sm := &StateMachine{
		definition:  def,
		activeState: def.InitialState,
		actions:     make(map[string]ActionFn),
		guards:      make(map[string]GuardFn),
		appCtx:      mockCtx,
	}

	// Act: process "start" event
	result := sm.ProcessEvent(Event{Type: "start"})

	// Assert: should transition to "running"
	if !result.Transitioned {
		t.Error("Expected transition to occur")
	}
	if result.ToState != "running" {
		t.Errorf("Expected state 'running', got '%s'", result.ToState)
	}
	if sm.activeState != "running" {
		t.Errorf("Expected activeState 'running', got '%s'", sm.activeState)
	}
}

// TestStateMachine_EntryExitActions verifies entry and exit actions execute on transition.
func TestStateMachine_EntryExitActions(t *testing.T) {
	var actionLog []string

	def := ChartDefinition{
		ID:      "test-actions",
		Version: "1.0.0",
		Root: &Node{
			ID:           "idle",
			Children:     nil,
			EntryActions: []string{"idleEntry"},
			ExitActions:  []string{"idleExit"},
			Transitions: []Transition{
				{Event: "go", Target: "active", Actions: []string{"transitionAction"}},
			},
		},
		InitialState: "idle",
	}

	// Add target state to definition
	def.Root.Children = map[string]*Node{
		"active": {
			ID:           "active",
			EntryActions: []string{"activeEntry"},
			ExitActions:  []string{"activeExit"},
		},
	}

	mockCtx := testutil.NewMockApplicationContext()
	actions := map[string]ActionFn{
		"idleEntry": func(rt RuntimeContext, ac ApplicationContext, ev Event) error {
			actionLog = append(actionLog, "idleEntry")
			return nil
		},
		"idleExit": func(rt RuntimeContext, ac ApplicationContext, ev Event) error {
			actionLog = append(actionLog, "idleExit")
			return nil
		},
		"transitionAction": func(rt RuntimeContext, ac ApplicationContext, ev Event) error {
			actionLog = append(actionLog, "transitionAction")
			return nil
		},
		"activeEntry": func(rt RuntimeContext, ac ApplicationContext, ev Event) error {
			actionLog = append(actionLog, "activeEntry")
			return nil
		},
	}

	sm := &StateMachine{
		definition:  def,
		activeState: def.InitialState,
		actions:     actions,
		guards:      make(map[string]GuardFn),
		appCtx:      mockCtx,
	}

	// Trigger initial entry action
	sm.executeEntryActions("idle", Event{})

	// Act: process transition
	sm.ProcessEvent(Event{Type: "go"})

	// Assert: action order should be: idleEntry, idleExit, transitionAction, activeEntry
	expected := []string{"idleEntry", "idleExit", "transitionAction", "activeEntry"}
	if len(actionLog) != len(expected) {
		t.Errorf("Expected %d actions, got %d: %v", len(expected), len(actionLog), actionLog)
		return
	}
	for i, exp := range expected {
		if actionLog[i] != exp {
			t.Errorf("Expected action[%d] = '%s', got '%s'", i, exp, actionLog[i])
		}
	}
}
