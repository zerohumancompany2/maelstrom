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

func TestMailRouter_RouteToTopic(t *testing.T) {
	router := NewMailRouter()

	topic := &Topic{Name: "market-data"}
	router.SubscribeTopic("market-data", topic)

	mail := Mail{
		ID:     "msg-002",
		Source: "sys:data-service",
		Target: "topic:market-data",
		Type:   MailTypeAssistant,
	}

	err := router.Route(mail)
	if err != nil {
		t.Errorf("Expected nil error, got %v", err)
	}
}

func TestMailRouter_RouteToService(t *testing.T) {
	router := NewMailRouter()

	serviceInbox := &ServiceInbox{ID: "heartbeat"}
	router.SubscribeService("heartbeat", serviceInbox)

	mail := Mail{
		ID:     "msg-003",
		Source: "agent:scheduler",
		Target: "sys:heartbeat",
		Type:   MailTypeHeartbeat,
	}

	err := router.Route(mail)
	if err != nil {
		t.Errorf("Expected nil error, got %v", err)
	}

	// Verify message was pushed to service inbox
	serviceInbox.mu.RLock()
	if len(serviceInbox.Messages) != 1 {
		t.Errorf("Expected 1 message in service inbox, got %d", len(serviceInbox.Messages))
	}
	serviceInbox.mu.RUnlock()
}

func TestMailRouter_RouteToUnknownAddress(t *testing.T) {
	router := NewMailRouter()

	// Test route to unregistered agent
	mail := Mail{
		ID:     "msg-004",
		Source: "agent:user-agent",
		Target: "agent:non-existent",
		Type:   MailTypeUser,
	}

	err := router.Route(mail)
	if err == nil {
		t.Error("Expected error for unregistered agent")
	}

	// Test route to unregistered topic
	mail.Target = "topic:non-existent"
	err = router.Route(mail)
	if err == nil {
		t.Error("Expected error for unregistered topic")
	}

	// Test route to unregistered service
	mail.Target = "sys:non-existent"
	err = router.Route(mail)
	if err == nil {
		t.Error("Expected error for unregistered service")
	}

	// Test route to invalid address format
	mail.Target = "invalid-format"
	err = router.Route(mail)
	if err == nil {
		t.Error("Expected error for invalid address format")
	}
}

func TestParseAddress_agent(t *testing.T) {
	addrType, id, err := ParseAddress("agent:user-123")
	if err != nil {
		t.Errorf("Expected nil error, got %v", err)
	}
	if addrType != AddressTypeAgent {
		t.Errorf("Expected AddressTypeAgent, got %v", addrType)
	}
	if id != "user-123" {
		t.Errorf("Expected id 'user-123', got %v", id)
	}
}

func TestParseAddress_topic(t *testing.T) {
	addrType, id, err := ParseAddress("topic:market-data")
	if err != nil {
		t.Errorf("Expected nil error, got %v", err)
	}
	if addrType != AddressTypeTopic {
		t.Errorf("Expected AddressTypeTopic, got %v", addrType)
	}
	if id != "market-data" {
		t.Errorf("Expected id 'market-data', got %v", id)
	}
}

func TestParseAddress_sys(t *testing.T) {
	addrType, id, err := ParseAddress("sys:heartbeat")
	if err != nil {
		t.Errorf("Expected nil error, got %v", err)
	}
	if addrType != AddressTypeSys {
		t.Errorf("Expected AddressTypeSys, got %v", addrType)
	}
	if id != "heartbeat" {
		t.Errorf("Expected id 'heartbeat', got %v", id)
	}
}
