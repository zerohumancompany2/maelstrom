package humangateway

import (
	"fmt"
	"time"

	"github.com/maelstrom/v3/pkg/mail"
	"strings"
)

type ChatSession struct {
	AgentID    string
	Messages   []mail.Mail
	ContextMap ContextMapSnapshot
}

type ContextMapSnapshot map[string]any

type ActionItem struct {
	Type    string
	Target  string
	Payload any
}

func (h *HumanGatewayService) ParseActionItem(message string) ([]ActionItem, error) {
	var items []ActionItem

	if strings.Contains(message, "@pause") {
		items = append(items, ActionItem{
			Type:    "pause",
			Target:  "",
			Payload: nil,
		})
	}

	if idx := strings.Index(message, "@inject-memory"); idx >= 0 {
		content := strings.TrimSpace(message[idx+14:])
		items = append(items, ActionItem{
			Type:    "inject-memory",
			Target:  "",
			Payload: content,
		})
	}

	return items, nil
}

func SanitizeContextForBoundary(ctx ContextMapSnapshot, boundary mail.BoundaryType) ContextMapSnapshot {
	sanitized := make(ContextMapSnapshot)
	
	forbiddenTaints := map[string]bool{
		"FORBIDDEN": true,
		"INTERNAL":  true,
		"SECRET":    true,
	}

	for k, v := range ctx {
		switch boundary {
		case mail.InnerBoundary:
			sanitized[k] = v
		case mail.DMZBoundary:
			if vm, ok := v.(map[string]any); ok {
				cleaned := make(map[string]any)
				for tk, tv := range vm {
					if !forbiddenTaints[tk] {
						cleaned[tk] = tv
					}
				}
				sanitized[k] = cleaned
			} else {
				sanitized[k] = v
			}
		case mail.OuterBoundary:
			if _, ok := v.(map[string]any); ok {
				continue
			}
			sanitized[k] = v
		}
	}
	return sanitized
}

func (h *HumanGatewayService) SendMessage(session *ChatSession, message string) error {
	if session == nil {
		return fmt.Errorf("nil session")
	}

	mailMsg := mail.Mail{
		ID:     fmt.Sprintf("human-%d", time.Now().UnixNano()),
		Type:   mail.MailTypeHumanFeedback,
		Source: "human:" + session.AgentID,
		Target: "agent:" + session.AgentID,
		Content: map[string]any{
			"message": message,
		},
		Metadata: mail.MailMetadata{
			Boundary: mail.InnerBoundary,
			Taints:   []string{"USER_SUPPLIED"},
		},
	}

	session.Messages = append(session.Messages, mailMsg)
	return nil
}
