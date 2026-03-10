package e2e

import (
	"testing"

	"github.com/maelstrom/v3/pkg/mail"
	security "github.com/maelstrom/v3/pkg/security"
)

func TestE2E_DataTainting_MessageCreationToExit(t *testing.T) {
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

	userMessage, err := runtime.SendUserMessage("dmz-agent", "Hello, agent")
	if err != nil {
		t.Fatalf("Failed to send user message: %v", err)
	}
	if userMessage == nil {
		t.Fatal("User message is nil")
	}

	hasUserSupplied := false
	for _, taint := range userMessage.Metadata.Taints {
		if taint == "USER_SUPPLIED" {
			hasUserSupplied = true
			break
		}
	}
	if !hasUserSupplied {
		t.Error("Expected USER_SUPPLIED taint on original message")
	}

	_, err = runtime.SendMail("dmz-agent", "outer-agent", "Response content", []string{"INNER_ONLY"})
	if err == nil {
		t.Error("Expected error when sending mail with INNER_ONLY taint to outer boundary")
	}

	deadLetterQueue := runtime.GetDeadLetterQueue()
	violationFound := false
	for _, v := range deadLetterQueue {
		if v.Type == mail.MailTypeTaintViolation {
			violationFound = true
			break
		}
	}
	if !violationFound {
		t.Error("Expected violation logged to dead-letter queue for INNER_ONLY taint")
	}
}

func TestE2E_DataTainting_AllowedExitPropagation(t *testing.T) {
	runtime := NewE2ERuntime()
	if err := runtime.Start(); err != nil {
		t.Fatalf("Failed to start runtime: %v", err)
	}
	defer runtime.Stop()

	taintPolicy := security.TaintPolicy{
		AllowedForBoundary: []security.BoundaryType{security.DMZBoundary, security.OuterBoundary},
	}
	dmzAgent := runtime.CreateAgent("dmz-agent", mail.DMZBoundary, taintPolicy)
	if dmzAgent == nil {
		t.Fatal("Failed to create DMZ agent")
	}

	outerAgent := runtime.CreateAgent("outer-agent", mail.OuterBoundary, security.TaintPolicy{})
	if outerAgent == nil {
		t.Fatal("Failed to create outer agent")
	}

	toolResult := map[string]interface{}{
		"_taints": []string{"TOOL_OUTPUT"},
		"data":    "search results",
	}

	mailWithToolOutput, err := runtime.SendMail("dmz-agent", "outer-agent", toolResult, []string{"TOOL_OUTPUT"})
	if err != nil {
		t.Fatalf("Expected TOOL_OUTPUT taint to be allowed on exit, got error: %v", err)
	}
	if mailWithToolOutput == nil {
		t.Fatal("Mail with TOOL_OUTPUT taint should be delivered")
	}

	hasToolOutput := false
	for _, taint := range mailWithToolOutput.Metadata.Taints {
		if taint == "TOOL_OUTPUT" {
			hasToolOutput = true
			break
		}
	}
	if !hasToolOutput {
		t.Error("Expected TOOL_OUTPUT taint to be preserved in mail")
	}

	violations := runtime.GetViolations()
	if len(violations) > 0 {
		t.Errorf("Expected no violations for allowed taint TOOL_OUTPUT, got %d violations", len(violations))
	}
}
