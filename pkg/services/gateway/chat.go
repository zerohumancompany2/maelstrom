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
	forbiddenTaints := map[string]bool{
		"SECRET": true,
		"PII":    true,
	}

	var sanitized []ChatMessage
	count := 0

	for i := len(s.Messages) - 1; i >= 0; i-- {
		msg := s.Messages[i]

		if msg.Boundary == "inner" {
			continue
		}

		if count >= n {
			break
		}

		var cleanTaints []string
		for _, taint := range msg.Taints {
			if !forbiddenTaints[taint] {
				cleanTaints = append(cleanTaints, taint)
			}
		}

		cleanMsg := msg
		cleanMsg.Taints = cleanTaints
		sanitized = append(sanitized, cleanMsg)
		count++
	}

	for i := 0; i < len(sanitized)/2; i++ {
		sanitized[i], sanitized[len(sanitized)-1-i] = sanitized[len(sanitized)-1-i], sanitized[i]
	}

	return sanitized
}

// SendHumanMessage sends a human message as mail_received
func (g *gatewayService) SendHumanMessage(agentID string, message string) (*mail.Mail, error) {
	actionItem, _ := g.ParseActionItem(message)

	m := &mail.Mail{
		ID:        fmt.Sprintf("mail-%d", time.Now().UnixNano()),
		Type:      mail.MailReceived,
		Source:    "user",
		Target:    agentID,
		Content:   message,
		CreatedAt: time.Now(),
		Metadata: mail.MailMetadata{
			Boundary:          mail.OuterBoundary,
			Taints:            []string{"USER_SUPPLIED"},
			HumanFeedbackType: "human_feedback",
		},
	}

	if actionItem != nil {
		m.Metadata.ActionItem = mail.ActionItem{
			Type:    actionItem.Type,
			Payload: actionItem.Payload,
		}
	}

	return m, nil
}

// RenderAgentReply renders an agent's mail reply as a chat message
func (g *gatewayService) RenderAgentReply(m *mail.Mail) ChatMessage {
	content := m.Content
	str, ok := content.(string)
	if !ok {
		str = fmt.Sprintf("%v", content)
	}

	return ChatMessage{
		ID:        m.ID,
		Content:   str,
		Taints:    m.Metadata.Taints,
		Boundary:  string(m.Metadata.Boundary),
		Type:      "assistant",
		Source:    m.Source,
		IsPartial: m.Type == mail.MailTypePartialAssistant,
	}
}

// ParseActionItem parses action item shorthands from chat input
func (g *gatewayService) ParseActionItem(message string) (*ActionItem, error) {
	if len(message) == 0 {
		return nil, nil
	}

	if message == "@pause" {
		return &ActionItem{
			Type:    "pause",
			Payload: nil,
		}, nil
	}

	if len(message) > 15 && message[:15] == "@inject-memory " {
		return &ActionItem{
			Type:    "inject-memory",
			Payload: message[15:],
		}, nil
	}

	if len(message) > 0 && message[0] == '@' {
		return nil, nil
	}

	return nil, nil
}
