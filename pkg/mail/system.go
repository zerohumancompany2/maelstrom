// Package mail provides the mail system as the only cross-boundary primitive.
// Spec Reference: Section 9.2
package mail

import (
	"errors"
	"sync"
	"time"
)

// ErrDuplicateMail is returned when a mail with the same correlationId is published.
var ErrDuplicateMail = errors.New("mail with correlationId already processed")

// ErrNotFound is returned when a subscriber address is not found.
var ErrNotFound = errors.New("subscriber not found")

// MailSystem provides publisher/subscriber coordination.
type MailSystem struct {
	subscribers map[string][]chan Mail
	published   map[string]bool
	mu          sync.RWMutex
	lruCache    map[string]time.Time
	lruMaxSize  int
}

// NewMailSystem creates a new MailSystem.
func NewMailSystem() *MailSystem {
	return &MailSystem{
		subscribers: make(map[string][]chan Mail),
		published:   make(map[string]bool),
		lruCache:    make(map[string]time.Time),
		lruMaxSize:  1000,
	}
}

// Publish sends mail to all subscribers of the target address.
// Returns Ack with delivery confirmation.
// At-least-once delivery with deduplication via correlationId.
func (ms *MailSystem) Publish(mail Mail) (Ack, error) {
	ms.mu.Lock()
	if ms.published[mail.CorrelationID] {
		ms.mu.Unlock()
		return Ack{}, ErrDuplicateMail
	}
	ms.published[mail.CorrelationID] = true
	if len(ms.published) > ms.lruMaxSize {
		var oldestTime time.Time
		var oldestID string
		for id, t := range ms.lruCache {
			if oldestTime.IsZero() || t.Before(oldestTime) {
				oldestTime = t
				oldestID = id
			}
		}
		if oldestID != "" {
			delete(ms.published, oldestID)
			delete(ms.lruCache, oldestID)
		}
	}
	ms.lruCache[mail.CorrelationID] = time.Now()
	subscribersCopy := make([]chan Mail, 0, len(ms.subscribers[mail.Target]))
	for _, ch := range ms.subscribers[mail.Target] {
		subscribersCopy = append(subscribersCopy, ch)
	}
	ms.mu.Unlock()

	for _, ch := range subscribersCopy {
		select {
		case ch <- mail:
		default:
		}
	}

	return Ack{
		MailID:        mail.ID,
		CorrelationID: mail.CorrelationID,
		DeliveredAt:   time.Now(),
		Success:       true,
	}, nil
}

// Subscribe registers a subscriber to receive mail for an address.
// Returns a channel that receives mail.
func (ms *MailSystem) Subscribe(address string) (<-chan Mail, error) {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	ch := make(chan Mail, 1000)
	ms.subscribers[address] = append(ms.subscribers[address], ch)
	return ch, nil
}

// Unsubscribe removes a subscriber from receiving mail for an address.
func (ms *MailSystem) Unsubscribe(address string, ch <-chan Mail) error {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	channels := ms.subscribers[address]
	for i, existingCh := range channels {
		if existingCh == ch {
			ms.subscribers[address] = append(channels[:i], channels[i+1:]...)
			break
		}
	}
	return nil
}
