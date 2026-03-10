package e2e

import (
	"strings"
	"testing"

	"github.com/maelstrom/v3/pkg/mail"
	security "github.com/maelstrom/v3/pkg/security"
)

func TestE2E_IsolationHook_DirectSyscallBlocked(t *testing.T) {
	runtime := NewE2ERuntime()
	if err := runtime.Start(); err != nil {
		t.Fatalf("Failed to start runtime: %v", err)
	}
	defer runtime.Stop()

	agent := runtime.CreateAgent("isolated-agent", mail.InnerBoundary, security.TaintPolicy{})
	if agent == nil {
		t.Fatal("Failed to create isolated agent")
	}

	blocked, err := runtime.AttemptDirectSyscall(agent.ID, "open", "/etc/passwd")

	if !blocked {
		t.Error("Expected direct syscall to be blocked")
	}

	if err == nil {
		t.Error("Expected error when blocking direct syscall")
	} else if !strings.Contains(err.Error(), "direct syscall not allowed") {
		t.Errorf("Expected error containing 'direct syscall not allowed - use mediated tool layer', got: %v", err)
	}

	deadLetterQueue := runtime.GetDeadLetterQueue()
	violationFound := false
	for _, v := range deadLetterQueue {
		if v.Type == mail.MailTypeTaintViolation {
			violationFound = true
			if v.Source != agent.ID {
				t.Errorf("Expected violation source to be agent ID, got: %s", v.Source)
			}
			if v.Target != "sys:observability" {
				t.Errorf("Expected violation target to be sys:observability, got: %s", v.Target)
			}
			break
		}
	}
	if !violationFound {
		t.Error("Expected taint_violation event in dead-letter queue")
	}
}
