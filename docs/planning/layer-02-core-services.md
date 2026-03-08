# Layer 2: Core Platform Services

**Reference**: `docs/arch-v1.md` Section 7.1  
**Parent Scope**: `docs/planning/implementation-scope.md`  
**Status**: ❌ PENDING

---

## Overview

Layer 2 implements the 4 hard-coded platform services that Bootstrap Chart spawns during kernel initialization. These services are **Charts** but compiled into the binary (not YAML-defined), providing foundational capabilities for all other Charts.

**Key Design Decision**: These are Charts running on the same Statechart Engine as user agents, but hard-coded (not YAML) to ensure deterministic bootstrap and prevent accidental hot-reload or tampering.

---

## Architectural Dependencies

```
Layer 2 depends on:
├── Layer 1 (Kernel & Bootstrap Infrastructure)
│   ├── Kernel spawns these services via Bootstrap Chart
│   └── ChartRegistry for YAML loading (post-kernel_ready)
└── Layer 0 (Statechart Engine) ✓ COMPLETE
    ├── Library interface for spawning runtimes
    └── Event system for service communication
```

**Why Layer 1 First**: The Kernel spawns these services through the Bootstrap Chart's entry actions. Layer 1 provides the bootstrap mechanism that brings these services online.

---

## Service Overview

| Service ID | Responsibility | Phase 1 Behavior |
|------------|---------------|------------------|
| `sys:security` | Boundary enforcement, taint validation, sanitization | Pass-through (no enforcement) |
| `sys:communication` | Mail pub/sub backbone | Basic pub/sub implementation |
| `sys:observability` | Traces, metrics, dead-letter queue | Basic trace storage |
| `sys:lifecycle` | Basic spawn/stop for bootstrap | Minimal spawn/stop |

---

## 1. sys:security

**ID**: `sys:security`  
**Boundary**: inner (hard-coded service)  
**Purpose**: Boundary enforcement, taint validation, sanitization

### Interface (arch-v1.md Section 14.2)

```go
type Security interface {
    // ValidateAndSanitize sanitizes mail crossing boundaries
    ValidateAndSanitize(mail Mail, sourceBoundary, targetBoundary BoundaryType) (Mail, error)
    
    // TaintPropagate tracks taints through copy/read/write operations
    TaintPropagate(obj any, newTaints []string) (any, error)
    
    // NamespaceIsolate provides isolated views for agents
    NamespaceIsolate(runtimeId string, operation string) (IsolatedView, error)
    
    // CheckTaintPolicy validates data against taint policy
    CheckTaintPolicy(data any, targetBoundary BoundaryType, policy TaintPolicy) (bool, error)
    
    // ReportTaints returns TaintMap (runtimeId → object→taints mapping)
    ReportTaints(runtimeId string) (TaintMap, error)
    
    // PrepareContextForBoundary called inside every assembleContextMap
    PrepareContextForBoundary(runtimeId string, boundary BoundaryType) error
}
```

### Mail Interface

```go
// HandleMail processes incoming mail
func (s *SecurityService) HandleMail(mail Mail) error {
    // Phase 1: Pass-through (no enforcement yet)
    // Phase 3: Add boundary enforcement
}
```

### Service Behavior

**Phase 1 (Bootstrap)**:
- Accept all mail without enforcement
- Register as `sys:security` for Mail addressing
- Provide basic taint tracking stub

**Phase 3 (Full Implementation)**:
- Enforce boundary transitions (inner/DMZ/outer)
- Apply taint propagation rules
- Sanitize outbound data (PII redaction, length caps)
- Call `prepareContextForBoundary` on ContextMap assembly
- Drop/redact forbidden taints at boundaries

### Service API

```go
type SecurityService struct {
    id          string  // "sys:security"
    boundaries  BoundaryConfig
    taintEngine TaintEngine  // Stub for Phase 3
}

func NewSecurityService() *SecurityService

func (s *SecurityService) ID() string

func (s *SecurityService) HandleMail(mail Mail) error

func (s *SecurityService) ValidateBoundary(caller, callee string) error

func (s *SecurityService) TagWithTaint(data any, taints []string) any
```

### Test Descriptions

**TestSecurityService_ID**
- Verify returns "sys:security"

**TestSecurityService_HandleMail**
- Accepts mail (Phase 1: no enforcement)

**TestSecurityService_StartStop**
- Lifecycle management works correctly

---

## 2. sys:communication

**ID**: `sys:communication`  
**Boundary**: inner (hard-coded service)  
**Purpose**: Mail pub/sub backbone

