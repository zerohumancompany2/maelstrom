package orchestrator

import (
	"context"
	"testing"

	"github.com/maelstrom/v3/internal/testutil"
)

func TestOrchestrator_ExecIsolated_SpawnsNewNamespace(t *testing.T) {
	// Given
	parentNs := "parent-chart-runtime-ns"
	orchestrator := New(parentNs)
	toolCall := &ToolCall{
		Name: "sys:admin:restart",
		Arguments: map[string]any{
			"target": "service-x",
		},
	}

	// When
	result, err := orchestrator.executeIsolated(context.Background(), toolCall)

	// Then
	if err != nil {
		t.Fatalf("executeIsolated returned unexpected error: %v", err)
	}

	if result == nil {
		t.Fatal("executeIsolated returned nil result")
	}

	if result.SourceNs == parentNs {
		t.Errorf("Expected isolated namespace to differ from parent, got same namespace: %s", result.SourceNs)
	}

	if result.SourceNs == "" {
		t.Error("Expected non-empty isolated namespace")
	}
}

func TestOrchestrator_ExecIsolated_ParentStateProtected(t *testing.T) {
	// Given
	parentNs := "parent-chart-runtime-ns"
	mockCtx := testutil.NewMockApplicationContext()
	mockCtx.Ns = parentNs

	sensitiveData := "secret-api-key-12345"
	mockCtx.Data["api_key"] = sensitiveData

	orchestrator := NewWithContext(parentNs, mockCtx)

	toolCall := &ToolCall{
		Name: "sys:admin:restart",
		Arguments: map[string]any{
			"target": "service-x",
		},
	}

	// When
	result, err := orchestrator.executeIsolated(context.Background(), toolCall)

	// Then
	if err != nil {
		t.Fatalf("executeIsolated returned unexpected error: %v", err)
	}

	if result == nil {
		t.Fatal("executeIsolated returned nil result")
	}

	if result.SourceNs == parentNs {
		t.Errorf("Expected isolated namespace to differ from parent, got same namespace: %s", result.SourceNs)
	}

	if result.CanAccessParent {
		t.Error("Expected isolated context to NOT have direct access to parent state")
	}

	if result.ParentNs != parentNs {
		t.Errorf("Expected parent namespace to be recorded as %s, got %s", parentNs, result.ParentNs)
	}
}
