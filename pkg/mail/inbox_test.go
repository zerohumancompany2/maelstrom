package mail

import "testing"

func TestAgentInbox_PushPop(t *testing.T) {
	inbox := &AgentInbox{ID: "test-agent"}

	// Push first message
	mail1 := Mail{ID: "msg-001", Type: MailTypeUser}
	err := inbox.Push(mail1)
	if err != nil {
		t.Errorf("Expected nil error, got %v", err)
	}

	// Push second message
	mail2 := Mail{ID: "msg-002", Type: MailTypeAssistant}
	err = inbox.Push(mail2)
	if err != nil {
		t.Errorf("Expected nil error, got %v", err)
	}

	// Pop first message (FIFO)
	popped, err := inbox.Pop()
	if err != nil {
		t.Errorf("Expected nil error, got %v", err)
	}
	if popped.ID != "msg-001" {
		t.Errorf("Expected msg-001, got %s", popped.ID)
	}

	// Pop second message
	popped, err = inbox.Pop()
	if err != nil {
		t.Errorf("Expected nil error, got %v", err)
	}
	if popped.ID != "msg-002" {
		t.Errorf("Expected msg-002, got %s", popped.ID)
	}

	// Inbox should be empty now
	_, err = inbox.Pop()
	if err == nil {
		t.Error("Expected error for empty inbox")
	}
}
