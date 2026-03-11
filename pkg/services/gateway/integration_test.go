package gateway

import (
	"encoding/json"
	"slices"
	"strings"
	"testing"
	"time"

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
	svc := NewGatewayService()
	agentID := "agent:dmz"

	// System Service: sys:human-gateway (arch-v1.md L725)
	// For any long-running top-level Agent, open an HTTPS chat session at /chat/{agentId} (arch-v1.md L728)
	session, err := svc.CreateChatSession(agentID)
	if err != nil {
		t.Fatalf("Expected no error creating chat session, got %v", err)
	}

	// Verify session created (arch-v1.md L728)
	if session == nil {
		t.Error("Expected session to be created")
	}

	// Verify agent ID matches (arch-v1.md L728)
	if session.AgentID != agentID {
		t.Errorf("Expected agentID '%s', got '%s'", agentID, session.AgentID)
	}

	// Verify HTTPS endpoint path format (arch-v1.md L728)
	expectedPath := "/chat/" + agentID
	actualPath := svc.GetChatPath(agentID)
	if actualPath != expectedPath {
		t.Errorf("Expected path '%s', got '%s'", expectedPath, actualPath)
	}

	// Session receives read-only snapshot of Agent's current ContextMap (arch-v1.md L731)
	if session.ContextMap == nil {
		t.Error("Expected session to have ContextMap snapshot")
	}

	// Verify snapshot is read-only (arch-v1.md L731)
	if !session.ContextMap.ReadOnly {
		t.Error("Expected session ContextMap to be read-only")
	}

	// Setup: Add messages to session for last N Messages test
	messages := []ChatMessage{
		{ID: "msg1", Content: "Initial message", Taints: []string{}, Boundary: "outer", Type: "user"},
		{ID: "msg2", Content: "Assistant response", Taints: []string{}, Boundary: "outer", Type: "assistant"},
		{ID: "msg3", Content: "Secret message", Taints: []string{"SECRET"}, Boundary: "inner", Type: "assistant"},
		{ID: "msg4", Content: "PII message", Taints: []string{"PII"}, Boundary: "dmz", Type: "user"},
		{ID: "msg5", Content: "Clean message", Taints: []string{}, Boundary: "outer", Type: "user"},
	}

	for _, msg := range messages {
		session.Messages = append(session.Messages, msg)
	}

	// Last N Messages (sanitized by boundary rules) (arch-v1.md L731)
	sanitized := session.GetLastNMessages(3)

	// Verify messages returned (inner-boundary messages filtered, so 2 instead of 3)
	if len(sanitized) < 2 {
		t.Errorf("Expected at least 2 messages, got %d", len(sanitized))
	}

	// Verify SECRET taint removed (arch-v1.md L731)
	for _, msg := range sanitized {
		if slices.Contains(msg.Taints, "SECRET") {
			t.Error("Expected SECRET taint to be sanitized")
		}
	}

	// Verify PII taint removed (arch-v1.md L731)
	for _, msg := range sanitized {
		if slices.Contains(msg.Taints, "PII") {
			t.Error("Expected PII taint to be sanitized")
		}
	}

	// Verify inner-boundary messages removed (arch-v1.md L731)
	for _, msg := range sanitized {
		if msg.Boundary == "inner" {
			t.Error("Expected inner-boundary messages to be sanitized")
		}
	}

	// Any message sent becomes mail_received (type: human_feedback or user) delivered to Agent's inbox (arch-v1.md L732)
	humanMessage := "Hello, agent! How can you help me?"
	m, err := svc.SendHumanMessage(agentID, humanMessage)
	if err != nil {
		t.Fatalf("Expected no error sending human message, got %v", err)
	}

	// Verify mail type is mail_received (arch-v1.md L732)
	if m.Type != mail.MailReceived {
		t.Errorf("Expected mail type 'mail_received', got '%s'", m.Type)
	}

	// Verify mail subtype is human_feedback (arch-v1.md L732)
	if m.Metadata.HumanFeedbackType != "human_feedback" {
		t.Errorf("Expected human_feedback type, got '%s'", m.Metadata.HumanFeedbackType)
	}

	// Verify delivered to Agent's inbox (arch-v1.md L732)
	if m.Target != agentID {
		t.Errorf("Expected target '%s', got '%s'", agentID, m.Target)
	}

	// Agent replies via normal mail → rendered back in chat UI (arch-v1.md L733)
	agentMail := &mail.Mail{
		ID:      "mail-001",
		Type:    mail.MailTypeAssistant,
		Source:  agentID,
		Target:  "user",
		Content: "I can help you with that request.",
		Metadata: mail.MailMetadata{
			Boundary: mail.OuterBoundary,
			Taints:   []string{},
		},
	}

	chatMessage := svc.RenderAgentReply(agentMail)

	// Verify chat message created (arch-v1.md L733)
	if chatMessage.ID == "" {
		t.Error("Expected chat message to be rendered")
	}

	// Verify content matches mail content (arch-v1.md L733)
	if chatMessage.Content != agentMail.Content {
		t.Errorf("Expected content '%s', got '%s'", agentMail.Content, chatMessage.Content)
	}

	// Verify type is assistant (arch-v1.md L733)
	if chatMessage.Type != "assistant" {
		t.Errorf("Expected type 'assistant', got '%s'", chatMessage.Type)
	}

	// Optional "action item" shorthand: @pause, @inject-memory X, etc. become special Mail messages (arch-v1.md L734)
	pauseMessage := "@pause"
	actionItem, err := svc.ParseActionItem(pauseMessage)
	if err != nil {
		t.Fatalf("Expected no error parsing @pause, got %v", err)
	}

	// Verify action item type (arch-v1.md L734)
	if actionItem.Type != "pause" {
		t.Errorf("Expected action type 'pause', got '%s'", actionItem.Type)
	}

	// Send @pause and verify it becomes special Mail (arch-v1.md L734)
	actionMail, err := svc.SendHumanMessage(agentID, pauseMessage)
	if err != nil {
		t.Fatalf("Expected no error sending @pause, got %v", err)
	}

	// Verify mail has action item metadata (arch-v1.md L734)
	if actionMail.Metadata.ActionItem.Type != "pause" {
		t.Errorf("Expected action item type 'pause' in mail metadata, got '%s'", actionMail.Metadata.ActionItem.Type)
	}

	// Test @inject-memory X action item (arch-v1.md L734)
	injectMessage := "@inject-memory This is important to remember"
	actionItem, err = svc.ParseActionItem(injectMessage)
	if err != nil {
		t.Fatalf("Expected no error parsing @inject-memory, got %v", err)
	}

	// Verify action item type (arch-v1.md L734)
	if actionItem.Type != "inject-memory" {
		t.Errorf("Expected action type 'inject-memory', got '%s'", actionItem.Type)
	}

	// Verify payload contains the memory content (arch-v1.md L734)
	if actionItem.Payload != "This is important to remember" {
		t.Errorf("Expected payload 'This is important to remember', got '%v'", actionItem.Payload)
	}

	// Verify full human chat flow completed
	t.Logf("Human chat with running agent completed: session → messages → mail → replies → action items")
}

