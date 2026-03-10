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

func TestE2E_ViolationReporting_MultiTypeAggregation(t *testing.T) {
	runtime := NewE2ERuntime()
	if err := runtime.Start(); err != nil {
		t.Fatalf("Failed to start runtime: %v", err)
	}
	defer runtime.Stop()

	agentA := runtime.CreateAgent("agent-a", mail.DMZBoundary, security.TaintPolicy{})
	if agentA == nil {
		t.Fatal("Failed to create agent A")
	}

	agentB := runtime.CreateAgent("agent-b", mail.DMZBoundary, security.TaintPolicy{})
	if agentB == nil {
		t.Fatal("Failed to create agent B")
	}

	agentC := runtime.CreateAgent("agent-c", mail.DMZBoundary, security.TaintPolicy{})
	if agentC == nil {
		t.Fatal("Failed to create agent C")
	}

	outerAgent := runtime.CreateAgent("outer-agent", mail.OuterBoundary, security.TaintPolicy{})
	if outerAgent == nil {
		t.Fatal("Failed to create outer agent")
	}

	runtime.TriggerViolation("agent-a", "INNER_ONLY", []string{"INNER_ONLY"})

	runtime.TriggerViolation("agent-b", "BOUNDARY", []string{"BOUNDARY"})

	runtime.TriggerViolation("agent-c", "ISOLATION", []string{"ISOLATION"})

	deadLetterQueue := runtime.GetDeadLetterQueue()
	violationCount := 0
	for _, v := range deadLetterQueue {
		if v.Type == mail.MailTypeTaintViolation {
			violationCount++
		}
	}
	if violationCount != 3 {
		t.Errorf("Expected 3 violations in dead-letter queue, got %d", violationCount)
	}

	violations := runtime.GetViolations()
	if len(violations) != 3 {
		t.Errorf("Expected 3 violations stored in sys:observability, got %d", len(violations))
	}

	metrics := runtime.GetMetrics()
	totalViolations, ok := metrics["taint_violations_total"].(int)
	if !ok || totalViolations != 3 {
		t.Errorf("Expected taint_violations_total to be 3, got %d", totalViolations)
	}

	byType, ok := metrics["taint_violations_by_type"].(map[string]int)
	if !ok {
		t.Error("Expected taint_violations_by_type metric")
	}
	innerOnlyCount, ok := byType["INNER_ONLY"]
	if !ok || innerOnlyCount != 1 {
		t.Errorf("Expected taint_violations_by_type{type=\"INNER_ONLY\"} = 1, got %d", innerOnlyCount)
	}
	boundaryCount, ok := byType["BOUNDARY"]
	if !ok || boundaryCount != 1 {
		t.Errorf("Expected taint_violations_by_type{type=\"BOUNDARY\"} = 1, got %d", boundaryCount)
	}
	isolationCount, ok := byType["ISOLATION"]
	if !ok || isolationCount != 1 {
		t.Errorf("Expected taint_violations_by_type{type=\"ISOLATION\"} = 1, got %d", isolationCount)
	}

	queryResults := runtime.QueryViolations(map[string]interface{}{})
	if len(queryResults) != 3 {
		t.Errorf("Expected 3 violations from query, got %d", len(queryResults))
	}

	foundTypes := make(map[string]bool)
	for _, v := range queryResults {
		foundTypes[v.Type] = true
	}
	if !foundTypes["INNER_ONLY"] {
		t.Error("Expected INNER_ONLY violation in query results")
	}
	if !foundTypes["BOUNDARY"] {
		t.Error("Expected BOUNDARY violation in query results")
	}
	if !foundTypes["ISOLATION"] {
		t.Error("Expected ISOLATION violation in query results")
	}

	runtime.TriggerViolation("agent-a", "INNER_ONLY", []string{"INNER_ONLY"})

	afterRetryMetrics := runtime.GetMetrics()
	afterRetryTotal, ok := afterRetryMetrics["taint_violations_total"].(int)
	if !ok || afterRetryTotal != 4 {
		t.Errorf("Expected taint_violations_total to be 4 after retry, got %d", afterRetryTotal)
	}
}
