# Layer 5: Hot-Reloadable Platform Services (YAML) - Phase Breakdown

**Generated**: 2026-03-09  
**Source Spec**: `docs/arch-v1.md` v1.3 (Section 7: L449-500, Section 12.3: L856-888, Section 13.7: L1110-1128)  
**Status**: Planning document for TDD implementation

---

## Executive Summary

Layer 5 implements YAML-defined platform services that are loaded via ChartRegistry after `kernel_ready`. These are Charts that happen to be system services, using the same YAML schema as user agents. All services are addressable by well-known IDs (`sys:*`) and are hot-reloadable.

**Key Guarantee** (arch-v1.md L462-463): "Platform services SHALL be hot-reloadable without manual intervention"

### Phase Overview

| Phase | Name | Tests | Files | Dependencies | Spec Reference |
|-------|------|-------|-------|--------------|----------------|
| P1.1 | GatewayService Core | 4 | 2 | None | L56-129, L462-474 |
| P1.2 | GatewayChannelAdapters | 6 | 6 | P1.1 | L59-79, L462-474 |
| P2.1 | AdminService Core | 4 | 2 | P1.1 | L135-174, L462-474 |
| P2.2 | AdminService 2FA | 3 | 1 | P2.1 | L143-144, L462-474 |
| P3.1 | PersistenceService Core | 5 | 3 | None | L179-218, L462-474 |
| P3.2 | PersistenceStorageBackend | 4 | 2 | P3.1 | L193-218, L462-474 |
| P4.1 | HeartbeatService Core | 3 | 2 | None | L224-248, L462-474 |
| P4.2 | HeartbeatAgentWake | 3 | 1 | P4.1 | L225, L462-474 |
| P5.1 | MemoryService VectorStore | 4 | 3 | P3.1 | L254-285, L462-474 |
| P5.2 | MemoryService GraphStore | 4 | 2 | P5.1 | L254-285, L462-474 |
| P6.1 | ToolRegistry Core | 4 | 2 | None | L291-328, L462-474 |
| P6.2 | ToolRegistry BoundaryFiltering | 3 | 1 | P6.1 | L324-326, L261-270 |
| P7.1 | DataSourceService Core | 4 | 2 | Layer 4 | L332-355, L277-281 |
| P7.2 | DataSourceService TaintTracking | 4 | 1 | P7.1 | L333, L1312 |
| P8.1 | HumanGatewayService Core | 4 | 2 | P1.1 | L361-399, L462-474 |
| P8.2 | HumanGatewayService ChatSession | 4 | 1 | P8.1 | L366-370, L462-474 |
| P9.1 | HotReloadProtocol Quiescence | 4 | 2 | P3.1, P6.1 | L856-860, L462-474 |
| P9.2 | HotReloadProtocol History | 4 | 1 | P9.1 | L877-880, L462-474 |
| P9.3 | HotReloadProtocol ContextTransform | 3 | 1 | P9.2 | L882-885, L462-474 |
| P10.1 | PlatformServiceYAML Schema | 3 | 2 | All | L1110-1128, L462-474 |
| P10.2 | ServicesIntegration Mail | 4 | 3 | P1-P9 | L23-24, L462-474 |
| P10.3 | ServicesIntegration E2E | 4 | 2 | P10.2 | L11-14, L462-474 |

**Total: 87 tests, ~45 files modified/created, 22 sub-phases**

### Current State (from implementation-assessment)

**Completed**:

- ✅ Service stubs exist for all 8 services (gateway, admin, persistence, heartbeat, memory, tools, datasources, human-gateway)
- ✅ Layer 4 security infrastructure complete (taint engine, boundary enforcement, data sources)
- ✅ Layer 3 mail system complete (publish/subscribe, routing, inbox)
- ✅ Layer 2 service registry complete (Service interface, ServiceRegistry)
- ✅ Layer 1 ChartRegistry exists (YAML loading infrastructure)
- ✅ Layer 0 Statechart Engine complete

**Missing**:

- ❌ HandleMail() methods on all services
- ❌ ID() methods on most services
- ❌ Actual channel adapter implementations (webhook, websocket, sse, pubsub, smtp, slack, whatsapp, telegram, internal_grpc)
- ❌ OpenAPI generation from chart specs
- ❌ 2FA middleware implementation
- ❌ Persistent storage backends (all in-memory)
- ❌ Cron scheduler integration (needs github.com/robfig/cron)
- ❌ VectorStore implementation (embedding + search)
- ❌ GraphStore implementation (edges + queries)
- ❌ Tool invocation logic
- ❌ Hot-reload protocol (quiescence detection, context transform)
- ❌ PlatformService YAML definitions for all sys:* services

---

## Phase P1: GatewayService Core

**Goal**: Implement GatewayService with core adapter registration and normalization

### P1.1: GatewayService Core

**Parent**: Phase P1 (GatewayService Core)  
**Spec References**: arch-v1.md L466, L657-671 (GatewayService), L462-474 (Hot-Reloadable Services)  
**Status**: ⚠️ PARTIAL

#### Overview

- Tests: 4
- Commits: 4 (1:1 ratio)
- Dependencies: None

#### Requirements

From `arch-v1.md L466, L657-671` - GatewayService:

1. GatewayService must return ID "sys:gateway"
2. GatewayService must register channel adapters
3. GatewayService must normalize inbound messages to mail_received
4. GatewayService must normalize outbound messages to channel-specific format

#### Implementation Details

**Files to create/modify**:

- `pkg/services/gateway/service.go` - Complete GatewayService implementation
- `pkg/services/gateway/service_test.go` - Add tests

**Functions to implement**:

```go
func NewGatewayService() *GatewayService
func (s *GatewayService) ID() string
func (s *GatewayService) HandleMail(mail Mail) error
func (s *GatewayService) RegisterAdapter(name string, adapter ChannelAdapter) error
func (s *GatewayService) GetOpenAPI() *OpenAPISpec
```

#### Tests to Write (4 tests, 4 commits)

**Test 1: TestGatewayService_ID**

```go
func TestGatewayService_ID(t *testing.T)
```

**Acceptance Criteria**:

- GatewayService.ID() returns "sys:gateway"
- ID is immutable after construction

**Test 2: TestGatewayService_RegisterAdapter**

```go
func TestGatewayService_RegisterAdapter(t *testing.T)
```

**Acceptance Criteria**:

- Adapter registered with unique name
- Duplicate registration returns error
- Registered adapters retrievable via GetAdapter

**Test 3: TestGatewayService_NormalizeInbound**

```go
func TestGatewayService_NormalizeInbound(t *testing.T)
```

**Acceptance Criteria**:

- Inbound traffic normalized to mail_received Message
- Adapter-specific metadata preserved in Mail.Metadata

**Test 4: TestGatewayService_NormalizeOutbound**

```go
func TestGatewayService_NormalizeOutbound(t *testing.T)
```

**Acceptance Criteria**:

- Outbound mail_send normalized to channel-specific format
- Boundary enforcement applied before normalization

---

### P1.2: GatewayChannelAdapters

