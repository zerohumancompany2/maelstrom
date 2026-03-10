package orchestrator

import (
	"testing"
)

func TestParallel_Continue_ExecutesConcurrently(t *testing.T) {
	// Given
	executor := NewParallelExecutor(PolicyParContinue)
	toolCalls := []ToolCall{
		{Name: "tool1", Arguments: map[string]any{"arg": "value1"}},
		{Name: "tool2", Arguments: map[string]any{"arg": "value2"}},
		{Name: "tool3", Arguments: map[string]any{"arg": "value3"}},
	}

	// When
	resultChan, err := executor.Execute(toolCalls)

	// Then
	if err != nil {
		t.Errorf("Expected Execute() to return nil error, got %v", err)
	}

	if resultChan == nil {
		t.Fatal("Expected non-nil result channel")
	}

	// Collect all results
	results := make([]ExecutionResult, 0, 3)
	for result := range resultChan {
		results = append(results, result)
	}

	if len(results) != 3 {
		t.Errorf("Expected 3 results, got %d", len(results))
	}
}
