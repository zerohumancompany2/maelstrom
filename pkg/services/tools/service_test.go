package tools

import (
	"testing"
)

func TestTools_BoundaryFilter(t *testing.T) {
	svc := NewToolsService()

	svc.Register(ToolDescriptor{Name: "inner-tool-1", Boundary: "inner", Isolation: "container"})
	svc.Register(ToolDescriptor{Name: "inner-tool-2", Boundary: "inner", Isolation: "process"})
	svc.Register(ToolDescriptor{Name: "outer-tool-1", Boundary: "outer", Isolation: "sandbox"})

	tools, err := svc.List("inner")
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}

	if len(tools) != 2 {
		t.Errorf("Expected 2 inner tools, got %d", len(tools))
	}

	for _, tool := range tools {
		if tool.Boundary != "inner" {
			t.Errorf("Expected boundary 'inner', got '%s'", tool.Boundary)
		}
	}
}
