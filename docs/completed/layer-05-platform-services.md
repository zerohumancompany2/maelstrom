# Layer 5: Hot-Reloadable Platform Services (YAML)

**Reference**: `docs/arch-v1.md` Section 7.2  
**Parent Scope**: `docs/planning/implementation-scope.md`  
**Status**: ❌ PENDING

---

## Overview

Layer 5 implements the YAML-defined platform services that are loaded via ChartRegistry after `kernel_ready`. These are Charts that happen to be system services, using the same YAML schema as user agents. They depend on Layer 4 (Security) for boundary enforcement.

**Key Characteristic**: All services are Charts running on the Statechart Engine, declared in YAML, hot-reloadable, and addressable by well-known IDs (`sys:*`).

---

## Architectural Dependencies

```
Layer 5 depends on:
├── Layer 4 (Security & Boundaries)
│   └── Security enforces boundaries on these services
├── Layer 3 (Communication & Mail System)
│   └── Mail backbone for service communication
├── Layer 2 (Core Platform Services)
│   └── sys:lifecycle for spawning services
├── Layer 1 (Kernel & Bootstrap)
│   └── ChartRegistry for YAML loading
└── Layer 0 (Statechart Engine) ✓ COMPLETE
    └── All services are Charts
```

**Why Layer 4 First**: Security enforces boundaries on these services. Without security, services could leak sensitive data across boundaries.

---

## Service Overview

All services are loaded post-`kernel_ready` via ChartRegistry. Each service is a Chart with a well-known ID.

| Service ID | Key Capability | Depends On |
|------------|---------------|------------|
| `sys:gateway` | HTTP/SSE/WS, 2FA, OpenAPI | sys:communication, sys:security |
| `sys:admin` | k9s-style terminal/web | sys:gateway (outer-only) |
| `sys:persistence` | Snapshots, event sourcing | sys:observability |
| `sys:heartbeat` | Scheduled wake-ups | sys:lifecycle |
| `sys:memory` | Vector/graph stores | sys:persistence |
| `sys:tools` | Tool registry | sys:lifecycle |
| `sys:datasources` | Storage plugins | sys:security, DataSources |
| `sys:human-gateway` | Human-in-the-loop chat | sys:gateway |

---

## 1. sys:gateway

**ID**: `sys:gateway`  
**Purpose**: Channel adapters (HTTP/SSE/WS/Email/Slack/etc.), OpenAPI generation, 2FA enforcement

### Channel Adapters

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

### Normalization Behavior

- **Inbound**: Each adapter normalizes traffic to `mail_received` Message
- **Outbound**: Each adapter normalizes outbound `mail_send` to channel's specific format

### HTTP Exposure & OpenAPI

Any Chart can declare HTTP endpoints:

```yaml
expose:
  http:
    path: /api/v1/agents/{id}/
    events:
      - trigger: user_query
        method: POST
        requestSchema: UserQuery
        responseSchema: AgentResponse
        stream: true
      - trigger: status
        method: GET
```

**Key Points**:
- `sys:gateway` automatically registers these as **OpenAPI 3.1 endpoints** with **auth middleware**
- The event surface of the chart becomes the API surface
- **Inner-boundary Charts are never directly exposed**; only DMZ/outer ones

### Service API

```go
type GatewayService struct {
    id          string  // "sys:gateway"
    adapters    map[string]ChannelAdapter
    openAPI     *OpenAPISpec
    mu          sync.RWMutex
}

type ChannelAdapter interface {
    NormalizeInbound(mail Mail) (Mail, error)
    NormalizeOutbound(mail Mail) (any, error)
    Start() error
    Stop() error
}

func NewGatewayService() *GatewayService

func (s *GatewayService) ID() string

func (s *GatewayService) HandleMail(mail Mail) error

func (s *GatewayService) RegisterAdapter(name string, adapter ChannelAdapter) error

func (s *GatewayService) GetOpenAPI() *OpenAPISpec
```

---

## 2. sys:admin

**ID**: `sys:admin`  
**Purpose**: k9s-style terminal/web console for debugging (outer-only, 2FA-gated)

### Features

- List/pause/resume/stop Charts
- View session/state/taints
- Inject events
- 2FA/MFA required for all access
- Operate only on outer boundary

### Service API

