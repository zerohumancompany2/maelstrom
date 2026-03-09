package bootstrap

import (
	"context"
	"fmt"
	"log"
	"sync"
)

// Sequence orchestrates the bootstrap state machine.
type Sequence struct {
	kernel       interface{}
	mu           sync.RWMutex
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

// NewSequenceWithKernel creates a new bootstrap sequence with kernel reference.
func NewSequenceWithKernel(kernel interface{}) *Sequence {
	return &Sequence{
		kernel:       kernel,
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
	s.mu.RLock()
	current := s.currentState
	s.mu.RUnlock()

	log.Printf("[bootstrap] Received event: %s in state: %s", event, current)

	// State machine transitions
	transitions := map[string]map[string]string{
		"security":      {"SECURITY_READY": "communication"},
		"communication": {"COMMUNICATION_READY": "observability"},
		"observability": {"OBSERVABILITY_READY": "lifecycle"},
		"lifecycle":     {"LIFECYCLE_READY": "handoff"},
		"handoff":       {"KERNEL_READY": "complete"},
	}

	stateTrans, ok := transitions[current]
	if !ok {
		return fmt.Errorf("no transitions defined for state: %s", current)
	}

	nextState, ok := stateTrans[event]
	if !ok {
		return fmt.Errorf("event %s not valid in state %s", event, current)
	}

	return s.transition(ctx, nextState)
}

func (s *Sequence) transition(ctx context.Context, nextState string) error {
	s.mu.RLock()
	prev := s.currentState
	onStateEnter := s.onStateEnter
	onComplete := s.onComplete
	s.mu.RUnlock()

	log.Printf("[bootstrap] Transitioning: %s -> %s", prev, nextState)

	// Update state BEFORE calling handler to avoid race
	s.mu.Lock()
	s.currentState = nextState
	isComplete := nextState == "complete"
	s.mu.Unlock()

	// Execute entry action for next state
	if onStateEnter != nil {
		if err := onStateEnter(nextState); err != nil {
			return fmt.Errorf("failed to enter state %s: %w", nextState, err)
		}
	}

	// Check if complete
	if isComplete {
		log.Println("[bootstrap] Bootstrap complete")
		if onComplete != nil {
			onComplete()
		}
	}

	return nil
}

// CurrentState returns the current bootstrap state.
func (s *Sequence) CurrentState() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.currentState
}

// IsComplete returns true if bootstrap has finished.
func (s *Sequence) IsComplete() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.currentState == "complete"
}
