# Layer 8: Streaming & Gateway - Phase Breakdown

## Executive Summary

Layer 8 implements the Streaming & Gateway layer that provides external interfaces for the Maelstrom runtime. Based on analysis of Layer 7 patterns, Layer 8 specification (arch-v1.md L653-736), and current codebase state, this document breaks down Layer 8 into **6 phases** with **~34 tests** across **~34 commits**.

### Current State (Layers 0-7 Complete/In Progress)
- ✅ Gateway service infrastructure exists with ChannelAdapter interface
- ✅ Basic adapter stubs: Webhook, WebSocket, SSE, SMTP, gRPC
- ✅ HumanGateway service with chat session handling
- ✅ StreamSession and StreamChunk types in pkg/mail
- ⚠️ Adapter implementations incomplete (Handle methods are stubs)
- ⚠️ HTTP(S) exposure & OpenAPI generation missing
- ⚠️ Streaming path end-to-end integration incomplete
- ⚠️ Human chat full implementation missing (action items, ContextMap snapshot)

### Layer 8 Goal
Complete Phase 1 behavior per spec (arch-v1.md L653-736):
1. **Channel Adapters** (L659-667): webhook, websocket, sse, pubsub, smtp, slack, whatsapp, telegram, internal_grpc
2. **Streaming Path** (L672-703): End-to-end flow from user input to SSE output
3. **HTTP(S) Exposure** (L705-723): Chart-declared endpoints, OpenAPI 3.1 auto-generation
4. **Human Chat** (L725-736): sys:human-gateway for long-running agents

---

## Phase Overview

| Phase | Name | Tests | Files | Branch | Dependencies |
|-------|------|-------|-------|--------|--------------|
| 8.1 | Channel Adapter Completion | 6 | 4 | `feat/layer8-channel-adapters` | None |
| 8.2 | Streaming Path Integration | 6 | 3 | `feat/layer8-streaming-path` | 8.1 |
| 8.3 | HTTP(S) Exposure & OpenAPI | 6 | 3 | `feat/layer8-http-openapi` | 8.1 |
| 8.4 | Human Chat Completion | 6 | 2 | `feat/layer8-human-chat` | 8.2 |
| 8.5 | Gateway Security & Tainting | 6 | 3 | `feat/layer8-gateway-security` | 8.2, 8.3 |
| 8.6 | Integration Tests | 4 | 1 | `feat/layer8-integration` | 8.4, 8.5 |

**Total: 34 tests, 16 files, 6 phases**

---

## Phase 8.1: Channel Adapter Completion

### Goal
Complete all 9 channel adapter implementations per spec (arch-v1.md L659-667).

### Scope
- Complete WebhookAdapter HTTP POST handling (L660)
- Complete WebSocketAdapter bidirectional handling (L661)
- Complete SSEAdapter server-sent events (L662)
- Complete PubSubAdapter message queue integration (L663)
- Complete SMTPAdapter email handling (L664)
- Complete messaging platform adapters: Slack, WhatsApp, Telegram (L665)
- Complete InternalGRPCAdapter service mesh (L666)
- Verify all adapters normalize to/from mail.Mail (L670)

### Current Status

#### Implementation Status
| Adapter | Status | Notes |
|---------|--------|-------|
| `WebhookAdapter` | ⚠️ 30% complete | Stub exists; needs HTTP handling, normalization |
| `WebSocketAdapter` | ⚠️ 30% complete | Stub exists; needs WebSocket upgrade, streaming |
| `SSEAdapter` | ⚠️ 30% complete | Stub exists; needs SSE streaming |
| `SMTPAdapter` | ⚠️ 30% complete | Stub exists; needs email send/receive |
| `PubSubAdapter` | ❌ Missing | Need to implement |
| `SlackAdapter` | ❌ Missing | Need to implement |
| `WhatsAppAdapter` | ❌ Missing | Need to implement |
| `TelegramAdapter` | ❌ Missing | Need to implement |
| `InternalGRPCAdapter` | ⚠️ 30% complete | Stub exists; needs gRPC handling |

#### Files Status
| File | Status |
|------|--------|
| `pkg/services/gateway/adapter.go` | ⚠️ Partial - complete existing adapters, add missing |
| `pkg/services/gateway/webhook_adapter_test.go` | ⚠️ Partial - add HTTP handling tests |
| `pkg/services/gateway/websocket_adapter_test.go` | ⚠️ Partial - add WebSocket tests |
| `pkg/services/gateway/sse_adapter_test.go` | ⚠️ Partial - add SSE tests |

