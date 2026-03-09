package security

import (
	"testing"
	"time"

	"github.com/maelstrom/v3/pkg/mail"
)

func TestReportViolation_CreateEvent(t *testing.T) {
	// Given
	violation := TaintViolation{
		RuntimeID:       "agent-123",
		SourceBoundary:  OuterBoundary,
		TargetBoundary:  InnerBoundary,
		ForbiddenTaints: []string{"PII"},
		Timestamp:       time.Now(),
	}

	// When
	resultMail := createViolationMail(violation)

	// Then
	if resultMail.Type != mail.MailTypeTaintViolation {
		t.Errorf("Expected Mail type to be MailTypeTaintViolation, got %v", resultMail.Type)
	}

	if resultMail.Source != "sys:security" {
		t.Errorf("Expected source to be 'sys:security', got %v", resultMail.Source)
	}

	if resultMail.Target != "sys:observability" {
		t.Errorf("Expected target to be 'sys:observability', got %v", resultMail.Target)
	}
}

func TestReportViolation_DeadLetter(t *testing.T) {
	// Given
	violation := TaintViolation{
		RuntimeID:       "agent-456",
		SourceBoundary:  DMZBoundary,
		TargetBoundary:  InnerBoundary,
		ForbiddenTaints: []string{"SECRET"},
		Timestamp:       time.Now(),
	}

	router := mail.NewMailRouter()
	observabilityInbox := &mail.ServiceInbox{ID: "observability"}
	router.SubscribeService("observability", observabilityInbox)
	SetViolationRouter(router)

	// When
	err := ReportViolation("agent-456", violation)

	// Then
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	receivedMail, err := observabilityInbox.Pop()
	if err != nil {
		t.Fatalf("Expected to receive mail in observability inbox, got error: %v", err)
	}

	if receivedMail.Type != mail.MailTypeTaintViolation {
		t.Errorf("Expected Mail type to be MailTypeTaintViolation, got %v", receivedMail.Type)
	}
}
