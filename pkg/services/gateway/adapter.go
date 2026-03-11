package gateway

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"net/http"
	"time"

	"github.com/maelstrom/v3/pkg/mail"
)

const (
	TaintUserSupplied = "USER_SUPPLIED"
	TaintExternal     = "EXTERNAL"
)

func generateID() string {
	b := make([]byte, 16)
	rand.Read(b)
	return hex.EncodeToString(b)
}

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
	port int
}

// NewWebhookAdapter creates a new webhook adapter with configurable port
func NewWebhookAdapter(port int) *WebhookAdapter {
	return &WebhookAdapter{port: port}
}

func (w *WebhookAdapter) Name() string {
	return "webhook"
}

func (w *WebhookAdapter) Handle(r *http.Request) error {
	if r.Method != http.MethodPost {
		return errors.New("webhook adapter only accepts POST requests")
	}
	return nil
}

func (w *WebhookAdapter) Stream() bool {
	return false
}

func (w *WebhookAdapter) NormalizeInbound(rawMessage any) (*mail.Mail, error) {
	if rawMessage == nil {
		return nil, errors.New("rawMessage cannot be nil")
	}

	return &mail.Mail{
		ID:            generateID(),
		CorrelationID: generateID(),
		Type:          mail.MailReceived,
		CreatedAt:     time.Now(),
		Source:        "gateway",
		Content:       rawMessage,
		Taints:        []string{TaintUserSupplied, TaintExternal},
		Metadata: mail.MailMetadata{
			Adapter:  "webhook",
			Boundary: mail.OuterBoundary,
			Taints:   []string{TaintUserSupplied, TaintExternal},
		},
	}, nil
}

func (w *WebhookAdapter) NormalizeOutbound(mail *mail.Mail) (any, error) {
	if mail == nil {
		return nil, errors.New("mail cannot be nil")
	}
	return mail.Content, nil
}

// WebSocketAdapter implements bidirectional WebSocket channel adapter
type WebSocketAdapter struct {
	port int
}

// NewWebSocketAdapter creates a new WebSocket adapter with configurable port
func NewWebSocketAdapter(port int) *WebSocketAdapter {
	return &WebSocketAdapter{port: port}
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
	if rawMessage == nil {
		return nil, errors.New("rawMessage cannot be nil")
	}

	return &mail.Mail{
		ID:            generateID(),
		CorrelationID: generateID(),
		Type:          mail.MailReceived,
		CreatedAt:     time.Now(),
		Source:        "gateway",
		Content:       rawMessage,
		Taints:        []string{TaintUserSupplied, TaintExternal},
		Metadata: mail.MailMetadata{
			Adapter:  "websocket",
			Boundary: mail.OuterBoundary,
			Taints:   []string{TaintUserSupplied, TaintExternal},
			Stream:   true,
		},
	}, nil
}

func (ws *WebSocketAdapter) NormalizeOutbound(mail *mail.Mail) (any, error) {
	if mail == nil {
		return nil, errors.New("mail cannot be nil")
	}
	return mail.Content, nil
}

// SSEAdapter implements Server-Sent Events channel adapter
type SSEAdapter struct {
	port int
}

// NewSSEAdapter creates a new SSE adapter with configurable port
func NewSSEAdapter(port int) *SSEAdapter {
	return &SSEAdapter{port: port}
}

func (sse *SSEAdapter) Name() string {
	return "sse"
}

func (sse *SSEAdapter) Handle(r *http.Request) error {
	if r.Method != http.MethodGet {
		return errors.New("SSE adapter only accepts GET requests")
	}
	return nil
}

func (sse *SSEAdapter) Stream() bool {
	return true
}

func (sse *SSEAdapter) NormalizeInbound(rawMessage any) (*mail.Mail, error) {
	if rawMessage == nil {
		return nil, errors.New("rawMessage cannot be nil")
	}

	return &mail.Mail{
		ID:            generateID(),
		CorrelationID: generateID(),
		Type:          mail.MailReceived,
		CreatedAt:     time.Now(),
		Source:        "gateway",
		Content:       rawMessage,
		Taints:        []string{TaintUserSupplied, TaintExternal},
		Metadata: mail.MailMetadata{
			Adapter:  "sse",
			Boundary: mail.OuterBoundary,
			Taints:   []string{TaintUserSupplied, TaintExternal},
			Stream:   true,
		},
	}, nil
}

func (sse *SSEAdapter) NormalizeOutbound(mail *mail.Mail) (any, error) {
	if mail == nil {
		return nil, errors.New("mail cannot be nil")
	}

	sseEvent := map[string]any{
		"event": mail.Type,
		"data":  mail.Content,
	}

	if mail.Metadata.Stream {
		if chunk := mail.Metadata.StreamChunk; chunk != nil {
			sseEvent["chunk"] = chunk.Chunk
			sseEvent["sequence"] = chunk.Sequence
			sseEvent["isFinal"] = chunk.IsFinal
		}
	}

	return sseEvent, nil
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
