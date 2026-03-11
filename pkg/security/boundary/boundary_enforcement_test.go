package boundary

import (
	"testing"
	"time"

	"github.com/maelstrom/v3/pkg/mail"
)

func TestTransitionRules_InnerToDMZ(t *testing.T) {
	rules := NewTransitionRules()
	rule := rules.GetRule(InnerBoundary, DMZBoundary)

	if rule == nil {
		t.Fatal("Expected rule for InnerBoundary -> DMZBoundary")
	}

	if len(rule.Forbidden) != 0 {
		t.Errorf("Expected 0 forbidden taints, got %d", len(rule.Forbidden))
	}

	expectedStrip := []string{"INNER_ONLY", "PII"}
	if len(rule.AutoStrip) != len(expectedStrip) {
		t.Errorf("Expected %v auto-strip taints, got %v", expectedStrip, rule.AutoStrip)
	}
}

func TestTransitionRules_InnerToOuter(t *testing.T) {
	rules := NewTransitionRules()
	rule := rules.GetRule(InnerBoundary, OuterBoundary)

	if rule == nil {
		t.Fatal("Expected rule for InnerBoundary -> OuterBoundary")
	}

	expectedStrip := []string{"SECRET", "PII", "TOOL_OUTPUT", "INNER_ONLY"}
	if len(rule.AutoStrip) != len(expectedStrip) {
		t.Errorf("Expected %v auto-strip taints, got %v", expectedStrip, rule.AutoStrip)
	}
}

func TestTransitionRules_OuterToInner(t *testing.T) {
	rules := NewTransitionRules()
	rule := rules.GetRule(OuterBoundary, InnerBoundary)

	if rule == nil {
		t.Fatal("Expected rule for OuterBoundary -> InnerBoundary")
	}

	expectedForbidden := []string{"PII", "SECRET", "INNER_ONLY"}
	if len(rule.Forbidden) != len(expectedForbidden) {
		t.Errorf("Expected %v forbidden taints, got %v", expectedForbidden, rule.Forbidden)
	}
}

func TestTransitionRules_DMZToInner(t *testing.T) {
	rules := NewTransitionRules()
	rule := rules.GetRule(DMZBoundary, InnerBoundary)

	if rule == nil {
		t.Fatal("Expected rule for DMZBoundary -> InnerBoundary")
	}

	expectedForbidden := []string{"SECRET"}
	if len(rule.Forbidden) != len(expectedForbidden) {
		t.Errorf("Expected %v forbidden taints, got %v", expectedForbidden, rule.Forbidden)
	}
}

func TestTransitionRules_IsForbidden(t *testing.T) {
	rules := NewTransitionRules()

	if !rules.IsForbidden(OuterBoundary, InnerBoundary, "PII") {
		t.Error("Expected PII to be forbidden for Outer -> Inner")
	}

	if !rules.IsForbidden(OuterBoundary, InnerBoundary, "SECRET") {
		t.Error("Expected SECRET to be forbidden for Outer -> Inner")
	}

	if !rules.IsForbidden(DMZBoundary, InnerBoundary, "SECRET") {
		t.Error("Expected SECRET to be forbidden for DMZ -> Inner")
	}

	if rules.IsForbidden(InnerBoundary, DMZBoundary, "TOOL_OUTPUT") {
		t.Error("Expected TOOL_OUTPUT to be allowed for Inner -> DMZ")
	}
}

func TestTransitionRules_ShouldStrip(t *testing.T) {
	rules := NewTransitionRules()

	if !rules.ShouldStrip(InnerBoundary, DMZBoundary, "INNER_ONLY") {
		t.Error("Expected INNER_ONLY to be stripped for Inner -> DMZ")
	}

	if !rules.ShouldStrip(InnerBoundary, DMZBoundary, "PII") {
		t.Error("Expected PII to be stripped for Inner -> DMZ")
	}

	if !rules.ShouldStrip(InnerBoundary, OuterBoundary, "SECRET") {
		t.Error("Expected SECRET to be stripped for Inner -> Outer")
	}

	if rules.ShouldStrip(OuterBoundary, DMZBoundary, "USER_SUPPLIED") {
		t.Error("Expected USER_SUPPLIED to NOT be stripped for Outer -> DMZ")
	}
}

