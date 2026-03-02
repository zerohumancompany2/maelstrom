package memory

import (
	"testing"
)

func TestMemory_Store(t *testing.T) {
	// Arrange
	svc := NewMemoryService()

	// Act
	_, err := svc.Store("runtime-1", "test content", map[string]any{
		"key": "value",
	})

	// Assert
	if err != nil {
		t.Fatalf("Store failed: %v", err)
	}
}
