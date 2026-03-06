package testutil

import (
	"sync"
	"time"
)

// MockMailSystem provides a mock implementation of mail.System for testing.
type MockMailSystem struct {
	mu          sync.RWMutex
	subscribers map[string][]chan Mail
	published   []Mail
	dedupCache  map[string]bool
}

// Mail is a simplified mail type for testing.
type Mail struct {
	ID            string
	Type          string
	From          string
	To            string
	Content       []byte
	CorrelationID string
	Timestamp     time.Time
}

// NewMockMailSystem creates a new mock mail system.
func NewMockMailSystem() *MockMailSystem {
	return &MockMailSystem{
		subscribers: make(map[string][]chan Mail),
		published:   make([]Mail, 0),
		dedupCache:  make(map[string]bool),
	}
}

// Publish simulates publishing mail.
func (m *MockMailSystem) Publish(mail Mail) (Ack, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Check for duplicates
	if m.dedupCache[mail.CorrelationID] {
		return Ack{}, ErrDuplicateMail
	}
	m.dedupCache[mail.CorrelationID] = true

	m.published = append(m.published, mail)

	// Deliver to subscribers
	for address, channels := range m.subscribers {
		if address == mail.To || address == "*" {
			for _, ch := range channels {
				select {
				case ch <- mail:
				default:
				}
			}
		}
	}

	return Ack{
		MailID:        mail.ID,
		CorrelationID: mail.CorrelationID,
		DeliveredAt:   time.Now(),
		Success:       true,
	}, nil
}

// Subscribe simulates subscribing to an address.
func (m *MockMailSystem) Subscribe(address string) (<-chan Mail, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	ch := make(chan Mail, 100)
	m.subscribers[address] = append(m.subscribers[address], ch)
	return ch, nil
}

// Unsubscribe simulates unsubscribing from an address.
func (m *MockMailSystem) Unsubscribe(address string, ch <-chan Mail) error {
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
func (m *MockMailSystem) GetPublished() []Mail {
	m.mu.RLock()
	defer m.mu.RUnlock()
	result := make([]Mail, len(m.published))
	copy(result, m.published)
	return result
}

// Ack represents an acknowledgment.
type Ack struct {
	MailID        string
	CorrelationID string
	DeliveredAt   time.Time
	Success       bool
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