```go
type AdminService struct {
    id       string  // "sys:admin"
    auth     AuthManager
    mu       sync.RWMutex
}

type AuthManager interface {
    Verify2FA(token string) (bool, error)
    CreateToken(userID string) (string, error)
}

func NewAdminService() *AdminService

func (s *AdminService) ID() string

func (s *AdminService) HandleMail(mail Mail) error

func (s *AdminService) AdminCommand(cmd AdminCommand, authToken string) (any, error)

type AdminCommand struct {
    Type       string  // list, pause, resume, stop, queryTaints, inject
    TargetID   string
    Event      *Event  // for inject
}
```

---

## 3. sys:persistence

**ID**: `sys:persistence`  
**Purpose**: Snapshots, event sourcing, version migration

### Snapshot Triggers

```yaml
persistence:
  enabled: true
  snapshotOn:
    - stateEntry: "decide"
    - every: 50 messages
    - cron: "*/15 * * * *"
```

### Service API

```go
type PersistenceService struct {
    id       string  // "sys:persistence"
    storage  StorageBackend
    mu       sync.RWMutex
}

type StorageBackend interface {
    SaveSnapshot(snapshot Snapshot) error
    LoadSnapshot(id string) (Snapshot, error)
    SaveEvent(event Event) error
    QueryEvents(filters EventFilters) ([]Event, error)
}

func NewPersistenceService() *PersistenceService

func (s *PersistenceService) ID() string

func (s *PersistenceService) HandleMail(mail Mail) error

func (s *PersistenceService) Snapshot(runtimeId string) (Snapshot, error)

func (s *PersistenceService) Restore(snapshotId string) (statechart.RuntimeID, error)
```

---

## 4. sys:heartbeat

**ID**: `sys:heartbeat`  
**Purpose**: Scheduled agent wake-ups with HEARTBEAT.md injection

### Service API

```go
type HeartbeatService struct {
    id       string  // "sys:heartbeat"
    scheduler *Scheduler
    mu       sync.RWMutex
}

type Scheduler interface {
    Schedule(cron string, handler func()) error
    Cancel(jobID string) error
}

func NewHeartbeatService() *HeartbeatService

func (s *HeartbeatService) ID() string

func (s *HeartbeatService) HandleMail(mail Mail) error

func (s *HeartbeatService) ScheduleAgentWake(agentID string, cron string) error
```

---

## 5. sys:memory

**ID**: `sys:memory`  
**Purpose**: Long-term memory (vector/graph stores) for ContextMap injection

### Service API

```go
type MemoryService struct {
    id       string  // "sys:memory"
    vectorDB VectorStore
    graphDB  GraphStore
    mu       sync.RWMutex
}

type VectorStore interface {
    Embed(content string) ([]float32, error)
    Search(query []float32, topK int) ([]MemoryItem, error)
    Store(item MemoryItem) error
}

type GraphStore interface {
    AddEdge(from, to string, properties any) error
    Query(pattern GraphPattern) ([]GraphNode, error)
}

func NewMemoryService() *MemoryService

func (s *MemoryService) ID() string

func (s *MemoryService) HandleMail(mail Mail) error

func (s *MemoryService) Query(vector []float32, topK int, boundaryFilter BoundaryType) ([]MemorySlice, error)
```

---

## 6. sys:tools

**ID**: `sys:tools`  
**Purpose**: Tool registry and resolution

### Tool Registry

```go
type ToolRegistry struct {
    tools     map[string]ToolDescriptor
    mu        sync.RWMutex
}

type ToolDescriptor struct {
    Name        string
    Boundary    BoundaryType
    Schema      JSONSchema
    Isolation   string  // container, process, sandbox, strict
    TaintOutput []string
    Type        string  // tool, attachedSubAgent
    ChartRef    string  // for sub-agents
}

func NewToolRegistry() *ToolRegistry

func (r *ToolRegistry) Register(tool ToolDescriptor) error

func (r *ToolRegistry) Resolve(name string, callerBoundary BoundaryType) (ToolDescriptor, error)

func (r *ToolRegistry) ListTools(boundary BoundaryType) []ToolDescriptor
```

### Tool Resolution

```go
resolveTool(name string, callerBoundary BoundaryType) → ToolDescriptor | notFound
```
- Returns **sanitized schema filtered by caller's boundary**
- Inner tools are **invisible to outer callers**

---

## 7. sys:datasources

**ID**: `sys:datasources`  
**Purpose**: Pluggable FS/object storage with tainting

### Service API

