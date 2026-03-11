package gateway

import (
	"fmt"
	"strings"

	"github.com/maelstrom/v3/pkg/mail"
	"github.com/maelstrom/v3/pkg/security"
)

// BorderTainter attaches taints to all inbound data at the border
type BorderTainter struct {
	DefaultTaints []string
}

// TaintInboundData attaches taints to data entering the runtime
func (bt *BorderTainter) TaintInboundData(data any, source string) (any, error) {
	switch v := data.(type) {
	case map[string]any:
		result := make(map[string]any)
		for k, val := range v {
			result[k] = val
		}
		taints := bt.determineTaints(source)
		result["taints"] = taints
		return result, nil
	default:
		return data, nil
	}
}

func (bt *BorderTainter) determineTaints(source string) []string {
	taints := make([]string, 0)

	for _, defaultTaint := range bt.DefaultTaints {
		taints = append(taints, defaultTaint)
	}

	if len(taints) == 0 {
		taints = append(taints, "USER_SUPPLIED", "OUTER_BOUNDARY")
	}

	switch {
	case source == "system:heartbeat":
		taints = append(taints, "SYSTEM_GENERATED")
	case source == "tool:registry":
		taints = append(taints, "TOOL_OUTPUT")
	case len(source) >= 5 && source[:5] == "user:":
		if !taintContains(taints, "USER_SUPPLIED") {
			taints = append(taints, "USER_SUPPLIED")
		}
	}

	return taints
}

func taintContains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// PropagateTaints propagates taints from source mail to target mail
func PropagateTaints(sourceMail *mail.Mail, targetMail *mail.Mail) {
	mail.PropagateTaints(sourceMail, targetMail)
}

// ForbiddenTaintStripper strips forbidden taints before emission
type ForbiddenTaintStripper struct {
	AllowedOnExit map[string]bool
}

// StripForbiddenTaints removes taints not allowed on exit
func (fts *ForbiddenTaintStripper) StripForbiddenTaints(data any, boundary string) (any, error) {
	switch v := data.(type) {
	case map[string]any:
		result := make(map[string]any)
		for k, val := range v {
			result[k] = val
		}

		if taints, ok := v["taints"].([]string); ok {
			allowed := make([]string, 0)
			for _, t := range taints {
				if fts.AllowedOnExit[t] {
					allowed = append(allowed, t)
				}
			}

			if len(allowed) == 0 && len(taints) > 0 {
				return nil, fmt.Errorf("all taints are forbidden for boundary %s", boundary)
			}

			result["taints"] = allowed
		}

		return result, nil
	default:
		return data, nil
	}
}

// BoundaryValidator validates mail on ingress
type BoundaryValidator struct {
	Policy *security.TaintPolicyConfig
}

// ValidateOnIngress validates mail entering the gateway
func (bv *BoundaryValidator) ValidateOnIngress(m *mail.Mail) error {
	if bv.Policy == nil {
		bv.Policy = &security.TaintPolicyConfig{
			Enforcement: security.EnforcementStrict,
		}
	}

	taints := m.GetTaints()
	if len(taints) == 0 {
		return nil
	}

	switch m.Metadata.Boundary {
	case mail.InnerBoundary:
		for _, t := range taints {
			if t == "INNER_ONLY" || t == "SECRET" {
				if strings.Contains(m.Target, "user:") || strings.Contains(m.Target, "outer") {
					return fmt.Errorf("taint_violation: forbidden taint %s on inner→outer transition", t)
				}
			}
		}
	case mail.DMZBoundary:
		for _, t := range taints {
			if t == "INNER_ONLY" || t == "SECRET" {
				return fmt.Errorf("taint_violation: forbidden taint %s on DMZ boundary", t)
			}
		}
	case mail.OuterBoundary:
		for _, t := range taints {
			if t == "PII" {
				return fmt.Errorf("taint_violation: PII not allowed on outer boundary")
			}
		}
	}

	return nil
}

// OuterInnerTaintEnforcer enforces tainting on outer→inner transitions
type OuterInnerTaintEnforcer struct {
	Policy *security.TaintPolicyConfig
}

// EnforceOuterToInner enforces tainting when data moves from outer to inner
func (oit *OuterInnerTaintEnforcer) EnforceOuterToInner(data any, fromBoundary string, toBoundary string) error {
	if oit.Policy == nil {
		oit.Policy = security.NewDefaultSecurityPolicy()
	}

	switch v := data.(type) {
	case map[string]any:
		taints, hasTaints := v["taints"].([]string)
		if !hasTaints {
			if fromBoundary == "outer" && toBoundary == "inner" {
				return fmt.Errorf("data must have taints attached for outer→inner transition")
			}
			return nil
		}

		forbidden := getForbiddenTaintsForTransition(fromBoundary, toBoundary)
		for _, t := range taints {
			for _, f := range forbidden {
				if t == f {
					return fmt.Errorf("taint %s is forbidden for %s→%s transition", t, fromBoundary, toBoundary)
				}
			}
		}

		if toBoundary == "disk" {
			if fromBoundary == "outer" {
				for _, t := range taints {
					if t == "SECRET" || t == "INNER_ONLY" {
						return fmt.Errorf("persistence policy violation: taint %s not allowed on disk from outer", t)
					}
				}
			}
		}

		return nil
	default:
		return nil
	}
}

func getForbiddenTaintsForTransition(from, to string) []string {
	switch {
	case from == "outer" && to == "inner":
		return []string{"PII", "SECRET", "INNER_ONLY"}
	case from == "dmz" && to == "inner":
		return []string{"SECRET"}
	case from == "inner" && to == "outer":
		return []string{"SECRET", "INNER_ONLY"}
	case from == "inner" && to == "dmz":
		return []string{"INNER_ONLY"}
	default:
		return nil
	}
}
