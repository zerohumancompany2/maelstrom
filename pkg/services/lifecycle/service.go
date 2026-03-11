package lifecycle

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"text/template"
	"time"

	"github.com/maelstrom/v3/pkg/mail"
	"github.com/maelstrom/v3/pkg/services"
	"github.com/maelstrom/v3/pkg/statechart"
)

type transformData struct {
	OldContext     any
	NewVersion     string
	ContextVersion string
}

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

func (l *LifecycleService) HandleMail(mail mail.Mail) *services.OutcomeEvent {
	return &services.OutcomeEvent{
		ServiceID: l.ID(),
		MailID:    mail.ID,
		Status:    "success",
		Timestamp: time.Now(),
	}
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
		l.mu.Lock()
		delete(l.runtimes, id)
		runtimeID := string(id)
		delete(l.stateHistory, runtimeID)
		delete(l.savedStates, runtimeID)
		l.mu.Unlock()
		return nil
	}
	err := l.engine.Control(id, statechart.CmdStop)
	if err == nil {
		l.mu.Lock()
		delete(l.runtimes, id)
		runtimeID := string(id)
		delete(l.stateHistory, runtimeID)
		delete(l.savedStates, runtimeID)
		l.mu.Unlock()
	}
	return err
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

func (l *LifecycleService) rollbackReload(serviceID string) error {
	l.mu.Lock()
	defer l.mu.Unlock()
	savedState, exists := l.savedStates[serviceID]
	if !exists {
		return statechart.ErrRuntimeNotFound
	}
	runtime, exists := l.runtimes[statechart.RuntimeID(serviceID)]
	if !exists {
		return statechart.ErrRuntimeNotFound
	}
	fromState := ""
	if len(runtime.ActiveStates) > 0 {
		fromState = runtime.ActiveStates[0]
	}
	runtime.ActiveStates = []string{savedState}
	l.runtimes[statechart.RuntimeID(serviceID)] = runtime
	l.stateHistory[serviceID] = append(l.stateHistory[serviceID], StateTransition{
		From:      fromState,
		To:        savedState,
		Timestamp: time.Now(),
	})
	return nil
}

func (l *LifecycleService) checkQuiescence(runtimeID string) (bool, error) {
	l.mu.Lock()
	defer l.mu.Unlock()
	_, exists := l.runtimes[statechart.RuntimeID(runtimeID)]
	if !exists {
		return false, statechart.ErrRuntimeNotFound
	}
	return true, nil
}

func (l *LifecycleService) prepareForReload(runtimeID string, timeoutMs int) error {
	l.mu.Lock()
	_, exists := l.runtimes[statechart.RuntimeID(runtimeID)]
	if !exists {
		l.mu.Unlock()
		return statechart.ErrRuntimeNotFound
	}
	l.mu.Unlock()

	timeout := time.Duration(timeoutMs) * time.Millisecond
	ticker := time.NewTicker(10 * time.Millisecond)
	defer ticker.Stop()

	done := make(chan bool)
	go func() {
		for {
			isQuiescent, err := l.checkQuiescence(runtimeID)
			if err == nil && isQuiescent {
				done <- true
				return
			}
		}
	}()

	select {
	case <-done:
		return nil
	case <-time.After(timeout):
		return statechart.ErrQuiescenceTimeout
	}
}

func (l *LifecycleService) restoreWithShallowHistory(snapshot statechart.Snapshot) (statechart.RuntimeID, error) {
	l.mu.Lock()
	defer l.mu.Unlock()

	id := statechart.RuntimeID(fmt.Sprintf("restored-%s-%d", snapshot.RuntimeID, len(l.runtimes)))
	runtimeID := string(id)

	l.runtimes[id] = RuntimeInfo{
		ID:           runtimeID,
		DefinitionID: snapshot.DefinitionID,
		Boundary:     mail.InnerBoundary,
		ActiveStates: snapshot.ActiveStates,
		IsRunning:    false,
	}

	l.stateHistory[runtimeID] = []StateTransition{
		{From: "", To: snapshot.ActiveStates[0], Timestamp: time.Now()},
	}

	return id, nil
}

