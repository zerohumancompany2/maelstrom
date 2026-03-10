package security

import (
	"testing"
)

func TestContextBlock_Redaction_AppliesPIIRules(t *testing.T) {
	// Given: A ContextBlock marked with taints=["PII"] and taintPolicy.redactRules configured
	ClearAuditLog()
	contextBlockRegistry = make(map[string]BlockTaintInfo)

	piiBlock := &ContextBlock{
		Name:    "user-email",
		Content: "User email: john.doe@example.com",
		TaintPolicy: TaintPolicy{
			RedactMode: "redact",
			RedactRules: []RedactRule{
				{Taint: "PII", Replacement: "[REDACTED]"},
			},
		},
	}

	contextBlockRegistry["user-email"] = BlockTaintInfo{
		Block:  piiBlock,
		Taints: []string{"PII"},
	}

	// When: prepareContextForBoundary is called
	err := PrepareContextForBoundary("runtime-1", OuterBoundary)

	// Then: PII content is replaced per rules
	if err != nil {
		t.Fatalf("PrepareContextForBoundary returned error: %v", err)
	}

	info, exists := contextBlockRegistry["user-email"]
	if !exists {
		t.Fatal("Expected PII block to remain in context after redaction")
	}

	// Redacted content logged to audit trail with original taint markers preserved in metadata
	if info.Block.Content == "User email: john.doe@example.com" {
		t.Error("Expected PII content to be redacted, but original email still present")
	}

	lastLog := GetLastAuditLog()
	if lastLog == "" {
		t.Error("Expected redaction to be logged to audit trail, but log is empty")
	}
}

func TestContextBlock_Redaction_DROPS_FORBIDDEN_TAINTS(t *testing.T) {
	// Given: A ContextBlock marked with taints=["INNER_ONLY"] attempting to cross to DMZ boundary
	ClearAuditLog()
	contextBlockRegistry = make(map[string]BlockTaintInfo)

	innerBlock := &ContextBlock{
		Name:    "inner-secret",
		Content: "This is inner-only data",
		TaintPolicy: TaintPolicy{
			RedactMode: "redact",
		},
	}

	contextBlockRegistry["inner-secret"] = BlockTaintInfo{
		Block:  innerBlock,
		Taints: []string{"INNER_ONLY"},
	}

	// When: prepareContextForBoundary is called
	err := PrepareContextForBoundary("runtime-1", DMZBoundary)

	// Then: entire block is removed from prompt
	if err != nil {
		t.Fatalf("PrepareContextForBoundary returned error: %v", err)
	}

	if _, exists := contextBlockRegistry["inner-secret"]; exists {
		t.Error("Expected INNER_ONLY block to be dropped from DMZ boundary, but block still exists")
	}

	// Drop action logged with reason: "forbidden taint INNER_ONLY for DMZ boundary"
	lastLog := GetLastAuditLog()
	if lastLog == "" {
		t.Error("Expected drop action to be logged to audit trail, but log is empty")
	}
}
