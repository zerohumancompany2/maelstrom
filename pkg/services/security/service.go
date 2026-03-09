package security

import (
	"sync"

	"github.com/maelstrom/v3/pkg/mail"
	"github.com/maelstrom/v3/pkg/security"
)

type SecurityService struct {
	mu sync.Mutex
}

func NewSecurityService() *SecurityService {
	return &SecurityService{}
}

func (s *SecurityService) ID() string {
	return "sys:security"
}

func (s *SecurityService) HandleMail(mail mail.Mail) error {
	return nil
}

func (s *SecurityService) ValidateBoundary(source, target mail.BoundaryType) error {
	return nil
}

func (s *SecurityService) ValidateAndSanitize(mail mail.Mail, sourceBoundary, targetBoundary mail.BoundaryType) (mail.Mail, error) {
	return mail, nil
}

func (s *SecurityService) TaintPropagate(obj any, newTaints []string) (any, error) {
	return obj, nil
}

func (s *SecurityService) ReportTaints(runtimeId string) (security.TaintMap, error) {
	return security.TaintMap{}, nil
}

func (s *SecurityService) PrepareContextForBoundary(runtimeId string, boundary mail.BoundaryType) error {
	return nil
}

func (s *SecurityService) Start() error {
	return nil
}

func (s *SecurityService) Stop() error {
	return nil
}
