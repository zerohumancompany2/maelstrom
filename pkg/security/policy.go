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

var auditLog []string

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
	if policy.Enforcement == EnforcementAudit {
		logAuditViolations(data, policy.AllowedOnExit, boundary)
		return data, nil
	}
	return data, nil
}

func logAuditViolations(data any, allowedOnExit []string, boundary BoundaryType) {
	forbidden := getPolicyForbiddenTaints(data, allowedOnExit)
	if len(forbidden) > 0 {
		auditLog = append(auditLog, fmt.Sprintf("VIOLATION at %s: forbidden taints %v", boundary, forbidden))
	}
}

func GetLastAuditLog() string {
	if len(auditLog) == 0 {
		return ""
	}
	return auditLog[len(auditLog)-1]
}

func ClearAuditLog() {
	auditLog = []string{}
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

	redactRules := make(map[string]string)
	for _, rule := range rules {
		redactRules[rule.Taint] = rule.Replacement
	}

	result := make(map[string]interface{})
	for k, val := range m {
		if k == "_taints" {
			continue
		}
		result[k] = applyFieldRedaction(val, taints, redactRules)
	}
	return result, nil
}

func applyFieldRedaction(val any, taints []string, redactRules map[string]string) any {
	switch v := val.(type) {
	case string:
		replacement := getMostRestrictiveRedaction(taints, redactRules)
		if replacement != "" {
			return replacement
		}
		return v
	default:
		return val
	}
}

func getMostRestrictiveRedaction(taints []string, redactRules map[string]string) string {
	for _, t := range taints {
		if replacement, ok := redactRules[t]; ok {
			return replacement
		}
	}
	return ""
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

type AllowedOnExitPolicy struct {
	AllowedOnExit []string
	Enforcement   EnforcementMode
	RedactRules   []RedactRule
}

func CheckSubAgentReturn(result any, policy *AllowedOnExitPolicy) (any, error) {
	if policy == nil {
		return result, nil
	}

	switch v := result.(type) {
	case map[string]interface{}:
		_, hasTaints := v["_taints"].([]string)
		if !hasTaints {
			return result, nil
		}

		forbidden := getPolicyForbiddenTaints(result, policy.AllowedOnExit)
		if len(forbidden) == 0 {
			return result, nil
		}

		if policy.Enforcement == EnforcementStrict {
			return nil, fmt.Errorf("forbidden taints in sub-agent return: %v", forbidden)
		}

		if policy.Enforcement == EnforcementRedact {
			return applyRedaction(result, policy.RedactRules)
		}

		if policy.Enforcement == EnforcementAudit {
			logAuditViolations(result, policy.AllowedOnExit, DMZBoundary)
			return result, nil
		}
		return result, nil
	default:
		return result, nil
	}
}
