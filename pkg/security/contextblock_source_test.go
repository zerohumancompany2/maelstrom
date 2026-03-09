package security

import (
	"fmt"
	"strings"
	"testing"

	"github.com/maelstrom/v3/pkg/mail"
)

func TestContextBlockSource_StaticContent(t *testing.T) {
	// Given: A ContextBlock with source: static and content field set to "You are a secure agent"
	block := &ContextBlock{
		Name:    "system-prompt",
		Source:  string(SourceStatic),
		Content: "You are a secure agent",
	}

	// When: AssembleSource is called on the block
	result, err := AssembleSource(block, nil, nil, nil)

	// Then: Block returns the static content exactly as configured
	if err != nil {
		t.Fatalf("AssembleSource returned error: %v", err)
	}

	expected := "You are a secure agent"
	if string(result) != expected {
		t.Errorf("Expected static content to be returned exactly, got: %s", string(result))
	}
}

func TestContextBlockSource_SessionLastN(t *testing.T) {
	// Given: A ContextBlock with source: session, strategy: lastN, n: 30 and session with 50 messages
	messages := make([]Message, 50)
	for i := 0; i < 50; i++ {
		messages[i] = Message{
			Role:    "user",
			Content: fmt.Sprintf("|||MSG-%03d|||", i),
		}
	}

	session := &Session{
		Messages: messages,
	}

	block := &ContextBlock{
		Name:     "conversation-history",
		Source:   string(SourceSession),
		Strategy: "lastN",
		N:        30,
	}

	// When: AssembleSource is called on the block
	result, err := AssembleSource(block, session, nil, nil)

	// Then: Block returns last 30 messages from session, in chronological order
	if err != nil {
		t.Fatalf("AssembleSource returned error: %v", err)
	}

	if len(result) == 0 {
		t.Fatal("Expected result to contain messages")
	}

	resultStr := string(result)
	for i := 20; i < 50; i++ {
		expectedMsg := fmt.Sprintf("|||MSG-%03d|||", i)
		if !strings.Contains(resultStr, expectedMsg) {
			t.Errorf("Expected result to contain last 30 messages including %s", expectedMsg)
			break
		}
	}

	for i := 0; i < 20; i++ {
		expectedMsg := fmt.Sprintf("|||MSG-%03d|||", i)
		if strings.Contains(resultStr, expectedMsg) {
			t.Errorf("Expected result to NOT contain first 20 messages including %s", expectedMsg)
			break
		}
	}
}

type mockMemoryService struct {
	results []MemoryResult
}

func (m *mockMemoryService) Query(query string, topK int) ([]MemoryResult, error) {
	if topK >= len(m.results) {
		return m.results, nil
	}
	return m.results[:topK], nil
}

func TestContextBlockSource_MemoryServiceRAG(t *testing.T) {
	// Given: A ContextBlock with source: memoryService, query: "{{currentGoal}}", topK: 8
	memorySvc := &mockMemoryService{
		results: make([]MemoryResult, 15),
	}
	for i := 0; i < 15; i++ {
		memorySvc.results[i] = MemoryResult{
			Content: fmt.Sprintf("|||MEMORY-%03d|||", i),
			Score:   float64(100 - i),
		}
	}

	block := &ContextBlock{
		Name:    "memory-context",
		Source:  string(SourceMemoryService),
		Content: "{{currentGoal}}",
		N:       8,
	}

	// When: AssembleSource is called with currentGoal = "implement feature X"
	result, err := AssembleSource(block, nil, memorySvc, nil)

	// Then: Block returns top 8 memory results matching the query
	if err != nil {
		t.Fatalf("AssembleSource returned error: %v", err)
	}

	if len(result) == 0 {
		t.Fatal("Expected result to contain memory results")
	}

	resultStr := string(result)
	for i := 0; i < 8; i++ {
		expectedMem := fmt.Sprintf("|||MEMORY-%03d|||", i)
		if !strings.Contains(resultStr, expectedMem) {
			t.Errorf("Expected result to contain top 8 memory results including %s", expectedMem)
			break
		}
	}

	for i := 8; i < 15; i++ {
		expectedMem := fmt.Sprintf("|||MEMORY-%03d|||", i)
		if strings.Contains(resultStr, expectedMem) {
			t.Errorf("Expected result to NOT contain memory results beyond topK including %s", expectedMem)
			break
		}
	}
}

func TestContextBlockSource_ToolRegistryBoundaryFilter(t *testing.T) {
	// Given: A ContextBlock with source: toolRegistry, boundaryFilter: dmz
	toolRegistry := NewToolRegistry()
	toolRegistry.RegisterTool(&ToolConfig{Name: "|||DMZ-TOOL-1|||", Boundary: mail.BoundaryType("dmz")})
	toolRegistry.RegisterTool(&ToolConfig{Name: "|||DMZ-TOOL-2|||", Boundary: mail.BoundaryType("dmz")})
	toolRegistry.RegisterTool(&ToolConfig{Name: "|||DMZ-TOOL-3|||", Boundary: mail.BoundaryType("dmz")})
	toolRegistry.RegisterTool(&ToolConfig{Name: "|||INNER-TOOL-1|||", Boundary: mail.BoundaryType("inner")})
	toolRegistry.RegisterTool(&ToolConfig{Name: "|||INNER-TOOL-2|||", Boundary: mail.BoundaryType("inner")})

	block := &ContextBlock{
		Name:           "tool-context",
		Source:         string(SourceToolRegistry),
		BoundaryFilter: DMZBoundary,
	}

	// When: AssembleSource is called at DMZ boundary
	result, err := AssembleSource(block, nil, nil, toolRegistry)

	// Then: Block returns only tools marked as visible at DMZ boundary
	if err != nil {
		t.Fatalf("AssembleSource returned error: %v", err)
	}

	if len(result) == 0 {
		t.Fatal("Expected result to contain tools")
	}

	resultStr := string(result)
	for i := 1; i <= 3; i++ {
		expectedTool := fmt.Sprintf("|||DMZ-TOOL-%d|||", i)
		if !strings.Contains(resultStr, expectedTool) {
			t.Errorf("Expected result to contain DMZ tools including %s", expectedTool)
			break
		}
	}

	for i := 1; i <= 2; i++ {
		expectedTool := fmt.Sprintf("|||INNER-TOOL-%d|||", i)
		if strings.Contains(resultStr, expectedTool) {
			t.Errorf("Expected result to NOT contain INNER tools including %s", expectedTool)
			break
		}
	}
}
