package mail

import (
	"errors"
	"sync"
	"time"
)

type MailRouter struct {
	agents   map[string]*AgentInbox
	topics   map[string]*Topic
	services map[string]*ServiceInbox
	mu       sync.RWMutex
}

func NewMailRouter() *MailRouter {
	return &MailRouter{
		agents:   make(map[string]*AgentInbox),
		topics:   make(map[string]*Topic),
		services: make(map[string]*ServiceInbox),
	}
}

func (r *MailRouter) Route(mail Mail) error {
	addrType, id, err := ParseAddress(mail.Target)
	if err != nil {
		return err
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	switch addrType {
	case AddressTypeAgent:
		inbox, exists := r.agents[id]
		if !exists {
			return errors.New("agent not found: " + id)
		}
		return inbox.Push(mail)
	case AddressTypeTopic:
		topic, exists := r.topics[id]
		if !exists {
			return errors.New("topic not found: " + id)
		}
		return topic.Publish(mail)
	case AddressTypeSys:
		inbox, exists := r.services[id]
		if !exists {
			return errors.New("service not found: " + id)
		}
		return inbox.Push(mail)
	default:
		return errors.New("unknown address type")
	}
}

func (r *MailRouter) SubscribeAgent(id string, inbox *AgentInbox) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.agents[id] = inbox
	return nil
}

func (r *MailRouter) SubscribeTopic(name string, topic *Topic) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.topics[name] = topic
	return nil
}

func (r *MailRouter) SubscribeService(name string, inbox *ServiceInbox) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.services[name] = inbox
	return nil
}

type TopicSubscriber interface {
	Receive() chan Mail
	Subscribe(topic string) error
	Unsubscribe(topic string) error
}

type Topic struct {
	Name        string
	Subscribers []TopicSubscriber
	mu          sync.RWMutex
}

func (t *Topic) Publish(mail Mail) error {
	t.mu.RLock()
	defer t.mu.RUnlock()

	for _, sub := range t.Subscribers {
		ch := sub.Receive()
		select {
		case ch <- mail:
		default:
		}
	}
	return nil
}

func (t *Topic) Subscribe(sub TopicSubscriber) error {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.Subscribers = append(t.Subscribers, sub)
	return nil
}

func (t *Topic) Unsubscribe(sub TopicSubscriber) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	for i, s := range t.Subscribers {
		if s == sub {
			t.Subscribers = append(t.Subscribers[:i], t.Subscribers[i+1:]...)
			return nil
		}
	}

	return errors.New("subscriber not found")
}

type SecurityService interface {
	ValidateAndSanitize(mail any, src, tgt BoundaryType, allowedOnExit []string) (any, error)
	MarkTaint(obj any, taints []string) (any, error)
}

func (r *MailRouter) RouteWithSecurity(mail Mail, securityService SecurityService) error {
	var allowedOnExit []string
	sanitized, err := securityService.ValidateAndSanitize(mail, mail.Metadata.Boundary, mail.Metadata.Boundary, allowedOnExit)
	if err != nil {
		if sanitizedMail, ok := sanitized.(Mail); ok {
			r.routeViolation(sanitizedMail)
		}
		return err
	}
	if sanitizedMail, ok := sanitized.(Mail); ok {
		mail = sanitizedMail
	}
	return r.Route(mail)
}

func (r *MailRouter) routeViolation(mail Mail) error {
	violationMail := Mail{
		Type:   MailTypeTaintViolation,
		Source: "sys:security",
		Target: "sys:observability",
		Content: map[string]interface{}{
			"sourceBoundary":  string(mail.Metadata.Boundary),
			"targetBoundary":  string(mail.Metadata.Boundary),
			"forbiddenTaints": mail.Metadata.Taints,
			"timestamp":       time.Now(),
		},
	}
	return r.Route(violationMail)
}
