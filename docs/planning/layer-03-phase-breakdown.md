# Layer 3: Communication & Mail System - Phase Breakdown

## Executive Summary

Layer 3 implements the Communication layer that provides cross-boundary async messaging via a pub-sub mail system. Based on analysis of Layer 2 patterns, Layer 3 specification, and the architectural requirements, this document breaks down Layer 3 into **8 phases** with **~38 tests** across **~38 commits**.

### Current State (Layer 0 Complete, Layer 2 In Progress)
- ✅ Layer 0 (Statechart Engine) complete
- ⏳ Layer 2 (Core Platform Services) in progress - provides CommunicationService stub
- ❌ Mail types and routing infrastructure not yet implemented
- ⚠️ Layer 2's sys:communication depends on Layer 3 Mail system

### Layer 3 Goal
Complete the Mail system infrastructure:
1. **Mail Types** - Core message types, addressing, metadata
2. **Mail Router** - Routing logic for agent, topic, sys addresses
3. **Inboxes** - AgentInbox, Topic, ServiceInbox implementations
4. **Publisher/Subscriber** - Pub/sub interfaces and implementations
5. **Streaming** - SSE/WS streaming with taint stripping
6. **Gateway Adapters** - Channel adapters for external I/O
7. **Human Gateway** - Human chat with running agents
8. **Integration** - Wire everything to CommunicationService

---

## Phase Overview

| Phase | Name | Tests | Files | Branch | Dependencies |
|-------|------|-------|-------|--------|--------------|
| 3.1 | Mail Core Types | 6 | 2 | `feat/layer3-mail-types` | None |
| 3.2 | Mail Router | 5 | 1 | `feat/layer3-mail-router` | 3.1 |
| 3.3 | Inboxes & Topics | 6 | 1 | `feat/layer3-inboxes` | 3.1 |
| 3.4 | Publisher/Subscriber | 4 | 2 | `feat/layer3-pubsub` | 3.1, 3.3 |
| 3.5 | Streaming Support | 5 | 1 | `feat/layer3-streaming` | 3.1 |
| 3.6 | Gateway Adapters | 6 | 5 | `feat/layer3-gateway` | 3.1, 3.5 |
| 3.7 | Human Gateway Service | 4 | 2 | `feat/layer3-human-gateway` | 3.1, 3.6 |
| 3.8 | Integration & CommunicationService | 2 | 2 | `feat/layer3-integration` | 3.2, 3.3, 3.4, 3.5 |

**Total: 38 tests, 16 files, 8 phases**

---

## Phase 3.1: Mail Core Types

### Goal
Define all Mail types, addressing formats, and metadata structures required by the communication layer.

### Scope
- Create `pkg/mail/types.go` with Mail, MailType, MailMetadata, Ack
- Create `pkg/mail/address.go` with address parsing and validation
- Define all 11 Mail types
- Implement address format validation (agent:<id>, topic:<name>, sys:<service>)

### Current Status

#### Implementation Status
| Type | Status | Notes |
|------|--------|-------|
| `Mail` | ❌ Missing | Core message structure |
| `MailType` | ❌ Missing | 11 mail types enum |
| `MailMetadata` | ❌ Missing | Token, cost, taint metadata |
| `Ack` | ❌ Missing | Delivery acknowledgment |
| `StreamChunk` | ⚠️ Exists in Layer 2 | Will be reused |

#### Files Status
| File | Status |
|------|-------|
| `pkg/mail/types.go` | ❌ MISSING - create |
| `pkg/mail/address.go` | ❌ MISSING - create |

### Required Implementation

#### Mail Type
```go
// pkg/mail/types.go
type Mail struct {
    ID            string
    CorrelationID string
    Type          MailType
    CreatedAt     time.Time
    Source        string
    Target        string
    Content       any
    Metadata      MailMetadata
}
```

#### MailType Enum
```go
type MailType string

const (
    MailTypeUser            MailType = "user"
    MailTypeAssistant       MailType = "assistant"
    MailTypeToolResult      MailType = "tool_result"
    MailTypeToolCall        MailType = "tool_call"
    MailTypeMailReceived    MailType = "mail_received"
    MailTypeHeartbeat       MailType = "heartbeat"
    MailTypeError           MailType = "error"
    MailTypeHumanFeedback   MailType = "human_feedback"
    MailTypePartialAssistant MailType = "partial_assistant"
    MailTypeSubagentDone    MailType = "subagent_done"
    MailTypeTaintViolation  MailType = "taint_violation"
)
```

#### MailMetadata
```go
type MailMetadata struct {
    Tokens   int
    Model    string
    Cost     float64
    Boundary BoundaryType
    Taints   []string
    Stream   bool
    IsFinal  bool
}

type BoundaryType string

const (
    InnerBoundary BoundaryType = "inner"
    DmzBoundary   BoundaryType = "dmz"
    OuterBoundary BoundaryType = "outer"
)
```

#### Ack
```go
type Ack struct {
    CorrelationID string
    DeliveredAt   time.Time
}
```

#### Address Validation
```go
// pkg/mail/address.go
func IsValidAgentAddress(addr string) bool
func IsValidTopicAddress(addr string) bool
func IsValidSysAddress(addr string) bool
func ParseAddress(addr string) (addrType AddressType, id string, err error)
```

