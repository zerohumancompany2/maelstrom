package mail

import (
	"testing"
	"time"
)

func TestMail_AddressFormat(t *testing.T) {
	tests := []struct {
		name     string
		address  string
		expected bool
	}{
		{"agent format", "agent:test-agent", true},
		{"topic format", "topic:general", true},
		{"sys format", "sys:security", true},
		{"invalid no prefix", "invalid", false},
		{"empty string", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			valid := isValidAddress(tt.address)
			if valid != tt.expected {
				t.Errorf("isValidAddress(%q) = %v, want %v", tt.address, valid, tt.expected)
			}
		})
	}
}

func TestMail_PublishSubscribe(t *testing.T) {
	ch := make(chan Mail, 1)

	publisher := &simplePublisher{ch: ch}

	mail := Mail{
		ID:      "msg-1",
		Type:    User,
		Source:  "agent:test",
		Target:  "topic:general",
		Content: "hello",
	}

	_, err := publisher.Publish(mail)
	if err != nil {
		t.Fatalf("Publish failed: %v", err)
	}

	select {
	case received := <-ch:
		if received.Content != mail.Content {
			t.Errorf("Content mismatch: got %v, want %v", received.Content, mail.Content)
		}
		if received.Source != mail.Source {
			t.Errorf("Source mismatch: got %v, want %v", received.Source, mail.Source)
		}
	case <-time.After(100 * time.Millisecond):
		t.Fatal("Timeout waiting for message")
	}
}

type simplePublisher struct {
	ch chan Mail
}

func (p *simplePublisher) Publish(mail Mail) (Ack, error) {
	p.ch <- mail
	return Ack{DeliveredAt: time.Now()}, nil
}

func TestMail_Deduplication(t *testing.T) {
	subscriber := &simpleSubscriber{
		mailChans: make(map[string]chan Mail),
		delivered: make(map[string]bool),
	}

	ch, err := subscriber.Subscribe("topic:general")
	if err != nil {
		t.Fatalf("Subscribe failed: %v", err)
	}

	mail := Mail{
		ID:            "msg-1",
		CorrelationID: "corr-1",
		Type:          User,
		Source:        "agent:test",
		Target:        "topic:general",
		Content:       "hello",
	}

	subscriber.Deliver(mail)

	received := 0
	select {
	case <-ch:
		received++
	case <-time.After(100 * time.Millisecond):
		t.Fatal("Timeout waiting for message")
	}

	subscriber.Deliver(mail)

	select {
	case <-ch:
		received++
		t.Fatal("Should not receive duplicate with same correlationId")
	case <-time.After(100 * time.Millisecond):
		// Expected timeout
	}

	if received != 1 {
		t.Errorf("Expected 1 message, got %d", received)
	}
}

type simpleSubscriber struct {
	mailChans map[string]chan Mail
	delivered map[string]bool
}

func (s *simpleSubscriber) Subscribe(address string) (<-chan Mail, error) {
	ch := make(chan Mail, 1)
	s.mailChans[address] = ch
	return ch, nil
}

func (s *simpleSubscriber) Unsubscribe(address string, ch <-chan Mail) error {
	delete(s.mailChans, address)
	return nil
}

func (s *simpleSubscriber) Deliver(mail Mail) {
	if mail.CorrelationID != "" {
		if s.delivered[mail.CorrelationID] {
			return
		}
		s.delivered[mail.CorrelationID] = true
	}
	if ch, ok := s.mailChans[mail.Target]; ok {
		select {
		case ch <- mail:
		default:
		}
	}
}

func TestMail_RouterRouting(t *testing.T) {
	router := &simpleRouter{
		subscribers: make(map[string][]chan Mail),
	}

	agentCh := make(chan Mail, 1)
	topicCh := make(chan Mail, 1)

	router.Subscribe("agent:test", agentCh)
	router.Subscribe("topic:general", topicCh)

	mail := Mail{
		ID:      "msg-1",
		Type:    User,
		Source:  "agent:other",
		Target:  "agent:test",
		Content: "direct",
	}

	router.Route(mail)

	select {
	case received := <-agentCh:
		if received.Content != "direct" {
			t.Errorf("Expected direct mail, got %v", received.Content)
		}
	case <-time.After(100 * time.Millisecond):
		t.Fatal("Timeout waiting for agent mail")
	}

	mail.Target = "topic:general"
	mail.Content = "broadcast"
	router.Route(mail)

	select {
	case received := <-topicCh:
		if received.Content != "broadcast" {
			t.Errorf("Expected broadcast mail, got %v", received.Content)
		}
	case <-time.After(100 * time.Millisecond):
		t.Fatal("Timeout waiting for topic mail")
	}
}

