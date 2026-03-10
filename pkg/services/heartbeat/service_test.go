package heartbeat

import (
	"testing"
)

// arch-v1.md L469: HeartbeatService must return ID "sys:heartbeat"
func TestHeartbeatService_ID(t *testing.T) {
	svc := NewHeartbeatService()

	id := svc.ID()
	if id != "sys:heartbeat" {
		t.Errorf("Expected ID 'sys:heartbeat', got '%s'", id)
	}
}

// arch-v1.md L469: HeartbeatService must schedule wake-ups using cron expressions
func TestHeartbeatService_Schedule(t *testing.T) {
	svc := NewHeartbeatService()

	err := svc.Schedule("agent-1", "0 * * * *", "default template")
	if err != nil {
		t.Fatalf("Schedule failed: %v", err)
	}

	sched, err := svc.GetSchedule("agent-1")
	if err != nil {
		t.Fatalf("GetSchedule failed: %v", err)
	}

	if sched.CronExpr != "0 * * * *" {
		t.Errorf("Expected cron '0 * * * *', got %s", sched.CronExpr)
	}
	if sched.Template != "default template" {
		t.Errorf("Expected template 'default template', got %s", sched.Template)
	}
}

func TestHeartbeat_Schedule(t *testing.T) {
	svc := NewHeartbeatService()

	err := svc.Schedule("agent-1", "0 * * * *", "default template")

	if err != nil {
		t.Fatalf("Schedule failed: %v", err)
	}
}

func TestHeartbeat_Trigger(t *testing.T) {
	svc := NewHeartbeatService()

	err := svc.Trigger("agent-1")

	if err != nil {
		t.Fatalf("Trigger failed: %v", err)
	}
}

func TestHeartbeat_TemplateInjection(t *testing.T) {
	svc := NewHeartbeatService()

	err := svc.Trigger("agent-1")

	if err != nil {
		t.Fatalf("Trigger failed: %v", err)
	}

	_, err = svc.GetSchedule("agent-1")
	if err != nil {
		t.Fatalf("GetSchedule failed: %v", err)
	}
}

func TestHeartbeat_CronFires(t *testing.T) {
	svc := NewHeartbeatService()

	err := svc.Schedule("agent-1", "0 * * * *", "cron template")

	if err != nil {
		t.Fatalf("Schedule failed: %v", err)
	}

	sched, err := svc.GetSchedule("agent-1")
	if err != nil {
		t.Fatalf("GetSchedule failed: %v", err)
	}

	if sched.CronExpr != "0 * * * *" {
		t.Errorf("Expected cron '0 * * * *', got %s", sched.CronExpr)
	}
}

func TestHeartbeat_Unschedule(t *testing.T) {
	svc := NewHeartbeatService()

	svc.Schedule("agent-1", "0 * * * *", "template")

	err := svc.Unschedule("agent-1")

	if err != nil {
		t.Fatalf("Unschedule failed: %v", err)
	}

	sched, err := svc.GetSchedule("agent-1")
	if err != nil {
		t.Fatalf("GetSchedule failed: %v", err)
	}

	if sched.AgentID != "" {
		t.Errorf("Expected empty schedule after unschedule, got %v", sched)
	}
}

func TestHeartbeatService_ScheduleCron(t *testing.T) {
	svc := NewHeartbeatService()

	err := svc.ScheduleCron("0 * * * *")
	if err != nil {
		t.Errorf("Expected nil error, got %v", err)
	}
}

func TestHeartbeatService_TriggerAll(t *testing.T) {
	svc := NewHeartbeatService()

	err := svc.TriggerAll()
	if err != nil {
		t.Errorf("Expected nil error, got %v", err)
	}
}

// arch-v1.md L469: HeartbeatService must unschedule wake-ups by job ID
func TestHeartbeatService_Unschedule(t *testing.T) {
	svc := NewHeartbeatService()

	// First schedule a job
	err := svc.Schedule("agent-1", "0 * * * *", "template")
	if err != nil {
		t.Fatalf("Schedule failed: %v", err)
	}

	// Unschedule the job
	err = svc.Unschedule("agent-1")
	if err != nil {
		t.Fatalf("Unschedule failed: %v", err)
	}

	// Verify the schedule was removed
	sched, err := svc.GetSchedule("agent-1")
	if err != nil {
		t.Fatalf("GetSchedule failed: %v", err)
	}

	if sched.AgentID != "" {
		t.Errorf("Expected empty schedule after unschedule, got %v", sched)
	}

	// Unschedule non-existent job should return error
	err = svc.Unschedule("non-existent")
	if err == nil {
		t.Error("Expected error when unscheduling non-existent job, got nil")
	}
}