### Interface (arch-v1.md Section 9.3)

```go
type Communication interface {
    // Publish sends mail to inbox/topic
    Publish(mail Mail) (ack Ack, err error)
    
    // Subscribe subscribes to address
    Subscribe(address string) (mailStream chan Mail, err error)
    
    // Unsubscribe removes subscription
    Unsubscribe(address string, stream chan Mail) error
    
    // UpgradeToStream upgrades connection to SSE/WS
    UpgradeToStream(sessionId string, lastEventId *string) (chunkChannel chan StreamChunk, err error)
}
```

### Mail Addressing

Three address types:
- `agent:<id>` - specific agent inbox
- `topic:<name>` - pub-sub topic (multiple subscribers)
- `sys:<service>` - system service inbox

**Inbox Rule**: Only top-level agents (long life-cycle) have inboxes. All agents can send mail to any inbox.

### Delivery Guarantees

- **At-least-once delivery**: Guaranteed by Communication layer
- **Deduplication**: Via `correlationId`
- **Dead-letter queue**: Managed by `sys:observability` for permanent failures

### Service Behavior

**Phase 1 (Bootstrap)**:
- Basic pub/sub implementation
- Accept mail, route to targets
- Track subscriptions per address

**Phase 2 (Full Implementation)**:
- Full at-least-once delivery
- Request-reply pattern via `correlationId`
- Dead-letter queue integration
- SSE/WS streaming support

### Service API

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

func (s *CommunicationService) HandleMail(mail Mail) error

func (s *CommunicationService) Subscribe(address string, sub Subscriber) error

func (s *CommunicationService) Publish(mail Mail) error

func (s *CommunicationService) UpgradeToStream(sessionId string, lastEventId *string) (chan StreamChunk, error)
```

### Test Descriptions

**TestCommunicationService_ID**
- Verify returns "sys:communication"

**TestCommunicationService_PublishSubscribe**
- Pub/sub works correctly

**TestCommunicationService_RouteToAgent**
- Direct routing to agent inbox

**TestCommunicationService_RouteToTopic**
- Topic broadcasting to subscribers

**TestCommunicationService_StartStop**
- Lifecycle management works correctly

---

## 3. sys:observability

**ID**: `sys:observability`  
**Boundary**: inner (hard-coded service)  
**Purpose**: Traces, metrics, dead-letter queue

### Interface (arch-v1.md Section 7.3)

```go
type Observability interface {
    // EmitTrace stores a trace
    EmitTrace(trace Trace) error
    
    // QueryTraces returns traces matching filters
    QueryTraces(filters TraceFilters) ([]Trace, error)
    
    // LogDeadLetter stores failed messages
    LogDeadLetter(mail Mail, reason string) error
}
```

### Trace Structure

```go
type Trace struct {
    ID          string    // unique trace ID
    RuntimeID   string    // originating ChartRuntime
    EventType   string    // transition, entry, exit, event
    StatePath   string    // state path at event time
    Timestamp   time.Time // event time
    Payload     any       // event payload
}

type TraceFilters struct {
    RuntimeID string
    EventType string
    FromTime  time.Time
    ToTime    time.Time
}
```

### Service Behavior

**Phase 1 (Bootstrap)**:
- Basic trace storage in memory
- Accept trace emit requests
- Simple query by RuntimeID

**Phase 3 (Full Implementation)**:
- Full trace collection (transitions, entries, exits, events)
- Metrics collection (state counts, transition rates)
- Dead-letter queue management
- Integration with sys:admin for inspection

### Service API

```go
type ObservabilityService struct {
    id         string  // "sys:observability"
    traces     []Trace
    metrics    MetricsCollector
    deadLetter []Mail
    mu         sync.RWMutex
}

type MetricsCollector struct {
    StateCounts    map[string]int
    TransitionRate float64
    EventRate      float64
}

func NewObservabilityService() *ObservabilityService

func (s *ObservabilityService) ID() string

func (s *ObservabilityService) HandleMail(mail Mail) error

func (s *ObservabilityService) EmitTrace(trace Trace) error

func (s *ObservabilityService) QueryTraces(filters TraceFilters) ([]Trace, error)

func (s *ObservabilityService) LogDeadLetter(mail Mail, reason string) error

