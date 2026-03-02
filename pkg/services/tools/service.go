package tools

type ToolsService interface {
	Register(tool ToolDescriptor) error
	Resolve(name string, callerBoundary string) (ToolDescriptor, error)
	List(boundaryFilter string) ([]ToolDescriptor, error)
	Invoke(toolName string, args map[string]any, callerBoundary string) (any, error)
	Unregister(name string) error
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
	s.registry[tool.Name] = tool
	return nil
}

func (s *toolsService) Resolve(name string, callerBoundary string) (ToolDescriptor, error) {
	tool, ok := s.registry[name]
	if !ok {
		var zero ToolDescriptor
		return zero, nil
	}
	return tool, nil
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
	return nil, nil
}

func (s *toolsService) Unregister(name string) error {
	return nil
}
