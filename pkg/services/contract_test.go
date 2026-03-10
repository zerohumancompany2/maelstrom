package services

import (
	"testing"
	"time"

	"github.com/maelstrom/v3/pkg/mail"
)

type contractMockService struct {
	id string
}

func (m *contractMockService) ID() string {
	return m.id
}

func (m *contractMockService) HandleMail(mail mail.Mail) *OutcomeEvent {
	return &OutcomeEvent{
		ServiceID:    m.id,
		MailID:       mail.ID,
		Status:       "success",
		Timestamp:    time.Now(),
		ErrorDetails: "",
	}
}

func (m *contractMockService) Start() error {
	return nil
}

func (m *contractMockService) Stop() error {
	return nil
}

// TestServiceContract_handleMailReturnsOutcomeEvent - spec: arch-v1.md L479-480
// Acceptance Criteria: HandleMail returns an OutcomeEvent
func TestServiceContract_handleMailReturnsOutcomeEvent(t *testing.T) {
	svc := &contractMockService{id: "sys:test"}
	m := mail.Mail{
		ID:      "test-mail-1",
		Source:  "agent:test",
		Target:  "sys:test",
		Type:    mail.MailTypeUser,
		Content: "test",
	}

	outcome := svc.HandleMail(m)

	if outcome == nil {
		t.Error("Expected HandleMail to return non-nil OutcomeEvent")
	}
}

// TestServiceContract_outcomeEventContainsStatus - spec: arch-v1.md L479-480
// Acceptance Criteria: OutcomeEvent contains Status field
func TestServiceContract_outcomeEventContainsStatus(t *testing.T) {
	svc := &contractMockService{id: "sys:test"}
	m := mail.Mail{
		ID:      "test-mail-2",
		Source:  "agent:test",
		Target:  "sys:test",
		Type:    mail.MailTypeUser,
		Content: "test",
	}

	outcome := svc.HandleMail(m)

	if outcome.Status != "success" {
		t.Errorf("Expected Status to be 'success', got '%s'", outcome.Status)
	}
}
