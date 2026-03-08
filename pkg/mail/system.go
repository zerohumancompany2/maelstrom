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
	subscribers map[string]chan Mail
	published   map[string]bool
	mu          sync.RWMutex
	lruCache    map[string]time.Time
	lruMaxSize  int
}

// NewMailSystem creates a new MailSystem.
func NewMailSystem() *MailSystem {
	return &MailSystem{
		subscribers: make(map[string]chan Mail),
		published:   make(map[string]bool),
		lruCache:    make(map[string]time.Time),
		lruMaxSize:  1000,
	}
}

// Publish sends mail to all subscribers of the target address.
// Returns Ack with delivery confirmation.
// At-least-once delivery with deduplication via correlationId.
func (ms *MailSystem) Publish(mail Mail) (Ack, error) {
	// TODO: implement
	return Ack{}, nil
}

// Subscribe registers a subscriber to receive mail for an address.
// Returns a channel that receives mail.
func (ms *MailSystem) Subscribe(address string) (<-chan Mail, error) {
	// TODO: implement
	return nil, nil
}

// Unsubscribe removes a subscriber from receiving mail for an address.
func (ms *MailSystem) Unsubscribe(address string, ch <-chan Mail) error {
	// TODO: implement
	return nil
}

// TODO: implement LRU cache management for deduplication
// TODO: implement thread-safe queue management
// TODO: implement concurrent safety for all operations
