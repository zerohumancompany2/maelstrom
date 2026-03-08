# Layer 3: Communication & Mail System

**Reference**: `docs/arch-v1.md` Section 9  
**Parent Scope**: `docs/planning/implementation-scope.md`  
**Status**: ❌ PENDING

---

## Overview

Layer 3 implements the Communication layer that provides cross-boundary async messaging via a pub-sub mail system. This is the **only cross-boundary primitive** in Maelstrom - Events drive internal transitions within Charts, but Mail is used for inter-agent communication.

**Key Principle**: Statecharts are driven by Events. Inter-agent communication uses a pub-sub mail system built on top of a standard messaging library.

---

## Architectural Dependencies

```
Layer 3 depends on:
├── Layer 2 (Core Platform Services)
│   └── sys:communication must be online to use Mail system
└── Layer 0 (Statechart Engine) ✓ COMPLETE
    └── Event system (internal Chart transitions)
```

**Why Layer 2 First**: The sys:communication service depends on the Mail system to function. Layer 3 provides the Mail types and routing infrastructure that Layer 2's CommunicationService uses.

---

## Events vs Mail (arch-v1.md 9.1)

### Events (Internal Only)

- **Purpose**: Drive internal transitions within a Chart
- **Scope**: Local to ChartRuntime only
- **Ordering**:
  - Synchronous ordering guaranteed within compound/parallel boundaries
  - "Eventually consistent" across parallel regions
- **Delivery**: Best-effort, at-least-once with deduplication via `correlationId`
- **Addressing** (local to ChartRuntime):
  - `targetPath` optional, used for:
    - Parallel regions: `region:foo`
    - Sub-charts: `child:bar` or `child:bar/region:baz`
    - Self: omitted or `.`
  - No global addressing, no topics, no inter-Chart routing

### Mail (Cross-Boundary Primitive)

- **Purpose**: The only cross-boundary communication primitive
- **Scope**: Global addressing across agents and services
- **Nature**: Async, fire-and-forget or request-reply via `correlationId`
- **Key Principle**: All inter-agent communication goes through Mail (never direct event passing)

---

## Mail Addressing (arch-v1.md 9.2)

Three address types:

| Format | Purpose | Example |
|--------|---------|---------|
| `agent:<id>` | Specific agent inbox | `agent:recommendation-agent` |
| `topic:<name>` | Pub-sub topic (multiple subscribers) | `topic:market-data` |
| `sys:<service>` | System service inbox | `sys:heartbeat`, `sys:persistence` |

**Inbox Rule**: Only top-level agents (expected to have long life-cycles) have inboxes. All agents can send mail to any inbox.

---

## Complete Mail Message Structure (arch-v1.md 3.3, 9.2)

```go
type Mail struct {
    ID            string    // unique message ID
    CorrelationID string    // correlation ID for request-reply
    Type          MailType  // see MailTypes below
    CreatedAt     time.Time // creation timestamp
    Source        string    // agent:<id> | sys:<service> | human | gateway | tool:<name>
    Target        string    // agent:<id> | topic:<name> | sys:<service>
    Content       any       // message payload
    Metadata      MailMetadata
}

type MailMetadata struct {
    Tokens    int       // token count
    Model     string    // LLM model used
    Cost      float64   // cost of operation
    Boundary  BoundaryType // inner | dmz | outer
    Taints    []string  // taint categories
    Stream    bool      // is streaming response
    IsFinal   bool      // last chunk marker
}
```

### All 11 Mail Types and Their Purposes

| Type | Purpose |
|------|---------|
| `user` | User input / external message |
| `assistant` | Assistant response |
| `tool_result` | Tool execution result |
| `tool_call` | Tool invocation request |
| `mail_received` | Mail arrival notification |
| `heartbeat` | Scheduled agent wake-up |
| `error` | Error condition |
| `human_feedback` | Human-in-the-loop feedback |
| `partial_assistant` | Streaming LLM output chunk |
| `subagent_done` | Sub-agent completion signal |
| `taint_violation` | Security policy violation |

