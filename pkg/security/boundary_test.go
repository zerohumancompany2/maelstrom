package security

import (
	"strings"
	"testing"
)

func TestValidateBoundary_Valid(t *testing.T) {
	tests := []BoundaryType{InnerBoundary, DMZBoundary, OuterBoundary}
	for _, boundary := range tests {
		err := ValidateBoundary(boundary)
		if err != nil {
			t.Errorf("ValidateBoundary(%q) = %v, want nil", boundary, err)
		}
	}
}

func TestValidateBoundary_Invalid(t *testing.T) {
	tests := []BoundaryType{"", "public", "unknown"}
	for _, boundary := range tests {
		err := ValidateBoundary(boundary)
		if err == nil {
			t.Errorf("ValidateBoundary(%q) = nil, want error", boundary)
		}
	}
}

func TestInnerCapabilities(t *testing.T) {
	caps := GetBoundaryCapabilities(InnerBoundary)
	expected := Capabilities{
		CanReadSecrets:       true,
		CanWriteSecrets:      true,
		RequiresSanitization: false,
		IngressOnly:          false,
		MailOnlyTransitions:  false,
	}
	if caps != expected {
		t.Errorf("GetBoundaryCapabilities(InnerBoundary) = %v, want %v", caps, expected)
	}
}

func TestDMZCapabilities(t *testing.T) {
	caps := GetBoundaryCapabilities(DMZBoundary)
	expected := Capabilities{
		CanReadSecrets:       false,
		CanWriteSecrets:      false,
		RequiresSanitization: true,
		IngressOnly:          false,
		MailOnlyTransitions:  true,
	}
	if caps != expected {
		t.Errorf("GetBoundaryCapabilities(DMZBoundary) = %v, want %v", caps, expected)
	}
}

func TestOuterCapabilities(t *testing.T) {
	caps := GetBoundaryCapabilities(OuterBoundary)
	expected := Capabilities{
		CanReadSecrets:       false,
		CanWriteSecrets:      false,
		RequiresSanitization: true,
		IngressOnly:          true,
		MailOnlyTransitions:  true,
	}
	if caps != expected {
		t.Errorf("GetBoundaryCapabilities(OuterBoundary) = %v, want %v", caps, expected)
	}
}

func TestTransition_InnerToDMZ(t *testing.T) {
	taints := []string{"INNER_ONLY", "PII", "TOOL_OUTPUT"}
	result, err := EnforceTransition(InnerBoundary, DMZBoundary, taints)
	if err != nil {
		t.Fatalf("EnforceTransition(InnerBoundary, DMZBoundary, %v) returned error: %v", taints, err)
	}
	expected := []string{"TOOL_OUTPUT"}
	if len(result) != len(expected) {
		t.Errorf("EnforceTransition(InnerBoundary, DMZBoundary, %v) = %v, want %v", taints, result, expected)
	}
	for i, v := range expected {
		if result[i] != v {
			t.Errorf("EnforceTransition(InnerBoundary, DMZBoundary, %v)[%d] = %v, want %v", taints, i, result[i], v)
		}
	}
}

func TestTransition_InnerToOuter(t *testing.T) {
	taints := []string{"SECRET", "PII", "TOOL_OUTPUT", "EXTERNAL"}
	result, err := EnforceTransition(InnerBoundary, OuterBoundary, taints)
	if err != nil {
		t.Fatalf("EnforceTransition(InnerBoundary, OuterBoundary, %v) returned error: %v", taints, err)
	}
	expected := []string{"EXTERNAL"}
	if len(result) != len(expected) {
		t.Errorf("EnforceTransition(InnerBoundary, OuterBoundary, %v) = %v, want %v", taints, result, expected)
	}
	for i, v := range expected {
		if result[i] != v {
			t.Errorf("EnforceTransition(InnerBoundary, OuterBoundary, %v)[%d] = %v, want %v", taints, i, result[i], v)
		}
	}
}

func TestTransition_OuterToInner(t *testing.T) {
	taints := []string{"USER_SUPPLIED", "PII"}
	_, err := EnforceTransition(OuterBoundary, InnerBoundary, taints)
	if err == nil {
		t.Fatalf("EnforceTransition(OuterBoundary, InnerBoundary, %v) returned nil error, want error for PII", taints)
	}
	if !strings.Contains(err.Error(), "PII") {
		t.Errorf("EnforceTransition error = %v, want error mentioning PII", err)
	}
}