### Tests to Write (6 tests, 6 commits)

#### Test 1: TestMail_AddressFormats
```go
func TestMail_AddressFormats(t *testing.T)
```
**Acceptance Criteria:**
- `agent:<id>` format validated correctly
- `topic:<name>` format validated correctly
- `sys:<service>` format validated correctly
- Invalid formats rejected

#### Test 2: TestMail_Types
```go
func TestMail_Types(t *testing.T)
```
**Acceptance Criteria:**
- All 11 mail types defined
- Each type has unique string value
- Types can be compared and matched

#### Test 3: TestMail_Metadata
```go
func TestMail_Metadata(t *testing.T)
```
**Acceptance Criteria:**
- MailMetadata has all required fields
- BoundaryType enum has inner, dmz, outer values
- Taints is a slice that can be empty or populated

#### Test 4: TestMail_Structure
```go
func TestMail_Structure(t *testing.T)
```
**Acceptance Criteria:**
- Mail struct has all required fields
- Can instantiate Mail with all fields populated
- Content field accepts any type

#### Test 5: TestAck_Structure
```go
func TestAck_Structure(t *testing.T)
```
**Acceptance Criteria:**
- Ack has CorrelationID and DeliveredAt fields
- Can instantiate Ack with values
- DeliveredAt is time.Time type

#### Test 6: TestAddress_ParseAddress
```go
func TestAddress_ParseAddress(t *testing.T)
```
**Acceptance Criteria:**
- ParseAddress extracts type and id correctly
- Returns error for invalid formats
- Handles all three address formats

### Dependencies

#### Test Dependencies
```
Test 1 → Test 6 (Address parsing)
Test 2 → Independent
Test 3 → Independent
Test 4 → Test 2, Test 3 (Mail structure)
Test 5 → Independent
```

### Files to Create/Modify

| File | Operation | Changes |
|------|-----------|---------|
| `pkg/mail/types.go` | CREATE | Mail, MailType, MailMetadata, Ack, BoundaryType |
| `pkg/mail/address.go` | CREATE | Address validation and parsing |

### Implementation Order (TDD)

1. **Test 1**: Write test → verify RED → Implement address validation functions → verify GREEN → commit
2. **Test 2**: Write test → verify RED → Define all 11 MailType constants → verify GREEN → commit
3. **Test 3**: Write test → verify RED → Implement MailMetadata and BoundaryType → verify GREEN → commit
4. **Test 4**: Write test → verify RED → Implement Mail struct → verify GREEN → commit
5. **Test 5**: Write test → verify RED → Implement Ack struct → verify GREEN → commit
6. **Test 6**: Write test → verify RED → Implement ParseAddress function → verify GREEN → commit

**Total: 6 tests, 6 commits**

### Deliverables
- ✅ Mail, MailType, MailMetadata, Ack types in `pkg/mail/types.go`
- ✅ Address validation in `pkg/mail/address.go`
- ✅ All 11 mail types defined
- ✅ 6 commits (one per test per TDD workflow)

---

## Phase 3.2: Mail Router

### Goal
Implement the MailRouter that routes messages to agent inboxes, topics, or system services.

### Scope
- Create `pkg/mail/router.go` with MailRouter struct
- Implement Route() method
- Implement SubscribeAgent, SubscribeTopic, SubscribeService methods
- Thread-safe routing with mutex protection

### Current Status

#### Implementation Status
| Component | Status | Notes |
|-----------|--------|-------|
| `MailRouter` | ❌ Missing | Core routing logic |
| `Route()` | ❌ Missing | Route to agent/topic/service |
| `SubscribeAgent()` | ❌ Missing | Subscribe to agent inbox |
| `SubscribeTopic()` | ❌ Missing | Subscribe to topic |
| `SubscribeService()` | ❌ Missing | Subscribe to sys service |

#### Files Status
| File | Status |
|------|-------|
| `pkg/mail/router.go` | ❌ MISSING - create |

### Required Implementation

#### MailRouter
```go
type MailRouter struct {
    agents   map[string]*AgentInbox
    topics   map[string]*Topic
    services map[string]*ServiceInbox
    mu       sync.RWMutex
}

func NewMailRouter() *MailRouter
func (r *MailRouter) Route(mail Mail) error
func (r *MailRouter) SubscribeAgent(id string, sub Subscriber) error
func (r *MailRouter) SubscribeTopic(name string, sub Subscriber) error
func (r *MailRouter) SubscribeService(name string, sub Subscriber) error
```

### Tests to Write (5 tests, 5 commits)

#### Test 1: TestMailRouter_RouteToAgent
```go
func TestMailRouter_RouteToAgent(t *testing.T)
```
**Acceptance Criteria:**
- Mail to `agent:<id>` delivered to correct inbox
- Returns nil error on success
- Agent must be registered first

#### Test 2: TestMailRouter_RouteToTopic
```go
func TestMailRouter_RouteToTopic(t *testing.T)
```
**Acceptance Criteria:**
- Mail to `topic:<name>` broadcast to all subscribers
- Multiple subscribers receive same mail
- Returns nil error on success

#### Test 3: TestMailRouter_RouteToService
```go
func TestMailRouter_RouteToService(t *testing.T)
```
**Acceptance Criteria:**
- Mail to `sys:<service>` delivered to service inbox
- Returns nil error on success
- Service must be registered first

