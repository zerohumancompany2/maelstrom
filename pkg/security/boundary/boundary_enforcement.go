package boundary

import (
	"fmt"
	"time"

	"github.com/maelstrom/v3/pkg/mail"
	"github.com/maelstrom/v3/pkg/security"
)

type BoundaryType string

const (
	InnerBoundary BoundaryType = "inner"
	DMZBoundary   BoundaryType = "dmz"
	OuterBoundary BoundaryType = "outer"
)

type TransitionRule struct {
	Source    BoundaryType
	Target    BoundaryType
	Forbidden []string
	AutoStrip []string
	Allowed   []string
}

type TransitionRules struct {
	rules []TransitionRule
}

func NewTransitionRules() *TransitionRules {
	return &TransitionRules{
		rules: []TransitionRule{
			{
				Source:    InnerBoundary,
				Target:    DMZBoundary,
				Forbidden: []string{},
				AutoStrip: []string{"INNER_ONLY", "PII"},
				Allowed:   []string{"TOOL_OUTPUT", "EXTERNAL", "USER_SUPPLIED"},
			},
			{
				Source:    InnerBoundary,
				Target:    OuterBoundary,
				Forbidden: []string{},
				AutoStrip: []string{"SECRET", "PII", "TOOL_OUTPUT", "INNER_ONLY"},
				Allowed:   []string{"EXTERNAL", "USER_SUPPLIED"},
			},
			{
				Source:    DMZBoundary,
				Target:    InnerBoundary,
				Forbidden: []string{"SECRET"},
				AutoStrip: []string{},
				Allowed:   []string{"TOOL_OUTPUT", "EXTERNAL", "USER_SUPPLIED", "PII"},
			},
			{
				Source:    DMZBoundary,
				Target:    OuterBoundary,
				Forbidden: []string{},
				AutoStrip: []string{"PII"},
				Allowed:   []string{"TOOL_OUTPUT", "EXTERNAL", "USER_SUPPLIED"},
			},
			{
				Source:    OuterBoundary,
				Target:    InnerBoundary,
				Forbidden: []string{"PII", "SECRET", "INNER_ONLY"},
				AutoStrip: []string{},
				Allowed:   []string{"USER_SUPPLIED", "EXTERNAL"},
			},
			{
				Source:    OuterBoundary,
				Target:    DMZBoundary,
				Forbidden: []string{},
				AutoStrip: []string{},
				Allowed:   []string{"USER_SUPPLIED", "EXTERNAL", "TOOL_OUTPUT"},
			},
		},
	}
}

func (tr *TransitionRules) GetRule(source, target BoundaryType) *TransitionRule {
	for _, rule := range tr.rules {
		if rule.Source == source && rule.Target == target {
			return &rule
		}
	}
	return nil
}

func (tr *TransitionRules) IsForbidden(source, target BoundaryType, taint string) bool {
	rule := tr.GetRule(source, target)
	if rule == nil {
		return false
	}
	for _, forbidden := range rule.Forbidden {
		if forbidden == taint {
			return true
		}
	}
	return false
}

func (tr *TransitionRules) ShouldStrip(source, target BoundaryType, taint string) bool {
	rule := tr.GetRule(source, target)
	if rule == nil {
		return false
	}
	for _, strip := range rule.AutoStrip {
		if strip == taint {
			return true
		}
	}
	return false
}

func (tr *TransitionRules) IsAllowed(source, target BoundaryType, taint string) bool {
	if source == target {
		return true
	}
	rule := tr.GetRule(source, target)
	if rule == nil {
		return false
	}
	for _, allowed := range rule.Allowed {
		if allowed == taint {
			return true
		}
	}
	return false
}

type Violation struct {
	RuntimeID       string
	SourceBoundary  BoundaryType
	TargetBoundary  BoundaryType
	ForbiddenTaints []string
	Timestamp       time.Time
	MailID          string
}

type ViolationHandler struct {
	router *mail.MailRouter
}

