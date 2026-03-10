package tools

import (
	"errors"

	"github.com/maelstrom/v3/pkg/security"
)

var (
	ErrToolNotFound      = errors.New("tool not found")
	ErrToolNotAccessible = errors.New("tool not accessible from this boundary")
	ErrBoundaryViolation = errors.New("boundary violation: unauthorized tool access")
	ErrDuplicateTool     = errors.New("tool with this name already registered")
)

type ToolsService interface {
	Register(tool ToolDescriptor) error
	Resolve(name string, callerBoundary string) (ToolDescriptor, error)
	List(boundaryFilter string) ([]ToolDescriptor, error)
	Invoke(toolName string, args map[string]any, callerBoundary string) (any, error)
	Unregister(name string) error
	RegisterByName(name string, tool ToolDescriptor) error
	ResolveByName(name string) (ToolDescriptor, error)
}

type ToolDescriptor struct {
	Name      string
	Boundary  string
	Schema    map[string]any
	Isolation string
}

type toolsService struct {
	registry map[string]ToolDescriptor
}

func NewToolsService() ToolsService {
	return &toolsService{
		registry: make(map[string]ToolDescriptor),
	}
}

func (s *toolsService) Register(tool ToolDescriptor) error {
	if _, exists := s.registry[tool.Name]; exists {
		return ErrDuplicateTool
	}
	s.registry[tool.Name] = tool
	return nil
}

func (s *toolsService) Resolve(name string, callerBoundary string) (ToolDescriptor, error) {
	tool, ok := s.registry[name]
	if !ok {
		return ToolDescriptor{}, ErrToolNotFound
	}

	if !canAccessTool(tool.Boundary, callerBoundary) {
		return ToolDescriptor{}, ErrToolNotAccessible
	}

	return tool, nil
}

func canAccessTool(toolBoundary, callerBoundary string) bool {
	toolBound := security.BoundaryType(toolBoundary)
	callerBound := security.BoundaryType(callerBoundary)

	switch callerBound {
	case security.InnerBoundary:
		return toolBound == security.InnerBoundary || toolBound == security.DMZBoundary || toolBound == security.OuterBoundary
	case security.DMZBoundary:
		return toolBound == security.DMZBoundary || toolBound == security.OuterBoundary
	case security.OuterBoundary:
		return toolBound == security.OuterBoundary
	default:
		return false
	}
}

func (s *toolsService) List(boundaryFilter string) ([]ToolDescriptor, error) {
	if boundaryFilter == "" {
		result := make([]ToolDescriptor, 0, len(s.registry))
		for _, tool := range s.registry {
			result = append(result, tool)
		}
		return result, nil
	}

	var result []ToolDescriptor
	for _, tool := range s.registry {
		if tool.Boundary == boundaryFilter {
			result = append(result, tool)
		}
	}
	return result, nil
}

func (s *toolsService) Invoke(toolName string, args map[string]any, callerBoundary string) (any, error) {
	tool, ok := s.registry[toolName]
	if !ok {
		return nil, nil
	}

	return map[string]any{
		"tool":      tool.Name,
		"args":      args,
		"boundary":  callerBoundary,
		"isolation": tool.Isolation,
		"status":    "executed",
	}, nil
}

func (s *toolsService) Unregister(name string) error {
	delete(s.registry, name)
	return nil
}

func (s *toolsService) RegisterByName(name string, tool ToolDescriptor) error {
	s.registry[name] = tool
	return nil
}

func (s *toolsService) ResolveByName(name string) (ToolDescriptor, error) {
	tool, ok := s.registry[name]
	if !ok {
		var zero ToolDescriptor
		return zero, nil
	}
	return tool, nil
}
