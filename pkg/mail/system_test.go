package mail

import (
	"sync"
	"testing"
	"time"

	"github.com/maelstrom/v3/internal/testutil"
)

func TestMailSystem_PublishDeliversMail(t *testing.T) {
	ms := NewMailSystem()

	subCh, err := ms.Subscribe("agent:test")
	if err != nil {
		t.Fatalf("Subscribe failed: %v", err)
	}

	mail := Mail{
		ID:            "mail-1",
		Type:          MailTypeCommand,
		From:          "agent:sender",
		To:            "agent:test",
		Content:       []byte("test content"),
		CorrelationID: "corr-1",
	}

	ack, err := ms.Publish(mail)
	if err != nil {
		t.Fatalf("Publish failed: %v", err)
	}

	if !ack.Success {
		t.Error("Expected ack.Success to be true")
	}

	received := testutil.MustReceiveMail(t, subCh, 1*time.Second)
	if received.Content == nil || string(received.Content) != "test content" {
		t.Errorf("Expected mail content 'test content', got %v", received.Content)
	}
}

func TestMailSystem_SubscribeReceivesMail(t *testing.T) {
	ms := NewMailSystem()

	subCh, err := ms.Subscribe("topic:test")
	if err != nil {
		t.Fatalf("Subscribe failed: %v", err)
	}

	mail := Mail{
		ID:            "mail-2",
		Type:          MailTypeEvent,
		From:          "system",
		To:            "topic:test",
		Content:       []byte("event data"),
		CorrelationID: "corr-2",
	}

	_, err = ms.Publish(mail)
	if err != nil {
		t.Fatalf("Publish failed: %v", err)
	}

	received := testutil.MustReceiveMail(t, subCh, 1*time.Second)
	if received.To != "topic:test" {
		t.Errorf("Expected mail To 'topic:test', got %v", received.To)
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

	mail := Mail{
		ID:            "mail-3",
		Type:          MailTypeNotification,
		From:          "system",
		To:            "agent:test",
		Content:       []byte("notify"),
		CorrelationID: "corr-3",
	}

	_, err = ms.Publish(mail)
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
			mail := Mail{
				ID:            "mail-concurrent-" + string(rune(idx)),
				Type:          MailTypeCommand,
				From:          "agent:sender",
				To:            "agent:concurrent",
				Content:       []byte("content-" + string(rune(idx))),
				CorrelationID: "corr-concurrent-" + string(rune(idx)),
			}
			_, err := ms.Publish(mail)
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