### Tests to Write (6 tests, 6 commits)

#### Test 1: TestChannelAdapters_WebhookHTTPHandling
```go
func TestChannelAdapters_WebhookHTTPHandling(t *testing.T)
```
**Acceptance Criteria:**
- `WebhookAdapter.Handle()` processes HTTP POST requests (arch-v1.md L660)
- Request body normalized to mail.Mail with Type=MailReceived
- Response sent as HTTP 200 with acknowledgment
- `Stream()` returns false for webhook

#### Test 2: TestChannelAdapters_WebSocketBidirectional
```go
func TestChannelAdapters_WebSocketBidirectional(t *testing.T)
```
**Acceptance Criteria:**
- `WebSocketAdapter.Handle()` upgrades HTTP to WebSocket (arch-v1.md L661)
- Inbound messages normalized to mail.Mail
- Outbound mail normalized and sent to client
- `Stream()` returns true for websocket

#### Test 3: TestChannelAdapters_SSEStreaming
```go
func TestChannelAdapters_SSEStreaming(t *testing.T)
```
**Acceptance Criteria:**
- `SSEAdapter.Handle()` sets Content-Type: text/event-stream (arch-v1.md L662)
- Stream chunks sent as SSE format (data: \n\n)
- Supports Last-Event-ID for reconnection
- `Stream()` returns true for sse

#### Test 4: TestChannelAdapters_SMTPEmail
```go
func TestChannelAdapters_SMTPEmail(t *testing.T)
```
**Acceptance Criteria:**
- `SMTPAdapter` sends emails via SMTP protocol (arch-v1.md L664)
- Inbound emails normalized to mail.Mail
- Outbound mail formatted as email
- `Stream()` returns false for smtp

#### Test 5: TestChannelAdapters_MessagingPlatforms
```go
func TestChannelAdapters_MessagingPlatforms(t *testing.T)
```
**Acceptance Criteria:**
- `SlackAdapter`, `WhatsAppAdapter`, `TelegramAdapter` normalize platform messages (arch-v1.md L665)
- Each adapter handles platform-specific message format
- Outbound mail normalized to platform format

#### Test 6: TestChannelAdapters_InternalGRPC
```go
func TestChannelAdapters_InternalGRPC(t *testing.T)
```
**Acceptance Criteria:**
- `InternalGRPCAdapter` handles gRPC service calls (arch-v1.md L666)
- Protobuf messages normalized to mail.Mail
- Service mesh integration via gRPC

### Dependencies

#### Test Dependencies
```
Test 1 (Webhook)
Test 2 (WebSocket)
Test 3 (SSE)
Test 4 (SMTP)
Test 5 (Messaging platforms)
Test 6 (gRPC)
```

### Files to Create/Modify

| File | Operation | Changes |
|------|-----------|---------|
| `pkg/services/gateway/adapter.go` | MODIFY | Complete existing adapters, add PubSub, Slack, WhatsApp, Telegram |
| `pkg/services/gateway/pubsub_adapter.go` | CREATE | New PubSub adapter implementation |
| `pkg/services/gateway/messaging_adapters.go` | CREATE | Slack, WhatsApp, Telegram adapters |
| `pkg/services/gateway/adapter_test.go` | MODIFY | Add comprehensive adapter tests |

### Implementation Order (TDD)

1. **Test 1**: Write test → verify RED → Implement Webhook HTTP handling → verify GREEN → commit
2. **Test 2**: Write test → verify RED → Implement WebSocket handling → verify GREEN → commit
3. **Test 3**: Write test → verify RED → Implement SSE streaming → verify GREEN → commit
4. **Test 4**: Write test → verify RED → Implement SMTP handling → verify GREEN → commit
5. **Test 5**: Write test → verify RED → Implement messaging adapters → verify GREEN → commit
6. **Test 6**: Write test → verify RED → Implement gRPC adapter → verify GREEN → commit

**Total: 6 tests, 6 commits**

### Deliverables
- ✅ All 9 channel adapters complete (arch-v1.md L659-667)
- ✅ Each adapter normalizes inbound to mail.Mail
- ✅ Each adapter normalizes outbound from mail.Mail
- ✅ 6 commits (one per test per TDD workflow)

---

## Phase 8.2: Streaming Path Integration

### Goal
Implement complete end-to-end streaming path from user input to SSE output (arch-v1.md L672-703).

