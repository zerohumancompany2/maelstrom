package humangateway

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/maelstrom/v3/pkg/mail"
)

// TestHumanGatewayService_ID - arch-v1.md L471
// HumanGatewayService.ID() returns "sys:human-gateway"
func TestHumanGatewayService_ID(t *testing.T) {
	svc := NewHumanGatewayService()

	id := svc.ID()
	if id != "sys:human-gateway" {
		t.Errorf("Expected ID 'sys:human-gateway', got '%s'", id)
	}

	// ID is immutable - multiple calls return same value
	id2 := svc.ID()
	if id != id2 {
		t.Error("Expected ID to be immutable across calls")
	}
}

// TestHumanGatewayService_CreateSession - arch-v1.md L728
// Create chat sessions for agents
func TestHumanGatewayService_CreateSession(t *testing.T) {
	svc := NewHumanGatewayService()

	session, err := svc.CreateChatSession("agent-123")
	if err != nil {
		t.Errorf("Expected nil error, got %v", err)
	}

	if session == nil {
		t.Error("Expected non-nil session")
	}

	if session.AgentID != "agent-123" {
		t.Errorf("Expected AgentID 'agent-123', got '%s'", session.AgentID)
	}

	// Multiple sessions can be created for different agents
	session2, err := svc.CreateChatSession("agent-456")
	if err != nil {
		t.Errorf("Expected nil error, got %v", err)
	}

	if session2.AgentID != "agent-456" {
		t.Errorf("Expected AgentID 'agent-456', got '%s'", session2.AgentID)
	}

	if session == session2 {
		t.Error("Expected different session instances for different agents")
	}
}

func TestHumanGatewayService_HandleChat(t *testing.T) {
	svc := NewHumanGatewayService()

	m, err := svc.HandleChat("recommendation-agent", "What do you think?")
	if err != nil {
		t.Errorf("Expected nil error, got %v", err)
	}

	if m.Type != mail.MailTypeUser {
		t.Errorf("Expected MailTypeUser, got %s", m.Type)
	}

	if m.Source != "human:recommendation-agent" {
		t.Errorf("Expected source 'human:recommendation-agent', got '%s'", m.Source)
	}

	if m.Target != "agent:recommendation-agent" {
		t.Errorf("Expected target 'agent:recommendation-agent', got '%s'", m.Target)
	}

	content, ok := m.Content.(map[string]any)
	if !ok {
		t.Error("Expected content to be map[string]any")
	}
	if content["message"] != "What do you think?" {
		t.Error("Expected message preserved in content")
	}
}

func TestHumanGatewayService_ParseActionItem(t *testing.T) {
	svc := NewHumanGatewayService()

	items, err := svc.ParseActionItem("Please @pause processing")
	if err != nil {
		t.Errorf("Expected nil error, got %v", err)
	}

	if len(items) != 1 {
		t.Errorf("Expected 1 action item, got %d", len(items))
	}
	if items[0].Type != "pause" {
		t.Errorf("Expected type 'pause', got '%s'", items[0].Type)
	}

	items, err = svc.ParseActionItem("@inject-memory This is important")
	if err != nil {
		t.Errorf("Expected nil error, got %v", err)
	}

	if len(items) != 1 {
		t.Errorf("Expected 1 action item, got %d", len(items))
	}
	if items[0].Type != "inject-memory" {
		t.Errorf("Expected type 'inject-memory', got '%s'", items[0].Type)
	}
	if items[0].Payload != "This is important" {
		t.Errorf("Expected payload 'This is important', got '%v'", items[0].Payload)
	}

	items, err = svc.ParseActionItem("Just a normal message")
	if err != nil {
		t.Errorf("Expected nil error, got %v", err)
	}
	if len(items) != 0 {
		t.Errorf("Expected 0 action items, got %d", len(items))
	}
}

func TestHumanGatewayService_SessionManagement(t *testing.T) {
	svc := NewHumanGatewayService()

	session := svc.CreateSession("test-agent")
	if session == nil {
		t.Error("Expected non-nil session")
	}

	if session.AgentID != "test-agent" {
		t.Errorf("Expected AgentID 'test-agent', got '%s'", session.AgentID)
	}

	retrieved := svc.GetSession("test-agent")
	if retrieved == nil {
		t.Error("Expected non-nil retrieved session")
	}
	if retrieved != session {
		t.Error("Expected same session instance")
	}

	missing := svc.GetSession("non-existent")
	if missing != nil {
		t.Error("Expected nil for non-existent session")
	}
}

func TestHumanGatewayService_SanitizedContext(t *testing.T) {
	ctx := ContextMapSnapshot{
		"conversation": []any{"msg1", "msg2"},
		"memory":       "important data",
		"internal":     "secret data",
	}

	sanitized := SanitizeContextForBoundary(ctx, mail.InnerBoundary)
	if len(sanitized) != len(ctx) {
		t.Errorf("Expected %d keys, got %d", len(ctx), len(sanitized))
	}

	sanitizedOuter := SanitizeContextForBoundary(ctx, mail.OuterBoundary)
	if len(sanitizedOuter) != len(ctx) {
		t.Errorf("Expected %d keys, got %d", len(ctx), len(sanitizedOuter))
	}
}

func TestHumanGatewayService_ChatEndpoint(t *testing.T) {
	svc := NewHumanGatewayService()

	req, _ := http.NewRequest(http.MethodGet, "/chat", nil)
	rr := httptest.NewRecorder()

	svc.ChatEndpoint(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, rr.Code)
	}
}

func TestHumanGatewayService_ChatSession(t *testing.T) {
	svc := NewHumanGatewayService()

	session, err := svc.CreateChatSession("test-agent")
	if err != nil {
		t.Errorf("Expected nil error, got %v", err)
	}

	if session == nil {
		t.Error("Expected non-nil session")
	}

	if session.AgentID != "test-agent" {
		t.Errorf("Expected AgentID 'test-agent', got '%s'", session.AgentID)
	}

	if session.ContextMap == nil {
		t.Error("Expected non-nil ContextMap")
	}
}
