package security

import (
	"fmt"
	"strings"
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

func TestContextBlockSource_SessionLastN(t *testing.T) {
	// Given: A ContextBlock with source: session, strategy: lastN, n: 30 and session with 50 messages
	messages := make([]Message, 50)
	for i := 0; i < 50; i++ {
		messages[i] = Message{
			Role:    "user",
			Content: fmt.Sprintf("|||MSG-%03d|||", i),
		}
	}

	session := &Session{
		Messages: messages,
	}

	block := &ContextBlock{
		Name:     "conversation-history",
		Source:   string(SourceSession),
		Strategy: "lastN",
		N:        30,
	}

	// When: AssembleSource is called on the block
	result, err := AssembleSource(block, session, nil, nil)

	// Then: Block returns last 30 messages from session, in chronological order
	if err != nil {
		t.Fatalf("AssembleSource returned error: %v", err)
	}

	if len(result) == 0 {
		t.Fatal("Expected result to contain messages")
	}

	resultStr := string(result)
	for i := 20; i < 50; i++ {
		expectedMsg := fmt.Sprintf("|||MSG-%03d|||", i)
		if !strings.Contains(resultStr, expectedMsg) {
			t.Errorf("Expected result to contain last 30 messages including %s", expectedMsg)
			break
		}
	}

	for i := 0; i < 20; i++ {
		expectedMsg := fmt.Sprintf("|||MSG-%03d|||", i)
		if strings.Contains(resultStr, expectedMsg) {
			t.Errorf("Expected result to NOT contain first 20 messages including %s", expectedMsg)
			break
		}
	}
}
