package mail

import "testing"

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
