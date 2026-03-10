package orchestrator

import (
	"testing"

	"github.com/maelstrom/v3/pkg/services/tools"
)

func TestOrchestratorService_NewCreatesService(t *testing.T) {
	// Given
	// No setup needed

	// When
	service := NewOrchestratorService()

	// Then
	if service == nil {
		t.Fatal("Expected NewOrchestratorService() to return non-nil service")
	}

	if service.id != "sys:orchestrator" {
		t.Errorf("Expected service ID to be 'sys:orchestrator', got '%s'", service.id)
	}

	if service.policies == nil {
		t.Error("Expected service policies map to be initialized")
	}
}

func TestOrchestratorService_RegisterPolicy(t *testing.T) {
	// Given
	service := NewOrchestratorService()
	policy := ExecutionPolicy{
		Mode:        "test_policy",
		MaxRetries:  3,
		Isolation:   "strict",
		MaxParallel: 4,
		TimeoutMs:   5000,
	}

	// When
	err := service.RegisterPolicy("test", policy)

	// Then
	if err != nil {
		t.Errorf("Expected RegisterPolicy() to return nil error, got %v", err)
	}

	retrieved, exists := service.policies["test"]
	if !exists {
		t.Error("Expected policy to be stored with key 'test'")
	}

	if retrieved == nil {
		t.Error("Expected retrieved policy to be non-nil")
	} else if retrieved.Mode != policy.Mode {
		t.Errorf("Expected policy Mode to be '%s', got '%s'", policy.Mode, retrieved.Mode)
	}
}

func TestOrchestratorService_ResolveTool(t *testing.T) {
	// Given
	expectedTool := tools.ToolDescriptor{
		Name:     "test-tool",
		Boundary: "outer",
		Schema:   map[string]any{"type": "object"},
	}

	// Create tools service and register a tool
	toolsService := tools.NewToolsService()
	_ = toolsService.Register(expectedTool)

	// Create orchestrator with tools service
	service := NewOrchestratorServiceWithTools(toolsService)

	// When
	result, err := service.ResolveTool("test-tool", "outer")

	// Then
	if err != nil {
		t.Errorf("Expected ResolveTool() to return nil error, got %v", err)
	}

	if result.Name != expectedTool.Name {
		t.Errorf("Expected tool Name to be '%s', got '%s'", expectedTool.Name, result.Name)
	}
}

func TestOrchestratorService_ResolveToolWithBoundary(t *testing.T) {
	// Given
	// Register an inner-boundary tool
	innerTool := tools.ToolDescriptor{
		Name:     "inner-tool",
		Boundary: "inner",
		Schema:   map[string]any{"type": "object"},
	}

	toolsService := tools.NewToolsService()
	_ = toolsService.Register(innerTool)

	service := NewOrchestratorServiceWithTools(toolsService)

	// When - outer boundary tries to access inner tool
	_, err := service.ResolveTool("inner-tool", "outer")

	// Then - should fail due to boundary restriction
	if err == nil {
		t.Error("Expected ResolveTool() to return error for boundary violation")
	}
}
