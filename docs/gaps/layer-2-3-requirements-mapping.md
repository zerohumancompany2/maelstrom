# Layer 2 & 3 Requirements Mapping

**Generated**: 2026-03-09  
**Purpose**: Detailed mapping of arch-v1.md requirements to implementation

---

## Section 7: Platform Services

### 7.1 Core Services (Hard-Coded)

#### Requirement: sys:security — Boundary enforcement, taint validation, sanitization

| Aspect | Spec | Implementation | Status |
|--------|------|----------------|--------|
| Service ID | `sys:security` | `pkg/services/security/service.go:15` returns `"sys:security"` | ✅ |
| HandleMail | `handleMail(mail: Mail) → outcomeEvent` | `pkg/services/security/service.go:21-23` returns `nil` | ⚠️ |
| ValidateAndSanitize | `ValidateAndSanitize(mail, sourceBoundary, targetBoundary) (Mail, error)` | `pkg/services/security/service.go:28-30` returns input unchanged | ❌ |
| TaintPropagate | `TaintPropagate(obj, newTaints) (any, error)` | `pkg/services/security/service.go:32-34` returns obj unchanged | ❌ |
| NamespaceIsolate | `NamespaceIsolate(runtimeId, operation) (IsolatedView, error)` | **NOT FOUND** | ❌ |
| CheckTaintPolicy | `CheckTaintPolicy(data, targetBoundary, policy) (bool, error)` | **NOT FOUND** | ❌ |
| ReportTaints | `ReportTaints(runtimeId) (TaintMap, error)` | `pkg/services/security/service.go:36-38` returns empty TaintMap | ⚠️ |
| PrepareContextForBoundary | `PrepareContextForBoundary(runtimeId, boundary) error` | `pkg/services/security/service.go:40-42` returns nil | ⚠️ |
| Start/Stop | Lifecycle methods | `pkg/services/security/service.go:44-50` return nil | ✅ |

**Notes**: 
- Service exists but all methods are stubs
- No actual boundary enforcement logic
- Missing 2 required methods (NamespaceIsolate, CheckTaintPolicy)

---

#### Requirement: sys:communication — Mail pub/sub backbone

| Aspect | Spec | Implementation | Status |
|--------|------|----------------|--------|
| Service ID | `sys:communication` | `pkg/services/communication/service.go:24` returns `"sys:communication"` | ✅ |
| Publish | `Publish(mail) (ack, err)` | `pkg/services/communication/service.go:30-56` | ✅ |
| Subscribe | `Subscribe(address) (chan Mail, err)` | `pkg/services/communication/service.go:58-78` | ✅ |
| Unsubscribe | `Unsubscribe(address, stream) error` | `pkg/services/communication/service.go:80-96` | ✅ |
| UpgradeToStream | `UpgradeToStream(sessionId, lastEventId) (chan StreamChunk, err)` | `pkg/services/communication/service.go:98-100` calls mail.UpgradeToStream | ✅ |
| HandleMail | `HandleMail(mail) error` | `pkg/services/communication/service.go:102-107` routes mail | ✅ |
| At-least-once delivery | Guaranteed | No retry logic found | ❌ |
| Deduplication | Via correlationId | Not implemented | ❌ |
| Dead-letter integration | Managed by sys:observability | Not wired | ❌ |

**Notes**:
- Basic pub/sub implemented
- Missing delivery guarantees
- Local subscribers tracked separately from router

---

#### Requirement: sys:observability — Traces, metrics, dead-letter queue

| Aspect | Spec | Implementation | Status |
|--------|------|----------------|--------|
| Service ID | `sys:observability` | `pkg/services/observability/service.go:21` returns `"sys:observability"` | ✅ |
| EmitTrace | `EmitTrace(trace) error` | `pkg/services/observability/service.go:33-41` | ✅ |
| QueryTraces | `QueryTraces(filters) ([]Trace, error)` | `pkg/services/observability/service.go:43-61` | ✅ |
| LogDeadLetter | `LogDeadLetter(mail, reason) error` | `pkg/services/observability/service.go:75-83` | ✅ |
| QueryDeadLetters | Returns dead letters | `pkg/services/observability/service.go:85-93` | ✅ |
| GetMetrics | Returns MetricsCollector | `pkg/services/observability/service.go:95-105` | ⚠️ |
| HandleMail | `HandleMail(mail) error` | `pkg/services/observability/service.go:28-30` returns nil | ✅ |
| Metrics - StateCounts | Track state counts | Implemented | ✅ |
| Metrics - TransitionRate | Track transition rate | Not implemented | ❌ |
| Metrics - EventRate | Track event rate | Not implemented | ❌ |

**Notes**:
- Trace storage and query implemented
- Dead-letter queue implemented
- Metrics incomplete (only state counts)

---