**Parent**: Phase P1 (GatewayService Core)  
**Spec References**: arch-v1.md L659-666 (Channel Adapters), L462-474 (Hot-Reloadable Services)  
**Status**: ❌ PENDING

#### Overview

- Tests: 6
- Commits: 6 (1:1 ratio)
- Dependencies: P1.1

#### Requirements

From `arch-v1.md L659-666` - Channel Adapters:

1. webhook adapter for HTTP POST endpoints
2. websocket adapter for full bidirectional communication
3. sse adapter for Server-Sent Events (firewall-friendly)
4. pubsub adapter for message queue integration
5. smtp adapter for email
6. platform adapters (slack, whatsapp, telegram)

#### Implementation Details

**Files to create/modify**:

- `pkg/services/gateway/adapters/webhook.go` - Webhook adapter
- `pkg/services/gateway/adapters/websocket.go` - WebSocket adapter
- `pkg/services/gateway/adapters/sse.go` - SSE adapter
- `pkg/services/gateway/adapters/pubsub.go` - PubSub adapter
- `pkg/services/gateway/adapters/smtp.go` - SMTP adapter
- `pkg/services/gateway/adapters/platform.go` - Slack, WhatsApp, Telegram adapters

**Functions to implement**:

```go
type ChannelAdapter interface {
    NormalizeInbound(mail Mail) (Mail, error)
    NormalizeOutbound(mail Mail) (any, error)
    Start() error
    Stop() error
}
```

#### Tests to Write (6 tests, 6 commits)

**Test 1: TestWebhookAdapter_NormalizeInbound**

```go
func TestWebhookAdapter_NormalizeInbound(t *testing.T)
```

**Acceptance Criteria**:

- HTTP POST normalized to mail_received
- Request body preserved in Mail.Content

**Test 2: TestWebSocketAdapter_Bidirectional**

```go
func TestWebSocketAdapter_Bidirectional(t *testing.T)
```

**Acceptance Criteria**:

- Bidirectional message flow works
- Connection state tracked

**Test 3: TestSSEAdapter_FirewallFriendly**

```go
func TestSSEAdapter_FirewallFriendly(t *testing.T)
```

**Acceptance Criteria**:

- Server-Sent Events formatted correctly
- Single HTTP connection maintained

**Test 4: TestPubSubAdapter_MessageQueue**

```go
func TestPubSubAdapter_MessageQueue(t *testing.T)
```

**Acceptance Criteria**:

- Message queue integration works
- Topic-based routing functional

**Test 5: TestSMTPAdapter_Email**

```go
func TestSMTPAdapter_Email(t *testing.T)
```

**Acceptance Criteria**:

- Email formatted correctly
- SMTP protocol compliance

**Test 6: TestPlatformAdapter_Slack**

```go
func TestPlatformAdapter_Slack(t *testing.T)
```

**Acceptance Criteria**:

- Slack API integration works
- Webhook and bot message formats supported

---

## Phase P2: AdminService

**Goal**: Implement AdminService with 2FA-gated agent control

### P2.1: AdminService Core

**Parent**: Phase P2 (AdminService)  
**Spec References**: arch-v1.md L467, L485 (AdminService), L462-474 (Hot-Reloadable Services)  
**Status**: ⚠️ PARTIAL

#### Overview

- Tests: 4
- Commits: 4 (1:1 ratio)
- Dependencies: P1.1

#### Requirements

From `arch-v1.md L467, L485` - AdminService:

1. AdminService must return ID "sys:admin"
2. AdminService must list agents
3. AdminService must control agents (pause/resume/stop)
4. AdminService must query taints

#### Implementation Details

**Files to create/modify**:

- `pkg/services/admin/service.go` - Complete AdminService implementation
- `pkg/services/admin/service_test.go` - Add tests

**Functions to implement**:

```go
func NewAdminService() *AdminService
func (s *AdminService) ID() string
func (s *AdminService) HandleMail(mail Mail) error
func (s *AdminService) AdminCommand(cmd AdminCommand, authToken string) (any, error)
func (s *AdminService) ListAgents() ([]RuntimeInfo, error)
func (s *AdminService) ControlAgent(agentID string, action string) error
```

#### Tests to Write (4 tests, 4 commits)

**Test 1: TestAdminService_ID**

```go
func TestAdminService_ID(t *testing.T)
```

**Acceptance Criteria**:

- AdminService.ID() returns "sys:admin"
- ID is immutable after construction

**Test 2: TestAdminService_ListAgents**

```go
func TestAdminService_ListAgents(t *testing.T)
```

**Acceptance Criteria**:

- All running agents listed with runtime info
- Agent state included in response

**Test 3: TestAdminService_ControlAgent**

```go
func TestAdminService_ControlAgent(t *testing.T)
```

**Acceptance Criteria**:

- Pause/resume/stop commands executed
- Agent state transitions correctly

**Test 4: TestAdminService_QueryTaints**

```go
func TestAdminService_QueryTaints(t *testing.T)
```

**Acceptance Criteria**:

- Taints queried for specified agent
- TaintMap returned with all taints

---

### P2.2: AdminService 2FA

**Parent**: Phase P2 (AdminService)  
**Spec References**: arch-v1.md L467 (AdminService 2FA), L462-474 (Hot-Reloadable Services)  
**Status**: ❌ PENDING

#### Overview

- Tests: 3
- Commits: 3 (1:1 ratio)
- Dependencies: P2.1

#### Requirements

From `arch-v1.md L467` - AdminService 2FA:

1. 2FA/MFA required for all access
2. Operate only on outer boundary

#### Implementation Details

**Files to create/modify**:

- `pkg/services/admin/auth.go` - AuthManager implementation
- `pkg/services/admin/service_test.go` - Add 2FA tests

**Functions to implement**:

```go
type AuthManager interface {
    Verify2FA(token string) (bool, error)
    CreateToken(userID string) (string, error)
}
```

#### Tests to Write (3 tests, 3 commits)

**Test 1: TestAdminService_2FARequired**

```go
func TestAdminService_2FARequired(t *testing.T)
```

**Acceptance Criteria**:

- All admin commands require valid 2FA token
- Invalid token returns error

**Test 2: TestAdminService_OuterBoundaryOnly**

```go
func TestAdminService_OuterBoundaryOnly(t *testing.T)
```

**Acceptance Criteria**:

- Admin commands only work on outer boundary agents
- Inner/DMZ boundary agents rejected

**Test 3: TestAdminService_TokenCreation**

```go
func TestAdminService_TokenCreation(t *testing.T)
```

**Acceptance Criteria**:

- Tokens created for valid users
- Tokens expire after configured duration

---

## Phase P3: PersistenceService

**Goal**: Implement PersistenceService with snapshots and event sourcing

### P3.1: PersistenceService Core

**Parent**: Phase P3 (PersistenceService)  
**Spec References**: arch-v1.md L468, L486 (PersistenceService), L462-474 (Hot-Reloadable Services)  
**Status**: ⚠️ PARTIAL

#### Overview