func NewViolationHandler(router *mail.MailRouter) *ViolationHandler {
	return &ViolationHandler{
		router: router,
	}
}

func (vh *ViolationHandler) HandleViolation(violation Violation) error {
	violationMail := vh.createViolationMail(violation)
	return vh.publishViolation(violationMail, violation)
}

func (vh *ViolationHandler) createViolationMail(violation Violation) mail.Mail {
	forbiddenTaints := make([]interface{}, len(violation.ForbiddenTaints))
	for i, taint := range violation.ForbiddenTaints {
		forbiddenTaints[i] = taint
	}

	return mail.Mail{
		ID:     fmt.Sprintf("violation-%s-%d", violation.RuntimeID, violation.Timestamp.UnixNano()),
		Type:   mail.MailTypeTaintViolation,
		Source: "sys:security",
		Target: "sys:observability",
		Content: map[string]interface{}{
			"runtimeId":       violation.RuntimeID,
			"sourceBoundary":  string(violation.SourceBoundary),
			"targetBoundary":  string(violation.TargetBoundary),
			"forbiddenTaints": forbiddenTaints,
			"timestamp":       violation.Timestamp,
			"mailId":          violation.MailID,
		},
		Metadata: mail.MailMetadata{
			Boundary: mail.OuterBoundary,
		},
	}
}

func (vh *ViolationHandler) publishViolation(mail mail.Mail, violation Violation) error {
	if vh.router == nil {
		return nil
	}
	err := vh.router.Route(mail)
	if err != nil {
		return fmt.Errorf("failed to publish violation: %w", err)
	}
	return nil
}

type BoundaryValidator struct {
	rules *TransitionRules
}

func NewBoundaryValidator() *BoundaryValidator {
	return &BoundaryValidator{
		rules: NewTransitionRules(),
	}
}

func (bv *BoundaryValidator) ValidateTransition(source, target BoundaryType, taints []string) error {
	if source == target {
		return nil
	}

	for _, taint := range taints {
		if bv.rules.IsForbidden(source, target, taint) {
			return &TransitionError{
				Source:          source,
				Target:          target,
				ForbiddenTaints: []string{taint},
			}
		}
	}
	return nil
}

func (bv *BoundaryValidator) ValidateTaintsForBoundary(taints []string, boundary BoundaryType) error {
	for _, taint := range taints {
		if taint == "INNER_ONLY" && (boundary == DMZBoundary || boundary == OuterBoundary) {
			return fmt.Errorf("taint %q is forbidden on boundary %q", taint, boundary)
		}
		if taint == "SECRET" && (boundary == DMZBoundary || boundary == OuterBoundary) {
			return fmt.Errorf("taint %q is forbidden on boundary %q", taint, boundary)
		}
		if taint == "PII" && boundary == OuterBoundary {
			return fmt.Errorf("taint %q is forbidden on boundary %q", taint, boundary)
		}
	}
	return nil
}

func (bv *BoundaryValidator) GetForbiddenTaints(source, target BoundaryType) []string {
	rule := bv.rules.GetRule(source, target)
	if rule == nil {
		return nil
	}
	return rule.Forbidden
}

func (bv *BoundaryValidator) GetAutoStripTaints(source, target BoundaryType) []string {
	rule := bv.rules.GetRule(source, target)
	if rule == nil {
		return nil
	}
	return rule.AutoStrip
}

type TransitionError struct {
	Source          BoundaryType
	Target          BoundaryType
	ForbiddenTaints []string
}

func (e *TransitionError) Error() string {
	return fmt.Sprintf("transition from %q to %q forbidden: taints %v", e.Source, e.Target, e.ForbiddenTaints)
}

type BoundaryEnforcer struct {
	validator *BoundaryValidator
	handler   *ViolationHandler
	sanitizer *Sanitizer
}

func NewBoundaryEnforcer(handler *ViolationHandler) *BoundaryEnforcer {
	return &BoundaryEnforcer{
		validator: NewBoundaryValidator(),
		handler:   handler,
		sanitizer: NewSanitizer(),
	}
}