func TestLayer8Integration_ChannelAdapterToMailToStream(t *testing.T) {
	svc := NewGatewayService()

	// Test WebhookAdapter (arch-v1.md L660)
	t.Run("WebhookAdapter", func(t *testing.T) {
		adapter := &WebhookAdapter{}

		// Inbound normalization (arch-v1.md L670)
		inboundMessage := map[string]any{
			"from":    "sender@example.com",
			"to":      []string{"recipient@example.com"},
			"subject": "Test message",
			"body":    "Hello, world!",
		}

		mailMsg, err := adapter.NormalizeInbound(inboundMessage)
		if err != nil {
			t.Fatalf("NormalizeInbound failed: %v", err)
		}

		// Verify mail type (arch-v1.md L670)
		if mailMsg.Type != mail.MailReceived {
			t.Errorf("Expected type mail_received, got %v", mailMsg.Type)
		}

		// Verify adapter metadata (arch-v1.md L660)
		if mailMsg.Metadata.Adapter != "webhook" {
			t.Errorf("Expected adapter 'webhook', got %v", mailMsg.Metadata.Adapter)
		}

		// Verify Stream() returns false for webhook (arch-v1.md L660)
		if adapter.Stream() {
			t.Error("Expected Stream() to return false for webhook")
		}

		// Outbound normalization (arch-v1.md L670)
		outboundMail := &mail.Mail{
			Type:    mail.MailSend,
			Content: map[string]any{"response": "acknowledged"},
		}

		normalized, err := adapter.NormalizeOutbound(outboundMail)
		if err != nil {
			t.Fatalf("NormalizeOutbound failed: %v", err)
		}

		if normalized == nil {
			t.Error("Expected normalized outbound content")
		}
	})

	// Test WebSocketAdapter (arch-v1.md L661)
	t.Run("WebSocketAdapter", func(t *testing.T) {
		adapter := &WebSocketAdapter{}

		// Inbound normalization (arch-v1.md L670)
		inboundMessage := map[string]any{
			"text":      "Hello from WebSocket client",
			"timestamp": time.Now().Unix(),
		}

		mailMsg, err := adapter.NormalizeInbound(inboundMessage)
		if err != nil {
			t.Fatalf("NormalizeInbound failed: %v", err)
		}

		// Verify mail type (arch-v1.md L670)
		if mailMsg.Type != mail.MailReceived {
			t.Errorf("Expected type mail_received, got %v", mailMsg.Type)
		}

		// Verify adapter metadata (arch-v1.md L661)
		if mailMsg.Metadata.Adapter != "websocket" {
			t.Errorf("Expected adapter 'websocket', got %v", mailMsg.Metadata.Adapter)
		}

		// Verify Stream() returns true for websocket (arch-v1.md L661)
		if !adapter.Stream() {
			t.Error("Expected Stream() to return true for websocket")
		}

		// Full streaming path: WebSocket → mail → StreamChunk → WebSocket
		chunk := &mail.StreamChunk{
			Chunk:       "Assistant response",
			Sequence:    1,
			IsFinal:     true,
			Taints:      []string{"USER_SUPPLIED"},
			MessageType: "partial_assistant",
		}

		wsData, err := svc.FormatWebSocketChunk(chunk)
		if err != nil {
			t.Fatalf("Expected no error formatting WebSocket chunk, got %v", err)
		}

		// Verify WebSocket format: raw JSON (arch-v1.md L696-701)
		var parsed map[string]any
		err = json.Unmarshal(wsData, &parsed)
		if err != nil {
			t.Fatalf("Expected valid JSON in WebSocket data, got %v", err)
		}

		if parsed["chunk"] != "Assistant response" {
			t.Error("Expected chunk field in WebSocket JSON")
		}
	})

	// Test SSEAdapter (arch-v1.md L662)
	t.Run("SSEAdapter", func(t *testing.T) {
		adapter := &SSEAdapter{}

		// Inbound normalization (arch-v1.md L670)
		inboundMessage := map[string]any{
			"event": "user_input",
			"data":  "Hello",
		}

		mailMsg, err := adapter.NormalizeInbound(inboundMessage)
		if err != nil {
			t.Fatalf("NormalizeInbound failed: %v", err)
		}

		// Verify mail type (arch-v1.md L670)
		if mailMsg.Type != mail.MailReceived {
			t.Errorf("Expected type mail_received, got %v", mailMsg.Type)
		}

		// Verify adapter metadata (arch-v1.md L662)
		if mailMsg.Metadata.Adapter != "sse" {
			t.Errorf("Expected adapter 'sse', got %v", mailMsg.Metadata.Adapter)
		}

		// Verify Stream() returns true for sse (arch-v1.md L662)
		if !adapter.Stream() {
			t.Error("Expected Stream() to return true for sse")
		}

		// Full streaming path: SSE → mail → StreamChunk → SSE
		chunk := &mail.StreamChunk{
			Chunk:       "Assistant response",
			Sequence:    1,
			IsFinal:     true,
			Taints:      []string{"USER_SUPPLIED"},
			MessageType: "partial_assistant",
		}

		sseData, err := svc.FormatSSEChunk(chunk)
		if err != nil {
			t.Fatalf("Expected no error formatting SSE chunk, got %v", err)
		}

		// Verify SSE format: data: <json>\n\n (arch-v1.md L696-701)
		if !strings.HasPrefix(sseData, "data: ") {
			t.Errorf("Expected SSE format starting with 'data: ', got '%s'", sseData[:min(20, len(sseData))])
		}
	})

	// Test SMTPAdapter (arch-v1.md L664)
	t.Run("SMTPAdapter", func(t *testing.T) {
		adapter := &SMTPAdapter{}

		// Inbound normalization (arch-v1.md L670)
		inboundEmail := map[string]any{
			"from":    "sender@example.com",
			"to":      []string{"recipient@example.com"},
			"subject": "Test email",
			"body":    "Email body content",
		}

		mailMsg, err := adapter.NormalizeInbound(inboundEmail)
		if err != nil {
			t.Fatalf("NormalizeInbound failed: %v", err)
		}

		// Verify mail type (arch-v1.md L670)
		if mailMsg.Type != mail.MailReceived {
			t.Errorf("Expected type mail_received, got %v", mailMsg.Type)
		}

		// Verify adapter metadata (arch-v1.md L664)
		if mailMsg.Metadata.Adapter != "smtp" {
			t.Errorf("Expected adapter 'smtp', got %v", mailMsg.Metadata.Adapter)
		}

		// Verify Stream() returns false for smtp (arch-v1.md L664)
		if adapter.Stream() {
			t.Error("Expected Stream() to return false for smtp")
		}
	})

	// Test InternalGRPCAdapter (arch-v1.md L666)
	t.Run("InternalGRPCAdapter", func(t *testing.T) {
		adapter := &InternalGRPCAdapter{}

		// Inbound normalization (arch-v1.md L670)
		protobufMessage := map[string]any{
			"service": "internal_service",
			"method":  "ProcessRequest",
			"payload": map[string]any{
				"id":   "req-001",
				"data": "test payload",
			},
		}

		mailMsg, err := adapter.NormalizeInbound(protobufMessage)
		if err != nil {
			t.Fatalf("NormalizeInbound failed: %v", err)
		}

		// Verify mail type (arch-v1.md L670)
		if mailMsg.Type != mail.MailReceived {
			t.Errorf("Expected type mail_received, got %v", mailMsg.Type)
		}

		// Verify adapter metadata (arch-v1.md L666)
		if mailMsg.Metadata.Adapter != "grpc" {
			t.Errorf("Expected adapter 'grpc', got %v", mailMsg.Metadata.Adapter)
		}

		// Verify Stream() returns false for grpc (arch-v1.md L666)
		if adapter.Stream() {
			t.Error("Expected Stream() to return false for grpc")
		}
	})

	// Verify all adapter types implement ChannelAdapter interface (arch-v1.md L659)
	t.Run("ChannelAdapterInterface", func(t *testing.T) {
		var _ ChannelAdapter = &WebhookAdapter{}
		var _ ChannelAdapter = &WebSocketAdapter{}
		var _ ChannelAdapter = &SSEAdapter{}
		var _ ChannelAdapter = &SMTPAdapter{}
		var _ ChannelAdapter = &InternalGRPCAdapter{}
	})

	// Verify full channel adapter → mail → streaming path completed
	t.Logf("Channel adapter to mail to stream completed for all adapters")
}

func TestLayer8Integration_SecurityEnforcedThroughout(t *testing.T) {
	// Stub - will be implemented in next TDD iteration
}
