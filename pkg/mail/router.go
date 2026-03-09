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
	// TODO: implement
	return errors.New("not implemented")
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

// Stub types for TDD workflow
type AgentInbox struct {
	ID       string
	Messages []Mail
	mu       sync.RWMutex
}

func (ai *AgentInbox) Push(mail Mail) error {
	// TODO: implement
	return nil
}

type Topic struct {
	Name        string
	Subscribers []chan Mail
	mu          sync.RWMutex
}

func (t *Topic) Publish(mail Mail) error {
	// TODO: implement
	return nil
}

type ServiceInbox struct {
	ID       string
	Messages []Mail
	mu       sync.RWMutex
}

func (si *ServiceInbox) Push(mail Mail) error {
	// TODO: implement
	return nil
}
