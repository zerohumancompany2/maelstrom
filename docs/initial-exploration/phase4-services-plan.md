# Phase 4: Platform Services Implementation Plan

**Date**: 2026-03-02  
**Status**: Planning Phase  
**Based on**: arch-v1.md Section 7 (Platform Services), Section 12 (Bootstrap Sequence)

---

## Executive Summary

This plan implements **Layer 5: Platform Services** - the YAML-defined services that run on top of the Security Layer. These services provide the operational backbone for Maelstrom: gateway, admin, persistence, heartbeat, memory, tools, datasources, and human-gateway.

**Current State**: Phase 3 complete with 169 tests passing (mail system + 4 core services + security layer).

**Target**: Platform services operational with ~35 new tests, bringing total to **~204 tests**.

---

## Architectural Dependencies

From arch-v1.md Section 7 and Section 12:

```
Platform Services Dependencies:
1. Layer 0: Statechart Engine ✓ COMPLETE
2. Layer 2: Core Services (sys:security, sys:communication, etc.) ✓ COMPLETE
3. Layer 3: Mail System ✓ COMPLETE
4. Layer 4: Security & Boundaries ✓ COMPLETE (Phase 3)

New Components:
- sys:gateway - Channel adapters (HTTP/SSE/WS/Email/Slack)
- sys:admin - k9s-style terminal/console (outer-only, 2FA-gated)
- sys:persistence - Snapshots, event sourcing, version migration
- sys:heartbeat - Scheduled agent wake-ups with HEARTBEAT.md injection
- sys:memory - Long-term memory (vector/graph stores) for ContextMap injection
- sys:tools - Tool registry and resolution
- sys:datasources - Pluggable FS/object storage with tainting
- sys:human-gateway - Chat interface for human-in-the-loop

Key Invariant: All services are Charts themselves, addressable by well-known IDs
```

---

## Service 1: sys:gateway (Channel Adapters)

**Goal**: Normalize all external I/O through pluggable Channel Adapters

### Architecture

```yaml
apiVersion: maelstrom.dev/v1
kind: PlatformService
metadata:
  name: sys:gateway
  core: true
  boundary: outer
spec:
  chartRef: gateway-v1
  requiredForKernelReady: true
  adapters:
    - type: webhook
      endpoint: /webhook/*
    - type: websocket
      path: /ws
    - type: sse
      path: /sse
    - type: smtp
      domain: agents.example.com
  expose:
    http:
      port: 8080
      tls: true
    openapi: true
```

### Components

| Component | File | Purpose | Lines |
|-----------|------|---------|-------|
| GatewayService | `pkg/services/gateway/service.go` | Main service orchestrator | ~150 |
| ChannelAdapter | `pkg/services/gateway/adapter.go` | Base adapter interface | ~50 |
| WebhookAdapter | `pkg/services/gateway/webhook.go` | HTTP POST endpoint adapter | ~100 |
| WebSocketAdapter | `pkg/services/gateway/websocket.go` | Bidirectional WS adapter | ~120 |
| SSEAdapter | `pkg/services/gateway/sse.go` | Server-Sent Events adapter | ~100 |
| SMTPAdapter | `pkg/services/gateway/smtp.go` | Email adapter | ~80 |
| Test Suite | `pkg/services/gateway/service_test.go` | 8 tests | ~200 |

### API Contract

```go
type GatewayService interface {
    RegisterAdapter(name string, adapter ChannelAdapter) error
    Publish(mail Mail) (Ack, error)
    Subscribe(address string) (<-chan Mail, error)
    Unsubscribe(address string, ch <-chan Mail) error
    GetOpenAPI() (*openapi.Spec, error)
}

type ChannelAdapter interface {
    Name() string
    Handle(r *http.Request) error
    Stream() bool
}
```

### Test Specifications

