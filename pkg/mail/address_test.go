package mail

import (
	"testing"
	"time"
)

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

func TestMail_Metadata(t *testing.T) {
	meta := MailMetadata{
		Tokens:   150,
		Model:    "gpt-4",
		Cost:     0.03,
		Boundary: InnerBoundary,
		Taints:   []string{"USER_SUPPLIED", "TOOL_OUTPUT"},
		Stream:   nil,
		IsFinal:  true,
	}

	if meta.Tokens != 150 {
		t.Errorf("Expected Tokens 150, got %d", meta.Tokens)
	}
	if meta.Model != "gpt-4" {
		t.Errorf("Expected Model 'gpt-4', got '%s'", meta.Model)
	}
	if meta.Boundary != InnerBoundary {
		t.Errorf("Expected Boundary InnerBoundary, got %s", meta.Boundary)
	}

	// Test boundary types
	if InnerBoundary != "inner" {
		t.Errorf("Expected InnerBoundary to be 'inner', got '%s'", InnerBoundary)
	}
	if DMZBoundary != "dmz" {
		t.Errorf("Expected DMZBoundary to be 'dmz', got '%s'", DMZBoundary)
	}
	if OuterBoundary != "outer" {
		t.Errorf("Expected OuterBoundary to be 'outer', got '%s'", OuterBoundary)
	}

	// Test empty taints
	emptyMeta := MailMetadata{Taints: []string{}}
	if len(emptyMeta.Taints) != 0 {
		t.Error("Expected empty Taints slice")
	}
}

func TestMail_Structure(t *testing.T) {
	mail := Mail{
		ID:            "msg-001",
		CorrelationID: "corr-001",
		Type:          MailTypeUser,
		CreatedAt:     time.Now(),
		Source:        "agent:user-agent",
		Target:        "agent:recommendation-agent",
		Content:       map[string]any{"text": "hello"},
		Metadata: MailMetadata{
			Tokens:   10,
			Boundary: OuterBoundary,
			Taints:   []string{"USER_SUPPLIED"},
		},
	}

	if mail.ID != "msg-001" {
		t.Errorf("Expected ID 'msg-001', got '%s'", mail.ID)
	}
	if mail.Type != MailTypeUser {
		t.Errorf("Expected Type MailTypeUser, got %s", mail.Type)
	}
	if mail.Source != "agent:user-agent" {
		t.Errorf("Expected Source 'agent:user-agent', got '%s'", mail.Source)
	}
	if mail.Target != "agent:recommendation-agent" {
		t.Errorf("Expected Target 'agent:recommendation-agent', got '%s'", mail.Target)
	}

	// Test Content accepts any type
	mail.Content = "string content"
	if mail.Content != "string content" {
		t.Error("Expected Content to accept string")
	}
	mail.Content = 42
	if mail.Content != 42 {
		t.Error("Expected Content to accept int")
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
