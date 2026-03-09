package observability

import (
	"sync"
	"time"

	"github.com/maelstrom/v3/pkg/mail"
	"github.com/maelstrom/v3/pkg/services"
)

type ObservabilityService struct {
	mu          sync.Mutex
	traces      []services.Trace
	deadLetters []DeadLetterEntry
}

func NewObservabilityService() *ObservabilityService {
	return &ObservabilityService{}
}

func (o *ObservabilityService) ID() string {
	return "sys:observability"
}

func (o *ObservabilityService) Boundary() mail.BoundaryType {
	return mail.InnerBoundary
}

func (o *ObservabilityService) HandleMail(mail mail.Mail) error {
	return nil
}

func (o *ObservabilityService) EmitTrace(trace services.Trace) error {
	o.mu.Lock()
	defer o.mu.Unlock()
	o.traces = append(o.traces, trace)
	return nil
}

func (o *ObservabilityService) QueryTraces(filters services.TraceFilters) ([]services.Trace, error) {
	o.mu.Lock()
	defer o.mu.Unlock()
	var result []services.Trace
	for _, trace := range o.traces {
		if filters.RuntimeID != "" && trace.RuntimeID != filters.RuntimeID {
			continue
		}
		if filters.EventType != "" && trace.EventType != filters.EventType {
			continue
		}
		if !filters.FromTime.IsZero() && trace.Timestamp.Before(filters.FromTime) {
			continue
		}
		if !filters.ToTime.IsZero() && trace.Timestamp.After(filters.ToTime) {
			continue
		}
		result = append(result, trace)
	}
	return result, nil
}

func (o *ObservabilityService) Start() error {
	return nil
}

func (o *ObservabilityService) Stop() error {
	return nil
}

func (o *ObservabilityService) LogDeadLetter(mail mail.Mail, reason string) error {
	o.mu.Lock()
	defer o.mu.Unlock()
	entry := DeadLetterEntry{
		Mail:   mail,
		Reason: reason,
		Logged: time.Now(),
	}
	o.deadLetters = append(o.deadLetters, entry)
	return nil
}

func (o *ObservabilityService) QueryDeadLetters() ([]DeadLetterEntry, error) {
	o.mu.Lock()
	defer o.mu.Unlock()
	result := make([]DeadLetterEntry, len(o.deadLetters))
	copy(result, o.deadLetters)
	return result, nil
}
