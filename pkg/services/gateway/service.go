package gateway

import (
	"encoding/json"
	"errors"

	"github.com/maelstrom/v3/pkg/mail"
	"github.com/maelstrom/v3/pkg/openapi"
)

// GatewayAck represents an acknowledgment of message delivery
type GatewayAck struct {
	MessageID string
	Status    string
}

// GatewayMail represents a message in the system
type GatewayMail struct {
	From     string
	To       []string
	Subject  string
	Body     string
	Taints   []string
	Metadata map[string]string
}

// GatewayService interface defines the gateway service API
type GatewayService interface {
	ID() string
	RegisterAdapter(name string, adapter ChannelAdapter) error
	Publish(mail GatewayMail) (GatewayAck, error)
	PublishTo(mail GatewayMail) error
	Subscribe(address string) (<-chan GatewayMail, error)
	Unsubscribe(address string, ch <-chan GatewayMail) error
	GetOpenAPI() (*openapi.Spec, error)
	GetAdapter(name string) (ChannelAdapter, bool)
	NormalizeInbound(adapterName string, rawMessage any) (*mail.Mail, error)
	NormalizeOutbound(mail *mail.Mail, adapterName string) (any, error)
}

// gatewayService implements GatewayService
type gatewayService struct {
	adapters map[string]ChannelAdapter
	mailChan chan GatewayMail
}

// NewGatewayService creates a new gateway service instance
func NewGatewayService() GatewayService {
	return &gatewayService{
		adapters: make(map[string]ChannelAdapter),
		mailChan: make(chan GatewayMail, 100),
	}
}

// RegisterAdapter registers a channel adapter
func (g *gatewayService) RegisterAdapter(name string, adapter ChannelAdapter) error {
	if _, exists := g.adapters[name]; exists {
		return errors.New("adapter already registered")
	}
	g.adapters[name] = adapter
	return nil
}

// Publish publishes a mail message
func (g *gatewayService) Publish(mail GatewayMail) (GatewayAck, error) {
	g.mailChan <- mail
	return GatewayAck{
		MessageID: mail.Subject,
		Status:    "published",
	}, nil
}

// Subscribe subscribes to messages at an address
func (g *gatewayService) Subscribe(address string) (<-chan GatewayMail, error) {
	return g.mailChan, nil
}

// PublishTo publishes a mail to a specific channel
func (g *gatewayService) PublishTo(mail GatewayMail) error {
	g.mailChan <- mail
	return nil
}

// Unsubscribe unsubscribes from an address
func (g *gatewayService) Unsubscribe(address string, ch <-chan GatewayMail) error {
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

// ID returns the service ID
func (g *gatewayService) ID() string {
	return "sys:gateway"
}

// NormalizeInbound normalizes inbound messages to mail_received
func (g *gatewayService) NormalizeInbound(adapterName string, rawMessage any) (*mail.Mail, error) {
	_, exists := g.adapters[adapterName]
	if !exists {
		return nil, errors.New("adapter not registered")
	}

	normalizedContent, err := normalizeContent(rawMessage)
	if err != nil {
		return nil, err
	}

	return &mail.Mail{
		Type:    mail.MailReceived,
		Content: normalizedContent,
		Metadata: mail.MailMetadata{
			Adapter: adapterName,
		},
	}, nil
}

func normalizeContent(rawMessage any) (string, error) {
	switch v := rawMessage.(type) {
	case string:
		return v, nil
	default:
		jsonBytes, err := json.Marshal(rawMessage)
		if err != nil {
			return "", err
		}
		return string(jsonBytes), nil
	}
}

// NormalizeOutbound normalizes outbound mail to channel-specific format
func (g *gatewayService) NormalizeOutbound(m *mail.Mail, adapterName string) (any, error) {
	result := map[string]any{
		"content":  m.Content,
		"boundary": string(m.Metadata.Boundary),
		"adapter":  adapterName,
	}

	boundary := m.Metadata.Boundary
	if boundary == mail.OuterBoundary {
		return stripSensitiveMetadata(result), nil
	}
	if boundary == mail.DMZBoundary {
		return allowLimitedMetadata(result), nil
	}
	if boundary == mail.InnerBoundary {
		return result, nil
	}
	return result, nil
}

func stripSensitiveMetadata(data map[string]any) map[string]any {
	result := make(map[string]any)
	for k, v := range data {
		if k != "tokens" && k != "internal" && k != "secret" {
			result[k] = v
		}
	}
	return result
}

func allowLimitedMetadata(data map[string]any) map[string]any {
	result := make(map[string]any)
	allowedKeys := []string{"content", "boundary", "adapter"}
	for _, key := range allowedKeys {
		if v, exists := data[key]; exists {
			result[key] = v
		}
	}
	return result
}

func (g *gatewayService) HandleMail(m mail.Mail) error {
	return nil
}

func (g *gatewayService) Start() error {
	return nil
}

func (g *gatewayService) Stop() error {
	return nil
}
