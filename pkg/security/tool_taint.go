package security

import (
	"github.com/maelstrom/v3/pkg/mail"
)

type ToolConfig struct {
	Name        string
	Boundary    mail.BoundaryType
	Isolation   string
	TaintOutput []string
}

type ToolRegistry struct {
	tools map[string]*ToolConfig
}

func NewToolRegistry() *ToolRegistry {
	return &ToolRegistry{
		tools: make(map[string]*ToolConfig),
	}
}

func (r *ToolRegistry) RegisterTool(tool *ToolConfig) {
	r.tools[tool.Name] = tool
}

func (r *ToolRegistry) GetTool(name string) (*ToolConfig, bool) {
	tool, ok := r.tools[name]
	return tool, ok
}

func AttachToolTaints(toolName string, result any, toolRegistry *ToolRegistry) (any, error) {
	tool, ok := toolRegistry.GetTool(toolName)
	if !ok {
		return result, nil
	}

	if len(tool.TaintOutput) == 0 {
		return result, nil
	}

	switch v := result.(type) {
	case *mail.Mail:
		return attachTaintsToMail(v, tool.TaintOutput, tool.Boundary), nil
	default:
		return result, nil
	}
}

func attachTaintsToMail(mail *mail.Mail, taints []string, boundary mail.BoundaryType) *mail.Mail {
	existing := make(map[string]bool)
	for _, t := range mail.Metadata.Taints {
		existing[t] = true
	}
	for _, t := range taints {
		existing[t] = true
	}
	merged := make([]string, 0, len(existing))
	for t := range existing {
		merged = append(merged, t)
	}
	mail.Metadata.Taints = merged
	mail.Metadata.Boundary = boundary
	return mail
}
