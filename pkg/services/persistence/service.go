package persistence

import (
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
	return Snapshot{
		ID:        id,
		RuntimeID: runtimeId,
		State:     make(map[string]any),
		Timestamp: time.Now(),
	}, nil
}

func (s *persistenceService) Restore(snapshotId string, def statechart.ChartDefinition) (statechart.RuntimeID, error) {
	panic("Not implemented")
}

func (s *persistenceService) AppendEvent(runtimeId string, event statechart.Event) error {
	panic("Not implemented")
}

func (s *persistenceService) GetEvents(runtimeId string, since string) ([]Event, error) {
	panic("Not implemented")
}

func (s *persistenceService) Migrate(runtimeId string, newDef statechart.ChartDefinition, policy MigrationPolicy) error {
	panic("Not implemented")
}
