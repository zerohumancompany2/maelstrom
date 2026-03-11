package security

// NewDefaultSecurityPolicy creates a default security policy
func NewDefaultSecurityPolicy() *TaintPolicyConfig {
	return &TaintPolicyConfig{
		Enforcement: EnforcementStrict,
		AllowedOnExit: []string{
			"USER_SUPPLIED",
			"TOOL_OUTPUT",
			"PUBLIC",
		},
	}
}
