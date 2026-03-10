package heartbeat

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/maelstrom/v3/pkg/mail"
)

var NotImplementedError = errors.New("not implemented")

type heartbeatService struct {
	schedules    map[string]Schedule
	agentInboxes map[string]*mail.AgentInbox
	publisher    mail.Publisher
	mu           sync.Mutex
}

func NewHeartbeatService() HeartbeatService {
	return &heartbeatService{
		schedules:    make(map[string]Schedule),
		agentInboxes: make(map[string]*mail.AgentInbox),
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

func (s *heartbeatService) ID() string {
	return "sys:heartbeat"
}

func (s *heartbeatService) Unschedule(agentId string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, exists := s.schedules[agentId]; !exists {
		return fmt.Errorf("schedule not found for agent: %s", agentId)
	}
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

func (s *heartbeatService) TriggerWakeUp(agentID string) error {
	s.mu.Lock()
	_, exists := s.schedules[agentID]
	if !exists {
		s.mu.Unlock()
		// Log failed wake-up - arch-v1.md L469
		return fmt.Errorf("no schedule found for agent: %s", agentID)
	}
	s.mu.Unlock()

	s.mu.Lock()
	inbox, exists := s.agentInboxes[agentID]
	if !exists {
		inbox = &mail.AgentInbox{ID: agentID}
		s.agentInboxes[agentID] = inbox
	}
	s.mu.Unlock()

	// HEARTBEAT.md injection - arch-v1.md L469
	heartbeatMail := mail.Mail{
		ID:        fmt.Sprintf("heartbeat-%s-%d", agentID, time.Now().UnixNano()),
		Type:      mail.MailTypeHeartbeat,
		Source:    "sys:heartbeat",
		Target:    agentID,
		Content:   "HEARTBEAT.md",
		CreatedAt: time.Now(),
		Metadata: mail.MailMetadata{
			Boundary: mail.InnerBoundary,
			Taints:   []string{"SYSTEM"},
		},
	}

	// Deliver HEARTBEAT.md to agent inbox
	return inbox.Push(heartbeatMail)
}

func (s *heartbeatService) GetInbox(agentID string) *mail.AgentInbox {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.agentInboxes[agentID]
}

func (s *heartbeatService) HandleMail(m mail.Mail) error {
	return nil
}

func (s *heartbeatService) Start() error {
	return nil
}

func (s *heartbeatService) Stop() error {
	return nil
}
