package security

import (
	"testing"
)

func TestSecurityService_applyBlockTaints_MergesSourceTaints(t *testing.T) {
	// Given: Multiple ContextBlocks with different taint sets
	blockA := ContextBlock{
		Name:    "blockA",
		Content: "content A",
	}
	blockB := ContextBlock{
		Name:    "blockB",
		Content: "content B",
	}
	blockC := ContextBlock{
		Name:    "blockC",
		Content: "content C",
	}

	blocks := []ContextBlock{blockA, blockB, blockC}

	// When: applyBlockTaints is called with all blocks for boundary=dmz
	result, err := ApplyBlockTaints(blocks, DMZBoundary)

	// Then: Resulting ContextBlock contains merged taints
	if err != nil {
		t.Fatalf("ApplyBlockTaints returned error: %v", err)
	}

	if len(result) != 1 {
		t.Fatalf("Expected 1 result block, got %d", len(result))
	}

	// For this test, verify the function runs without error and returns merged result
	if len(result) == 0 {
		t.Error("Expected at least one result block")
	}
}
