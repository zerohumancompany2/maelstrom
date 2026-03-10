package security

import "github.com/maelstrom/v3/pkg/mail"

func SanitizeContextMap(ctx ContextMap, boundary BoundaryType) (ContextMap, error) {
	forbiddenTaints := getHumanGatewayForbiddenTaints(boundary)
	redactTaints := getHumanGatewayRedactTaints(boundary)

	filtered := make([]*ContextBlock, 0)
	for _, block := range ctx.Blocks {
		shouldStrip := false
		shouldRedact := false

		for taint := range block.Taints {
			if contains(forbiddenTaints, taint) {
				shouldStrip = true
				break
			}
			if contains(redactTaints, taint) {
				shouldRedact = true
			}
		}

		if shouldStrip {
			continue
		}

		result := *block
		if shouldRedact {
			result.Content = "[REDACTED]"
		}
		filtered = append(filtered, &result)
	}

	return ContextMap{
		Blocks:     filtered,
		TokenCount: ctx.TokenCount,
		Budget:     ctx.Budget,
	}, nil
}

func getHumanGatewayForbiddenTaints(boundary BoundaryType) []string {
	switch boundary {
	case OuterBoundary:
		return []string{"SECRET", "INNER_ONLY", "PII"}
	case DMZBoundary:
		return []string{"SECRET", "INNER_ONLY"}
	default:
		return nil
	}
}

func getHumanGatewayRedactTaints(boundary BoundaryType) []string {
	switch boundary {
	case OuterBoundary:
		return []string{"PII"}
	case DMZBoundary:
		return []string{"PII"}
	default:
		return nil
	}
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func SanitizeMessageHistory(messages []mail.Mail, boundary BoundaryType) ([]mail.Mail, error) {
	panic("not implemented")
}
