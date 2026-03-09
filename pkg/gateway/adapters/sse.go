package adapters

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/maelstrom/v3/pkg/mail"
)

type SSEAdapter struct {
	name string
}

func NewSSEAdapter() *SSEAdapter {
	return &SSEAdapter{name: "sse"}
}

func (a *SSEAdapter) Name() string {
	return a.name
}

func (a *SSEAdapter) NormalizeInbound(data []byte) (mail.Mail, error) {
	return mail.Mail{
		ID:      generateID(),
		Type:    mail.MailTypeMailReceived,
		Source:  "gateway:sse",
		Content: string(data),
		Metadata: mail.MailMetadata{
			Boundary: mail.OuterBoundary,
			Taints:   []string{"USER_SUPPLIED"},
		},
	}, nil
}

func (a *SSEAdapter) NormalizeOutbound(mailObj mail.Mail) ([]byte, error) {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("event: %s\n", mailObj.Type))

	dataJSON, err := json.Marshal(mailObj.Content)
	if err != nil {
		return nil, err
	}
	sb.WriteString(fmt.Sprintf("data: %s\n", string(dataJSON)))

	if mailObj.Metadata.StreamChunk != nil {
		chunk := mailObj.Metadata.StreamChunk
		sb.WriteString(fmt.Sprintf("id: %d\n", chunk.Sequence))
		if chunk.IsFinal {
			sb.WriteString("event: end\n")
		}
	}

	sb.WriteString("\n")
	return []byte(sb.String()), nil
}
