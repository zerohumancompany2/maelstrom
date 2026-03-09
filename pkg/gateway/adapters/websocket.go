package adapters

import (
	"encoding/json"

	"github.com/maelstrom/v3/pkg/mail"
)

type WebSocketAdapter struct {
	name string
}

func NewWebSocketAdapter() *WebSocketAdapter {
	return &WebSocketAdapter{name: "websocket"}
}

func (a *WebSocketAdapter) Name() string {
	return a.name
}

func (a *WebSocketAdapter) NormalizeInbound(data []byte) (mail.Mail, error) {
	var payload map[string]any
	if err := json.Unmarshal(data, &payload); err != nil {
		return mail.Mail{}, err
	}

	return mail.Mail{
		ID:      generateID(),
		Type:    mail.MailTypeMailReceived,
		Source:  "gateway:websocket",
		Content: payload,
		Metadata: mail.MailMetadata{
			Boundary: mail.OuterBoundary,
			Taints:   []string{"USER_SUPPLIED"},
		},
	}, nil
}

func (a *WebSocketAdapter) NormalizeOutbound(mailObj mail.Mail) ([]byte, error) {
	return json.Marshal(map[string]any{
		"type":    mailObj.Type,
		"id":      mailObj.ID,
		"content": mailObj.Content,
		"source":  mailObj.Source,
		"stream":  mailObj.Metadata.Stream != nil,
	})
}