#### Requirement: sys:lifecycle (minimal) — Basic spawn/stop for bootstrap

| Aspect | Spec | Implementation | Status |
|--------|------|----------------|--------|
| Service ID | `sys:lifecycle` | `pkg/services/lifecycle/service.go:25` returns `"sys:lifecycle"` | ✅ |
| Spawn | `Spawn(def) (RuntimeID, error)` | `pkg/services/lifecycle/service.go:40-58` | ✅ |
| Stop | `Stop(id) error` | `pkg/services/lifecycle/service.go:60-65` | ✅ |
| Control | `Control(id, cmd) error` | `pkg/services/lifecycle/service.go:67-72` | ✅ |
| List | `List() ([]RuntimeInfo, error)` | `pkg/services/lifecycle/service.go:74-84` | ✅ |
| HandleMail | `HandleMail(mail) error` | `pkg/services/lifecycle/service.go:35-37` returns nil | ✅ |
| Runtime tracking | Track active runtimes | `pkg/services/lifecycle/service.go:13-14` | ⚠️ |

**Notes**:
- Basic spawn/stop implemented
- Runtime tracking exists but not updated on state changes
- Hot-reload not implemented (expected in Layer 5)

---

### 7.2 Hot-Reloadable Services (YAML-Defined)

| Service | Spec | Implementation | Status |
|---------|------|----------------|--------|
| sys:gateway | Channel adapters, OpenAPI, 2FA | `pkg/services/gateway/service.go` stub exists | ⚠️ |
| sys:admin | k9s-style console, 2FA-gated | `pkg/services/admin/service.go` stub exists | ⚠️ |
| sys:persistence | Snapshots, event sourcing | `pkg/services/persistence/service.go` stub exists | ⚠️ |
| sys:heartbeat | Scheduled wake-ups | `pkg/services/heartbeat/service.go` stub exists | ⚠️ |
| sys:memory | Vector/graph stores | `pkg/services/memory/service.go` stub exists | ⚠️ |
| sys:human-gateway | Chat interface | `pkg/services/humangateway/service.go` basic impl | ⚠️ |
| sys:tools | Tool registry | `pkg/services/tools/service.go` stub exists | ⚠️ |
| sys:datasources | FS/object storage | `pkg/services/datasources/service.go` stub exists | ⚠️ |

**Notes**: All hot-reloadable services exist as stubs but not fully implemented

---

### 7.3 Platform Service Contract

| Contract | Spec | Implementation | Status |
|----------|------|----------------|--------|
| handleMail(mail) → outcomeEvent | Every sys:* service | All services have HandleMail method | ✅ |
| adminCommand (sys:admin) | 2FA-gated admin commands | Not implemented | ❌ |
| snapshot/restore (sys:persistence) | Snapshot and restore | Not implemented | ❌ |
| emit/query (sys:observability) | Trace emit and query | Implemented | ✅ |
| resolve (sys:tools) | Tool resolution | Not implemented | ❌ |
| query (sys:memory) | Memory query | Not implemented | ❌ |
| tagOnWrite/getTaints (sys:datasources) | Taint tagging | Not implemented | ❌ |

---

## Section 9: Inter-Agent Communication

### 9.1 Events vs Mail

| Concept | Spec | Implementation | Status |
|---------|------|----------------|--------|
| Events drive internal transitions | Within Chart | Statechart engine (Layer 0) | ✅ |
| Mail is cross-boundary primitive | Only cross-boundary method | Mail system implemented | ✅ |
| Mail is async | Fire-and-forget or request-reply | Fire-and-forget only | ⚠️ |

---

### 9.2 Mail System

#### Mail Addressing

| Address Type | Spec | Implementation | Status |
|--------------|------|----------------|--------|
| agent:<id> | Specific agent inbox | `pkg/mail/router.go:26-32` | ✅ |
| topic:<name> | Pub-sub topic | `pkg/mail/router.go:34-40` | ✅ |
| sys:<service> | System service inbox | `pkg/mail/router.go:42-48` | ✅ |
| IsValidAgentAddress | Validation helper | **NOT FOUND** | ❌ |
| IsValidTopicAddress | Validation helper | **NOT FOUND** | ❌ |
| IsValidSysAddress | Validation helper | **NOT FOUND** | ❌ |
| ParseAddress | Parse address type/id | **NOT FOUND** (referenced in router.go:12) | ❌ |

#### Mail Format

