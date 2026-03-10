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

// arch-v1.md L472: Tool registry and resolution
// arch-v1.md L488: resolve(name, callerBoundary) → ToolDescriptor | notFound
func TestToolRegistry_ListTools(t *testing.T) {
	svc := NewToolsService()

	tools, err := svc.List("")
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}

	if len(tools) != 0 {
		t.Errorf("Expected empty list, got %d tools", len(tools))
	}

	tool1 := ToolDescriptor{
		Name:      "tool-1",
		Boundary:  "inner",
		Schema:    map[string]any{"type": "object"},
		Isolation: "container",
	}

	tool2 := ToolDescriptor{
		Name:      "tool-2",
		Boundary:  "outer",
		Schema:    map[string]any{"type": "object"},
		Isolation: "sandbox",
	}

	err = svc.Register(tool1)
	if err != nil {
		t.Fatalf("Register failed: %v", err)
	}

	err = svc.Register(tool2)
	if err != nil {
		t.Fatalf("Register failed: %v", err)
	}

	tools, err = svc.List("")
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}

	if len(tools) != 2 {
		t.Errorf("Expected 2 tools, got %d", len(tools))
	}
}

// arch-v1.md L472: Tool registry and resolution
// arch-v1.md L488: resolve(name, callerBoundary) → ToolDescriptor | notFound
// arch-v1.md L261-270: Boundary types (outer, DMZ, inner) and enforcement rules
func TestToolRegistry_Invoke(t *testing.T) {
	svc := NewToolsService()

	tool := ToolDescriptor{
		Name:      "invoke-test-tool",
		Boundary:  "inner",
		Schema:    map[string]any{"type": "object", "properties": map[string]any{"input": map[string]any{"type": "string"}}},
		Isolation: "container",
	}

	err := svc.Register(tool)
	if err != nil {
		t.Fatalf("Register failed: %v", err)
	}

	result, err := svc.Invoke("invoke-test-tool", map[string]any{"input": "test-data"}, "inner")
	if err != nil {
		t.Fatalf("Invoke failed: %v", err)
	}

	if result == nil {
		t.Fatal("Expected non-nil result")
	}

	resultMap, ok := result.(map[string]any)
	if !ok {
		t.Fatal("Expected result to be map[string]any")
	}

	if resultMap["tool"] != "invoke-test-tool" {
		t.Errorf("Expected tool name 'invoke-test-tool', got '%v'", resultMap["tool"])
	}

	_, err = svc.Invoke("nonexistent-tool", map[string]any{}, "inner")
	if err != nil {
		t.Errorf("Invoke for unknown tool should return nil error, got %v", err)
	}
}

// arch-v1.md L472: Tool registry and resolution
// arch-v1.md L488: resolve(name, callerBoundary) → ToolDescriptor | notFound
// arch-v1.md L261-270: Boundary types (outer, DMZ, inner) and enforcement rules
func TestToolRegistry_BoundaryFiltering(t *testing.T) {
	svc := NewToolsService()

	innerTool := ToolDescriptor{
		Name:      "inner-tool",
		Boundary:  "inner",
		Schema:    map[string]any{"type": "object"},
		Isolation: "container",
	}

	dmzTool := ToolDescriptor{
		Name:      "dmz-tool",
		Boundary:  "dmz",
		Schema:    map[string]any{"type": "object"},
		Isolation: "sandbox",
	}

	outerTool := ToolDescriptor{
		Name:      "outer-tool",
		Boundary:  "outer",
		Schema:    map[string]any{"type": "object"},
		Isolation: "sandbox",
	}

	if err := svc.Register(innerTool); err != nil {
		t.Fatalf("Register inner tool failed: %v", err)
	}
	if err := svc.Register(dmzTool); err != nil {
		t.Fatalf("Register dmz tool failed: %v", err)
	}
	if err := svc.Register(outerTool); err != nil {
		t.Fatalf("Register outer tool failed: %v", err)
	}

	tools, err := svc.List("inner")
	if err != nil {
		t.Fatalf("List for inner boundary failed: %v", err)
	}

	if len(tools) != 1 {
		t.Errorf("Expected 1 tool for inner boundary, got %d", len(tools))
	}

	if tools[0].Name != "inner-tool" {
		t.Errorf("Expected inner-tool, got %s", tools[0].Name)
	}

	tools, err = svc.List("dmz")
	if err != nil {
		t.Fatalf("List for dmz boundary failed: %v", err)
	}

	if len(tools) != 1 {
		t.Errorf("Expected 1 tool for dmz boundary, got %d", len(tools))
	}

	tools, err = svc.List("outer")
	if err != nil {
		t.Fatalf("List for outer boundary failed: %v", err)
	}

	if len(tools) != 1 {
		t.Errorf("Expected 1 tool for outer boundary, got %d", len(tools))
	}

	tools, err = svc.List("")
	if err != nil {
		t.Fatalf("List for all tools failed: %v", err)
	}

	if len(tools) != 3 {
		t.Errorf("Expected 3 tools for empty filter, got %d", len(tools))
	}
}