### Scope
- Implement streaming path flow (L674-682)
- Implement StreamChunk format (L696-701)
- Implement contextMap streaming configuration (L685-691)
- Integrate with CommunicationService for mail routing
- Integrate with SecurityService for taint stripping

### Current Status

#### Implementation Status
| Component | Status | Notes |
|-----------|--------|-------|
| `StreamSession` | ⚠️ 60% complete | Exists in pkg/mail; needs full integration |
| `StreamChunk` | ⚠️ 60% complete | Exists in pkg/mail; needs boundary enforcement |
| `Streaming path` | ❌ Missing | Need to implement end-to-end flow |
| `ContextMap streaming` | ❌ Missing | Need to implement lastN strategy |

#### Files Status
| File | Status |
|------|--------|
| `pkg/mail/stream.go` | ⚠️ Partial - add streaming path integration |
| `pkg/mail/stream_test.go` | ⚠️ Partial - add streaming tests |
| `pkg/services/gateway/service.go` | ⚠️ Partial - add streaming integration |

### Tests to Write (6 tests, 6 commits)

#### Test 1: TestStreamingPath_UserInputToMail
```go
func TestStreamingPath_UserInputToMail(t *testing.T)
```
**Acceptance Criteria:**
- User input via Gateway creates mail_received Mail (arch-v1.md L675)
- Mail has taints=["USER_SUPPLIED"], boundary=outer
- Communication.publish() delivers to DMZ Agent inbox

#### Test 2: TestStreamingPath_AgentProcessing
```go
func TestStreamingPath_AgentProcessing(t *testing.T)
```
**Acceptance Criteria:**
- Engine.dispatchEvent enters LLMReason state (arch-v1.md L677)
- AgentExtensions.assembleContextMap pulls session + Memory blocks
- Taints propagated through context assembly

#### Test 3: TestStreamingPath_ToolExecution
```go
func TestStreamingPath_ToolExecution(t *testing.T)
```
**Acceptance Criteria:**
- LLM tool calls routed to Orchestrator (arch-v1.md L679)
- Tool Registry.resolve executes in isolated namespace
- Tool result taintPropagated with taints=["TOOL_OUTPUT"]

#### Test 4: TestStreamingPath_PartialAssistantMessages
```go
func TestStreamingPath_PartialAssistantMessages(t *testing.T)
```
**Acceptance Criteria:**
- Engine emits partial_assistant Messages with stream: true (arch-v1.md L681)
- Communication routes to Gateway
- Security.stripForbiddenTaints applied before emission

#### Test 5: TestStreamingPath_StreamChunkFormat
```go
func TestStreamingPath_StreamChunkFormat(t *testing.T)
```
**Acceptance Criteria:**
- StreamChunk has chunk, sequence, isFinal, taints fields (arch-v1.md L696-701)
- taints array stripped before emission to user
- sequence increments for each chunk

#### Test 6: TestStreamingPath_ContextMapStreaming
```go
func TestStreamingPath_ContextMapStreaming(t *testing.T)
```
**Acceptance Criteria:**
- contextMap with streamCommit: false prevents partials entering next prompt (arch-v1.md L685-691)
- lastN strategy limits to configured number of messages
- Conversation source pulls from session

### Dependencies

#### Test Dependencies
```
Test 1 → Test 2 (User input to agent processing)
Test 3 (Tool execution)
Test 4 → Test 5 (Partial messages and chunk format)
Test 6 (ContextMap streaming)
```

### Files to Create/Modify

| File | Operation | Changes |
|------|-----------|---------|
| `pkg/mail/stream.go` | MODIFY | Add streaming path integration |
| `pkg/mail/stream_test.go` | MODIFY | Add streaming path tests |
| `pkg/services/gateway/service.go` | MODIFY | Add streaming integration |

### Implementation Order (TDD)

1. **Test 1**: Write test → verify RED → Implement user input to mail → verify GREEN → commit
2. **Test 2**: Write test → verify RED → Implement agent processing → verify GREEN → commit
3. **Test 3**: Write test → verify RED → Implement tool execution → verify GREEN → commit
4. **Test 4**: Write test → verify RED → Implement partial messages → verify GREEN → commit
5. **Test 5**: Write test → verify RED → Implement StreamChunk format → verify GREEN → commit
6. **Test 6**: Write test → verify RED → Implement ContextMap streaming → verify GREEN → commit

