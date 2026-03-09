package services

import "time"

type Trace struct {
	ID        string    // unique trace ID
	RuntimeID string    // originating ChartRuntime
	EventType string    // transition, entry, exit, event
	StatePath string    // state path at event time
	Timestamp time.Time // event time
	Payload   any       // event payload
}

type TraceFilters struct {
	RuntimeID string
	EventType string
	FromTime  time.Time
	ToTime    time.Time
}

type MetricsCollector struct {
	StateCounts    map[string]int
	TransitionRate float64
	EventRate      float64
	LastUpdate     time.Time
}
