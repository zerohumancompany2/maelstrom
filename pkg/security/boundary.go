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
	panic("not implemented")
}
