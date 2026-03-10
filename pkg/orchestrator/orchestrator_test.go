package orchestrator

import (
	"testing"
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
