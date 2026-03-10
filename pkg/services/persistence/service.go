package persistence

import (
	"fmt"
	"time"

	"github.com/maelstrom/v3/pkg/mail"
	"github.com/maelstrom/v3/pkg/security"
	"github.com/maelstrom/v3/pkg/statechart"
)

type SnapshotRecord struct {
	ID        string
	RuntimeID string
	State     map[string]any
	Taints    []string
	Timestamp time.Time
}

type EventLogEntry struct {
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
	Snapshot(runtimeId string, policy security.EnforcementPolicy) (SnapshotRecord, error)
	Restore(snapshotId string, def statechart.ChartDefinition) (statechart.RuntimeID, error)
	AppendEvent(runtimeId string, event statechart.Event) error
	GetEvents(runtimeId string, since string) ([]EventLogEntry, error)
	Migrate(runtimeId string, newDef statechart.ChartDefinition, policy MigrationPolicy) error
	HandleMail(mail mail.Mail) error
	Start() error
	Stop() error
}

type persistenceService struct {
	snapshots map[string]SnapshotRecord
	events    map[string][]EventLogEntry
	state     map[string]map[string]any
	taints    map[string][]string
	security  security.TaintEngine
}

func NewPersistenceService() PersistenceService {
	return &persistenceService{
		snapshots: make(map[string]SnapshotRecord),
		events:    make(map[string][]EventLogEntry),
		state:     make(map[string]map[string]any),
		taints:    make(map[string][]string),
		security:  security.NewTaintEngine(),
	}
}

func (s *persistenceService) ID() string {
	return "sys:persistence"
}

func (s *persistenceService) Snapshot(runtimeId string, policy security.EnforcementPolicy) (SnapshotRecord, error) {
	id := runtimeId + "-snap-" + time.Now().Format("20060102150405")

	runtimeTaints := s.taints[runtimeId]
	if err := s.security.EnforceAllowedOnExit(runtimeTaints, policy); err != nil {
		return SnapshotRecord{}, err
	}

	snap := SnapshotRecord{
		ID:        id,
		RuntimeID: runtimeId,
		State:     s.state[runtimeId],
		Taints:    runtimeTaints,
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
	s.snapshots[snapshotId] = SnapshotRecord{
		ID:        snapshotId,
		RuntimeID: snap.RuntimeID,
		State:     snap.State,
		Timestamp: time.Now(),
	}
	s.snapshots[string(newID)] = SnapshotRecord{
		ID:        string(newID),
		RuntimeID: snap.RuntimeID,
		State:     snap.State,
		Timestamp: time.Now(),
	}
	return newID, nil
}

func (s *persistenceService) AppendEvent(runtimeId string, event statechart.Event) error {
	id := runtimeId + "-evt-" + time.Now().Format("20060102150405") + "-" + fmt.Sprintf("%d", len(s.events[runtimeId]))
	ev := EventLogEntry{
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

func (s *persistenceService) GetEvents(runtimeId string, since string) ([]EventLogEntry, error) {
	events, ok := s.events[runtimeId]
	if !ok {
		return []EventLogEntry{}, nil
	}
	if since == "" {
		return events, nil
	}
	for i, ev := range events {
		if ev.ID == since {
			return events[i+1:], nil
		}
	}
	return []EventLogEntry{}, nil
}

func (s *persistenceService) Migrate(runtimeId string, newDef statechart.ChartDefinition, policy MigrationPolicy) error {
	switch policy {
	case ShallowHistory:
		_ = newDef
		_ = runtimeId
	case DeepHistory:
		_ = newDef
		_ = runtimeId
	case CleanStart:
		s.events[runtimeId] = []EventLogEntry{}
	}
	return nil
}

func (s *persistenceService) HandleMail(m mail.Mail) error {
	return nil
}

func (s *persistenceService) Start() error {
	return nil
}

func (s *persistenceService) Stop() error {
	return nil
}
