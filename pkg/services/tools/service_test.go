package tools

import (
	"testing"
)

func TestTools_Invoke(t *testing.T) {
	svc := NewToolsService()

	tool := ToolDescriptor{
		Name:      "invoke-test",
		Boundary:  "inner",
		Schema:    map[string]any{"type": "object", "properties": map[string]any{"input": map[string]any{"type": "string"}}},
		Isolation: "container",
	}

	err := svc.Register(tool)
	if err != nil {
		t.Fatalf("Register failed: %v", err)
	}

	result, err := svc.Invoke("invoke-test", map[string]any{"input": "test-data"}, "inner")
	if err != nil {
		t.Fatalf("Invoke failed: %v", err)
	}

	if result == nil {
		t.Fatal("Expected non-nil result")
	}
}
