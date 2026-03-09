package lifecycle

import (
	"sync"

	"github.com/maelstrom/v3/pkg/mail"
	"github.com/maelstrom/v3/pkg/statechart"
)

type LifecycleService struct {
	mu sync.Mutex
}

func NewLifecycleService() *LifecycleService {
	return &LifecycleService{}
}

func (l *LifecycleService) ID() string {
	return "sys:lifecycle"
}

func (l *LifecycleService) HandleMail(mail mail.Mail) error {
	return nil
}

func (l *LifecycleService) Spawn(def statechart.ChartDefinition) (statechart.RuntimeID, error) {
	return "", nil
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
