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

func TestHumanGateway_SendMessage(t *testing.T) {
	// Test: Send message to agent
	svc := NewHumanGatewayService()

	// Open a session
	sessionID, err := svc.OpenSession("agent-002")
	if err != nil {
		t.Fatalf("Failed to open session: %v", err)
	}

	// Send a message
	content := "Hello, agent!"
	err = svc.SendMessage(string(sessionID), content)
	if err != nil {
		t.Fatalf("Failed to send message: %v", err)
	}

	// Verify session exists
	if !svc.SessionExists(sessionID) {
		t.Error("Session should still exist after sending message")
	}
}

func TestHumanGateway_StreamResponse(t *testing.T) {
	// Test: Stream agent response
	svc := NewHumanGatewayService()

	// Open a session
	sessionID, err := svc.OpenSession("agent-003")
	if err != nil {
		t.Fatalf("Failed to open session: %v", err)
	}

	// Stream response
	ch, err := svc.StreamResponse(string(sessionID))
	if err != nil {
		t.Fatalf("Failed to stream response: %v", err)
	}

	if ch == nil {
		t.Fatal("Stream channel should not be nil")
	}

	// Verify session still exists
	if !svc.SessionExists(sessionID) {
		t.Error("Session should exist during streaming")
	}
}

func TestHumanGateway_BoundaryOuter(t *testing.T) {
	// Test: Human gateway only accessible from outer
	svc := NewHumanGatewayService()

	// Open session (outer boundary operation)
	sessionID, err := svc.OpenSession("agent-004")
	if err != nil {
		t.Fatalf("Failed to open session: %v", err)
	}

	// Send message (outer boundary operation)
	err = svc.SendMessage(string(sessionID), "Hello from outer")
	if err != nil {
		t.Fatalf("Failed to send message: %v", err)
	}

	// Stream response (outer boundary operation)
	ch, err := svc.StreamResponse(string(sessionID))
	if err != nil {
		t.Fatalf("Failed to stream: %v", err)
	}

	// Verify we can receive from stream
	<-ch

	// Close session (outer boundary operation)
	err = svc.CloseSession(string(sessionID))
	if err != nil {
		t.Fatalf("Failed to close session: %v", err)
	}

	// Session should be closed (not active)
	if svc.SessionActive(sessionID) {
		t.Error("Session should be closed after CloseSession")
	}
}
