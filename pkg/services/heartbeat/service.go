package heartbeat

import (
	"errors"
	"sync"
)

var NotImplementedError = errors.New("not implemented")

type heartbeatService struct {
	schedules map[string]Schedule
	mu        sync.Mutex
}

func NewHeartbeatService() HeartbeatService {
	return &heartbeatService{
		schedules: make(map[string]Schedule),
	}
}

func (s *heartbeatService) Schedule(agentId string, cron string, template string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.schedules[agentId] = Schedule{
		AgentID:  agentId,
		CronExpr: cron,
		Template: template,
	}
	return nil
}

func (s *heartbeatService) Unschedule(agentId string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.schedules, agentId)
	return nil
}

func (s *heartbeatService) Trigger(agentId string) error {
	return nil
}

func (s *heartbeatService) GetSchedule(agentId string) (Schedule, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.schedules[agentId], nil
}

func (s *heartbeatService) ScheduleCron(cron string) error {
	return nil
}

func (s *heartbeatService) TriggerAll() error {
	return nil
}
