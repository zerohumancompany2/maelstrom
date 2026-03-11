package gateway

import (
	"slices"
	"strings"
	"testing"

	"github.com/maelstrom/v3/pkg/mail"
)

func TestLayer8Integration_FullStreamingPath(t *testing.T) {
	svc := NewGatewayService()
	sessionID := "test-session-001"
	agentID := "agent:dmz"

	// User types → Gateway (SSE connect) → Communication.publish(mail: type=user, taints=["USER_SUPPLIED"], boundary=outer) (arch-v1.md L675)
	userInput := "Hello, agent! What can you do?"
	m, err := svc.HandleUserInput(userInput, sessionID)
	if err != nil {
		t.Fatalf("Expected no error handling user input, got %v", err)
	}

	// Verify mail_received created (arch-v1.md L675)
	if m.Type != mail.MailReceived {
		t.Errorf("Expected mail type 'mail_received', got '%s'", m.Type)
	}

	// Verify taints=["USER_SUPPLIED"] (arch-v1.md L675)
	if !slices.Contains(m.Taints, "USER_SUPPLIED") {
		t.Error("Expected USER_SUPPLIED taint")
	}

	// Verify boundary=outer (arch-v1.md L675)
	if m.Metadata.Boundary != mail.OuterBoundary {
		t.Errorf("Expected boundary 'outer', got '%s'", m.Metadata.Boundary)
	}

	// Verify delivered to DMZ Agent inbox (arch-v1.md L676)
	if m.Target != agentID {
		t.Errorf("Expected target '%s', got '%s'", agentID, m.Target)
	}

	// Engine.dispatchEvent → enters LLMReason state (arch-v1.md L677)
	state, err := svc.DispatchEvent(agentID, *m)
	if err != nil {
		t.Fatalf("Expected no error dispatching event, got %v", err)
	}

	// Verify LLMReason state entered (arch-v1.md L677)
	if state.Name != "LLMReason" {
		t.Errorf("Expected state 'LLMReason', got '%s'", state.Name)
	}

	// AgentExtensions.assembleContextMap (pulls session + Memory blocks; taints propagated) → LLM call (arch-v1.md L678)
	contextMap, err := svc.AssembleContextMap(agentID)
	if err != nil {
		t.Fatalf("Expected no error assembling context map, got %v", err)
	}

	// Verify contextMap includes session data (arch-v1.md L678)
	if contextMap.Session == nil {
		t.Error("Expected contextMap to include session")
	}

	// Verify taints propagated through context assembly (arch-v1.md L678)
	if len(contextMap.Taints) == 0 {
		t.Error("Expected contextMap to have propagated taints")
	}

	// LLM returns tool calls → Orchestrator (isolated sub-chart) → Tool Registry.resolve → executeIsolated (new namespace) (arch-v1.md L679)
	toolCall := ToolCall{
		Name:      "get_weather",
		Arguments: map[string]any{"location": "San Francisco"},
	}

	toolResult, err := svc.ExecuteTool(toolCall, "isolated-namespace-001")
	if err != nil {
		t.Fatalf("Expected no error executing tool, got %v", err)
	}

	// Tool result → taintPropagate → new tool_result Mail (taints=["TOOL_OUTPUT"]) injected to session (arch-v1.md L680)
	if toolResult.Type != mail.ToolResult {
		t.Errorf("Expected mail type 'tool_result', got '%s'", toolResult.Type)
	}

	// Verify taints=["TOOL_OUTPUT"] (arch-v1.md L680)
	if !slices.Contains(toolResult.Taints, "TOOL_OUTPUT") {
		t.Error("Expected TOOL_OUTPUT taint on tool result")
	}

	// Engine emits partial_assistant Messages (stream: true) → Communication → Gateway → Security.stripForbiddenTaints → SSE chunks to user (arch-v1.md L681)
	assistantResponse := "The weather in San Francisco is currently sunny with a high of 72°F."
	chunk, err := svc.EmitPartialAssistant(assistantResponse, 1)
	if err != nil {
		t.Fatalf("Expected no error emitting partial assistant, got %v", err)
	}

	// Verify StreamChunk format (arch-v1.md L696-701)
	if chunk.Chunk != assistantResponse {
		t.Errorf("Expected chunk content '%s', got '%s'", assistantResponse, chunk.Chunk)
	}

	// Verify sequence number
	if chunk.Sequence != 1 {
		t.Errorf("Expected sequence 1, got %d", chunk.Sequence)
	}

	// Security.stripForbiddenTaints before emission (arch-v1.md L681)
	strippedChunk, err := svc.StripForbiddenTaints(chunk)
	if err != nil {
		t.Fatalf("Expected no error stripping taints, got %v", err)
	}

	// Verify forbidden taints removed (arch-v1.md L681)
	if slices.Contains(strippedChunk.Taints, "SECRET") {
		t.Error("Expected SECRET taint to be stripped")
	}

	if slices.Contains(strippedChunk.Taints, "INNER_ONLY") {
		t.Error("Expected INNER_ONLY taint to be stripped")
	}

	// SSE chunks to user (arch-v1.md L681)
	sseData, err := svc.FormatSSEChunk(strippedChunk)
	if err != nil {
		t.Fatalf("Expected no error formatting SSE chunk, got %v", err)
	}

	// Verify SSE format (arch-v1.md L662)
	if !strings.HasPrefix(sseData, "data: ") {
		t.Errorf("Expected SSE format starting with 'data: ', got '%s'", sseData[:min(20, len(sseData))])
	}

	// Verify full streaming path completed
	t.Logf("Full streaming path completed: user → Gateway → mail → LLM → StreamChunk → SSE")
}

func TestLayer8Integration_HumanChatWithRunningAgent(t *testing.T) {
	// Stub - will be implemented in next TDD iteration
}

func TestLayer8Integration_ChannelAdapterToMailToStream(t *testing.T) {
	// Stub - will be implemented in next TDD iteration
}

func TestLayer8Integration_SecurityEnforcedThroughout(t *testing.T) {
	// Stub - will be implemented in next TDD iteration
}
