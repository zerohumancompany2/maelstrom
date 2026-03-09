package mail

import (
	"errors"
	"sync"
)

var ErrInboxEmpty = errors.New("inbox is empty")

type AgentInbox struct {
	ID       string
	Messages []Mail
	mu       sync.RWMutex
}

func (a *AgentInbox) Push(mail Mail) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.Messages = append(a.Messages, mail)
	return nil
}

func (a *AgentInbox) Pop() (Mail, error) {
	a.mu.Lock()
	defer a.mu.Unlock()

	if len(a.Messages) == 0 {
		return Mail{}, ErrInboxEmpty
	}

	mail := a.Messages[0]
	a.Messages = a.Messages[1:]
	return mail, nil
}

func (a *AgentInbox) Subscribe() <-chan Mail {
	ch := make(chan Mail, 100)
	return ch
}

type ServiceInbox struct {
	ID       string
	Messages []Mail
	mu       sync.RWMutex
}

func (s *ServiceInbox) Push(mail Mail) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Messages = append(s.Messages, mail)
	return nil
}

func (s *ServiceInbox) Pop() (Mail, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if len(s.Messages) == 0 {
		return Mail{}, ErrInboxEmpty
	}

	mail := s.Messages[0]
	s.Messages = s.Messages[1:]
	return mail, nil
}
