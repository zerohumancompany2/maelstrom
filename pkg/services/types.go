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
