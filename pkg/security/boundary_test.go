package security

import "testing"

func TestValidateBoundary_Valid(t *testing.T) {
	tests := []BoundaryType{InnerBoundary, DMZBoundary, OuterBoundary}
	for _, boundary := range tests {
		err := ValidateBoundary(boundary)
		if err != nil {
			t.Errorf("ValidateBoundary(%q) = %v, want nil", boundary, err)
		}
	}
}

func TestValidateBoundary_Invalid(t *testing.T) {
	tests := []BoundaryType{"", "public", "unknown"}
	for _, boundary := range tests {
		err := ValidateBoundary(boundary)
		if err == nil {
			t.Errorf("ValidateBoundary(%q) = nil, want error", boundary)
		}
	}
}

func TestInnerCapabilities(t *testing.T) {
	caps := GetBoundaryCapabilities(InnerBoundary)
	expected := Capabilities{
		CanReadSecrets:       true,
		CanWriteSecrets:      true,
		RequiresSanitization: false,
		IngressOnly:          false,
		MailOnlyTransitions:  false,
	}
	if caps != expected {
		t.Errorf("GetBoundaryCapabilities(InnerBoundary) = %v, want %v", caps, expected)
	}
}

func TestDMZCapabilities(t *testing.T) {
	panic("not implemented")
}

func TestOuterCapabilities(t *testing.T) {
	panic("not implemented")
}
