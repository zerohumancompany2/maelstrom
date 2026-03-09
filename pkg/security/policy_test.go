package security

import (
	"strings"
	"testing"
)

func TestPolicyEnforcement_Strict(t *testing.T) {
	// Given: A data object with taints ["SECRET", "PII"]
	data := map[string]interface{}{
		"content": "sensitive data",
		"_taints": []string{"SECRET", "PII"},
	}

	// Given: A TaintPolicy with enforcement: strict and allowedOnExit: ["TOOL_OUTPUT"]
	policy := TaintPolicyConfig{
		Enforcement:   EnforcementStrict,
		AllowedOnExit: []string{"TOOL_OUTPUT"},
	}

	// When: EnforcePolicy is called with the data and policy
	result, err := EnforcePolicy(data, policy, OuterBoundary)

	// Then: An error is returned blocking the data, no data passes through
	if err == nil {
		t.Fatal("Expected error blocking forbidden taints, got nil")
	}

	if result != nil {
		t.Error("Expected nil result when blocked, got non-nil")
	}

	expectedErr := "forbidden taints"
	if !strings.Contains(err.Error(), expectedErr) {
		t.Errorf("Expected error containing '%s', got '%s'", expectedErr, err.Error())
	}
}

func TestPolicyEnforcement_Redact(t *testing.T) {
	// Given: A data object with taints ["SECRET"]
	data := map[string]interface{}{
		"password": "secret123",
		"_taints":  []string{"SECRET"},
	}

	// Given: A TaintPolicy with enforcement: redact and redactRules
	policy := TaintPolicyConfig{
		Enforcement: EnforcementRedact,
		RedactRules: []RedactRule{
			{Taint: "SECRET", Replacement: "[REDACTED]"},
		},
	}

	// When: EnforcePolicy is called with the data and policy
	result, err := EnforcePolicy(data, policy, OuterBoundary)

	// Then: The SECRET-tainted values are replaced with [REDACTED], data passes through
	if err != nil {
		t.Fatalf("Expected no error in redact mode, got %v", err)
	}

	resultMap, ok := result.(map[string]interface{})
	if !ok {
		t.Fatal("Expected result to be map[string]interface{}")
	}

	if resultMap["password"] != "[REDACTED]" {
		t.Errorf("Expected password to be redacted to '[REDACTED]', got '%v'", resultMap["password"])
	}

	if _, ok := resultMap["_taints"]; ok {
		t.Error("Expected _taints to be removed after redaction")
	}
}

func TestPolicyEnforcement_Audit(t *testing.T) {
	// Given: A data object with taints ["PII"]
	data := map[string]interface{}{
		"email":   "user@example.com",
		"_taints": []string{"PII"},
	}

	// Given: A TaintPolicy with enforcement: audit and allowedOnExit: ["TOOL_OUTPUT"]
	policy := TaintPolicyConfig{
		Enforcement:   EnforcementAudit,
		AllowedOnExit: []string{"TOOL_OUTPUT"},
	}

	// When: EnforcePolicy is called with the data and policy
	result, err := EnforcePolicy(data, policy, OuterBoundary)

	// Then: Data passes through unchanged, violation is logged to audit trail
	if err != nil {
		t.Fatalf("Expected no error in audit mode, got %v", err)
	}

	resultMap, ok := result.(map[string]interface{})
	if !ok {
		t.Fatal("Expected result to be map[string]interface{}")
	}

	if resultMap["email"] != "user@example.com" {
		t.Errorf("Expected email unchanged in audit mode, got '%v'", resultMap["email"])
	}

	if !strings.Contains(GetLastAuditLog(), "PII") {
		t.Error("Expected violation to be logged to audit trail")
	}
}

func TestPolicyEnforcement_AllowedOnExit(t *testing.T) {
	ClearAuditLog()

	// Given: A data object with only allowed taint ["TOOL_OUTPUT"]
	data := map[string]interface{}{
		"tool_result": "search results",
		"_taints":     []string{"TOOL_OUTPUT"},
	}

	// Given: A TaintPolicy with allowedOnExit: ["TOOL_OUTPUT"] in redact mode
	policy := TaintPolicyConfig{
		Enforcement:   EnforcementRedact,
		AllowedOnExit: []string{"TOOL_OUTPUT"},
		RedactRules: []RedactRule{
			{Taint: "PII", Replacement: "[PERSONAL_INFO]"},
			{Taint: "SECRET", Replacement: "[REDACTED]"},
		},
	}

	// When: EnforcePolicy is called with the data and policy
	result, err := EnforcePolicy(data, policy, OuterBoundary)

	// Then: allowedOnExit acts as whitelist for outbound taints
	// Fields with only allowed taints pass through unchanged
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	resultMap, ok := result.(map[string]interface{})
	if !ok {
		t.Fatal("Expected result to be map[string]interface{}")
	}

	if resultMap["tool_result"] != "search results" {
		t.Errorf("Expected tool_result unchanged (TOOL_OUTPUT is allowed), got '%v'", resultMap["tool_result"])
	}
}