- Tests: 5
- Commits: 5 (1:1 ratio)
- Dependencies: None

#### Requirements

From `arch-v1.md L468, L486` - PersistenceService:

1. PersistenceService must return ID "sys:persistence"
2. PersistenceService must create snapshots
3. PersistenceService must restore from snapshots
4. PersistenceService must append events
5. PersistenceService must query events

#### Implementation Details

**Files to create/modify**:

- `pkg/services/persistence/service.go` - Complete PersistenceService implementation
- `pkg/services/persistence/service_test.go` - Add tests
- `pkg/services/persistence/storage.go` - Storage backend interface

**Functions to implement**:

```go
func NewPersistenceService() *PersistenceService
func (s *PersistenceService) ID() string
func (s *PersistenceService) HandleMail(mail Mail) error
func (s *PersistenceService) Snapshot(runtimeId string) (Snapshot, error)
func (s *PersistenceService) Restore(snapshotId string) (statechart.RuntimeID, error)
func (s *PersistenceService) AppendEvent(event Event) error
func (s *PersistenceService) GetEvents(filters EventFilters) ([]Event, error)
```

#### Tests to Write (5 tests, 5 commits)

**Test 1: TestPersistenceService_ID**

```go
func TestPersistenceService_ID(t *testing.T)
```

**Acceptance Criteria**:

- PersistenceService.ID() returns "sys:persistence"
- ID is immutable after construction

**Test 2: TestPersistenceService_Snapshot**

```go
func TestPersistenceService_Snapshot(t *testing.T)
```

**Acceptance Criteria**:

- Snapshot created with runtime state
- Snapshot includes taints and context

**Test 3: TestPersistenceService_Restore**

```go
func TestPersistenceService_Restore(t *testing.T)
```

**Acceptance Criteria**:

- Runtime restored from snapshot
- New RuntimeID returned

**Test 4: TestPersistenceService_AppendEvent**

```go
func TestPersistenceService_AppendEvent(t *testing.T)
```

**Acceptance Criteria**:

- Event appended to event log
- Event timestamp recorded

**Test 5: TestPersistenceService_QueryEvents**

```go
func TestPersistenceService_QueryEvents(t *testing.T)
```

**Acceptance Criteria**:

- Events queried with filters
- Events returned in chronological order

---

### P3.2: PersistenceStorageBackend

**Parent**: Phase P3 (PersistenceService)  
**Spec References**: arch-v1.md L468, L486 (PersistenceService Storage), L462-474 (Hot-Reloadable Services)  
**Status**: ❌ PENDING

#### Overview

- Tests: 4
- Commits: 4 (1:1 ratio)
- Dependencies: P3.1

#### Requirements

From `arch-v1.md L468, L486` - PersistenceService Storage:

1. StorageBackend must save snapshots
2. StorageBackend must load snapshots
3. StorageBackend must save events
4. StorageBackend must query events

#### Implementation Details

**Files to create/modify**:

- `pkg/services/persistence/storage.go` - StorageBackend interface and in-memory implementation
- `pkg/services/persistence/storage_test.go` - Add tests

**Functions to implement**:

```go
type StorageBackend interface {
    SaveSnapshot(snapshot Snapshot) error
    LoadSnapshot(id string) (Snapshot, error)
    SaveEvent(event Event) error
    QueryEvents(filters EventFilters) ([]Event, error)
}
```

#### Tests to Write (4 tests, 4 commits)

**Test 1: TestStorageBackend_SaveSnapshot**

```go
func TestStorageBackend_SaveSnapshot(t *testing.T)
```

**Acceptance Criteria**:

- Snapshot saved with unique ID
- Snapshot retrievable by ID

**Test 2: TestStorageBackend_LoadSnapshot**

```go
func TestStorageBackend_LoadSnapshot(t *testing.T)
```

**Acceptance Criteria**:

- Snapshot loaded correctly
- All snapshot data preserved

**Test 3: TestStorageBackend_SaveEvent**

```go
func TestStorageBackend_SaveEvent(t *testing.T)
```

**Acceptance Criteria**:

- Event saved with timestamp
- Event associated with runtime ID

**Test 4: TestStorageBackend_QueryEvents**

```go
func TestStorageBackend_QueryEvents(t *testing.T)
```

**Acceptance Criteria**:

- Events queried by runtime ID
- Events filtered by time range

---

## Phase P4: HeartbeatService

**Goal**: Implement HeartbeatService with scheduled agent wake-ups

### P4.1: HeartbeatService Core

**Parent**: Phase P4 (HeartbeatService)  
**Spec References**: arch-v1.md L469 (HeartbeatService), L462-474 (Hot-Reloadable Services)  
**Status**: ⚠️ PARTIAL

#### Overview

- Tests: 3
- Commits: 3 (1:1 ratio)
- Dependencies: None

#### Requirements

From `arch-v1.md L469` - HeartbeatService:

1. HeartbeatService must return ID "sys:heartbeat"
2. HeartbeatService must schedule wake-ups
3. HeartbeatService must unschedule wake-ups

#### Implementation Details

**Files to create/modify**:

- `pkg/services/heartbeat/service.go` - Complete HeartbeatService implementation
- `pkg/services/heartbeat/service_test.go` - Add tests

**Functions to implement**:

```go
func NewHeartbeatService() *HeartbeatService
func (s *HeartbeatService) ID() string
func (s *HeartbeatService) HandleMail(mail Mail) error
func (s *HeartbeatService) ScheduleAgentWake(agentID string, cron string) error
func (s *HeartbeatService) Unschedule(jobID string) error
```

#### Tests to Write (3 tests, 3 commits)

**Test 1: TestHeartbeatService_ID**

```go
func TestHeartbeatService_ID(t *testing.T)
```

**Acceptance Criteria**:

- HeartbeatService.ID() returns "sys:heartbeat"
- ID is immutable after construction

**Test 2: TestHeartbeatService_Schedule**

```go
func TestHeartbeatService_Schedule(t *testing.T)
```

**Acceptance Criteria**:

- Cron schedule parsed correctly
- Job registered with scheduler

**Test 3: TestHeartbeatService_Unschedule**

```go
func TestHeartbeatService_Unschedule(t *testing.T)
```

**Acceptance Criteria**:

- Job removed from scheduler
- No more wake-ups for agent

---

### P4.2: HeartbeatAgentWake

**Parent**: Phase P4 (HeartbeatService)  
**Spec References**: arch-v1.md L469 (HeartbeatService HEARTBEAT.md), L462-474 (Hot-Reloadable Services)  
**Status**: ❌ PENDING

#### Overview

- Tests: 3
- Commits: 3 (1:1 ratio)
- Dependencies: P4.1

#### Requirements

From `arch-v1.md L469` - HeartbeatService HEARTBEAT.md:

1. Scheduled wake-ups inject HEARTBEAT.md
2. Wake-up triggers agent processing
3. Wake-up respects boundary rules

#### Implementation Details

**Files to create/modify**:

- `pkg/services/heartbeat/service.go` - Add wake-up logic
- `pkg/services/heartbeat/service_test.go` - Add wake-up tests