#### Test 4: TestMailRouter_RouteToUnknownAddress
```go
func TestMailRouter_RouteToUnknownAddress(t *testing.T)
```
**Acceptance Criteria:**
- Route to unregistered agent returns error
- Route to unregistered topic returns error
- Route to unregistered service returns error

#### Test 5: TestMailRouter_SubscribeMethods
```go
func TestMailRouter_SubscribeMethods(t *testing.T)
```
**Acceptance Criteria:**
- SubscribeAgent registers agent inbox
- SubscribeTopic registers topic subscribers
- SubscribeService registers service inbox

### Dependencies

#### Test Dependencies
```
Test 5 → Test 1, Test 2, Test 3 (Subscribe must work first)
Test 4 (Error handling - independent)
```

### Files to Create/Modify

| File | Operation | Changes |
|------|-----------|---------|
| `pkg/mail/router.go` | CREATE | MailRouter struct and methods |

### Implementation Order (TDD)

1. **Test 5**: Write test → verify RED → Implement SubscribeAgent/Topic/Service → verify GREEN → commit
2. **Test 1**: Write test → verify RED → Implement Route() for agent addresses → verify GREEN → commit
3. **Test 2**: Write test → verify RED → Implement Route() for topic addresses → verify GREEN → commit
4. **Test 3**: Write test → verify RED → Implement Route() for sys addresses → verify GREEN → commit
5. **Test 4**: Write test → verify RED → Add error handling for unknown addresses → verify GREEN → commit

**Total: 5 tests, 5 commits**

### Deliverables
- ✅ MailRouter in `pkg/mail/router.go`
- ✅ Route() method for all address types
- ✅ SubscribeAgent, SubscribeTopic, SubscribeService methods
- ✅ 5 commits (one per test per TDD workflow)

---

## Phase 3.3: Inboxes & Topics

### Goal
Implement AgentInbox, Topic, and ServiceInbox data structures for message storage and delivery.

### Scope
- Create `pkg/mail/inbox.go` with inbox implementations
- Implement AgentInbox with Push/Pop/Subscribe
- Implement Topic with Publish/Subscribe/Unsubscribe
- Implement ServiceInbox (similar to AgentInbox)
- Channel-based pub/sub

### Current Status

#### Implementation Status
| Component | Status | Notes |
|-----------|--------|-------|
| `AgentInbox` | ❌ Missing | Agent message queue |
| `Topic` | ❌ Missing | Pub-sub topic |
| `ServiceInbox` | ❌ Missing | System service queue |

#### Files Status
| File | Status |
|------|-------|
| `pkg/mail/inbox.go` | ❌ MISSING - create |

### Required Implementation

#### AgentInbox
```go
type AgentInbox struct {
    ID       string
    Messages []Mail
    mu       sync.RWMutex
}

func (a *AgentInbox) Push(mail Mail) error
func (a *AgentInbox) Pop() (Mail, error)
func (a *AgentInbox) Subscribe() <-chan Mail
```

#### Topic
```go
type Topic struct {
    Name        string
    Subscribers []Subscriber
    mu          sync.RWMutex
}

func (t *Topic) Publish(mail Mail) error
func (t *Topic) Subscribe(sub Subscriber) error
func (t *Topic) Unsubscribe(sub Subscriber) error
```

#### ServiceInbox
```go
type ServiceInbox struct {
    ID       string
    Messages []Mail
    mu       sync.RWMutex
}
```

### Tests to Write (6 tests, 6 commits)

#### Test 1: TestAgentInbox_PushPop
```go
func TestAgentInbox_PushPop(t *testing.T)
```
**Acceptance Criteria:**
- Messages pushed to inbox
- Messages popped in FIFO order
- Returns nil error on success

#### Test 2: TestAgentInbox_Subscribe
```go
func TestAgentInbox_Subscribe(t *testing.T)
```
**Acceptance Criteria:**
- Subscribe returns a channel
- Messages sent to channel when pushed

#### Test 3: TestTopic_SubscribeUnsubscribe
```go
func TestTopic_SubscribeUnsubscribe(t *testing.T)
```
**Acceptance Criteria:**
- Subscribe adds subscriber
- Unsubscribe removes subscriber
- Returns nil error on success

#### Test 4: TestTopic_Broadcast
```go
func TestTopic_Broadcast(t *testing.T)
```
**Acceptance Criteria:**
- Publish delivers to all subscribers
- Each subscriber receives mail once

#### Test 5: TestServiceInbox_PushPop
```go
func TestServiceInbox_PushPop(t *testing.T)
```
**Acceptance Criteria:**
- ServiceInbox Push/Pop works like AgentInbox
- FIFO ordering maintained

#### Test 6: TestTopic_UnsubscribeNotFound
```go
func TestTopic_UnsubscribeNotFound(t *testing.T)
```
**Acceptance Criteria:**
- Unsubscribe non-existent subscriber returns error
- Error indicates subscriber not found

### Dependencies

#### Test Dependencies
```
Test 1 → Test 2 (Push/Pop before Subscribe)
Test 3 → Test 4, Test 6 (Subscribe before Broadcast/Unsubscribe)
Test 5 (Independent - ServiceInbox)
```

