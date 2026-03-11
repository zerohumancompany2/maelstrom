package gateway

import (
	"fmt"

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
func (bv *BoundaryValidator) ValidateOnIngress(mail *mail.Mail) error {
	return fmt.Errorf("not implemented")
}

// OuterInnerTaintEnforcer enforces tainting on outer→inner transitions
type OuterInnerTaintEnforcer struct {
	Policy *security.TaintPolicyConfig
}

// EnforceOuterToInner enforces tainting when data moves from outer to inner
func (oit *OuterInnerTaintEnforcer) EnforceOuterToInner(data any, fromBoundary string, toBoundary string) error {
	return fmt.Errorf("not implemented")
}
