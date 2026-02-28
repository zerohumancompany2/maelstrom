package bootstrap

import (
	"context"
	"fmt"
	"log"
)

// Sequence orchestrates the bootstrap state machine.
type Sequence struct {
	currentState string
	services     map[string]bool // Track loaded services
	onStateEnter func(state string) error
	onComplete   func()
}

// NewSequence creates a new bootstrap sequence starting at "initializing".
func NewSequence() *Sequence {
	return &Sequence{
		currentState: "initializing",
		services:     make(map[string]bool),
	}
}

// OnStateEnter registers a callback for state entry (to load services).
func (s *Sequence) OnStateEnter(fn func(state string) error) {
	s.onStateEnter = fn
}

// OnComplete registers a callback for when bootstrap finishes.
func (s *Sequence) OnComplete(fn func()) {
	s.onComplete = fn
}

// Start begins the bootstrap sequence.
func (s *Sequence) Start(ctx context.Context) error {
	log.Println("[bootstrap] Starting bootstrap sequence")

	// Initial transition
	return s.transition(ctx, "security")
}

// HandleEvent processes an event and transitions if valid.
func (s *Sequence) HandleEvent(ctx context.Context, event string) error {
	log.Printf("[bootstrap] Received event: %s in state: %s", event, s.currentState)

	// State machine transitions
	transitions := map[string]map[string]string{
		"security":      {"SECURITY_READY": "communication"},
		"communication": {"COMMUNICATION_READY": "observability"},
		"observability": {"OBSERVABILITY_READY": "lifecycle"},
		"lifecycle":     {"LIFECYCLE_READY": "handoff"},
		"handoff":       {"KERNEL_READY": "complete"},
	}

	stateTrans, ok := transitions[s.currentState]
	if !ok {
		return fmt.Errorf("no transitions defined for state: %s", s.currentState)
	}

	nextState, ok := stateTrans[event]
	if !ok {
		return fmt.Errorf("event %s not valid in state %s", event, s.currentState)
	}

	return s.transition(ctx, nextState)
}

func (s *Sequence) transition(ctx context.Context, nextState string) error {
	log.Printf("[bootstrap] Transitioning: %s -> %s", s.currentState, nextState)

	// Execute entry action for next state
	if s.onStateEnter != nil {
		if err := s.onStateEnter(nextState); err != nil {
			return fmt.Errorf("failed to enter state %s: %w", nextState, err)
		}
	}

	s.currentState = nextState

	// Check if complete
	if nextState == "complete" {
		log.Println("[bootstrap] Bootstrap complete")
		if s.onComplete != nil {
			s.onComplete()
		}
	}

	return nil
}

// CurrentState returns the current bootstrap state.
func (s *Sequence) CurrentState() string {
	return s.currentState
}

// IsComplete returns true if bootstrap has finished.
func (s *Sequence) IsComplete() bool {
	return s.currentState == "complete"
}
