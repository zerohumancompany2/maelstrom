package security

import (
	"fmt"
	"sync"
	"time"

	"github.com/maelstrom/v3/pkg/mail"
	"github.com/maelstrom/v3/pkg/security"
)

type SecurityService struct {
	mu          sync.Mutex
	taintTracks map[string]security.TaintMap
	publisher   mail.Publisher
}

type TaintViolationEvent struct {
	RuntimeID string
	ObjectID  string
	Taints    []string
	Reason    string
}

type NotImplementedError struct{}

func (NotImplementedError) Error() string {
	return "not implemented"
}

func NewSecurityService() *SecurityService {
	return &SecurityService{
		taintTracks: make(map[string]security.TaintMap),
	}
}

func (s *SecurityService) SetPublisher(publisher mail.Publisher) {
	s.publisher = publisher
}

func (s *SecurityService) EmitTaintViolation(runtimeId, objectID string, taints []string, reason string) error {
	if s.publisher == nil {
		return nil
	}

	event := TaintViolationEvent{
		RuntimeID: runtimeId,
		ObjectID:  objectID,
		Taints:    taints,
		Reason:    reason,
	}

	mail := mail.Mail{
		ID:        fmt.Sprintf("taint-violation-%s-%s", runtimeId, objectID),
		Type:      mail.MailTypeTaintViolation,
		Source:    "sys:security",
		Target:    "sys:observability",
		Content:   event,
		Metadata:  mail.MailMetadata{},
		CreatedAt: time.Now(),
	}

	_, err := s.publisher.Publish(mail)
	return err
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
	s.mu.Lock()
	defer s.mu.Unlock()

	taintMap, ok := s.taintTracks[runtimeId]
	if !ok {
		return make(security.TaintMap), nil
	}

	result := make(security.TaintMap)
	for k, v := range taintMap {
		result[k] = make([]string, len(v))
		copy(result[k], v)
	}

	return result, nil
}

func (s *SecurityService) TrackTaint(runtimeId, objectID, taint string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.taintTracks[runtimeId] == nil {
		s.taintTracks[runtimeId] = make(security.TaintMap)
	}

	s.taintTracks[runtimeId][objectID] = append(s.taintTracks[runtimeId][objectID], taint)

	return nil
}

func (s *SecurityService) PrepareContextForBoundary(runtimeId string, boundary mail.BoundaryType) error {
	return nil
}

func (s *SecurityService) CheckTaintPolicy(data any, targetBoundary mail.BoundaryType, policy security.TaintPolicy) (bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	taints := s.extractTaints(data)

	if policy.RedactMode == "strict" {
		for _, taint := range taints {
			if taint == "INNER_ONLY" || taint == "SECRET" {
				if targetBoundary != mail.InnerBoundary {
					return false, nil
				}
			}
			if taint == "PII" {
				if targetBoundary == mail.OuterBoundary {
					return false, nil
				}
			}
		}
	}

	allowedSet := make(map[security.BoundaryType]bool)
	for _, b := range policy.AllowedForBoundary {
		allowedSet[b] = true
	}

	secTargetBoundary := security.BoundaryType(targetBoundary)
	if !allowedSet[secTargetBoundary] {
		return false, nil
	}

	return true, nil
}

func (s *SecurityService) extractTaints(data any) []string {
	dataMap, ok := data.(map[string]interface{})
	if !ok {
		return nil
	}

	taints, ok := dataMap["_taints"].([]string)
	if !ok {
		return nil
	}

	return taints
}

func (s *SecurityService) NamespaceIsolate(runtimeId string, operation string) (security.IsolatedView, error) {
	return security.IsolatedView{
		RuntimeID:   runtimeId,
		Operation:   operation,
		Boundary:    security.DMZBoundary,
		ContextData: make(map[string]any),
	}, nil
}

func (s *SecurityService) NamespaceIsolateWithFilter(runtimeId string, operation string, data map[string]interface{}) (security.IsolatedView, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	filteredData := make(map[string]interface{})

	for key, value := range data {
		if key == "items" {
			items, ok := value.([]interface{})
			if !ok {
				filteredData[key] = value
				continue
			}

			var filteredItems []interface{}
			for _, item := range items {
				itemMap, ok := item.(map[string]interface{})
				if !ok {
					filteredItems = append(filteredItems, item)
					continue
				}

				if itemMap["agentID"] == runtimeId {
					filteredItems = append(filteredItems, item)
				}
			}

			filteredData[key] = filteredItems
		} else {
			filteredData[key] = value
		}
	}

	return security.IsolatedView{
		RuntimeID:   runtimeId,
		Operation:   operation,
		Boundary:    security.DMZBoundary,
		ContextData: filteredData,
	}, nil
}

func (s *SecurityService) Start() error {
	return nil
}

func (s *SecurityService) Stop() error {
	return nil
}
