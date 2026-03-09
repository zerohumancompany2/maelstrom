package security

import "fmt"

type EnforcementMode string

const (
	EnforcementStrict EnforcementMode = "strict"
	EnforcementRedact EnforcementMode = "redact"
	EnforcementAudit  EnforcementMode = "audit"
)

type TaintPolicyConfig struct {
	Enforcement   EnforcementMode
	AllowedOnExit []string
	RedactRules   []RedactRule
}

func EnforcePolicy(data any, policy TaintPolicyConfig, boundary BoundaryType) (any, error) {
	if policy.Enforcement == EnforcementStrict {
		forbidden := getPolicyForbiddenTaints(data, policy.AllowedOnExit)
		if len(forbidden) > 0 {
			return nil, fmt.Errorf("forbidden taints: %v", forbidden)
		}
	}
	return data, nil
}

func getPolicyForbiddenTaints(data any, allowedOnExit []string) []string {
	allowedSet := make(map[string]bool)
	for _, t := range allowedOnExit {
		allowedSet[t] = true
	}

	switch v := data.(type) {
	case map[string]interface{}:
		if taints, ok := v["_taints"].([]string); ok {
			forbidden := make([]string, 0)
			for _, t := range taints {
				if !allowedSet[t] {
					forbidden = append(forbidden, t)
				}
			}
			return forbidden
		}
	}
	return nil
}