| Test | Semantics Covered |
|------|-------------------|
| `TestGateway_RegisterAdapter` | Register webhook, websocket, sse, smtp adapters |
| `TestGateway_WebhookIngress` | HTTP POST → mail_received event |
| `TestGateway_WebSocketBidir` | Bidirectional WS messaging |
| `TestGateway_SSEStreaming` | SSE chunk streaming |
| `TestGateway_SMTPIngress` | Email → mail_received |
| `TestGateway_OpenAPI` | Auto-generate OpenAPI from chart events |
| `TestGateway_BoundaryOuter` | Gateway only accepts outer boundary mail |
| `TestGateway_AdapterNotFound` | Unknown adapter returns error |

**Commit**: `feat: sys:gateway with channel adapters (webhook, websocket, sse, smtp)`

---

## Service 2: sys:admin (Terminal Console)

**Goal**: k9s-style terminal/web console for debugging (outer-only, 2FA-gated)

### Architecture

```yaml
apiVersion: maelstrom.dev/v1
kind: PlatformService
metadata:
  name: sys:admin
  core: false
  boundary: outer
spec:
  chartRef: admin-v1
  requiredForKernelReady: false
  auth:
    method: 2fa
    provider: oauth2
  endpoints:
    - path: /admin/*
      methods: [GET, POST]
```

### Components

| Component | File | Purpose | Lines |
|-----------|------|---------|-------|
| AdminService | `pkg/services/admin/service.go` | Main service + 2FA middleware | ~150 |
| TerminalAdapter | `pkg/services/admin/terminal.go` | k9s-style TUI | ~200 |
| AuthMiddleware | `pkg/services/admin/auth.go` | 2FA authentication | ~100 |
| Test Suite | `pkg/services/admin/service_test.go` | 6 tests | ~150 |

### API Contract

```go
type AdminService interface {
    ListAgents() ([]AgentInfo, error)
    ControlAgent(id string, cmd string) error
    QueryTaints(agentId string) (TaintMap, error)
    InjectEvent(agentId string, event Event) error
    GetLogs(agentId string, limit int) ([]LogEntry, error)
}
```

### Test Specifications

| Test | Semantics Covered |
|------|-------------------|
| `TestAdmin_2FARequired` | Admin endpoints require 2FA |
| `TestAdmin_ListAgents` | List all running agents |
| `TestAdmin_ControlAgent` | Pause/resume/stop agents |
| `TestAdmin_QueryTaints` | Report taints per agent |
| `TestAdmin_InjectEvent` | Inject event into agent |
| `TestAdmin_BoundaryOuter` | Admin only accessible from outer |

**Commit**: `feat: sys:admin with k9s-style terminal and 2FA`

---

## Service 3: sys:persistence (Snapshots & Event Sourcing)

**Goal**: Snapshots, event sourcing, version migration

### Architecture

```yaml
apiVersion: maelstrom.dev/v1
kind: PlatformService
metadata:
  name: sys:persistence
  core: false
  boundary: inner
spec:
  chartRef: persistence-v1
  requiredForKernelReady: false
  storage:
    type: postgres
    connection: ${DB_URL}
  snapshot:
    enabled: true
    onStateEntry: true
    everyMessages: 50
    cron: "*/15 * * * *"
```

### Components

| Component | File | Purpose | Lines |
|-----------|------|---------|-------|
| PersistenceService | `pkg/services/persistence/service.go` | Main service + storage | ~150 |
| SnapshotStore | `pkg/services/persistence/snapshot.go` | Snapshot CRUD operations | ~100 |
| EventStore | `pkg/services/persistence/events.go` | Append-only event log | ~120 |
| MigrationManager | `pkg/services/persistence/migrate.go` | Version migration logic | ~150 |
| Test Suite | `pkg/services/persistence/service_test.go` | 7 tests | ~200 |

### API Contract

```go
type PersistenceService interface {
    Snapshot(runtimeId string) (Snapshot, error)
    Restore(snapshotId string, def ChartDefinition) (RuntimeID, error)
    AppendEvent(runtimeId string, event Event) error
    GetEvents(runtimeId string, since string) ([]Event, error)
    Migrate(runtimeId string, newDef ChartDefinition) error
}
```

