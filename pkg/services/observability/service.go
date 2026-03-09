package observability

import (
	"sync"

	"github.com/maelstrom/v3/pkg/mail"
	"github.com/maelstrom/v3/pkg/services"
)

type ObservabilityService struct {
	mu sync.Mutex
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
	return nil
}

func (o *ObservabilityService) QueryTraces(runtimeID string) ([]services.Trace, error) {
	return nil, nil
}

func (o *ObservabilityService) Start() error {
	return nil
}

func (o *ObservabilityService) Stop() error {
	return nil
}
