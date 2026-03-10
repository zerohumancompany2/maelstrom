package security

import (
	"testing"
)

func TestSubAgentBoundary_InheritsParent(t *testing.T) {
	// Given: Parent agent at inner boundary spawns sub-agent with inner boundary
	parentBoundary := InnerBoundary
	childBoundary := InnerBoundary

	// When: Sub-agent is created and boundary is validated
	err := ValidateSubAgentBoundary(parentBoundary, childBoundary)

	// Then: Sub-agent inherits parent's inner boundary successfully
	if err != nil {
		t.Errorf("ValidateSubAgentBoundary(InnerBoundary, InnerBoundary) = %v, want nil", err)
	}
}

func TestSubAgentBoundary_CannotElevate(t *testing.T) {
	// Given: Parent agent at dmz boundary attempts to spawn sub-agent with inner boundary
	parentBoundary := DMZBoundary
	childBoundary := InnerBoundary

	// When: Sub-agent creation is attempted with elevation to inner
	result := CheckSubAgentElevation(parentBoundary, childBoundary)

	// Then: Sub-agent creation is rejected with boundary elevation error
	if result {
		t.Errorf("CheckSubAgentElevation(DMZBoundary, InnerBoundary) = true, want false (dmz→inner forbidden)")
	}
}

func TestSubAgentBoundary_ReturnDataTainted(t *testing.T) {
	// Given: Sub-agent at dmz boundary completes and returns data to parent at inner boundary
	subAgentBoundary := DMZBoundary
	data := map[string]interface{}{"result": "test data"}

	// When: Return data is passed through boundary validation
	taintedData, err := TaintSubAgentReturn(data, subAgentBoundary)

	// Then: Data is tainted with DMZ boundary marker before reaching parent
	if err != nil {
		t.Fatalf("TaintSubAgentReturn returned error: %v", err)
	}

	taintedMap, ok := taintedData.(map[string]interface{})
	if !ok {
		t.Errorf("TaintSubAgentReturn returned non-map type: %T", taintedData)
	}

	taints, hasTaints := taintedMap["_taints"].([]string)
	if !hasTaints {
		t.Errorf("Tainted data does not have _taints field")
	}

	found := false
	for _, t := range taints {
		if t == "DMZ" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("Returned data does not carry DMZ taint, got: %v", taints)
	}
}

func TestSubAgentBoundary_ViolationOnElevation(t *testing.T) {
	// Given: Sub-agent at outer boundary attempts to return inner-boundary data to parent
	runtimeId := "test-runtime-123"
	violation := SubAgentViolation{
		RuntimeID:       runtimeId,
		ParentBoundary:  InnerBoundary,
		ChildBoundary:   OuterBoundary,
		ForbiddenTaints: []string{"INNER_ONLY", "SECRET"},
	}

	// When: Return data contains inner-boundary taints not allowed for outer boundary
	err := EmitSubAgentViolation(runtimeId, violation)

	// Then: taint_violation event is emitted to dead-letter queue
	if err != nil {
		t.Errorf("EmitSubAgentViolation returned error: %v", err)
	}

	count := GetViolationCount(runtimeId)
	if count == 0 {
		t.Errorf("Expected violation count > 0, got %d", count)
	}
}

func TestAllowedOnExit_SubAgentReturn_CleanDataPasses(t *testing.T) {
	// Given: A sub-agent returns a result containing only clean (untainted) data
	cleanData := map[string]interface{}{
		"result": "computation output",
		"value":  42,
	}

	// Given: A policy that allows certain taints on exit
	policy := &AllowedOnExitPolicy{
		AllowedOnExit: []string{"TOOL_OUTPUT", "PUBLIC"},
		Enforcement:   EnforcementStrict,
	}

	// When: allowedOnExit check is performed on the sub-agent return value
	result, err := CheckSubAgentReturn(cleanData, policy)

	// Then: Data is returned to parent agent without modification or blocking
	if err != nil {
		t.Fatalf("Expected no error for clean data, got %v", err)
	}

	resultMap, ok := result.(map[string]interface{})
	if !ok {
		t.Fatal("Expected result to be map[string]interface{}")
	}

	if resultMap["result"] != "computation output" {
		t.Errorf("Expected result unchanged, got '%v'", resultMap["result"])
	}

	if resultMap["value"] != 42 {
		t.Errorf("Expected value unchanged, got '%v'", resultMap["value"])
	}
}
