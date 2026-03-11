package communication

import (
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/maelstrom/v3/pkg/mail"
)

func TestDeliveryTracker_TrackCreatesRecord(t *testing.T) {
	tracker := NewDeliveryTracker()

	mailID := "mail-123"
	correlationID := "corr-456"
	target := "agent:test"

	err := tracker.Track(mailID, correlationID, target)

	if err != nil {
		t.Errorf("Expected nil error, got %v", err)
	}

	record, err := tracker.GetRecord(mailID)
	if err != nil {
		t.Errorf("Expected to find record, got error %v", err)
	}
	if record.MailID != mailID {
		t.Errorf("Expected MailID %s, got %s", mailID, record.MailID)
	}
	if record.CorrelationID != correlationID {
		t.Errorf("Expected CorrelationID %s, got %s", correlationID, record.CorrelationID)
	}
	if record.Target != target {
		t.Errorf("Expected Target %s, got %s", target, record.Target)
	}
	if record.State != DeliveryStatePending {
		t.Errorf("Expected State pending, got %s", record.State)
	}
}

func TestDeliveryTracker_AcknowledgeMarksAsDelivered(t *testing.T) {
	tracker := NewDeliveryTracker()

	mailID := "mail-789"
	tracker.Track(mailID, "corr-abc", "agent:test")

	err := tracker.Acknowledge(mailID)

	if err != nil {
		t.Errorf("Expected nil error, got %v", err)
	}

	record, err := tracker.GetRecord(mailID)
	if err != nil {
		t.Errorf("Expected to find record, got error %v", err)
	}
	if record.State != DeliveryStateDelivered {
		t.Errorf("Expected State delivered, got %s", record.State)
	}
	if record.LastAttempt.IsZero() {
		t.Error("Expected LastAttempt to be set")
	}
}

func TestDeliveryTracker_GetPendingReturnsOnlyPending(t *testing.T) {
	tracker := NewDeliveryTracker()

	tracker.Track("mail-1", "corr-1", "agent:test")
	tracker.Track("mail-2", "corr-2", "agent:test")
	tracker.Track("mail-3", "corr-3", "agent:test")

	tracker.Acknowledge("mail-2")

	pending := tracker.GetPending()

	if len(pending) != 2 {
		t.Errorf("Expected 2 pending mails, got %d", len(pending))
	}

	hasMail1 := false
	hasMail3 := false
	for _, id := range pending {
		if id == "mail-1" {
			hasMail1 = true
		}
		if id == "mail-3" {
			hasMail3 = true
		}
	}

	if !hasMail1 {
		t.Error("Expected mail-1 in pending list")
	}
	if !hasMail3 {
		t.Error("Expected mail-3 in pending list")
	}
}

func TestRetryPolicy_ExponentialBackoff(t *testing.T) {
	policy := NewRetryPolicy()

	backoff0 := policy.Backoff(0)
	backoff1 := policy.Backoff(1)
	backoff2 := policy.Backoff(2)

	if backoff0 != policy.InitialBackoff {
		t.Errorf("Expected initial backoff %v, got %v", policy.InitialBackoff, backoff0)
	}
	if backoff1 != 2*backoff0 {
		t.Errorf("Expected backoff at attempt 1 to be 2x initial (%v), got %v", 2*backoff0, backoff1)
	}
	if backoff2 != 4*backoff0 {
		t.Errorf("Expected backoff at attempt 2 to be 4x initial (%v), got %v", 4*backoff0, backoff2)
	}
}

func TestRetryPolicy_RespectsMaxBackoff(t *testing.T) {
	policy := NewRetryPolicy()

	backoff10 := policy.Backoff(10)

	if backoff10 > policy.MaxBackoff {
		t.Errorf("Expected backoff to not exceed max backoff %v, got %v", policy.MaxBackoff, backoff10)
	}
}

func TestRetryPolicy_ShouldRetry(t *testing.T) {
	policy := NewRetryPolicy()

	if !policy.ShouldRetry(0) {
		t.Error("Should retry on attempt 0")
	}
	if !policy.ShouldRetry(1) {
		t.Error("Should retry on attempt 1")
	}
	if !policy.ShouldRetry(2) {
		t.Error("Should retry on attempt 2")
	}
	if policy.ShouldRetry(3) {
		t.Error("Should not retry on attempt 3 (max retries reached)")
	}
}

func TestDeliveryGuarantee_DeliverWithRetrySuccess(t *testing.T) {
	dg := NewDeliveryGuarantee()

	attemptCount := 0
	deliverFunc := func(mail mail.Mail) (bool, error) {
		attemptCount++
		return true, nil
	}

	mail := mail.Mail{
		ID:            "mail-success",
		CorrelationID: "corr-success",
		Source:        "agent:sender",
		Target:        "agent:receiver",
		Type:          mail.MailTypeUser,
		Content:       "test content",
	}

	err := dg.DeliverWithRetry(mail, deliverFunc)

	if err != nil {
		t.Errorf("Expected nil error, got %v", err)
	}
	if attemptCount != 1 {
		t.Errorf("Expected 1 attempt, got %d", attemptCount)
	}

	record, err := dg.tracker.GetRecord(mail.ID)
	if err != nil {
		t.Errorf("Expected to find record, got error %v", err)
	}
	if record.State != DeliveryStateDelivered {
		t.Errorf("Expected State delivered, got %s", record.State)
	}
}

