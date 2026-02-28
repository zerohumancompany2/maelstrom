package runtime

import (
	"context"
	"fmt"
	"time"

	"github.com/maelstrom/v3/pkg/statechart"
)

// ChartRuntime represents an instantiated, running chart.
type ChartRuntime struct {
	ID      string
	Context RuntimeContext
	def     interface{}
	events  chan Event
	done    chan struct{}
	// Statechart integration (Phase 2)
	engine     interface{ Dispatch(id string, ev statechart.Event) error }
	runtimeID  string
}

// RuntimeContext provides read-only access to chart runtime information.
type RuntimeContext struct {
	ChartID      string
	ParentID     string
	StartTime    int64
	ActiveStates []string
	Metrics      map[string]interface{}
}

// Event represents a statechart event.
type Event struct {
	Type          string
	Payload       map[string]interface{}
	CorrelationID string
	Source        string
}

// HasID is an interface for types that can provide an ID.
type HasID interface {
	GetID() string
}

// NewChartRuntime creates a new runtime from a hydrated chart definition.
func NewChartRuntime(id string, def interface{}) (*ChartRuntime, error) {
	var chartID string

	switch d := def.(type) {
	case map[string]interface{}:
		chartID, _ = d["id"].(string)
	case interface{ GetID() string }:
		chartID = d.GetID()
	case string:
		chartID = d
	default:
		chartID = "unknown"
	}

	if chartID == "" {
		chartID = "unknown"
	}

	return &ChartRuntime{
		ID:     id,
		def:    def,
		events: make(chan Event, 100),
		done:   make(chan struct{}),
		Context: RuntimeContext{
			ChartID:      chartID,
			StartTime:    time.Now().Unix(),
			ActiveStates: []string{"idle"},
			Metrics:      make(map[string]interface{}),
		},
	}, nil
}

// SendEvent dispatches an event to the runtime (non-blocking).
// If connected to a statechart engine, forwards to the engine.
func (r *ChartRuntime) SendEvent(evt Event) error {
	// If connected to statechart engine, forward there
	if r.engine != nil && r.runtimeID != "" {
		statechartEvt := statechart.Event{
			Type:          evt.Type,
			Payload:       evt.Payload,
			CorrelationID: evt.CorrelationID,
			Source:        evt.Source,
		}
		return r.engine.Dispatch(r.runtimeID, statechartEvt)
	}

	// Otherwise, use internal event queue
	select {
	case r.events <- evt:
		return nil
	default:
		return fmt.Errorf("event queue full")
	}
}

// ActiveStates returns the currently active states.
func (r *ChartRuntime) ActiveStates() []string {
	return r.Context.ActiveStates
}

// Start begins the runtime's event processing loop.
func (r *ChartRuntime) Start(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-r.done:
			return nil
		case evt := <-r.events:
			r.processEvent(evt)
		}
	}
}

func (r *ChartRuntime) processEvent(evt Event) {
	// Stub: in real implementation, this would use the statechart engine
	r.Context.Metrics["last_event"] = evt.Type
}

// Stop gracefully shuts down the runtime.
func (r *ChartRuntime) Stop() error {
	close(r.done)
	return nil
}
