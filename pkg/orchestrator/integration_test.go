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
