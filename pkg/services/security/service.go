package security

import (
	"sync"

	"github.com/maelstrom/v3/pkg/mail"
	"github.com/maelstrom/v3/pkg/security"
)

type SecurityService struct {
	mu sync.Mutex
}

type NotImplementedError struct{}

func (NotImplementedError) Error() string {
	return "not implemented"
}

func NewSecurityService() *SecurityService {
	return &SecurityService{}
}

func (s *SecurityService) ID() string {
	return "sys:security"
}

func (s *SecurityService) HandleMail(m *mail.Mail) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	sourceBoundary := m.Metadata.Boundary
	targetBoundary := mail.InnerBoundary

	if sourceBoundary == mail.OuterBoundary && targetBoundary == mail.InnerBoundary {
		if m.Metadata.Taints == nil {
			m.Metadata.Taints = []string{}
		}
		m.Metadata.Taints = append(m.Metadata.Taints, "EXTERNAL")
		m.Metadata.Boundary = targetBoundary
	}

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
	result, ok := obj.(map[string]interface{})
	if !ok {
		return obj, nil
	}

	seen := make(map[string]bool)
	merged := make([]string, 0)

	if existing, ok := result["_taints"].([]string); ok {
		for _, t := range existing {
			if !seen[t] {
				seen[t] = true
				merged = append(merged, t)
			}
		}
	}

	for _, t := range newTaints {
		if !seen[t] {
			seen[t] = true
			merged = append(merged, t)
		}
	}

	result["_taints"] = merged

	for key, value := range result {
		if key == "_taints" {
			continue
		}
		nested, ok := value.(map[string]interface{})
		if ok {
			propagated, err := s.TaintPropagate(nested, newTaints)
			if err != nil {
				return nil, err
			}
			result[key] = propagated
			continue
		}

		slice, ok := value.([]interface{})
		if ok {
			var propagatedSlice []interface{}
			for _, item := range slice {
				itemMap, ok := item.(map[string]interface{})
				if ok {
					propagated, err := s.TaintPropagate(itemMap, newTaints)
					if err != nil {
						return nil, err
					}
					propagatedSlice = append(propagatedSlice, propagated)
				} else {
					propagatedSlice = append(propagatedSlice, item)
				}
			}
			result[key] = propagatedSlice
		}
	}

	return result, nil
}

func (s *SecurityService) ReportTaints(runtimeId string) (security.TaintMap, error) {
	return nil, NotImplementedError{}
}

func (s *SecurityService) PrepareContextForBoundary(runtimeId string, boundary mail.BoundaryType) error {
	return nil
}

func (s *SecurityService) CheckTaintPolicy(data any, targetBoundary mail.BoundaryType, policy security.TaintPolicy) (bool, error) {
	return false, NotImplementedError{}
}

func (s *SecurityService) NamespaceIsolate(runtimeId string, operation string) (security.IsolatedView, error) {
	return security.IsolatedView{}, NotImplementedError{}
}

func (s *SecurityService) Start() error {
	return nil
}

func (s *SecurityService) Stop() error {
	return nil
}
