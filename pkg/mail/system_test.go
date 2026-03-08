package mail

import (
	"sync"
	"testing"
	"time"
)

func TestMailSystem_PublishDeliversMail(t *testing.T) {
	ms := NewMailSystem()

	subCh, err := ms.Subscribe("agent:test")
	if err != nil {
		t.Fatalf("Subscribe failed: %v", err)
	}

	mailMsg := Mail{
		ID:            "mail-1",
		CorrelationID: "corr-1",
		Type:          Heartbeat,
		Source:        "agent:sender",
		Target:        "agent:test",
		Content:       []byte("test content"),
		CreatedAt:     time.Now(),
	}

	ack, err := ms.Publish(mailMsg)
	if err != nil {
		t.Fatalf("Publish failed: %v", err)
	}

	if !ack.Success {
		t.Error("Expected ack.Success to be true")
	}

	select {
	case received := <-subCh:
		if received.Content == nil {
			t.Error("Expected mail content to not be nil")
		}
	case <-time.After(1 * time.Second):
		t.Error("Timeout waiting for mail")
	}
}

func TestMailSystem_SubscribeReceivesMail(t *testing.T) {
	ms := NewMailSystem()

	subCh, err := ms.Subscribe("topic:test")
	if err != nil {
		t.Fatalf("Subscribe failed: %v", err)
	}

	mailMsg := Mail{
		ID:            "mail-2",
		CorrelationID: "corr-2",
		Type:          Heartbeat,
		Source:        "system",
		Target:        "topic:test",
		Content:       []byte("event data"),
		CreatedAt:     time.Now(),
	}

	_, err = ms.Publish(mailMsg)
	if err != nil {
		t.Fatalf("Publish failed: %v", err)
	}

	select {
	case received := <-subCh:
		if received.Target != "topic:test" {
			t.Errorf("Expected mail Target 'topic:test', got %v", received.Target)
		}
	case <-time.After(1 * time.Second):
		t.Error("Timeout waiting for mail")
	}
}

func TestMailSystem_UnsubscribeRemovesSubscriber(t *testing.T) {
	ms := NewMailSystem()

	subCh1, err := ms.Subscribe("agent:test")
	if err != nil {
		t.Fatalf("Subscribe failed: %v", err)
	}

	subCh2, err := ms.Subscribe("agent:test")
	if err != nil {
		t.Fatalf("Subscribe failed: %v", err)
	}

	err = ms.Unsubscribe("agent:test", subCh1)
	if err != nil {
		t.Fatalf("Unsubscribe failed: %v", err)
	}

	mailMsg := Mail{
		ID:            "mail-3",
		CorrelationID: "corr-3",
		Type:          Heartbeat,
		Source:        "system",
		Target:        "agent:test",
		Content:       []byte("notify"),
		CreatedAt:     time.Now(),
	}

	_, err = ms.Publish(mailMsg)
	if err != nil {
		t.Fatalf("Publish failed: %v", err)
	}

	select {
	case <-subCh1:
		t.Error("Unsubscribed channel should not receive mail")
	case <-subCh2:
		// Expected
	case <-time.After(500 * time.Millisecond):
		t.Error("Timeout waiting for mail on subscribed channel")
	}
}

func TestMailSystem_ConcurrentPublish(t *testing.T) {
	ms := NewMailSystem()

	numSubscribers := 10
	numMails := 100
	var wg sync.WaitGroup

	subChs := make([]<-chan Mail, numSubscribers)
	for i := 0; i < numSubscribers; i++ {
		ch, err := ms.Subscribe("agent:concurrent")
		if err != nil {
			t.Fatalf("Subscribe failed: %v", err)
		}
		subChs[i] = ch
	}

	for i := 0; i < numMails; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			mailMsg := Mail{
				ID:            "mail-concurrent-" + string(rune(idx)),
				CorrelationID: "corr-concurrent-" + string(rune(idx)),
				Type:          Heartbeat,
				Source:        "agent:sender",
				Target:        "agent:concurrent",
				Content:       []byte("content-" + string(rune(idx))),
				CreatedAt:     time.Now(),
			}
			_, err := ms.Publish(mailMsg)
			if err != nil {
				t.Errorf("Publish failed: %v", err)
			}
		}(i)
	}

	wg.Wait()

	time.Sleep(100 * time.Millisecond)

	totalReceived := 0
	for _, ch := range subChs {
		select {
		case <-ch:
			totalReceived++
		default:
		}
	}

	if totalReceived < numMails {
		t.Errorf("Expected at least %d mails received, got %d", numMails, totalReceived)
	}
}