### Files to Create/Modify

| File | Operation | Changes |
|------|-----------|---------|
| `pkg/mail/inbox.go` | CREATE | AgentInbox, Topic, ServiceInbox |

### Implementation Order (TDD)

1. **Test 1**: Write test → verify RED → Implement AgentInbox Push/Pop → verify GREEN → commit
2. **Test 2**: Write test → verify RED → Implement AgentInbox Subscribe → verify GREEN → commit
3. **Test 3**: Write test → verify RED → Implement Topic Subscribe/Unsubscribe → verify GREEN → commit
4. **Test 4**: Write test → verify RED → Implement Topic Publish → verify GREEN → commit
5. **Test 5**: Write test → verify RED → Implement ServiceInbox → verify GREEN → commit
6. **Test 6**: Write test → verify RED → Add error handling to Unsubscribe → verify GREEN → commit

**Total: 6 tests, 6 commits**

### Deliverables
- ✅ AgentInbox in `pkg/mail/inbox.go`
- ✅ Topic with pub/sub in `pkg/mail/inbox.go`
- ✅ ServiceInbox in `pkg/mail/inbox.go`
- ✅ 6 commits (one per test per TDD workflow)

---

## Phase 3.4: Publisher/Subscriber

### Goal
Implement Publisher and Subscriber interfaces with concrete implementations.

### Scope
- Create `pkg/mail/publisher.go` with Publisher interface
- Create `pkg/mail/subscriber.go` with Subscriber interface
- Implement concrete types that satisfy interfaces
- Integration with MailRouter

### Current Status

#### Implementation Status
| Component | Status | Notes |
|-----------|--------|-------|
| `Publisher` interface | ❌ Missing | Publish method |
| `Subscriber` interface | ❌ Missing | Receive method |
| Concrete implementations | ❌ Missing | Channel-based |

#### Files Status
| File | Status |
|------|-------|
| `pkg/mail/publisher.go` | ❌ MISSING - create |
| `pkg/mail/subscriber.go` | ❌ MISSING - create |

### Required Implementation

#### Publisher Interface
```go
type Publisher interface {
    Publish(mail Mail) (Ack, error)
}
```

#### Subscriber Interface
```go
type Subscriber interface {
    Receive() <-chan Mail
}
```

### Tests to Write (4 tests, 4 commits)

#### Test 1: TestPublisher_Publish
```go
func TestPublisher_Publish(t *testing.T)
```
**Acceptance Criteria:**
- Publisher.Publish sends mail correctly
- Returns Ack with correct fields
- Returns nil error on success

#### Test 2: TestSubscriber_Receives
```go
func TestSubscriber_Receives(t *testing.T)
```
**Acceptance Criteria:**
- Subscriber.Receive returns channel
- Mail received on channel after publish

#### Test 3: TestPublisher_SubscriberIntegration
```go
func TestPublisher_SubscriberIntegration(t *testing.T)
```
**Acceptance Criteria:**
- Publisher and Subscriber work together
- Mail flows from publisher to subscriber

#### Test 4: TestPublisherInterface_ImplementedByRouter
```go
func TestPublisherInterface_ImplementedByRouter(t *testing.T)
```
**Acceptance Criteria:**
- MailRouter implements Publisher interface
- Can use MailRouter as Publisher

### Dependencies

#### Test Dependencies
```
Test 1 → Test 2 (Publish before Receive)
Test 3 (Integration - depends on 1 and 2)
Test 4 (Interface check - depends on 1)
```

### Files to Create/Modify

| File | Operation | Changes |
|------|-----------|---------|
| `pkg/mail/publisher.go` | CREATE | Publisher interface and implementation |
| `pkg/mail/subscriber.go` | CREATE | Subscriber interface and implementation |

### Implementation Order (TDD)

1. **Test 1**: Write test → verify RED → Implement Publisher interface and concrete type → verify GREEN → commit
2. **Test 2**: Write test → verify RED → Implement Subscriber interface and concrete type → verify GREEN → commit
3. **Test 3**: Write test → verify RED → Wire Publisher and Subscriber together → verify GREEN → commit
4. **Test 4**: Write test → verify RED → Verify MailRouter implements Publisher → verify GREEN → commit

**Total: 4 tests, 4 commits**

### Deliverables
- ✅ Publisher interface in `pkg/mail/publisher.go`
- ✅ Subscriber interface in `pkg/mail/subscriber.go`
- ✅ Concrete implementations
- ✅ 4 commits (one per test per TDD workflow)

---

## Phase 3.5: Streaming Support

### Goal
Implement streaming support with StreamChunk format and taint stripping.

### Scope
- Create `pkg/mail/stream.go` with streaming support
- Implement StreamChunk structure
- Implement UpgradeToStream method
- Implement StripForbiddenTaints function
- SSE/WS upgrade logic

### Current Status

#### Implementation Status
| Component | Status | Notes |
|-----------|--------|-------|
| `StreamChunk` | ⚠️ Defined in Layer 2 | Reuse from Layer 2 |
| `UpgradeToStream()` | ❌ Missing | SSE/WS upgrade |
| `StripForbiddenTaints()` | ❌ Missing | Taint filtering |

