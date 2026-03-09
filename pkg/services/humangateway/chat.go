package humangateway

import (
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
	for k, v := range ctx {
		sanitized[k] = v
	}
	return sanitized
}
