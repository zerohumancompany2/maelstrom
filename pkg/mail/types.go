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
	Taints        []string
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
	MailTypeMailSend         MailType = "mail_send"
	MailTypeContextBlock     MailType = "context_block"
	MailTypeSnapshot         MailType = "snapshot"
	MailTypeKernelReady      MailType = "kernel_ready"

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
	MailSend         = MailTypeMailSend
	ContextBlock     = MailTypeContextBlock
	Snapshot         = MailTypeSnapshot
	KernelReady      = MailTypeKernelReady
)

type MailMetadata struct {
	Tokens            int
	Model             string
	Cost              float64
	Boundary          BoundaryType
	Taints            []string
	Stream            bool
	StreamChunk       *StreamChunk
	IsFinal           bool
	Adapter           string
	ActionItem        ActionItem
	HumanFeedbackType string
}

// ActionItem represents a parsed action item from chat input
type ActionItem struct {
	Type    string
	Payload any
}

type BoundaryType string

const (
	InnerBoundary BoundaryType = "inner"
	DMZBoundary   BoundaryType = "dmz"
	OuterBoundary BoundaryType = "outer"
)

type StreamChunk struct {
	Data        string   `json:"data"`
	Chunk       string   `json:"chunk"`
	Sequence    int      `json:"sequence"`
	IsFinal     bool     `json:"isFinal"`
	Taints      []string `json:"taints"`
	MessageType string   `json:"messageType"`
}

type Ack struct {
	MailID        string
	CorrelationID string
	DeliveredAt   time.Time
	Success       bool
	ErrorMessage  string
}

// GetTaints returns the taints associated with the mail
func (m *Mail) GetTaints() []string {
	if m.Taints != nil {
		return m.Taints
	}
	return m.Metadata.Taints
}

// PropagateTaints propagates taints from source mail to target mail (arch-v1.md L283)
func PropagateTaints(sourceMail *Mail, targetMail *Mail) {
	if sourceMail == nil || targetMail == nil {
		return
	}

	sourceTaints := sourceMail.GetTaints()
	if len(sourceTaints) == 0 {
		return
	}

	seen := make(map[string]bool)
	for _, t := range targetMail.Taints {
		seen[t] = true
	}
	for _, t := range targetMail.Metadata.Taints {
		seen[t] = true
	}

	for _, t := range sourceTaints {
		if !seen[t] {
			targetMail.Taints = append(targetMail.Taints, t)
			seen[t] = true
		}
	}

	targetMail.Metadata.Taints = targetMail.Taints
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
