package communication

import (
	"fmt"
	"strings"
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

func TestCommunicationService_UnsubscribeRemovesSubscriber(t *testing.T) {
	svc := NewCommunicationService()

	ch, err := svc.Subscribe("test-topic")
	if err != nil {
		t.Fatalf("Subscribe failed: %v", err)
	}

	m := mail.Mail{Source: "test", Target: "test-topic"}
	_, err = svc.Publish(m)
	if err != nil {
		t.Fatalf("Publish failed: %v", err)
	}

	select {
	case <-ch:
	case <-time.After(100 * time.Millisecond):
		t.Fatal("Timeout waiting for mail before unsubscribe")
	}

	for len(ch) > 0 {
		select {
		case <-ch:
		default:
			goto unsubscribed
		}
	}
unsubscribed:
	err = svc.Unsubscribe("test-topic", ch)
	if err != nil {
		t.Errorf("Unsubscribe should return nil error, got %v", err)
	}

	_, err = svc.Publish(m)
	if err != nil {
		t.Fatalf("Publish failed: %v", err)
	}

	select {
	case _, ok := <-ch:
		if ok {
			t.Error("Should not receive mail after unsubscribe")
		}
	case <-time.After(50 * time.Millisecond):
	}
}

func TestCommunicationService_UnsubscribeNotFoundReturnsError(t *testing.T) {
	svc := NewCommunicationService()

	ch := make(chan mail.Mail)

	err := svc.Unsubscribe("non-existent", ch)

	if err == nil {
		t.Error("Expected error for non-existent address, got nil")
	}
	if !strings.Contains(err.Error(), "no subscribers") {
		t.Errorf("Expected error mentioning 'no subscribers', got %v", err)
	}
}

func TestCommunicationService_RetryOnFailure(t *testing.T) {
	svc := NewCommunicationService()

	m := mail.Mail{Source: "test", Target: "non-existent:address"}

	err := svc.PublishWithRetry(&m, 3)

	if err == nil {
		t.Error("Expected error after retries, got nil")
	}
}

func TestCommunicationService_ExponentialBackoff(t *testing.T) {
	svc := NewCommunicationService()

	start := time.Now()
	m := mail.Mail{Source: "test", Target: "non-existent:address"}

	err := svc.PublishWithRetry(&m, 3)
	elapsed := time.Since(start)

	if err == nil {
		t.Error("Expected error after retries, got nil")
	}
	if elapsed < 1*time.Second {
		t.Errorf("Expected exponential backoff delays, got %v", elapsed)
	}
}

func TestCommunicationService_MaxRespectsLimit(t *testing.T) {
	svc := NewCommunicationService()

	start := time.Now()
	m := mail.Mail{Source: "test", Target: "non-existent:address"}

	err := svc.PublishWithRetry(&m, 1)
	elapsed := time.Since(start)

	if err == nil {
		t.Error("Expected error after max retries, got nil")
	}
	if elapsed > 2*time.Second {
		t.Errorf("Expected to respect max retries limit, got %v elapsed", elapsed)
	}
}

func TestCommunicationService_DeliveryTracking(t *testing.T) {
	svc := NewCommunicationService()

	correlationID := "test-tracking-123"
	m := mail.Mail{Source: "test", Target: "non-existent:address", CorrelationID: correlationID}

	svc.PublishWithRetry(&m, 2)

	svc.trackDeliveryAttempt(correlationID)
	svc.trackDeliveryAttempt(correlationID)
	svc.trackDeliveryAttempt(correlationID)
}

func TestCommunicationService_RequestReply(t *testing.T) {
	svc := NewCommunicationService()

	mailChan, err := svc.Subscribe("agent:reply")
	if err != nil {
		t.Fatalf("Subscribe failed: %v", err)
	}

	replyChan := make(chan *mail.Mail)
	go func() {
		for m := range mailChan {
			replyChan <- &m
		}
	}()

	correlationID := "test-correlation-456"
	requestMail := mail.Mail{
		ID:            "request-1",
		CorrelationID: correlationID,
		Source:        "agent:requester",
		Target:        "agent:handler",
		Type:          mail.MailTypeUser,
		Content:       "test request",
	}

	go func() {
		time.Sleep(10 * time.Millisecond)
		replyMail := mail.Mail{
			ID:            "reply-1",
			CorrelationID: correlationID,
			Source:        "agent:handler",
			Target:        "agent:reply",
			Type:          mail.MailTypeAssistant,
			Content:       "test reply",
		}
		_, _ = svc.Publish(replyMail)
	}()

	reply, err := svc.Request(replyChan, 1*time.Second)

	if err != nil {
		t.Errorf("Request failed: %v", err)
	}
	if reply == nil {
		t.Error("Expected non-nil reply")
	}
	if reply.CorrelationID != correlationID {
		t.Errorf("Expected CorrelationID %s, got %s", correlationID, reply.CorrelationID)
	}
	if reply.Content != "test reply" {
		t.Errorf("Expected content 'test reply', got %v", reply.Content)
	}
	_ = requestMail
}

func TestCommunicationService_CorrelationIdMatching(t *testing.T) {
	svc := NewCommunicationService()

	correlationID := "corr-match-789"
	otherCorrelationID := "other-corr-abc"

	matchingMail := &mail.Mail{
		ID:            "mail-1",
		CorrelationID: correlationID,
		Source:        "agent:sender",
		Target:        "agent:receiver",
		Type:          mail.MailTypeUser,
		Content:       "matching content",
	}

	nonMatchingMail := &mail.Mail{
		ID:            "mail-2",
		CorrelationID: otherCorrelationID,
		Source:        "agent:sender",
		Target:        "agent:receiver",
		Type:          mail.MailTypeUser,
		Content:       "non-matching content",
	}

	matchResult := svc.matchReply(correlationID, matchingMail)
	if !matchResult {
		t.Error("Expected match for matching correlationId")
	}

	nonMatchResult := svc.matchReply(correlationID, nonMatchingMail)
	if nonMatchResult {
		t.Error("Expected no match for different correlationId")
	}

	emptyCorrelationMail := &mail.Mail{
		ID:            "mail-3",
		CorrelationID: "",
		Source:        "agent:sender",
		Target:        "agent:receiver",
		Type:          mail.MailTypeUser,
		Content:       "empty correlation content",
	}

	emptyMatchResult := svc.matchReply("", emptyCorrelationMail)
	if !emptyMatchResult {
		t.Error("Expected match for empty correlationId")
	}
}

func TestCommunicationService_RequestTimeout(t *testing.T) {
	svc := NewCommunicationService()

	mailChan, err := svc.Subscribe("agent:timeout-test")
	if err != nil {
		t.Fatalf("Subscribe failed: %v", err)
	}

	replyChan := make(chan *mail.Mail)
	go func() {
		for m := range mailChan {
			replyChan <- &m
		}
	}()

	_, err = svc.Request(replyChan, 50*time.Millisecond)

	if err == nil {
		t.Error("Expected timeout error, got nil")
	}
	if err.Error() != "request timeout" {
		t.Errorf("Expected 'request timeout' error, got %v", err)
	}
}

func TestCommunicationService_MultipleRequests(t *testing.T) {
	svc := NewCommunicationService()

	numRequests := 5
	results := make(chan string, numRequests)

	for i := 0; i < numRequests; i++ {
		correlationID := fmt.Sprintf("multi-corr-%d", i)

		mailChan, err := svc.Subscribe(fmt.Sprintf("agent:multi-reply-%d", i))
		if err != nil {
			t.Fatalf("Subscribe failed: %v", err)
		}

		replyChan := make(chan *mail.Mail)
		go func() {
			for m := range mailChan {
				replyChan <- &m
			}
		}()

		go func() {
			time.Sleep(10 * time.Millisecond)
			replyMail := mail.Mail{
				ID:            fmt.Sprintf("reply-%d", i),
				CorrelationID: correlationID,
				Source:        "agent:handler",
				Target:        fmt.Sprintf("agent:multi-reply-%d", i),
				Type:          mail.MailTypeAssistant,
				Content:       fmt.Sprintf("reply-%d", i),
			}
			_, _ = svc.Publish(replyMail)
		}()

		go func() {
			reply, err := svc.Request(replyChan, 1*time.Second)
			if err != nil {
				results <- fmt.Sprintf("error-%d", i)
				return
			}
			if reply.CorrelationID != correlationID {
				results <- fmt.Sprintf("wrong-corr-%d", i)
				return
			}
			results <- fmt.Sprintf("success-%d", i)
		}()
	}

	successCount := 0
	for i := 0; i < numRequests; i++ {
		result := <-results
		if strings.HasPrefix(result, "success-") {
			successCount++
		}
	}

	if successCount != numRequests {
		t.Errorf("Expected %d successful requests, got %d", numRequests, successCount)
	}
}
