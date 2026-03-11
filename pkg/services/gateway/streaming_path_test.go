package gateway

import (
	"testing"
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
