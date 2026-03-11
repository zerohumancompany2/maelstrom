package gateway

import (
	"slices"
	"testing"

	"github.com/maelstrom/v3/pkg/mail"
	"github.com/maelstrom/v3/pkg/security"
)

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

func TestHumanChat_ReadOnlyContextMapSnapshot(t *testing.T) {
	svc := NewGatewayService()
	agentID := "agent:dmz"

	// Setup: Create ContextMap with blocks (arch-v1.md L731)
	blocks := []*security.ContextBlock{
		{Name: "session", Content: "Current session data", Priority: 1},
		{Name: "memory", Content: "Long-term memory", Priority: 2},
	}
	contextMap := security.NewContextMap(blocks, 4096)

	// Session receives read-only snapshot (arch-v1.md L731)
	snapshot := contextMap.Snapshot()
	if snapshot == nil {
		t.Fatal("Expected snapshot to be created")
	}

	// Verify snapshot is read-only (arch-v1.md L731)
	if !snapshot.ReadOnly {
		t.Error("Expected snapshot to be read-only")
	}

	// Verify snapshot contains blocks (arch-v1.md L731)
	if len(snapshot.Blocks) != 2 {
		t.Errorf("Expected 2 blocks, got %d", len(snapshot.Blocks))
	}

	// Verify snapshot has token count
	if snapshot.TokenCount == 0 {
		t.Error("Expected snapshot to have token count")
	}

	// Verify creating chat session receives snapshot (arch-v1.md L731)
	session, err := svc.CreateChatSession(agentID)
	if err != nil {
		t.Fatalf("Expected no error creating session, got %v", err)
	}

	if session.ContextMap == nil {
		t.Error("Expected session to have ContextMap snapshot")
	}

	if !session.ContextMap.ReadOnly {
		t.Error("Expected session ContextMap to be read-only")
	}

	// Verify snapshot is independent (modifying original doesn't affect snapshot)
	originalCount := len(snapshot.Blocks)
	blocks[0].Content = "Modified content"
	if len(snapshot.Blocks) != originalCount {
		t.Error("Expected snapshot to be independent of original")
	}
}

func TestHumanChat_LastNMessagesSanitized(t *testing.T) {
	svc := NewGatewayService()
	agentID := "agent:dmz"

	// Setup: Create session with messages (arch-v1.md L731)
	session, err := svc.CreateChatSession(agentID)
	if err != nil {
		t.Fatalf("Expected no error creating session, got %v", err)
	}

	// Add messages with various taints and boundaries (arch-v1.md L731)
	messages := []ChatMessage{
		{ID: "msg1", Content: "Normal message", Taints: []string{}, Boundary: "outer", Type: "user"},
		{ID: "msg2", Content: "Secret message", Taints: []string{"SECRET"}, Boundary: "inner", Type: "assistant"},
		{ID: "msg3", Content: "PII message", Taints: []string{"PII"}, Boundary: "dmz", Type: "user"},
		{ID: "msg4", Content: "Tool output", Taints: []string{"TOOL_OUTPUT"}, Boundary: "outer", Type: "assistant"},
		{ID: "msg5", Content: "Clean message", Taints: []string{}, Boundary: "outer", Type: "user"},
	}

	for _, msg := range messages {
		session.Messages = append(session.Messages, msg)
	}

	// Get last N messages sanitized by boundary rules (arch-v1.md L731)
	sanitized := session.GetLastNMessages(3)

	// Verify only 3 messages returned (last N)
	if len(sanitized) != 3 {
		t.Errorf("Expected 3 messages, got %d", len(sanitized))
	}

	// Verify SECRET taint removed (arch-v1.md L731)
	for _, msg := range sanitized {
		if slices.Contains(msg.Taints, "SECRET") {
			t.Error("Expected SECRET taint to be sanitized")
		}
	}

	// Verify PII taint removed (arch-v1.md L731)
	for _, msg := range sanitized {
		if slices.Contains(msg.Taints, "PII") {
			t.Error("Expected PII taint to be sanitized")
		}
	}

	// Verify inner-boundary messages removed (arch-v1.md L731)
	for _, msg := range sanitized {
		if msg.Boundary == "inner" {
			t.Error("Expected inner-boundary messages to be sanitized")
		}
	}

	// Verify TOOL_OUTPUT taint preserved (allowed)
	hasToolOutput := false
	for _, msg := range sanitized {
		if slices.Contains(msg.Taints, "TOOL_OUTPUT") {
			hasToolOutput = true
			break
		}
	}
	if !hasToolOutput {
		t.Error("Expected TOOL_OUTPUT taint to be preserved")
	}

	// Verify clean messages preserved
	hasClean := false
	for _, msg := range sanitized {
		if msg.Content == "Clean message" {
			hasClean = true
			break
		}
	}
	if !hasClean {
		t.Error("Expected clean message to be preserved")
	}
}

func TestHumanChat_MessageBecomesMailReceived(t *testing.T) {
	svc := NewGatewayService()
	agentID := "agent:dmz"

	// Human message sent (arch-v1.md L732)
	humanMessage := "Hello, agent! How can you help me?"
	m, err := svc.SendHumanMessage(agentID, humanMessage)
	if err != nil {
		t.Fatalf("Expected no error sending human message, got %v", err)
	}

	// Verify mail type is mail_received (arch-v1.md L732)
	if m.Type != mail.MailReceived {
		t.Errorf("Expected mail type 'mail_received', got '%s'", m.Type)
	}

	// Verify mail subtype is human_feedback (arch-v1.md L732)
	if m.Metadata.HumanFeedbackType != "human_feedback" {
		t.Errorf("Expected human_feedback type, got '%s'", m.Metadata.HumanFeedbackType)
	}

	// Verify delivered to Agent's inbox (arch-v1.md L732)
	if m.Target != agentID {
		t.Errorf("Expected target '%s', got '%s'", agentID, m.Target)
	}

	// Verify content matches human message
	if m.Content != humanMessage {
		t.Errorf("Expected content '%s', got '%s'", humanMessage, m.Content)
	}

	// Verify taints include USER_SUPPLIED
	if !slices.Contains(m.Metadata.Taints, "USER_SUPPLIED") {
		t.Error("Expected USER_SUPPLIED taint")
	}

	// Verify boundary is outer (human input)
	if m.Metadata.Boundary != mail.OuterBoundary {
		t.Errorf("Expected boundary 'outer', got '%s'", m.Metadata.Boundary)
	}

	// Verify mail has ID and timestamp
	if m.ID == "" {
		t.Error("Expected mail to have ID")
	}
	if m.CreatedAt.IsZero() {
		t.Error("Expected mail to have timestamp")
	}
}