**Total: 6 tests, 6 commits**

### Deliverables
- ✅ End-to-end streaming path complete (arch-v1.md L672-703)
- ✅ StreamChunk format with taints, sequence, isFinal
- ✅ ContextMap streaming with streamCommit configuration
- ✅ 6 commits (one per test per TDD workflow)

---

## Phase 8.3: HTTP(S) Exposure & OpenAPI

### Goal
Implement HTTP endpoint declaration and OpenAPI 3.1 auto-generation (arch-v1.md L705-723).

### Scope
- Implement Chart HTTP endpoint declaration (L707-720)
- Implement OpenAPI 3.1 auto-generation (L722)
- Implement auth middleware integration
- Implement boundary-based exposure (inner never exposed)

### Current Status

#### Implementation Status
| Component | Status | Notes |
|-----------|--------|-------|
| `OpenAPI generation` | ⚠️ 20% complete | Basic spec exists; needs auto-generation |
| `HTTP endpoint declaration` | ❌ Missing | Need to implement YAML parsing |
| `Auth middleware` | ❌ Missing | Need to integrate with security |
| `Boundary exposure` | ❌ Missing | Need to enforce inner never exposed |

#### Files Status
| File | Status |
|------|--------|
| `pkg/openapi/spec.go` | ⚠️ Partial - add auto-generation |
| `pkg/services/gateway/openapi_test.go` | ⚠️ Partial - add generation tests |
| `pkg/bootstrap/chart.go` | ⚠️ Partial - add HTTP expose parsing |

### Tests to Write (6 tests, 6 commits)

#### Test 1: TestHTTPExposure_ChartEndpointDeclaration
```go
func TestHTTPExposure_ChartEndpointDeclaration(t *testing.T)
```
**Acceptance Criteria:**
- Chart YAML expose.http declares path, events, methods (arch-v1.md L707-720)
- trigger events mapped to HTTP methods
- requestSchema and responseSchema defined

#### Test 2: TestHTTPExposure_OpenAPIAutoGeneration
```go
func TestHTTPExposure_OpenAPIAutoGeneration(t *testing.T)
```
**Acceptance Criteria:**
- sys:gateway auto-generates OpenAPI 3.1 spec (arch-v1.md L722)
- All declared endpoints included in spec
- Schema definitions auto-generated from request/response

#### Test 3: TestHTTPExposure_AuthMiddleware
```go
func TestHTTPExposure_AuthMiddleware(t *testing.T)
```
**Acceptance Criteria:**
- Auth middleware applied to all endpoints
- Bearer token validation
- API key validation

#### Test 4: TestHTTPExposure_BoundaryEnforcement
```go
func TestHTTPExposure_BoundaryEnforcement(t *testing.T)
```
**Acceptance Criteria:**
- Inner-boundary Charts never directly exposed (arch-v1.md L723)
- Only DMZ/outer Charts exposed via gateway
- Boundary check enforced at registration time

#### Test 5: TestHTTPExposure_StreamEndpoints
```go
func TestHTTPExposure_StreamEndpoints(t *testing.T)
```
**Acceptance Criteria:**
- stream: true endpoints return SSE/WebSocket responses
- Content-Type set correctly for streaming
- Chunked transfer encoding supported

#### Test 6: TestHTTPExposure_EventSurfaceToAPI
```go
func TestHTTPExposure_EventSurfaceToAPI(t *testing.T)
```
**Acceptance Criteria:**
- Chart event surface becomes API surface (arch-v1.md L723)
- Each trigger event creates corresponding endpoint
- Request/response schemas derived from event payload

### Dependencies

#### Test Dependencies
```
Test 1 → Test 2 (Endpoint declaration to OpenAPI)
Test 3 (Auth middleware)
Test 4 (Boundary enforcement)
Test 5 (Stream endpoints)
Test 6 (Event to API mapping)
```

### Files to Create/Modify

| File | Operation | Changes |
|------|-----------|---------|
| `pkg/openapi/generator.go` | CREATE | OpenAPI auto-generation from chart definitions |
| `pkg/bootstrap/chart.go` | MODIFY | Add HTTP expose YAML parsing |
| `pkg/services/gateway/openapi_test.go` | MODIFY | Add comprehensive OpenAPI tests |

### Implementation Order (TDD)