### Test Specifications

| Test | Semantics Covered |
|------|-------------------|
| `TestPersistence_SnapshotCreate` | Create snapshot of runtime |
| `TestPersistence_SnapshotRestore` | Restore from snapshot |
| `TestPersistence_EventAppend` | Append event to event store |
| `TestPersistence_EventQuery` | Query events since ID |
| `TestPersistence_MigrateShallow` | Migrate with shallowHistory |
| `TestPersistence_MigrateDeep` | Migrate with deepHistory |
| `TestPersistence_MigrateClean` | Migrate with cleanStart |

**Commit**: `feat: sys:persistence with snapshots, event sourcing, migration`

---

## Service 4: sys:heartbeat (Scheduled Wake-ups)

**Goal**: Scheduled agent wake-ups with HEARTBEAT.md injection

### Architecture

```yaml
apiVersion: maelstrom.dev/v1
kind: PlatformService
metadata:
  name: sys:heartbeat
  core: false
  boundary: dmz
spec:
  chartRef: heartbeat-v1
  requiredForKernelReady: false
  schedules:
    - agentId: agent-1
      cron: "*/5 * * * *"
      template: HEARTBEAT.md
    - agentId: agent-2
      cron: "0 * * * *"
      template: DAILY_SUMMARY.md
```

### Components

| Component | File | Purpose | Lines |
|-----------|------|---------|-------|
| HeartbeatService | `pkg/services/heartbeat/service.go` | Main service + scheduler | ~150 |
| CronScheduler | `pkg/services/heartbeat/scheduler.go` | Cron-based scheduling | ~100 |
| TemplateEngine | `pkg/services/heartbeat/templates.go` | HEARTBEAT.md injection | ~80 |
| Test Suite | `pkg/services/heartbeat/service_test.go` | 5 tests | ~150 |

### API Contract

```go
type HeartbeatService interface {
    Schedule(agentId string, cron string, template string) error
    Unschedule(agentId string) error
    Trigger(agentId string) error
    GetSchedule(agentId string) (Schedule, error)
}
```

### Test Specifications

| Test | Semantics Covered |
|------|-------------------|
| `TestHeartbeat_Schedule` | Schedule cron-based heartbeat |
| `TestHeartbeat_Trigger` | Trigger heartbeat manually |
| `TestHeartbeat_TemplateInjection` | Inject HEARTBEAT.md into session |
| `TestHeartbeat_CronFires` | Cron fires at scheduled time |
| `TestHeartbeat_Unschedule` | Remove schedule |

**Commit**: `feat: sys:heartbeat with cron scheduling and template injection`

---

## Service 5: sys:memory (Long-term Memory)

**Goal**: Long-term memory (vector/graph stores) for ContextMap injection

### Architecture

```yaml
apiVersion: maelstrom.dev/v1
kind: PlatformService
metadata:
  name: sys:memory
  core: false
  boundary: inner
spec:
  chartRef: memory-v1
  requiredForKernelReady: false
  store:
    type: vector
    provider: pinecone
    dimensions: 1536
  queries:
    - name: recent
      topK: 10
    - name: contextual
      topK: 5
      threshold: 0.7
```

### Components

| Component | File | Purpose | Lines |
|-----------|------|---------|-------|
| MemoryService | `pkg/services/memory/service.go` | Main service + vector store | ~150 |
| VectorStore | `pkg/services/memory/vector.go` | Vector DB operations | ~120 |
| GraphStore | `pkg/services/memory/graph.go` | Graph DB operations | ~100 |
| QueryEngine | `pkg/services/memory/query.go` | Query assembly and execution | ~100 |
| Test Suite | `pkg/services/memory/service_test.go` | 6 tests | ~180 |

### API Contract

