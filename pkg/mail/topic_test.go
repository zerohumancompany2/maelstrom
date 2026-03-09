package mail

import "testing"

type mockSubscriber struct {
	ch chan Mail
}

func (m *mockSubscriber) Receive() chan Mail {
	return m.ch
}

func TestTopic_SubscribeUnsubscribe(t *testing.T) {
	topic := &Topic{Name: "test-topic"}

	sub1 := &mockSubscriber{ch: make(chan Mail, 10)}
	sub2 := &mockSubscriber{ch: make(chan Mail, 10)}

	err := topic.Subscribe(sub1)
	if err != nil {
		t.Errorf("Expected nil error, got %v", err)
	}

	err = topic.Subscribe(sub2)
	if err != nil {
		t.Errorf("Expected nil error, got %v", err)
	}

	err = topic.Unsubscribe(sub1)
	if err != nil {
		t.Errorf("Expected nil error, got %v", err)
	}

	topic.mu.RLock()
	if len(topic.Subscribers) != 1 {
		t.Errorf("Expected 1 subscriber, got %d", len(topic.Subscribers))
	}
	if topic.Subscribers[0] != sub2 {
		t.Error("Expected sub2 to remain")
	}
	topic.mu.RUnlock()
}

func TestTopic_Broadcast(t *testing.T) {
	topic := &Topic{Name: "market-data"}

	sub1 := &mockSubscriber{ch: make(chan Mail, 10)}
	sub2 := &mockSubscriber{ch: make(chan Mail, 10)}
	sub3 := &mockSubscriber{ch: make(chan Mail, 10)}

	topic.Subscribe(sub1)
	topic.Subscribe(sub2)
	topic.Subscribe(sub3)

	mail := Mail{ID: "msg-001", Type: MailTypeAssistant}
	err := topic.Publish(mail)
	if err != nil {
		t.Errorf("Expected nil error, got %v", err)
	}

	for i, sub := range []*mockSubscriber{sub1, sub2, sub3} {
		select {
		case received := <-sub.ch:
			if received.ID != "msg-001" {
				t.Errorf("Subscriber %d: Expected msg-001, got %s", i, received.ID)
			}
		default:
			t.Errorf("Subscriber %d: Expected to receive message", i)
		}
	}
}

func TestTopic_UnsubscribeNotFound(t *testing.T) {
	topic := &Topic{Name: "test-topic"}

	unsubscribedSub := &mockSubscriber{ch: make(chan Mail, 10)}

	err := topic.Unsubscribe(unsubscribedSub)
	if err == nil {
		t.Error("Expected error for unsubscribing non-subscriber")
	}

	if err.Error() != "subscriber not found" {
		t.Errorf("Expected 'subscriber not found', got '%s'", err.Error())
	}
}
