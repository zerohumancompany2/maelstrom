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
