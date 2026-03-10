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

func TestTools_Isolation(t *testing.T) {
	svc := NewToolsService()

	svc.Register(ToolDescriptor{Name: "isolated-tool", Boundary: "inner", Isolation: "container"})
	svc.Register(ToolDescriptor{Name: "strict-tool", Boundary: "dmz", Isolation: "strict"})

	result1, err := svc.Invoke("isolated-tool", map[string]any{"mode": "test"}, "inner")
	if err != nil {
		t.Fatalf("Invoke failed: %v", err)
	}

	result2, err := svc.Invoke("strict-tool", map[string]any{"mode": "test"}, "dmz")
	if err != nil {
		t.Fatalf("Invoke failed: %v", err)
	}

	if result1 == nil || result2 == nil {
		t.Fatal("Expected non-nil results")
	}

	m1, ok1 := result1.(map[string]any)
	m2, ok2 := result2.(map[string]any)
	if !ok1 || !ok2 {
		t.Fatal("Expected map results")
	}

	if m1["isolation"] != "container" {
		t.Errorf("Expected isolation 'container', got '%v'", m1["isolation"])
	}

	if m2["isolation"] != "strict" {
		t.Errorf("Expected isolation 'strict', got '%v'", m2["isolation"])
	}
}

func TestTools_Unregister(t *testing.T) {
	svc := NewToolsService()

	svc.Register(ToolDescriptor{Name: "unregister-test", Boundary: "inner", Isolation: "process"})

	_, err := svc.Resolve("unregister-test", "inner")
	if err != nil {
		t.Fatalf("Resolve before unregister failed: %v", err)
	}

	err = svc.Unregister("unregister-test")
	if err != nil {
		t.Fatalf("Unregister failed: %v", err)
	}

	_, err = svc.Resolve("unregister-test", "inner")
	if err != ErrToolNotFound {
		t.Fatalf("Expected ErrToolNotFound after unregister, got %v", err)
	}
}

func TestTools_NotFound(t *testing.T) {
	svc := NewToolsService()

	_, err := svc.Resolve("nonexistent-tool", "inner")
	if err != ErrToolNotFound {
		t.Fatalf("Expected ErrToolNotFound, got %v", err)
	}

	_, err = svc.Invoke("nonexistent-tool", map[string]any{}, "inner")
	if err != nil {
		t.Fatalf("Invoke for nonexistent tool failed: %v", err)
	}

	err = svc.Unregister("nonexistent-tool")
	if err != nil {
		t.Fatalf("Unregister nonexistent tool failed: %v", err)
	}
}

func TestToolsService_Register(t *testing.T) {
	svc := NewToolsService()

	err := svc.RegisterByName("test-tool", ToolDescriptor{Name: "test-tool"})
	if err != nil {
		t.Errorf("Expected nil error, got %v", err)
	}
}

func TestToolsService_Resolve(t *testing.T) {
	svc := NewToolsService()

	svc.RegisterByName("test-tool", ToolDescriptor{Name: "test-tool"})

	tool, err := svc.ResolveByName("test-tool")
	if err != nil {
		t.Errorf("Expected nil error, got %v", err)
	}

	if tool.Name != "test-tool" {
		t.Errorf("Expected name 'test-tool', got '%s'", tool.Name)
	}
}

func TestBoundaryAwareTool_ResolvesInnerTools(t *testing.T) {
	svc := NewToolsService()

	innerTool := ToolDescriptor{
		Name:      "innerDbQuery",
		Boundary:  "inner",
		Schema:    map[string]any{"type": "object", "properties": map[string]any{"query": map[string]any{"type": "string"}}},
		Isolation: "container",
	}

	err := svc.Register(innerTool)
	if err != nil {
		t.Fatalf("Register failed: %v", err)
	}

	resolved, err := svc.Resolve("innerDbQuery", "inner")
	if err != nil {
		t.Fatalf("Resolve failed for inner caller accessing inner tool: %v", err)
	}

	if resolved.Name != "innerDbQuery" {
		t.Errorf("Expected name 'innerDbQuery', got '%s'", resolved.Name)
	}

	if resolved.Boundary != "inner" {
		t.Errorf("Expected boundary 'inner', got '%s'", resolved.Boundary)
	}

	if len(resolved.Schema) == 0 {
		t.Error("Expected full schema to be returned for inner caller")
	}

	if _, ok := resolved.Schema["properties"]; !ok {
		t.Error("Expected schema to contain all fields for inner caller")
	}
}

