package communication

import (
	"testing"
	"time"

	"github.com/maelstrom/v3/pkg/mail"
)

func TestCommunicationService_NewCommunicationServiceReturnsNonNil(t *testing.T) {
	svc := NewCommunicationService()

	if svc == nil {
		t.Error("Expected NewCommunicationService to return non-nil")
	}
}

func TestCommunicationService_IDReturnsCorrectString(t *testing.T) {
	svc := NewCommunicationService()

	id := svc.ID()

	if id != "sys:communication" {
		t.Errorf("Expected ID sys:communication, got %s", id)
	}
}

func TestCommunicationService_HandleMailReturnsNil(t *testing.T) {
	svc := NewCommunicationService()

	err := svc.HandleMail(mail.Mail{})

	if err != nil {
		t.Errorf("Expected HandleMail to return nil, got %v", err)
	}
}

func TestCommunicationService_PublishReturnsNil(t *testing.T) {
	svc := NewCommunicationService()

	_, err := svc.Publish(mail.Mail{})

	if err != nil {
		t.Errorf("Expected Publish to return nil, got %v", err)
	}
}

func TestCommunicationService_SubscribeReturnsNonNilChannel(t *testing.T) {
	svc := NewCommunicationService()

	ch, err := svc.Subscribe("agent:test")

	if err != nil {
		t.Errorf("Expected Subscribe to return nil error, got %v", err)
	}

	if ch == nil {
		t.Error("Expected Subscribe to return non-nil channel")
	}
}

func TestCommunicationService_StartReturnsNil(t *testing.T) {
	svc := NewCommunicationService()

	err := svc.Start()

	if err != nil {
		t.Errorf("Expected Start to return nil, got %v", err)
	}
}

func TestCommunicationService_StopReturnsNil(t *testing.T) {
	svc := NewCommunicationService()

	err := svc.Stop()

	if err != nil {
		t.Errorf("Expected Stop to return nil, got %v", err)
	}
}

func TestCommunicationService_BootstrapChart(t *testing.T) {
	chart := BootstrapChart()

	if chart.ID != "sys:communication" {
		t.Errorf("Expected ID sys:communication, got %s", chart.ID)
	}

	if chart.Version != "1.0.0" {
		t.Errorf("Expected version 1.0.0, got %s", chart.Version)
	}
}

func TestCommunicationService_PubSub(t *testing.T) {
	svc := NewCommunicationService()
	ch, err := svc.Subscribe("test-topic")
	if err != nil {
		t.Errorf("Subscribe should return nil error, got: %v", err)
	}
	if ch == nil {
		t.Fatal("Subscribe should return non-nil channel")
	}

	mail := mail.Mail{Source: "test", Target: "test-topic"}
	_, err = svc.Publish(mail)
	if err != nil {
		t.Errorf("Publish should return nil, got: %v", err)
	}

	select {
	case received := <-ch:
		if received.Source != mail.Source {
			t.Errorf("Expected source %s, got %s", mail.Source, received.Source)
		}
	case <-time.After(100 * time.Millisecond):
		t.Error("Timeout waiting for mail")
	}
}

func TestCommunicationService_RoutesMail(t *testing.T) {
	svc := NewCommunicationService()

	agentCh, _ := svc.Subscribe("agent:test-agent")
	topicCh, _ := svc.Subscribe("topic:test-topic")
	sysCh, _ := svc.Subscribe("sys:security")

	_, err := svc.Publish(mail.Mail{Source: "test", Target: "agent:test-agent"})
	if err != nil {
		t.Errorf("Publish to agent failed: %v", err)
	}

	select {
	case <-agentCh:
	case <-time.After(100 * time.Millisecond):
		t.Error("Timeout waiting for agent mail")
	}

	_, err = svc.Publish(mail.Mail{Source: "test", Target: "topic:test-topic"})
	if err != nil {
		t.Errorf("Publish to topic failed: %v", err)
	}

	select {
	case <-topicCh:
	case <-time.After(100 * time.Millisecond):
		t.Error("Timeout waiting for topic mail")
	}

	_, err = svc.Publish(mail.Mail{Source: "test", Target: "sys:security"})
	if err != nil {
		t.Errorf("Publish to sys failed: %v", err)
	}

	select {
	case <-sysCh:
	case <-time.After(100 * time.Millisecond):
		t.Error("Timeout waiting for sys mail")
	}
}

func TestCommunicationService_ID(t *testing.T) {
	chart := BootstrapChart()

	if chart.ID != "sys:communication" {
		t.Errorf("Expected ID sys:communication, got %s", chart.ID)
	}
}

func TestCommunicationService_PublishReturnsAck(t *testing.T) {
	svc := NewCommunicationService()

	ch, _ := svc.Subscribe("test-topic")
	m := mail.Mail{Source: "test", Target: "test-topic"}

	ack, err := svc.Publish(m)

	if err != nil {
		t.Errorf("Expected nil error, got %v", err)
	}
	if ack.MailID != m.ID {
		t.Errorf("Expected MailID %s, got %s", m.ID, ack.MailID)
	}
	if !ack.Success {
		t.Error("Expected Success to be true")
	}
	_ = ch
}

func TestCommunicationService_PublishAckHasCorrelationID(t *testing.T) {
	svc := NewCommunicationService()

	ch, _ := svc.Subscribe("test-topic")
	correlationID := "test-correlation-123"
	m := mail.Mail{Source: "test", Target: "test-topic", CorrelationID: correlationID}

	ack, err := svc.Publish(m)

	if err != nil {
		t.Errorf("Expected nil error, got %v", err)
	}
	if ack.CorrelationID != correlationID {
		t.Errorf("Expected CorrelationID %s, got %s", correlationID, ack.CorrelationID)
	}
	if ack.DeliveredAt.IsZero() {
		t.Error("Expected DeliveredAt to be set")
	}
	_ = ch
}

func TestCommunicationService_PublishToNonExistentAddress(t *testing.T) {
	svc := NewCommunicationService()

	m := mail.Mail{Source: "test", Target: "non-existent:address"}

	ack, err := svc.Publish(m)

	if err != nil {
		t.Errorf("Expected nil error, got %v", err)
	}
	if ack.Success {
		t.Error("Expected Success to be false for non-existent address")
	}
	if ack.ErrorMessage != "no subscribers" {
		t.Errorf("Expected ErrorMessage 'no subscribers', got %s", ack.ErrorMessage)
	}
}
