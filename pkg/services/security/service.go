package security

import (
	"sync"

	"github.com/maelstrom/v3/pkg/mail"
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

func (s *SecurityService) Start() error {
	return nil
}

func (s *SecurityService) Stop() error {
	return nil
}
