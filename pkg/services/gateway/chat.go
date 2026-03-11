package gateway

import (
	"fmt"
	"time"

	"github.com/maelstrom/v3/pkg/mail"
	"github.com/maelstrom/v3/pkg/security"
)

// ChatSession represents a chat session for human-in-the-loop interaction
type ChatSession struct {
	AgentID     string
	SessionID   string
	SessionType string
	ContextMap  *security.ContextMapSnapshot
	Messages    []ChatMessage
}

// ChatMessage represents a message in the chat session
type ChatMessage struct {
	ID        string
	Content   string
	Taints    []string
	Boundary  string
	Type      string
	Source    string
	IsPartial bool
}

// ActionItem represents a parsed action item from chat input
type ActionItem struct {
	Type    string
	Payload any
}

// CreateChatSession creates a new chat session for an agent
func (g *gatewayService) CreateChatSession(agentID string) (*ChatSession, error) {
	sessionID := fmt.Sprintf("session-%s-%d", agentID, time.Now().UnixNano())
	contextMap := &security.ContextMap{
		Blocks:     []*security.ContextBlock{},
		TokenCount: 0,
		Budget:     4096,
	}
	snapshot := contextMap.Snapshot()

	return &ChatSession{
		AgentID:     agentID,
		SessionID:   sessionID,
		SessionType: "chat",
		ContextMap:  snapshot,
		Messages:    []ChatMessage{},
	}, nil
}

// GetChatPath returns the HTTPS endpoint path for a chat session
func (g *gatewayService) GetChatPath(agentID string) string {
	return "/chat/" + agentID
}

// GetLastNMessages returns the last N messages sanitized by boundary rules
func (s *ChatSession) GetLastNMessages(n int) []ChatMessage {
	panic("not implemented")
}

// SendHumanMessage sends a human message as mail_received
func (g *gatewayService) SendHumanMessage(agentID string, message string) (*mail.Mail, error) {
	panic("not implemented")
}

// RenderAgentReply renders an agent's mail reply as a chat message
func (g *gatewayService) RenderAgentReply(mail *mail.Mail) ChatMessage {
	panic("not implemented")
}

// ParseActionItem parses action item shorthands from chat input
func (g *gatewayService) ParseActionItem(message string) (*ActionItem, error) {
	panic("not implemented")
}