```go
type MemoryService interface {
    Store(runtimeId string, content string, metadata map[string]any) (string, error)
    Query(vector []float32, topK int, boundaryFilter string) ([]MemoryResult, error)
    QueryByQuery(query string, topK int, boundaryFilter string) ([]MemoryResult, error)
    Delete(memoryId string) error
    List(runtimeId string) ([]MemoryResult, error)
}
```

### Test Specifications

| Test | Semantics Covered |
|------|-------------------|
| `TestMemory_Store` | Store vector embedding |
| `TestMemory_QueryVector` | Similarity search by vector |
| `TestMemory_QueryText` | Similarity search by text query |
| `TestMemory_BoundaryFilter` | Filter by boundary type |
| `TestMemory_Delete` | Delete memory entry |
| `TestMemory_List` | List all memories for agent |

**Commit**: `feat: sys:memory with vector and graph stores`

---

## Service 6: sys:tools (Tool Registry)

**Goal**: Tool registry and resolution

### Architecture

```yaml
apiVersion: maelstrom.dev/v1
kind: PlatformService
metadata:
  name: sys:tools
  core: false
  boundary: dmz
spec:
  chartRef: tools-v1
  requiredForKernelReady: false
  tools:
    - name: webSearch
      boundary: dmz
      isolation: container
    - name: innerDbQuery
      boundary: inner
      isolation: strict
```

### Components

| Component | File | Purpose | Lines |
|-----------|------|---------|-------|
| ToolsService | `pkg/services/tools/service.go` | Main service + registry | ~150 |
| ToolRegistry | `pkg/services/tools/registry.go` | Tool registration and lookup | ~100 |
| ToolInvoker | `pkg/services/tools/invoker.go` | Isolated tool execution | ~150 |
| SchemaValidator | `pkg/services/tools/schema.go` | Validate tool schemas | ~80 |
| Test Suite | `pkg/services/tools/service_test.go` | 7 tests | ~200 |

### API Contract

```go
type ToolsService interface {
    Register(tool ToolDescriptor) error
    Resolve(name string, callerBoundary string) (ToolDescriptor, error)
    List(boundaryFilter string) ([]ToolDescriptor, error)
    Invoke(toolName string, args map[string]any, callerBoundary string) (any, error)
    Unregister(name string) error
}
```

### Test Specifications

| Test | Semantics Covered |
|------|-------------------|
| `TestTools_Register` | Register tool with schema |
| `TestTools_Resolve` | Resolve tool by name |
| `TestTools_BoundaryFilter` | Filter tools by caller boundary |
| `TestTools_Invoke` | Invoke tool in isolated context |
| `TestTools_Isolation` | Tool runs in isolated namespace |
| `TestTools_Unregister` | Remove tool from registry |
| `TestTools_NotFound` | Unknown tool returns error |

**Commit**: `feat: sys:tools with registry, isolation, and invocation`

---

## Service 7: sys:datasources (Pluggable Storage)

**Goal**: Pluggable FS/object storage with tainting (wraps pkg/datasource)

### Architecture

```yaml
apiVersion: maelstrom.dev/v1
kind: PlatformService
metadata:
  name: sys:datasources
  core: false
  boundary: inner
spec:
  chartRef: datasources-v1
  requiredForKernelReady: false
  sources:
    - name: agentWorkspace
      type: localDisk
      path: /agents/{{agentId}}/workspace
      taintMode: inheritFromXattr
    - name: s3Bucket
      type: objectStorage
      bucket: my-bucket
      taintMode: inheritFromObjectTags
```

### Components

| Component | File | Purpose | Lines |
|-----------|------|---------|-------|
| DatasourceService | `pkg/services/datasources/service.go` | Main service + registry | ~150 |
| DataSourceRegistry | `pkg/services/datasources/registry.go` | Wrap pkg/datasource registry | ~80 |
| LocalDiskSource | `pkg/services/datasources/local.go` | localDisk wrapper | ~80 |
| ObjectStorageSource | `pkg/services/datasources/object.go` | S3/GCS wrapper | ~100 |
| Test Suite | `pkg/services/datasources/service_test.go` | 5 tests | ~150 |