func TestDeliveryGuarantee_DeliverWithRetryRetriesOnFailure(t *testing.T) {
	dg := NewDeliveryGuarantee()

	attemptCount := 0
	deliverFunc := func(mail mail.Mail) (bool, error) {
		attemptCount++
		if attemptCount < 3 {
			return false, errors.New("temporary failure")
		}
		return true, nil
	}

	mail := mail.Mail{
		ID:            "mail-retry",
		CorrelationID: "corr-retry",
		Source:        "agent:sender",
		Target:        "agent:receiver",
		Type:          mail.MailTypeUser,
		Content:       "test content",
	}

	err := dg.DeliverWithRetry(mail, deliverFunc)

	if err != nil {
		t.Errorf("Expected nil error, got %v", err)
	}
	if attemptCount != 3 {
		t.Errorf("Expected 3 attempts, got %d", attemptCount)
	}
}

func TestDeliveryGuarantee_DeadLetterOnMaxRetries(t *testing.T) {
	dg := NewDeliveryGuarantee()

	deliverFunc := func(mail mail.Mail) (bool, error) {
		return false, errors.New("permanent failure")
	}

	mail := mail.Mail{
		ID:            "mail-fail",
		CorrelationID: "corr-fail",
		Source:        "agent:sender",
		Target:        "agent:receiver",
		Type:          mail.MailTypeUser,
		Content:       "test content",
	}

	err := dg.DeliverWithRetry(mail, deliverFunc)

	if err != ErrMaxRetriesExceeded {
		t.Errorf("Expected ErrMaxRetriesExceeded, got %v", err)
	}

	entries := dg.GetDeadLetterEntries()
	if len(entries) != 1 {
		t.Errorf("Expected 1 dead letter entry, got %d", len(entries))
	}
	if entries[0].Mail.ID != mail.ID {
		t.Errorf("Expected mail ID %s, got %s", mail.ID, entries[0].Mail.ID)
	}
	if entries[0].Reason != "delivery failed after max retries" {
		t.Errorf("Expected reason 'delivery failed after max retries', got %s", entries[0].Reason)
	}
}

func TestDeliveryGuarantee_AtLeastOnceDelivery(t *testing.T) {
	dg := NewDeliveryGuarantee()

	deliveryCount := 0
	var mu sync.Mutex

	deliverFunc := func(mail mail.Mail) (bool, error) {
		mu.Lock()
		deliveryCount++
		mu.Unlock()
		return true, nil
	}

	mail := mail.Mail{
		ID:            "mail-once",
		CorrelationID: "corr-once",
		Source:        "agent:sender",
		Target:        "agent:receiver",
		Type:          mail.MailTypeUser,
		Content:       "test content",
	}

	err := dg.DeliverWithRetry(mail, deliverFunc)

	if err != nil {
		t.Errorf("Expected nil error, got %v", err)
	}
	if deliveryCount < 1 {
		t.Error("Expected mail to be delivered at least once")
	}
}

func TestDeliveryGuarantee_IdempotencyAtReceiver(t *testing.T) {
	dg := NewDeliveryGuarantee()

	seenCorrelations := make(map[string]bool)
	var mu sync.Mutex

	deliverFunc := func(m mail.Mail) (bool, error) {
		mu.Lock()
		seenCorrelations[m.CorrelationID] = true
		mu.Unlock()
		return true, nil
	}

	mail1 := mail.Mail{
		ID:            "mail-idempotent-1",
		CorrelationID: "corr-idempotent",
		Source:        "agent:sender",
		Target:        "agent:receiver",
		Type:          mail.MailTypeUser,
		Content:       "test content",
	}

	mail2 := mail.Mail{
		ID:            "mail-idempotent-2",
		CorrelationID: "corr-idempotent",
		Source:        "agent:sender",
		Target:        "agent:receiver",
		Type:          mail.MailTypeUser,
		Content:       "test content",
	}

	err1 := dg.DeliverWithRetry(mail1, deliverFunc)
	if err1 != nil {
		t.Errorf("First delivery failed: %v", err1)
	}

	err2 := dg.DeliverWithRetry(mail2, deliverFunc)
	if err2 != nil {
		t.Errorf("Second delivery failed: %v", err2)
	}

	mu.Lock()
	count := len(seenCorrelations)
	mu.Unlock()

	if count != 1 {
		t.Errorf("Expected 1 unique correlation ID, got %d", count)
	}
}

func TestDeliveryGuarantee_BackoffTiming(t *testing.T) {
	dg := NewDeliveryGuarantee()

	dg.retryPolicy.InitialBackoff = 10 * time.Millisecond
	dg.retryPolicy.MaxBackoff = 100 * time.Millisecond

	attemptCount := 0
	deliverFunc := func(mail mail.Mail) (bool, error) {
		attemptCount++
		if attemptCount < 3 {
			return false, errors.New("temporary failure")
		}
		return true, nil
	}

	mail := mail.Mail{
		ID:            "mail-timing",
		CorrelationID: "corr-timing",
		Source:        "agent:sender",
		Target:        "agent:receiver",
		Type:          mail.MailTypeUser,
		Content:       "test content",
	}

	start := time.Now()
	err := dg.DeliverWithRetry(mail, deliverFunc)
	elapsed := time.Since(start)

	if err != nil {
		t.Errorf("Expected nil error, got %v", err)
	}

	expectedMin := 10*time.Millisecond + 20*time.Millisecond
	if elapsed < expectedMin {
		t.Errorf("Expected at least %v elapsed for backoff, got %v", expectedMin, elapsed)
	}
}
