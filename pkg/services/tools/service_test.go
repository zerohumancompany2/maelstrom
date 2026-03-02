package tools

import (
	"testing"
)

func TestTools_Resolve(t *testing.T) {
	svc := NewToolsService()

	tool := ToolDescriptor{
		Name:      "resolve-test-tool",
		Boundary:  "outer",
		Schema:    map[string]any{"param": "string"},
		Isolation: "process",
	}

	err := svc.Register(tool)
	if err != nil {
		t.Fatalf("Register failed: %v", err)
	}

	resolved, err := svc.Resolve("resolve-test-tool", "inner")
	if err != nil {
		t.Fatalf("Resolve failed: %v", err)
	}

	if resolved.Name != "resolve-test-tool" {
		t.Errorf("Expected name 'resolve-test-tool', got '%s'", resolved.Name)
	}

	if resolved.Boundary != "outer" {
		t.Errorf("Expected boundary 'outer', got '%s'", resolved.Boundary)
	}

	if resolved.Isolation != "process" {
		t.Errorf("Expected isolation 'process', got '%s'", resolved.Isolation)
	}
}
