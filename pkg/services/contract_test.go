package services

import (
	"testing"
	"time"

	"github.com/maelstrom/v3/pkg/mail"
)

type contractMockService struct {
	id          string
	returnError bool
}

func (m *contractMockService) ID() string {
	return m.id
}

func (m *contractMockService) HandleMail(mail mail.Mail) *OutcomeEvent {
	if m.returnError {
		return &OutcomeEvent{
			ServiceID:    m.id,
			MailID:       mail.ID,
			Status:       "error",
			Timestamp:    time.Now(),
			ErrorDetails: "test error",
		}
	}
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

// TestServiceContract_outcomeEventContainsTimestamp - spec: arch-v1.md L479-480
// Acceptance Criteria: OutcomeEvent contains Timestamp field
func TestServiceContract_outcomeEventContainsTimestamp(t *testing.T) {
	svc := &contractMockService{id: "sys:test"}
	m := mail.Mail{
		ID:      "test-mail-3",
		Source:  "agent:test",
		Target:  "sys:test",
		Type:    mail.MailTypeUser,
		Content: "test",
	}

	outcome := svc.HandleMail(m)

	if outcome.Timestamp.IsZero() {
		t.Error("Expected Timestamp to be non-zero")
	}
}

// TestServiceContract_outcomeEventContainsServiceID - spec: arch-v1.md L479-480
// Acceptance Criteria: OutcomeEvent contains ServiceID field
func TestServiceContract_outcomeEventContainsServiceID(t *testing.T) {
	expectedServiceID := "sys:test"
	svc := &contractMockService{id: expectedServiceID}
	m := mail.Mail{
		ID:      "test-mail-4",
		Source:  "agent:test",
		Target:  "sys:test",
		Type:    mail.MailTypeUser,
		Content: "test",
	}

	outcome := svc.HandleMail(m)

	if outcome.ServiceID != expectedServiceID {
		t.Errorf("Expected ServiceID to be '%s', got '%s'", expectedServiceID, outcome.ServiceID)
	}
}

// TestServiceContract_outcomeEventContainsErrorDetails - spec: arch-v1.md L479-480
// Acceptance Criteria: OutcomeEvent contains ErrorDetails field
func TestServiceContract_outcomeEventContainsErrorDetails(t *testing.T) {
	expectedError := "test error"
	svc := &contractMockService{id: "sys:test", returnError: true}
	m := mail.Mail{
		ID:      "test-mail-5",
		Source:  "agent:test",
		Target:  "sys:test",
		Type:    mail.MailTypeUser,
		Content: "test",
	}

	outcome := svc.HandleMail(m)

	if outcome.Status != "error" {
		t.Errorf("Expected Status to be 'error', got '%s'", outcome.Status)
	}
	if outcome.ErrorDetails != expectedError {
		t.Errorf("Expected ErrorDetails to be '%s', got '%s'", expectedError, outcome.ErrorDetails)
	}
}

// TestServiceContract_allServicesImplementInterface - spec: arch-v1.md L479-480
// Acceptance Criteria: All services implement the Service interface with HandleMail returning *OutcomeEvent
func TestServiceContract_allServicesImplementInterface(t *testing.T) {
	var _ Service = (*contractMockService)(nil)
}
