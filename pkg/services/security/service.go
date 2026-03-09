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

func (s *SecurityService) ValidateAndSanitize(m mail.Mail, sourceBoundary, targetBoundary mail.BoundaryType) (mail.Mail, error) {
	result := m
	result.Metadata.Boundary = targetBoundary

	if sourceBoundary == mail.InnerBoundary && targetBoundary == mail.OuterBoundary {
		return result, nil
	}

	if sourceBoundary == mail.OuterBoundary && targetBoundary == mail.InnerBoundary {
		if result.Metadata.Taints == nil {
			result.Metadata.Taints = []string{}
		}
		result.Metadata.Taints = append(result.Metadata.Taints, "EXTERNAL")
		return result, nil
	}

	return result, nil
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

func (s *SecurityService) CheckTaintPolicy(taints []string, action string) bool {
	for _, taint := range taints {
		if taint == "UNTRUSTED" {
			return false
		}
	}
	return true
}

func (s *SecurityService) NamespaceIsolate(data interface{}, agentID string) interface{} {
	dataSlice, ok := data.([]interface{})
	if !ok {
		return data
	}

	var result []interface{}
	for _, item := range dataSlice {
		itemMap, ok := item.(map[string]interface{})
		if !ok {
			continue
		}

		if itemMap["agentID"] == agentID {
			result = append(result, item)
		}
	}

	return result
}

func (s *SecurityService) Start() error {
	return nil
}

func (s *SecurityService) Stop() error {
	return nil
}
