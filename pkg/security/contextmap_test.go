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

func TestFilterContextBlock_Redact(t *testing.T) {
	contextBlockRegistry = make(map[string]BlockTaintInfo)

	block := ContextBlock{
		Name:    "secret-block",
		Content: "This contains SECRET data",
		TaintPolicy: TaintPolicy{
			RedactMode: "redact",
			RedactRules: []RedactRule{
				{Taint: "SECRET", Replacement: "[REDACTED]"},
			},
		},
	}

	filtered, err := FilterContextBlock(block, OuterBoundary)

	if err != nil {
		t.Fatalf("FilterContextBlock returned error: %v", err)
	}

	if filtered.Content != "This contains [REDACTED] data" {
		t.Errorf("Expected SECRET to be redacted, got: %s", filtered.Content)
	}
	if filtered.Name != "secret-block" {
		t.Errorf("Expected block structure to be preserved, got name: %s", filtered.Name)
	}
}

func TestFilterContextBlock_DropBlock(t *testing.T) {
	contextBlockRegistry = make(map[string]BlockTaintInfo)

	block := ContextBlock{
		Name:    "pii-block",
		Content: "This contains PII data",
		TaintPolicy: TaintPolicy{
			RedactMode: "dropBlock",
		},
	}

	filtered, err := FilterContextBlock(block, OuterBoundary)

	if err != nil {
		t.Fatalf("FilterContextBlock returned error: %v", err)
	}

	if filtered.Name != "" {
		t.Errorf("Expected block to be dropped (empty name), got: %s", filtered.Name)
	}
	if filtered.Content != "" {
		t.Errorf("Expected block content to be empty after drop, got: %s", filtered.Content)
	}
}

func TestFilterContextBlock_PerBlockOverride(t *testing.T) {
	ClearAuditLog()
	contextBlockRegistry = make(map[string]BlockTaintInfo)

	block := ContextBlock{
		Name:    "audit-block",
		Content: "This contains SECRET data",
		TaintPolicy: TaintPolicy{
			RedactMode: "audit",
		},
	}

	globalPolicy := TaintPolicyConfig{
		Enforcement: EnforcementStrict,
	}

	filtered, err := FilterContextBlockWithGlobalPolicy(block, DMZBoundary, globalPolicy)

	if err != nil {
		t.Fatalf("FilterContextBlockWithGlobalPolicy returned error (per-block audit should override global strict): %v", err)
	}

	if filtered.Name != "audit-block" {
		t.Errorf("Expected per-block audit mode to allow block through, got name: %s", filtered.Name)
	}
	if filtered.Content != "This contains SECRET data" {
		t.Errorf("Expected block content to be preserved with audit mode, got: %s", filtered.Content)
	}
}

func TestFilterContextBlock_AllowedForBoundary(t *testing.T) {
	contextBlockRegistry = make(map[string]BlockTaintInfo)

	block := ContextBlock{
		Name:    "inner-only-block",
		Content: "This contains SECRET data",
		TaintPolicy: TaintPolicy{
			AllowedForBoundary: []BoundaryType{InnerBoundary},
		},
	}

	filtered, err := FilterContextBlock(block, DMZBoundary)

	if err != nil {
		t.Fatalf("FilterContextBlock returned error: %v", err)
	}

	if filtered.Name != "" {
		t.Errorf("Expected block to be filtered out (dmz not in allowedForBoundary), got name: %s", filtered.Name)
	}
}
