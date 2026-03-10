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
