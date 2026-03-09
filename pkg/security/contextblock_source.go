package security

type Session struct {
	Messages []Message
}

type Message struct {
	Role    string
	Content string
}

type MemoryService interface {
	Query(query string, topK int) ([]MemoryResult, error)
}

type MemoryResult struct {
	Content string
	Score   float64
}

type SourceType string

const (
	SourceStatic        SourceType = "static"
	SourceSession       SourceType = "session"
	SourceMemoryService SourceType = "memoryService"
	SourceToolRegistry  SourceType = "toolRegistry"
	SourceRuntime       SourceType = "runtime"
)

func AssembleSource(block *ContextBlock, session *Session, memorySvc MemoryService, toolRegistry *ToolRegistry) ([]byte, error) {
	if block.Source == string(SourceStatic) {
		return assembleStatic(block)
	}
	return nil, nil
}

func assembleStatic(block *ContextBlock) ([]byte, error) {
	return []byte(block.Content), nil
}

func assembleSession(block *ContextBlock, session *Session) ([]byte, error) {
	return nil, nil
}

func assembleMemoryService(block *ContextBlock, memorySvc MemoryService) ([]byte, error) {
	return nil, nil
}

func assembleToolRegistry(block *ContextBlock, toolRegistry *ToolRegistry) ([]byte, error) {
	return nil, nil
}