func (s *ObservabilityService) GetMetrics() MetricsCollector
```

### Test Descriptions

**TestObservabilityService_ID**
- Verify returns "sys:observability"

**TestObservabilityService_EmitTrace**
- Traces stored correctly

**TestObservabilityService_QueryTraces**
- Filtering works correctly

**TestObservabilityService_DeadLetter**
- Failed messages logged correctly

**TestObservabilityService_StartStop**
- Lifecycle management works correctly

---

## 4. sys:lifecycle

**ID**: `sys:lifecycle`  
**Boundary**: inner (hard-coded service)  
**Purpose**: Basic spawn/stop for bootstrap (full implementation in Layer 5)

### Interface

```go
type Lifecycle interface {
    // Spawn creates a new ChartRuntime
    Spawn(def statechart.ChartDefinition) (statechart.RuntimeID, error)
    
    // Stop halts a ChartRuntime
    Stop(id statechart.RuntimeID) error
    
    // List returns all active runtimes
    List() ([]RuntimeInfo, error)
}
```

### RuntimeInfo Structure

```go
type RuntimeInfo struct {
    ID           string
    DefinitionID string
    Boundary     BoundaryType
    ActiveStates []string
    IsRunning    bool
}
```

### Service Behavior

**Phase 1 (Bootstrap)**:
- Minimal spawn/stop for bootstrap services
- Accept spawn requests from Bootstrap Chart
- Track runtimes for listing

**Phase 5 (Full Implementation)**:
- Full lifecycle management (start, pause, resume, stop)
- Hot-reload via replaceDefinition
- Snapshot/restore support
- Transient sub-chart lifecycle binding

### Service API

```go
type LifecycleService struct {
    id       string  // "sys:lifecycle"
    library  statechart.Library
    registry *ChartRegistry
    runtimes map[statechart.RuntimeID]*statechart.ChartRuntime
    mu       sync.RWMutex
}

func NewLifecycleService(library statechart.Library, registry *ChartRegistry) *LifecycleService

func (s *LifecycleService) ID() string

func (s *LifecycleService) HandleMail(mail Mail) error

func (s *LifecycleService) Spawn(def statechart.ChartDefinition) (statechart.RuntimeID, error)

func (s *LifecycleService) Stop(id statechart.RuntimeID) error

func (s *LifecycleService) List() ([]RuntimeInfo, error)

func (s *LifecycleService) Control(id statechart.RuntimeID, cmd statechart.ControlCmd) error
```

### Test Descriptions

**TestLifecycleService_ID**
- Verify returns "sys:lifecycle"

**TestLifecycleService_Spawn**
- Can spawn charts correctly

**TestLifecycleService_Stop**
- Can stop charts correctly

**TestLifecycleService_List**
- Returns runtime list correctly

**TestLifecycleService_StartStop**
- Lifecycle management works correctly

---

## Service Orchestration

### Bootstrap Chart Integration

The Bootstrap Chart spawns these services via entry actions:

```yaml
states:
  security:
    type: atomic
    entryActions:
      - action: securityBootstrap
        params:
          boundaries: [inner, dmz, outer]
    transitions:
      - event: securityReady
        target: sys:bootstrap/communication

  communication:
    type: atomic
    entryActions:
      - action: communicationBootstrap
        params:
          mailBackbone: true
    transitions:
      - event: communicationReady
        target: sys:bootstrap/observability

  observability:
    type: atomic
    entryActions:
      - action: observabilityBootstrap
        params:
          tracing: true
          metrics: true
    transitions:
      - event: observabilityReady
        target: sys:bootstrap/lifecycle

  lifecycle:
    type: atomic
    entryActions:
      - action: lifecycleBootstrap
        params:
          enableSpawn: true
    transitions:
      - event: lifecycleReady
        target: sys:bootstrap/ready
        actions:
          - action: emitKernelReady