---

## Delivery Guarantees (arch-v1.md 9.2, 14.2)

- **At-least-once delivery**: Guaranteed by Communication layer
- **Deduplication**: Via `correlationId`
- **Dead-letter queue**: Managed by `sys:observability` for permanent failures
- **Request-reply pattern**: Supported via `correlationId` correlation
- **Routing**: To `agent:<id>`, `topic:<name>`, or `sys:<service>`

---

## Communication Interface (arch-v1.md 9.3)

```go
type Communication interface {
    // Publish sends mail to inbox/topic
    Publish(mail Mail) (ack Ack, err error)
    
    // Subscribe subscribes to address
    Subscribe(address string) (mailStream chan Mail, err error)
    
    // Unsubscribe removes subscription
    Unsubscribe(address string, stream chan Mail) error
    
    // UpgradeToStream upgrades connection to SSE/WS streaming
    UpgradeToStream(sessionId string, lastEventId *string) (chunkChannel chan StreamChunk, err error)
}
```

### Ack Structure

```go
type Ack struct {
    CorrelationID string    // correlation ID for request-reply
    DeliveredAt   time.Time // delivery timestamp
}
```

### StreamChunk Structure

```go
type StreamChunk struct {
    Data     string    // chunk data
    Sequence int       // sequence number
    IsFinal  bool      // last chunk flag
    Taints   []string  // taint categories (stripped before emission)
}
```

---

## End-to-End Streaming Path (arch-v1.md 10.2)

```
User types → Gateway (SSE connect) → Communication.publish(mail: type=user, taints=["USER_SUPPLIED"], boundary=outer)
Security.validate → DMZ Agent inbox
Engine.dispatchEvent → enters LLMReason state
AgentExtensions.assembleContextMap (pulls session + Memory blocks; taints propagated) → LLM call
LLM returns tool calls → Orchestrator (isolated sub-chart) → Tool Registry.resolve → executeIsolated (new namespace)
Tool result → taintPropagate → new tool_result Mail (taints=["TOOL_OUTPUT"]) injected to session
Engine emits partial_assistant Messages (stream: true) → Communication → Gateway → Security.stripForbiddenTaints → SSE chunks to user
```

### Stream Configuration (ContextBlock)

```yaml
contextMap:
  - name: conversation
    source: session
    strategy: lastN
    n: 30
    streamCommit: false  # false = partials never enter next prompt
```

### Stream Chunk Format (JSON)

```json
{
  "chunk": "string",
  "sequence": 1,
  "isFinal": false,
  "taints": ["TOOL_OUTPUT"]
}
```

**Stream Sanitization**:
- Every outgoing chunk passes through DMZ sanitizers (PII redaction, length caps, schema validation, inner-data stripping)
- Sanitization is **per-chunk (stateless)** so latency stays **<50ms**
- No buffering of entire response unless chart explicitly requests it

---

## Gateway Channel Adapters (arch-v1.md 10.1)

YAML-declared, hot-loaded adapters that normalize external I/O:

| Adapter | Purpose |
|---------|---------|
| `webhook` | HTTP POST endpoints |
| `websocket` | Full bidirectional |
| `sse` | Server-Sent Events (firewall-friendly) |
| `pubsub` | Message queue integration |
| `smtp` | Email |
| `slack` | Messaging platform |
| `whatsapp` | Messaging platform |
| `telegram` | Messaging platform |
| `internal_grpc` | Internal service mesh |

**Normalization Behavior**:
- **Inbound**: Each adapter normalizes traffic to `mail_received` Message
- **Outbound**: Each adapter normalizes outbound `mail_send` to channel's specific format

---

## Human Chat with Running Agents (arch-v1.md 10.4)

**System Service**: `sys:human-gateway`

