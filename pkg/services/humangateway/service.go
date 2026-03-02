package humangateway

// HumanGatewayService interface defines the human gateway service API
type HumanGatewayService interface {
	OpenSession(agentId string) (SessionID, error)
	SendMessage(sessionId string, content string) error
	StreamResponse(sessionId string) (<-chan StreamChunk, error)
	CloseSession(sessionId string) error
	SessionExists(sessionID SessionID) bool
}

// humanGatewayService implements HumanGatewayService
type humanGatewayService struct {
	sessionMgr *sessionManager
}

// NewHumanGatewayService creates a new human gateway service instance
func NewHumanGatewayService() HumanGatewayService {
	return &humanGatewayService{
		sessionMgr: NewSessionManager(),
	}
}

// OpenSession opens a chat session for an agent
func (h *humanGatewayService) OpenSession(agentId string) (SessionID, error) {
	return h.sessionMgr.CreateSession(agentId), nil
}

// SendMessage sends a message to a session
func (h *humanGatewayService) SendMessage(sessionId string, content string) error {
	return nil
}

// StreamResponse streams a response from an agent
func (h *humanGatewayService) StreamResponse(sessionId string) (<-chan StreamChunk, error) {
	return nil, nil
}

// CloseSession closes a chat session
func (h *humanGatewayService) CloseSession(sessionId string) error {
	return nil
}

// SessionExists checks if a session exists
func (h *humanGatewayService) SessionExists(sessionID SessionID) bool {
	return h.sessionMgr.SessionExists(sessionID)
}
