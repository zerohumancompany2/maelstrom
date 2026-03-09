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
