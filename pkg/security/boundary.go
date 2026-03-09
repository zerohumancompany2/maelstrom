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
