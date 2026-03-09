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
	if policy.Enforcement == EnforcementRedact {
		return applyRedaction(data, policy.RedactRules)
	}
	return data, nil
}

func applyRedaction(data any, rules []RedactRule) (any, error) {
	switch v := data.(type) {
	case map[string]interface{}:
		return redactMap(v, rules)
	}
	return data, nil
}

func redactMap(m map[string]interface{}, rules []RedactRule) (map[string]interface{}, error) {
	taints, hasTaints := m["_taints"].([]string)
	if !hasTaints {
		return m, nil
	}

	redactMap := make(map[string]string)
	for _, rule := range rules {
		redactMap[rule.Taint] = rule.Replacement
	}

	result := make(map[string]interface{})
	for k, val := range m {
		if k == "_taints" {
			continue
		}
		shouldRedact := false
		for _, t := range taints {
			if _, ok := redactMap[t]; ok {
				shouldRedact = true
				break
			}
		}
		if shouldRedact {
			switch val.(type) {
			case string:
				result[k] = getRedactionForTaint(taints, redactMap)
			default:
				result[k] = val
			}
		} else {
			result[k] = val
		}
	}
	return result, nil
}

func getRedactionForTaint(taints []string, redactMap map[string]string) string {
	for _, t := range taints {
		if replacement, ok := redactMap[t]; ok {
			return replacement
		}
	}
	return "[REDACTED]"
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
