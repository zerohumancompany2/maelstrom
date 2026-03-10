package orchestrator

import (
	"testing"
)

func TestSequential_Continue_ExecutesAllTools(t *testing.T) {
	// Given
	executor := NewSequentialExecutor(PolicySeqContinue)
	toolCalls := []ToolCall{
		{Name: "tool1", Arguments: map[string]any{"arg": "value1"}},
		{Name: "tool2", Arguments: map[string]any{"arg": "value2"}},
		{Name: "tool3", Arguments: map[string]any{"arg": "value3"}},
	}

	// When
	results, err := executor.Execute(toolCalls)

	// Then
	if err != nil {
		t.Errorf("Expected Execute() to return nil error, got %v", err)
	}

	if len(results) != 3 {
		t.Errorf("Expected 3 results, got %d", len(results))
	}

	for i, result := range results {
		if result.Output == nil {
			t.Errorf("Expected result %d to have non-nil output", i)
		}
	}
}

func TestSequential_Continue_AccumulatesAllResults(t *testing.T) {
	// Given
	executor := NewSequentialExecutor(PolicySeqContinue)
	toolCalls := []ToolCall{
		{Name: "success-tool", Arguments: map[string]any{"arg": "value1"}},
		{Name: "fail-tool", Arguments: map[string]any{"should": "fail"}},
		{Name: "another-success", Arguments: map[string]any{"arg": "value3"}},
	}

	// When
	results, err := executor.Execute(toolCalls)

	// Then
	if err != nil {
		t.Errorf("Expected Execute() to return nil error, got %v", err)
	}

	if len(results) != 3 {
		t.Errorf("Expected 3 results, got %d", len(results))
	}

	expectedNames := []string{"success-tool", "fail-tool", "another-success"}
	for i, result := range results {
		output, ok := result.Output.(map[string]any)
		if !ok {
			t.Errorf("Expected result %d output to be map[string]any", i)
			continue
		}
		if tool, ok := output["tool"].(string); ok {
			if tool != expectedNames[i] {
				t.Errorf("Expected result %d to be for tool '%s', got '%s'", i, expectedNames[i], tool)
			}
		} else {
			t.Errorf("Expected result %d to have 'tool' field", i)
		}
	}
}

func TestSequential_FailFast_StopsOnFirstFailure(t *testing.T) {
	// Given
	executor := NewSequentialExecutor(PolicySeqFailFast)
	toolCalls := []ToolCall{
		{Name: "tool1", Arguments: map[string]any{"arg": "value1"}},
		{Name: "tool2-fail", Arguments: map[string]any{"should": "fail"}},
		{Name: "tool3", Arguments: map[string]any{"arg": "value3"}},
		{Name: "tool4", Arguments: map[string]any{"arg": "value4"}},
	}

	// When
	results, err := executor.Execute(toolCalls)

	// Then
	if err != nil {
		t.Errorf("Expected Execute() to return nil error, got %v", err)
	}

	if len(results) != 2 {
		t.Errorf("Expected 2 results (stopped at first failure), got %d", len(results))
	}

	if results[1].Error == nil {
		t.Error("Expected second result to have error (fail-fast stopped here)")
	}
}

func TestSequential_FailFast_ReturnsPartialResults(t *testing.T) {
	// Given
	executor := NewSequentialExecutor(PolicySeqFailFast)
	toolCalls := []ToolCall{
		{Name: "tool1-success", Arguments: map[string]any{"arg": "value1"}},
		{Name: "tool2-success", Arguments: map[string]any{"arg": "value2"}},
		{Name: "tool3-fail", Arguments: map[string]any{"should": "fail"}},
		{Name: "tool4", Arguments: map[string]any{"arg": "value4"}},
	}

	// When
	results, err := executor.Execute(toolCalls)

	// Then
	if err != nil {
		t.Errorf("Expected Execute() to return nil error, got %v", err)
	}

	if len(results) != 3 {
		t.Errorf("Expected 3 results (2 successful + 1 failure), got %d", len(results))
	}

	if results[0].Error != nil {
		t.Error("Expected first result to have no error")
	}

	if results[1].Error != nil {
		t.Error("Expected second result to have no error")
	}

	if results[2].Error == nil {
		t.Error("Expected third result to have error")
	}
}

func TestSequential_EmptyToolList_ReturnsEmptyResults(t *testing.T) {
	// Given
	executor := NewSequentialExecutor(PolicySeqContinue)
	toolCalls := []ToolCall{}

	// When
	results, err := executor.Execute(toolCalls)

	// Then
	if err != nil {
		t.Errorf("Expected Execute() to return nil error, got %v", err)
	}

	if len(results) != 0 {
		t.Errorf("Expected 0 results for empty tool list, got %d", len(results))
	}
}
