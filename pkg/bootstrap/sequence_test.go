package bootstrap

import (
	"context"
	"testing"
)

// TestBootstrapSequence_SecurityFirst verifies bootstrap starts with security service.
func TestBootstrapSequence_SecurityFirst(t *testing.T) {
	seq := NewSequence()

	var enteredStates []string
	seq.OnStateEnter(func(state string) error {
		enteredStates = append(enteredStates, state)
		return nil
	})

	ctx := context.Background()
	if err := seq.Start(ctx); err != nil {
		t.Fatalf("failed to start: %v", err)
	}

	// Should have entered security state
	if len(enteredStates) != 1 || enteredStates[0] != "security" {
		t.Errorf("expected [security], got %v", enteredStates)
	}

	if seq.CurrentState() != "security" {
		t.Errorf("expected state 'security', got %q", seq.CurrentState())
	}
}

// TestBootstrapSequence_CompleteSequence verifies full bootstrap flow.
func TestBootstrapSequence_CompleteSequence(t *testing.T) {
	seq := NewSequence()

	var enteredStates []string
	seq.OnStateEnter(func(state string) error {
		enteredStates = append(enteredStates, state)
		return nil
	})

	var completeCalled bool
	seq.OnComplete(func() {
		completeCalled = true
	})

	ctx := context.Background()
	seq.Start(ctx)

	// Simulate service ready events
	events := []string{"SECURITY_READY", "COMMUNICATION_READY", "OBSERVABILITY_READY", "LIFECYCLE_READY", "KERNEL_READY"}
	for _, event := range events {
		if err := seq.HandleEvent(ctx, event); err != nil {
			t.Fatalf("failed to handle %s: %v", event, err)
		}
	}

	// Should have entered all states
	expected := []string{"security", "communication", "observability", "lifecycle", "handoff", "complete"}
	if len(enteredStates) != len(expected) {
		t.Errorf("expected %d states, got %d: %v", len(expected), len(enteredStates), enteredStates)
	}

	for i, exp := range expected {
		if i >= len(enteredStates) || enteredStates[i] != exp {
			t.Errorf("state %d: expected %q, got %v", i, exp, enteredStates)
			break
		}
	}

	if !completeCalled {
		t.Error("OnComplete callback was not called")
	}

	if !seq.IsComplete() {
		t.Error("IsComplete should return true")
	}
}

// TestBootstrapSequence_InvalidEvent verifies error on invalid event.
func TestBootstrapSequence_InvalidEvent(t *testing.T) {
	seq := NewSequence()
	ctx := context.Background()
	seq.Start(ctx)

	// Invalid event in security state
	err := seq.HandleEvent(ctx, "INVALID_EVENT")
	if err == nil {
		t.Error("expected error for invalid event")
	}
}