1. **Test 1**: Write test → verify RED → Implement endpoint declaration → verify GREEN → commit
2. **Test 2**: Write test → verify RED → Implement OpenAPI generation → verify GREEN → commit
3. **Test 3**: Write test → verify RED → Implement auth middleware → verify GREEN → commit
4. **Test 4**: Write test → verify RED → Implement boundary enforcement → verify GREEN → commit
5. **Test 5**: Write test → verify RED → Implement stream endpoints → verify GREEN → commit
6. **Test 6**: Write test → verify RED → Implement event to API mapping → verify GREEN → commit

**Total: 6 tests, 6 commits**

### Deliverables
- ✅ Chart HTTP endpoint declaration complete (arch-v1.md L705-723)
- ✅ OpenAPI 3.1 auto-generation complete
- ✅ Auth middleware integration complete
- ✅ Boundary-based exposure enforced
- ✅ 6 commits (one per test per TDD workflow)

---

## Phase 8.4: Human Chat Completion

### Goal
Complete sys:human-gateway for human chat with running agents (arch-v1.md L725-736).

### Scope
- Implement /chat/{agentId} endpoint (L728)
- Implement ContextMap snapshot with boundary sanitization (L731)
- Implement action item parsing (@pause, @inject-memory) (L735)
- Implement read-only session snapshot

### Current Status

#### Implementation Status
| Component | Status | Notes |
|-----------|--------|-------|
| `HumanGatewayService` | ⚠️ 50% complete | Basic structure exists; needs full implementation |
| `ChatSession` | ⚠️ 50% complete | Exists; needs ContextMap integration |
| `ActionItem parsing` | ⚠️ 30% complete | Stub exists; needs full parsing |
| `ContextMap snapshot` | ❌ Missing | Need to implement read-only snapshot |

#### Files Status
| File | Status |
|------|--------|
| `pkg/services/humangateway/service.go` | ⚠️ Partial - add full implementation |
| `pkg/services/humangateway/chat.go` | ⚠️ Partial - add ContextMap snapshot |
| `pkg/services/humangateway/service_test.go` | ⚠️ Partial - add comprehensive tests |

### Tests to Write (6 tests, 6 commits)

#### Test 1: TestHumanChat_AgentEndpoint
```go
func TestHumanChat_AgentEndpoint(t *testing.T)
```
**Acceptance Criteria:**
- /chat/{agentId} endpoint opens HTTPS chat session (arch-v1.md L728)
- Session created for long-running top-level Agent
- AgentID validated and resolved

#### Test 2: TestHumanChat_ContextMapSnapshot
```go
func TestHumanChat_ContextMapSnapshot(t *testing.T)
```
**Acceptance Criteria:**
- Session receives read-only snapshot of Agent's ContextMap (arch-v1.md L731)
- Last N Messages included in snapshot
- Sanitized by boundary rules (no SECRET/PII taints)

#### Test 3: TestHumanChat_HumanFeedbackMail
```go
func TestHumanChat_HumanFeedbackMail(t *testing.T)
```
**Acceptance Criteria:**
- Any message sent becomes mail_received (arch-v1.md L732)
- Type: human_feedback or user
- Delivered to Agent's inbox

#### Test 4: TestHumanChat_AgentReplyRendering
```go
func TestHumanChat_AgentReplyRendering(t *testing.T)
```
**Acceptance Criteria:**
- Agent replies via normal mail (arch-v1.md L733)
- Mail rendered back in chat UI
- Taints stripped based on boundary

#### Test 5: TestHumanChat_ActionItemParsing
```go
func TestHumanChat_ActionItemParsing(t *testing.T)
```
**Acceptance Criteria:**
- @pause becomes special Mail message (arch-v1.md L735)
- @inject-memory X becomes special Mail message
- ActionItems parsed and attached to mail content

#### Test 6: TestHumanChat_SessionManagement
```go
func TestHumanChat_SessionManagement(t *testing.T)
```
**Acceptance Criteria:**
- CreateChatSession creates new session for agent
- GetSession retrieves existing session
- Sessions isolated per agent

### Dependencies

#### Test Dependencies
```
Test 1 → Test 2 (Endpoint to ContextMap snapshot)
Test 3 (Human feedback mail)
Test 4 (Agent reply rendering)
Test 5 (Action item parsing)
Test 6 (Session management)
```

### Files to Create/Modify

| File | Operation | Changes |
|------|-----------|---------|
| `pkg/services/humangateway/service.go` | MODIFY | Complete human gateway implementation |
| `pkg/services/humangateway/chat.go` | MODIFY | Add ContextMap snapshot, action item parsing |
| `pkg/services/humangateway/service_test.go` | MODIFY | Add comprehensive tests |

