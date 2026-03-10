package admin

import (
	"fmt"

	"github.com/maelstrom/v3/pkg/security"
	"github.com/maelstrom/v3/pkg/services/lifecycle"
	"github.com/maelstrom/v3/pkg/statechart"
)

type AdminService interface {
	ID() string
	ListAgents() ([]lifecycle.RuntimeInfo, error)
	ControlAgent(id string, cmd string) error
	QueryTaints(agentId string) (security.TaintMap, error)
	InjectEvent(agentId string, event statechart.Event) error
	ExecuteCommand(cmd string, token string) error
}

type adminService struct {
	lifecycleService *lifecycle.LifecycleService
	taintEngine      security.TaintEngine
}

func NewAdminService() AdminService {
	return &adminService{
		lifecycleService: lifecycle.NewLifecycleServiceWithoutEngine(),
		taintEngine:      security.NewTaintEngine(),
	}
}

func (s *adminService) ID() string {
	return "sys:admin"
}

func (s *adminService) ListAgents() ([]lifecycle.RuntimeInfo, error) {
	return s.lifecycleService.List()
}

func (s *adminService) ControlAgent(id string, action string) error {
	var cmd statechart.ControlCmd
	switch action {
	case "pause":
		cmd = statechart.CmdPause
	case "resume":
		cmd = statechart.CmdResume
	case "stop":
		cmd = statechart.CmdStop
	default:
		return fmt.Errorf("unknown action: %s", action)
	}

	agents, err := s.lifecycleService.List()
	if err != nil {
		return err
	}

	found := false
	for _, agent := range agents {
		if agent.ID == id {
			found = true
			break
		}
	}

	if !found {
		return statechart.ErrRuntimeNotFound
	}

	return s.lifecycleService.Control(statechart.RuntimeID(id), cmd)
}

func (s *adminService) QueryTaints(agentId string) (security.TaintMap, error) {
	return s.taintEngine.ReportTaints(agentId)
}

func (s *adminService) InjectEvent(agentId string, event statechart.Event) error {
	return nil
}

func (s *adminService) ExecuteCommand(cmd string, token string) error {
	if token == "" {
		return fmt.Errorf("2FA token required")
	}
	return nil
}
