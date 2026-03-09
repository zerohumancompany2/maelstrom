package mail

import (
	"errors"
	"sync"
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

type Topic struct {
	Name        string
	Subscribers []chan Mail
	mu          sync.RWMutex
}

func (t *Topic) Publish(mail Mail) error {
	return nil
}