### Implementation Order (TDD)

1. **Test 1**: Write test → verify RED → Implement agent endpoint → verify GREEN → commit
2. **Test 2**: Write test → verify RED → Implement ContextMap snapshot → verify GREEN → commit
3. **Test 3**: Write test → verify RED → Implement human feedback mail → verify GREEN → commit
4. **Test 4**: Write test → verify RED → Implement agent reply rendering → verify GREEN → commit
5. **Test 5**: Write test → verify RED → Implement action item parsing → verify GREEN → commit
6. **Test 6**: Write test → verify RED → Implement session management → verify GREEN → commit

**Total: 6 tests, 6 commits**

### Deliverables
- ✅ Human chat endpoint complete (arch-v1.md L725-736)
- ✅ ContextMap snapshot with boundary sanitization
- ✅ Action item parsing (@pause, @inject-memory)
- ✅ Session management complete
- ✅ 6 commits (one per test per TDD workflow)

---

## Phase 8.5: Gateway Security & Tainting

### Goal
Implement security and tainting for gateway operations (arch-v1.md L675-682, L696-701).

### Scope
- Implement taint propagation through gateway (L675, L680)
- Implement forbidden taint stripping (L681, L696-701)
- Implement boundary transitions at gateway
- Implement security validation for all adapters

### Current Status

#### Implementation Status
| Component | Status | Notes |
|-----------|--------|-------|
| `Taint propagation` | ⚠️ 40% complete | Basic exists; needs gateway integration |
| `Forbidden taint stripping` | ⚠️ 40% complete | Exists in stream.go; needs enforcement |
| `Boundary transitions` | ⚠️ 30% complete | Basic exists; needs gateway enforcement |
| `Adapter security` | ❌ Missing | Need to implement per-adapter security |

#### Files Status
| File | Status |
|------|--------|
| `pkg/security/security.go` | ⚠️ Partial - add gateway security |
| `pkg/mail/stream.go` | ⚠️ Partial - add taint enforcement |
| `pkg/services/gateway/boundary_adapter_test.go` | ⚠️ Partial - add security tests |

### Tests to Write (6 tests, 6 commits)

#### Test 1: TestGatewaySecurity_TaintPropagation
```go
func TestGatewaySecurity_TaintPropagation(t *testing.T)
```
**Acceptance Criteria:**
- USER_SUPPLIED taint added at gateway entry (arch-v1.md L675)
- TOOL_OUTPUT taint added for tool results (arch-v1.md L680)
- Taints propagated through mail content

#### Test 2: TestGatewaySecurity_ForbiddenTaintStripping
```go
func TestGatewaySecurity_ForbiddenTaintStripping(t *testing.T)
```
**Acceptance Criteria:**
- Security.stripForbiddenTaints removes SECRET/PII before emission (arch-v1.md L681)
- StreamChunk taints array filtered
- Only allowed taints in output

#### Test 3: TestGatewaySecurity_BoundaryTransitions
```go
func TestGatewaySecurity_BoundaryTransitions(t *testing.T)
```
**Acceptance Criteria:**
- Outer→Inner transitions taint data
- Inner→Outer transitions blocked for sensitive taints
- DMZ boundary handled correctly

#### Test 4: TestGatewaySecurity_AdapterSecurity
```go
func TestGatewaySecurity_AdapterSecurity(t *testing.T)
```
**Acceptance Criteria:**
- Each adapter validates incoming data
- Malformed requests rejected
- Authentication required for sensitive operations

#### Test 5: TestGatewaySecurity_StreamTaintHandling
```go
func TestGatewaySecurity_StreamTaintHandling(t *testing.T)
```
**Acceptance Criteria:**
- Stream chunks have taints field (arch-v1.md L696-701)
- Taints stripped before sending to user
- Taint propagation tracked per chunk

#### Test 6: TestGatewaySecurity_NormalizationSecurity
```go
func TestGatewaySecurity_NormalizationSecurity(t *testing.T)
```
**Acceptance Criteria:**
- NormalizeInbound validates adapter name
- NormalizeOutbound respects boundary
- Security metadata stripped for outer boundary

### Dependencies

#### Test Dependencies
```
Test 1 → Test 2 (Taint propagation to stripping)
Test 3 (Boundary transitions)
Test 4 (Adapter security)
Test 5 (Stream taint handling)
Test 6 (Normalization security)
```

