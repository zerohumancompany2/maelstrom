package e2e

import (
	"testing"

	"github.com/maelstrom/v3/pkg/mail"
	security "github.com/maelstrom/v3/pkg/security"
)

func TestE2E_ViolationReporting_TaintViolationFlow(t *testing.T) {
	runtime := NewE2ERuntime()
	if err := runtime.Start(); err != nil {
		t.Fatalf("Failed to start runtime: %v", err)
	}
	defer runtime.Stop()

	dmzAgent := runtime.CreateAgent("dmz-agent", mail.DMZBoundary, security.TaintPolicy{})
	if dmzAgent == nil {
		t.Fatal("Failed to create DMZ agent")
	}

	outerAgent := runtime.CreateAgent("outer-agent", mail.OuterBoundary, security.TaintPolicy{})
	if outerAgent == nil {
		t.Fatal("Failed to create outer agent")
	}

	_, err := runtime.SendMail("dmz-agent", "outer-agent", "INNER_ONLY tainted data", []string{"INNER_ONLY"})
	if err == nil {
		t.Error("Expected error when sending mail with INNER_ONLY taint from DMZ to outer boundary")
	}

	deadLetterQueue := runtime.GetDeadLetterQueue()
	violationFound := false
	var violationMail *mail.Mail
	for _, v := range deadLetterQueue {
		if v.Type == mail.MailTypeTaintViolation {
			violationFound = true
			violationMail = v
			break
		}
	}
	if !violationFound {
		t.Error("Expected taint_violation event in dead-letter queue")
	}

	if violationMail != nil {
		if violationMail.Target != "sys:observability" {
			t.Error("Expected violation to be sent to sys:observability")
		}
	}

	violations := runtime.GetViolations()
	if len(violations) == 0 {
		t.Error("Expected violations to be stored in sys:observability")
	}

	metrics := runtime.GetMetrics()
	totalViolations, ok := metrics["taint_violations_total"].(int)
	if !ok || totalViolations == 0 {
		t.Error("Expected taint_violations_total metric to be incremented")
	}

	byType, ok := metrics["taint_violations_by_type"].(map[string]int)
	if !ok {
		t.Error("Expected taint_violations_by_type metric")
	}
	innerOnlyCount, ok := byType["INNER_ONLY"]
	if !ok || innerOnlyCount == 0 {
		t.Error("Expected taint_violations_by_type{type=\"INNER_ONLY\"} metric to be incremented")
	}

	queryResults := runtime.QueryViolations(map[string]interface{}{})
	if len(queryResults) == 0 {
		t.Error("Expected violations to be queryable via sys:observability.query")
	}

	if len(queryResults) > 0 {
		v := queryResults[0]
		if v.Source != "dmz-agent" {
			t.Errorf("Expected violation source to be 'dmz-agent', got '%s'", v.Source)
		}
		if v.Target != "sys:observability" {
			t.Errorf("Expected violation target to be 'sys:observability', got '%s'", v.Target)
		}
		if len(v.ForbiddenTaints) == 0 {
			t.Error("Expected violation to have forbidden taints")
		}
		if v.Timestamp.IsZero() {
			t.Error("Expected violation to have timestamp")
		}
	}
}
