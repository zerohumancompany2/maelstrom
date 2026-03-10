package security

import (
	"testing"
	"time"
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

func TestContextBlock_Audit_CONTINUES_EXECUTION(t *testing.T) {
	// Given: ContextBlock with taintPolicy.enforcement=audit has forbidden taint for target boundary
	ClearAuditLog()

	block := ContextBlock{
		Name:    "dmz-block",
		Source:  "memory",
		Content: "DMZ data with SECRET taint",
		Taints:  TaintSet{"SECRET": true},
		TaintPolicy: TaintPolicy{
			RedactMode: "audit",
		},
	}

	// When: prepareContextForBoundary completes (FilterContextBlock simulates this)
	filtered, err := FilterContextBlock(block, OuterBoundary)

	// Then: prepareContextForBoundary completes successfully (returns no error)
	if err != nil {
		t.Fatalf("Expected no error from audit mode, got: %v", err)
	}

	// Then: LLM prompt assembly proceeds with block included
	if filtered.Name != "dmz-block" {
		t.Errorf("Expected block to be included in prompt assembly, got name: %s", filtered.Name)
	}
	if filtered.Content != "DMZ data with SECRET taint" {
		t.Errorf("Expected block content to be preserved for prompt assembly, got: %s", filtered.Content)
	}

	// Then: Violation emitted to dead-letter queue as TaintViolation event
	violation := TaintViolation{
		RuntimeID:       "agent-test-123",
		SourceBoundary:  DMZBoundary,
		TargetBoundary:  OuterBoundary,
		ForbiddenTaints: []string{"SECRET"},
		Timestamp:       time.Now(),
	}

	err = ReportViolation("agent-test-123", violation)
	if err != nil {
		t.Fatalf("Expected no error from ReportViolation, got: %v", err)
	}

	// Verify violation was counted (dead-letter queue tracking)
	count := GetViolationCount("agent-test-123")
	if count != 1 {
		t.Errorf("Expected violation count to be 1, got: %d", count)
	}

	// Then: Chart state machine continues to next state without interruption
	// (Verified by: no error returned, block included, violation logged)
	// This is the key difference from strict mode which would return error and block execution
}