### API Contract

```go
type DatasourceService interface {
    Get(name string) (datasource.DataSource, error)
    List() []string
    TagOnWrite(path string, taints []string) error
    GetTaints(path string) ([]string, error)
    ValidateAccess(path string, boundary string) error
}
```

### Test Specifications

| Test | Semantics Covered |
|------|-------------------|
| `TestDatasources_Register` | Register localDisk, s3 sources |
| `TestDatasources_Get` | Get datasource by name |
| `TestDatasources_TagOnWrite` | Tag file with taints |
| `TestDatasources_GetTaints` | Read taints from file |
| `TestDatasources_ValidateAccess` | Check boundary access |

**Commit**: `feat: sys:datasources wrapping pkg/datasource with tainting`

---

## Service 8: sys:human-gateway (Chat Interface)

**Goal**: Chat interface for human-in-the-loop with running agents

### Architecture

```yaml
apiVersion: maelstrom.dev/v1
kind: PlatformService
metadata:
  name: sys:human-gateway
  core: false
  boundary: outer
spec:
  chartRef: human-gateway-v1
  requiredForKernelReady: false
  endpoints:
    - path: /chat/{agentId}
      methods: [GET, POST]
      stream: true
```

### Components

| Component | File | Purpose | Lines |
|-----------|------|---------|-------|
| HumanGatewayService | `pkg/services/humangateway/service.go` | Main service + routes | ~150 |
| ChatHandler | `pkg/services/humangateway/chat.go` | Chat session management | ~150 |
| SessionStore | `pkg/services/humangateway/sessions.go` | Session persistence | ~100 |
| Test Suite | `pkg/services/humangateway/service_test.go` | 4 tests | ~120 |

### API Contract

```go
type HumanGatewayService interface {
    OpenSession(agentId string) (SessionID, error)
    SendMessage(sessionId string, content string) error
    StreamResponse(sessionId string) (<-chan StreamChunk, error)
    CloseSession(sessionId string) error
}
```

### Test Specifications

| Test | Semantics Covered |
|------|-------------------|
| `TestHumanGateway_OpenSession` | Open chat session for agent |
| `TestHumanGateway_SendMessage` | Send message to agent |
| `TestHumanGateway_StreamResponse` | Stream agent response |
| `TestHumanGateway_BoundaryOuter` | Human gateway only accessible from outer |

**Commit**: `feat: sys:human-gateway with chat sessions and streaming`

---

## File Structure

```
pkg/services/
├── gateway/
│   ├── service.go           # Main service + adapter registry
│   ├── adapter.go           # Base ChannelAdapter interface
│   ├── webhook.go           # HTTP webhook adapter
│   ├── websocket.go         # WebSocket bidirectional adapter
│   ├── sse.go               # Server-Sent Events adapter
│   ├── smtp.go              # Email/SMTP adapter
│   └── service_test.go      # 8 tests
├── admin/
│   ├── service.go           # Main service + 2FA middleware
│   ├── terminal.go          # k9s-style TUI
│   ├── auth.go              # 2FA authentication
│   └── service_test.go      # 6 tests
├── persistence/
│   ├── service.go           # Main service + storage
│   ├── snapshot.go          # Snapshot CRUD
│   ├── events.go            # Append-only event log
│   ├── migrate.go           # Version migration
│   └── service_test.go      # 7 tests
├── heartbeat/
│   ├── service.go           # Main service + scheduler
│   ├── scheduler.go         # Cron-based scheduling
│   ├── templates.go         # HEARTBEAT.md injection
│   └── service_test.go      # 5 tests
├── memory/
│   ├── service.go           # Main service + vector store
│   ├── vector.go            # Vector DB operations
│   ├── graph.go             # Graph DB operations
│   ├── query.go             # Query assembly
│   └── service_test.go      # 6 tests
├── tools/
│   ├── service.go           # Main service + registry
│   ├── registry.go          # Tool registration
│   ├── invoker.go           # Isolated execution
│   ├── schema.go            # Schema validation
│   └── service_test.go      # 7 tests
├── datasources/
│   ├── service.go           # Main service + registry
│   ├── registry.go          # Wrap pkg/datasource
│   ├── local.go             # localDisk wrapper
│   ├── object.go            # S3/GCS wrapper
│   └── service_test.go      # 5 tests
└── humangateway/
    ├── service.go           # Main service + routes
    ├── chat.go              # Chat session management
    ├── sessions.go          # Session persistence
    └── service_test.go      # 4 tests
```

