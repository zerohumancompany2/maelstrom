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
