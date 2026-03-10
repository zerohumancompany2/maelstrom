package heartbeat

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/maelstrom/v3/pkg/mail"
)

type InjectedContent struct {
	Type      string
	Timestamp time.Time
	Content   string
}

var NotImplementedError = errors.New("not implemented")

type heartbeatService struct {
	schedules       map[string]Schedule
	agentInboxes    map[string]*mail.AgentInbox
	publisher       mail.Publisher
	injectedContent map[string]InjectedContent
	wakeUpChannels  map[string]chan time.Time
	mu              sync.Mutex
}

func NewHeartbeatService() HeartbeatService {
	return &heartbeatService{
		schedules:       make(map[string]Schedule),
		agentInboxes:    make(map[string]*mail.AgentInbox),
		injectedContent: make(map[string]InjectedContent),
		wakeUpChannels:  make(map[string]chan time.Time),
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

func (s *heartbeatService) InjectHEARTBEAT(runtimeId string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.injectedContent[runtimeId] = InjectedContent{
		Type:      "HEARTBEAT",
		Timestamp: time.Now(),
		Content:   "HEARTBEAT.md",
	}

	return nil
}

func (s *heartbeatService) GetInjectedContent(runtimeId string) (InjectedContent, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	content, exists := s.injectedContent[runtimeId]
	if !exists {
		return InjectedContent{}, fmt.Errorf("no injected content for runtime: %s", runtimeId)
	}

	return content, nil
}

func (s *heartbeatService) NextWakeUp(runtimeId string, interval time.Duration) chan time.Time {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.wakeUpChannels[runtimeId]; !exists {
		ch := make(chan time.Time, 1)
		s.wakeUpChannels[runtimeId] = ch

		go func() {
			ticker := time.NewTicker(interval)
			defer ticker.Stop()
			select {
			case <-ticker.C:
				select {
				case ch <- time.Now():
				default:
				}
			case <-time.After(interval * 2):
			}
		}()
	}

	return s.wakeUpChannels[runtimeId]
}
