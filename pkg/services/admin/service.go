package admin

import "github.com/maelstrom/v3/pkg/statechart"

type AgentInfo struct {
	ID           string
	State        string
	ActiveStates []string
	Boundary     string
}

type TaintMap map[string][]string

type AdminService interface {
	ListAgents() ([]AgentInfo, error)
	ControlAgent(id string, cmd string) error
	QueryTaints(agentId string) (TaintMap, error)
	InjectEvent(agentId string, event statechart.Event) error
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
	panic("Not implemented")
}

func (s *adminService) QueryTaints(agentId string) (TaintMap, error) {
	panic("Not implemented")
}

func (s *adminService) InjectEvent(agentId string, event statechart.Event) error {
	panic("Not implemented")
}