---

## Total Test Count

- **sys:gateway**: 8 tests
- **sys:admin**: 6 tests
- **sys:persistence**: 7 tests
- **sys:heartbeat**: 5 tests
- **sys:memory**: 6 tests
- **sys:tools**: 7 tests
- **sys:datasources**: 5 tests
- **sys:human-gateway**: 4 tests
- **Total**: 48 new tests

**Current Total**: 169 (existing) + 48 (new) = **217 tests**

---

## Implementation Order (TDD Workflow)

### Week 1: Gateway & Admin (14 tests)

1. `TestGateway_RegisterAdapter` → Define ChannelAdapter interface
2. `TestGateway_WebhookIngress` → Implement webhook adapter
3. `TestGateway_WebSocketBidir` → Implement WS adapter
4. `TestGateway_SSEStreaming` → Implement SSE adapter
5. `TestGateway_SMTPIngress` → Implement SMTP adapter
6. `TestGateway_OpenAPI` → Auto-generate OpenAPI spec
7. `TestGateway_BoundaryOuter` → Enforce outer boundary
8. `TestGateway_AdapterNotFound` → Handle unknown adapters
9. `TestAdmin_2FARequired` → Implement 2FA middleware
10. `TestAdmin_ListAgents` → List running agents
11. `TestAdmin_ControlAgent` → Pause/resume/stop agents
12. `TestAdmin_QueryTaints` → Report taints per agent
13. `TestAdmin_InjectEvent` → Inject events into agents
14. `TestAdmin_BoundaryOuter` → Enforce outer boundary

**Commit**: `feat: sys:gateway and sys:admin services`

### Week 2: Persistence & Heartbeat (12 tests)

15. `TestPersistence_SnapshotCreate` → Snapshot CRUD
16. `TestPersistence_SnapshotRestore` → Restore from snapshot
17. `TestPersistence_EventAppend` → Append events
18. `TestPersistence_EventQuery` → Query events
19. `TestPersistence_MigrateShallow` → Shallow migration
20. `TestPersistence_MigrateDeep` → Deep migration
21. `TestPersistence_MigrateClean` → Clean migration
22. `TestHeartbeat_Schedule` → Schedule cron
23. `TestHeartbeat_Trigger` → Trigger manually
24. `TestHeartbeat_TemplateInjection` → Inject templates
25. `TestHeartbeat_CronFires` → Cron scheduling
26. `TestHeartbeat_Unschedule` → Remove schedule

**Commit**: `feat: sys:persistence and sys:heartbeat services`

### Week 3: Memory & Tools (13 tests)

27. `TestMemory_Store` → Store embeddings
28. `TestMemory_QueryVector` → Vector search
29. `TestMemory_QueryText` → Text search
30. `TestMemory_BoundaryFilter` → Filter by boundary
31. `TestMemory_Delete` → Delete entries
32. `TestMemory_List` → List memories
33. `TestTools_Register` → Register tools
34. `TestTools_Resolve` → Resolve by name
35. `TestTools_BoundaryFilter` → Filter by boundary
36. `TestTools_Invoke` → Invoke tools
37. `TestTools_Isolation` → Isolated execution
38. `TestTools_Unregister` → Remove tools
39. `TestTools_NotFound` → Handle unknown tools

**Commit**: `feat: sys:memory and sys:tools services`

### Week 4: Datasources & Human-Gateway (9 tests)

