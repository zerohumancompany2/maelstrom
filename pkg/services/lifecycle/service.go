package lifecycle

import (
	"fmt"
	"sync"
	"time"

	"github.com/maelstrom/v3/pkg/mail"
	"github.com/maelstrom/v3/pkg/statechart"
)

type LifecycleService struct {
	mu           sync.Mutex
	engine       statechart.Library
	runtimes     map[statechart.RuntimeID]RuntimeInfo
	stateHistory map[string][]StateTransition
	savedStates  map[string]string
}

func NewLifecycleService(engine statechart.Library) *LifecycleService {
	return &LifecycleService{
		engine:       engine,
		runtimes:     make(map[statechart.RuntimeID]RuntimeInfo),
		stateHistory: make(map[string][]StateTransition),
		savedStates:  make(map[string]string),
	}
}

func NewLifecycleServiceWithoutEngine() *LifecycleService {
	return &LifecycleService{
		runtimes:     make(map[statechart.RuntimeID]RuntimeInfo),
		stateHistory: make(map[string][]StateTransition),
		savedStates:  make(map[string]string),
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
		l.mu.Lock()
		id := statechart.RuntimeID(fmt.Sprintf("fake-runtime-%d", len(l.runtimes)))
		runtimeID := string(id)
		l.runtimes[id] = RuntimeInfo{
			ID:           runtimeID,
			DefinitionID: def.ID,
			Boundary:     mail.InnerBoundary,
			ActiveStates: []string{def.InitialState},
			IsRunning:    false,
		}
		l.stateHistory[runtimeID] = []StateTransition{
			{From: "", To: def.InitialState, Timestamp: time.Now()},
		}
		l.mu.Unlock()
		return id, nil
	}

	id, err := l.engine.Spawn(def, nil)
	if err != nil {
		return "", err
	}

	l.mu.Lock()
	runtimeID := string(id)
	l.runtimes[id] = RuntimeInfo{
		ID:           runtimeID,
		DefinitionID: def.ID,
		Boundary:     mail.InnerBoundary,
		ActiveStates: []string{def.InitialState},
		IsRunning:    false,
	}
	l.stateHistory[runtimeID] = []StateTransition{
		{From: "", To: def.InitialState, Timestamp: time.Now()},
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

func (l *LifecycleService) updateRuntimeState(runtimeID, state string) error {
	l.mu.Lock()
	defer l.mu.Unlock()
	for id, info := range l.runtimes {
		if info.ID == runtimeID {
			fromState := ""
			if len(info.ActiveStates) > 0 {
				fromState = info.ActiveStates[0]
			}
			info.ActiveStates = []string{state}
			l.runtimes[id] = info
			l.stateHistory[runtimeID] = append(l.stateHistory[runtimeID], StateTransition{
				From:      fromState,
				To:        state,
				Timestamp: time.Now(),
			})
			return nil
		}
	}
	return statechart.ErrRuntimeNotFound
}

func (l *LifecycleService) getStateHistory(runtimeID string) []StateTransition {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.stateHistory[runtimeID]
}

func (l *LifecycleService) HotReload(serviceID string) error {
	l.mu.Lock()
	defer l.mu.Unlock()
	_, exists := l.runtimes[statechart.RuntimeID(serviceID)]
	if !exists {
		return statechart.ErrRuntimeNotFound
	}
	return nil
}

func (l *LifecycleService) preserveState(serviceID string) error {
	l.mu.Lock()
	defer l.mu.Unlock()
	runtime, exists := l.runtimes[statechart.RuntimeID(serviceID)]
	if !exists {
		return statechart.ErrRuntimeNotFound
	}
	if len(runtime.ActiveStates) > 0 {
		l.savedStates[serviceID] = runtime.ActiveStates[0]
	}
	return nil
}

func (l *LifecycleService) getSavedState(serviceID string) string {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.savedStates[serviceID]
}
