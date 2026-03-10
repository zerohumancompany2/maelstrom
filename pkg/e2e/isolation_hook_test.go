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

func TestE2E_IsolationHook_ToolLayerMediation(t *testing.T) {
	runtime := NewE2ERuntime()
	if err := runtime.Start(); err != nil {
		t.Fatalf("Failed to start runtime: %v", err)
	}
	defer runtime.Stop()

	researcher := runtime.CreateAgent("researcher", mail.InnerBoundary, security.TaintPolicy{})
	if researcher == nil {
		t.Fatal("Failed to create researcher agent")
	}

	analyst := runtime.CreateAgent("analyst", mail.InnerBoundary, security.TaintPolicy{})
	if analyst == nil {
		t.Fatal("Failed to create analyst agent")
	}

	err := runtime.SetIsolationPolicy(researcher.ID, "strict")
	if err != nil {
		t.Fatalf("Failed to set isolation policy: %v", err)
	}

	err = runtime.SetIsolationPolicy(analyst.ID, "strict")
	if err != nil {
		t.Fatalf("Failed to set isolation policy: %v", err)
	}

	err = runtime.WriteFileViaDataSource(researcher.ID, "/agents/researcher/workspace/data.txt", []byte("researcher data"), []string{"USER_SUPPLIED"})
	if err != nil {
		t.Fatalf("Failed to write file: %v", err)
	}

	err = runtime.WriteFileViaDataSource(analyst.ID, "/agents/analyst/workspace/secret.txt", []byte("analyst secret"), []string{"SECRET"})
	if err != nil {
		t.Fatalf("Failed to write file: %v", err)
	}

	toolResult, err := runtime.CallTool(researcher.ID, "readFile", "/agents/researcher/workspace/data.txt")
	if err != nil {
		t.Fatalf("Expected legitimate read to succeed, got error: %v", err)
	}
	if toolResult == nil {
		t.Error("Expected tool result for legitimate read")
	}

	crossNamespaceResult, crossNamespaceErr := runtime.CallTool(researcher.ID, "readFile", "/agents/analyst/workspace/secret.txt")
	if crossNamespaceErr == nil {
		t.Error("Expected error when reading cross-namespace file")
	} else if !strings.Contains(crossNamespaceErr.Error(), "namespace isolation violation") {
		t.Errorf("Expected error containing 'namespace isolation violation', got: %v", crossNamespaceErr)
	}
	if crossNamespaceResult != nil {
		t.Error("Expected nil result for blocked cross-namespace read")
	}

	deadLetterQueue := runtime.GetDeadLetterQueue()
	crossNamespaceViolationFound := false
	for _, v := range deadLetterQueue {
		if v.Type == mail.MailTypeTaintViolation && v.Source == researcher.ID {
			if content, ok := v.Content.(map[string]interface{}); ok {
				if vtype, ok := content["type"].(string); ok && vtype == "namespace_isolation_violation" {
					crossNamespaceViolationFound = true
					if v.Target != "sys:observability" {
						t.Errorf("Expected violation target to be sys:observability, got: %s", v.Target)
					}
					break
				}
			}
		}
	}
	if !crossNamespaceViolationFound {
		t.Error("Expected namespace_isolation_violation in dead-letter queue")
	}
}
