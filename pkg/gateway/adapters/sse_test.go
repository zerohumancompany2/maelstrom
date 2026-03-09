package adapters

import (
	"strings"
	"testing"

	"github.com/maelstrom/v3/pkg/mail"
)

func TestSSEAdapter_NormalizeOutbound(t *testing.T) {
	adapter := NewSSEAdapter()

	// Test regular message
	mailObj := mail.Mail{
		ID:      "msg-001",
		Type:    mail.MailTypeAssistant,
		Content: "Hello, world!",
		Metadata: mail.MailMetadata{
			Stream: nil,
		},
	}

	data, err := adapter.NormalizeOutbound(mailObj)
	if err != nil {
		t.Errorf("Expected nil error, got %v", err)
	}

	str := string(data)
	if !strings.Contains(str, "event: assistant") {
		t.Errorf("Expected 'event: assistant' in output, got '%s'", str)
	}
	if !strings.Contains(str, "data:") {
		t.Errorf("Expected 'data:' in output, got '%s'", str)
	}

	// Test streaming message
	streamMail := mail.Mail{
		ID:      "msg-002",
		Type:    mail.MailTypePartialAssistant,
		Content: "chunk data",
		Metadata: mail.MailMetadata{
			Stream: &mail.StreamChunk{
				Sequence: 1,
				IsFinal:  false,
			},
		},
	}

	streamData, _ := adapter.NormalizeOutbound(streamMail)
	streamStr := string(streamData)
	if !strings.Contains(streamStr, "id: 1") {
		t.Errorf("Expected 'id: 1' in streaming output, got '%s'", streamStr)
	}
}
