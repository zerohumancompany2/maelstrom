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
	if l.engine == nil {
		return statechart.RuntimeID("fake-runtime-id"), nil
	}

	id, err := l.engine.Spawn(def, nil)
	if err != nil {
		return "", err
	}

	l.mu.Lock()
	l.runtimes[id] = RuntimeInfo{
		ID:           string(id),
		DefinitionID: def.ID,
		Boundary:     mail.InnerBoundary,
		ActiveStates: []string{def.InitialState},
		IsRunning:    false,
	}
	l.mu.Unlock()

	return id, nil
}

func (l *LifecycleService) Stop(id statechart.RuntimeID) error {
	if l.engine == nil {
		return nil
	}
	return l.engine.Control(id, statechart.CmdStop)
}

func (l *LifecycleService) Control(id statechart.RuntimeID, cmd statechart.ControlCmd) error {
	if l.engine == nil {
		return statechart.ErrRuntimeNotFound
	}
	return l.engine.Control(id, cmd)
}

func (l *LifecycleService) List() ([]RuntimeInfo, error) {
	l.mu.Lock()
	defer l.mu.Unlock()

	result := make([]RuntimeInfo, 0, len(l.runtimes))
	for _, info := range l.runtimes {
		result = append(result, info)
	}
	return result, nil
}

func (l *LifecycleService) Start() error {
	return nil
}