```go
type DataSourceService struct {
    id        string  // "sys:datasources"
    registry  *DataSourceRegistry
    mu        sync.RWMutex
}

func NewDataSourceService() *DataSourceService

func (s *DataSourceService) ID() string

func (s *DataSourceService) HandleMail(mail Mail) error

func (s *DataSourceService) TagOnWrite(path string, taints []string) error

func (s *DataSourceService) GetTaints(path string) ([]string, error)

func (s *DataSourceService) RegisterDataSource(name string, ds DataSource) error
```

---

## 8. sys:human-gateway

**ID**: `sys:human-gateway`  
**Purpose**: Chat interface for human-in-the-loop with running agents

### Features

- HTTPS chat session at `/chat/{agentId}`
- Session receives read-only snapshot of Agent's ContextMap + last N Messages (sanitized by boundary rules)
- Messages sent become `mail_received` (type: `human_feedback` or `user`) delivered to Agent's inbox
- Agent replies via normal mail → rendered back in chat UI
- **Optional "action item" shorthand**: `@pause`, `@inject-memory X`, etc. become special Mail messages

### Service API

```go
type HumanGatewayService struct {
    id       string  // "sys:human-gateway"
    sessions map[string]*ChatSession
    mu       sync.RWMutex
}

type ChatSession struct {
    AgentID    string
    Messages   []Mail
    ContextMap []ContextBlock
    mu         sync.RWMutex
}

func NewHumanGatewayService() *HumanGatewayService

func (s *HumanGatewayService) ID() string

func (s *HumanGatewayService) HandleMail(mail Mail) error

func (s *HumanGatewayService) CreateSession(agentID string) (*ChatSession, error)

func (s *HumanGatewayService) SendMessage(sessionID string, message Mail) error

func (s *HumanGatewayService) GetSession(sessionID string) (*ChatSession, error)
```

---

## Service YAML Schema (arch-v1.md 13.7)

```yaml
apiVersion: maelstrom.dev/v1
kind: PlatformService
metadata:
  name: sys:gateway
  core: true                    # treated specially by Registry
spec:
  chartRef: gateway-v1
  requiredForKernelReady: true
  replicas: 1
  persistence:
    snapshotEvery: "100 messages"
  expose:
    http:
      port: 8080
      tls: true
```

---

## Hot-Reload Protocol (arch-v1.md 12.3)

```
1. New ChartDefinition loaded by Registry
2. Registry signals current ChartRuntime: prepareForReload
3. ChartRuntime attempts to reach quiescence within timeoutMs
4. IF quiescence reached:
   - Stop current runtime
   - Spawn new runtime with history (shallow/deep)
   - Apply contextTransform if version changed
5. IF timeout expires:
   - Force-stop current runtime
   - cleanStart (no history preserved)
   - Increment reload attempt counter
6. IF maxWaitAttempts exceeded:
   - Log permanent failure to sys:observability
   - Require manual intervention via sys:admin
```

### Quiescence Definition

A ChartRuntime is quiescent when:
1. Event queue is empty (no pending events)
2. No active parallel regions are processing events
3. No inflight tool calls or sub-agent invocations (Orchestrator idle)

### History Mechanisms

- **shallowHistory**: Restore to parent state's default sub-state
- **deepHistory**: Restore to specific sub-state (if still exists)
- **Deleted state fallback**: If deepHistory target deleted, fall back to shallow

### Context Transform

- Go template with access to `oldContext`, `newVersion`, `contextVersion`
- Executed on version change; failure → cleanStart fallback
- Registry SHALL validate template syntax at load time

---

## File Structure

```
maelstrom/
├── pkg/services/
│   ├── gateway/
│   │   ├── service.go         # GatewayService (~200 lines)
│   │   ├── adapters/
│   │   │   ├── webhook.go
│   │   │   ├── websocket.go
│   │   │   ├── sse.go
│   │   │   ├── pubsub.go
│   │   │   ├── smtp.go
│   │   │   └── platform.go    # slack, whatsapp, telegram
│   │   └── service_test.go
│   ├── admin/
│   │   ├── service.go         # AdminService (~150 lines)
│   │   └── service_test.go
│   ├── persistence/
│   │   ├── service.go         # PersistenceService (~150 lines)
│   │   ├── storage.go         # Storage backend (~100 lines)
│   │   └── service_test.go
│   ├── heartbeat/
│   │   ├── service.go         # HeartbeatService (~100 lines)
│   │   └── service_test.go
│   ├── memory/
│   │   ├── service.go         # MemoryService (~150 lines)
│   │   ├── vector.go          # Vector store (~100 lines)
│   │   ├── graph.go           # Graph store (~100 lines)
│   │   └── service_test.go
│   ├── tools/
│   │   ├── service.go         # ToolRegistry (~150 lines)
│   │   └── service_test.go
│   ├── datasources/
│   │   ├── service.go         # DataSourceService (~100 lines)
│   │   └── service_test.go
│   └── human-gateway/
│       ├── service.go         # HumanGatewayService (~150 lines)
│       └── service_test.go
```