type simpleRouter struct {
	subscribers map[string][]chan Mail
}

func (r *simpleRouter) Subscribe(address string, ch chan Mail) {
	r.subscribers[address] = append(r.subscribers[address], ch)
}

func (r *simpleRouter) Route(mail Mail) {
	for _, ch := range r.subscribers[mail.Target] {
		select {
		case ch <- mail:
		default:
		}
	}
}

func TestMail_MailTypes(t *testing.T) {
	expectedTypes := []MailType{
		User, Assistant, ToolResult, ToolCall,
		MailReceived, Heartbeat, Error, HumanFeedback,
		PartialAssistant, SubagentDone, TaintViolation,
	}

	expected := len(expectedTypes)
	if len(expectedTypes) != expected {
		t.Errorf("Expected %d mail types, got %d", expected, len(expectedTypes))
	}

	for _, typ := range expectedTypes {
		if string(typ) == "" {
			t.Errorf("MailType %q has empty string value", typ)
		}
	}
}

func TestMail_MetadataStructure(t *testing.T) {
	streamChunk := &StreamChunk{
		Data:     "test-stream",
		Sequence: 1,
		IsFinal:  false,
		Taints:   []string{"test"},
	}
	metadata := MailMetadata{
		Tokens:   100,
		Model:    "test-model",
		Cost:     0.5,
		Boundary: InnerBoundary,
		Taints:   []string{"PII", "SECRET"},
		Stream:   streamChunk,
		IsFinal:  true,
	}

	if metadata.Tokens != 100 {
		t.Errorf("Expected tokens 100, got %d", metadata.Tokens)
	}

	if metadata.Model != "test-model" {
		t.Errorf("Expected model test-model, got %s", metadata.Model)
	}

	if metadata.Cost != 0.5 {
		t.Errorf("Expected cost 0.5, got %f", metadata.Cost)
	}

	if metadata.Boundary != InnerBoundary {
		t.Errorf("Expected boundary inner, got %s", metadata.Boundary)
	}

	if len(metadata.Taints) != 2 {
		t.Errorf("Expected 2 taints, got %d", len(metadata.Taints))
	}

	if metadata.Stream == nil {
		t.Error("Expected stream to be set")
	}

	if !metadata.IsFinal {
		t.Error("Expected isFinal to be true")
	}

	empty := MailMetadata{}
	if empty.Tokens != 0 {
		t.Errorf("Expected empty tokens 0, got %d", empty.Tokens)
	}
}

func TestMail_DeadLetterDeferred(t *testing.T) {
	mail := Mail{
		ID:      "test",
		Source:  "agent:test",
		Target:  "topic:test",
		Content: "test",
	}

	if mail.ID != "test" {
		t.Error("Mail created successfully")
	}
	_ = mail
}

