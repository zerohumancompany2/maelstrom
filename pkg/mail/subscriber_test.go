package mail

import (
	"testing"
)

func TestSubscriber_Receives(t *testing.T) {
	sub := NewChannelSubscriber(10)

	ch := sub.Receive()
	if ch == nil {
		t.Error("Expected non-nil channel")
	}

	// Send mail through internal channel via GetMail verification
	mail := Mail{ID: "msg-001", Type: MailTypeUser}
	sub.ch <- mail

	// Receive mail
	received, ok := sub.GetMail()
	if !ok {
		t.Error("Expected to receive mail")
	}

	if received.ID != "msg-001" {
		t.Errorf("Expected msg-001, got %s", received.ID)
	}
}

func TestPublisher_SubscriberIntegration(t *testing.T) {
	router := NewMailRouter()
	publisher := NewRouterPublisher(router)

	// Setup subscriber
	_ = NewChannelSubscriber(10)
	inbox := &AgentInbox{ID: "test-agent"}
	router.SubscribeAgent("test-agent", inbox)

	// Publish mail
	mail := Mail{
		ID:     "msg-001",
		Source: "agent:user",
		Target: "agent:test-agent",
		Type:   MailTypeUser,
	}

	_, err := publisher.Publish(mail)
	if err != nil {
		t.Errorf("Expected nil error, got %v", err)
	}

	// Verify mail reached inbox
	inbox.mu.RLock()
	if len(inbox.Messages) != 1 {
		t.Errorf("Expected 1 message in inbox, got %d", len(inbox.Messages))
	}
	inbox.mu.RUnlock()
}