package humangateway

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/maelstrom/v3/pkg/mail"
	"sync"
)

type HumanGatewayService struct {
	id       string
	sessions map[string]*ChatSession
	mu       sync.RWMutex
}

func NewHumanGatewayService() *HumanGatewayService {
	return &HumanGatewayService{
		id:       "sys:human-gateway",
		sessions: make(map[string]*ChatSession),
	}
}

func (h *HumanGatewayService) ID() string {
	return h.id
}

func (h *HumanGatewayService) HandleChat(agentID, message string) (mail.Mail, error) {
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

func (h *HumanGatewayService) GetSession(agentID string) *ChatSession {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.sessions[agentID]
}

func (h *HumanGatewayService) CreateSession(agentID string) *ChatSession {
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

func (h *HumanGatewayService) ChatEndpoint(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status": "ok",
	})
}
