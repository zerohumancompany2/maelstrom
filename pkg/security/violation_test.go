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
