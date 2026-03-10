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

// TestHumanGatewayService_GetSession - arch-v1.md L728
// Get chat session by ID
func TestHumanGatewayService_GetSession(t *testing.T) {
	svc := NewHumanGatewayService()

	// Create a session first
	session, err := svc.CreateChatSession("agent-789")
	if err != nil {
		t.Errorf("Expected nil error, got %v", err)
	}

	// Retrieve session by ID
	retrieved := svc.GetSession("agent-789")
	if retrieved == nil {
		t.Error("Expected non-nil session")
	}

	if retrieved.AgentID != "agent-789" {
		t.Errorf("Expected AgentID 'agent-789', got '%s'", retrieved.AgentID)
	}

	// Session data preserved between calls
	if retrieved != session {
		t.Error("Expected same session instance")
	}

	// Error for non-existent session (returns nil)
	missing := svc.GetSession("non-existent")
	if missing != nil {
		t.Error("Expected nil for non-existent session")
	}
}

// TestHumanGatewayService_HandleChat - arch-v1.md L732
// Handle chat messages
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

// TestHumanGatewayService_SendMessage - arch-v1.md L732
// Send messages to agent inbox with human_feedback type
func TestHumanGatewayService_SendMessage(t *testing.T) {
	svc := NewHumanGatewayService()

	session, err := svc.CreateChatSession("agent-123")
	if err != nil {
		t.Errorf("Expected nil error, got %v", err)
	}

	msg := "Hello agent"
	err = svc.SendMessage(session, msg)
	if err != nil {
		t.Errorf("Expected nil error, got %v", err)
	}

	if len(session.Messages) != 1 {
		t.Errorf("Expected 1 message, got %d", len(session.Messages))
	}

	m := session.Messages[0]
	if m.Type != mail.MailTypeHumanFeedback {
		t.Errorf("Expected MailTypeHumanFeedback, got %s", m.Type)
	}

	if m.Target != "agent:agent-123" {
		t.Errorf("Expected target 'agent:agent-123', got '%s'", m.Target)
	}

	content, ok := m.Content.(map[string]any)
	if !ok {
		t.Error("Expected content to be map[string]any")
	}
	if content["message"] != msg {
		t.Errorf("Expected message '%s', got '%v'", msg, content["message"])
	}

	err = svc.SendMessage(nil, "test")
	if err == nil {
		t.Error("Expected error for nil session")
	}
}

// TestHumanGatewayService_ContextMapSanitization - arch-v1.md L731
// Sanitize ContextMap by boundary rules
func TestHumanGatewayService_ContextMapSanitization(t *testing.T) {
	ctx := ContextMapSnapshot{
		"conversation": []any{"msg1", "msg2"},
		"memory":       "important data",
		"internal":     "secret data",
	}

	ctx["taints"] = map[string]any{
		"FORBIDDEN": "should be redacted",
		"USER":      "allowed",
	}

	sanitizedInner := SanitizeContextForBoundary(ctx, mail.InnerBoundary)
	if len(sanitizedInner) != len(ctx) {
		t.Errorf("Expected %d keys for inner boundary, got %d", len(ctx), len(sanitizedInner))
	}

	sanitizedOuter := SanitizeContextForBoundary(ctx, mail.OuterBoundary)
	if len(sanitizedOuter) != 3 {
		t.Errorf("Expected 3 keys for outer boundary, got %d", len(sanitizedOuter))
	}

	if sanitizedOuter["taints"] != nil {
		t.Error("Expected taints to be redacted for outer boundary")
	}
}

// TestHumanGatewayService_ParseActionItem - arch-v1.md L734
// Parse action item shorthands
func TestHumanGatewayService_ParseActionItem(t *testing.T) {
	svc := NewHumanGatewayService()

	items, err := svc.ParseActionItem("@pause")
	if err != nil {
		t.Errorf("Expected nil error, got %v", err)
	}

	if len(items) != 1 {
		t.Errorf("Expected 1 action item, got %d", len(items))
	}

	if items[0].Type != "pause" {
		t.Errorf("Expected type 'pause', got '%s'", items[0].Type)
	}

	items, err = svc.ParseActionItem("@inject-memory important data")
	if err != nil {
		t.Errorf("Expected nil error, got %v", err)
	}

	if len(items) != 1 {
		t.Errorf("Expected 1 action item, got %d", len(items))
	}

	if items[0].Type != "inject-memory" {
		t.Errorf("Expected type 'inject-memory', got '%s'", items[0].Type)
	}

	payload, ok := items[0].Payload.(string)
	if !ok {
		t.Error("Expected payload to be string")
	}

	if payload != "important data" {
		t.Errorf("Expected payload 'important data', got '%s'", payload)
	}

	items, err = svc.ParseActionItem("Just a normal message")
	if err != nil {
		t.Errorf("Expected nil error, got %v", err)
	}

	if len(items) != 0 {
		t.Errorf("Expected 0 action items, got %d", len(items))
	}
}
