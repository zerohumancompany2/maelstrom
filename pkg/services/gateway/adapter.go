package gateway

import (
	"net/http"

	"github.com/maelstrom/v3/pkg/mail"
)

// ChannelAdapter interface defines the contract for channel adapters
type ChannelAdapter interface {
	Name() string
	Handle(r *http.Request) error
	Stream() bool
	NormalizeInbound(rawMessage any) (*mail.Mail, error)
	NormalizeOutbound(mail *mail.Mail) (any, error)
}

// WebhookAdapter implements HTTP webhook channel adapter
type WebhookAdapter struct {
	// Add fields as needed
}

func (w *WebhookAdapter) Name() string {
	return "webhook"
}

func (w *WebhookAdapter) Handle(r *http.Request) error {
	return nil
}

func (w *WebhookAdapter) Stream() bool {
	return false
}

func (w *WebhookAdapter) NormalizeInbound(rawMessage any) (*mail.Mail, error) {
	return &mail.Mail{
		Type:    mail.MailReceived,
		Content: rawMessage,
		Metadata: mail.MailMetadata{
			Adapter: "webhook",
		},
	}, nil
}

func (w *WebhookAdapter) NormalizeOutbound(mail *mail.Mail) (any, error) {
	return mail.Content, nil
}

// WebSocketAdapter implements bidirectional WebSocket channel adapter
type WebSocketAdapter struct {
	// Add fields as needed
}

func (ws *WebSocketAdapter) Name() string {
	return "websocket"
}

func (ws *WebSocketAdapter) Handle(r *http.Request) error {
	return nil
}

func (ws *WebSocketAdapter) Stream() bool {
	return true
}

func (ws *WebSocketAdapter) NormalizeInbound(rawMessage any) (*mail.Mail, error) {
	return &mail.Mail{
		Type:    mail.MailReceived,
		Content: rawMessage,
		Metadata: mail.MailMetadata{
			Adapter: "websocket",
		},
	}, nil
}

func (ws *WebSocketAdapter) NormalizeOutbound(mail *mail.Mail) (any, error) {
	return mail.Content, nil
}

// SSEAdapter implements Server-Sent Events channel adapter
type SSEAdapter struct {
	// Add fields as needed
}

func (sse *SSEAdapter) Name() string {
	return "sse"
}

func (sse *SSEAdapter) Handle(r *http.Request) error {
	return nil
}

func (sse *SSEAdapter) Stream() bool {
	return true
}

func (sse *SSEAdapter) NormalizeInbound(rawMessage any) (*mail.Mail, error) {
	return &mail.Mail{
		Type:    mail.MailReceived,
		Content: rawMessage,
		Metadata: mail.MailMetadata{
			Adapter: "sse",
		},
	}, nil
}

func (sse *SSEAdapter) NormalizeOutbound(mail *mail.Mail) (any, error) {
	return mail.Content, nil
}

// SMTPAdapter implements email/SMTP channel adapter
type SMTPAdapter struct {
	// Add fields as needed
}

func (s *SMTPAdapter) Name() string {
	return "smtp"
}

func (s *SMTPAdapter) Handle(r *http.Request) error {
	return nil
}

func (s *SMTPAdapter) Stream() bool {
	return false
}

func (s *SMTPAdapter) NormalizeInbound(rawMessage any) (*mail.Mail, error) {
	return &mail.Mail{
		Type:    mail.MailReceived,
		Content: rawMessage,
		Metadata: mail.MailMetadata{
			Adapter: "smtp",
		},
	}, nil
}

func (s *SMTPAdapter) NormalizeOutbound(mail *mail.Mail) (any, error) {
	return mail.Content, nil
}

// InternalGRPCAdapter implements internal gRPC channel adapter
type InternalGRPCAdapter struct {
	// Add fields as needed
}

func (g *InternalGRPCAdapter) Name() string {
	return "grpc"
}

func (g *InternalGRPCAdapter) Handle(r *http.Request) error {
	return nil
}

func (g *InternalGRPCAdapter) Stream() bool {
	return false
}

func (g *InternalGRPCAdapter) NormalizeInbound(rawMessage any) (*mail.Mail, error) {
	return &mail.Mail{
		Type:    mail.MailReceived,
		Content: rawMessage,
		Metadata: mail.MailMetadata{
			Adapter: "grpc",
		},
	}, nil
}

func (g *InternalGRPCAdapter) NormalizeOutbound(mail *mail.Mail) (any, error) {
	return mail.Content, nil
}