```

### Service Registration

Each service registers its `HandleMail` handler with the Communication layer:

```go
// Kernel registers services after Bootstrap Chart spawns them
func (k *Kernel) registerService(id string, svc PlatformService) {
    k.services[id] = svc
    k.registry.OnChange(func(name string, def any) {
        // Handle service updates
    })
}
```

---

## Mail Flow

### Service-to-Service Communication

```
Service A → Communication.Publish(mail) → Communication.Route() → Service B.HandleMail()
```

### Mail Types for Services

| Type | Purpose |
|------|---------|
| `heartbeat` | Scheduled service wake-up |
| `error` | Service error reporting |
| `taint_violation` | Security violations |
| `mail_received` | Mail arrival notification |

---

## File Structure

```
maelstrom/
├── pkg/services/
│   ├── security/
│   │   ├── service.go         # SecurityService (~100 lines)
│   │   └── service_test.go
│   ├── communication/
│   │   ├── service.go         # CommunicationService (~150 lines)
│   │   ├── publisher.go       # Mail publisher (~50 lines)
│   │   ├── subscriber.go      # Mail subscriber (~50 lines)
│   │   └── service_test.go
│   ├── observability/
│   │   ├── service.go         # ObservabilityService (~100 lines)
│   │   ├── trace.go           # Trace types (~50 lines)
│   │   └── service_test.go
│   └── lifecycle/
│       ├── service.go         # LifecycleService (~100 lines)
│       └── service_test.go
```

---

## TDD Implementation Plan

### Phase 2.1: SecurityService Stub

**Test**: `TestSecurityService_ID`
- Verify returns "sys:security"

**Test**: `TestSecurityService_HandleMail`
- Accepts mail (Phase 1: no enforcement)

**Test**: `TestSecurityService_StartStop`
- Lifecycle management works

**Implementation**:
- ~100 lines
- Basic pass-through behavior

### Phase 2.2: CommunicationService

**Test**: `TestCommunicationService_PublishSubscribe`
- Pub/sub works

**Test**: `TestCommunicationService_RouteToAgent`
- Direct routing

**Test**: `TestCommunicationService_RouteToTopic`
- Topic broadcasting

**Implementation**:
- ~250 lines (service + publisher + subscriber)
- Basic in-memory pub/sub

### Phase 2.3: ObservabilityService

**Test**: `TestObservabilityService_EmitTrace`
- Traces stored

**Test**: `TestObservabilityService_QueryTraces`
- Filtering works

**Test**: `TestObservabilityService_DeadLetter`
- Failed messages logged

**Implementation**:
- ~150 lines
- Basic trace storage

### Phase 2.4: LifecycleService

**Test**: `TestLifecycleService_Spawn`
- Can spawn charts

**Test**: `TestLifecycleService_Stop`
- Can stop charts

**Test**: `TestLifecycleService_List`
- Returns runtime list

**Implementation**:
- ~100 lines
- Minimal spawn/stop

### Phase 2.5: Integration

**Test**: `TestAllServicesSpawned`
- All 4 services spawned by Bootstrap Chart

**Test**: `TestServicesCommunicateViaMail`
- Services exchange mail via Communication

**Implementation**:
- Wire services to Bootstrap Chart
- Verify full bootstrap flow

---

## Dependencies & External Libraries

- `sync` - Mutex for thread-safe service state
- `github.com/maelstrom/v3/pkg/statechart` - Library interface for spawning

---

## Risk Assessment

| Risk | Likelihood | Impact | Mitigation |
|------|-----------|--------|------------|
| Service startup order bugs | Medium | Medium | Explicit sequential compound state in Bootstrap Chart |
| Mail routing deadlocks | Low | High | Comprehensive tests with timeouts |
| Service state race conditions | Low | Medium | Mutex-protected state, single goroutine |

---

## Open Questions

1. **Mail system scope**: Full implementation in Phase 2, or stub for Phase 3?
2. **Persistence**: In-memory only for Phase 2, or add disk persistence?
3. **Metrics depth**: Basic counts only, or full metrics collection?
4. **Dead-letter queue**: Simple list, or integration with persistence?

---

## Integration with Layer 1 Bootstrap

These 4 services are spawned by the Bootstrap Chart during Layer 1's kernel bootstrap sequence. Specifically:

1. **sys:security** - Spawned by `securityBootstrap` action in Bootstrap Chart's `security` state entry
2. **sys:communication** - Spawned by `communicationBootstrap` action in Bootstrap Chart's `communication` state entry
3. **sys:observability** - Spawned by `observabilityBootstrap` action in Bootstrap Chart's `observability` state entry
4. **sys:lifecycle** - Spawned by `lifecycleBootstrap` action in Bootstrap Chart's `lifecycle` state entry

Each service must implement the `PlatformService` interface to integrate with the Bootstrap Chart's entry actions:

```go
type PlatformService interface {
    ID() string
    HandleMail(mail Mail) error
    Start() error
    Stop() error
}
```

The Kernel coordinates these services through the statechart engine - each service runs as a ChartRuntime with well-known IDs for Mail addressing (`sys:security`, `sys:communication`, etc.).

---

## References

- `docs/arch-v1.md` Section 7.1 - Core Platform Services
- `docs/arch-v1.md` Section 9.3 - Communication Layer API
- `docs/arch-v1.md` Section 14.2 - Maelstrom Functional Unit Contracts
- `docs/planning/layer-01-kernel-bootstrap.md` - Layer 1 (dependency)
- `docs/planning/implementation-scope.md` - Parent scope document

---
