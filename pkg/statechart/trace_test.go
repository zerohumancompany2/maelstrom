package statechart

import (
	"testing"

	"github.com/maelstrom/v3/internal/testutil"
)

// mockTraceCallback implements TraceCallback for testing.
type mockTraceCallback struct {
	transitions []transitionTrace
	entries     []stateTrace
	exits       []stateTrace
	dispatches  []Event
}

type transitionTrace struct {
	from  string
	to    string
	event Event
}

type stateTrace struct {
	state string
	event Event
}

func (m *mockTraceCallback) OnTransition(runtimeID RuntimeID, fromState, toState string, event Event) {
	m.transitions = append(m.transitions, transitionTrace{from: fromState, to: toState, event: event})
}

func (m *mockTraceCallback) OnStateEntry(runtimeID RuntimeID, state string, event Event) {
	m.entries = append(m.entries, stateTrace{state: state, event: event})
}

func (m *mockTraceCallback) OnStateExit(runtimeID RuntimeID, state string, event Event) {
	m.exits = append(m.exits, stateTrace{state: state, event: event})
}

func (m *mockTraceCallback) OnEventDispatch(runtimeID RuntimeID, event Event) {
	m.dispatches = append(m.dispatches, event)
}

// TestTraceCallback_Transition fires trace callback on state transition.
func TestTraceCallback_Transition(t *testing.T) {
	mock := &mockTraceCallback{}

	def := ChartDefinition{
		ID:      "test-trace",
		Version: "1.0.0",
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
	}

	mockCtx := testutil.NewMockApplicationContext()
	sm := &StateMachine{
		definition:    def,
		activeState:   def.InitialState,
		actions:       make(map[string]ActionFn),
		guards:        make(map[string]GuardFn),
		appCtx:        mockCtx,
		TraceCallback: mock,
	}

	sm.ProcessEvent(Event{Type: "go"})

	if len(mock.transitions) != 1 {
		t.Errorf("Expected 1 transition trace, got %d", len(mock.transitions))
	}
	if mock.transitions[0].from != "idle" {
		t.Errorf("Expected from='idle', got '%s'", mock.transitions[0].from)
	}
	if mock.transitions[0].to != "idle/active" {
		t.Errorf("Expected to='idle/active', got '%s'", mock.transitions[0].to)
	}
}

// TestTraceCallback_EntryExit fires on entry and exit actions.
func TestTraceCallback_EntryExit(t *testing.T) {
	mock := &mockTraceCallback{}

	def := ChartDefinition{
		ID:      "test-trace-entry",
		Version: "1.0.0",
		Root: &Node{
			ID:           "idle",
			EntryActions: []string{"entryIdle"},
			ExitActions:  []string{"exitIdle"},
			Children: map[string]*Node{
				"active": {
					ID:           "active",
					EntryActions: []string{"entryActive"},
				},
			},
			Transitions: []Transition{
				{Event: "go", Target: "idle/active", Actions: []string{"transAction"}},
			},
		},
		InitialState: "idle",
	}

	mockCtx := testutil.NewMockApplicationContext()
	sm := &StateMachine{
		definition:  def,
		activeState: def.InitialState,
		actions: map[string]ActionFn{
			"entryIdle":   func(rt RuntimeContext, ac ApplicationContext, ev Event) error { return nil },
			"exitIdle":    func(rt RuntimeContext, ac ApplicationContext, ev Event) error { return nil },
			"transAction": func(rt RuntimeContext, ac ApplicationContext, ev Event) error { return nil },
			"entryActive": func(rt RuntimeContext, ac ApplicationContext, ev Event) error { return nil },
		},
		guards:        make(map[string]GuardFn),
		appCtx:        mockCtx,
		TraceCallback: mock,
	}

	sm.ProcessEvent(Event{Type: "go"})

	// Should have exit for idle, then entry for active
	if len(mock.exits) != 1 || mock.exits[0].state != "idle" {
		t.Errorf("Expected 1 exit for 'idle', got %v", mock.exits)
	}
	if len(mock.entries) != 1 || mock.entries[0].state != "idle/active" {
		t.Errorf("Expected 1 entry for 'idle/active', got %v", mock.entries)
	}
}

// TestTraceCallback_EventDispatch fires on event dispatch.
func TestTraceCallback_EventDispatch(t *testing.T) {
	mock := &mockTraceCallback{}

	def := ChartDefinition{
		ID:      "test-trace-dispatch",
		Version: "1.0.0",
		Root: &Node{
			ID:       "idle",
			Children: nil,
			Transitions: []Transition{
				{Event: "go", Target: "running"},
			},
		},
		InitialState: "idle",
	}

	mockCtx := testutil.NewMockApplicationContext()
	sm := &StateMachine{
		definition:    def,
		activeState:   def.InitialState,
		actions:       make(map[string]ActionFn),
		guards:        make(map[string]GuardFn),
		appCtx:        mockCtx,
		TraceCallback: mock,
	}

	sm.ProcessEvent(Event{Type: "go"})

	if len(mock.dispatches) != 1 {
		t.Errorf("Expected 1 dispatch trace, got %d", len(mock.dispatches))
	}
	if mock.dispatches[0].Type != "go" {
		t.Errorf("Expected event type 'go', got '%s'", mock.dispatches[0].Type)
	}
}
