package e2e

import (
	"testing"

	"github.com/maelstrom/v3/pkg/mail"
	security "github.com/maelstrom/v3/pkg/security"
)

func TestE2E_BoundaryEnforcement_ForbiddenTransitionBlocked(t *testing.T) {
	runtime := NewE2ERuntime()
	if err := runtime.Start(); err != nil {
		t.Fatalf("Failed to start runtime: %v", err)
	}
	defer runtime.Stop()

	innerAgent := runtime.CreateAgent("inner-agent", mail.InnerBoundary, security.TaintPolicy{})
	if innerAgent == nil {
		t.Fatal("Failed to create inner agent")
	}

	outerAgent := runtime.CreateAgent("outer-agent", mail.OuterBoundary, security.TaintPolicy{})
	if outerAgent == nil {
		t.Fatal("Failed to create outer agent")
	}

	_, err := runtime.SendMail("inner-agent", "outer-agent", "Sensitive inner data", []string{"INNER_ONLY"})
	if err == nil {
		t.Error("Expected error when sending mail with INNER_ONLY taint from inner to outer boundary")
	}

	deadLetterQueue := runtime.GetDeadLetterQueue()
	violationFound := false
	for _, v := range deadLetterQueue {
		if v.Type == mail.MailTypeTaintViolation {
			violationFound = true
			if v.Target != "sys:observability" {
				t.Error("Expected violation to be sent to sys:observability")
			}
			break
		}
	}
	if !violationFound {
		t.Error("Expected taint_violation event in dead-letter queue")
	}

	outerInbox := outerAgent.Inbox
	if outerInbox != nil && len(outerInbox.Messages) > 0 {
		t.Error("Expected outer agent to not receive the blocked message")
	}
}

func TestE2E_BoundaryEnforcement_ContextMapFiltering(t *testing.T) {
	runtime := NewE2ERuntime()
	if err := runtime.Start(); err != nil {
		t.Fatalf("Failed to start runtime: %v", err)
	}
	defer runtime.Stop()

	taintPolicy := security.TaintPolicy{
		RedactMode: "redact",
		RedactRules: []security.RedactRule{
			{Taint: "INNER_ONLY", Replacement: "[REDACTED]"},
			{Taint: "PII", Replacement: "[REDACTED]"},
		},
		AllowedForBoundary: []security.BoundaryType{security.DMZBoundary},
	}
	dmzAgent := runtime.CreateAgent("dmz-agent", mail.DMZBoundary, taintPolicy)
	if dmzAgent == nil {
		t.Fatal("Failed to create DMZ agent")
	}

	_, err := runtime.AssembleContextMap("dmz-agent")
	if err != nil {
		t.Fatalf("Failed to assemble ContextMap: %v", err)
	}

	taintedData := map[string]interface{}{
		"data":    "sensitive content",
		"_taints": []string{"INNER_ONLY", "PII"},
	}

	allowed, err := runtime.securityService.CheckTaintPolicy(taintedData, mail.DMZBoundary, taintPolicy)
	if err != nil {
		t.Fatalf("CheckTaintPolicy failed: %v", err)
	}

	if allowed {
		t.Error("Expected tainted data to be filtered for DMZ boundary")
	}
}
