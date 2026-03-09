package security

import (
	"fmt"
	"strings"
)

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
	switch SourceType(block.Source) {
	case SourceStatic:
		return assembleStatic(block)
	case SourceSession:
		return assembleSession(block, session)
	case SourceMemoryService:
		return assembleMemoryService(block, memorySvc)
	case SourceToolRegistry:
		return assembleToolRegistry(block, toolRegistry)
	default:
		return nil, fmt.Errorf("unsupported source type: %s", block.Source)
	}
}

func assembleStatic(block *ContextBlock) ([]byte, error) {
	return []byte(block.Content), nil
}

func assembleSession(block *ContextBlock, session *Session) ([]byte, error) {
	if session == nil || len(session.Messages) == 0 {
		return []byte(""), nil
	}

	n := block.N
	if n <= 0 {
		n = 10
	}

	messages := session.Messages
	if n >= len(messages) {
		n = len(messages)
	}

	lastN := messages[len(messages)-n:]

	var builder strings.Builder
	for i, msg := range lastN {
		if i > 0 {
			builder.WriteString("\n")
		}
		builder.WriteString(fmt.Sprintf("%s: %s", msg.Role, msg.Content))
	}

	return []byte(builder.String()), nil
}

func assembleMemoryService(block *ContextBlock, memorySvc MemoryService) ([]byte, error) {
	return nil, nil
}

func assembleToolRegistry(block *ContextBlock, toolRegistry *ToolRegistry) ([]byte, error) {
	return nil, nil
}
