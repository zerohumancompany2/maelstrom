package gateway

import (
	"github.com/maelstrom/v3/pkg/openapi"
)

// Ack represents an acknowledgment of message delivery
type Ack struct {
	MessageID string
	Status    string
}

// Mail represents a message in the system
type Mail struct {
	From     string
	To       []string
	Subject  string
	Body     string
	Taints   []string
	Metadata map[string]string
}

// GatewayService interface defines the gateway service API
type GatewayService interface {
	RegisterAdapter(name string, adapter ChannelAdapter) error
	Publish(mail Mail) (Ack, error)
	PublishTo(mail Mail) error
	Subscribe(address string) (<-chan Mail, error)
	Unsubscribe(address string, ch <-chan Mail) error
	GetOpenAPI() (*openapi.Spec, error)
	GetAdapter(name string) (ChannelAdapter, bool)
}

// gatewayService implements GatewayService
type gatewayService struct {
	adapters map[string]ChannelAdapter
	mailChan chan Mail
}

// NewGatewayService creates a new gateway service instance
func NewGatewayService() GatewayService {
	return &gatewayService{
		adapters: make(map[string]ChannelAdapter),
		mailChan: make(chan Mail, 100),
	}
}

// RegisterAdapter registers a channel adapter
func (g *gatewayService) RegisterAdapter(name string, adapter ChannelAdapter) error {
	g.adapters[name] = adapter
	return nil
}

// Publish publishes a mail message
func (g *gatewayService) Publish(mail Mail) (Ack, error) {
	g.mailChan <- mail
	return Ack{
		MessageID: mail.Subject,
		Status:    "published",
	}, nil
}

// Subscribe subscribes to messages at an address
func (g *gatewayService) Subscribe(address string) (<-chan Mail, error) {
	return g.mailChan, nil
}

// PublishTo publishes a mail to a specific channel
func (g *gatewayService) PublishTo(mail Mail) error {
	g.mailChan <- mail
	return nil
}

// Unsubscribe unsubscribes from an address
func (g *gatewayService) Unsubscribe(address string, ch <-chan Mail) error {
	return nil
}

// GetOpenAPI returns the OpenAPI specification
func (g *gatewayService) GetOpenAPI() (*openapi.Spec, error) {
	return &openapi.Spec{
		Version: "3.0.0",
		Info: openapi.Info{
			Title:   "Gateway Service API",
			Version: "1.0.0",
		},
		Paths: make(map[string]interface{}),
	}, nil
}

// GetAdapter returns a registered adapter by name
func (g *gatewayService) GetAdapter(name string) (ChannelAdapter, bool) {
	adapter, ok := g.adapters[name]
	return adapter, ok
}