---

## TDD Implementation Plan

### Phase 5.1: GatewayService

**Test**: `TestGatewayService_ID`
- Returns "sys:gateway"

**Test**: `TestGatewayService_RegisterAdapter`
- Adapters registered correctly

**Test**: `TestGatewayService_NormalizeInbound`
- Inbound normalization works

**Test**: `TestGatewayService_NormalizeOutbound`
- Outbound normalization works

**Implementation**:
- ~200 lines service.go + adapters

### Phase 5.2: AdminService

**Test**: `TestAdminService_ID`
- Returns "sys:admin"

**Test**: `TestAdminService_AdminCommand`
- Commands executed correctly

**Test**: `TestAdminService_2FARequired`
- 2FA enforced

**Implementation**:
- ~150 lines service.go

### Phase 5.3: PersistenceService

**Test**: `TestPersistenceService_Snapshot`
- Snapshots saved correctly

**Test**: `TestPersistenceService_Restore`
- Restores work correctly

**Test**: `TestPersistenceService_EventSourcing`
- Events stored and queried

**Implementation**:
- ~250 lines service.go + storage.go

### Phase 5.4: HeartbeatService

**Test**: `TestHeartbeatService_Schedule`
- Scheduling works

**Test**: `TestHeartbeatService_WakeAgent`
- Agent wake-ups work

**Implementation**:
- ~100 lines service.go

### Phase 5.5: MemoryService

**Test**: `TestMemoryService_StoreQuery`
- Vector storage/query works

**Test**: `TestMemoryService_GraphOperations`
- Graph operations work

**Implementation**:
- ~350 lines service.go + vector.go + graph.go

### Phase 5.6: ToolRegistry

**Test**: `TestToolRegistry_RegisterResolve`
- Registration and resolution work

**Test**: `TestToolRegistry_BoundaryFiltering`
- Boundary filtering works

**Implementation**:
- ~150 lines service.go

### Phase 5.7: DataSourceService

**Test**: `TestDataSourceService_TagOnWrite`
- Taints set on write

**Test**: `TestDataSourceService_GetTaints`
- Taints retrieved correctly

**Implementation**:
- ~100 lines service.go

### Phase 5.8: HumanGatewayService

**Test**: `TestHumanGatewayService_CreateSession`
- Sessions created correctly

**Test**: `TestHumanGatewayService_SendMessage`
- Messages sent correctly

**Implementation**:
- ~150 lines service.go

### Phase 5.9: Integration

**Test**: `TestAllServicesLoaded`
- All services loaded by ChartRegistry

**Test**: `TestServicesCommunicateViaMail`
- Services exchange mail correctly

**Implementation**:
- Wire all services to ChartRegistry

---

## Dependencies & External Libraries

- `github.com/robfig/cron` - Cron scheduling (for heartbeat)
- `github.com/google/uuid` - UUID generation
- `github.com/maelstrom/v3/pkg/statechart` - Library interface
- `github.com/maelstrom/v3/pkg/mail` - Mail types

---

## Risk Assessment

| Risk | Likelihood | Impact | Mitigation |
|------|-----------|--------|------------|
| Service startup order bugs | Medium | Medium | Explicit dependencies in YAML |
| Hot-reload race conditions | Low | High | Quiescence-based reload |
| Context transform failures | Medium | Medium | Validate templates at load time |
| Storage backend failures | Low | Medium | Error handling and fallbacks |

---

## Open Questions

1. **Storage backends**: In-memory only, or persistent (postgres, s3, etc.)?
2. **Vector database**: Which implementation (pgvector, milvus, etc.)?
3. **Graph database**: Which implementation (neo4j, graphdb, etc.)?
4. **Cron scheduler**: Simple in-memory, or persistent (hashicorp-consul)?

---

## References

- `docs/arch-v1.md` Section 7.2 - Hot-Reloadable Services
- `docs/arch-v1.md` Section 12.3 - Hot-Reload Protocol
- `docs/arch-v1.md` Section 13.7 - PlatformService YAML Schema
- `docs/planning/layer-04-security-datasources.md` - Layer 4 (dependency)
- `docs/planning/implementation-scope.md` - Parent scope document

---
