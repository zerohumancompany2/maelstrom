package gateway

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/maelstrom/v3/pkg/mail"
)

func TestStreamingPath_UserInputToMail(t *testing.T) {
	svc := NewGatewayService()
	sessionID := "test-session-001"
	userInput := "Hello, agent!"

	// User input creates mail_received Mail (arch-v1.md L675)
	m, err := svc.HandleUserInput(userInput, sessionID)
	if err != nil {
		t.Fatalf("Expected no error handling user input, got %v", err)
	}

	// Verify mail type (arch-v1.md L675)
	if m.Type != "mail_received" {
		t.Errorf("Expected mail type 'mail_received', got '%s'", m.Type)
	}

	// Verify taints=["USER_SUPPLIED"] (arch-v1.md L675)
	if len(m.Metadata.Taints) != 1 {
		t.Errorf("Expected 1 taint, got %d", len(m.Metadata.Taints))
	}
	if m.Metadata.Taints[0] != "USER_SUPPLIED" {
		t.Errorf("Expected taint 'USER_SUPPLIED', got '%s'", m.Metadata.Taints[0])
	}

	// Verify boundary=outer (arch-v1.md L675)
	if m.Metadata.Boundary != "outer" {
		t.Errorf("Expected boundary 'outer', got '%s'", m.Metadata.Boundary)
	}

	// Verify delivered to DMZ Agent inbox via Communication.publish
	if m.Target != "agent:dmz" {
		t.Errorf("Expected target 'agent:dmz', got '%s'", m.Target)
	}

	// Verify content matches user input
	if m.Content != userInput {
		t.Errorf("Expected content '%s', got '%s'", userInput, m.Content)
	}
}

func TestStreamingPath_MailToLLMStream(t *testing.T) {
	svc := NewGatewayService()
	agentID := "test-agent-001"

	// Setup: Create mail_received in agent inbox
	m := mail.Mail{
		ID:      "mail-001",
		Type:    "mail_received",
		Source:  "user",
		Target:  "agent:" + agentID,
		Content: "Process this request",
		Metadata: mail.MailMetadata{
			Taints:   []string{"USER_SUPPLIED"},
			Boundary: "outer",
		},
	}

	// Engine.dispatchEvent enters LLMReason state (arch-v1.md L677)
	state, err := svc.DispatchEvent(agentID, m)
	if err != nil {
		t.Fatalf("Expected no error dispatching event, got %v", err)
	}

	// Verify LLMReason state entered (arch-v1.md L677)
	if state.Name != "LLMReason" {
		t.Errorf("Expected state 'LLMReason', got '%s'", state.Name)
	}

	// AgentExtensions.assembleContextMap pulls session + Memory blocks (arch-v1.md L678)
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

	// Verify streaming enabled for LLM call
	if !contextMap.StreamingEnabled {
		t.Error("Expected streaming enabled for LLM call")
	}
}

func TestStreamingPath_LLMAssistantToStreamChunk(t *testing.T) {
	svc := NewGatewayService()

	// LLM returns assistant response
	llmResponse := "This is the assistant response to your query."

	// Engine emits partial_assistant Messages (stream: true) (arch-v1.md L681)
	chunk, err := svc.EmitPartialAssistant(llmResponse, 1)
	if err != nil {
		t.Fatalf("Expected no error emitting partial assistant, got %v", err)
	}

	// Verify StreamChunk format (arch-v1.md L696-701)
	if chunk.Chunk != llmResponse {
		t.Errorf("Expected chunk content '%s', got '%s'", llmResponse, chunk.Chunk)
	}

	// Verify sequence number
	if chunk.Sequence != 1 {
		t.Errorf("Expected sequence 1, got %d", chunk.Sequence)
	}

	// Verify isFinal flag for single chunk
	if !chunk.IsFinal {
		t.Error("Expected isFinal to be true for single chunk")
	}

	// Verify taints array exists (arch-v1.md L696-701)
	if chunk.Taints == nil {
		t.Error("Expected taints array to exist")
	}

	// Verify partial_assistant message type
	if chunk.MessageType != "partial_assistant" {
		t.Errorf("Expected message type 'partial_assistant', got '%s'", chunk.MessageType)
	}
}

