package humangateway

// SessionID represents a unique chat session identifier
type SessionID string

// StreamChunk represents a chunk of streamed response data
type StreamChunk struct {
	Data     string
	Sequence int
	IsFinal  bool
	Taints   []string
}

// ChatSession represents a chat session with an agent
type ChatSession struct {
	SessionID SessionID
	AgentID   string
	Messages  []string
	Active    bool
}
