# Core Services & Mail System Implementation Plan

**Date**: 2026-03-01  
**Status**: Planning Phase  
**Based on**: arch-v1.md Section 9 (Communication), Section 12 (Bootstrap), and Section 7 (Platform Services)

---

## Executive Summary

This plan implements the **Mail System** (Layer 3) and **Core Platform Services** (Layer 2) as Charts, enabling cross-Chart communication and establishing the foundation for all subsequent layers.

**Current State**: Layer 0 (Statechart Engine) complete with 94 tests passing.

**Target**: Mail system operational, 4 core services spawned during bootstrap, `kernel_ready` event emitted.

---

## Architectural Dependencies

From arch-v1.md Section 9 and Section 12:

```
Bootstrap Sequence (Section 12.1):
1. Kernel starts (bare binary, no Charts yet)
2. Kernel spawns Bootstrap Chart (hard-coded YAML)
3. Bootstrap Chart sequentially spawns:
   - Region 1: sys:security (boundary enforcement, tainting)
   - Region 2: sys:communication (mail pub/sub backbone)
   - Region 3: sys:observability (traces, metrics, dead-letter)
   - Region 4: sys:lifecycle (spawn/stop charts)
4. Bootstrap Chart emits kernel_ready event
5. Kernel registers remaining actions/guards, loads YAML services

Key Constraint: Mail system MUST exist before services, since services use Mail to communicate (Section 9.1).
```

**Critical Design Decision**: All 4 core services are **Charts themselves** with `handleMail(mail) → outcome` interface, not standalone Go structs.

---

## Layer 3: Mail System

**Goal**: Cross-Chart async messaging backbone

**Dependencies**: None (foundation layer)

### Components

| Component | File | Purpose | Lines |
|-----------|------|---------|-------|
| Mail Types | `pkg/mail/types.go` | Mail struct, addressing, types | ~80 |
| Publisher | `pkg/mail/publisher.go` | Publish interface | ~40 |
| Subscriber | `pkg/mail/subscriber.go` | Subscribe/unsubscribe interface | ~40 |
| Router | `pkg/mail/router.go` | Route mail to targets | ~100 |

### Mail Types (arch-v1.md Section 9.2)

```go
type Mail struct {
    ID            string    // UUID
    CorrelationID string    // For request-reply
    Type          MailType  // See below
    CreatedAt     time.Time
    Source        string    // agent:<id> | sys:<service> | human | gateway | tool:<name>
    Target        string    // agent:<id> | topic:<name> | sys:<service>
    Content       any
    Metadata      MailMetadata
}

type MailType string
const (
    User            MailType = "user"
    Assistant       MailType = "assistant"
    ToolResult      MailType = "tool_result"
    ToolCall        MailType = "tool_call"
    MailReceived    MailType = "mail_received"
    Heartbeat       MailType = "heartbeat"
    Error           MailType = "error"
    HumanFeedback   MailType = "human_feedback"
    PartialAssistant MailType = "partial_assistant"
    SubagentDone    MailType = "subagent_done"
    TaintViolation  MailType = "taint_violation"
)

type MailMetadata struct {
    Tokens   int
    Model    string
    Cost     float64
    Boundary BoundaryType  // inner | dmz | outer
    Taints   []string      // e.g., ["PII", "SECRET", "USER_SUPPLIED"]
    Stream   bool
    IsFinal  bool
}
```

### Addressing Format (arch-v1.md Section 9.2)

- `agent:<id>` - Specific agent inbox
- `topic:<name>` - Pub-sub topic (multiple subscribers)
- `sys:<service>` - System service inbox (sys:security, sys:communication, etc.)

### Publisher Interface

```go
type Publisher interface {
    Publish(mail Mail) (Ack, error)
}

type Ack struct {
    CorrelationID string
    DeliveredAt   time.Time
}
```

### Subscriber Interface

```go
type Subscriber interface {
    Subscribe(address string) (<-chan Mail, error)
    Unsubscribe(address string, ch <-chan Mail) error
}
```

### Router Responsibilities

- Route mail to correct target (agent inbox, topic, sys service)
- Deduplicate via correlationId (at-least-once delivery)
- Integrate with dead-letter queue (Phase 2 deferred)

---

## Layer 2: Core Platform Services as Charts

