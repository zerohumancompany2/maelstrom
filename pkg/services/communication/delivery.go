package communication

import (
	"errors"
	"sync"
	"time"

	"github.com/maelstrom/v3/pkg/mail"
)

var (
	ErrMailNotFound        = errors.New("mail not found")
	ErrAlreadyAcknowledged = errors.New("mail already acknowledged")
	ErrMaxRetriesExceeded  = errors.New("max retries exceeded")
)

type DeliveryState string

const (
	DeliveryStatePending   DeliveryState = "pending"
	DeliveryStateDelivered DeliveryState = "delivered"
	DeliveryStateFailed    DeliveryState = "failed"
)

type DeliveryRecord struct {
	MailID        string
	CorrelationID string
	Target        string
	State         DeliveryState
	AttemptCount  int
	LastAttempt   time.Time
	CreatedAt     time.Time
	ErrorMessage  string
}

type DeliveryTracker struct {
	records map[string]*DeliveryRecord
	mu      sync.RWMutex
}

func NewDeliveryTracker() *DeliveryTracker {
	return &DeliveryTracker{
		records: make(map[string]*DeliveryRecord),
	}
}

func (dt *DeliveryTracker) Track(mailID string, correlationID string, target string) error {
	dt.mu.Lock()
	defer dt.mu.Unlock()

	if _, exists := dt.records[mailID]; exists {
		return nil
	}

	dt.records[mailID] = &DeliveryRecord{
		MailID:        mailID,
		CorrelationID: correlationID,
		Target:        target,
		State:         DeliveryStatePending,
		AttemptCount:  0,
		CreatedAt:     time.Now(),
		LastAttempt:   time.Time{},
	}

	return nil
}

func (dt *DeliveryTracker) Acknowledge(mailID string) error {
	dt.mu.Lock()
	defer dt.mu.Unlock()

	record, exists := dt.records[mailID]
	if !exists {
		return ErrMailNotFound
	}

	if record.State == DeliveryStateDelivered {
		return ErrAlreadyAcknowledged
	}

	record.State = DeliveryStateDelivered
	record.LastAttempt = time.Now()

	return nil
}

func (dt *DeliveryTracker) GetPending() []string {
	dt.mu.RLock()
	defer dt.mu.RUnlock()

	var pending []string
	for mailID, record := range dt.records {
		if record.State == DeliveryStatePending {
			pending = append(pending, mailID)
		}
	}

	return pending
}

func (dt *DeliveryTracker) GetRecord(mailID string) (*DeliveryRecord, error) {
	dt.mu.RLock()
	defer dt.mu.RUnlock()

	record, exists := dt.records[mailID]
	if !exists {
		return nil, ErrMailNotFound
	}

	return record, nil
}

func (dt *DeliveryTracker) Fail(mailID string, errorMessage string) error {
	dt.mu.Lock()
	defer dt.mu.Unlock()

	record, exists := dt.records[mailID]
	if !exists {
		return ErrMailNotFound
	}

	record.State = DeliveryStateFailed
	record.ErrorMessage = errorMessage
	record.LastAttempt = time.Now()

	return nil
}

func (dt *DeliveryTracker) IncrementAttempt(mailID string) error {
	dt.mu.Lock()
	defer dt.mu.Unlock()

	record, exists := dt.records[mailID]
	if !exists {
		return ErrMailNotFound
	}

	record.AttemptCount++
	record.LastAttempt = time.Now()

	return nil
}

func (dt *DeliveryTracker) ClearDelivered() {
	dt.mu.Lock()
	defer dt.mu.Unlock()

	for mailID, record := range dt.records {
		if record.State == DeliveryStateDelivered {
			delete(dt.records, mailID)
		}
	}
}

type RetryPolicy struct {
	MaxRetries     int
	InitialBackoff time.Duration
	MaxBackoff     time.Duration
	BackoffFactor  float64
}

func NewRetryPolicy() *RetryPolicy {
	return &RetryPolicy{
		MaxRetries:     3,
		InitialBackoff: time.Second,
		MaxBackoff:     30 * time.Second,
		BackoffFactor:  2.0,
	}
}

func (rp *RetryPolicy) Backoff(attempt int) time.Duration {
	backoff := float64(rp.InitialBackoff) * pow(rp.BackoffFactor, float64(attempt))
	if backoff > float64(rp.MaxBackoff) {
		return rp.MaxBackoff
	}
	return time.Duration(backoff)
}

func (rp *RetryPolicy) ShouldRetry(attempt int) bool {
	return attempt < rp.MaxRetries
}

func (rp *RetryPolicy) GetMaxRetries() int {
	return rp.MaxRetries
}

func pow(base, exponent float64) float64 {
	result := 1.0
	for i := 0; i < int(exponent); i++ {
		result *= base
	}
	return result
}

