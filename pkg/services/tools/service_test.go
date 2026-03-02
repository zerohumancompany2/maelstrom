package tools

import (
	"testing"
)

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
