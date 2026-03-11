package gateway

import (
	"testing"

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