func TestBoundaryAwareTool_BlocksOuterAccess(t *testing.T) {
	svc := NewToolsService()

	innerTool := ToolDescriptor{
		Name:      "innerDbQuery",
		Boundary:  "inner",
		Schema:    map[string]any{"type": "object"},
		Isolation: "container",
	}

	err := svc.Register(innerTool)
	if err != nil {
		t.Fatalf("Register failed: %v", err)
	}

	_, err = svc.Resolve("innerDbQuery", "outer")
	if err == nil {
		t.Fatal("Expected error when outer caller tries to access inner tool")
	}

	if err != ErrToolNotAccessible {
		t.Errorf("Expected ErrToolNotAccessible, got %v", err)
	}
}

func TestBoundaryAwareTool_EnforcesBoundary(t *testing.T) {
	svc := NewToolsService()

	dmzTool := ToolDescriptor{
		Name:      "webSearch",
		Boundary:  "dmz",
		Schema:    map[string]any{"type": "object", "properties": map[string]any{"query": map[string]any{"type": "string"}}},
		Isolation: "sandbox",
	}

	err := svc.Register(dmzTool)
	if err != nil {
		t.Fatalf("Register failed: %v", err)
	}

	resolved, err := svc.Resolve("webSearch", "dmz")
	if err != nil {
		t.Fatalf("Resolve failed for DMZ caller accessing DMZ tool: %v", err)
	}

	if resolved.Name != "webSearch" {
		t.Errorf("Expected name 'webSearch', got '%s'", resolved.Name)
	}

	if resolved.Boundary != "dmz" {
		t.Errorf("Expected boundary 'dmz', got '%s'", resolved.Boundary)
	}

	if len(resolved.Schema) == 0 {
		t.Error("Expected schema to be returned for DMZ caller")
	}

	if _, ok := resolved.Schema["properties"]; !ok {
		t.Error("Expected schema to contain DMZ-appropriate fields")
	}
}

func TestBoundaryAwareTool_ViolationOnBreach(t *testing.T) {
	svc := NewToolsService()

	innerTool := ToolDescriptor{
		Name:      "innerDbQuery",
		Boundary:  "inner",
		Schema:    map[string]any{"type": "object"},
		Isolation: "container",
	}

	err := svc.Register(innerTool)
	if err != nil {
		t.Fatalf("Register failed: %v", err)
	}

	_, err = svc.Resolve("innerDbQuery", "dmz")
	if err == nil {
		t.Fatal("Expected error when DMZ caller tries to access inner tool")
	}

	if err != ErrToolNotAccessible {
		t.Errorf("Expected ErrToolNotAccessible, got %v", err)
	}
}

// arch-v1.md L472: Tool registry and resolution
// arch-v1.md L488: resolve(name, callerBoundary) → ToolDescriptor | notFound
func TestToolRegistry_Register(t *testing.T) {
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

	err = svc.Register(tool)
	if err == nil {
		t.Fatal("Duplicate registration should return error, got nil")
	}
}

// arch-v1.md L472: Tool registry and resolution
// arch-v1.md L488: resolve(name, callerBoundary) → ToolDescriptor | notFound
func TestToolRegistry_Resolve(t *testing.T) {
	svc := NewToolsService()

	tool := ToolDescriptor{
		Name:      "resolve-test-tool",
		Boundary:  "inner",
		Schema:    map[string]any{"type": "object", "properties": map[string]any{"query": map[string]any{"type": "string"}}},
		Isolation: "container",
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

	if resolved.Boundary != "inner" {
		t.Errorf("Expected boundary 'inner', got '%s'", resolved.Boundary)
	}

	_, err = svc.Resolve("nonexistent-tool", "inner")
	if err != ErrToolNotFound {
		t.Errorf("Expected ErrToolNotFound for unknown tool, got %v", err)
	}
}
