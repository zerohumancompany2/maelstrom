package lifecycle

import (
	"sync"

	"github.com/maelstrom/v3/pkg/mail"
	"github.com/maelstrom/v3/pkg/statechart"
)

type LifecycleService struct {
	mu       sync.Mutex
	engine   statechart.Library
	runtimes map[statechart.RuntimeID]RuntimeInfo
}

func NewLifecycleService(engine statechart.Library) *LifecycleService {
	return &LifecycleService{
		engine:   engine,
		runtimes: make(map[statechart.RuntimeID]RuntimeInfo),
	}
}

func NewLifecycleServiceWithoutEngine() *LifecycleService {
	return &LifecycleService{
		runtimes: make(map[statechart.RuntimeID]RuntimeInfo),
	}
}

func (l *LifecycleService) ID() string {
	return "sys:lifecycle"
}

func (l *LifecycleService) Boundary() mail.BoundaryType {
	return mail.InnerBoundary
}

func (l *LifecycleService) HandleMail(mail mail.Mail) error {
	return nil
}

func (l *LifecycleService) Spawn(def statechart.ChartDefinition) (statechart.RuntimeID, error) {
	return statechart.RuntimeID("fake-runtime-id"), nil
}

func (l *LifecycleService) Stop(id statechart.RuntimeID) error {
	return nil
}

func (l *LifecycleService) List() ([]string, error) {
	return nil, nil
}

func (l *LifecycleService) Start() error {
	return nil
}
