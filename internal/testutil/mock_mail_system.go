package testutil

import (
	"github.com/maelstrom/v3/pkg/mail"
	"sync"
	"time"
)

// MockMailSystem provides a mock implementation of mail.System for testing.
type MockMailSystem struct {
	mu          sync.RWMutex
	subscribers map[string][]chan mail.Mail
	published   []mail.Mail
	dedupCache  map[string]bool
}

// NewMockMailSystem creates a new mock mail system.
func NewMockMailSystem() *MockMailSystem {
	return &MockMailSystem{
		subscribers: make(map[string][]chan mail.Mail),
		published:   make([]mail.Mail, 0),
		dedupCache:  make(map[string]bool),
	}
}

// Publish simulates publishing mail.
func (m *MockMailSystem) Publish(mailMsg mail.Mail) (mail.Ack, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Check for duplicates
	if m.dedupCache[mailMsg.CorrelationID] {
		return mail.Ack{}, ErrDuplicateMail
	}
	m.dedupCache[mailMsg.CorrelationID] = true

	m.published = append(m.published, mailMsg)

	// Deliver to subscribers
	for address, channels := range m.subscribers {
		if address == mailMsg.Target || address == "*" {
			for _, ch := range channels {
				select {
				case ch <- mailMsg:
				default:
				}
			}
		}
	}

	return mail.Ack{
		MailID:        mailMsg.ID,
		CorrelationID: mailMsg.CorrelationID,
		DeliveredAt:   time.Now(),
		Success:       true,
	}, nil
}

// Subscribe simulates subscribing to an address.
func (m *MockMailSystem) Subscribe(address string) (<-chan mail.Mail, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	ch := make(chan mail.Mail, 100)
	m.subscribers[address] = append(m.subscribers[address], ch)
	return ch, nil
}

// Unsubscribe simulates unsubscribing from an address.
func (m *MockMailSystem) Unsubscribe(address string, ch <-chan mail.Mail) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if channels, ok := m.subscribers[address]; ok {
		for i, c := range channels {
			if c == ch {
				m.subscribers[address] = append(channels[:i], channels[i+1:]...)
				return nil
			}
		}
	}
	return ErrNotFound
}

// PublishedCount returns the number of published mails.
func (m *MockMailSystem) PublishedCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.published)
}

// GetPublished returns all published mails.
func (m *MockMailSystem) GetPublished() []mail.Mail {
	m.mu.RLock()
	defer m.mu.RUnlock()
	result := make([]mail.Mail, len(m.published))
	copy(result, m.published)
	return result
}

// ErrDuplicateMail is returned when a mail with the same correlationId is published.
var ErrDuplicateMail = &Error{"duplicate mail"}

// ErrNotFound is returned when a subscriber address is not found.
var ErrNotFound = &Error{"subscriber not found"}

// Error is a custom error type.
type Error struct {
	msg string
}

func (e *Error) Error() string {
	return e.msg
}