#### Files Status
| File | Status |
|------|-------|
| `pkg/mail/stream.go` | ❌ MISSING - create |

### Required Implementation

#### StreamChunk
```go
type StreamChunk struct {
    Data     string
    Sequence int
    IsFinal  bool
    Taints   []string
}
```

#### UpgradeToStream
```go
func UpgradeToStream(sessionId string, lastEventId *string) (chan StreamChunk, error)
```

#### StripForbiddenTaints
```go
func StripForbiddenTaints(chunk StreamChunk, allowed []string) StreamChunk
```

### Tests to Write (5 tests, 5 commits)

#### Test 1: TestStreamUpgrade
```go
func TestStreamUpgrade(t *testing.T)
```
**Acceptance Criteria:**
- UpgradeToStream returns channel
- Channel is non-nil
- sessionId is used for tracking

#### Test 2: TestStreamChunkFormat
```go
func TestStreamChunkFormat(t *testing.T)
```
**Acceptance Criteria:**
- StreamChunk has Data, Sequence, IsFinal, Taints fields
- Can instantiate with all fields
- Sequence increments correctly

#### Test 3: TestTaintStripping
```go
func TestTaintStripping(t *testing.T)
```
**Acceptance Criteria:**
- StripForbiddenTaints removes disallowed taints
- Allowed taints preserved
- Empty allowed list strips all taints

#### Test 4: TestStreamChunk_IsFinal
```go
func TestStreamChunk_IsFinal(t *testing.T)
```
**Acceptance Criteria:**
- IsFinal flag indicates last chunk
- Can be used to detect stream completion

#### Test 5: TestStreamChunk_Sequence
```go
func TestStreamChunk_Sequence(t *testing.T)
```
**Acceptance Criteria:**
- Sequence numbers are unique
- Sequence can be used for ordering
- Sequence starts at 1

### Dependencies

#### Test Dependencies
```
Test 1 → Independent
Test 2 → Test 3, Test 4, Test 5 (Chunk structure)
```

### Files to Create/Modify

| File | Operation | Changes |
|------|-----------|---------|
| `pkg/mail/stream.go` | CREATE | StreamChunk, UpgradeToStream, StripForbiddenTaints |

### Implementation Order (TDD)

1. **Test 1**: Write test → verify RED → Implement UpgradeToStream stub → verify GREEN → commit
2. **Test 2**: Write test → verify RED → Define StreamChunk structure → verify GREEN → commit
3. **Test 3**: Write test → verify RED → Implement StripForbiddenTaints → verify GREEN → commit
4. **Test 4**: Write test → verify RED → Add IsFinal flag handling → verify GREEN → commit
5. **Test 5**: Write test → verify RED → Add sequence number tracking → verify GREEN → commit

**Total: 5 tests, 5 commits**

### Deliverables
- ✅ StreamChunk in `pkg/mail/stream.go`
- ✅ UpgradeToStream function
- ✅ StripForbiddenTaints function
- ✅ 5 commits (one per test per TDD workflow)

---

## Phase 3.6: Gateway Adapters

### Goal
Implement gateway channel adapters for normalizing external I/O.

### Scope
- Create `pkg/gateway/gateway.go` with Gateway service
- Create adapter implementations in `pkg/gateway/adapters/`
- Implement webhook, websocket, sse adapters (core)
- Stub pubsub, smtp, platform adapters
- Normalize inbound to mail_received
- Normalize outbound to channel-specific format

### Current Status

#### Implementation Status
| Adapter | Status | Notes |
|---------|--------|-------|
| `webhook` | ❌ Missing | HTTP POST |
| `websocket` | ❌ Missing | Bidirectional |
| `sse` | ❌ Missing | Server-Sent Events |
| `pubsub` | ❌ Missing | Message queue |
| `smtp` | ❌ Missing | Email |
| `platform` | ❌ Missing | Slack, WhatsApp, Telegram |

#### Files Status
| File | Status |
|------|-------|
| `pkg/gateway/gateway.go` | ❌ MISSING - create |
| `pkg/gateway/adapters/webhook.go` | ❌ MISSING - create |
| `pkg/gateway/adapters/websocket.go` | ❌ MISSING - create |
| `pkg/gateway/adapters/sse.go` | ❌ MISSING - create |
| `pkg/gateway/adapters/pubsub.go` | ❌ MISSING - create |
| `pkg/gateway/adapters/smtp.go` | ❌ MISSING - create |
| `pkg/gateway/adapters/platform.go` | ❌ MISSING - create |

### Required Implementation

#### Gateway Service
```go
type Gateway struct {
    adapters map[string]Adapter
    mu       sync.RWMutex
}

type Adapter interface {
    Name() string
    NormalizeInbound(data []byte) (Mail, error)
    NormalizeOutbound(mail Mail) ([]byte, error)
}
```

### Tests to Write (6 tests, 6 commits)

#### Test 1: TestGateway_RegisterAdapter
```go
func TestGateway_RegisterAdapter(t *testing.T)
```
**Acceptance Criteria:**
- Gateway can register adapters
- Adapters stored by name
- Returns nil error on success

#### Test 2: TestWebhookAdapter_NormalizeInbound
```go
func TestWebhookAdapter_NormalizeInbound(t *testing.T)
```
**Acceptance Criteria:**
- Webhook adapter normalizes HTTP POST to mail_received
- Source set to gateway
- Content extracted from request body