**Functions to implement**:

```go
func (s *HeartbeatService) triggerWakeUp(agentID string) error
```

#### Tests to Write (3 tests, 3 commits)

**Test 1: TestHeartbeatService_WakeAgent**

```go
func TestHeartbeatService_WakeAgent(t *testing.T)
```

**Acceptance Criteria**:

- Agent receives HEARTBEAT.md injection
- Agent processing triggered

**Test 2: TestHeartbeatService_WakeUpBoundaryEnforcement**

```go
func TestHeartbeatService_WakeUpBoundaryEnforcement(t *testing.T)
```

**Acceptance Criteria**:

- HEARTBEAT.md respects boundary rules
- Taints attached correctly

**Test 3: TestHeartbeatService_WakeUpFailure**

```go
func TestHeartbeatService_WakeUpFailure(t *testing.T)
```

**Acceptance Criteria**:

- Failed wake-up logged to observability
- Schedule continues for future wake-ups

---

## Phase P5: MemoryService

**Goal**: Implement MemoryService with vector and graph stores

### P5.1: MemoryService VectorStore

**Parent**: Phase P5 (MemoryService)  
**Spec References**: arch-v1.md L470, L489 (MemoryService), L462-474 (Hot-Reloadable Services)  
**Status**: ❌ PENDING

#### Overview

- Tests: 4
- Commits: 4 (1:1 ratio)
- Dependencies: P3.1

#### Requirements

From `arch-v1.md L470, L489` - MemoryService:

1. MemoryService must return ID "sys:memory"
2. VectorStore must embed content
3. VectorStore must search by vector
4. VectorStore must store items

#### Implementation Details

**Files to create/modify**:

- `pkg/services/memory/service.go` - MemoryService implementation
- `pkg/services/memory/vector.go` - VectorStore implementation
- `pkg/services/memory/service_test.go` - Add tests

**Functions to implement**:

```go
type VectorStore interface {
    Embed(content string) ([]float32, error)
    Search(query []float32, topK int) ([]MemoryItem, error)
    Store(item MemoryItem) error
}
```

#### Tests to Write (4 tests, 4 commits)

**Test 1: TestMemoryService_ID**

```go
func TestMemoryService_ID(t *testing.T)
```

**Acceptance Criteria**:

- MemoryService.ID() returns "sys:memory"
- ID is immutable after construction

**Test 2: TestMemoryService_Embed**

```go
func TestMemoryService_Embed(t *testing.T)
```

**Acceptance Criteria**:

- Content embedded to vector
- Vector dimension consistent

**Test 3: TestMemoryService_VectorSearch**

```go
func TestMemoryService_VectorSearch(t *testing.T)
```

**Acceptance Criteria**:

- Search returns topK results
- Results ranked by similarity

**Test 4: TestMemoryService_StoreItem**

```go
func TestMemoryService_StoreItem(t *testing.T)
```

**Acceptance Criteria**:

- Item stored with metadata
- Item retrievable by ID

---

### P5.2: MemoryService GraphStore

**Parent**: Phase P5 (MemoryService)  
**Spec References**: arch-v1.md L470, L489 (MemoryService), L462-474 (Hot-Reloadable Services)  
**Status**: ❌ PENDING

#### Overview

- Tests: 4
- Commits: 4 (1:1 ratio)
- Dependencies: P5.1

#### Requirements

From `arch-v1.md L470, L489` - MemoryService:

1. GraphStore must add edges
2. GraphStore must query patterns
3. GraphStore must traverse relationships

#### Implementation Details

**Files to create/modify**:

- `pkg/services/memory/graph.go` - GraphStore implementation
- `pkg/services/memory/service_test.go` - Add graph tests

**Functions to implement**:

```go
type GraphStore interface {
    AddEdge(from, to string, properties any) error
    Query(pattern GraphPattern) ([]GraphNode, error)
}
```

#### Tests to Write (4 tests, 4 commits)

**Test 1: TestMemoryService_AddEdge**

```go
func TestMemoryService_AddEdge(t *testing.T)
```

**Acceptance Criteria**:

- Edge added between nodes
- Properties stored with edge

**Test 2: TestMemoryService_QueryPattern**

```go
func TestMemoryService_QueryPattern(t *testing.T)
```

**Acceptance Criteria**:

- Pattern query returns matching nodes
- Relationships traversed correctly

**Test 3: TestMemoryService_TraverseRelationships**

```go
func TestMemoryService_TraverseRelationships(t *testing.T)
```

**Acceptance Criteria**:

- Relationships traversed from node
- All connected nodes returned

**Test 4: TestMemoryService_BoundaryFilteredQuery**

```go
func TestMemoryService_BoundaryFilteredQuery(t *testing.T)
```

**Acceptance Criteria**:

- Query respects boundary filter
- Forbidden taints excluded from results

---

## Phase P6: ToolRegistry

**Goal**: Implement ToolRegistry with boundary-aware resolution

### P6.1: ToolRegistry Core

**Parent**: Phase P6 (ToolRegistry)  
**Spec References**: arch-v1.md L472, L488 (ToolRegistry), L462-474 (Hot-Reloadable Services)  
**Status**: ⚠️ PARTIAL

#### Overview

- Tests: 4
- Commits: 4 (1:1 ratio)
- Dependencies: None

#### Requirements

From `arch-v1.md L472, L488` - ToolRegistry:

1. ToolRegistry must register tools
2. ToolRegistry must resolve tools by name
3. ToolRegistry must list tools
4. ToolRegistry must invoke tools

#### Implementation Details

**Files to create/modify**:

- `pkg/services/tools/service.go` - Complete ToolRegistry implementation
- `pkg/services/tools/service_test.go` - Add tests

**Functions to implement**:

```go
func NewToolRegistry() *ToolRegistry
func (r *ToolRegistry) Register(tool ToolDescriptor) error
func (r *ToolRegistry) Resolve(name string, callerBoundary BoundaryType) (ToolDescriptor, error)
func (r *ToolRegistry) ListTools(boundary BoundaryType) []ToolDescriptor
func (r *ToolRegistry) Invoke(name string, input any) (any, error)
```

#### Tests to Write (4 tests, 4 commits)

**Test 1: TestToolRegistry_Register**

```go
func TestToolRegistry_Register(t *testing.T)
```

**Acceptance Criteria**:

- Tool registered with unique name
- ToolDescriptor stored with metadata

**Test 2: TestToolRegistry_Resolve**

```go
func TestToolRegistry_Resolve(t *testing.T)
```

**Acceptance Criteria**:

- Tool resolved by name
- ToolDescriptor returned with schema

**Test 3: TestToolRegistry_ListTools**

```go
func TestToolRegistry_ListTools(t *testing.T)
```

**Acceptance Criteria**:

- All tools listed
- ToolDescriptors returned with metadata

**Test 4: TestToolRegistry_Invoke**

```go
func TestToolRegistry_Invoke(t *testing.T)
```

**Acceptance Criteria**:

- Tool invoked with input
- Output returned with taints

