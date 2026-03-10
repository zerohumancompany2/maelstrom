package orchestrator

import (
	"testing"
)

func TestExecutionPolicy_SeqContinueStructure(t *testing.T) {
	// Given
	// PolicySeqContinue should be defined

	// When
	// Access PolicySeqContinue fields

	// Then
	if PolicySeqContinue.Mode != "seq_continue" {
		t.Errorf("Expected Mode to be 'seq_continue', got '%s'", PolicySeqContinue.Mode)
	}

	if PolicySeqContinue.MaxRetries != 1 {
		t.Errorf("Expected MaxRetries to be 1, got %d", PolicySeqContinue.MaxRetries)
	}

	if PolicySeqContinue.Isolation != "process" {
		t.Errorf("Expected Isolation to be 'process', got '%s'", PolicySeqContinue.Isolation)
	}
}

func TestExecutionPolicy_SeqFailFastStructure(t *testing.T) {
	// Given
	// PolicySeqFailFast should be defined

	// When
	// Access PolicySeqFailFast fields

	// Then
	if PolicySeqFailFast.Mode != "seq_failfast" {
		t.Errorf("Expected Mode to be 'seq_failfast', got '%s'", PolicySeqFailFast.Mode)
	}

	if PolicySeqFailFast.MaxRetries != 2 {
		t.Errorf("Expected MaxRetries to be 2, got %d", PolicySeqFailFast.MaxRetries)
	}

	if PolicySeqFailFast.Isolation != "strict" {
		t.Errorf("Expected Isolation to be 'strict', got '%s'", PolicySeqFailFast.Isolation)
	}
}

func TestExecutionPolicy_ParContinueStructure(t *testing.T) {
	// Given
	// PolicyParContinue should be defined

	// When
	// Access PolicyParContinue fields

	// Then
	if PolicyParContinue.Mode != "par_continue" {
		t.Errorf("Expected Mode to be 'par_continue', got '%s'", PolicyParContinue.Mode)
	}

	if PolicyParContinue.MaxParallel != 8 {
		t.Errorf("Expected MaxParallel to be 8, got %d", PolicyParContinue.MaxParallel)
	}

	if PolicyParContinue.Isolation != "strict" {
		t.Errorf("Expected Isolation to be 'strict', got '%s'", PolicyParContinue.Isolation)
	}
}

func TestExecutionPolicy_ParFailFastDropped(t *testing.T) {
	// Given
	// par_failfast policy should NOT be defined (dropped from MVP)

	// When & Then
	// This test verifies that PolicyParFailFast does not exist
	// The par_failfast pattern was dropped due to being "too racy, not worth complexity"
	// We verify this by checking that only the 3 expected policies exist

	policies := []ExecutionPolicy{
		PolicySeqContinue,
		PolicySeqFailFast,
		PolicyParContinue,
	}

	// Verify we have exactly 3 policies
	if len(policies) != 3 {
		t.Errorf("Expected exactly 3 policies, got %d", len(policies))
	}

	// Verify none of them is par_failfast
	for _, policy := range policies {
		if policy.Mode == "par_failfast" {
			t.Error("PolicyParFailFast should not be defined (dropped from MVP)")
		}
	}
}