#### Test 3: TestSSEAdapter_NormalizeOutbound
```go
func TestSSEAdapter_NormalizeOutbound(t *testing.T)
```
**Acceptance Criteria:**
- SSE adapter formats mail for SSE output
- Returns properly formatted SSE data
- Handles streaming chunks

#### Test 4: TestWebSocketAdapter_Bidirectional
```go
func TestWebSocketAdapter_Bidirectional(t *testing.T)
```
**Acceptance Criteria:**
- WebSocket adapter handles inbound normalization
- WebSocket adapter handles outbound normalization
- Both directions work correctly

#### Test 5: TestGateway_AdapterNotFound
```go
func TestGateway_AdapterNotFound(t *testing.T)
```
**Acceptance Criteria:**
- Using non-registered adapter returns error
- Error indicates adapter not found

#### Test 6: TestAdapter_NormalizationRoundTrip
```go
func TestAdapter_NormalizationRoundTrip(t *testing.T)
```
**Acceptance Criteria:**
- Inbound then outbound preserves content
- Mail type preserved through normalization

### Dependencies

#### Test Dependencies
```
Test 1 → Test 2, Test 3, Test 4 (Register before use)
Test 5 (Error handling - independent)
Test 6 (Round-trip - depends on 2 and 3)
```

### Files to Create/Modify

| File | Operation | Changes |
|------|-----------|---------|
| `pkg/gateway/gateway.go` | CREATE | Gateway service and Adapter interface |
| `pkg/gateway/adapters/webhook.go` | CREATE | Webhook adapter |
| `pkg/gateway/adapters/websocket.go` | CREATE | WebSocket adapter |
| `pkg/gateway/adapters/sse.go` | CREATE | SSE adapter |
| `pkg/gateway/adapters/pubsub.go` | CREATE | Pubsub adapter stub |
| `pkg/gateway/adapters/smtp.go` | CREATE | SMTP adapter stub |
| `pkg/gateway/adapters/platform.go` | CREATE | Platform adapter stub |

### Implementation Order (TDD)

1. **Test 1**: Write test → verify RED → Implement Gateway and Adapter interface → verify GREEN → commit
2. **Test 2**: Write test → verify RED → Implement Webhook adapter → verify GREEN → commit
3. **Test 3**: Write test → verify RED → Implement SSE adapter → verify GREEN → commit
4. **Test 4**: Write test → verify RED → Implement WebSocket adapter → verify GREEN → commit
5. **Test 5**: Write test → verify RED → Add error handling for missing adapters → verify GREEN → commit
6. **Test 6**: Write test → verify RED → Verify round-trip normalization → verify GREEN → commit

**Total: 6 tests, 6 commits**

### Deliverables
- ✅ Gateway service in `pkg/gateway/gateway.go`
- ✅ Webhook, WebSocket, SSE adapters implemented
- ✅ Pubsub, SMTP, Platform adapter stubs
- ✅ 6 commits (one per test per TDD workflow)

---

## Phase 3.7: Human Gateway Service

### Goal
Implement sys:human-gateway service for human chat with running agents.

### Scope
- Create `pkg/services/humangateway/service.go`
- Implement HTTPS chat session at `/chat/{agentId}`
- Session receives sanitized ContextMap snapshot
- Messages sent as human_feedback or user mail type
- Optional action item shorthand (@pause, @inject-memory)

### Current Status

#### Implementation Status
| Component | Status | Notes |
|-----------|--------|-------|
| `HumanGatewayService` | ❌ Missing | Chat service |
| Chat session handling | ❌ Missing | HTTPS endpoint |
| Action item parsing | ❌ Missing | @pause, @inject-memory |

#### Files Status
| File | Status |
|------|-------|
| `pkg/services/humangateway/service.go` | ❌ MISSING - create |
| `pkg/services/humangateway/chat.go` | ❌ MISSING - create |

### Required Implementation

#### HumanGatewayService
```go
type HumanGatewayService struct {
    id       string
    sessions map[string]*ChatSession
    mu       sync.RWMutex
}

type ChatSession struct {
    AgentID   string
    Messages  []Mail
    ContextMap ContextMapSnapshot
}

func (h *HumanGatewayService) HandleChat(agentId string, message string) (Mail, error)
func (h *HumanGatewayService) ParseActionItem(message string) ([]ActionItem, error)
```

### Tests to Write (4 tests, 4 commits)

#### Test 1: TestHumanGatewayService_HandleChat
```go
func TestHumanGatewayService_HandleChat(t *testing.T)
```
**Acceptance Criteria:**
- HandleChat creates mail for agent
- Mail type is human_feedback or user
- Returns nil error on success

#### Test 2: TestHumanGatewayService_ParseActionItem
```go
func TestHumanGatewayService_ParseActionItem(t *testing.T)
```
**Acceptance Criteria:**
- @pause parsed correctly
- @inject-memory X parsed correctly
- Returns ActionItem slice

#### Test 3: TestHumanGatewayService_SessionManagement
```go
func TestHumanGatewayService_SessionManagement(t *testing.T)
```
**Acceptance Criteria:**
- Chat sessions created per agent
- Sessions store messages
- ContextMap snapshot available

