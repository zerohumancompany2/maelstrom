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
