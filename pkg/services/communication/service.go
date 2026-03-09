package communication

import (
	"sync"

	"github.com/maelstrom/v3/pkg/mail"
)

type CommunicationService struct {
	mu sync.Mutex
}

func NewCommunicationService() *CommunicationService {
	return &CommunicationService{}
}

func (c *CommunicationService) ID() string {
	return "sys:communication"
}

func (c *CommunicationService) HandleMail(mail mail.Mail) error {
	return nil
}

func (c *CommunicationService) Publish(mail mail.Mail) error {
	return nil
}

func (c *CommunicationService) Subscribe(address string) (<-chan mail.Mail, error) {
	return nil, nil
}

func (c *CommunicationService) Start() error {
	return nil
}

func (c *CommunicationService) Stop() error {
	return nil
}
