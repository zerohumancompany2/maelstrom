package security

import (
	"strings"
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

func TestNamespaceIsolate_CreatesIsolatedContext(t *testing.T) {
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

	if view.Boundary != DMZBoundary {
		t.Errorf("Expected Boundary DMZ, got %s", view.Boundary)
	}

	if view.ContextData == nil {
		t.Error("Expected ContextData to be initialized")
	}
}

func TestNamespaceIsolate_BlocksCrossNamespaceAccess(t *testing.T) {
	view, _ := NamespaceIsolate("agent-123", "read")
	view.Boundary = DMZBoundary

	err := view.AccessData("agent-456")
	if err == nil {
		t.Fatal("Expected error for cross-namespace access, got nil")
	}

	if !strings.Contains(err.Error(), "namespace violation") {
		t.Errorf("Expected namespace violation error, got: %v", err)
	}
}

func TestNamespaceIsolate_Operation(t *testing.T) {
	t.Run("WriteOperation", func(t *testing.T) {
		view, _ := NamespaceIsolate("agent-123", "write")

		operation := view.GetOperation()
		if operation != "write" {
			t.Errorf("Expected operation 'write', got %s", operation)
		}

		if !view.IsWriteOperation() {
			t.Error("Expected IsWriteOperation to return true for write operation")
		}

		if view.IsReadOperation() {
			t.Error("Expected IsReadOperation to return false for write operation")
		}
	})

	t.Run("ReadOperation", func(t *testing.T) {
		view, _ := NamespaceIsolate("agent-123", "read")

		operation := view.GetOperation()
		if operation != "read" {
			t.Errorf("Expected operation 'read', got %s", operation)
		}

		if view.IsWriteOperation() {
			t.Error("Expected IsWriteOperation to return false for read operation")
		}

		if !view.IsReadOperation() {
			t.Error("Expected IsReadOperation to return true for read operation")
		}
	})

	t.Run("ListOperation", func(t *testing.T) {
		view, _ := NamespaceIsolate("agent-123", "list")

		operation := view.GetOperation()
		if operation != "list" {
			t.Errorf("Expected operation 'list', got %s", operation)
		}

		if view.IsWriteOperation() {
			t.Error("Expected IsWriteOperation to return false for list operation")
		}

		if !view.IsReadOperation() {
			t.Error("Expected IsReadOperation to return true for list operation")
		}
	})
}
