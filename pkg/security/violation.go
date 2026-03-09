package security

import (
	"time"

	"github.com/maelstrom/v3/pkg/mail"
)

type TaintViolation struct {
	RuntimeID       string
	SourceBoundary  BoundaryType
	TargetBoundary  BoundaryType
	ForbiddenTaints []string
	Timestamp       time.Time
}

type ViolationReport struct {
	Violations []TaintViolation
	Count      int
}

func ReportViolation(runtimeId string, violation TaintViolation) error {
	return nil
}

func GetViolationCount(runtimeId string) int {
	return 0
}

func createViolationMail(violation TaintViolation) mail.Mail {
	return mail.Mail{
		Type:   mail.MailTypeTaintViolation,
		Source: "sys:security",
		Target: "sys:observability",
	}
}
