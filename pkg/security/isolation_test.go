package security

import (
	"testing"
)

func TestNamespaceIsolate_CreateView(t *testing.T) {
	runtimeId := "agent-123"
	operation := "read"

	view, err := NamespaceIsolate(runtimeId, operation)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if view.RuntimeID != runtimeId {
		t.Errorf("Expected RuntimeID %s, got %s", runtimeId, view.RuntimeID)
	}

	if view.Operation != operation {
		t.Errorf("Expected Operation %s, got %s", operation, view.Operation)
	}
}

func TestNamespaceIsolate_FilterData(t *testing.T) {
	view, _ := NamespaceIsolate("agent-123", "read")
	view.Boundary = DMZBoundary

	data := map[string]interface{}{
		"_taints": []string{"INNER_ONLY", "TOOL_OUTPUT"},
		"secret":  "should be filtered",
		"public":  "should remain",
	}

	result := view.FilterData(data)
	resultMap, ok := result.(map[string]interface{})
	if !ok {
		t.Fatal("Expected FilterData to return map[string]interface{}")
	}

	taints, ok := resultMap["_taints"].([]string)
	if !ok {
		t.Fatal("Expected _taints field in result")
	}

	hasInnerOnly := false
	hasToolOutput := false
	for _, t := range taints {
		if t == "INNER_ONLY" {
			hasInnerOnly = true
		}
		if t == "TOOL_OUTPUT" {
			hasToolOutput = true
		}
	}

	if hasInnerOnly {
		t.Error("INNER_ONLY taint should be filtered on DMZ boundary")
	}

	if !hasToolOutput {
		t.Error("TOOL_OUTPUT taint should remain on DMZ boundary")
	}
}
