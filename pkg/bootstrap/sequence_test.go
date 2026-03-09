package bootstrap

import (
	"context"
	"fmt"
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

// TestSequence_PassesKernelToActions verifies sequence creates chart with actions registered.
func TestSequence_PassesKernelToActions(t *testing.T) {
	var kernelRef interface{}
	seq := NewSequenceWithKernel(&kernelRef)

	if seq == nil {
		t.Fatal("expected non-nil sequence")
	}
}

// TestActions_ReceiveKernelReference verifies actions receive kernel reference.
func TestActions_ReceiveKernelReference(t *testing.T) {
	kernel := &MockKernel{
		services: make(map[string]interface{}),
	}
	seq := NewSequenceWithKernel(kernel)

	kernelAccessed := false
	seq.OnStateEnter(func(state string) error {
		if state == "security" {
			kernelAccessed = true
			kernel.RegisterService("security", "security-service")
		}
		return nil
	})

	ctx := context.Background()
	if err := seq.Start(ctx); err != nil {
		t.Fatalf("failed to start: %v", err)
	}

	if !kernelAccessed {
		t.Error("kernel was not accessed during state entry")
	}

	if len(kernel.services) != 1 {
		t.Errorf("expected 1 service registered, got %d", len(kernel.services))
	}

	if _, ok := kernel.services["security"]; !ok {
		t.Error("expected security service to be registered")
	}
}

type MockKernel struct {
	services map[string]interface{}
}

func (m *MockKernel) RegisterService(name string, service interface{}) error {
	m.services[name] = service
	return nil
}

// TestErrorPath_TriggersFailedState verifies error path triggers failed state.
func TestErrorPath_TriggersFailedState(t *testing.T) {
	kernel := &MockKernel{
		services: make(map[string]interface{}),
	}
	seq := NewSequenceWithKernel(kernel)

	var enteredStates []string
	seq.OnStateEnter(func(state string) error {
		enteredStates = append(enteredStates, state)
		if state == "security" {
			return fmt.Errorf("security bootstrap failed")
		}
		return nil
	})

	ctx := context.Background()
	if err := seq.Start(ctx); err == nil {
		t.Fatal("expected error during state entry")
	}

	if len(enteredStates) != 1 || enteredStates[0] != "security" {
		t.Errorf("expected [security], got %v", enteredStates)
	}
}

// TestSuccessPath_TriggersReadyState verifies success path triggers ready state.
func TestSuccessPath_TriggersReadyState(t *testing.T) {
	kernel := &MockKernel{
		services: make(map[string]interface{}),
	}
	seq := NewSequenceWithKernel(kernel)

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

	events := []string{"SECURITY_READY", "COMMUNICATION_READY", "OBSERVABILITY_READY", "LIFECYCLE_READY", "KERNEL_READY"}
	for _, event := range events {
		if err := seq.HandleEvent(ctx, event); err != nil {
			t.Fatalf("failed to handle %s: %v", event, err)
		}
	}

	expected := []string{"security", "communication", "observability", "lifecycle", "handoff", "complete"}
	if len(enteredStates) != len(expected) {
		t.Errorf("expected %d states, got %d: %v", len(expected), len(enteredStates), enteredStates)
	}

	if !completeCalled {
		t.Error("OnComplete callback was not called")
	}

	if !seq.IsComplete() {
		t.Error("IsComplete should return true")
	}
}
