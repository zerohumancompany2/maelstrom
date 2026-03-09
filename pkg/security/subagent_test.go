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
