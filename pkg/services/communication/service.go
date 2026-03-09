package communication

import (
	"sync"

	"github.com/maelstrom/v3/pkg/mail"
)

type CommunicationService struct {
	mu          sync.Mutex
	subscribers map[string][]chan mail.Mail
}

func NewCommunicationService() *CommunicationService {
	return &CommunicationService{
		subscribers: make(map[string][]chan mail.Mail),
	}
}

func (c *CommunicationService) ID() string {
	return "sys:communication"
}

func (c *CommunicationService) HandleMail(mail mail.Mail) error {
	return nil
}

func (c *CommunicationService) Publish(mail mail.Mail) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	target := mail.Target
	if target == "" {
		target = mail.Source
	}
	for _, ch := range c.subscribers[target] {
		select {
		case ch <- mail:
		default:
		}
	}
	return nil
}

func (c *CommunicationService) Subscribe(address string) (<-chan mail.Mail, error) {
	ch := make(chan mail.Mail, 10)
	c.mu.Lock()
	c.subscribers[address] = append(c.subscribers[address], ch)
	c.mu.Unlock()
	return ch, nil
}

func (c *CommunicationService) Start() error {
	return nil
}

func (c *CommunicationService) Stop() error {
	return nil
}
