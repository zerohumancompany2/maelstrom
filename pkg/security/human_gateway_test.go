package security

import (
	"testing"

	"github.com/maelstrom/v3/pkg/mail"
)

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
	messages := []mail.Mail{
		{ID: "1", Type: mail.MailTypeUser, Content: "user message", Metadata: mail.MailMetadata{Taints: []string{}}},
		{ID: "2", Type: mail.MailTypeToolResult, Content: "secret tool output", Metadata: mail.MailMetadata{Taints: []string{"SECRET"}}},
		{ID: "3", Type: mail.MailTypeToolResult, Content: "inner tool output", Metadata: mail.MailMetadata{Taints: []string{"INNER_ONLY"}}},
		{ID: "4", Type: mail.MailTypeAssistant, Content: "assistant response", Metadata: mail.MailMetadata{Taints: []string{}}},
	}

	result, err := SanitizeMessageHistory(messages, DMZBoundary)
	if err != nil {
		t.Fatalf("SanitizeMessageHistory returned error: %v", err)
	}

	if len(result) != len(messages) {
		t.Errorf("SanitizeMessageHistory returned %d messages, want %d", len(result), len(messages))
	}

	if result[0].Content != "user message" {
		t.Errorf("Message 0 content = %q, want %q", result[0].Content, "user message")
	}

	if result[1].Content != "[REDACTED: SECRET]" {
		t.Errorf("Message 1 content = %q, want %q", result[1].Content, "[REDACTED: SECRET]")
	}

	if result[2].Content != "[REDACTED: INNER_ONLY]" {
		t.Errorf("Message 2 content = %q, want %q", result[2].Content, "[REDACTED: INNER_ONLY]")
	}

	if result[3].Content != "assistant response" {
		t.Errorf("Message 3 content = %q, want %q", result[3].Content, "assistant response")
	}
}
