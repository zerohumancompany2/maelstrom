package security

import (
	"testing"
)

func TestStreamSchemaValidation_Validate_ValidChunk(t *testing.T) {
	schema := Schema{
		Fields: map[string]FieldSchema{
			"source": {Type: "string", Required: true},
			"target": {Type: "string", Required: true},
			"data":   {Type: "string", Required: false},
		},
	}

	chunk := map[string]interface{}{
		"source": "agent-1",
		"target": "agent-2",
		"data":   "test content",
	}

	err := Validate(chunk, schema)

	if err != nil {
		t.Errorf("Expected validation to pass, got error: %v", err)
	}

	if chunk["source"] != "agent-1" {
		t.Error("Original chunk was modified")
	}
}

func TestStreamSchemaValidation_Validate_InvalidChunk(t *testing.T) {
	schema := Schema{
		Fields: map[string]FieldSchema{
			"source": {Type: "string", Required: true},
			"target": {Type: "string", Required: true},
			"data":   {Type: "string", Required: false},
		},
	}

	originalChunk := map[string]interface{}{
		"target": "agent-2",
		"data":   12345,
	}

	chunk := make(map[string]interface{})
	for k, v := range originalChunk {
		chunk[k] = v
	}

	err := Validate(chunk, schema)

	if err == nil {
		t.Error("Expected validation to fail, got nil error")
	}

	errors, ok := err.(ValidationErrors)
	if !ok {
		t.Fatalf("Expected ValidationErrors, got %T", err)
	}

	if len(errors) < 2 {
		t.Errorf("Expected at least 2 validation errors, got %d: %v", len(errors), errors)
	}

	foundMissingField := false
	foundTypeError := false

	for _, e := range errors {
		if e.Field == "source" && e.Reason == "missing required field" {
			foundMissingField = true
		}
		if e.Field == "data" && e.Reason == "type mismatch" {
			foundTypeError = true
		}
	}

	if !foundMissingField {
		t.Error("Expected missing field error for 'source'")
	}

	if !foundTypeError {
		t.Error("Expected type error for 'data'")
	}

	if originalChunk["data"] != 12345 {
		t.Error("Original chunk was modified")
	}
}
