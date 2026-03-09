package security

import (
	"testing"
	"time"

	"github.com/maelstrom/v3/pkg/mail"
)

func TestReportViolation_CreateEvent(t *testing.T) {
	// Given
	violation := TaintViolation{
		RuntimeID:       "agent-123",
		SourceBoundary:  OuterBoundary,
		TargetBoundary:  InnerBoundary,
		ForbiddenTaints: []string{"PII"},
		Timestamp:       time.Now(),
	}

	// When
	resultMail := createViolationMail(violation)

	// Then
	if resultMail.Type != mail.MailTypeTaintViolation {
		t.Errorf("Expected Mail type to be MailTypeTaintViolation, got %v", resultMail.Type)
	}

	if resultMail.Source != "sys:security" {
		t.Errorf("Expected source to be 'sys:security', got %v", resultMail.Source)
	}

	if resultMail.Target != "sys:observability" {
		t.Errorf("Expected target to be 'sys:observability', got %v", resultMail.Target)
	}
}

func TestReportViolation_DeadLetter(t *testing.T) {
	// Given
	violation := TaintViolation{
		RuntimeID:       "agent-456",
		SourceBoundary:  DMZBoundary,
		TargetBoundary:  InnerBoundary,
		ForbiddenTaints: []string{"SECRET"},
		Timestamp:       time.Now(),
	}

	router := mail.NewMailRouter()
	observabilityInbox := &mail.ServiceInbox{ID: "observability"}
	router.SubscribeService("observability", observabilityInbox)
	SetViolationRouter(router)

	// When
	err := ReportViolation("agent-456", violation)

	// Then
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	receivedMail, err := observabilityInbox.Pop()
	if err != nil {
		t.Fatalf("Expected to receive mail in observability inbox, got error: %v", err)
	}

	if receivedMail.Type != mail.MailTypeTaintViolation {
		t.Errorf("Expected Mail type to be MailTypeTaintViolation, got %v", receivedMail.Type)
	}
}

func TestReportViolation_Details(t *testing.T) {
	// Given
	violation := TaintViolation{
		RuntimeID:       "agent-789",
		SourceBoundary:  OuterBoundary,
		TargetBoundary:  InnerBoundary,
		ForbiddenTaints: []string{"PII", "SECRET"},
		Timestamp:       time.Now(),
	}

	router := mail.NewMailRouter()
	observabilityInbox := &mail.ServiceInbox{ID: "observability"}
	router.SubscribeService("observability", observabilityInbox)
	SetViolationRouter(router)

	// When
	err := ReportViolation("agent-789", violation)

	// Then
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	receivedMail, err := observabilityInbox.Pop()
	if err != nil {
		t.Fatalf("Expected to receive mail in observability inbox, got error: %v", err)
	}

	content, ok := receivedMail.Content.(map[string]interface{})
	if !ok {
		t.Fatalf("Expected content to be map[string]interface{}, got %T", receivedMail.Content)
	}

	if content["runtimeId"] != "agent-789" {
		t.Errorf("Expected runtimeId to be 'agent-789', got %v", content["runtimeId"])
	}

	if content["sourceBoundary"] != "outer" {
		t.Errorf("Expected sourceBoundary to be 'outer', got %v", content["sourceBoundary"])
	}

	if content["targetBoundary"] != "inner" {
		t.Errorf("Expected targetBoundary to be 'inner', got %v", content["targetBoundary"])
	}

	forbiddenTaints, ok := content["forbiddenTaints"].([]interface{})
	if !ok {
		t.Fatalf("Expected forbiddenTaints to be []interface{}, got %T", content["forbiddenTaints"])
	}

	if len(forbiddenTaints) != 2 {
		t.Errorf("Expected 2 forbidden taints, got %d", len(forbiddenTaints))
	}

	if _, ok := content["timestamp"]; !ok {
		t.Error("Expected timestamp to be present in content")
	}
}

func TestReportViolation_Count(t *testing.T) {
	// Given
	engine := NewTaintEngine()

	violation1 := TaintViolation{
		RuntimeID:       "agent-count-test",
		SourceBoundary:  OuterBoundary,
		TargetBoundary:  InnerBoundary,
		ForbiddenTaints: []string{"PII"},
		Timestamp:       time.Now(),
	}

	violation2 := TaintViolation{
		RuntimeID:       "agent-count-test",
		SourceBoundary:  OuterBoundary,
		TargetBoundary:  InnerBoundary,
		ForbiddenTaints: []string{"SECRET"},
		Timestamp:       time.Now(),
	}

	// When
	_, _ = engine.ReportTaints("agent-count-test")

	err1 := ReportViolation("agent-count-test", violation1)
	if err1 != nil {
		t.Fatalf("Expected no error from first ReportViolation, got %v", err1)
	}

	err2 := ReportViolation("agent-count-test", violation2)
	if err2 != nil {
		t.Fatalf("Expected no error from second ReportViolation, got %v", err2)
	}

	taintMap, err := engine.ReportTaints("agent-count-test")

	// Then
	if err != nil {
		t.Fatalf("Expected no error from ReportTaints, got %v", err)
	}

	count := GetViolationCount("agent-count-test")
	if count != 2 {
		t.Errorf("Expected violation count to be 2, got %d", count)
	}

	if taintMap == nil {
		t.Error("Expected non-nil TaintMap from ReportTaints")
	}
}
