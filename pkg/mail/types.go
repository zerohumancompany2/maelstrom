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
	User             MailType = "user"
	Assistant        MailType = "assistant"
	ToolResult       MailType = "tool_result"
	ToolCall         MailType = "tool_call"
	MailReceived     MailType = "mail_received"
	Heartbeat        MailType = "heartbeat"
	Error            MailType = "error"
	HumanFeedback    MailType = "human_feedback"
	PartialAssistant MailType = "partial_assistant"
	SubagentDone     MailType = "subagent_done"
	TaintViolation   MailType = "taint_violation"
)

type MailMetadata struct {
	Tokens   int
	Model    string
	Cost     float64
	Boundary BoundaryType
	Taints   []string
	Stream   bool
	IsFinal  bool
}

type BoundaryType string

const (
	InnerBoundary BoundaryType = "inner"
	DMZBoundary   BoundaryType = "dmz"
	OuterBoundary BoundaryType = "outer"
)

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
