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
