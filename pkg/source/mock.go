package source

// ManualSource is a test helper that allows manual injection of events.
type ManualSource struct {
	events chan SourceEvent
	err    error
}

// NewManualSource creates a source for testing that can receive manual events.
func NewManualSource() *ManualSource {
	return &ManualSource{
		events: make(chan SourceEvent, 10),
	}
}

// Send delivers an event to the source (non-blocking up to buffer size).
func (m *ManualSource) Send(evt SourceEvent) {
	m.events <- evt
}

// Close signals end of events.
func (m *ManualSource) Close(err error) {
	m.err = err
	close(m.events)
}

// Events implements Source interface.
func (m *ManualSource) Events() <-chan SourceEvent {
	return m.events
}

// Err implements Source interface.
func (m *ManualSource) Err() error {
	return m.err
}