| Field | Spec | Implementation | Status |
|-------|------|----------------|--------|
| id | uuid | `pkg/mail/types.go:4` string | ✅ |
| correlationId | uuid | `pkg/mail/types.go:5` string | ✅ |
| type | 11 types | `pkg/mail/types.go:9-29` all 11 defined | ✅ |
| createdAt | timestamp | `pkg/mail/types.go:6` time.Time | ✅ |
| source | agent/sys/human/gateway/tool | `pkg/mail/types.go:7` string | ✅ |
| target | agent/topic/sys | `pkg/mail/types.go:8` string | ✅ |
| content | any | `pkg/mail/types.go:9` any | ✅ |
| metadata.tokens | int | `pkg/mail/types.go:35` int | ✅ |
| metadata.model | string | `pkg/mail/types.go:36` string | ✅ |
| metadata.cost | float | `pkg/mail/types.go:37` float64 | ✅ |
| metadata.boundary | inner/dmz/outer | `pkg/mail/types.go:38` BoundaryType | ✅ |
| metadata.taints | []string | `pkg/mail/types.go:39` []string | ✅ |
| metadata.stream | boolean | `pkg/mail/types.go:40` *StreamChunk | ⚠️ |
| metadata.isFinal | boolean | `pkg/mail/types.go:41` bool | ✅ |

**Notes**: 
- `metadata.stream` type mismatch (spec says bool, impl says *StreamChunk)
- All 11 mail types implemented with aliases

#### Mail Types

| Type | Spec | Implementation | Status |
|------|------|----------------|--------|
| user | User input | `MailTypeUser` | ✅ |
| assistant | Assistant response | `MailTypeAssistant` | ✅ |
| tool_result | Tool result | `MailTypeToolResult` | ✅ |
| tool_call | Tool call | `MailTypeToolCall` | ✅ |
| mail_received | Mail arrival | `MailTypeMailReceived` | ✅ |
| heartbeat | Scheduled wake-up | `MailTypeHeartbeat` | ✅ |
| error | Error condition | `MailTypeError` | ✅ |
| human_feedback | Human feedback | `MailTypeHumanFeedback` | ✅ |
| partial_assistant | Streaming chunk | `MailTypePartialAssistant` | ✅ |
| subagent_done | Sub-agent done | `MailTypeSubagentDone` | ✅ |
| taint_violation | Security violation | `MailTypeTaintViolation` | ✅ |

#### Delivery Guarantees

| Guarantee | Spec | Implementation | Status |
|-----------|------|----------------|--------|
| At-least-once | Guaranteed | Not implemented | ❌ |
| Deduplication | Via correlationId | Not implemented | ❌ |
| Dead-letter queue | On permanent failure | `pkg/services/observability/service.go:75-93` | ✅ |

---

### 9.3 Communication Layer API

| Method | Spec | Implementation | Status |
|--------|------|----------------|--------|
| Publish(mail) (ack, err) | Send mail | `pkg/services/communication/service.go:30-56` | ✅ |
| Subscribe(address) (chan Mail, err) | Subscribe | `pkg/services/communication/service.go:58-78` | ✅ |
| Unsubscribe(address, stream) error | Unsubscribe | `pkg/services/communication/service.go:80-96` | ✅ |
| UpgradeToStream(sessionId, lastEventId) (chan StreamChunk, err) | Streaming | `pkg/services/communication/service.go:98-100` | ✅ |

#### Ack Structure

| Field | Spec | Implementation | Status |
|-------|------|----------------|--------|
| CorrelationID | string | `pkg/mail/types.go:57` | ✅ |
| DeliveredAt | time.Time | `pkg/mail/types.go:58` | ✅ |
| Success | (extra) | `pkg/mail/types.go:59` | ✅ |
| ErrorMessage | (extra) | `pkg/mail/types.go:60` | ✅ |

#### StreamChunk Structure

| Field | Spec | Implementation | Status |
|-------|------|----------------|--------|
| Data | string | `pkg/mail/types.go:45` | ✅ |
| Sequence | int | `pkg/mail/types.go:46` | ✅ |
| IsFinal | bool | `pkg/mail/types.go:47` | ✅ |
| Taints | []string | `pkg/mail/types.go:48` | ✅ |

---

## Section 10: Streaming & Gateway

### 10.1 Gateway Layer

#### Channel Adapters

| Adapter | Spec | Implementation | Status |
|---------|------|----------------|--------|
| webhook | HTTP POST | `pkg/gateway/adapters/webhook.go` stub | ⚠️ |
| websocket | Bidirectional | `pkg/gateway/adapters/websocket.go` stub | ⚠️ |
| sse | Server-Sent Events | `pkg/gateway/adapters/sse.go` stub | ⚠️ |
| pubsub | Message queue | `pkg/gateway/adapters/pubsub.go` stub | ⚠️ |
| smtp | Email | `pkg/gateway/adapters/smtp.go` stub | ⚠️ |
| slack | Messaging | `pkg/gateway/adapters/platform.go` stub | ⚠️ |
| whatsapp | Messaging | `pkg/gateway/adapters/platform.go` stub | ⚠️ |
| telegram | Messaging | `pkg/gateway/adapters/platform.go` stub | ⚠️ |
| internal_grpc | Service mesh | **NOT FOUND** | ❌ |