func (l *LifecycleService) restoreWithDeepHistory(snapshot statechart.Snapshot, targetState string, def statechart.ChartDefinition) (statechart.RuntimeID, error) {
	l.mu.Lock()
	defer l.mu.Unlock()

	activeState := targetState
	if regionState, ok := snapshot.RegionStates[snapshot.ActiveStates[0]]; ok {
		activeState = regionState
	}

	fallbackToShallow := false
	if def.Root != nil {
		fullStatePath := snapshot.ActiveStates[0] + "/" + activeState
		if !l.stateExistsInDefinition(def, fullStatePath) {
			activeState = snapshot.ActiveStates[0]
			fallbackToShallow = true
		}
	}

	id := statechart.RuntimeID(fmt.Sprintf("restored-%s-%d", snapshot.RuntimeID, len(l.runtimes)))
	runtimeID := string(id)

	l.runtimes[id] = RuntimeInfo{
		ID:           runtimeID,
		DefinitionID: snapshot.DefinitionID,
		Boundary:     mail.InnerBoundary,
		ActiveStates: []string{activeState},
		IsRunning:    false,
	}

	l.stateHistory[runtimeID] = []StateTransition{
		{From: "", To: activeState, Timestamp: time.Now()},
	}

	if fallbackToShallow {
		fmt.Printf("Warning: state %s not found, falling back to shallow history\n", targetState)
	}

	return id, nil
}

func (l *LifecycleService) stateExistsInDefinition(def statechart.ChartDefinition, statePath string) bool {
	root := def.Root
	if root == nil {
		return false
	}

	parts := strings.Split(statePath, "/")
	if len(parts) == 0 {
		return false
	}

	current := root
	if parts[0] == root.ID {
		current = root
	} else {
		if child, ok := root.Children[parts[0]]; ok {
			current = child
		} else {
			return false
		}
	}

	for i := 1; i < len(parts); i++ {
		part := parts[i]
		if child, ok := current.Children[part]; ok {
			current = child
		} else {
			return false
		}
	}

	return true
}

func (l *LifecycleService) applyContextTransform(oldContext any, newVersion string, templateStr string) (any, error) {
	data := transformData{
		OldContext:     oldContext,
		NewVersion:     newVersion,
		ContextVersion: newVersion,
	}

	funcMap := template.FuncMap{
		"GetMapValue": func(m map[string]any, key string) any {
			if val, ok := m[key]; ok {
				return val
			}
			return ""
		},
	}

	tmpl, err := template.New("contextTransform").Funcs(funcMap).Parse(templateStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse template: %w", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return nil, fmt.Errorf("failed to execute template: %w", err)
	}

	var result map[string]any
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal result: %w", err)
	}

	return result, nil
}

func (l *LifecycleService) applyContextTransformWithFallback(oldContext any, newVersion string, templateStr string, cleanStart func() (any, error)) (any, error) {
	result, err := l.applyContextTransform(oldContext, newVersion, templateStr)
	if err != nil {
		fmt.Printf("Warning: transform failed, falling back to cleanStart: %v\n", err)
		return cleanStart()
	}
	return result, nil
}

func validateTransformTemplate(templateStr string) error {
	funcMap := template.FuncMap{
		"GetMapValue": func(m map[string]any, key string) any {
			if val, ok := m[key]; ok {
				return val
			}
			return ""
		},
	}

	_, err := template.New("contextTransform").Funcs(funcMap).Parse(templateStr)
	if err != nil {
		return fmt.Errorf("invalid template syntax: %w", err)
	}
	return nil
}

func (l *LifecycleService) validateTransformTemplate(templateStr string) error {
	return validateTransformTemplate(templateStr)
}
