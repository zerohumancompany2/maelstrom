package humangateway

import "github.com/maelstrom/v3/pkg/mail"

// SessionID represents a unique chat session identifier
type SessionID string

// ChatSession represents a chat session with an agent
type ChatSession struct {
	SessionID SessionID
	AgentID   string
	Messages  []string
	Active    bool
}

// StreamChunk is an alias for mail.StreamChunk
type StreamChunk = mail.StreamChunk
