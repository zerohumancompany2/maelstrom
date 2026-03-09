package mail

import "testing"

func TestMail_Types(t *testing.T) {
	types := []MailType{
		MailTypeUser,
		MailTypeAssistant,
		MailTypeToolResult,
		MailTypeToolCall,
		MailTypeMailReceived,
		MailTypeHeartbeat,
		MailTypeError,
		MailTypeHumanFeedback,
		MailTypePartialAssistant,
		MailTypeSubagentDone,
		MailTypeTaintViolation,
	}

	if len(types) != 11 {
		t.Errorf("Expected 11 mail types, got %d", len(types))
	}

	// Verify unique values
	seen := make(map[MailType]bool)
	for _, mt := range types {
		if seen[mt] {
			t.Errorf("Duplicate mail type: %s", mt)
		}
		seen[mt] = true
	}

	// Verify specific values
	if MailTypeUser != "user" {
		t.Errorf("Expected MailTypeUser to be 'user', got '%s'", MailTypeUser)
	}
	if MailTypeAssistant != "assistant" {
		t.Errorf("Expected MailTypeAssistant to be 'assistant', got '%s'", MailTypeAssistant)
	}
}

func TestMail_AddressFormats(t *testing.T) {
	// Test agent:<id> format
	if !IsValidAgentAddress("agent:recommendation-agent") {
		t.Error("Expected agent:recommendation-agent to be valid")
	}
	if IsValidAgentAddress("topic:market-data") {
		t.Error("Expected topic:market-data to be invalid for agent address")
	}

	// Test topic:<name> format
	if !IsValidTopicAddress("topic:market-data") {
		t.Error("Expected topic:market-data to be valid")
	}

	// Test sys:<service> format
	if !IsValidSysAddress("sys:heartbeat") {
		t.Error("Expected sys:heartbeat to be valid")
	}
	if !IsValidSysAddress("sys:persistence") {
		t.Error("Expected sys:persistence to be valid")
	}

	// Invalid formats
	if IsValidAgentAddress("invalid-format") {
		t.Error("Expected invalid-format to be rejected")
	}
}
