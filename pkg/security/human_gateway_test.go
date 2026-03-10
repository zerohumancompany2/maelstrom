package security

import "testing"

func TestHumanGatewaySanitization_SanitizeContextMap_StripsForbiddenTaints(t *testing.T) {
	ctx := ContextMap{
		Blocks: []*ContextBlock{
			{Name: "secret_block", Content: "secret data", Taints: TaintSet{"SECRET": true}},
			{Name: "inner_block", Content: "inner data", Taints: TaintSet{"INNER_ONLY": true}},
			{Name: "pii_block", Content: "user@email.com", Taints: TaintSet{"PII": true}},
			{Name: "safe_block", Content: "safe data", Taints: TaintSet{}},
		},
	}

	result, err := SanitizeContextMap(ctx, DMZBoundary)
	if err != nil {
		t.Fatalf("SanitizeContextMap returned error: %v", err)
	}

	if len(result.Blocks) != 2 {
		t.Errorf("SanitizeContextMap returned %d blocks, want 2 (SECRET and INNER_ONLY stripped)", len(result.Blocks))
	}

	hasSafe := false
	hasPII := false
	for _, block := range result.Blocks {
		if block.Name == "safe_block" {
			hasSafe = true
		}
		if block.Name == "pii_block" {
			hasPII = true
			if block.Content != "[REDACTED]" {
				t.Errorf("PII block content = %q, want [REDACTED]", block.Content)
			}
		}
	}

	if !hasSafe {
		t.Error("SanitizeContextMap did not include safe_block")
	}
	if !hasPII {
		t.Error("SanitizeContextMap did not include pii_block (should be redacted, not stripped)")
	}
}

func TestHumanGatewaySanitization_SanitizeMessageHistory_BoundaryEnforcement(t *testing.T) {
	panic("not implemented")
}
