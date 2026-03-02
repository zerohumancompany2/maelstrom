package heartbeat

import "errors"

var NotImplementedError = errors.New("not implemented")

type heartbeatService struct{}

func NewHeartbeatService() HeartbeatService {
	return &heartbeatService{}
}

func (s *heartbeatService) Schedule(agentId string, cron string, template string) error {
	return nil
}

func (s *heartbeatService) Unschedule(agentId string) error {
	return NotImplementedError
}

func (s *heartbeatService) Trigger(agentId string) error {
	return nil
}

func (s *heartbeatService) GetSchedule(agentId string) (Schedule, error) {
	return Schedule{}, nil
}
