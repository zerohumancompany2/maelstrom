package adapters

import (
	"crypto/rand"
	"encoding/json"
	"fmt"

	"github.com/maelstrom/v3/pkg/mail"
)

type WebhookAdapter struct {
	name string
}

func NewWebhookAdapter() *WebhookAdapter {
	return &WebhookAdapter{name: "webhook"}
}

func (a *WebhookAdapter) Name() string {
	return a.name
}

func (a *WebhookAdapter) NormalizeInbound(data []byte) (mail.Mail, error) {
	var payload map[string]any
	if err := json.Unmarshal(data, &payload); err != nil {
		return mail.Mail{}, err
	}

	return mail.Mail{
		ID:      generateID(),
		Type:    mail.MailTypeMailReceived,
		Source:  "gateway:webhook",
		Target:  "agent:default",
		Content: payload,
		Metadata: mail.MailMetadata{
			Boundary: mail.OuterBoundary,
			Taints:   []string{"USER_SUPPLIED"},
		},
	}, nil
}

func (a *WebhookAdapter) NormalizeOutbound(mail mail.Mail) ([]byte, error) {
	return json.Marshal(map[string]any{
		"type":    mail.Type,
		"content": mail.Content,
		"source":  mail.Source,
	})
}

func generateID() string {
	b := make([]byte, 8)
	rand.Read(b)
	return fmt.Sprintf("webhook-%x", b)
}
