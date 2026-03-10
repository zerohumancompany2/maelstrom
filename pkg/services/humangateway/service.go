package humangateway

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/maelstrom/v3/pkg/mail"
	"sync"
)

// HumanGatewayService interface for human gateway operations
type HumanGatewayService interface {
	ID() string
	HandleChat(agentID, message string) (mail.Mail, error)
	GetSession(agentID string) *ChatSession
	CreateSession(agentID string) *ChatSession
	ChatEndpoint(w http.ResponseWriter, r *http.Request)
	CreateChatSession(agentID string) (*ChatSession, error)
	HandleMail(mail mail.Mail) error
	Start() error
	Stop() error
}

type humanGatewayService struct {
	id       string
	sessions map[string]*ChatSession
	mu       sync.RWMutex
}

func NewHumanGatewayService() HumanGatewayService {
	return &humanGatewayService{
		id:       "sys:human-gateway",
		sessions: make(map[string]*ChatSession),
	}
}

func (h *humanGatewayService) ID() string {
	return h.id
}

func (h *humanGatewayService) HandleChat(agentID, message string) (mail.Mail, error) {
	b := make([]byte, 8)
	rand.Read(b)
	id := fmt.Sprintf("human-gateway-%x", b)

	actionItems, _ := h.ParseActionItem(message)

	mailType := mail.MailTypeUser
	if len(actionItems) > 0 {
		mailType = mail.MailTypeHumanFeedback
	}

	return mail.Mail{
		ID:     id,
		Type:   mailType,
		Source: "human:" + agentID,
		Target: "agent:" + agentID,
		Content: map[string]any{
			"message":     message,
			"actionItems": actionItems,
		},
		Metadata: mail.MailMetadata{
			Boundary: mail.InnerBoundary,
			Taints:   []string{"USER_SUPPLIED"},
		},
	}, nil
}

func (h *humanGatewayService) GetSession(agentID string) *ChatSession {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.sessions[agentID]
}

func (h *humanGatewayService) CreateSession(agentID string) *ChatSession {
	h.mu.Lock()
	defer h.mu.Unlock()

	session := &ChatSession{
		AgentID:    agentID,
		Messages:   make([]mail.Mail, 0),
		ContextMap: make(ContextMapSnapshot),
	}
	h.sessions[agentID] = session
	return session
}

func (h *humanGatewayService) ChatEndpoint(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status": "ok",
	})
}

func (h *humanGatewayService) CreateChatSession(agentID string) (*ChatSession, error) {
	h.mu.Lock()
	defer h.mu.Unlock()

	session := &ChatSession{
		AgentID:    agentID,
		Messages:   make([]mail.Mail, 0),
		ContextMap: make(ContextMapSnapshot),
	}
	h.sessions[agentID] = session
	return session, nil
}

func (h *humanGatewayService) HandleMail(m mail.Mail) error {
	return nil
}

func (h *humanGatewayService) Start() error {
	return nil
}

func (h *humanGatewayService) Stop() error {
	return nil
}