#### Test 4: TestHumanGatewayService_SanitizedContext
```go
func TestHumanGatewayService_SanitizedContext(t *testing.T)
```
**Acceptance Criteria:**
- ContextMap snapshot is sanitized
- Boundary rules applied
- Forbidden taints stripped

### Dependencies

#### Test Dependencies
```
Test 1 → Test 3 (HandleChat before SessionManagement)
Test 2 (Independent - ActionItem parsing)
Test 4 (Depends on 3 - Context sanitization)
```

### Files to Create/Modify

| File | Operation | Changes |
|------|-----------|---------|
| `pkg/services/humangateway/service.go` | CREATE | HumanGatewayService |
| `pkg/services/humangateway/chat.go` | CREATE | ChatSession and action items |

### Implementation Order (TDD)

1. **Test 1**: Write test → verify RED → Implement HandleChat method → verify GREEN → commit
2. **Test 2**: Write test → verify RED → Implement ParseActionItem method → verify GREEN → commit
3. **Test 3**: Write test → verify RED → Implement session management → verify GREEN → commit
4. **Test 4**: Write test → verify RED → Implement context sanitization → verify GREEN → commit

**Total: 4 tests, 4 commits**

### Deliverables
- ✅ HumanGatewayService in `pkg/services/humangateway/service.go`
- ✅ ChatSession handling in `pkg/services/humangateway/chat.go`
- ✅ Action item parsing (@pause, @inject-memory)
- ✅ 4 commits (one per test per TDD workflow)

---

## Phase 3.8: Integration & CommunicationService

### Goal
Wire Mail system to CommunicationService and verify full integration.

### Scope
- Update CommunicationService to use MailRouter
- Verify full mail flow end-to-end
- Integration tests for all components

### Current Status

#### Implementation Status
| Component | Status | Notes |
|-----------|--------|-------|
| CommunicationService integration | ⚠️ Partial | Stub exists in Layer 2 |
| Full mail flow | ❌ Missing | End-to-end test |

#### Files Status
| File | Status |
|------|-------|
| `pkg/services/communication/service.go` | ⚠️ MODIFY - integrate MailRouter |
| `pkg/mail/mail_test.go` | ❌ MISSING - create comprehensive tests |

### Required Implementation

#### CommunicationService Integration
```go
type CommunicationService struct {
    id         string
    router     *MailRouter
    publisher  Publisher
    subscribers map[string][]Subscriber
    mu         sync.RWMutex
}

func (s *CommunicationService) HandleMail(mail Mail) error
```

### Tests to Write (2 tests, 2 commits)

#### Test 1: TestFullMailFlow
```go
func TestFullMailFlow(t *testing.T)
```
**Acceptance Criteria:**
- Mail created with all fields
- Mail routed through MailRouter
- Mail delivered to subscriber
- Ack returned with correct fields

#### Test 2: TestCommunicationService_Integration
```go
func TestCommunicationService_Integration(t *testing.T)
```
**Acceptance Criteria:**
- CommunicationService uses MailRouter
- Publish/Subscribe flow works end-to-end
- All components integrated correctly

### Dependencies

#### Test Dependencies
```
Test 1 → Test 2 (Full flow before Integration)
```

### Files to Create/Modify

| File | Operation | Changes |
|------|-----------|---------|
| `pkg/services/communication/service.go` | MODIFY | Integrate MailRouter |
| `pkg/mail/mail_test.go` | CREATE | Comprehensive integration tests |

### Implementation Order (TDD)

1. **Test 1**: Write test → verify RED → Wire MailRouter into CommunicationService → verify GREEN → commit
2. **Test 2**: Write test → verify RED → Verify full integration → verify GREEN → commit

**Total: 2 tests, 2 commits**

### Deliverables
- ✅ CommunicationService integrated with MailRouter
- ✅ Full mail flow working end-to-end
- ✅ 2 commits (one per test per TDD workflow)

---

## Summary

### Phase Summary Table

| Phase | Tests | Files | Branch | Dependencies | Doc |
|-------|-------|-------|--------|--------------|-----|
| 3.1 | 6 | 2 | `feat/layer3-mail-types` | None | [phase-3.1-mail-types.md](phase-3.1-mail-types.md) |
| 3.2 | 5 | 1 | `feat/layer3-mail-router` | 3.1 | [phase-3.2-mail-router.md](phase-3.2-mail-router.md) |
| 3.3 | 6 | 1 | `feat/layer3-inboxes` | 3.1 | [phase-3.3-inboxes.md](phase-3.3-inboxes.md) |
| 3.4 | 4 | 2 | `feat/layer3-pubsub` | 3.1, 3.3 | [phase-3.4-pubsub.md](phase-3.4-pubsub.md) |
| 3.5 | 5 | 1 | `feat/layer3-streaming` | 3.1 | [phase-3.5-streaming.md](phase-3.5-streaming.md) |
| 3.6 | 6 | 5 | `feat/layer3-gateway` | 3.1, 3.5 | [phase-3.6-gateway.md](phase-3.6-gateway.md) |
| 3.7 | 4 | 2 | `feat/layer3-human-gateway` | 3.1, 3.6 | [phase-3.7-human-gateway.md](phase-3.7-human-gateway.md) |
| 3.8 | 2 | 2 | `feat/layer3-integration` | 3.2, 3.3, 3.4, 3.5 | [phase-3.8-integration.md](phase-3.8-integration.md) |