### Files to Create/Modify

| File | Operation | Changes |
|------|-----------|---------|
| `pkg/security/security.go` | MODIFY | Add gateway security functions |
| `pkg/mail/stream.go` | MODIFY | Add taint enforcement |
| `pkg/services/gateway/boundary_adapter_test.go` | MODIFY | Add comprehensive security tests |

### Implementation Order (TDD)

1. **Test 1**: Write test → verify RED → Implement taint propagation → verify GREEN → commit
2. **Test 2**: Write test → verify RED → Implement forbidden taint stripping → verify GREEN → commit
3. **Test 3**: Write test → verify RED → Implement boundary transitions → verify GREEN → commit
4. **Test 4**: Write test → verify RED → Implement adapter security → verify GREEN → commit
5. **Test 5**: Write test → verify RED → Implement stream taint handling → verify GREEN → commit
6. **Test 6**: Write test → verify RED → Implement normalization security → verify GREEN → commit

**Total: 6 tests, 6 commits**

### Deliverables
- ✅ Taint propagation through gateway complete (arch-v1.md L675, L680)
- ✅ Forbidden taint stripping complete (arch-v1.md L681)
- ✅ Boundary transitions enforced
- ✅ Adapter security implemented
- ✅ 6 commits (one per test per TDD workflow)

---

## Phase 8.6: Integration Tests

### Goal
Verify all gateway components integrate correctly via comprehensive integration tests.

### Scope
- End-to-end channel adapter integration
- Streaming path end-to-end flow
- Human chat end-to-end flow
- Security enforcement end-to-end

### Tests to Write (4 tests, 4 commits)

#### Test 1: TestIntegration_ChannelAdapterAllTypes
```go
func TestIntegration_ChannelAdapterAllTypes(t *testing.T)
```
**Acceptance Criteria:**
- All 9 channel adapters can receive and send messages
- Normalization works correctly for each type
- Gateway service routes to correct adapter

#### Test 2: TestIntegration_StreamingPathEndToEnd
```go
func TestIntegration_StreamingPathEndToEnd(t *testing.T)
```
**Acceptance Criteria:**
- Complete streaming path from user input to SSE output
- ContextMap assembled correctly
- Tool execution integrated
- Stream chunks formatted correctly

#### Test 3: TestIntegration_HumanChatEndToEnd
```go
func TestIntegration_HumanChatEndToEnd(t *testing.T)
```
**Acceptance Criteria:**
- Human can connect to agent via /chat/{agentId}
- ContextMap snapshot provided
- Messages exchanged correctly
- Action items processed

#### Test 4: TestIntegration_GatewaySecurityEndToEnd
```go
func TestIntegration_GatewaySecurityEndToEnd(t *testing.T)
```
**Acceptance Criteria:**
- Taints propagated through entire flow
- Forbidden taints stripped at boundary
- Boundary transitions enforced
- Security validated end-to-end

### Dependencies

#### Test Dependencies
```
Test 1 (Channel adapter integration)
Test 2 (Streaming path integration)
Test 3 (Human chat integration)
Test 4 (Security integration)
```

### Files to Create/Modify

| File | Operation | Changes |
|------|-----------|---------|
| `pkg/services/integration/gateway_integration_test.go` | CREATE | Comprehensive gateway integration tests |

### Implementation Order (TDD)

1. **Test 1**: Write test → verify RED → Integrate all channel adapters → verify GREEN → commit
2. **Test 2**: Write test → verify RED → Integrate streaming path → verify GREEN → commit
3. **Test 3**: Write test → verify RED → Integrate human chat → verify GREEN → commit
4. **Test 4**: Write test → verify RED → Integrate security → verify GREEN → commit

**Total: 4 tests, 4 commits**

### Deliverables
- ✅ All channel adapters integrated
- ✅ Streaming path end-to-end verified
- ✅ Human chat end-to-end verified
- ✅ Security enforcement end-to-end verified
- ✅ 4 commits (one per test per TDD workflow)

---

## Summary

### Phase Summary Table

| Phase | Tests | Files | Branch | Dependencies |
|-------|-------|-------|--------|--------------|
| 8.1 | 6 | 4 | `feat/layer8-channel-adapters` | None |
| 8.2 | 6 | 3 | `feat/layer8-streaming-path` | 8.1 |
| 8.3 | 6 | 3 | `feat/layer8-http-openapi` | 8.1 |
| 8.4 | 6 | 2 | `feat/layer8-human-chat` | 8.2 |
| 8.5 | 6 | 3 | `feat/layer8-gateway-security` | 8.2, 8.3 |
| 8.6 | 4 | 1 | `feat/layer8-integration` | 8.4, 8.5 |

