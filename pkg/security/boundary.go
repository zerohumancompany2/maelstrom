package security

import "fmt"

type Capabilities struct {
	CanReadSecrets       bool
	CanWriteSecrets      bool
	RequiresSanitization bool
	IngressOnly          bool
	MailOnlyTransitions  bool
}

func ValidateBoundary(boundary BoundaryType) error {
	switch boundary {
	case InnerBoundary, DMZBoundary, OuterBoundary:
		return nil
	default:
		return fmt.Errorf("invalid boundary type: %q", boundary)
	}
}

func GetBoundaryCapabilities(boundary BoundaryType) Capabilities {
	switch boundary {
	case InnerBoundary:
		return Capabilities{
			CanReadSecrets:       true,
			CanWriteSecrets:      true,
			RequiresSanitization: false,
			IngressOnly:          false,
			MailOnlyTransitions:  false,
		}
	case DMZBoundary:
		return Capabilities{
			CanReadSecrets:       false,
			CanWriteSecrets:      false,
			RequiresSanitization: true,
			IngressOnly:          false,
			MailOnlyTransitions:  true,
		}
	case OuterBoundary:
		return Capabilities{
			CanReadSecrets:       false,
			CanWriteSecrets:      false,
			RequiresSanitization: true,
			IngressOnly:          true,
			MailOnlyTransitions:  true,
		}
	default:
		return Capabilities{}
	}
}

type TransitionResult struct {
	AllowedTaints  []string
	StrippedTaints []string
	BlockedTaints  []string
	Error          error
}

func EnforceTransition(source, target BoundaryType, taints []string) ([]string, error) {
	if source == target {
		return taints, nil
	}

	forbidden := getForbiddenTaints(source, target)
	for _, t := range taints {
		for _, f := range forbidden {
			if t == f {
				return nil, fmt.Errorf("taint %q is forbidden for transition %s → %s", t, source, target)
			}
		}
	}

	strip := getAutoStripTaints(source, target)
	stripSet := make(map[string]bool)
	for _, s := range strip {
		stripSet[s] = true
	}

	result := make([]string, 0)
	for _, t := range taints {
		if !stripSet[t] {
			result = append(result, t)
		}
	}

	return result, nil
}

func getForbiddenTaints(source, target BoundaryType) []string {
	switch {
	case source == OuterBoundary && target == InnerBoundary:
		return []string{"PII", "SECRET", "INNER_ONLY"}
	case source == DMZBoundary && target == InnerBoundary:
		return []string{"SECRET"}
	default:
		return nil
	}
}

func getAutoStripTaints(source, target BoundaryType) []string {
	switch {
	case source == InnerBoundary && target == DMZBoundary:
		return []string{"INNER_ONLY", "PII"}
	case source == InnerBoundary && target == OuterBoundary:
		return []string{"SECRET", "PII", "TOOL_OUTPUT", "INNER_ONLY"}
	case source == OuterBoundary && target == InnerBoundary:
		return []string{}
	case source == DMZBoundary && target == InnerBoundary:
		return []string{}
	default:
		return nil
	}
}
