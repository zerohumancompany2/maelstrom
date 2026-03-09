package security

import (
	"testing"
)

func TestContextBlockSource_StaticContent(t *testing.T) {
	// Given: A ContextBlock with source: static and content field set to "You are a secure agent"
	block := &ContextBlock{
		Name:    "system-prompt",
		Source:  string(SourceStatic),
		Content: "You are a secure agent",
	}

	// When: AssembleSource is called on the block
	result, err := AssembleSource(block, nil, nil, nil)

	// Then: Block returns the static content exactly as configured
	if err != nil {
		t.Fatalf("AssembleSource returned error: %v", err)
	}

	expected := "You are a secure agent"
	if string(result) != expected {
		t.Errorf("Expected static content to be returned exactly, got: %s", string(result))
	}
}
