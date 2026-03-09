package communication

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/maelstrom/v3/pkg/mail"
	"github.com/maelstrom/v3/pkg/services/observability"
)

type pendingReply struct {
	ch        chan *mail.Mail
	createdAt time.Time
}

type CommunicationService struct {
	id               string
	router           *mail.MailRouter
	publisher        mail.Publisher
	subscribers      map[string][]chan mail.Mail
	deliveryAttempts map[string]int
	pendingReplies   map[string]pendingReply
	observability    *observability.ObservabilityService
	mu               sync.RWMutex
}

func NewCommunicationService() *CommunicationService {
	router := mail.NewMailRouter()
	return &CommunicationService{
		id:               "sys:communication",
		router:           router,
		publisher:        mail.NewRouterPublisher(router),
		subscribers:      make(map[string][]chan mail.Mail),
		deliveryAttempts: make(map[string]int),
		pendingReplies:   make(map[string]pendingReply),
	}
}

func (c *CommunicationService) ID() string {
	return c.id
}

func (c *CommunicationService) Publish(m mail.Mail) (mail.Ack, error) {
	target := m.Target
	if target == "" {
		target = m.Source
	}

	c.mu.RLock()
	localSubscribers, localExists := c.subscribers[target]
	c.mu.RUnlock()

	if localExists && len(localSubscribers) > 0 {
		ack := mail.Ack{
			MailID:        m.ID,
			CorrelationID: m.CorrelationID,
			DeliveredAt:   time.Now(),
			Success:       false,
		}
		for _, ch := range localSubscribers {
			select {
			case ch <- m:
				ack.Success = true
			default:
			}
		}
		return ack, nil
	}

	if mail.IsValidAgentAddress(target) || mail.IsValidTopicAddress(target) || mail.IsValidSysAddress(target) {
		return c.publisher.Publish(m)
	}

	return mail.Ack{
		MailID:        m.ID,
		CorrelationID: m.CorrelationID,
		DeliveredAt:   time.Now(),
		Success:       false,
		ErrorMessage:  "no subscribers",
	}, nil
}

func (c *CommunicationService) Subscribe(address string) (chan mail.Mail, error) {
	ch := make(chan mail.Mail, 100)

	if mail.IsValidAgentAddress(address) || mail.IsValidSysAddress(address) {
		inbox := &mail.AgentInbox{ID: address}
		c.router.SubscribeAgent(address, inbox)
	} else if mail.IsValidTopicAddress(address) {
		topic := &mail.Topic{Name: address}
		c.router.SubscribeTopic(address, topic)
	}

	c.mu.Lock()
	c.subscribers[address] = append(c.subscribers[address], ch)
	c.mu.Unlock()

	return ch, nil
}

func (c *CommunicationService) Unsubscribe(address string, stream chan mail.Mail) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	subs, exists := c.subscribers[address]
	if !exists {
		return errors.New("no subscribers for address: " + address)
	}

	for i, sub := range subs {
		if sub == stream {
			c.subscribers[address] = append(subs[:i], subs[i+1:]...)
			return nil
		}
	}

	return errors.New("stream not found")
}

func (c *CommunicationService) UpgradeToStream(sessionID string, lastEventID *string) (chan mail.StreamChunk, error) {
	return mail.UpgradeToStream(sessionID, lastEventID)
}

func (c *CommunicationService) HandleMail(m mail.Mail) error {
	if mail.IsValidAgentAddress(m.Target) || mail.IsValidTopicAddress(m.Target) || mail.IsValidSysAddress(m.Target) {
		return c.router.Route(m)
	}
	return nil
}

func (c *CommunicationService) Start() error {
	return nil
}

func (c *CommunicationService) Stop() error {
	return nil
}

func (c *CommunicationService) PublishWithRetry(mail *mail.Mail, maxRetries int) error {
	for attempt := 0; attempt <= maxRetries; attempt++ {
		ack, err := c.Publish(*mail)
		if err != nil {
			return err
		}
		if ack.Success {
			return nil
		}
		if attempt < maxRetries {
			backoff := time.Duration(1<<uint(attempt)) * time.Second
			time.Sleep(backoff)
		}
	}
	c.sendToDeadLetter(mail, "delivery failed after max retries")
	return errors.New("delivery failed after max retries")
}

func (c *CommunicationService) trackDeliveryAttempt(correlationID string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.deliveryAttempts[correlationID]++
}

func (c *CommunicationService) Request(replyChan chan *mail.Mail, timeout time.Duration) (*mail.Mail, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	select {
	case reply, ok := <-replyChan:
		if !ok {
			return nil, errors.New("reply channel closed")
		}
		return reply, nil
	case <-ctx.Done():
		return nil, errors.New("request timeout")
	}
}

func (c *CommunicationService) matchReply(correlationID string, mail *mail.Mail) bool {
	return mail.CorrelationID == correlationID
}

func (c *CommunicationService) SetObservability(obs *observability.ObservabilityService) {
	c.observability = obs
}

func (c *CommunicationService) sendToDeadLetter(mail *mail.Mail, reason string) {
	if c.observability != nil {
		c.observability.LogDeadLetter(*mail, reason)
	}
}