**Features**:
- HTTPS chat session at `/chat/{agentId}`
- Session receives read-only snapshot of Agent's ContextMap + last N Messages (sanitized by boundary rules)
- Messages sent become `mail_received` (type: `human_feedback` or `user`) delivered to Agent's inbox
- Agent replies via normal mail → rendered back in chat UI
- **Optional "action item" shorthand**: `@pause`, `@inject-memory X`, etc. become special Mail messages

---

## Service Implementation Details

### CommunicationService (Layer 2 Integration)

```go
type CommunicationService struct {
    id          string  // "sys:communication"
    publisher   Publisher
    subscribers map[string][]Subscriber
    mu          sync.RWMutex
}

type Publisher interface {
    Publish(mail Mail) (Ack, error)
}

type Subscriber interface {
    Receive() <-chan Mail
}

func NewCommunicationService() *CommunicationService

func (s *CommunicationService) ID() string

func (s *CommunicationService) HandleMail(mail Mail) error {
    // Route mail to target (agent, topic, or sys service)
}

func (s *CommunicationService) Subscribe(address string, sub Subscriber) error

func (s *CommunicationService) Publish(mail Mail) (Ack, error)

func (s *CommunicationService) UpgradeToStream(sessionId string, lastEventId *string) (chan StreamChunk, error)
```

### Mail Router

```go
type MailRouter struct {
    agents    map[string]*AgentInbox      // agent:<id> → inbox
    topics    map[string]*Topic           // topic:<name> → subscribers
    services  map[string]*ServiceInbox    // sys:<service> → inbox
    mu        sync.RWMutex
}

func NewMailRouter() *MailRouter

func (r *MailRouter) Route(mail Mail) error

func (r *MailRouter) SubscribeAgent(id string, sub Subscriber) error

func (r *MailRouter) SubscribeTopic(name string, sub Subscriber) error

func (r *MailRouter) SubscribeService(name string, sub Subscriber) error
```

### AgentInbox

```go
type AgentInbox struct {
    ID        string
    Messages  []Mail
    mu        sync.RWMutex
}

func (a *AgentInbox) Push(mail Mail) error

func (a *AgentInbox) Pop() (Mail, error)

func (a *AgentInbox) Subscribe() <-chan Mail
```

### Topic

```go
type Topic struct {
    Name      string
    Subscribers []Subscriber
    mu        sync.RWMutex
}

func (t *Topic) Publish(mail Mail) error

func (t *Topic) Subscribe(sub Subscriber) error

func (t *Topic) Unsubscribe(sub Subscriber) error
```

---

## Taint Handling

### Stream Taint Stripping

Every outgoing chunk passes through DMZ sanitizers:

```go
func StripForbiddenTaints(chunk StreamChunk, allowed []string) StreamChunk {
    var filtered []string
    for _, taint := range chunk.Taints {
        for _, allowedTaint := range allowed {
            if taint == allowedTaint {
                filtered = append(filtered, taint)
                break
            }
        }
    }
    chunk.Taints = filtered
    return chunk
}
```

### Taint Attachment Points

1. **At Message creation** (Gateway → user, Orchestrator → tool_result, MemoryService → injected block, Heartbeat, etc.)
2. **At ContextBlock assembly** (Security calls `applyBlockTaints`)
3. **On any data leaving the runtime** (Mail, stream chunks, snapshot, sub-agent result) — uses `allowedOnExit`

---

## File Structure

