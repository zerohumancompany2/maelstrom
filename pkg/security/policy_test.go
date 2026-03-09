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
