package adapters

import (
	"crypto/rand"
	"fmt"

	"github.com/maelstrom/v3/pkg/mail"
)

type SMTPAdapter struct {
	name string
}

func NewSMTPAdapter() *SMTPAdapter {
	return &SMTPAdapter{name: "smtp"}
}

func (a *SMTPAdapter) Name() string {
	return a.name
}

func (a *SMTPAdapter) NormalizeInbound(data []byte) (mail.Mail, error) {
	return mail.Mail{
		ID:      generateSMTPID(),
		Type:    mail.MailTypeMailReceived,
		Source:  "gateway:smtp",
		Content: string(data),
		Metadata: mail.MailMetadata{
			Boundary: mail.OuterBoundary,
			Taints:   []string{"USER_SUPPLIED"},
		},
	}, nil
}

func (a *SMTPAdapter) NormalizeOutbound(mail mail.Mail) ([]byte, error) {
	return []byte(mail.Content.(string)), nil
}

func generateSMTPID() string {
	b := make([]byte, 8)
	rand.Read(b)
	return fmt.Sprintf("smtp-%x", b)
}
