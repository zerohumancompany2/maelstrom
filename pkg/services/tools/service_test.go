package tools

import (
	"testing"
)

func TestTools_Register(t *testing.T) {
	svc := NewToolsService()

	tool := ToolDescriptor{
		Name:      "test-tool",
		Boundary:  "inner",
		Schema:    map[string]any{"type": "object"},
		Isolation: "container",
	}

	err := svc.Register(tool)
	if err != nil {
		t.Fatalf("Register failed: %v", err)
	}

	// Verify tool was registered by resolving it
	resolved, err := svc.Resolve("test-tool", "inner")
	if err != nil {
		t.Fatalf("Resolve failed: %v", err)
	}

	if resolved.Name != "test-tool" {
		t.Errorf("Expected name 'test-tool', got '%s'", resolved.Name)
	}

	if resolved.Boundary != "inner" {
		t.Errorf("Expected boundary 'inner', got '%s'", resolved.Boundary)
	}
}
