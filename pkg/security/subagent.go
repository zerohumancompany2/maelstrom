package security

import (
	"fmt"
	"time"
)

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
	marker := createSubAgentTaintMarker(subAgentBoundary)
	switch v := data.(type) {
	case map[string]interface{}:
		existingTaints, _ := v["_taints"].([]string)
		taints := make([]string, len(existingTaints)+1)
		copy(taints, existingTaints)
		taints = append(taints, marker)
		v["_taints"] = taints
		return v, nil
	default:
		return data, nil
	}
}

func CheckSubAgentElevation(parentBoundary, childBoundary BoundaryType) bool {
	parentLevel := getBoundaryLevel(parentBoundary)
	childLevel := getBoundaryLevel(childBoundary)
	return childLevel > parentLevel
}

func EmitSubAgentViolation(runtimeId string, violation SubAgentViolation) error {
	taintViolation := TaintViolation{
		RuntimeID:       runtimeId,
		SourceBoundary:  violation.ChildBoundary,
		TargetBoundary:  violation.ParentBoundary,
		ForbiddenTaints: violation.ForbiddenTaints,
		Timestamp:       time.Now(),
	}
	return ReportViolation(runtimeId, taintViolation)
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
	switch boundary {
	case InnerBoundary:
		return "INNER"
	case DMZBoundary:
		return "DMZ"
	case OuterBoundary:
		return "OUTER"
	default:
		return string(boundary)
	}
}

func formatSubAgentViolationError(parentBoundary, childBoundary BoundaryType) error {
	panic("not implemented")
}
