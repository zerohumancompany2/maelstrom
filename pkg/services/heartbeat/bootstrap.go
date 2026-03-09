package heartbeat

import "github.com/maelstrom/v3/pkg/statechart"

// Schedule represents a heartbeat schedule
type Schedule struct {
	AgentID  string
	CronExpr string
	Template string
}

// HeartbeatService interface for heartbeat operations
type HeartbeatService interface {
	Schedule(agentId string, cron string, template string) error
	Unschedule(agentId string) error
	Trigger(agentId string) error
	GetSchedule(agentId string) (Schedule, error)
	ScheduleCron(cron string) error
	TriggerAll() error
}

// BootstrapChart returns the chart definition for sys:heartbeat
func BootstrapChart() statechart.ChartDefinition {
	return statechart.ChartDefinition{
		ID:      "sys:heartbeat",
		Version: "1.0.0",
	}
}
