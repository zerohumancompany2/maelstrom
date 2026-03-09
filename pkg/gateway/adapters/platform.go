package adapters

import (
	"crypto/rand"
	"fmt"

	"github.com/maelstrom/v3/pkg/mail"
)

type PlatformAdapter struct {
	name string
}

func NewPlatformAdapter(platform string) *PlatformAdapter {
	return &PlatformAdapter{name: platform}
}

func (a *PlatformAdapter) Name() string {
	return a.name
}

func (a *PlatformAdapter) NormalizeInbound(data []byte) (mail.Mail, error) {
	return mail.Mail{
		ID:      generatePlatformID(),
		Type:    mail.MailTypeMailReceived,
		Source:  "gateway:" + a.name,
		Content: string(data),
		Metadata: mail.MailMetadata{
			Boundary: mail.OuterBoundary,
			Taints:   []string{"USER_SUPPLIED"},
		},
	}, nil
}

func (a *PlatformAdapter) NormalizeOutbound(mail mail.Mail) ([]byte, error) {
	return []byte(mail.Content.(string)), nil
}

func generatePlatformID() string {
	b := make([]byte, 8)
	rand.Read(b)
	return fmt.Sprintf("platform-%x", b)
}

// Convenience constructors
func NewSlackAdapter() *PlatformAdapter {
	return NewPlatformAdapter("slack")
}

func NewWhatsAppAdapter() *PlatformAdapter {
	return NewPlatformAdapter("whatsapp")
}

func NewTelegramAdapter() *PlatformAdapter {
	return NewPlatformAdapter("telegram")
}
