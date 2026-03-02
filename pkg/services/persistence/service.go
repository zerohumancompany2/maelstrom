package persistence

import (
	"fmt"
	"time"

	"github.com/maelstrom/v3/pkg/statechart"
)

type Snapshot struct {
	ID        string
	RuntimeID string
	State     map[string]any
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

type MigrationPolicy string

const (
	ShallowHistory MigrationPolicy = "shallowHistory"
	DeepHistory    MigrationPolicy = "deepHistory"
	CleanStart     MigrationPolicy = "cleanStart"
)

type PersistenceService interface {
	Snapshot(runtimeId string) (Snapshot, error)
	Restore(snapshotId string, def statechart.ChartDefinition) (statechart.RuntimeID, error)
	AppendEvent(runtimeId string, event statechart.Event) error
	GetEvents(runtimeId string, since string) ([]Event, error)
	Migrate(runtimeId string, newDef statechart.ChartDefinition, policy MigrationPolicy) error
}

type persistenceService struct {
	snapshots map[string]Snapshot
	events    map[string][]Event
}

func NewPersistenceService() PersistenceService {
	return &persistenceService{
		snapshots: make(map[string]Snapshot),
		events:    make(map[string][]Event),
	}
}

func (s *persistenceService) Snapshot(runtimeId string) (Snapshot, error) {
	id := runtimeId + "-snap-" + time.Now().Format("20060102150405")
	snap := Snapshot{
		ID:        id,
		RuntimeID: runtimeId,
		State:     make(map[string]any),
		Timestamp: time.Now(),
	}
	s.snapshots[id] = snap
	return snap, nil
}

func (s *persistenceService) Restore(snapshotId string, def statechart.ChartDefinition) (statechart.RuntimeID, error) {
	snap, ok := s.snapshots[snapshotId]
	if !ok {
		return "", fmt.Errorf("snapshot not found")
	}
	newID := statechart.RuntimeID(snapshotId + "-restored-" + time.Now().Format("20060102150405"))
	s.snapshots[snapshotId] = Snapshot{
		ID:        snapshotId,
		RuntimeID: snap.RuntimeID,
		State:     snap.State,
		Timestamp: time.Now(),
	}
	s.snapshots[string(newID)] = Snapshot{
		ID:        string(newID),
		RuntimeID: snap.RuntimeID,
		State:     snap.State,
		Timestamp: time.Now(),
	}
	return newID, nil
}

func (s *persistenceService) AppendEvent(runtimeId string, event statechart.Event) error {
	id := runtimeId + "-evt-" + time.Now().Format("20060102150405") + "-" + fmt.Sprintf("%d", len(s.events[runtimeId]))
	ev := Event{
		ID:            id,
		RuntimeID:     runtimeId,
		Type:          event.Type,
		Payload:       event.Payload,
		CorrelationID: event.CorrelationID,
		Source:        event.Source,
		Timestamp:     time.Now(),
	}
	s.events[runtimeId] = append(s.events[runtimeId], ev)
	return nil
}

func (s *persistenceService) GetEvents(runtimeId string, since string) ([]Event, error) {
	events, ok := s.events[runtimeId]
	if !ok {
		return []Event{}, nil
	}
	if since == "" {
		return events, nil
	}
	for i, ev := range events {
		if ev.ID == since {
			return events[i+1:], nil
		}
	}
	return []Event{}, nil
}

func (s *persistenceService) Migrate(runtimeId string, newDef statechart.ChartDefinition, policy MigrationPolicy) error {
	panic("Not implemented")
}
