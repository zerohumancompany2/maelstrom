package mail

import (
	"testing"
	"time"
)

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

func TestAgentInbox_Subscribe(t *testing.T) {
	inbox := &AgentInbox{ID: "test-agent"}

	mail := Mail{ID: "msg-001", Type: MailTypeUser}
	inbox.Push(mail)

	ch := inbox.Subscribe()
	if ch == nil {
		t.Error("Expected non-nil channel")
	}

	time.Sleep(10 * time.Millisecond)

	select {
	case received := <-ch:
		if received.ID != "msg-001" {
			t.Errorf("Expected msg-001, got %s", received.ID)
		}
	default:
		t.Error("Expected channel to receive message without blocking")
	}
}

func TestServiceInbox_PushPop(t *testing.T) {
	inbox := &ServiceInbox{ID: "test-service"}

	mail := Mail{ID: "msg-001", Type: MailTypeHeartbeat}
	err := inbox.Push(mail)
	if err != nil {
		t.Errorf("Expected nil error, got %v", err)
	}

	popped, err := inbox.Pop()
	if err != nil {
		t.Errorf("Expected nil error, got %v", err)
	}
	if popped.ID != "msg-001" {
		t.Errorf("Expected msg-001, got %s", popped.ID)
	}

	inbox.Push(Mail{ID: "msg-002"})
	inbox.Push(Mail{ID: "msg-003"})

	popped, _ = inbox.Pop()
	if popped.ID != "msg-002" {
		t.Errorf("Expected msg-002, got %s", popped.ID)
	}

	popped, _ = inbox.Pop()
	if popped.ID != "msg-003" {
		t.Errorf("Expected msg-003, got %s", popped.ID)
	}
}