**Goal**: Implement 4 hard-coded services that Bootstrap Chart spawns

**Dependencies**: Layer 3 (Mail system must exist)

### Services Overview

| Service | ID | Boundary | Purpose | YAML Constant |
|---------|-----|----------|---------|---------------|
| sys:security | `sys:security` | inner | Boundary enforcement, taint validation | `pkg/services/security/bootstrap.go` |
| sys:communication | `sys:communication` | inner | Mail pub/sub backbone | `pkg/services/communication/bootstrap.go` |
| sys:observability | `sys:observability` | inner | Traces, metrics, dead-letter | `pkg/services/observability/bootstrap.go` |
| sys:lifecycle | `sys:lifecycle` | inner | Spawn/stop charts | `pkg/services/lifecycle/bootstrap.go` |

### Service Contract (arch-v1.md Section 7.3)

Every `sys:*` service exposes:

```go
handleMail(mail Mail) → outcomeEvent
```

Specific additions per service:

- **sys:security**: `ValidateAndSanitize(mail, sourceBoundary, targetBoundary) → (Mail, error)`
- **sys:communication**: `Subscribe(address) → (<-chan Mail, error)`
- **sys:observability**: `EmitTrace(trace) / QueryTraces(filters) → []Trace`
- **sys:lifecycle**: `Spawn(def ChartDefinition) → RuntimeID, error`

### Service Chart Definitions (Hard-Coded YAML)

#### sys:security

```yaml
id: sys:security
version: 1.0.0
boundary: inner
nodes:
  root:
    type: compound
    children:
      initializing:
        type: atomic
        onEnter: [initBoundaryService]
        transitions:
          - event: SECURITY_INIT
            to: running
      running:
        type: parallel
        regions:
          - name: enforcement
            child: boundaryChecker
          - name: taintTracking
            child: taintEngine
        onEnter: [registerMailHandler]
```

#### sys:communication

```yaml
id: sys:communication
version: 1.0.0
boundary: inner
nodes:
  root:
    type: compound
    children:
      initializing:
        type: atomic
        onEnter: [initMailBackbone]
        transitions:
          - event: COMM_INIT
            to: running
      running:
        type: atomic
        onEnter: [startPubSub, registerMailHandler]
```

#### sys:observability

```yaml
id: sys:observability
version: 1.0.0
boundary: inner
nodes:
  root:
    type: compound
    children:
      initializing:
        type: atomic
        onEnter: [initTraces, initMetrics]
        transitions:
          - event: OBS_INIT
            to: running
      running:
        type: atomic
        onEnter: [registerMailHandler]
```

#### sys:lifecycle

```yaml
id: sys:lifecycle
version: 1.0.0
boundary: inner
nodes:
  root:
    type: compound
    children:
      initializing:
        type: atomic
        onEnter: [initChartRegistry]
        transitions:
          - event: LIFE_INIT
            to: running
      running:
        type: atomic
        onEnter: [registerMailHandler]
```

---

## Bootstrap Integration

### Updated Kernel Flow

```go
func (k *Kernel) onBootstrapStateEnter(ctx context.Context, state string) error {
    switch state {
    case "security":
        def := security.BootstrapChart()
        id, err := k.library.Spawn(def, nil)
        if err != nil { return err }
        // Emit ready immediately (no ceremony)
        k.factory.DispatchEvent(id, Event{Type: "SECURITY_READY"})
    
    case "communication":
        def := communication.BootstrapChart()
        id, err := k.library.Spawn(def, nil)
        if err != nil { return err }
        k.factory.DispatchEvent(id, Event{Type: "COMMUNICATION_READY"})
    
    case "observability":
        def := observability.BootstrapChart()
        id, err := k.library.Spawn(def, nil)
        if err != nil { return err }
        k.factory.DispatchEvent(id, Event{Type: "OBSERVABILITY_READY"})
    
    case "lifecycle":
        def := lifecycle.BootstrapChart()
        id, err := k.library.Spawn(def, nil)
        if err != nil { return err }
        k.factory.DispatchEvent(id, Event{Type: "LIFECYCLE_READY"})
    }
    return nil
}
```

---

## Test Specifications

### Mail System Tests (7 tests)

