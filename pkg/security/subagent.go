package security

import "fmt"

type SubAgentBoundary struct {
	ParentBoundary  BoundaryType
	ChildBoundary   BoundaryType
	InheritanceRule string
}

type SubAgentViolation struct {
	RuntimeID       string
	ParentBoundary  BoundaryType
	ChildBoundary   BoundaryType
	ForbiddenTaints []string
}

func ValidateSubAgentBoundary(parentBoundary, childBoundary BoundaryType) error {
	if parentBoundary == childBoundary {
		return nil
	}
	if !isStricterOrEqual(childBoundary, parentBoundary) {
		return fmt.Errorf("sub-agent boundary %q cannot be less strict than parent boundary %q", childBoundary, parentBoundary)
	}
	return nil
}

func TaintSubAgentReturn(data any, subAgentBoundary BoundaryType) (any, error) {
	panic("not implemented")
}

func CheckSubAgentElevation(parentBoundary, childBoundary BoundaryType) bool {
	panic("not implemented")
}

func EmitSubAgentViolation(runtimeId string, violation SubAgentViolation) error {
	panic("not implemented")
}

func getBoundaryLevel(boundary BoundaryType) int {
	switch boundary {
	case InnerBoundary:
		return 1
	case DMZBoundary:
		return 2
	case OuterBoundary:
		return 3
	default:
		return 999
	}
}

func isStricterOrEqual(childBoundary, parentBoundary BoundaryType) bool {
	childLevel := getBoundaryLevel(childBoundary)
	parentLevel := getBoundaryLevel(parentBoundary)
	return childLevel >= parentLevel
}

func createSubAgentTaintMarker(boundary BoundaryType) string {
	panic("not implemented")
}

func formatSubAgentViolationError(parentBoundary, childBoundary BoundaryType) error {
	panic("not implemented")
}
