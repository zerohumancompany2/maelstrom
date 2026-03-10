package orchestrator

import (
	"sync"

	"github.com/maelstrom/v3/pkg/security"
	"github.com/maelstrom/v3/pkg/services/tools"
	"github.com/maelstrom/v3/pkg/statechart"
)

type OrchestratorService struct {
	id           string
	registry     *security.ToolRegistry
	library      statechart.Library
	policies     map[string]*ExecutionPolicy
	toolsService tools.ToolsService
	mu           sync.RWMutex
}

func NewOrchestratorService() *OrchestratorService {
	return &OrchestratorService{
		id:           "sys:orchestrator",
		registry:     security.NewToolRegistry(),
		policies:     make(map[string]*ExecutionPolicy),
		toolsService: tools.NewToolsService(),
	}
}

func NewOrchestratorServiceWithTools(toolsService tools.ToolsService) *OrchestratorService {
	return &OrchestratorService{
		id:           "sys:orchestrator",
		registry:     security.NewToolRegistry(),
		policies:     make(map[string]*ExecutionPolicy),
		toolsService: toolsService,
	}
}

func (s *OrchestratorService) ID() string {
	return s.id
}

func (s *OrchestratorService) RegisterPolicy(name string, policy ExecutionPolicy) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.policies[name] = &policy
	return nil
}

func (s *OrchestratorService) ResolveTool(name string, callerBoundary string) (tools.ToolDescriptor, error) {
	return s.toolsService.Resolve(name, callerBoundary)
}

func (s *OrchestratorService) Execute(toolCalls []ToolCall, policy ExecutionPolicy) (statechart.RuntimeID, error) {
	// TODO: Implement execution
	return "", nil
}