| Test | Semantics Covered |
|------|-------------------|
| `TestMail_AddressFormat` | Validates `agent:<id>`, `topic:<name>`, `sys:<service>` patterns |
| `TestMail_PublishSubscribe` | Publisher sends, subscriber receives via channel |
| `TestMail_Deduplication` | Same correlationId delivered only once |
| `TestMail_RouterRouting` | Routes to correct target based on address |
| `TestMail_MailTypes` | All 11 mail types exist and are valid |
| `TestMail_MetadataStructure` | Metadata fields present (tokens, model, cost, boundary, taints) |
| `TestMail_DeadLetterDeferred` | Dead-letter integration noted as Phase 2 |

### Core Services Tests (16 tests)

| Test | Service | Semantics Covered |
|------|---------|-------------------|
| `TestSecurityService_BootstrapChart` | security | YAML loads, ID is "sys:security" |
| `TestSecurityService_HandleMail` | security | Processes mail, returns outcome |
| `TestSecurityService_BoundaryInner` | security | Boundary is "inner" |
| `TestCommunicationService_BootstrapChart` | communication | YAML loads, ID is "sys:communication" |
| `TestCommunicationService_PubSub` | communication | Publish/subscribe works |
| `TestCommunicationService_RoutesMail` | communication | Routes mail to agents/topics |
| `TestCommunicationService_BoundaryInner` | communication | Boundary is "inner" |
| `TestObservabilityService_BootstrapChart` | observability | YAML loads, ID is "sys:observability" |
| `TestObservabilityService_EmitTrace` | observability | Traces recorded |
| `TestObservabilityService_BoundaryInner` | observability | Boundary is "inner" |
| `TestLifecycleService_BootstrapChart` | lifecycle | YAML loads, ID is "sys:lifecycle" |
| `TestLifecycleService_SpawnChart` | lifecycle | Can spawn other charts |
| `TestLifecycleService_BoundaryInner` | lifecycle | Boundary is "inner" |
| `TestSecurityService_ID` | security | Returns "sys:security" |
| `TestCommunicationService_ID` | communication | Returns "sys:communication" |
| `TestObservabilityService_ID` | observability | Returns "sys:observability" |
| `TestLifecycleService_ID` | lifecycle | Returns "sys:lifecycle" |

### Integration Tests (4 tests)

| Test | Semantics Covered |
|------|-------------------|
| `TestKernel_SpawnsAllServices` | All 4 services spawned sequentially during bootstrap |
| `TestKernel_ServicesReady` | All services emit ready events immediately |
| `TestKernel_KernelReadyEvent` | Final KERNEL_READY emitted after all services ready |
| `TestKernel_MailSystemRequired` | Mail system must exist before services spawn |

---

## Total Test Count

- **Mail System**: 7 tests
- **Core Services**: 16 tests (4 services × 4 tests each)
- **Integration**: 4 tests
- **Total**: 27 new tests

**Current Total**: 94 (existing) + 27 (new) = 121 tests

---

## Implementation Order (TDD Workflow)

### Phase 1: Mail System (Layer 3)

1. `TestMail_AddressFormat` → Implement `pkg/mail/types.go`
2. `TestMail_PublishSubscribe` → Implement `pkg/mail/publisher.go`
3. `TestMail_RouterRouting` → Implement `pkg/mail/router.go`
4. Remaining mail tests → Flesh out implementation
5. Commit: `feat: mail system (Layer 3)`

### Phase 2: Core Services (Layer 2)

1. `TestSecurityService_BootstrapChart` → Create `pkg/services/security/bootstrap.go`
2. `TestCommunicationService_BootstrapChart` → Create `pkg/services/communication/bootstrap.go`
3. `TestObservabilityService_BootstrapChart` → Create `pkg/services/observability/bootstrap.go`
4. `TestLifecycleService_BootstrapChart` → Create `pkg/services/lifecycle/bootstrap.go`
5. Service behavior tests → Implement `handleMail()` methods
6. Commit: `feat: core services as Charts (Layer 2)`

### Phase 3: Bootstrap Integration

1. `TestKernel_SpawnsAllServices` → Update `kernel.go` to spawn real services
2. `TestKernel_ServicesReady` → Wire up service ready events
3. `TestKernel_KernelReadyEvent` → Complete bootstrap sequence
4. `TestKernel_MailSystemRequired` → Add dependency check
5. Commit: `feat: bootstrap integration`

---

