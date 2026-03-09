package orchestrator

import (
	"context"
	"testing"
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
