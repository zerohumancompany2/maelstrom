package gateway

import "testing"

func TestHumanChat_HTTPSEndpointAgentId(t *testing.T) {
	svc := NewGatewayService()
	agentID := "agent:dmz"

	// Create chat session at /chat/{agentId} (arch-v1.md L728)
	session, err := svc.CreateChatSession(agentID)
	if err != nil {
		t.Fatalf("Expected no error creating chat session, got %v", err)
	}

	// Verify session created (arch-v1.md L728)
	if session == nil {
		t.Error("Expected session to be created")
	}

	// Verify agent ID matches (arch-v1.md L728)
	if session.AgentID != agentID {
		t.Errorf("Expected agentID '%s', got '%s'", agentID, session.AgentID)
	}

	// Verify session ID generated
	if session.SessionID == "" {
		t.Error("Expected session ID to be generated")
	}

	// Verify HTTPS endpoint path format (arch-v1.md L728)
	expectedPath := "/chat/" + agentID
	actualPath := svc.GetChatPath(agentID)
	if actualPath != expectedPath {
		t.Errorf("Expected path '%s', got '%s'", expectedPath, actualPath)
	}

	// Verify session is for long-running top-level Agent (arch-v1.md L728)
	if session.SessionType != "chat" {
		t.Errorf("Expected session type 'chat', got '%s'", session.SessionType)
	}
}