// arch-v1.md L472: Tool registry and resolution
// arch-v1.md L488: resolve(name, callerBoundary) → ToolDescriptor | notFound
// arch-v1.md L261-270: Boundary types (outer, DMZ, inner) and enforcement rules
func TestToolRegistry_InnerToolsInvisible(t *testing.T) {
	svc := NewToolsService()

	innerTool := ToolDescriptor{
		Name:      "inner-db-query",
		Boundary:  "inner",
		Schema:    map[string]any{"type": "object"},
		Isolation: "container",
	}

	dmzTool := ToolDescriptor{
		Name:      "dmz-web-search",
		Boundary:  "dmz",
		Schema:    map[string]any{"type": "object"},
		Isolation: "sandbox",
	}

	outerTool := ToolDescriptor{
		Name:      "outer-http-client",
		Boundary:  "outer",
		Schema:    map[string]any{"type": "object"},
		Isolation: "sandbox",
	}

	if err := svc.Register(innerTool); err != nil {
		t.Fatalf("Register inner tool failed: %v", err)
	}
	if err := svc.Register(dmzTool); err != nil {
		t.Fatalf("Register dmz tool failed: %v", err)
	}
	if err := svc.Register(outerTool); err != nil {
		t.Fatalf("Register outer tool failed: %v", err)
	}

	_, err := svc.Resolve("inner-db-query", "outer")
	if err == nil {
		t.Fatal("Expected error when outer caller tries to access inner tool")
	}

	if err != ErrToolNotAccessible {
		t.Errorf("Expected ErrToolNotAccessible, got %v", err)
	}

	_, err = svc.Resolve("inner-db-query", "dmz")
	if err == nil {
		t.Fatal("Expected error when dmz caller tries to access inner tool")
	}

	if err != ErrToolNotAccessible {
		t.Errorf("Expected ErrToolNotAccessible, got %v", err)
	}

	_, err = svc.Resolve("dmz-web-search", "outer")
	if err == nil {
		t.Fatal("Expected error when outer caller tries to access dmz tool")
	}

	if err != ErrToolNotAccessible {
		t.Errorf("Expected ErrToolNotAccessible, got %v", err)
	}

	_, err = svc.Resolve("dmz-web-search", "dmz")
	if err != nil {
		t.Errorf("DMZ caller should access DMZ tool: %v", err)
	}

	_, err = svc.Resolve("dmz-web-search", "inner")
	if err != nil {
		t.Errorf("Inner caller should access DMZ tool: %v", err)
	}

	_, err = svc.Resolve("outer-http-client", "outer")
	if err != nil {
		t.Errorf("Outer caller should access outer tool: %v", err)
	}

	_, err = svc.Resolve("outer-http-client", "dmz")
	if err != nil {
		t.Errorf("DMZ caller should access outer tool: %v", err)
	}

	_, err = svc.Resolve("outer-http-client", "inner")
	if err != nil {
		t.Errorf("Inner caller should access outer tool: %v", err)
	}
}

