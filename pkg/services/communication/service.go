package communication

import (
	"fmt"
	"sync"
	"time"

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

func (c *CommunicationService) Publish(m mail.Mail) (mail.Ack, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	target := m.Target
	if target == "" {
		target = m.Source
	}
	ack := mail.Ack{
		MailID:        m.ID,
		CorrelationID: m.CorrelationID,
		DeliveredAt:   time.Now(),
		Success:       false,
	}
	subscribers, exists := c.subscribers[target]
	if !exists || len(subscribers) == 0 {
		ack.ErrorMessage = "no subscribers"
		return ack, nil
	}
	for _, ch := range subscribers {
		select {
		case ch <- m:
			ack.Success = true
		default:
		}
	}
	return ack, nil
}

func (c *CommunicationService) Subscribe(address string) (chan mail.Mail, error) {
	ch := make(chan mail.Mail, 10)
	c.mu.Lock()
	c.subscribers[address] = append(c.subscribers[address], ch)
	c.mu.Unlock()
	return ch, nil
}

func (c *CommunicationService) Unsubscribe(address string, ch chan mail.Mail) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	subscribers, exists := c.subscribers[address]
	if !exists {
		return fmt.Errorf("no subscribers for address %s", address)
	}

	found := false
	newSubscribers := make([]chan mail.Mail, 0, len(subscribers))
	for _, subCh := range subscribers {
		if subCh == ch {
			found = true
			close(ch)
		} else {
			newSubscribers = append(newSubscribers, subCh)
		}
	}

	if !found {
		return fmt.Errorf("subscriber not found for address %s", address)
	}

	if len(newSubscribers) == 0 {
		delete(c.subscribers, address)
	} else {
		c.subscribers[address] = newSubscribers
	}

	return nil
}

func (c *CommunicationService) Start() error {
	return nil
}

func (c *CommunicationService) Stop() error {
	return nil
}