func TestStreamingPath_StreamChunkTaintStripping(t *testing.T) {
	svc := NewGatewayService()

	// Create StreamChunk with forbidden taints
	chunk := &mail.StreamChunk{
		Chunk:    "This contains SECRET and PII data",
		Sequence: 1,
		IsFinal:  false,
		Taints:   []string{"USER_SUPPLIED", "SECRET", "PII", "TOOL_OUTPUT"},
	}

	// Security.stripForbiddenTaints applied before emission (arch-v1.md L681)
	strippedChunk, err := svc.StripForbiddenTaints(chunk)
	if err != nil {
		t.Fatalf("Expected no error stripping taints, got %v", err)
	}

	// Verify forbidden taints removed (arch-v1.md L681)
	if contains(strippedChunk.Taints, "SECRET") {
		t.Error("Expected SECRET taint to be stripped")
	}
	if contains(strippedChunk.Taints, "PII") {
		t.Error("Expected PII taint to be stripped")
	}

	// Verify allowed taints preserved
	if !contains(strippedChunk.Taints, "USER_SUPPLIED") {
		t.Error("Expected USER_SUPPLIED taint to be preserved")
	}
	if !contains(strippedChunk.Taints, "TOOL_OUTPUT") {
		t.Error("Expected TOOL_OUTPUT taint to be preserved")
	}

	// Verify chunk content unchanged (only taints stripped)
	if strippedChunk.Chunk != chunk.Chunk {
		t.Errorf("Expected chunk content unchanged, got '%s'", strippedChunk.Chunk)
	}

	// Verify sequence and isFinal preserved
	if strippedChunk.Sequence != chunk.Sequence {
		t.Errorf("Expected sequence %d, got %d", chunk.Sequence, strippedChunk.Sequence)
	}
	if strippedChunk.IsFinal != chunk.IsFinal {
		t.Errorf("Expected isFinal %v, got %v", chunk.IsFinal, strippedChunk.IsFinal)
	}
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func TestStreamingPath_SSEChunkFormat(t *testing.T) {
	svc := NewGatewayService()

	// Create StreamChunk for SSE output
	chunk := &mail.StreamChunk{
		Chunk:    "Assistant response part 1",
		Sequence: 1,
		IsFinal:  false,
		Taints:   []string{"USER_SUPPLIED"},
	}

	// Format as SSE chunk (arch-v1.md L696-701)
	sseData, err := svc.FormatSSEChunk(chunk)
	if err != nil {
		t.Fatalf("Expected no error formatting SSE chunk, got %v", err)
	}

	// Verify SSE format: data: <json>\n\n
	expectedPrefix := "data: "
	if !strings.HasPrefix(sseData, expectedPrefix) {
		t.Errorf("Expected SSE format starting with '%s', got '%s'", expectedPrefix, sseData[:min(20, len(sseData))])
	}

	// Verify JSON payload can be parsed
	jsonStr := strings.TrimPrefix(sseData, "data: ")
	var parsed map[string]any
	err = json.Unmarshal([]byte(jsonStr), &parsed)
	if err != nil {
		t.Fatalf("Expected valid JSON in SSE data, got %v", err)
	}

	// Verify chunk field in JSON
	if parsed["chunk"] != "Assistant response part 1" {
		t.Error("Expected chunk field in SSE JSON")
	}

	// Verify sequence field in JSON
	if parsed["sequence"].(float64) != 1 {
		t.Error("Expected sequence field in SSE JSON")
	}

	// Verify isFinal field in JSON
	if parsed["isFinal"] != false {
		t.Error("Expected isFinal field in SSE JSON")
	}

	// Verify taints stripped from SSE output (arch-v1.md L696-701)
	if _, exists := parsed["taints"]; exists {
		t.Error("Expected taints to be stripped from SSE output")
	}
}

func TestStreamingPath_WebSocketChunkFormat(t *testing.T) {
	svc := NewGatewayService()

	// Create StreamChunk for WebSocket output
	chunk := &mail.StreamChunk{
		Chunk:    "Assistant response part 1",
		Sequence: 1,
		IsFinal:  false,
		Taints:   []string{"USER_SUPPLIED"},
	}

	// Format as WebSocket chunk (arch-v1.md L696-701)
	wsData, err := svc.FormatWebSocketChunk(chunk)
	if err != nil {
		t.Fatalf("Expected no error formatting WebSocket chunk, got %v", err)
	}

	// Verify WebSocket format: raw JSON
	var parsed map[string]any
	err = json.Unmarshal(wsData, &parsed)
	if err != nil {
		t.Fatalf("Expected valid JSON in WebSocket data, got %v", err)
	}

	// Verify chunk field in JSON
	if parsed["chunk"] != "Assistant response part 1" {
		t.Error("Expected chunk field in WebSocket JSON")
	}

	// Verify sequence field in JSON
	if parsed["sequence"].(float64) != 1 {
		t.Error("Expected sequence field in WebSocket JSON")
	}

	// Verify isFinal field in JSON
	if parsed["isFinal"] != false {
		t.Error("Expected isFinal field in WebSocket JSON")
	}

	// Verify taints stripped from WebSocket output (arch-v1.md L696-701)
	if _, exists := parsed["taints"]; exists {
		t.Error("Expected taints to be stripped from WebSocket output")
	}

	// Test final chunk format
	finalChunk := &mail.StreamChunk{
		Chunk:    "Assistant response part 2 (final)",
		Sequence: 2,
		IsFinal:  true,
		Taints:   []string{},
	}

	finalData, err := svc.FormatWebSocketChunk(finalChunk)
	if err != nil {
		t.Fatalf("Expected no error formatting final WebSocket chunk, got %v", err)
	}

	var finalParsed map[string]any
	json.Unmarshal(finalData, &finalParsed)
	if finalParsed["isFinal"] != true {
		t.Error("Expected isFinal to be true for final chunk")
	}
}