func (be *BoundaryEnforcer) Enforce(m mail.Mail, source, target BoundaryType) (mail.Mail, error) {
	taints := m.GetTaints()

	err := be.validator.ValidateTransition(source, target, taints)
	if err != nil {
		if transitionErr, ok := err.(*TransitionError); ok {
			violation := Violation{
				RuntimeID:       m.ID,
				SourceBoundary:  source,
				TargetBoundary:  target,
				ForbiddenTaints: transitionErr.ForbiddenTaints,
				Timestamp:       time.Now(),
				MailID:          m.ID,
			}
			_ = be.handler.HandleViolation(violation)
			return m, err
		}
		return m, err
	}

	sanitizedMail := be.sanitizer.Sanitize(m, source, target)
	sanitizedMail.Metadata.Boundary = mail.BoundaryType(string(target))

	return sanitizedMail, nil
}

func (be *BoundaryEnforcer) EnforceWithPropagation(sourceMail, targetMail mail.Mail, source, target BoundaryType) (mail.Mail, error) {
	taints := sourceMail.GetTaints()

	err := be.validator.ValidateTransition(source, target, taints)
	if err != nil {
		if transitionErr, ok := err.(*TransitionError); ok {
			violation := Violation{
				RuntimeID:       sourceMail.ID,
				SourceBoundary:  source,
				TargetBoundary:  target,
				ForbiddenTaints: transitionErr.ForbiddenTaints,
				Timestamp:       time.Now(),
				MailID:          sourceMail.ID,
			}
			_ = be.handler.HandleViolation(violation)
			return targetMail, err
		}
		return targetMail, err
	}

	propagatedMail := be.propagateTaints(sourceMail, targetMail)
	sanitizedMail := be.sanitizer.Sanitize(propagatedMail, source, target)
	sanitizedMail.Metadata.Boundary = mail.BoundaryType(string(target))

	return sanitizedMail, nil
}

func (be *BoundaryEnforcer) propagateTaints(sourceMail, targetMail mail.Mail) mail.Mail {
	mail.PropagateTaints(&sourceMail, &targetMail)
	return targetMail
}

func (be *BoundaryEnforcer) ValidateAndSanitize(m mail.Mail, source, target BoundaryType) (mail.Mail, []string, error) {
	taints := m.GetTaints()

	err := be.validator.ValidateTransition(source, target, taints)
	if err != nil {
		return m, nil, err
	}

	strippedTaints := be.validator.GetAutoStripTaints(source, target)
	sanitizedMail := be.sanitizer.Sanitize(m, source, target)
	sanitizedMail.Metadata.Boundary = mail.BoundaryType(string(target))

	return sanitizedMail, strippedTaints, nil
}

type Sanitizer struct {
}

func NewSanitizer() *Sanitizer {
	return &Sanitizer{}
}

func (s *Sanitizer) Sanitize(m mail.Mail, source, target BoundaryType) mail.Mail {
	if source == target {
		return m
	}

	strippedTaints := getAutoStripTaints(security.BoundaryType(source), security.BoundaryType(target))
	strippedSet := make(map[string]bool)
	for _, t := range strippedTaints {
		strippedSet[t] = true
	}

	remainingTaints := make([]string, 0)
	for _, t := range m.GetTaints() {
		if !strippedSet[t] {
			remainingTaints = append(remainingTaints, t)
		}
	}

	m.Taints = remainingTaints
	m.Metadata.Taints = remainingTaints

	return m
}

func getAutoStripTaints(source, target security.BoundaryType) []string {
	switch {
	case source == security.InnerBoundary && target == security.DMZBoundary:
		return []string{"INNER_ONLY", "PII"}
	case source == security.InnerBoundary && target == security.OuterBoundary:
		return []string{"SECRET", "PII", "TOOL_OUTPUT", "INNER_ONLY"}
	case source == security.DMZBoundary && target == security.OuterBoundary:
		return []string{"PII"}
	default:
		return nil
	}
}