#### Gateway Service

| Feature | Spec | Implementation | Status |
|---------|------|----------------|--------|
| Adapter interface | Normalize inbound/outbound | `pkg/gateway/gateway.go:12-16` | ✅ |
| RegisterAdapter | Register adapters | `pkg/gateway/gateway.go:26-32` | ✅ |
| GetAdapter | Get by name | `pkg/gateway/gateway.go:34-42` | ✅ |
| ListAdapters | List all | `pkg/gateway/gateway.go:44-52` | ✅ |
| Actual HTTP server | Serve requests | Not implemented | ❌ |
| Actual WebSocket server | Serve WS connections | Not implemented | ❌ |
| Actual SSE endpoint | Serve SSE | Not implemented | ❌ |

---

### 10.2 Streaming Path

| Component | Spec | Implementation | Status |
|-----------|------|----------------|--------|
| StreamSession | Session management | `pkg/mail/stream.go:11-20` | ✅ |
| StreamSession.Send | Send chunk | `pkg/mail/stream.go:32-34` panics | ❌ |
| StreamSession.Close | Close session | `pkg/mail/stream.go:36-38` panics | ❌ |
| UpgradeToStream | Upgrade connection | `pkg/mail/stream.go:40-43` | ✅ |
| StripForbiddenTaints | Taint stripping | `pkg/mail/stream.go:46-61` | ✅ |
| contextMap streaming config | streamCommit: false | Not implemented | ❌ |

---

### 10.3 HTTP(S) Exposure & OpenAPI

| Feature | Spec | Implementation | Status |
|---------|------|----------------|--------|
| Chart HTTP endpoint declaration | YAML expose section | Not implemented | ❌ |
| OpenAPI 3.1 generation | Auto-generate from charts | Not implemented | ❌ |
| Auth middleware | Auth on endpoints | Not implemented | ❌ |
| Boundary-based exposure | Only DMZ/outer exposed | Not implemented | ❌ |

---

### 10.4 Human Chat with Running Agents

| Feature | Spec | Implementation | Status |
|---------|------|----------------|--------|
| sys:human-gateway service | Service exists | `pkg/services/humangateway/service.go` | ✅ |
| Chat session at /chat/{agentId} | HTTPS endpoint | Not implemented | ❌ |
| ContextMap snapshot | Read-only snapshot | `pkg/services/humangateway/service.go:60-70` stub | ⚠️ |
| Message as human_feedback/user | Mail delivery | `pkg/services/humangateway/service.go:27-50` | ✅ |
| Action item parsing | @pause, @inject-memory | `ParseActionItem` method referenced but not found | ❌ |

---

## Summary Tables

### Implementation Coverage by File

| File | Lines | Purpose | Test Coverage |
|------|-------|---------|---------------|
| pkg/services/security/service.go | ~50 | Security service stub | service_test.go exists |
| pkg/services/communication/service.go | ~110 | Communication service | service_test.go exists |
| pkg/services/observability/service.go | ~110 | Observability service | service_test.go exists |
| pkg/services/lifecycle/service.go | ~85 | Lifecycle service | service_test.go exists |
| pkg/services/registry.go | ~70 | Service registry | registry_test.go exists |
| pkg/mail/types.go | ~60 | Mail types | types_test.go exists |
| pkg/mail/router.go | ~70 | Mail routing | router_test.go exists |
| pkg/mail/inbox.go | ~50 | Inboxes | inbox_test.go exists |
| pkg/mail/publisher.go | ~20 | Publisher | publisher_test.go exists |
| pkg/mail/subscriber.go | ~20 | Subscriber | (no dedicated test) |
| pkg/mail/stream.go | ~60 | Streaming | stream_test.go exists |
| pkg/gateway/gateway.go | ~55 | Gateway service | (tests in services/gateway/) |
| pkg/gateway/adapters/*.go | ~350 | All adapters | Tests exist for webhook, sse, websocket |
| pkg/services/humangateway/service.go | ~70 | Human gateway | service_test.go exists |

### Critical Missing Implementations

1. `ParseAddress` function (referenced in router.go:12)
2. `StreamSession.Send()` method (panics)
3. `StreamSession.Close()` method (panics)
4. `NamespaceIsolate` method (not found in security service)
5. `CheckTaintPolicy` method (not found in security service)
6. `ParseActionItem` method (referenced but not found in human gateway)
7. Address validation helpers (IsValidAgentAddress, etc.)

### Partial Implementations

1. Security service - all methods are stubs
2. Streaming - basic structure but Send/Close panic
3. Gateway adapters - normalization logic exists but no actual servers
4. Observability metrics - only state counts, missing rates
5. Human gateway - service exists but no chat endpoint

---

**Mapping End**