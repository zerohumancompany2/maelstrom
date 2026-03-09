package observability

import (
	"sync"

	"github.com/maelstrom/v3/pkg/mail"
	"github.com/maelstrom/v3/pkg/services"
)

type ObservabilityService struct {
	mu     sync.Mutex
	traces []services.Trace
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
