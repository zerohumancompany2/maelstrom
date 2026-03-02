package humangateway

import (
	"testing"
)

func TestHumanGateway_OpenSession(t *testing.T) {
	// Test: Open chat session for agent
	svc := NewHumanGatewayService()

	// Open a session for an agent
	sessionID, err := svc.OpenSession("agent-001")
	if err != nil {
		t.Fatalf("Failed to open session: %v", err)
	}

	if sessionID == "" {
		t.Fatal("Session ID is empty")
	}

	// Verify session was created
	exists := svc.SessionExists(sessionID)
	if !exists {
		t.Error("Session should exist after opening")
	}
}