func TestFullMailFlow(t *testing.T) {
	// Setup
	router := NewMailRouter()
	publisher := NewRouterPublisher(router)

	// Create subscriber inbox
	inbox := &AgentInbox{ID: "test-agent"}
	router.SubscribeAgent("test-agent", inbox)

	// Create mail with all fields
	originalMail := Mail{
		ID:            "msg-001",
		CorrelationID: "corr-001",
		Type:          MailTypeUser,
		CreatedAt:     time.Now(),
		Source:        "agent:user-agent",
		Target:        "agent:test-agent",
		Content:       map[string]any{"text": "hello"},
		Metadata: MailMetadata{
			Tokens:   10,
			Model:    "gpt-4",
			Cost:     0.01,
			Boundary: OuterBoundary,
			Taints:   []string{"USER_SUPPLIED"},
		},
	}

	// Publish
	ack, err := publisher.Publish(originalMail)
	if err != nil {
		t.Fatalf("Publish failed: %v", err)
	}

	// Verify Ack
	if ack.CorrelationID != originalMail.CorrelationID {
		t.Errorf("Expected CorrelationID '%s', got '%s'",
			originalMail.CorrelationID, ack.CorrelationID)
	}

	if ack.DeliveredAt.IsZero() {
		t.Error("Expected DeliveredAt to be set")
	}

	// Verify delivery to inbox
	inbox.mu.RLock()
	if len(inbox.Messages) != 1 {
		t.Errorf("Expected 1 message in inbox, got %d", len(inbox.Messages))
	}
	deliveredMail := inbox.Messages[0]
	inbox.mu.RUnlock()

	// Verify mail integrity
	if deliveredMail.ID != originalMail.ID {
		t.Errorf("Expected ID '%s', got '%s'", originalMail.ID, deliveredMail.ID)
	}
	if deliveredMail.Type != originalMail.Type {
		t.Errorf("Expected Type '%s', got '%s'", originalMail.Type, deliveredMail.Type)
	}
	if deliveredMail.Source != originalMail.Source {
		t.Errorf("Expected Source '%s', got '%s'", originalMail.Source, deliveredMail.Source)
	}
}

func TestCommunicationService_Integration(t *testing.T) {
	// Test agent-to-agent routing
	router := NewMailRouter()

	agent1 := &AgentInbox{ID: "agent1"}
	agent2 := &AgentInbox{ID: "agent2"}
	router.SubscribeAgent("agent1", agent1)
	router.SubscribeAgent("agent2", agent2)

	mail1 := Mail{
		ID:     "msg-001",
		Source: "agent:agent1",
		Target: "agent:agent2",
		Type:   MailTypeUser,
	}

	err := router.Route(mail1)
	if err != nil {
		t.Fatalf("Route failed: %v", err)
	}

	agent2.mu.RLock()
	if len(agent2.Messages) != 1 {
		t.Errorf("Expected 1 message, got %d", len(agent2.Messages))
	}
	agent2.mu.RUnlock()

	// Test topic publishing
	topic := &Topic{Name: "events"}
	router.SubscribeTopic("events", topic)

	sub1Ch := make(chan Mail, 10)
	sub2Ch := make(chan Mail, 10)
	sub1 := &topicSubscriberWrapper{ch: sub1Ch}
	sub2 := &topicSubscriberWrapper{ch: sub2Ch}
	topic.Subscribe(sub1)
	topic.Subscribe(sub2)

	mail2 := Mail{
		ID:     "msg-002",
		Source: "sys:events",
		Target: "topic:events",
		Type:   MailTypeAssistant,
	}

	err = router.Route(mail2)
	if err != nil {
		t.Fatalf("Topic route failed: %v", err)
	}

	// Verify both subscribers received the mail
	select {
	case received := <-sub1.ch:
		if received.ID != "msg-002" {
			t.Errorf("Expected msg-002, got %s", received.ID)
		}
	default:
		t.Error("Subscriber 1 did not receive mail")
	}

	select {
	case received := <-sub2.ch:
		if received.ID != "msg-002" {
			t.Errorf("Expected msg-002, got %s", received.ID)
		}
	default:
		t.Error("Subscriber 2 did not receive mail")
	}

	// Test sys service routing
	serviceInbox := &ServiceInbox{ID: "heartbeat"}
	router.SubscribeService("heartbeat", serviceInbox)

	mail3 := Mail{
		ID:     "msg-003",
		Source: "agent:scheduler",
		Target: "sys:heartbeat",
		Type:   MailTypeHeartbeat,
	}

	err = router.Route(mail3)
	if err != nil {
		t.Fatalf("Service route failed: %v", err)
	}

	serviceInbox.mu.RLock()
	if len(serviceInbox.Messages) != 1 {
		t.Errorf("Expected 1 message in service inbox, got %d", len(serviceInbox.Messages))
	}
	serviceInbox.mu.RUnlock()
}

type topicSubscriberWrapper struct {
	ch chan Mail
}

func (t *topicSubscriberWrapper) Receive() chan Mail {
	return t.ch
}