---

### P6.2: ToolRegistry BoundaryFiltering

**Parent**: Phase P6 (ToolRegistry)  
**Spec References**: arch-v1.md L472, L488 (ToolRegistry Boundary), L261-270 (Boundary Model)  
**Status**: ❌ PENDING

#### Overview

- Tests: 3
- Commits: 3 (1:1 ratio)
- Dependencies: P6.1, Layer 4

#### Requirements

From `arch-v1.md L472, L488` - ToolRegistry Boundary:

1. Returns sanitized schema filtered by caller's boundary
2. Inner tools are invisible to outer callers
3. Boundary enforcement on tool resolution

#### Implementation Details

**Files to create/modify**:

- `pkg/services/tools/service.go` - Add boundary filtering
- `pkg/services/tools/service_test.go` - Add boundary tests

**Functions to implement**:

```go
func (r *ToolRegistry) filterByBoundary(tools []ToolDescriptor, boundary BoundaryType) []ToolDescriptor
```

#### Tests to Write (3 tests, 3 commits)

**Test 1: TestToolRegistry_BoundaryFiltering**

```go
func TestToolRegistry_BoundaryFiltering(t *testing.T)
```

**Acceptance Criteria**:

- Tools filtered by caller boundary
- Schema sanitized for boundary

**Test 2: TestToolRegistry_InnerToolsInvisible**

```go
func TestToolRegistry_InnerToolsInvisible(t *testing.T)
```

**Acceptance Criteria**:

- Inner boundary tools hidden from outer callers
- DMZ tools visible to DMZ and outer

**Test 3: TestToolRegistry_BoundaryEnforcement**

```go
func TestToolRegistry_BoundaryEnforcement(t *testing.T)
```

**Acceptance Criteria**:

- Boundary violation logged
- Tool resolution fails for forbidden boundaries

---

## Phase P7: DataSourceService

**Goal**: Implement DataSourceService with taint tracking

### P7.1: DataSourceService Core

**Parent**: Phase P7 (DataSourceService)  
**Spec References**: arch-v1.md L473, L490 (DataSourceService), L462-474 (Hot-Reloadable Services)  
**Status**: ⚠️ PARTIAL

#### Overview

- Tests: 4
- Commits: 4 (1:1 ratio)
- Dependencies: Layer 4 (pkg/datasource)

#### Requirements

From `arch-v1.md L473, L490` - DataSourceService:

1. DataSourceService must return ID "sys:datasources"
2. DataSourceService must get data sources
3. DataSourceService must list data sources
4. DataSourceService must register data sources

#### Implementation Details

**Files to create/modify**:

- `pkg/services/datasources/service.go` - Complete DataSourceService implementation
- `pkg/services/datasources/service_test.go` - Add tests

**Functions to implement**:

```go
func NewDataSourceService() *DataSourceService
func (s *DataSourceService) ID() string
func (s *DataSourceService) HandleMail(mail Mail) error
func (s *DataSourceService) Get(name string) (DataSource, error)
func (s *DataSourceService) List() []string
func (s *DataSourceService) RegisterDataSource(name string, ds DataSource) error
```

#### Tests to Write (4 tests, 4 commits)

**Test 1: TestDataSourceService_ID**

```go
func TestDataSourceService_ID(t *testing.T)
```

**Acceptance Criteria**:

- DataSourceService.ID() returns "sys:datasources"
- ID is immutable after construction

**Test 2: TestDataSourceService_Register**

```go
func TestDataSourceService_Register(t *testing.T)
```

**Acceptance Criteria**:

- DataSource registered with unique name
- DataSource retrievable by name

**Test 3: TestDataSourceService_Get**

```go
func TestDataSourceService_Get(t *testing.T)
```

**Acceptance Criteria**:

- DataSource retrieved by name
- DataSource interface returned

**Test 4: TestDataSourceService_List**

```go
func TestDataSourceService_List(t *testing.T)
```

**Acceptance Criteria**:

- All registered data sources listed
- Names returned as strings

---

### P7.2: DataSourceService TaintTracking

**Parent**: Phase P7 (DataSourceService)  
**Spec References**: arch-v1.md L473 (DataSourceService Tainting), L277-281 (Taint Attachment Points), L1312 (SHALL attach taints)  
**Status**: ❌ PENDING

#### Overview

- Tests: 4
- Commits: 4 (1:1 ratio)
- Dependencies: P7.1, Layer 4

#### Requirements

From `arch-v1.md L473, L277-281, L1312` - Taint Tracking:

1. TagOnWrite must set taints on data
2. GetTaints must retrieve taints from data
3. ValidateAccess must enforce boundary rules
4. Taints must be attached at file read operations

#### Implementation Details

**Files to create/modify**:

- `pkg/services/datasources/service.go` - Add taint tracking
- `pkg/services/datasources/service_test.go` - Add taint tests

**Functions to implement**:

```go
func (s *DataSourceService) TagOnWrite(path string, taints []string) error
func (s *DataSourceService) GetTaints(path string) ([]string, error)
func (s *DataSourceService) ValidateAccess(boundary security.BoundaryType) error
```

#### Tests to Write (4 tests, 4 commits)

**Test 1: TestDataSourceService_TagOnWrite**

```go
func TestDataSourceService_TagOnWrite(t *testing.T)
```

**Acceptance Criteria**:

- Taints set on write operation
- Taints persisted with data

**Test 2: TestDataSourceService_GetTaints**

```go
func TestDataSourceService_GetTaints(t *testing.T)
```

**Acceptance Criteria**:

- Taints retrieved from data
- All taints returned

**Test 3: TestDataSourceService_ValidateAccess**

```go
func TestDataSourceService_ValidateAccess(t *testing.T)
```

**Acceptance Criteria**:

- Access validated against boundary
- Forbidden access rejected

**Test 4: TestDataSourceService_TaintAttachment**

```go
func TestDataSourceService_TaintAttachment(t *testing.T)
```

**Acceptance Criteria**:

- Taints attached at file read
- Taints included in Mail.Metadata

---

## Phase P8: HumanGatewayService

**Goal**: Implement HumanGatewayService with chat sessions

### P8.1: HumanGatewayService Core

**Parent**: Phase P8 (HumanGatewayService)  
**Spec References**: arch-v1.md L471, L726-736 (HumanGatewayService), L462-474 (Hot-Reloadable Services)  
**Status**: ⚠️ PARTIAL

#### Overview

- Tests: 4
- Commits: 4 (1:1 ratio)
- Dependencies: P1.1

#### Requirements

From `arch-v1.md L471, L726-736` - HumanGatewayService:

1. HumanGatewayService must return ID "sys:human-gateway"
2. HumanGatewayService must create chat sessions
3. HumanGatewayService must get chat sessions
4. HumanGatewayService must handle chat messages

#### Implementation Details

**Files to create/modify**:

- `pkg/services/humangateway/service.go` - Complete HumanGatewayService implementation
- `pkg/services/humangateway/service_test.go` - Add tests

**Functions to implement**:

```go
func NewHumanGatewayService() *HumanGatewayService
func (s *HumanGatewayService) ID() string
func (s *HumanGatewayService) HandleMail(mail Mail) error
func (s *HumanGatewayService) CreateSession(agentID string) (*ChatSession, error)
func (s *HumanGatewayService) GetSession(sessionID string) (*ChatSession, error)
func (s *HumanGatewayService) HandleChat(sessionID string, message string) error
```

#### Tests to Write (4 tests, 4 commits)

**Test 1: TestHumanGatewayService_ID**

```go
func TestHumanGatewayService_ID(t *testing.T)
```

**Acceptance Criteria**:

- HumanGatewayService.ID() returns "sys:human-gateway"
- ID is immutable after construction

**Test 2: TestHumanGatewayService_CreateSession**

```go
func TestHumanGatewayService_CreateSession(t *testing.T)
```

**Acceptance Criteria**:

- Chat session created for agent
- SessionID returned

**Test 3: TestHumanGatewayService_GetSession**

```go
func TestHumanGatewayService_GetSession(t *testing.T)
```

**Acceptance Criteria**:

- Session retrieved by ID
- ChatSession with messages returned

**Test 4: TestHumanGatewayService_HandleChat**

```go
func TestHumanGatewayService_HandleChat(t *testing.T)
```

**Acceptance Criteria**:

- Chat message handled
- Message added to session

---

### P8.2: HumanGatewayService ChatSession

**Parent**: Phase P8 (HumanGatewayService)  
**Spec References**: arch-v1.md L471, L726-736 (HumanGatewayService Chat), L462-474 (Hot-Reloadable Services)  
**Status**: ❌ PENDING

#### Overview

- Tests: 4
- Commits: 4 (1:1 ratio)
- Dependencies: P8.1

#### Requirements

From `arch-v1.md L471, L726-736` - HumanGatewayService Chat:

1. HTTPS chat session at /chat/{agentId}
2. Session receives read-only snapshot of ContextMap + last N Messages (sanitized by boundary rules)
3. Messages become mail_received (type: human_feedback or user)
4. Optional action item shorthand (@pause, @inject-memory X)

#### Implementation Details

**Files to create/modify**:

- `pkg/services/humangateway/service.go` - Add chat session logic
- `pkg/services/humangateway/service_test.go` - Add chat tests

**Functions to implement**:

```go
func (s *HumanGatewayService) SendMessage(sessionID string, message Mail) error
func (s *HumanGatewayService) ParseActionItem(message string) (*Mail, error)
```

#### Tests to Write (4 tests, 4 commits)

**Test 1: TestHumanGatewayService_SendMessage**

```go
func TestHumanGatewayService_SendMessage(t *testing.T)
```

**Acceptance Criteria**:

- Message sent to agent inbox
- Message type set to human_feedback

**Test 2: TestHumanGatewayService_ContextMapSanitization**

```go
func TestHumanGatewayService_ContextMapSanitization(t *testing.T)
```

**Acceptance Criteria**:

- ContextMap sanitized by boundary rules
- Forbidden taints redacted

**Test 3: TestHumanGatewayService_ParseActionItem**

```go
func TestHumanGatewayService_ParseActionItem(t *testing.T)
```

**Acceptance Criteria**:

- Action items parsed (@pause, @inject-memory)
- Special Mail messages created

**Test 4: TestHumanGatewayService_AgentReply**

```go
func TestHumanGatewayService_AgentReply(t *testing.T)
```

**Acceptance Criteria**:

- Agent reply rendered in chat UI
- Reply added to session messages

---

## Phase P9: HotReloadProtocol

**Goal**: Implement hot-reload protocol with quiescence detection and history management

### P9.1: HotReloadProtocol Quiescence

**Parent**: Phase P9 (HotReloadProtocol)  
**Spec References**: arch-v1.md L856-860 (Quiescence Definition), L462-474 (Hot-Reloadable Services)  
**Status**: ❌ PENDING

#### Overview

- Tests: 4
- Commits: 4 (1:1 ratio)
- Dependencies: P3.1, P6.1

#### Requirements

From `arch-v1.md L856-860` - Quiescence Definition:

1. Event queue must be empty (no pending events)
2. No active parallel regions processing events
3. No inflight tool calls or sub-agent invocations (Orchestrator idle)

#### Implementation Details

**Files to create/modify**:

- `pkg/services/lifecycle/service.go` - Add quiescence detection
- `pkg/services/lifecycle/service_test.go` - Add quiescence tests

**Functions to implement**:

```go
func (s *LifecycleService) checkQuiescence(runtimeID string) (bool, error)
func (s *LifecycleService) prepareForReload(runtimeID string, timeoutMs int) error
```

#### Tests to Write (4 tests, 4 commits)

**Test 1: TestHotReload_QuiescenceEmptyQueue**

```go
func TestHotReload_QuiescenceEmptyQueue(t *testing.T)
```

**Acceptance Criteria**:

- Empty event queue detected
- Quiescence true when queue empty

**Test 2: TestHotReload_QuiescenceNoActiveRegions**

```go
func TestHotReload_QuiescenceNoActiveRegions(t *testing.T)
```

**Acceptance Criteria**:

- Active parallel regions detected
- Quiescence false when regions active

**Test 3: TestHotReload_QuiescenceNoInflightTools**

```go
func TestHotReload_QuiescenceNoInflightTools(t *testing.T)
```

**Acceptance Criteria**:

- Inflight tool calls detected
- Quiescence false when tools inflight

**Test 4: TestHotReload_PrepareForReload**

```go
func TestHotReload_PrepareForReload(t *testing.T)
```

**Acceptance Criteria**:

- Runtime prepared for reload
- Timeout enforced

---

### P9.2: HotReloadProtocol History

**Parent**: Phase P9 (HotReloadProtocol)  
**Spec References**: arch-v1.md L877-880 (History Mechanisms), L462-474 (Hot-Reloadable Services)  
**Status**: ❌ PENDING

#### Overview

- Tests: 4
- Commits: 4 (1:1 ratio)
- Dependencies: P9.1

#### Requirements

From `arch-v1.md L877-880` - History Mechanisms:

1. shallowHistory: Restore to parent state's default sub-state
2. deepHistory: Restore to specific sub-state (if still exists)
3. Deleted state fallback: If deepHistory target deleted, fall back to shallow

#### Implementation Details

**Files to create/modify**:

- `pkg/services/lifecycle/service.go` - Add history mechanisms
- `pkg/services/lifecycle/service_test.go` - Add history tests

**Functions to implement**:

```go
func (s *LifecycleService) restoreWithShallowHistory(snapshot Snapshot) (RuntimeID, error)
func (s *LifecycleService) restoreWithDeepHistory(snapshot Snapshot, targetState string) (RuntimeID, error)
```

#### Tests to Write (4 tests, 4 commits)

**Test 1: TestHotReload_ShallowHistory**

```go
func TestHotReload_ShallowHistory(t *testing.T)
```

**Acceptance Criteria**:

- Parent state's default sub-state restored
- New runtime spawned with history

