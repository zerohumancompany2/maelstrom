package security

import (
	"sync"
	"time"

	"github.com/maelstrom/v3/pkg/mail"
)

var (
	violationRouter *mail.MailRouter
	routerMu        sync.RWMutex
	violationCounts map[string]int
	countsMu        sync.RWMutex
)

func init() {
	violationCounts = make(map[string]int)
}

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
	countsMu.Lock()
	violationCounts[runtimeId]++
	countsMu.Unlock()

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
	countsMu.RLock()
	defer countsMu.RUnlock()
	return violationCounts[runtimeId]
}

func createViolationMail(violation TaintViolation) mail.Mail {
	forbiddenTaints := make([]interface{}, len(violation.ForbiddenTaints))
	for i, taint := range violation.ForbiddenTaints {
		forbiddenTaints[i] = taint
	}

	return mail.Mail{
		Type:   mail.MailTypeTaintViolation,
		Source: "sys:security",
		Target: "sys:observability",
		Content: map[string]interface{}{
			"runtimeId":       violation.RuntimeID,
			"sourceBoundary":  string(violation.SourceBoundary),
			"targetBoundary":  string(violation.TargetBoundary),
			"forbiddenTaints": forbiddenTaints,
			"timestamp":       violation.Timestamp,
		},
	}
}
