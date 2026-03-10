package security

import "fmt"

type Schema struct {
	Fields map[string]FieldSchema
}

type FieldSchema struct {
	Type      string
	Required  bool
	MaxLength int
}

type ValidationError struct {
	Field    string
	Reason   string
	Expected string
	Actual   string
}

func (e ValidationError) Error() string {
	return fmt.Sprintf("field %q: %s (expected %s, got %s)", e.Field, e.Reason, e.Expected, e.Actual)
}

type ValidationErrors []ValidationError

func (e ValidationErrors) Error() string {
	if len(e) == 0 {
		return "validation passed"
	}
	msg := "validation failed:\n"
	for _, err := range e {
		msg += fmt.Sprintf("  - %s\n", err.Error())
	}
	return msg
}

func Validate(chunk any, schema Schema) error {
	if chunk == nil {
		return ValidationError{
			Field:    "chunk",
			Reason:   "missing",
			Expected: "non-nil chunk",
			Actual:   "nil",
		}
	}

	chunkMap, ok := chunk.(map[string]interface{})
	if !ok {
		return ValidationError{
			Field:    "chunk",
			Reason:   "invalid type",
			Expected: "map[string]interface{}",
			Actual:   fmt.Sprintf("%T", chunk),
		}
	}

	var errors ValidationErrors

	for fieldName, fieldSchema := range schema.Fields {
		value, exists := chunkMap[fieldName]

		if !exists {
			if fieldSchema.Required {
				errors = append(errors, ValidationError{
					Field:    fieldName,
					Reason:   "missing required field",
					Expected: fieldSchema.Type,
					Actual:   "not present",
				})
			}
			continue
		}

		if fieldSchema.Type != "any" {
			expectedType := getTypeName(fieldSchema.Type)
			actualType := fmt.Sprintf("%T", value)

			if !typeMatches(value, fieldSchema.Type) {
				errors = append(errors, ValidationError{
					Field:    fieldName,
					Reason:   "type mismatch",
					Expected: expectedType,
					Actual:   actualType,
				})
			}
		}

		if fieldSchema.MaxLength > 0 {
			switch v := value.(type) {
			case string:
				if len(v) > fieldSchema.MaxLength {
					errors = append(errors, ValidationError{
						Field:    fieldName,
						Reason:   "exceeds maximum length",
						Expected: fmt.Sprintf("max %d characters", fieldSchema.MaxLength),
						Actual:   fmt.Sprintf("%d characters", len(v)),
					})
				}
			case []interface{}:
				if len(v) > fieldSchema.MaxLength {
					errors = append(errors, ValidationError{
						Field:    fieldName,
						Reason:   "exceeds maximum length",
						Expected: fmt.Sprintf("max %d elements", fieldSchema.MaxLength),
						Actual:   fmt.Sprintf("%d elements", len(v)),
					})
				}
			}
		}
	}

	if len(errors) > 0 {
		return errors
	}

	return nil
}

func getTypeName(typeName string) string {
	switch typeName {
	case "string":
		return "string"
	case "int", "integer":
		return "int"
	case "float", "number":
		return "float64"
	case "bool", "boolean":
		return "bool"
	case "array", "slice":
		return "[]interface{}"
	case "object", "map":
		return "map[string]interface{}"
	default:
		return typeName
	}
}

func typeMatches(value interface{}, expectedType string) bool {
	switch expectedType {
	case "string":
		_, ok := value.(string)
		return ok
	case "int", "integer":
		_, ok := value.(int)
		return ok
	case "float", "number":
		_, ok := value.(float64)
		return ok
	case "bool", "boolean":
		_, ok := value.(bool)
		return ok
	case "array", "slice":
		_, ok := value.([]interface{})
		return ok
	case "object", "map":
		_, ok := value.(map[string]interface{})
		return ok
	case "any":
		return true
	default:
		return false
	}
}
