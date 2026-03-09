package security

import (
	"testing"
)

func TestSecurityService_applyBlockTaints_MergesSourceTaints(t *testing.T) {
	// Given: Multiple ContextBlocks with different taint sets
	blockA := ContextBlock{
		Name:    "blockA",
		Content: "content A",
		Taints:  TaintSet{"USER_SUPPLIED": true},
	}
	blockB := ContextBlock{
		Name:    "blockB",
		Content: "content B",
		Taints:  TaintSet{"TOOL_OUTPUT": true},
	}
	blockC := ContextBlock{
		Name:    "blockC",
		Content: "content C",
		Taints:  TaintSet{"MEMORY_INJECTED": true},
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

	if !result[0].Taints.Has("USER_SUPPLIED") {
		t.Error("Expected merged taints to contain USER_SUPPLIED")
	}
	if !result[0].Taints.Has("TOOL_OUTPUT") {
		t.Error("Expected merged taints to contain TOOL_OUTPUT")
	}
	if !result[0].Taints.Has("MEMORY_INJECTED") {
		t.Error("Expected merged taints to contain MEMORY_INJECTED")
	}
}

func TestSecurityService_applyBlockTaints_EnforcesTaintPolicy(t *testing.T) {
	// Given: ContextBlocks with mixed taints for boundary=outer
	blockA := ContextBlock{
		Name:    "blockA",
		Content: "content A",
		Taints: TaintSet{
			"USER_SUPPLIED": true,
			"INNER_ONLY":    true,
		},
	}
	blockB := ContextBlock{
		Name:    "blockB",
		Content: "content B",
		Taints:  TaintSet{"TOOL_OUTPUT": true},
	}

	blocks := []ContextBlock{blockA, blockB}

	// When: applyBlockTaints is called with blocks for boundary=outer (INNER_ONLY forbidden)
	result, err := ApplyBlockTaints(blocks, OuterBoundary)

	// Then: Block A is dropped/redacted (contains forbidden INNER_ONLY), Block B passes through
	if err != nil {
		t.Fatalf("ApplyBlockTaints returned error: %v", err)
	}

	if len(result) != 1 {
		t.Fatalf("Expected 1 result block (blockA dropped), got %d", len(result))
	}

	if result[0].Name != "blockB" {
		t.Errorf("Expected blockB to pass through, got %s", result[0].Name)
	}

	if result[0].Taints.Has("INNER_ONLY") {
		t.Error("Expected INNER_ONLY taint to be filtered out")
	}
}
