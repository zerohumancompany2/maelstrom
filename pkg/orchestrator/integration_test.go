package orchestrator

import (
	"testing"
)

func TestIntegration_Orchestrator_EndToEndSequentialFlow(t *testing.T) {
	// Given - Create OrchestratorService with policies registered
	service := NewOrchestratorService()
	_ = service.RegisterPolicy("seq_continue", PolicySeqContinue)
	_ = service.RegisterPolicy("seq_failfast", PolicySeqFailFast)

	// Register test tools
	tools := []ToolCall{
		{Name: "tool-alpha", Arguments: map[string]any{"step": 1}},
		{Name: "tool-beta", Arguments: map[string]any{"step": 2}},
		{Name: "tool-gamma", Arguments: map[string]any{"step": 3}},
	}

	// When - Execute sequential flow with continue policy
	executor := NewSequentialExecutor(PolicySeqContinue)
	results, err := executor.Execute(tools)

	// Then - Verify all tools executed and results collected
	if err != nil {
		t.Fatalf("Expected Execute() to return nil error, got %v", err)
	}

	if len(results) != 3 {
		t.Errorf("Expected 3 results, got %d", len(results))
	}

	expectedNames := []string{"tool-alpha", "tool-beta", "tool-gamma"}
	for i, result := range results {
		if result.Output == nil {
			t.Errorf("Expected result %d to have non-nil output", i)
			continue
		}

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

func TestIntegration_Orchestrator_EndToEndParallelWithFailure(t *testing.T) {
	// Given - Create OrchestratorService with policies registered
	service := NewOrchestratorService()
	_ = service.RegisterPolicy("par_continue", PolicyParContinue)

	// Register test tools (some succeed, some fail)
	tools := []ToolCall{
		{Name: "success-tool-alpha", Arguments: map[string]any{"value": "result1"}},
		{Name: "fail-tool-beta", Arguments: map[string]any{"should": "fail"}},
		{Name: "success-tool-gamma", Arguments: map[string]any{"value": "result2"}},
	}

	// When - Execute parallel flow with continue policy
	executor := NewParallelExecutor(PolicyParContinue)
	resultChan, err := executor.Execute(tools)

	// Then - Verify all tools executed concurrently despite failures
	if err != nil {
		t.Fatalf("Expected Execute() to return nil error, got %v", err)
	}

	// Collect all results
	results := make([]ExecutionResultWithTool, 0, 3)
	for result := range resultChan {
		results = append(results, result)
	}

	// All tools should execute even if one fails
	if len(results) != 3 {
		t.Errorf("Expected 3 results (all tools executed), got %d", len(results))
	}

	// Verify results collected and ordered correctly by tool name
	expectedOrder := []string{"fail-tool-beta", "success-tool-alpha", "success-tool-gamma"}
	for i, expected := range expectedOrder {
		if results[i].ToolName != expected {
			t.Errorf("Expected result %d to be for tool '%s', got '%s'", i, expected, results[i].ToolName)
		}
	}

	// Verify all results have output
	for i, result := range results {
		if result.Output == nil {
			t.Errorf("Expected result %d to have non-nil output", i)
		}
	}
}