**Test 2: TestHotReload_DeepHistory**

```go
func TestHotReload_DeepHistory(t *testing.T)
```

**Acceptance Criteria**:

- Specific sub-state restored
- State hierarchy preserved

**Test 3: TestHotReload_DeletedStateFallback**

```go
func TestHotReload_DeletedStateFallback(t *testing.T)
```

**Acceptance Criteria**:

- Deleted state detected
- Fallback to shallow history

**Test 4: TestHotReload_HistoryPreservation**

```go
func TestHotReload_HistoryPreservation(t *testing.T)
```

**Acceptance Criteria**:

- History preserved across reload
- Context maintained

---

### P9.3: HotReloadProtocol ContextTransform

**Parent**: Phase P9 (HotReloadProtocol)  
**Spec References**: arch-v1.md L882-885 (Context Transform), L462-474 (Hot-Reloadable Services)  
**Status**: ❌ PENDING

#### Overview

- Tests: 3
- Commits: 3 (1:1 ratio)
- Dependencies: P9.2

#### Requirements

From `arch-v1.md L882-885` - Context Transform:

1. Go template with access to oldContext, newVersion, contextVersion
2. Executed on version change; failure → cleanStart fallback
3. Registry SHALL validate template syntax at load time

#### Implementation Details

**Files to create/modify**:

- `pkg/services/lifecycle/service.go` - Add context transform
- `pkg/services/lifecycle/service_test.go` - Add transform tests

**Functions to implement**:

```go
func (s *LifecycleService) applyContextTransform(oldContext any, newVersion string, template string) (any, error)
func validateTransformTemplate(template string) error
```

#### Tests to Write (3 tests, 3 commits)

**Test 1: TestHotReload_ContextTransform**

```go
func TestHotReload_ContextTransform(t *testing.T)
```

**Acceptance Criteria**:

- Template executed with oldContext
- New context returned

**Test 2: TestHotReload_TransformFailureFallback**

```go
func TestHotReload_TransformFailureFallback(t *testing.T)
```

**Acceptance Criteria**:

- Transform failure detected
- Fallback to cleanStart

**Test 3: TestHotReload_TemplateValidation**

```go
func TestHotReload_TemplateValidation(t *testing.T)
```

**Acceptance Criteria**:

- Template syntax validated at load time
- Invalid template rejected

---

## Phase P10: Integration

**Goal**: Integrate all services with ChartRegistry and verify end-to-end functionality

### P10.1: PlatformServiceYAML Schema

**Parent**: Phase P10 (Integration)  
**Spec References**: arch-v1.md L1110-1128 (PlatformService YAML Schema), L462-474 (Hot-Reloadable Services)  
**Status**: ❌ PENDING

#### Overview

- Tests: 3
- Commits: 3 (1:1 ratio)
- Dependencies: All previous phases

#### Requirements

From `arch-v1.md L1110-1128` - PlatformService YAML Schema:

1. PlatformService YAML schema defined
2. ChartRegistry loads PlatformService YAMLs
3. Services marked as core: true treated specially

#### Implementation Details

**Files to create/modify**:

- `charts/platform-services/gateway.yaml` - Gateway PlatformService YAML
- `charts/platform-services/*.yaml` - All other service YAMLs

**YAML Schema**:

```yaml
apiVersion: maelstrom.dev/v1
kind: PlatformService
metadata:
  name: sys:gateway
  core: true
spec:
  chartRef: gateway-v1
  requiredForKernelReady: true
  replicas: 1
```

#### Tests to Write (3 tests, 3 commits)

**Test 1: TestPlatformServiceYAML_Schema**

```go
func TestPlatformServiceYAML_Schema(t *testing.T)
```

**Acceptance Criteria**:

- YAML schema validated
- Required fields present

**Test 2: TestPlatformServiceYAML_ChartRegistryLoad**

```go
func TestPlatformServiceYAML_ChartRegistryLoad(t *testing.T)
```

**Acceptance Criteria**:

- ChartRegistry loads PlatformService YAMLs
- Services registered correctly

**Test 3: TestPlatformServiceYAML_CoreServices**

```go
func TestPlatformServiceYAML_CoreServices(t *testing.T)
```

**Acceptance Criteria**:

- Core services treated specially
- requiredForKernelReady enforced

---

### P10.2: ServicesIntegration Mail

**Parent**: Phase P10 (Integration)  
**Spec References**: arch-v1.md L23-24 (Mail Backbone), L462-474 (Hot-Reloadable Services)  
**Status**: ❌ PENDING

#### Overview

- Tests: 4
- Commits: 4 (1:1 ratio)
- Dependencies: P1-P9

#### Requirements

From `arch-v1.md L23-24` - Mail Backbone:

1. All services communicate via mail
2. HandleMail implemented on all services
3. Services exchange messages correctly
4. Mail routing functional

#### Implementation Details

**Files to create/modify**:

- `pkg/services/integration/integration_test.go` - Integration tests
- All service files - Ensure HandleMail complete

**Functions to verify**:

```go
func (s *GatewayService) HandleMail(mail Mail) error
func (s *AdminService) HandleMail(mail Mail) error
// ... all services
```

#### Tests to Write (4 tests, 4 commits)

**Test 1: TestServicesIntegration_HandleMail**

```go
func TestServicesIntegration_HandleMail(t *testing.T)
```

**Acceptance Criteria**:

- All services implement HandleMail
- Messages handled correctly

**Test 2: TestServicesIntegration_MailExchange**

```go
func TestServicesIntegration_MailExchange(t *testing.T)
```

**Acceptance Criteria**:

- Services exchange mail correctly
- Responses routed properly

**Test 3: TestServicesIntegration_MailRouting**

```go
func TestServicesIntegration_MailRouting(t *testing.T)
```

**Acceptance Criteria**:

- Mail routing functional
- Messages delivered to correct service

**Test 4: TestServicesIntegration_MailBoundaryEnforcement**

```go
func TestServicesIntegration_MailBoundaryEnforcement(t *testing.T)
```

**Acceptance Criteria**:

- Boundary enforcement on mail exchange
- Taints tracked across services

---

### P10.3: ServicesIntegration E2E

**Parent**: Phase P10 (Integration)  
**Spec References**: arch-v1.md L11-14 (Layer 5 Overview), L462-474 (Hot-Reloadable Services)  
**Status**: ❌ PENDING

#### Overview

- Tests: 4
- Commits: 4 (1:1 ratio)
- Dependencies: P10.2

#### Requirements

From `arch-v1.md L11-14` - Layer 5 Overview:

1. All services are Charts running on Statechart Engine
2. All services are declared in YAML
3. All services are hot-reloadable
4. All services are addressable by well-known IDs (sys:*)

#### Implementation Details

**Files to create/modify**:

- `pkg/e2e/services_test.go` - E2E tests
- `pkg/e2e/testruntime.go` - E2E test runtime

**Test scenarios**:

- Full service startup
- Service communication
- Hot-reload scenario
- Service discovery

#### Tests to Write (4 tests, 4 commits)

