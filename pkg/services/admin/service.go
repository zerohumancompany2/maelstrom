package admin

import (
	"fmt"

	"github.com/maelstrom/v3/pkg/security"
	"github.com/maelstrom/v3/pkg/statechart"
)

type AgentInfo struct {
	ID           string
	State        string
	ActiveStates []string
	Boundary     string
}

type AdminService interface {
	ListAgents() ([]AgentInfo, error)
	ControlAgent(id string, cmd string) error
	QueryTaints(agentId string) (security.TaintMap, error)
	InjectEvent(agentId string, event statechart.Event) error
	ExecuteCommand(cmd string, token string) error
}

type adminService struct {
	agents map[string]AgentInfo
}

func NewAdminService() AdminService {
	return &adminService{
		agents: make(map[string]AgentInfo),
	}
}

func (s *adminService) ListAgents() ([]AgentInfo, error) {
	return []AgentInfo{}, nil
}

func (s *adminService) ControlAgent(id string, cmd string) error {
	return nil
}

func (s *adminService) QueryTaints(agentId string) (security.TaintMap, error) {
	return security.TaintMap{}, nil
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