**Total: 34 tests, 16 files, 6 phases**

### Execution Order

```
Phase 8.1 (Channel Adapter Completion)
      ↓              ↘
      ↓        Phase 8.3 (HTTP & OpenAPI)
Phase 8.2 (Streaming Path Integration)
      ↓
Phase 8.4 (Human Chat Completion)
      ↓
Phase 8.5 (Gateway Security)
      ↓
Phase 8.6 (Integration Tests)
```

### Files to Create

| File | Phase | Purpose |
|------|-------|---------|
| `pkg/services/gateway/pubsub_adapter.go` | 8.1 | PubSub adapter implementation |
| `pkg/services/gateway/messaging_adapters.go` | 8.1 | Slack, WhatsApp, Telegram adapters |
| `pkg/openapi/generator.go` | 8.3 | OpenAPI auto-generation |
| `pkg/services/integration/gateway_integration_test.go` | 8.6 | Gateway integration tests |

### Files to Modify

| File | Phases | Changes |
|------|--------|---------|
| `pkg/services/gateway/adapter.go` | 8.1 | Complete existing adapters |
| `pkg/services/gateway/service.go` | 8.1, 8.2 | Add streaming integration |
| `pkg/mail/stream.go` | 8.2, 8.5 | Add streaming path, taint enforcement |
| `pkg/bootstrap/chart.go` | 8.3 | Add HTTP expose parsing |
| `pkg/services/humangateway/service.go` | 8.4 | Complete human gateway |
| `pkg/services/humangateway/chat.go` | 8.4 | Add ContextMap snapshot |
| `pkg/security/security.go` | 8.5 | Add gateway security |

### Next Steps

1. **Start Phase 8.1**: Create branch `feat/layer8-channel-adapters`
2. **Follow TDD workflow**: Write test → verify RED → Implement → verify GREEN → commit
3. **Repeat for each phase**: One branch per phase, merge after completion
4. **Run full test suite** after Phase 8.6 to verify integration

---

## Appendix A: Patterns Applied

This phase breakdown follows established patterns from Layer 7:
- ✅ **1:1 test-to-commit ratio** (34 tests = 34 commits)
- ✅ **One branch per phase** (6 branches total)
- ✅ **Descriptive test names** (`Test[Component]_[Behavior]_[ExpectedResult]`)
- ✅ **Small, focused phases** (4-6 tests per phase)
- ✅ **Explicit dependencies** between phases
- ✅ **Clear acceptance criteria** for each test
- ✅ **arch-v1.md line number references** for all requirements

## Appendix B: Gap Analysis Summary

| Gap | Resolution | Phase | Spec Reference |
|-----|------------|-------|----------------|
| Channel adapters incomplete | Complete all 9 adapters | 8.1 | L659-667 |
| Streaming path missing | Implement end-to-end flow | 8.2 | L672-703 |
| HTTP exposure missing | Implement endpoint declaration | 8.3 | L705-723 |
| OpenAPI generation missing | Implement auto-generation | 8.3 | L722 |
| Human chat incomplete | Complete implementation | 8.4 | L725-736 |
| Gateway security incomplete | Implement taint/stripping | 8.5 | L675-682, L696-701 |
| Integration tests missing | Add end-to-end tests | 8.6 | All |

## Appendix C: Layer Dependencies

### Layer Dependencies
- **Layer 0 (Statechart Engine)**: Complete - provides runtime foundation
- **Layer 1 (Mail Types)**: Complete - provides Mail, StreamChunk types
- **Layer 2-3 (Core Services)**: Complete - provides service infrastructure
- **Layer 4 (Security)**: Complete - provides taint/boundary types
- **Layer 5-6 (Registry/Tools)**: In Progress - provides chart registry
- **Layer 7 (Platform Services)**: In Progress - provides sys:gateway foundation

### Layer 8 Completes
- Channel adapter interface implementation
- Streaming path end-to-end
- HTTP(S) exposure with OpenAPI
- Human chat interface

### Layers 9+ Build On Layer 8
- Layer 9 (Agent Charts): Uses gateway for external communication
- Layer 10 (LLM Integration): Uses streaming for LLM output