type DeadLetterEntry struct {
	Mail      mail.Mail
	Reason    string
	Attempts  int
	CreatedAt time.Time
}

type DeadLetterQueue struct {
	entries []DeadLetterEntry
	mu      sync.RWMutex
}

func NewDeadLetterQueue() *DeadLetterQueue {
	return &DeadLetterQueue{
		entries: make([]DeadLetterEntry, 0),
	}
}

func (dlq *DeadLetterQueue) Add(entry DeadLetterEntry) {
	dlq.mu.Lock()
	defer dlq.mu.Unlock()

	dlq.entries = append(dlq.entries, entry)
}

func (dlq *DeadLetterQueue) GetEntries() []DeadLetterEntry {
	dlq.mu.RLock()
	defer dlq.mu.RUnlock()

	result := make([]DeadLetterEntry, len(dlq.entries))
	copy(result, dlq.entries)
	return result
}

func (dlq *DeadLetterQueue) Clear() {
	dlq.mu.Lock()
	defer dlq.mu.Unlock()

	dlq.entries = make([]DeadLetterEntry, 0)
}

type DeliveryGuarantee struct {
	tracker       *DeliveryTracker
	retryPolicy   *RetryPolicy
	deadLetter    *DeadLetterQueue
	observability ObservabilityInterface
	mu            sync.RWMutex
}

type ObservabilityInterface interface {
	LogDeadLetter(mail mail.Mail, reason string)
}

func NewDeliveryGuarantee() *DeliveryGuarantee {
	return &DeliveryGuarantee{
		tracker:     NewDeliveryTracker(),
		retryPolicy: NewRetryPolicy(),
		deadLetter:  NewDeadLetterQueue(),
	}
}

func (dg *DeliveryGuarantee) SetObservability(obs ObservabilityInterface) {
	dg.mu.Lock()
	defer dg.mu.Unlock()
	dg.observability = obs
}

func (dg *DeliveryGuarantee) Track(mail mail.Mail) error {
	return dg.tracker.Track(mail.ID, mail.CorrelationID, mail.Target)
}

func (dg *DeliveryGuarantee) Acknowledge(mailID string) error {
	return dg.tracker.Acknowledge(mailID)
}

func (dg *DeliveryGuarantee) GetPending() []string {
	return dg.tracker.GetPending()
}

func (dg *DeliveryGuarantee) IncrementAttempt(mailID string) error {
	return dg.tracker.IncrementAttempt(mailID)
}

func (dg *DeliveryGuarantee) Fail(mailID string, errorMessage string) error {
	return dg.tracker.Fail(mailID, errorMessage)
}

func (dg *DeliveryGuarantee) DeliverWithRetry(mail mail.Mail, deliverFunc func(mail mail.Mail) (bool, error)) error {
	if err := dg.Track(mail); err != nil {
		return err
	}

	for attempt := 0; attempt <= dg.retryPolicy.GetMaxRetries(); attempt++ {
		if attempt > 0 {
			backoff := dg.retryPolicy.Backoff(attempt - 1)
			time.Sleep(backoff)
		}

		success, err := deliverFunc(mail)
		if err != nil {
			dg.IncrementAttempt(mail.ID)
			if !dg.retryPolicy.ShouldRetry(attempt) {
				dg.sendToDeadLetter(mail, "delivery failed after max retries")
				return ErrMaxRetriesExceeded
			}
			continue
		}

		if success {
			return dg.Acknowledge(mail.ID)
		}

		if !dg.retryPolicy.ShouldRetry(attempt) {
			dg.sendToDeadLetter(mail, "delivery failed after max retries")
			return ErrMaxRetriesExceeded
		}
	}

	dg.sendToDeadLetter(mail, "delivery failed after max retries")
	return ErrMaxRetriesExceeded
}

func (dg *DeliveryGuarantee) sendToDeadLetter(mail mail.Mail, reason string) {
	dg.mu.RLock()
	obs := dg.observability
	dg.mu.RUnlock()

	record, err := dg.tracker.GetRecord(mail.ID)
	if err != nil {
		record = &DeliveryRecord{AttemptCount: 0}
	}

	entry := DeadLetterEntry{
		Mail:      mail,
		Reason:    reason,
		Attempts:  record.AttemptCount,
		CreatedAt: time.Now(),
	}

	dg.deadLetter.Add(entry)

	if obs != nil {
		obs.LogDeadLetter(mail, reason)
	}
}

func (dg *DeliveryGuarantee) GetDeadLetterEntries() []DeadLetterEntry {
	return dg.deadLetter.GetEntries()
}

func (dg *DeliveryGuarantee) ClearDeadLetter() {
	dg.deadLetter.Clear()
}

func (dg *DeliveryGuarantee) ClearDelivered() {
	dg.tracker.ClearDelivered()
}
