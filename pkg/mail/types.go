package mail

import "time"

type Mail struct {
	ID            string
	CorrelationID string
	Type          MailType
	CreatedAt     time.Time
	Source        string
	Target        string
	Content       any
	Metadata      MailMetadata
}

type MailType string

const (
	MailTypeUser             MailType = "user"
	MailTypeAssistant        MailType = "assistant"
	MailTypeToolResult       MailType = "tool_result"
	MailTypeToolCall         MailType = "tool_call"
	MailTypeMailReceived     MailType = "mail_received"
	MailTypeHeartbeat        MailType = "heartbeat"
	MailTypeError            MailType = "error"
	MailTypeHumanFeedback    MailType = "human_feedback"
	MailTypePartialAssistant MailType = "partial_assistant"
	MailTypeSubagentDone     MailType = "subagent_done"
	MailTypeTaintViolation   MailType = "taint_violation"

	// Aliases for backward compatibility
	User             = MailTypeUser
	Assistant        = MailTypeAssistant
	ToolResult       = MailTypeToolResult
	ToolCall         = MailTypeToolCall
	MailReceived     = MailTypeMailReceived
	Heartbeat        = MailTypeHeartbeat
	Error            = MailTypeError
	HumanFeedback    = MailTypeHumanFeedback
	PartialAssistant = MailTypePartialAssistant
	SubagentDone     = MailTypeSubagentDone
	TaintViolation   = MailTypeTaintViolation
)

type MailMetadata struct {
	Tokens      int
	Model       string
	Cost        float64
	Boundary    BoundaryType
	Taints      []string
	Stream      bool
	StreamChunk *StreamChunk
	IsFinal     bool
	Adapter     string
}

type BoundaryType string

const (
	InnerBoundary BoundaryType = "inner"
	DMZBoundary   BoundaryType = "dmz"
	OuterBoundary BoundaryType = "outer"
)

type StreamChunk struct {
	Data     string
	Sequence int
	IsFinal  bool
	Taints   []string
}

type Ack struct {
	MailID        string
	CorrelationID string
	DeliveredAt   time.Time
	Success       bool
	ErrorMessage  string
}

func isValidAddress(address string) bool {
	if address == "" {
		return false
	}
	prefixes := []string{"agent:", "topic:", "sys:"}
	for _, prefix := range prefixes {
		if len(address) >= len(prefix) && address[:len(prefix)] == prefix {
			return true
		}
	}
	return false
}