**Total: 38 tests, 16 files, 8 phases**

### Execution Order

```
Phase 3.1 (Mail Core Types)
    ↓
┌──────────────────────────────────────────────────┐
│     Phases 3.2-3.5 (can run in parallel)         │
│  ┌──────────┬─────────────┬──────────┬─────────┐ │
│  │ 3.2 Router│ 3.3 Inboxes │ 3.4 PubSub│ 3.5 Str│ │
│  └──────────┴─────────────┴──────────┴─────────┘ │
└──────────────────────────────────────────────────┘
    ↓
┌──────────────────────────────────────────────────┐
│     Phases 3.6-3.7 (can run in parallel)         │
│  ┌──────────┬──────────────────────────────────┐ │
│  │ 3.6 Gw   │ 3.7 Human Gateway                │ │
│  └──────────┴──────────────────────────────────┘ │
└──────────────────────────────────────────────────┘
    ↓
Phase 3.8 (Integration)
```

### Files to Create

| File | Phase | Purpose |
|------|-------|---------|
| `pkg/mail/types.go` | 3.1 | Mail, MailType, MailMetadata, Ack |
| `pkg/mail/address.go` | 3.1 | Address validation |
| `pkg/mail/router.go` | 3.2 | MailRouter |
| `pkg/mail/inbox.go` | 3.3 | AgentInbox, Topic, ServiceInbox |
| `pkg/mail/publisher.go` | 3.4 | Publisher interface |
| `pkg/mail/subscriber.go` | 3.4 | Subscriber interface |
| `pkg/mail/stream.go` | 3.5 | StreamChunk, streaming |
| `pkg/gateway/gateway.go` | 3.6 | Gateway service |
| `pkg/gateway/adapters/*.go` | 3.6 | Channel adapters |
| `pkg/services/humangateway/service.go` | 3.7 | HumanGatewayService |
| `pkg/services/humangateway/chat.go` | 3.7 | ChatSession |
| `pkg/mail/mail_test.go` | 3.8 | Integration tests |

### Files to Modify

| File | Phase | Changes |
|------|-------|---------|
| `pkg/services/communication/service.go` | 3.8 | Integrate MailRouter |

### Next Steps

1. **Start Phase 3.1**: Create branch `feat/layer3-mail-types`
2. **Follow TDD workflow**: Write test → verify RED → Implement → verify GREEN → commit
3. **Repeat for each phase**: One branch per phase, merge after completion
4. **Run full test suite** after Phase 3.8 to verify integration

---

## Appendix A: Layer 2 Patterns Applied

This phase breakdown follows Layer 2 patterns:
- ✅ **1:1 test-to-commit ratio** (38 tests = 38 commits)
- ✅ **One branch per phase** (8 branches total)
- ✅ **Descriptive test names** (`Test[Component]_[Behavior]_[ExpectedResult]`)
- ✅ **Small, focused phases** (2-6 tests per phase)
- ✅ **Explicit dependencies** between phases
- ✅ **Clear acceptance criteria** for each test

## Appendix B: Recursive Analysis

### Phases That Could Be Split Further

| Phase | Current Tests | Analysis | Recommendation |
|-------|---------------|----------|----------------|
| 3.1 | 6 | Multiple independent types | Could split into 3.1a (types), 3.1b (addressing) |
| 3.6 | 6 | Multiple adapters | Could split per adapter type |
| 3.3 | 6 | Three inbox types | Could split into 3.3a (AgentInbox), 3.3b (Topic), 3.3c (ServiceInbox) |

### Recursive Split Decisions

**Phase 3.1: Keep as-is**
- 6 tests is within acceptable range (2-5 recommended, 6 is close)
- Types are tightly coupled (Mail depends on MailType, MailMetadata)
- Single coherent feature: Mail type definitions

**Phase 3.3: Split recommended**
- 6 tests across 3 independent inbox types
- AgentInbox, Topic, ServiceInbox are loosely coupled
- Split into:
  - 3.3a: AgentInbox (2 tests)
  - 3.3b: Topic (3 tests)
  - 3.3c: ServiceInbox (1 test) - could merge with 3.3a

**Phase 3.6: Split recommended**
- 6 tests across multiple adapters
- Adapters are independent implementations
- Split into:
  - 3.6a: Gateway core + Webhook (2 tests)
  - 3.6b: SSE + WebSocket (2 tests)
  - 3.6c: Stubs + Integration (2 tests)

### Final Phase Count After Recursive Split

| Original Phase | Split Into | New Test Count |
|----------------|------------|----------------|
| 3.1 | 3.1 (keep) | 6 |
| 3.2 | 3.2 (keep) | 5 |
| 3.3 | 3.3a, 3.3b | 3 + 3 |
| 3.4 | 3.4 (keep) | 4 |
| 3.5 | 3.5 (keep) | 5 |
| 3.6 | 3.6a, 3.6b, 3.6c | 2 + 2 + 2 |
| 3.7 | 3.7 (keep) | 4 |
| 3.8 | 3.8 (keep) | 2 |

**Final Total: 10 phases, 38 tests**