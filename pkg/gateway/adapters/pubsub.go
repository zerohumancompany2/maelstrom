package adapters

import (
	"crypto/rand"
	"fmt"

	"github.com/maelstrom/v3/pkg/mail"
)

type PubSubAdapter struct {
	name string
}

func NewPubSubAdapter() *PubSubAdapter {
	return &PubSubAdapter{name: "pubsub"}
}

func (a *PubSubAdapter) Name() string {
	return a.name
}

func (a *PubSubAdapter) NormalizeInbound(data []byte) (mail.Mail, error) {
	return mail.Mail{
		ID:      generatePubSubID(),
		Type:    mail.MailTypeMailReceived,
		Source:  "gateway:pubsub",
		Content: string(data),
		Metadata: mail.MailMetadata{
			Boundary: mail.OuterBoundary,
			Taints:   []string{"USER_SUPPLIED"},
		},
	}, nil
}

func (a *PubSubAdapter) NormalizeOutbound(mail mail.Mail) ([]byte, error) {
	return []byte(mail.Content.(string)), nil
}

func generatePubSubID() string {
	b := make([]byte, 8)
	rand.Read(b)
	return fmt.Sprintf("pubsub-%x", b)
}
