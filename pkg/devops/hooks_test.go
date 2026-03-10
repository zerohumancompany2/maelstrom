package devops

import (
	"testing"
)

func TestIsolationHooks_ReplaceDefinition_Accept(t *testing.T) {
	// Given: A running tool instance with isolation level strict
	hooks := NewIsolationHooks()
	oldDef := &ToolDefinition{
		Name:           "test-tool",
		Signature:      "func(input string) (string, error)",
		Isolation:      IsolationStrict,
		Implementation: func(input string) (string, error) { return input, nil },
	}
	newDef := &ToolDefinition{
		Name:           "test-tool",
		Signature:      "func(input string) (string, error)",
		Isolation:      IsolationStrict,
		Implementation: func(input string) (string, error) { return "new: " + input, nil },
	}

	// When: replaceDefinition hook is called with new definition that maintains same signature
	err := hooks.ReplaceDefinition(oldDef, newDef)

	// Then: Running instance accepts the new definition and reloads without interruption
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Verify isolation boundary maintained
	if newDef.Isolation != IsolationStrict {
		t.Errorf("Expected isolation to remain strict, got %v", newDef.Isolation)
	}
}

func TestIsolationHooks_OnToolExecution_TaintPropagation(t *testing.T) {
	// Given: A tool execution hook with input data tainted ["USER_SUPPLIED"]
	hooks := NewIsolationHooks()
	input := map[string]interface{}{
		"_taints": []string{"USER_SUPPLIED"},
		"data":    "test input",
	}

	// When: onToolExecution hook processes the input and produces output
	output, err := hooks.OnToolExecution(input)

	// Then: Output data carries taint ["USER_SUPPLIED", "TOOL_OUTPUT"]
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	outputMap, ok := output.(map[string]interface{})
	if !ok {
		t.Fatal("Expected output to be map[string]interface{}")
	}

	taints, ok := outputMap["_taints"].([]string)
	if !ok {
		t.Fatal("Expected _taints field in output")
	}

	hasUserSupplied := false
	hasToolOutput := false
	for _, t := range taints {
		if t == "USER_SUPPLIED" {
			hasUserSupplied = true
		}
		if t == "TOOL_OUTPUT" {
			hasToolOutput = true
		}
	}

	if !hasUserSupplied {
		t.Error("Expected USER_SUPPLIED taint to be preserved")
	}

	if !hasToolOutput {
		t.Error("Expected TOOL_OUTPUT taint to be added")
	}
}
