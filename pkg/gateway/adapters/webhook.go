package adapters

import (
	"github.com/maelstrom/v3/pkg/mail"
)

type WebhookAdapter struct {
	name string
}

func NewWebhookAdapter() *WebhookAdapter {
	panic("not implemented")
}

func (a *WebhookAdapter) Name() string {
	panic("not implemented")
}

func (a *WebhookAdapter) NormalizeInbound(data []byte) (mail.Mail, error) {
	panic("not implemented")
}

func (a *WebhookAdapter) NormalizeOutbound(mail mail.Mail) ([]byte, error) {
	panic("not implemented")
}
