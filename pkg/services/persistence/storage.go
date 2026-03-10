package persistence

import (
	"errors"
	"sync"
	"time"
)

var (
	ErrSnapshotNotFound = errors.New("snapshot not found")
	ErrInvalidFilter    = errors.New("invalid filter")
)

type Snapshot struct {
	ID        string
	RuntimeID string
	State     map[string]any
	Taints    []string
	Timestamp time.Time
}

type Event struct {
	ID            string
	RuntimeID     string
	Type          string
	Payload       any
	CorrelationID string
	Source        string
	Timestamp     time.Time
}

type EventFilters struct {
	RuntimeID string
	StartTime time.Time
	EndTime   time.Time
}

type StorageBackend interface {
	SaveSnapshot(snapshot Snapshot) error
	LoadSnapshot(id string) (Snapshot, error)
	SaveEvent(event Event) error
	QueryEvents(filters EventFilters) ([]Event, error)
}

type storageBackend struct {
	mu        sync.RWMutex
	snapshots map[string]Snapshot
	events    map[string][]Event
}

func NewStorageBackend() StorageBackend {
	return &storageBackend{
		snapshots: make(map[string]Snapshot),
		events:    make(map[string][]Event),
	}
}

func (s *storageBackend) SaveSnapshot(snapshot Snapshot) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.snapshots[snapshot.ID] = snapshot
	return nil
}

func (s *storageBackend) LoadSnapshot(id string) (Snapshot, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	snap, ok := s.snapshots[id]
	if !ok {
		return Snapshot{}, ErrSnapshotNotFound
	}
	return snap, nil
}

func (s *storageBackend) SaveEvent(event Event) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.events[event.RuntimeID] = append(s.events[event.RuntimeID], event)
	return nil
}

func (s *storageBackend) QueryEvents(filters EventFilters) ([]Event, error) {
	return nil, nil
}
