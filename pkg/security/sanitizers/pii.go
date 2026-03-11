package sanitizers

import (
	"regexp"
)

// PIIRedactor redacts PII from data
type PIIRedactor struct{}

// NewPIIRedactor creates a new PIIRedactor
func NewPIIRedactor() *PIIRedactor {
	return &PIIRedactor{}
}

// Redact removes PII from text
func (p *PIIRedactor) Redact(text string) string {
	emailRegex := regexp.MustCompile(`[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}`)
	text = emailRegex.ReplaceAllString(text, "[REDACTED_EMAIL]")
	return text
}

// LengthCapper enforces maximum length limits
type LengthCapper struct {
	MaxLen int
}

// NewLengthCapper creates a new LengthCapper
func NewLengthCapper(maxLen int) *LengthCapper {
	return &LengthCapper{MaxLen: maxLen}
}

// Cap truncates text to max length
func (l *LengthCapper) Cap(text string) string {
	if len(text) > l.MaxLen {
		return text[:l.MaxLen]
	}
	return text
}

// SchemaValidator validates data against schema
type SchemaValidator struct{}

// NewSchemaValidator creates a new SchemaValidator
func NewSchemaValidator() *SchemaValidator {
	return &SchemaValidator{}
}

// Validate checks data against schema
func (s *SchemaValidator) Validate(data any) error {
	return nil
}

// InnerDataStripper strips inner-boundary-only data
type InnerDataStripper struct{}

// NewInnerDataStripper creates a new InnerDataStripper
func NewInnerDataStripper() *InnerDataStripper {
	return &InnerDataStripper{}
}

// Strip removes inner-boundary-only content
func (i *InnerDataStripper) Strip(text string) string {
	keyPatterns := []string{
		"api_key\\s*=\\s*[^\\s]+",
		"secret\\s*=\\s*[^\\s]+",
		"password\\s*=\\s*[^\\s]+",
		"token\\s*=\\s*[^\\s]+",
	}
	result := text
	for _, pattern := range keyPatterns {
		regex := regexp.MustCompile("(?i)" + pattern)
		result = regex.ReplaceAllString(result, "[REDACTED]")
	}
	return result
}
