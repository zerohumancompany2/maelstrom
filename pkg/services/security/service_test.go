package security

import (
	"testing"

	"github.com/maelstrom/v3/pkg/mail"
)

func TestSecurityService_NewSecurityServiceReturnsNonNil(t *testing.T) {
	svc := NewSecurityService()

	if svc == nil {
		t.Error("Expected NewSecurityService to return non-nil")
	}
}

func TestSecurityService_IDReturnsCorrectString(t *testing.T) {
	svc := NewSecurityService()

	id := svc.ID()

	if id != "sys:security" {
		t.Errorf("Expected ID sys:security, got %s", id)
	}
}

func TestSecurityService_HandleMailReturnsNil(t *testing.T) {
	svc := NewSecurityService()

	err := svc.HandleMail(mail.Mail{})

	if err != nil {
		t.Errorf("Expected HandleMail to return nil, got %v", err)
	}
}

func TestSecurityService_ValidateBoundaryReturnsNil(t *testing.T) {
	svc := NewSecurityService()

	err := svc.ValidateBoundary(mail.InnerBoundary, mail.DMZBoundary)

	if err != nil {
		t.Errorf("Expected ValidateBoundary to return nil, got %v", err)
	}
}

func TestSecurityService_StartReturnsNil(t *testing.T) {
	svc := NewSecurityService()

	err := svc.Start()

	if err != nil {
		t.Errorf("Expected Start to return nil, got %v", err)
	}
}

func TestSecurityService_StopReturnsNil(t *testing.T) {
	svc := NewSecurityService()

	err := svc.Stop()

	if err != nil {
		t.Errorf("Expected Stop to return nil, got %v", err)
	}
}

func TestSecurityService_BootstrapChart(t *testing.T) {
	chart := BootstrapChart()

	if chart.ID != "sys:security" {
		t.Errorf("Expected ID sys:security, got %s", chart.ID)
	}

	if chart.Version != "1.0.0" {
		t.Errorf("Expected version 1.0.0, got %s", chart.Version)
	}
}

func TestSecurityService_HandleMail(t *testing.T) {
	svc := NewSecurityService()
	mail := mail.Mail{
		ID:     "test-mail-1",
		Source: "test-source",
		Target: "sys:security",
	}
	err := svc.HandleMail(mail)
	if err != nil {
		t.Errorf("HandleMail should return nil for Phase 1 pass-through, got: %v", err)
	}
}

func TestSecurityService_ID(t *testing.T) {
	chart := BootstrapChart()

	if chart.ID != "sys:security" {
		t.Errorf("Expected ID sys:security, got %s", chart.ID)
	}
}
