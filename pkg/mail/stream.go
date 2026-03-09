package mail

import (
	"fmt"
	"sync"
	"time"
)

// StreamSession represents an active streaming session
type StreamSession struct {
	ID          string
	LastEventID *string
	Chunks      chan StreamChunk
	Closed      bool
	mu          sync.RWMutex
	CreatedAt   time.Time
}

// NewStreamSession creates a new streaming session
func NewStreamSession(sessionID string, lastEventID *string) *StreamSession {
	return &StreamSession{
		ID:          sessionID,
		LastEventID: lastEventID,
		Chunks:      make(chan StreamChunk, 100),
		Closed:      false,
		CreatedAt:   time.Now(),
	}
}

// Send sends a chunk to the session
func (s *StreamSession) Send(chunk StreamChunk) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.Closed {
		return fmt.Errorf("session is closed")
	}
	s.Chunks <- chunk
	return nil
}

// Close closes the session
func (s *StreamSession) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.Closed {
		return nil
	}
	s.Closed = true
	close(s.Chunks)
	return nil
}

// UpgradeToStream upgrades a connection to streaming mode
func UpgradeToStream(sessionID string, lastEventID *string) (chan StreamChunk, error) {
	session := NewStreamSession(sessionID, lastEventID)
	return session.Chunks, nil
}

// StripForbiddenTaints removes taints not in the allowed list
func StripForbiddenTaints(chunk StreamChunk, allowed []string) StreamChunk {
	if len(allowed) == 0 {
		chunk.Taints = nil
		return chunk
	}

	allowedMap := make(map[string]bool)
	for _, a := range allowed {
		allowedMap[a] = true
	}

	var filtered []string
	for _, taint := range chunk.Taints {
		if allowedMap[taint] {
			filtered = append(filtered, taint)
		}
	}
	chunk.Taints = filtered
	return chunk
}
