package security

import (
	"testing"
)

func TestContextBlock_Audit_LOGS_VIOLATIONS(t *testing.T) {
	// Given: ContextBlock marked with taints=["SECRET"] and taintPolicy.enforcement=audit
	ClearAuditLog()

	block := ContextBlock{
		Name:    "secret-block",
		Source:  "session",
		Content: "Secret API key: sk-12345",
		Taints:  TaintSet{"SECRET": true},
		TaintPolicy: TaintPolicy{
			RedactMode: "audit",
		},
	}

	// When: block crosses to boundary where SECRET is forbidden (OuterBoundary)
	filtered, err := FilterContextBlock(block, OuterBoundary)

	// Then: violation is logged
	if err != nil {
		t.Fatalf("Expected no error (audit mode should not block), got: %v", err)
	}

	// Verify violation was logged to audit trail
	lastLog := GetLastAuditLog()
	if lastLog == "" {
		t.Fatal("Expected audit log to contain violation entry, but log is empty")
	}

	// Verify audit log entry contains required fields
	if !containsString(lastLog, "VIOLATION") {
		t.Errorf("Expected audit log to contain VIOLATION, got: %s", lastLog)
	}
	if !containsString(lastLog, "secret-block") {
		t.Errorf("Expected audit log to contain block ID 'secret-block', got: %s", lastLog)
	}
	if !containsString(lastLog, "outer") {
		t.Errorf("Expected audit log to contain boundary 'outer', got: %s", lastLog)
	}

	// Then: Block content still included in LLM prompt (not dropped/redacted)
	if filtered.Name != "secret-block" {
		t.Errorf("Expected block to pass through with name 'secret-block', got: %s", filtered.Name)
	}
	if filtered.Content != "Secret API key: sk-12345" {
		t.Errorf("Expected block content to be preserved, got: %s", filtered.Content)
	}
}
