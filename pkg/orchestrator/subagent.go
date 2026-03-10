package orchestrator

import (
	"github.com/maelstrom/v3/pkg/mail"
	"github.com/maelstrom/v3/pkg/statechart"
)

type SubAgentType string

const (
	SubAgentAttached SubAgentType = "attached"
	SubAgentDetached SubAgentType = "detached"
)

type SubAgentConfig struct {
	Type           SubAgentType
	ChartRef       string
	MaxIterations  int
	InheritContext bool
	CorrelationId  string
}

type SubAgentExecutor struct {
	config        SubAgentConfig
	parentNs      string
	parentRuntime statechart.RuntimeID
	engine        statechart.Library
}

func NewSubAgentExecutor(config SubAgentConfig, parentNs string, parentRuntime statechart.RuntimeID, engine statechart.Library) *SubAgentExecutor {
	return &SubAgentExecutor{
		config:        config,
		parentNs:      parentNs,
		parentRuntime: parentRuntime,
		engine:        engine,
	}
}

func (e *SubAgentExecutor) Execute() (any, error) {
	panic("not implemented")
}

func (e *SubAgentExecutor) spawnAttached() (statechart.RuntimeID, error) {
	def := statechart.ChartDefinition{
		ID: e.config.ChartRef,
		Root: &statechart.Node{
			ID: "root",
		},
	}

	childID, err := e.engine.(interface {
		SpawnTransient(statechart.ChartDefinition, statechart.ApplicationContext, statechart.RuntimeID) (statechart.RuntimeID, error)
	}).SpawnTransient(def, nil, e.parentRuntime)
	if err != nil {
		return "", err
	}

	if err := e.setupAutoTermination(); err != nil {
		return "", err
	}

	return childID, nil
}

func (e *SubAgentExecutor) spawnDetached() (statechart.RuntimeID, error) {
	def := statechart.ChartDefinition{
		ID: e.config.ChartRef,
		Root: &statechart.Node{
			ID: "root",
		},
	}

	childID, err := e.engine.Spawn(def, nil)
	if err != nil {
		return "", err
	}

	return childID, nil
}

func (e *SubAgentExecutor) emitSubAgentDone(result any, message []mail.Mail) error {
	panic("not implemented")
}

func (e *SubAgentExecutor) setupAutoTermination() error {
	return nil
}
