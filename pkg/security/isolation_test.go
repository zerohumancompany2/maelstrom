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