```
maelstrom/
├── pkg/mail/
│   ├── types.go           # Mail types, Addressing (~100 lines)
│   ├── router.go          # MailRouter implementation (~150 lines)
│   ├── publisher.go       # Publisher interface + impl (~50 lines)
│   ├── subscriber.go      # Subscriber interface + impl (~50 lines)
│   ├── inbox.go           # AgentInbox, Topic, ServiceInbox (~100 lines)
│   ├── stream.go          # StreamChunk, streaming support (~50 lines)
│   └── mail_test.go       # Comprehensive tests (~200 lines)
│
└── pkg/gateway/
    ├── gateway.go         # Gateway service (~150 lines)
    ├── adapters/
    │   ├── webhook.go     # HTTP POST adapter
    │   ├── websocket.go   # WebSocket adapter
    │   ├── sse.go         # Server-Sent Events adapter
    │   ├── pubsub.go      # Message queue adapter
    │   ├── smtp.go        # Email adapter
    │   └── platform.go    # Slack, WhatsApp, Telegram adapters
    └── gateway_test.go
```

---

## TDD Implementation Plan

### Phase 3.1: Mail Types

**Test**: `TestMail_AddressFormats`
- Verify agent:<id>, topic:<name>, sys:<service> addressing

**Test**: `TestMail_Types`
- All 11 mail types defined and usable

**Test**: `TestMail_Metadata`
- Metadata fields populated correctly

**Implementation**:
- ~100 lines types.go
- Define Mail, MailMetadata, MailType, Ack, StreamChunk

### Phase 3.2: Mail Router

**Test**: `TestMailRouter_RouteToAgent`
- Direct routing to agent inbox

**Test**: `TestMailRouter_RouteToTopic`
- Topic broadcasting to subscribers

**Test**: `TestMailRouter_RouteToService`
- Routing to sys:<service>

**Implementation**:
- ~150 lines router.go
- Map-based routing with mutex protection

### Phase 3.3: Inboxes

**Test**: `TestAgentInbox_PushPop`
- Messages pushed and popped correctly

**Test**: `TestTopic_SubscribeUnsubscribe`
- Subscribers added/removed correctly

**Test**: `TestTopic_Broadcast`
- Broadcast to all subscribers

**Implementation**:
- ~100 lines inbox.go
- Channel-based pub/sub

### Phase 3.4: Publisher/Subscriber

**Test**: `TestPublisher_Publish`
- Publish sends mail correctly

**Test**: `TestSubscriber_Receives`
- Subscriber receives mail

**Implementation**:
- ~100 lines publisher.go + subscriber.go
- Interface + concrete implementations

### Phase 3.5: Streaming

**Test**: `TestStreamUpgrade`
- Upgrade to SSE/WS streaming

**Test**: `TestStreamChunkFormat`
- Chunk format correct

**Test**: `TestTaintStripping`
- Taints stripped correctly

**Implementation**:
- ~50 lines stream.go
- SSE/WS upgrade logic

### Phase 3.6: Integration

**Test**: `TestFullMailFlow`
- End-to-end mail flow works

**Test**: `TestCommunicationService_Integration`
- CommunicationService uses Mail system

**Implementation**:
- Wire Mail system to CommunicationService
- Verify full integration

---

## Dependencies & External Libraries

- `sync` - Mutex for thread-safe routing
- `github.com/maelstrom/v3/pkg/statechart` - Mail types compatibility

---

## Risk Assessment

| Risk | Likelihood | Impact | Mitigation |
|------|-----------|--------|------------|
| Mail routing deadlocks | Low | High | Comprehensive tests with timeouts |
| Subscriber channel leaks | Medium | Medium | Proper Unsubscribe implementation |
| Stream upgrade failures | Low | Medium | Error handling and fallback |

---

## Open Questions

1. **Persistence**: Store mail history, or ephemeral only?
2. **Topic scope**: Global topics, or per-agent?
3. **Delivery confirmation**: Sync ack, or async?
4. **Backpressure**: Handle subscriber overload?

---

## References

- `docs/arch-v1.md` Section 9 - Inter-Agent Communication
- `docs/arch-v1.md` Section 10.2 - Streaming Path
- `docs/arch-v1.md` Section 14.2 - Communication SHALL Statements
- `docs/planning/layer-02-core-services.md` - Layer 2 (dependency)
- `docs/planning/implementation-scope.md` - Parent scope document

---