## File Structure

```
pkg/
├── mail/
│   ├── types.go           # Mail struct, types, metadata
│   ├── publisher.go       # Publisher interface
│   ├── subscriber.go      # Subscriber interface
│   ├── router.go          # Routing logic
│   └── mail_test.go       # 7 tests
│
├── services/
│   ├── security/
│   │   ├── bootstrap.go   # Hard-coded YAML constant
│   │   └── service_test.go # 4 tests
│   ├── communication/
│   │   ├── bootstrap.go   # Hard-coded YAML constant
│   │   └── service_test.go # 4 tests
│   ├── observability/
│   │   ├── bootstrap.go   # Hard-coded YAML constant
│   │   └── service_test.go # 4 tests
│   └── lifecycle/
│       ├── bootstrap.go   # Hard-coded YAML constant
│       └── service_test.go # 4 tests
│
├── kernel/
│   ├── kernel.go          # Updated to spawn services
│   └── kernel_test.go     # 4 new integration tests
```

---

## Deferred Features (Update to Implementation Plan)

**From maelstrom-full-stack-implementation-plan.md**:

### Phase 1: Foundation (Layers 1-2) - UPDATED

1. Kernel bootstrap sequence
2. ChartRegistry YAML loading
3. Bootstrap Chart (hard-coded)
4. **Mail System** (Layer 3) - NEW ORDER
5. 4 core services (sys:security, sys:communication, sys:observability, sys:lifecycle)

**Milestone**: `kernel_ready` event fires. System can spawn charts.

### Phase 2: Communication Backbone (Layer 3) - DELETED

**DELETED**: Mail system now implemented in Phase 1 (Foundation), not Phase 2.

### Phase 3: Security (Layer 4) - RENAMED

Now covers **Security Service Implementation** (not just boundary system):
1. Boundary service (inner/DMZ/outer enforcement)
2. Taint engine (proactive + reactive tainting)
3. DataSources with xattr (Phase 3 deferred)
4. ContextMap filtering (Phase 3 deferred)

**Milestone**: Security enforced on all mail and data access.

**Note**: DataSources with xattr and ContextMap filtering deferred to Phase 3 (Security Layer), not Phase 1.

### Phase 4: Remaining Platform Services (Layer 5) - UPDATED

1. YAML-defined services loadable
2. sys:gateway, sys:admin, **sys:persistence** (now has dead-letter), sys:heartbeat
3. sys:memory, sys:tools, sys:datasources (now has xattr), sys:human-gateway

**Milestone**: Full platform operational.

### Phase 5: Execution (Layer 6) - NO CHANGES

### Phase 6: Agents (Layer 7) - NO CHANGES

### Phase 7: External Interface (Layer 8) - NO CHANGES

---

## Risk Assessment

| Risk | Likelihood | Impact | Mitigation |
|------|-----------|--------|------------|
| Mail system too complex for Phase 1 | Low | Medium | Keep router simple, defer advanced features |
| Services don't communicate correctly | Medium | High | Extensive integration tests |
| Bootstrap deadlock | Low | High | Services emit ready immediately (no ceremony) |
| YAML constants too verbose | Medium | Low | Keep service charts minimal (3-4 states) |
| Test coverage gaps | Medium | Medium | Track semantics vs tests (1:1 ratio target) |

---

## Questions for Review

1. **Service Chart Complexity**: Are the proposed YAML definitions minimal enough, or should we simplify further?

2. **Mail System Scope**: Should we include any dead-letter queue logic now, or truly defer to Phase 2?

3. **Test Count**: 27 new tests for mail + services. Does this feel appropriate given the semantics covered?

4. **Bootstrap Ready Events**: Should services truly emit ready immediately, or add a small async delay to simulate real initialization?

5. **Service IDs**: Are the proposed IDs (`sys:security`, etc.) consistent with arch-v1.md naming conventions?

---

## Success Criteria

1. ✅ Mail system operational (7 tests pass)
2. ✅ All 4 core services spawn as Charts (16 tests pass)
3. ✅ Bootstrap sequence completes with `kernel_ready` event (4 tests pass)
4. ✅ 1:1 semantics-to-tests ratio maintained
5. ✅ All 27 new tests pass + 94 existing tests pass = 121 total
6. ✅ Documentation updated to reflect new implementation order