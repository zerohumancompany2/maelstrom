package gateway

import "net/http"

// ChannelAdapter interface defines the contract for channel adapters
type ChannelAdapter interface {
	Name() string
	Handle(r *http.Request) error
	Stream() bool
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
