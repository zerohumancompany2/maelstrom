package security

import (
	"testing"
)

func TestPrepareContextForBoundary_Filter(t *testing.T) {
	contextBlockRegistry = make(map[string]BlockTaintInfo)

	innerBlock := &ContextBlock{
		Name: "inner-block",
		TaintPolicy: TaintPolicy{
			RedactMode: "strict",
		},
	}
	toolBlock := &ContextBlock{
		Name: "tool-block",
		TaintPolicy: TaintPolicy{
			RedactMode: "strict",
		},
	}

	contextBlockRegistry["inner-block"] = BlockTaintInfo{
		Block:  innerBlock,
		Taints: []string{"INNER_ONLY"},
	}
	contextBlockRegistry["tool-block"] = BlockTaintInfo{
		Block:  toolBlock,
		Taints: []string{"TOOL_OUTPUT"},
	}

	err := PrepareContextForBoundary("runtime-1", DMZBoundary)

	if err != nil {
		t.Fatalf("PrepareContextForBoundary returned error: %v", err)
	}

	if _, exists := contextBlockRegistry["inner-block"]; exists {
		t.Errorf("Expected INNER_ONLY blocks to be filtered out, but inner-block still exists")
	}
	if _, exists := contextBlockRegistry["tool-block"]; !exists {
		t.Errorf("Expected TOOL_OUTPUT block to be preserved, but tool-block was removed")
	}
}