**Test 1: TestServicesE2E_AllServicesLoaded**

```go
func TestServicesE2E_AllServicesLoaded(t *testing.T)
```

**Acceptance Criteria**:

- All 8 services loaded by ChartRegistry
- Services running on Statechart Engine

**Test 2: TestServicesE2E_ServicesAddressable**

```go
func TestServicesE2E_ServicesAddressable(t *testing.T)
```

**Acceptance Criteria**:

- All services addressable by sys:* IDs
- Service discovery functional

**Test 3: TestServicesE2E_HotReload**

```go
func TestServicesE2E_HotReload(t *testing.T)
```

**Acceptance Criteria**:

- Services hot-reloadable
- State preserved across reload

**Test 4: TestServicesE2E_FullWorkflow**

```go
func TestServicesE2E_FullWorkflow(t *testing.T)
```

**Acceptance Criteria**:

- End-to-end service workflow functional
- All dependencies satisfied

---

## Summary

### Phase Summary Table

| Phase | Tests | Files | Dependencies | Status |
|-------|-------|-------|--------------|--------|
| P1.1 | 4 | 2 | None | ⚠️ PARTIAL |
| P1.2 | 6 | 6 | P1.1 | ❌ PENDING |
| P2.1 | 4 | 2 | P1.1 | ⚠️ PARTIAL |
| P2.2 | 3 | 1 | P2.1 | ❌ PENDING |
| P3.1 | 5 | 3 | None | ⚠️ PARTIAL |
| P3.2 | 4 | 2 | P3.1 | ❌ PENDING |
| P4.1 | 3 | 2 | None | ⚠️ PARTIAL |
| P4.2 | 3 | 1 | P4.1 | ❌ PENDING |
| P5.1 | 4 | 3 | P3.1 | ❌ PENDING |
| P5.2 | 4 | 2 | P5.1 | ❌ PENDING |
| P6.1 | 4 | 2 | None | ⚠️ PARTIAL |
| P6.2 | 3 | 1 | P6.1 | ❌ PENDING |
| P7.1 | 4 | 2 | Layer 4 | ⚠️ PARTIAL |
| P7.2 | 4 | 1 | P7.1 | ❌ PENDING |
| P8.1 | 4 | 2 | P1.1 | ⚠️ PARTIAL |
| P8.2 | 4 | 1 | P8.1 | ❌ PENDING |
| P9.1 | 4 | 2 | P3.1, P6.1 | ❌ PENDING |
| P9.2 | 4 | 1 | P9.1 | ❌ PENDING |
| P9.3 | 3 | 1 | P9.2 | ❌ PENDING |
| P10.1 | 3 | 2 | All | ❌ PENDING |
| P10.2 | 4 | 3 | P1-P9 | ❌ PENDING |
| P10.3 | 4 | 2 | P10.2 | ❌ PENDING |
| **Total** | **81** | **~45** | | |

### Execution Order

```
P1 (GatewayService)
├── P1.1: GatewayService Core (4 tests)
└── P1.2: GatewayChannelAdapters (6 tests)
    ↓
P2 (AdminService)
├── P2.1: AdminService Core (4 tests)
└── P2.2: AdminService 2FA (3 tests)
    ↓
P3 (PersistenceService)
├── P3.1: PersistenceService Core (5 tests)
└── P3.2: PersistenceStorageBackend (4 tests)
    ↓
P4 (HeartbeatService)
├── P4.1: HeartbeatService Core (3 tests)
└── P4.2: HeartbeatAgentWake (3 tests)
    ↓
P5 (MemoryService)
├── P5.1: MemoryService VectorStore (4 tests)
└── P5.2: MemoryService GraphStore (4 tests)
    ↓
P6 (ToolRegistry)
├── P6.1: ToolRegistry Core (4 tests)
└── P6.2: ToolRegistry BoundaryFiltering (3 tests)
    ↓
P7 (DataSourceService)
├── P7.1: DataSourceService Core (4 tests)
└── P7.2: DataSourceService TaintTracking (4 tests)
    ↓
P8 (HumanGatewayService)
├── P8.1: HumanGatewayService Core (4 tests)
└── P8.2: HumanGatewayService ChatSession (4 tests)
    ↓
P9 (HotReloadProtocol)
├── P9.1: HotReloadProtocol Quiescence (4 tests)
├── P9.2: HotReloadProtocol History (4 tests)
└── P9.3: HotReloadProtocol ContextTransform (3 tests)
    ↓
P10 (Integration)
├── P10.1: PlatformServiceYAML Schema (3 tests)
├── P10.2: ServicesIntegration Mail (4 tests)
└── P10.3: ServicesIntegration E2E (4 tests)
```

### Next Steps

1. **Start Phase P1.1**: Create branch `feat/layer5-gateway-core`
2. **Follow TDD workflow**: Write test → verify RED → Implement → verify GREEN → commit
3. **Repeat for each phase**: One branch per phase, merge after completion
4. **Run full test suite** after final phase to verify integration

### Key Files to Create

```
pkg/services/
├── gateway/
│   ├── service.go          (P1.1)
│   ├── service_test.go     (P1.1)
│   └── adapters/           (P1.2)
│       ├── webhook.go
│       ├── websocket.go
│       ├── sse.go
│       ├── pubsub.go
│       ├── smtp.go
│       └── platform.go
├── admin/
│   ├── service.go          (P2.1)
│   ├── service_test.go     (P2.1)
│   └── auth.go             (P2.2)
├── persistence/
│   ├── service.go          (P3.1)
│   ├── service_test.go     (P3.1)
│   ├── storage.go          (P3.2)
│   └── storage_test.go     (P3.2)
├── heartbeat/
│   ├── service.go          (P4.1, P4.2)
│   └── service_test.go     (P4.1, P4.2)
├── memory/
│   ├── service.go          (P5.1)
│   ├── service_test.go     (P5.1, P5.2)
│   ├── vector.go           (P5.1)
│   └── graph.go            (P5.2)
├── tools/
│   ├── service.go          (P6.1, P6.2)
│   └── service_test.go     (P6.1, P6.2)
├── datasources/
│   ├── service.go          (P7.1, P7.2)
│   └── service_test.go     (P7.1, P7.2)
├── humangateway/
│   ├── service.go          (P8.1, P8.2)
│   └── service_test.go     (P8.1, P8.2)
└── lifecycle/
    ├── service.go          (P9.1, P9.2, P9.3)
    └── service_test.go     (P9.1, P9.2, P9.3)

charts/
└── platform-services/      (P10.1)
    ├── gateway.yaml
    ├── admin.yaml
    ├── persistence.yaml
    ├── heartbeat.yaml
    ├── memory.yaml
    ├── tools.yaml
    ├── datasources.yaml
    └── human-gateway.yaml

pkg/e2e/
├── services_test.go        (P10.3)
└── testruntime.go          (P10.3)
```

---

*Document generated: 2026-03-09*  
*Based on: arch-v1.md v1.3, layer-05-platform-services.md, implementation-assessment, implementation-patterns*