40. `TestDatasources_Register` → Register sources
41. `TestDatasources_Get` → Get by name
42. `TestDatasources_TagOnWrite` → Tag files
43. `TestDatasources_GetTaints` → Read taints
44. `TestDatasources_ValidateAccess` → Check access
45. `TestHumanGateway_OpenSession` → Open sessions
46. `TestHumanGateway_SendMessage` → Send messages
47. `TestHumanGateway_StreamResponse` → Stream responses
48. `TestHumanGateway_BoundaryOuter` → Enforce outer

**Commit**: `feat: sys:datasources and sys:human-gateway services`

---

## Key Design Decisions

### 1. Service Registration Pattern

All services register themselves in `pkg/bootstrap/bootstrap.go`:

```go
func init() {
    // Register core services
    services.Register("sys:gateway", gateway.BootstrapChart)
    services.Register("sys:admin", admin.BootstrapChart)
    services.Register("sys:persistence", persistence.BootstrapChart)
    // ... etc
}
```

**Why**: Clear dependency graph, testable, configurable.

### 2. Mail-Based Communication

All services communicate via the Mail system:

```go
// Service receives mail
func (s *Service) HandleMail(mail mail.Mail) error {
    switch mail.Type {
    case mail.MailReceived:
        // Process external ingress
    case mail.ToolResult:
        // Process tool results
    case mail.Error:
        // Handle errors
    }
    return nil
}
```

**Why**: Decoupled, testable, follows Maelstrom architecture.

### 3. Boundary Enforcement

Each service declares its boundary and enforces it:

```go
// Gateway is outer-only
if callerBoundary != outer {
    return errors.New("gateway only accessible from outer")
}
```

**Why**: Security invariant, prevents unauthorized access.

### 4. Test Count (48 Tests)

**Why 48 Tests?**
- 1:1 semantics-to-tests ratio maintained
- Each service has 4-8 focused tests
- Covers registration, mail handling, boundary, integration

**Acceptable**: Yes, matches Phase 3's scope.

---

## Risk Assessment

| Risk | Likelihood | Impact | Mitigation |
|------|-----------|--------|------------|
| Services too complex | Medium | High | Keep each service focused on single responsibility |
| Mail system bottlenecks | Low | Medium | Use async processing, background goroutines |
| Boundary enforcement gaps | Medium | High | Integration tests for cross-service communication |
| Test coverage gaps | Medium | Medium | Track semantics vs tests (1:1 ratio target) |
| Bootstrap ordering | Medium | High | Explicit registration order in bootstrap.go |

---

## Success Criteria

1. ✅ All 8 services implemented with working bootstrap charts
2. ✅ Services communicate via Mail system
3. ✅ Boundary enforcement works for all services
4. ✅ 48 new tests pass + 161 existing tests pass = 209 total
5. ✅ 1:1 semantics-to-tests ratio maintained
6. ✅ Services loadable via ChartRegistry
7. ✅ Hot-reload works for all services

---

## Questions for Review

1. **Service Scope**: Are the 8 services appropriate for Phase 4?
2. **Test Count**: 48 tests total - acceptable given semantics covered?
3. **Registration**: Should services register in bootstrap.go or dynamically?
4. **Mail Flow**: Should services process mail synchronously or asynchronously?
5. **Integration Tests**: Should they live in pkg/services/ or pkg/integration/?

---

## Deferred Features

### Phase 5: Execution Layer (Layer 6)
- Tool execution policies (seq_failfast, seq_continue, par_continue)
- Orchestrator sub-chart
- Sub-agent invocation

**Milestone**: Can execute tools and sub-agents.

### Phase 6: Agent Layer (Layer 7)
- Agent Charts (OODA loop)
- LLM bridge integration
- ContextMap assembly for LLM calls

**Milestone**: Agents can reason and use tools.

### Phase 7: External Interface (Layer 8)
- HTTP gateway with OpenAPI
- SSE/WS streaming endpoints
- 2FA for admin

**Milestone**: External clients can interact with agents.