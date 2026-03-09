package security

import (
	"sync"
	"time"

	"github.com/maelstrom/v3/pkg/mail"
)

var (
	violationRouter *mail.MailRouter
	routerMu        sync.RWMutex
)

func SetViolationRouter(router *mail.MailRouter) {
	routerMu.Lock()
	defer routerMu.Unlock()
	violationRouter = router
}

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
	mail := createViolationMail(violation)

	routerMu.RLock()
	router := violationRouter
	routerMu.RUnlock()

	if router == nil {
		return nil
	}

	return router.Route(mail)
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
