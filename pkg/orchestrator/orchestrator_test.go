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