func TestTransitionRules_IsAllowed(t *testing.T) {
	rules := NewTransitionRules()

	if !rules.IsAllowed(InnerBoundary, DMZBoundary, "TOOL_OUTPUT") {
		t.Error("Expected TOOL_OUTPUT to be allowed for Inner -> DMZ")
	}

	if !rules.IsAllowed(OuterBoundary, DMZBoundary, "USER_SUPPLIED") {
		t.Error("Expected USER_SUPPLIED to be allowed for Outer -> DMZ")
	}

	if !rules.IsAllowed(DMZBoundary, DMZBoundary, "TOOL_OUTPUT") {
		t.Error("Expected TOOL_OUTPUT to be allowed for same boundary")
	}
}

func TestBoundaryValidator_ValidateTransition_Allowed(t *testing.T) {
	validator := NewBoundaryValidator()

	err := validator.ValidateTransition(InnerBoundary, DMZBoundary, []string{"TOOL_OUTPUT"})
	if err != nil {
		t.Errorf("Expected no error for allowed transition, got %v", err)
	}

	err = validator.ValidateTransition(OuterBoundary, DMZBoundary, []string{"USER_SUPPLIED"})
	if err != nil {
		t.Errorf("Expected no error for allowed transition, got %v", err)
	}
}

func TestBoundaryValidator_ValidateTransition_Forbidden(t *testing.T) {
	validator := NewBoundaryValidator()

	err := validator.ValidateTransition(OuterBoundary, InnerBoundary, []string{"PII"})
	if err == nil {
		t.Error("Expected error for forbidden PII transition")
	}

	transitionErr, ok := err.(*TransitionError)
	if !ok {
		t.Errorf("Expected TransitionError, got %T", err)
	}

	if len(transitionErr.ForbiddenTaints) != 1 {
		t.Errorf("Expected 1 forbidden taint, got %d", len(transitionErr.ForbiddenTaints))
	}
}

func TestBoundaryValidator_ValidateTransition_SameBoundary(t *testing.T) {
	validator := NewBoundaryValidator()

	err := validator.ValidateTransition(InnerBoundary, InnerBoundary, []string{"SECRET", "INNER_ONLY"})
	if err != nil {
		t.Errorf("Expected no error for same boundary transition, got %v", err)
	}
}

func TestBoundaryValidator_ValidateTaintsForBoundary(t *testing.T) {
	validator := NewBoundaryValidator()

	err := validator.ValidateTaintsForBoundary([]string{"INNER_ONLY"}, OuterBoundary)
	if err == nil {
		t.Error("Expected error for INNER_ONLY on OuterBoundary")
	}

	err = validator.ValidateTaintsForBoundary([]string{"SECRET"}, DMZBoundary)
	if err == nil {
		t.Error("Expected error for SECRET on DMZBoundary")
	}

	err = validator.ValidateTaintsForBoundary([]string{"PII"}, OuterBoundary)
	if err == nil {
		t.Error("Expected error for PII on OuterBoundary")
	}

	err = validator.ValidateTaintsForBoundary([]string{"TOOL_OUTPUT"}, OuterBoundary)
	if err != nil {
		t.Errorf("Expected no error for TOOL_OUTPUT on OuterBoundary, got %v", err)
	}
}

func TestBoundaryValidator_GetForbiddenTaints(t *testing.T) {
	validator := NewBoundaryValidator()

	forbidden := validator.GetForbiddenTaints(OuterBoundary, InnerBoundary)
	expected := []string{"PII", "SECRET", "INNER_ONLY"}
	if len(forbidden) != len(expected) {
		t.Errorf("Expected %d forbidden taints, got %d", len(expected), len(forbidden))
	}
}

func TestBoundaryValidator_GetAutoStripTaints(t *testing.T) {
	validator := NewBoundaryValidator()

	stripped := validator.GetAutoStripTaints(InnerBoundary, OuterBoundary)
	expected := []string{"SECRET", "PII", "TOOL_OUTPUT", "INNER_ONLY"}
	if len(stripped) != len(expected) {
		t.Errorf("Expected %d auto-strip taints, got %d", len(expected), len(stripped))
	}
}

func TestViolationHandler_CreateViolationMail(t *testing.T) {
	handler := NewViolationHandler(nil)

	violation := Violation{
		RuntimeID:       "agent-123",
		SourceBoundary:  OuterBoundary,
		TargetBoundary:  InnerBoundary,
		ForbiddenTaints: []string{"PII", "SECRET"},
		Timestamp:       time.Now(),
		MailID:          "mail-456",
	}

	m := handler.createViolationMail(violation)

	if m.Type != mail.MailTypeTaintViolation {
		t.Errorf("Expected MailTypeTaintViolation, got %v", m.Type)
	}

	if m.Source != "sys:security" {
		t.Errorf("Expected source 'sys:security', got %v", m.Source)
	}

	if m.Target != "sys:observability" {
		t.Errorf("Expected target 'sys:observability', got %v", m.Target)
	}

	content, ok := m.Content.(map[string]interface{})
	if !ok {
		t.Fatal("Expected content to be map[string]interface{}")
	}

	if content["runtimeId"] != "agent-123" {
		t.Errorf("Expected runtimeId 'agent-123', got %v", content["runtimeId"])
	}
}

