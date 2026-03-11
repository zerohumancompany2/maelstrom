package gateway

import (
	"testing"

	"github.com/maelstrom/v3/pkg/mail"
)

func TestStreamingPath_UserInputToMail(t *testing.T) {
	svc := NewGatewayService()
	sessionID := "test-session-001"
	userInput := "Hello, agent!"

	// User input creates mail_received Mail (arch-v1.md L675)
	m, err := svc.HandleUserInput(userInput, sessionID)
	if err != nil {
		t.Fatalf("Expected no error handling user input, got %v", err)
	}

	// Verify mail type (arch-v1.md L675)
	if m.Type != "mail_received" {
		t.Errorf("Expected mail type 'mail_received', got '%s'", m.Type)
	}

	// Verify taints=["USER_SUPPLIED"] (arch-v1.md L675)
	if len(m.Metadata.Taints) != 1 {
		t.Errorf("Expected 1 taint, got %d", len(m.Metadata.Taints))
	}
	if m.Metadata.Taints[0] != "USER_SUPPLIED" {
		t.Errorf("Expected taint 'USER_SUPPLIED', got '%s'", m.Metadata.Taints[0])
	}

	// Verify boundary=outer (arch-v1.md L675)
	if m.Metadata.Boundary != "outer" {
		t.Errorf("Expected boundary 'outer', got '%s'", m.Metadata.Boundary)
	}

	// Verify delivered to DMZ Agent inbox via Communication.publish
	if m.Target != "agent:dmz" {
		t.Errorf("Expected target 'agent:dmz', got '%s'", m.Target)
	}

	// Verify content matches user input
	if m.Content != userInput {
		t.Errorf("Expected content '%s', got '%s'", userInput, m.Content)
	}
}

func TestStreamingPath_MailToLLMStream(t *testing.T) {
	svc := NewGatewayService()
	agentID := "test-agent-001"

	// Setup: Create mail_received in agent inbox
	m := mail.Mail{
		ID:      "mail-001",
		Type:    "mail_received",
		Source:  "user",
		Target:  "agent:" + agentID,
		Content: "Process this request",
		Metadata: mail.MailMetadata{
			Taints:   []string{"USER_SUPPLIED"},
			Boundary: "outer",
		},
	}

	// Engine.dispatchEvent enters LLMReason state (arch-v1.md L677)
	state, err := svc.DispatchEvent(agentID, m)
	if err != nil {
		t.Fatalf("Expected no error dispatching event, got %v", err)
	}

	// Verify LLMReason state entered (arch-v1.md L677)
	if state.Name != "LLMReason" {
		t.Errorf("Expected state 'LLMReason', got '%s'", state.Name)
	}

	// AgentExtensions.assembleContextMap pulls session + Memory blocks (arch-v1.md L678)
	contextMap, err := svc.AssembleContextMap(agentID)
	if err != nil {
		t.Fatalf("Expected no error assembling context map, got %v", err)
	}

	// Verify contextMap includes session data (arch-v1.md L678)
	if contextMap.Session == nil {
		t.Error("Expected contextMap to include session")
	}

	// Verify taints propagated through context assembly (arch-v1.md L678)
	if len(contextMap.Taints) == 0 {
		t.Error("Expected contextMap to have propagated taints")
	}

	// Verify streaming enabled for LLM call
	if !contextMap.StreamingEnabled {
		t.Error("Expected streaming enabled for LLM call")
	}
}
