package mail

import "testing"

func TestMailRouter_SubscribeMethods(t *testing.T) {
	router := NewMailRouter()

	// Test SubscribeAgent
	agentInbox := &AgentInbox{ID: "test-agent"}
	err := router.SubscribeAgent("test-agent", agentInbox)
	if err != nil {
		t.Errorf("Expected nil error, got %v", err)
	}

	// Test SubscribeTopic
	topic := &Topic{Name: "test-topic"}
	err = router.SubscribeTopic("test-topic", topic)
	if err != nil {
		t.Errorf("Expected nil error, got %v", err)
	}

	// Test SubscribeService
	serviceInbox := &ServiceInbox{ID: "test-service"}
	err = router.SubscribeService("test-service", serviceInbox)
	if err != nil {
		t.Errorf("Expected nil error, got %v", err)
	}
}

func TestMailRouter_RouteToAgent(t *testing.T) {
	router := NewMailRouter()

	agentInbox := &AgentInbox{ID: "recommendation-agent"}
	router.SubscribeAgent("recommendation-agent", agentInbox)

	mail := Mail{
		ID:     "msg-001",
		Source: "agent:user-agent",
		Target: "agent:recommendation-agent",
		Type:   MailTypeUser,
	}

	err := router.Route(mail)
	if err != nil {
		t.Errorf("Expected nil error, got %v", err)
	}

	// Verify message was pushed to inbox
	agentInbox.mu.RLock()
	if len(agentInbox.Messages) != 1 {
		t.Errorf("Expected 1 message in inbox, got %d", len(agentInbox.Messages))
	}
	agentInbox.mu.RUnlock()
}