func TestViolationHandler_HandleViolation_Publish(t *testing.T) {
	router := mail.NewMailRouter()
	observabilityInbox := &mail.ServiceInbox{ID: "observability"}
	router.SubscribeService("observability", observabilityInbox)

	handler := NewViolationHandler(router)

	violation := Violation{
		RuntimeID:       "agent-456",
		SourceBoundary:  DMZBoundary,
		TargetBoundary:  InnerBoundary,
		ForbiddenTaints: []string{"SECRET"},
		Timestamp:       time.Now(),
		MailID:          "mail-789",
	}

	err := handler.HandleViolation(violation)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	receivedMail, err := observabilityInbox.Pop()
	if err != nil {
		t.Fatalf("Expected to receive mail, got error: %v", err)
	}

	if receivedMail.Type != mail.MailTypeTaintViolation {
		t.Errorf("Expected MailTypeTaintViolation, got %v", receivedMail.Type)
	}
}

func TestViolationHandler_HandleViolation_NoRouter(t *testing.T) {
	handler := NewViolationHandler(nil)

	violation := Violation{
		RuntimeID:       "agent-no-router",
		SourceBoundary:  OuterBoundary,
		TargetBoundary:  InnerBoundary,
		ForbiddenTaints: []string{"PII"},
		Timestamp:       time.Now(),
		MailID:          "mail-no-router",
	}

	err := handler.HandleViolation(violation)
	if err != nil {
		t.Errorf("Expected no error when router is nil, got %v", err)
	}
}

func TestBoundaryEnforcer_Enforce_Allowed(t *testing.T) {
	handler := NewViolationHandler(nil)
	enforcer := NewBoundaryEnforcer(handler)

	m := mail.Mail{
		ID:     "mail-allowed",
		Type:   mail.MailTypeUser,
		Taints: []string{"TOOL_OUTPUT"},
	}

	result, err := enforcer.Enforce(m, InnerBoundary, DMZBoundary)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if result.Metadata.Boundary != mail.DMZBoundary {
		t.Errorf("Expected boundary DMZBoundary, got %v", result.Metadata.Boundary)
	}
}

func TestBoundaryEnforcer_Enforce_Forbidden(t *testing.T) {
	handler := NewViolationHandler(nil)
	enforcer := NewBoundaryEnforcer(handler)

	m := mail.Mail{
		ID:     "mail-forbidden",
		Type:   mail.MailTypeUser,
		Taints: []string{"PII"},
	}

	_, err := enforcer.Enforce(m, OuterBoundary, InnerBoundary)
	if err == nil {
		t.Error("Expected error for forbidden transition")
	}
}

func TestBoundaryEnforcer_EnforceWithPropagation(t *testing.T) {
	handler := NewViolationHandler(nil)
	enforcer := NewBoundaryEnforcer(handler)

	sourceMail := mail.Mail{
		ID:     "source-mail",
		Type:   mail.MailTypeUser,
		Taints: []string{"TOOL_OUTPUT"},
	}

	targetMail := mail.Mail{
		ID:     "target-mail",
		Type:   mail.MailTypeAssistant,
		Taints: []string{},
	}

	result, err := enforcer.EnforceWithPropagation(sourceMail, targetMail, InnerBoundary, DMZBoundary)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(result.Taints) != 1 {
		t.Errorf("Expected 1 propagated taint, got %d", len(result.Taints))
	}

	if result.Taints[0] != "TOOL_OUTPUT" {
		t.Errorf("Expected TOOL_OUTPUT taint, got %v", result.Taints[0])
	}
}

func TestBoundaryEnforcer_ValidateAndSanitize(t *testing.T) {
	handler := NewViolationHandler(nil)
	enforcer := NewBoundaryEnforcer(handler)

	m := mail.Mail{
		ID:     "mail-sanitize",
		Type:   mail.MailTypeUser,
		Taints: []string{"INNER_ONLY", "TOOL_OUTPUT"},
	}

	result, stripped, err := enforcer.ValidateAndSanitize(m, InnerBoundary, DMZBoundary)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(stripped) != 2 {
		t.Errorf("Expected 2 auto-strip taints, got %d", len(stripped))
	}

	if len(result.Taints) != 1 {
		t.Errorf("Expected 1 remaining taint, got %d", len(result.Taints))
	}
}

