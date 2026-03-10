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