// arch-v1.md L472: Tool registry and resolution
// arch-v1.md L488: resolve(name, callerBoundary) → ToolDescriptor | notFound
// arch-v1.md L261-270: Boundary types (outer, DMZ, inner) and enforcement rules
func TestToolRegistry_BoundaryEnforcement(t *testing.T) {
	svc := NewToolsService()

	innerTool := ToolDescriptor{
		Name:      "inner-db-query",
		Boundary:  "inner",
		Schema:    map[string]any{"type": "object"},
		Isolation: "container",
	}

	if err := svc.Register(innerTool); err != nil {
		t.Fatalf("Register inner tool failed: %v", err)
	}

	_, err := svc.Resolve("inner-db-query", "outer")
	if err == nil {
		t.Fatal("Expected error when outer caller tries to access inner tool")
	}

	if err != ErrToolNotAccessible {
		t.Errorf("Expected ErrToolNotAccessible, got %v", err)
	}

	_, err = svc.Resolve("inner-db-query", "dmz")
	if err == nil {
		t.Fatal("Expected error when dmz caller tries to access inner tool")
	}

	if err != ErrToolNotAccessible {
		t.Errorf("Expected ErrToolNotAccessible, got %v", err)
	}

	_, err = svc.Resolve("inner-db-query", "inner")
	if err != nil {
		t.Errorf("Inner caller should access inner tool: %v", err)
	}
}

// arch-v1.md L562, L566: TaintOutput field for tool descriptors
func TestToolDescriptor_HasTaintOutputField(t *testing.T) {
	tool := ToolDescriptor{
		Name:        "webSearch",
		Boundary:    "dmz",
		Schema:      map[string]any{"type": "object"},
		Isolation:   "container",
		TaintOutput: []string{"TOOL_OUTPUT"},
	}

	if len(tool.TaintOutput) != 1 {
		t.Errorf("Expected 1 taint output, got %d", len(tool.TaintOutput))
	}

	if tool.TaintOutput[0] != "TOOL_OUTPUT" {
		t.Errorf("Expected 'TOOL_OUTPUT', got '%s'", tool.TaintOutput[0])
	}

	innerTool := ToolDescriptor{
		Name:        "innerDbQuery",
		Boundary:    "inner",
		Schema:      map[string]any{"type": "object"},
		Isolation:   "strict",
		TaintOutput: []string{"TOOL_OUTPUT", "INNER_ONLY"},
	}

	if len(innerTool.TaintOutput) != 2 {
		t.Errorf("Expected 2 taint outputs, got %d", len(innerTool.TaintOutput))
	}
}

// arch-v1.md L568: Type field for sub-agent identification
func TestToolDescriptor_TypeField(t *testing.T) {
	tool := ToolDescriptor{
		Name: "subAgentResearch",
		Type: "attachedSubAgent",
	}

	if tool.Type != "attachedSubAgent" {
		t.Errorf("Expected 'attachedSubAgent', got '%s'", tool.Type)
	}
}

// arch-v1.md L569: ChartRef field for sub-agent chart reference
func TestToolDescriptor_ChartRefField(t *testing.T) {
	tool := ToolDescriptor{
		Name:     "subAgentResearch",
		Type:     "attachedSubAgent",
		ChartRef: "research-mini-v1",
	}

	if tool.ChartRef != "research-mini-v1" {
		t.Errorf("Expected 'research-mini-v1', got '%s'", tool.ChartRef)
	}
}

// arch-v1.md L569: MaxIterations field for sub-agent iteration limits
func TestToolDescriptor_MaxIterationsField(t *testing.T) {
	tool := ToolDescriptor{
		Name:          "subAgentResearch",
		Type:          "attachedSubAgent",
		ChartRef:      "research-mini-v1",
		MaxIterations: 10,
	}

	if tool.MaxIterations != 10 {
		t.Errorf("Expected 10, got %d", tool.MaxIterations)
	}
}