func TestSanitizer_Sanitize_StripTaints(t *testing.T) {
	sanitizer := NewSanitizer()

	m := mail.Mail{
		ID:     "mail-strip",
		Type:   mail.MailTypeUser,
		Taints: []string{"SECRET", "PII", "TOOL_OUTPUT", "INNER_ONLY", "EXTERNAL"},
	}

	result := sanitizer.Sanitize(m, InnerBoundary, OuterBoundary)

	expectedTaints := []string{"EXTERNAL"}
	if len(result.Taints) != len(expectedTaints) {
		t.Errorf("Expected %v taints, got %v", expectedTaints, result.Taints)
	}
}

func TestSanitizer_Sanitize_SameBoundary(t *testing.T) {
	sanitizer := NewSanitizer()

	m := mail.Mail{
		ID:     "mail-same",
		Type:   mail.MailTypeUser,
		Taints: []string{"SECRET", "INNER_ONLY"},
	}

	result := sanitizer.Sanitize(m, InnerBoundary, InnerBoundary)

	if len(result.Taints) != 2 {
		t.Errorf("Expected 2 taints for same boundary, got %d", len(result.Taints))
	}
}

func TestTransitionError_ErrorMessage(t *testing.T) {
	err := &TransitionError{
		Source:          OuterBoundary,
		Target:          InnerBoundary,
		ForbiddenTaints: []string{"PII", "SECRET"},
	}

	expectedMsg := "transition from \"outer\" to \"inner\" forbidden: taints [PII SECRET]"
	if err.Error() != expectedMsg {
		t.Errorf("Expected error message %q, got %q", expectedMsg, err.Error())
	}
}

func TestBoundaryEnforcer_Enforce_ViolationCreated(t *testing.T) {
	router := mail.NewMailRouter()
	observabilityInbox := &mail.ServiceInbox{ID: "observability"}
	router.SubscribeService("observability", observabilityInbox)

	handler := NewViolationHandler(router)
	enforcer := NewBoundaryEnforcer(handler)

	m := mail.Mail{
		ID:     "mail-violation",
		Type:   mail.MailTypeUser,
		Taints: []string{"SECRET"},
	}

	_, err := enforcer.Enforce(m, DMZBoundary, InnerBoundary)
	if err == nil {
		t.Error("Expected error for forbidden transition")
	}

	receivedMail, err := observabilityInbox.Pop()
	if err != nil {
		t.Fatalf("Expected violation mail to be published, got error: %v", err)
	}

	if receivedMail.Type != mail.MailTypeTaintViolation {
		t.Errorf("Expected MailTypeTaintViolation, got %v", receivedMail.Type)
	}
}

func TestTransitionRules_DMZToOuter(t *testing.T) {
	rules := NewTransitionRules()
	rule := rules.GetRule(DMZBoundary, OuterBoundary)

	if rule == nil {
		t.Fatal("Expected rule for DMZBoundary -> OuterBoundary")
	}

	expectedStrip := []string{"PII"}
	if len(rule.AutoStrip) != len(expectedStrip) {
		t.Errorf("Expected %v auto-strip taints, got %v", expectedStrip, rule.AutoStrip)
	}

	if len(rule.Forbidden) != 0 {
		t.Errorf("Expected 0 forbidden taints, got %d", len(rule.Forbidden))
	}
}

func TestBoundaryValidator_MultipleForbiddenTaints(t *testing.T) {
	validator := NewBoundaryValidator()

	err := validator.ValidateTransition(OuterBoundary, InnerBoundary, []string{"PII", "SECRET", "USER_SUPPLIED"})
	if err == nil {
		t.Error("Expected error for forbidden taints")
	}

	transitionErr, ok := err.(*TransitionError)
	if !ok {
		t.Errorf("Expected TransitionError, got %T", err)
	}

	if len(transitionErr.ForbiddenTaints) != 1 {
		t.Errorf("Expected 1 forbidden taint in error, got %d", len(transitionErr.ForbiddenTaints))
	}

	if transitionErr.ForbiddenTaints[0] != "PII" {
		t.Errorf("Expected PII as first forbidden taint, got %v", transitionErr.ForbiddenTaints[0])
	}
}
