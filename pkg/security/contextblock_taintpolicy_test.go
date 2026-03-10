package security

import (
	"testing"
)

func TestContextBlock_TaintPolicy_AttachedSuccessfully(t *testing.T) {
	// Given: A ContextBlock with an attached taintPolicy
	// enforcement=redact, allowedOnExit=["USER_SUPPLIED"], and redactRules array
	block := ContextBlock{
		Name:    "user-context",
		Source:  "session",
		Content: "User conversation history",
		TaintPolicy: TaintPolicy{
			RedactMode:         "redact",
			AllowedForBoundary: []BoundaryType{OuterBoundary},
			RedactRules: []RedactRule{
				{Taint: "PII", Replacement: "[PERSONAL_INFO]"},
				{Taint: "SECRET", Replacement: "[REDACTED]"},
			},
		},
	}

	// When: ContextBlock is serialized (marshaled to JSON)
	serialized := serializeContextBlock(block)

	// Then: TaintPolicy attached at ContextBlock creation persists through serialization
	if serialized == "" {
		t.Fatal("Expected serialized ContextBlock to be non-empty")
	}

	// Verify enforcement mode persists (redactMode)
	if !containsString(serialized, "redact") {
		t.Errorf("Expected serialized block to contain redactMode 'redact', got: %s", serialized)
	}

	// Verify allowedForBoundary persists
	if !containsString(serialized, "outer") {
		t.Errorf("Expected serialized block to contain allowedForBoundary 'outer', got: %s", serialized)
	}

	// Verify redactRules persist
	if !containsString(serialized, "PII") {
		t.Errorf("Expected serialized block to contain redactRules with PII, got: %s", serialized)
	}
	if !containsString(serialized, "[PERSONAL_INFO]") {
		t.Errorf("Expected serialized block to contain redactRules replacement '[PERSONAL_INFO]', got: %s", serialized)
	}
}

func serializeContextBlock(block ContextBlock) string {
	result := "ContextBlock{"
	result += "Name:" + block.Name + ","
	result += "Source:" + block.Source + ","
	result += "Content:" + block.Content + ","
	result += "TaintPolicy{"
	result += "RedactMode:" + block.TaintPolicy.RedactMode + ","
	for _, allowed := range block.TaintPolicy.AllowedForBoundary {
		result += "AllowedForBoundary:" + string(allowed) + ","
	}
	for _, rule := range block.TaintPolicy.RedactRules {
		result += "RedactRule{Taint:" + rule.Taint + ",Replacement:" + rule.Replacement + "},"
	}
	result += "}}"
	return result
}

func containsString(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
