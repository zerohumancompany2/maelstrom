package humangateway

import "sync"

// sessionManager manages chat sessions
type sessionManager struct {
	sessions map[SessionID]*ChatSession
	mu       sync.RWMutex
}

// NewSessionManager creates a new session manager
func NewSessionManager() *sessionManager {
	return &sessionManager{
		sessions: make(map[SessionID]*ChatSession),
	}
}

// CreateSession creates a new chat session
func (m *sessionManager) CreateSession(agentID string) SessionID {
	id := SessionID("sess-" + agentID + "-" + string(rune(len(m.sessions))))
	m.sessions[id] = &ChatSession{
		SessionID: id,
		AgentID:   agentID,
		Messages:  make([]string, 0),
		Active:    true,
	}
	return id
}

// GetSession retrieves a session by ID
func (m *sessionManager) GetSession(sessionID SessionID) (*ChatSession, bool) {
	session, ok := m.sessions[sessionID]
	return session, ok
}

// SessionExists checks if a session exists
func (m *sessionManager) SessionExists(sessionID SessionID) bool {
	_, ok := m.sessions[sessionID]
	return ok
}

// CloseSession marks a session as closed
func (m *sessionManager) CloseSession(sessionID SessionID) bool {
	session, ok := m.sessions[sessionID]
	if ok {
		session.Active = false
	}
	return ok
}
