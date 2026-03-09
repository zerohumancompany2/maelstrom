package humangateway

import (
	"github.com/maelstrom/v3/pkg/mail"
	"testing"
)

func TestHumanGatewayService_HandleChat(t *testing.T) {
	svc := NewHumanGatewayService()

	m, err := svc.HandleChat("recommendation-agent", "What do you think?")
	if err != nil {
		t.Errorf("Expected nil error, got %v", err)
	}

	if m.Type != mail.MailTypeUser {
		t.Errorf("Expected MailTypeUser, got %s", m.Type)
	}

	if m.Source != "human:recommendation-agent" {
		t.Errorf("Expected source 'human:recommendation-agent', got '%s'", m.Source)
	}

	if m.Target != "agent:recommendation-agent" {
		t.Errorf("Expected target 'agent:recommendation-agent', got '%s'", m.Target)
	}

	content, ok := m.Content.(map[string]any)
	if !ok {
		t.Error("Expected content to be map[string]any")
	}
	if content["message"] != "What do you think?" {
		t.Error("Expected message preserved in content")
	}
}
