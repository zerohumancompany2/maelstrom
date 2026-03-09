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
