package security

type Capabilities struct {
	CanReadSecrets       bool
	CanWriteSecrets      bool
	RequiresSanitization bool
	IngressOnly          bool
	MailOnlyTransitions  bool
}

func ValidateBoundary(boundary BoundaryType) error {
	panic("not implemented")
}

func GetBoundaryCapabilities(boundary BoundaryType) Capabilities {
	panic("not implemented")
}
