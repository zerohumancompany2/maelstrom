package humangateway

import (
	"crypto/rand"
	"fmt"
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
	panic("not implemented")
}

func (h *HumanGatewayService) CreateSession(agentID string) *ChatSession {
	panic("not implemented")
}
