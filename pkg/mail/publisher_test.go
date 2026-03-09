package mail

import (
	"testing"
)

func TestPublisher_Publish(t *testing.T) {
	router := NewMailRouter()
	publisher := NewRouterPublisher(router)

	// Setup: register an agent
	inbox := &AgentInbox{ID: "test-agent"}
	router.SubscribeAgent("test-agent", inbox)

	// Publish mail
	mail := Mail{
		ID:     "msg-001",
		Source: "agent:user",
		Target: "agent:test-agent",
		Type:   MailTypeUser,
	}

	ack, err := publisher.Publish(mail)
	if err != nil {
		t.Errorf("Expected nil error, got %v", err)
	}

	if ack.CorrelationID != mail.CorrelationID {
		t.Errorf("Expected CorrelationID to match, got %s", ack.CorrelationID)
	}

	if ack.DeliveredAt.IsZero() {
		t.Error("Expected DeliveredAt to be set")
	}
}

func TestPublisherInterface_ImplementedByRouter(t *testing.T) {
	router := NewMailRouter()
	publisher := NewRouterPublisher(router)

	// Verify RouterPublisher implements Publisher interface
	var _ Publisher = publisher

	// Use as Publisher
	var p Publisher = publisher
	if p == nil {
		t.Error("Expected non-nil Publisher")
	}
}