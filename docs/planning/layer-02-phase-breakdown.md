# Layer 2: Core Platform Services - Phase Breakdown

## Executive Summary

Layer 2 implements the 4 hard-coded platform services spawned by the Bootstrap Chart. Based on analysis of Layer 1 patterns, Layer 2 specification, and current codebase state, this document breaks down Layer 2 into **6 phases** with **~45 tests** across **~45 commits**.

### Current State (Layer 1 Complete)
- ✅ All 4 services have stub implementations
- ✅ All 41 existing tests pass
- ✅ Bootstrap integration complete for Phase 1
- ⚠️ Critical types and methods missing from spec

### Layer 2 Goal
Complete Phase 1 behavior per spec:
1. **sys:security** - Pass-through with basic boundary validation
2. **sys:communication** - Basic pub/sub with Ack support
3. **sys:observability** - Trace storage with filtering and dead-letter
4. **sys:lifecycle** - Basic spawn/stop with runtime listing

---

## Phase Overview

| Phase | Name | Tests | Files | Branch | Dependencies |
|-------|------|-------|-------|--------|--------------|
| 2.1 | Type Definitions | 6 | 3 | `feat/layer2-type-definitions` | None |
| 2.2 | Communication Service Enhancement | 8 | 3 | `feat/layer2-communication` | 2.1 |
| 2.3 | Observability Service Enhancement | 9 | 2 | `feat/layer2-observability` | 2.1 |
| 2.4 | Lifecycle Service Enhancement | 7 | 1 | `feat/layer2-lifecycle` | 2.1 |
| 2.5 | Security Service Enhancement | 10 | 1 | `feat/layer2-security` | 2.1 |
| 2.6 | Service Registry & Integration | 5 | 2 | `feat/layer2-registry` | 2.2, 2.3, 2.4, 2.5 |

**Total: 45 tests, 12 files, 6 phases**

---

## Phase 2.1: Type Definitions

### Goal
Define missing types required by all services before implementing functionality.

### Scope
- Move `StreamChunk` from `humangateway` to `mail` package
- Add `TraceFilters` type for observability queries
- Add `MetricsCollector` type for observability metrics
- Add `RuntimeInfo` type for lifecycle listing
- Add `TaintMap` type for security reporting

### Current Status

#### Implementation Status
| Type | Status | Notes |
|------|--------|-------|
| `StreamChunk` | ⚠️ Exists in wrong location | In `pkg/services/humangateway/chat.go` |
| `TraceFilters` | ❌ Missing | Needed for `QueryTraces()` |
| `MetricsCollector` | ❌ Missing | Needed for `GetMetrics()` |
| `RuntimeInfo` | ❌ Missing | Needed for `List()` |
| `TaintMap` | ⚠️ Exists in `pkg/security/types.go` | But not in `pkg/services/types.go` |

#### Files Status
| File | Status |
|------|--------|
| `pkg/mail/types.go` | ⚠️ Partial - add StreamChunk |
| `pkg/services/types.go` | ⚠️ Partial - add TraceFilters, MetricsCollector |
| `pkg/services/lifecycle/types.go` | ❌ MISSING - create with RuntimeInfo |

### Required Implementation

#### StreamChunk (move to mail package)
```go
// pkg/mail/types.go
type StreamChunk struct {
    Data     string
    Sequence int
    IsFinal  bool
    Taints   []string
}
```

#### TraceFilters
```go
// pkg/services/types.go
type TraceFilters struct {
    RuntimeID string
    EventType string
    FromTime  time.Time
    ToTime    time.Time
}
```

#### MetricsCollector
```go
// pkg/services/types.go
type MetricsCollector struct {
    StateCounts      map[string]int
    TransitionRate   float64
    EventRate        float64
    LastUpdate       time.Time
}
```

#### RuntimeInfo
```go
// pkg/services/lifecycle/types.go
type RuntimeInfo struct {
    ID           string
    DefinitionID string
    Boundary     mail.BoundaryType
    ActiveStates []string
    IsRunning    bool
}
```

### Tests to Write (6 tests, 6 commits)

#### Test 1: TestMail_StreamChunkHasRequiredFields
```go
func TestMail_StreamChunkHasRequiredFields(t *testing.T)
```
**Acceptance Criteria:**
- StreamChunk type exists in `pkg/mail/types.go`
- Has Data, Sequence, IsFinal, Taints fields
- Can be instantiated with all fields

#### Test 2: TestServices_TraceFiltersHasRequiredFields
```go
func TestServices_TraceFiltersHasRequiredFields(t *testing.T)
```
**Acceptance Criteria:**
- TraceFilters type exists in `pkg/services/types.go`
- Has RuntimeID, EventType, FromTime, ToTime fields
- Can be instantiated with all fields

#### Test 3: TestServices_MetricsCollectorHasRequiredFields
```go
func TestServices_MetricsCollectorHasRequiredFields(t *testing.T)
```
**Acceptance Criteria:**
- MetricsCollector type exists in `pkg/services/types.go`
- Has StateCounts, TransitionRate, EventRate, LastUpdate fields
- Can be instantiated with all fields

#### Test 4: TestLifecycle_RuntimeInfoHasRequiredFields
```go
func TestLifecycle_RuntimeInfoHasRequiredFields(t *testing.T)
```
**Acceptance Criteria:**
- RuntimeInfo type exists in `pkg/services/lifecycle/types.go`
- Has ID, DefinitionID, Boundary, ActiveStates, IsRunning fields
- Can be instantiated with all fields

#### Test 5: TestMail_StreamChunkCanBeUsedInMetadata
```go
func TestMail_StreamChunkCanBeUsedInMetadata(t *testing.T)
```
**Acceptance Criteria:**
- StreamChunk can be embedded in MailMetadata
- Stream field in MailMetadata accepts StreamChunk

#### Test 6: TestServices_TypesAreExported
```go
func TestServices_TypesAreExported(t *testing.T)
```
**Acceptance Criteria:**
- All new types are exported (capitalized)
- Types can be imported from other packages

### Dependencies

#### Test Dependencies
```
Test 1 → Test 5 (StreamChunk structure)
Test 2 → Test 6 (TraceFilters structure)
Test 3 → Test 6 (MetricsCollector structure)
Test 4 → Test 6 (RuntimeInfo structure)
```

### Files to Create/Modify

| File | Operation | Changes |
|------|-----------|---------|
| `pkg/mail/types.go` | MODIFY | Add StreamChunk type |
| `pkg/services/types.go` | MODIFY | Add TraceFilters, MetricsCollector |
| `pkg/services/lifecycle/types.go` | CREATE | Add RuntimeInfo type |

### Implementation Order (TDD)

1. **Test 1**: Write test → verify RED → Add StreamChunk to mail/types.go → verify GREEN → commit
2. **Test 2**: Write test → verify RED → Add TraceFilters to services/types.go → verify GREEN → commit
3. **Test 3**: Write test → verify RED → Add MetricsCollector to services/types.go → verify GREEN → commit
4. **Test 4**: Write test → verify RED → Create lifecycle/types.go with RuntimeInfo → verify GREEN → commit
5. **Test 5**: Write test → verify RED → Update MailMetadata.Stream field → verify GREEN → commit
6. **Test 6**: Write test → verify RED → Verify all types exported → verify GREEN → commit

**Total: 6 tests, 6 commits**

### Deliverables
- ✅ `StreamChunk` type in `pkg/mail/types.go`
- ✅ `TraceFilters` type in `pkg/services/types.go`
- ✅ `MetricsCollector` type in `pkg/services/types.go`
- ✅ `RuntimeInfo` type in `pkg/services/lifecycle/types.go`
- ✅ 6 commits (one per test per TDD workflow)

---

## Phase 2.2: Communication Service Enhancement

### Goal
Enhance Communication Service with Ack support, unsubscribe, and proper pub/sub interfaces.

### Scope
- Update `Publish()` to return `Ack`
- Implement `Unsubscribe()` method
- Create `publisher.go` with Publisher interface
- Create `subscriber.go` with Subscriber interface
- Add streaming support stub

### Current Status

#### Implementation Status
| Test | Status | Notes |
|------|--------|-------|
| TestCommunicationService_PublishReturnsAck | ❌ Missing | Publish returns error only |
| TestCommunicationService_Unsubscribe | ❌ Missing | Method doesn't exist |
| TestCommunicationService_UpgradeToStream | ❌ Missing | Method doesn't exist |
| TestCommunication_PublisherInterface | ❌ Missing | publisher.go missing |
| TestCommunication_SubscriberInterface | ❌ Missing | subscriber.go missing |

#### Files Status
| File | Status |
|------|--------|
| `pkg/services/communication/publisher.go` | ❌ MISSING - must create |
| `pkg/services/communication/subscriber.go` | ❌ MISSING - must create |
| `pkg/services/communication/service.go` | ⚠️ Partial - update Publish() signature |

### Required Implementation

#### Publisher Interface
```go
// pkg/services/communication/publisher.go
type Publisher interface {
    Publish(mail mail.Mail) (ack mail.Ack, err error)
}
```

#### Subscriber Interface
```go
// pkg/services/communication/subscriber.go
type Subscriber interface {
    Subscribe(address string) (<-chan mail.Mail, error)
    Unsubscribe(address string, stream chan mail.Mail) error
}
```

#### Updated Publish Method
```go
// pkg/services/communication/service.go
func (c *CommunicationService) Publish(mail mail.Mail) (ack mail.Ack, err error) {
    // Return Ack with delivery confirmation
}
```

### Tests to Write (8 tests, 8 commits)

#### Test 1: TestCommunicationService_PublishReturnsAck
```go
func TestCommunicationService_PublishReturnsAck(t *testing.T)
```
**Acceptance Criteria:**
- Publish() returns (mail.Ack, error)
- Ack has MailID matching published mail
- Ack.Success is true when delivered

#### Test 2: TestCommunicationService_PublishAckHasCorrelationID
```go
func TestCommunicationService_PublishAckHasCorrelationID(t *testing.T)
```
**Acceptance Criteria:**
- Ack.CorrelationID matches mail.CorrelationID
- Ack.DeliveredAt is set to current time

#### Test 3: TestCommunicationService_UnsubscribeRemovesSubscriber
```go
func TestCommunicationService_UnsubscribeRemovesSubscriber(t *testing.T)
```
**Acceptance Criteria:**
- Unsubscribe() removes stream from address
- After unsubscribe, no more mail received
- Returns nil error on success

#### Test 4: TestCommunicationService_UnsubscribeNotFoundReturnsError
```go
func TestCommunicationService_UnsubscribeNotFoundReturnsError(t *testing.T)
```
**Acceptance Criteria:**
- Unsubscribe() with non-existent stream returns error
- Error message indicates stream not found

#### Test 5: TestCommunication_PublisherInterfaceImplemented
```go
func TestCommunication_PublisherInterfaceImplemented(t *testing.T)
```
**Acceptance Criteria:**
- CommunicationService implements Publisher interface
- Publisher interface defined in publisher.go

#### Test 6: TestCommunication_SubscriberInterfaceImplemented
```go
func TestCommunication_SubscriberInterfaceImplemented(t *testing.T)
```
**Acceptance Criteria:**
- CommunicationService implements Subscriber interface
- Subscriber interface defined in subscriber.go

#### Test 7: TestCommunicationService_UpgradeToStreamReturnsChannel
```go
func TestCommunicationService_UpgradeToStreamReturnsChannel(t *testing.T)
```
**Acceptance Criteria:**
- UpgradeToStream() returns channel of StreamChunk
- Channel is non-nil

#### Test 8: TestCommunicationService_PublishToNonExistentAddress
```go
func TestCommunicationService_PublishToNonExistentAddress(t *testing.T)
```
**Acceptance Criteria:**
- Publish() to address with no subscribers succeeds
- Ack.Success is false (no delivery)
- Ack.ErrorMessage indicates no subscribers

### Dependencies

#### Test Dependencies
```
Test 1 → Test 2 (Ack structure)
Test 5 (Publisher interface)
Test 3 → Test 4 (Unsubscribe behavior)
Test 6 (Subscriber interface)
Test 7 (UpgradeToStream)
Test 8 (Publish edge case)
```

### Files to Create/Modify

| File | Operation | Changes |
|------|-----------|---------|
| `pkg/services/communication/publisher.go` | CREATE | Publisher interface |
| `pkg/services/communication/subscriber.go` | CREATE | Subscriber interface |
| `pkg/services/communication/service.go` | MODIFY | Update Publish(), add Unsubscribe(), UpgradeToStream() |

### Implementation Order (TDD)

1. **Test 1**: Write test → verify RED → Update Publish() to return Ack → verify GREEN → commit
2. **Test 2**: Write test → verify RED → Populate Ack fields → verify GREEN → commit
3. **Test 3**: Write test → verify RED → Implement Unsubscribe() → verify GREEN → commit
4. **Test 4**: Write test → verify RED → Add error handling to Unsubscribe() → verify GREEN → commit
5. **Test 5**: Write test → verify RED → Create publisher.go with interface → verify GREEN → commit
6. **Test 6**: Write test → verify RED → Create subscriber.go with interface → verify GREEN → commit
7. **Test 7**: Write test → verify RED → Implement UpgradeToStream() stub → verify GREEN → commit
8. **Test 8**: Write test → verify RED → Handle no-subscriber case in Publish() → verify GREEN → commit

**Total: 8 tests, 8 commits**

### Deliverables
- ✅ `Publisher` interface in `publisher.go`
- ✅ `Subscriber` interface in `subscriber.go`
- ✅ `Publish()` returns Ack
- ✅ `Unsubscribe()` implemented
- ✅ `UpgradeToStream()` stub
- ✅ 8 commits (one per test per TDD workflow)

---

## Phase 2.3: Observability Service Enhancement

### Goal
Enhance Observability Service with trace filtering, dead-letter queue, and metrics collection.

### Scope
- Update `QueryTraces()` to use `TraceFilters`
- Implement `LogDeadLetter()` method
- Implement `GetMetrics()` method
- Add dead-letter storage

### Current Status

#### Implementation Status
| Test | Status | Notes |
|------|--------|-------|
| TestObservabilityService_QueryTracesWithFilters | ❌ Missing | Uses string instead of TraceFilters |
| TestObservabilityService_LogDeadLetter | ❌ Missing | Method doesn't exist |
| TestObservabilityService_GetMetrics | ❌ Missing | Method doesn't exist |
| TestObservabilityService_QueryTracesByEventType | ❌ Missing | No event type filtering |
| TestObservabilityService_QueryTracesByTimeRange | ❌ Missing | No time range filtering |

#### Files Status
| File | Status |
|------|--------|
| `pkg/services/observability/trace.go` | ❌ MISSING - must create |
| `pkg/services/observability/service.go` | ⚠️ Partial - update QueryTraces() signature |

### Required Implementation

#### DeadLetterEntry
```go
// pkg/services/observability/trace.go
type DeadLetterEntry struct {
    Mail   mail.Mail
    Reason string
    Logged time.Time
}
```

#### Updated QueryTraces Method
```go
// pkg/services/observability/service.go
func (o *ObservabilityService) QueryTraces(filters TraceFilters) ([]Trace, error) {
    // Filter by RuntimeID, EventType, FromTime, ToTime
}
```

#### LogDeadLetter Method
```go
// pkg/services/observability/service.go
func (o *ObservabilityService) LogDeadLetter(mail mail.Mail, reason string) error {
    // Store in dead-letter queue
}
```

#### GetMetrics Method
```go
// pkg/services/observability/service.go
func (o *ObservabilityService) GetMetrics() MetricsCollector {
    // Return current metrics
}
```

### Tests to Write (9 tests, 9 commits)

#### Test 1: TestObservabilityService_QueryTracesWithFilters
```go
func TestObservabilityService_QueryTracesWithFilters(t *testing.T)
```
**Acceptance Criteria:**
- QueryTraces() accepts TraceFilters
- Filters by RuntimeID correctly

#### Test 2: TestObservabilityService_QueryTracesByEventType
```go
func TestObservabilityService_QueryTracesByEventType(t *testing.T)
```
**Acceptance Criteria:**
- QueryTraces() filters by EventType
- Empty EventType returns all events

#### Test 3: TestObservabilityService_QueryTracesByTimeRange
```go
func TestObservabilityService_QueryTracesByTimeRange(t *testing.T)
```
**Acceptance Criteria:**
- QueryTraces() filters by FromTime and ToTime
- Returns traces within time range

#### Test 4: TestObservabilityService_LogDeadLetter
```go
func TestObservabilityService_LogDeadLetter(t *testing.T)
```
**Acceptance Criteria:**
- LogDeadLetter() stores mail and reason
- Logged timestamp is set

#### Test 5: TestObservabilityService_QueryDeadLetters
```go
func TestObservabilityService_QueryDeadLetters(t *testing.T)
```
**Acceptance Criteria:**
- Can query dead-letter entries
- Returns all logged entries

#### Test 6: TestObservabilityService_GetMetrics
```go
func TestObservabilityService_GetMetrics(t *testing.T)
```
**Acceptance Criteria:**
- GetMetrics() returns MetricsCollector
- MetricsCollector has StateCounts map

#### Test 7: TestObservabilityService_MetricsUpdateOnTrace
```go
func TestObservabilityService_MetricsUpdateOnTrace(t *testing.T)
```
**Acceptance Criteria:**
- EmitTrace() updates metrics
- EventRate increases after emit

#### Test 8: TestObservabilityService_DeadLetterWithReason
```go
func TestObservabilityService_DeadLetterWithReason(t *testing.T)
```
**Acceptance Criteria:**
- DeadLetterEntry preserves reason
- Reason is non-empty string

#### Test 9: TestObservabilityService_EmptyFiltersReturnsAll
```go
func TestObservabilityService_EmptyFiltersReturnsAll(t *testing.T)
```
**Acceptance Criteria:**
- Empty TraceFilters returns all traces
- No filtering applied

### Dependencies

#### Test Dependencies
```
Test 1 → Test 2, Test 3, Test 9 (QueryTraces with filters)
Test 4 → Test 5, Test 8 (Dead-letter functionality)
Test 6 → Test 7 (Metrics functionality)
```

### Files to Create/Modify

| File | Operation | Changes |
|------|-----------|---------|
| `pkg/services/observability/trace.go` | CREATE | DeadLetterEntry type |
| `pkg/services/observability/service.go` | MODIFY | Update QueryTraces(), add LogDeadLetter(), GetMetrics() |

### Implementation Order (TDD)

1. **Test 1**: Write test → verify RED → Update QueryTraces() to use TraceFilters → verify GREEN → commit
2. **Test 2**: Write test → verify RED → Add EventType filtering → verify GREEN → commit
3. **Test 3**: Write test → verify RED → Add time range filtering → verify GREEN → commit
4. **Test 4**: Write test → verify RED → Implement LogDeadLetter() → verify GREEN → commit
5. **Test 5**: Write test → verify RED → Add QueryDeadLetters() → verify GREEN → commit
6. **Test 6**: Write test → verify RED → Implement GetMetrics() stub → verify GREEN → commit
7. **Test 7**: Write test → verify RED → Update metrics on EmitTrace() → verify GREEN → commit
8. **Test 8**: Write test → verify RED → Ensure reason preserved in dead-letter → verify GREEN → commit
9. **Test 9**: Write test → verify RED → Handle empty filters → verify GREEN → commit

**Total: 9 tests, 9 commits**

### Deliverables
- ✅ `DeadLetterEntry` type in `trace.go`
- ✅ `QueryTraces()` uses TraceFilters
- ✅ `LogDeadLetter()` implemented
- ✅ `GetMetrics()` implemented
- ✅ 9 commits (one per test per TDD workflow)

---

## Phase 2.4: Lifecycle Service Enhancement

### Goal
Enhance Lifecycle Service with proper runtime listing and control methods.

### Scope
- Update `List()` to return `[]RuntimeInfo`
- Implement `Control()` method
- Track spawned runtimes

### Current Status

#### Implementation Status
| Test | Status | Notes |
|------|--------|-------|
| TestLifecycleService_ListReturnsRuntimeInfo | ❌ Missing | Returns []string instead of []RuntimeInfo |
| TestLifecycleService_ControlStart | ❌ Missing | Method doesn't exist |
| TestLifecycleService_ControlStop | ❌ Missing | Method doesn't exist |
| TestLifecycleService_SpawnTracksRuntime | ❌ Missing | Spawn doesn't track |

#### Files Status
| File | Status |
|------|--------|
| `pkg/services/lifecycle/service.go` | ⚠️ Partial - update List() signature |

### Required Implementation

#### Updated List Method
```go
// pkg/services/lifecycle/service.go
func (l *LifecycleService) List() ([]RuntimeInfo, error) {
    // Return list of runtime info
}
```

#### Control Method
```go
// pkg/services/lifecycle/service.go
func (l *LifecycleService) Control(id statechart.RuntimeID, cmd statechart.ControlCmd) error {
    // Forward control command to engine
}
```

### Tests to Write (7 tests, 7 commits)

#### Test 1: TestLifecycleService_ListReturnsRuntimeInfo
```go
func TestLifecycleService_ListReturnsRuntimeInfo(t *testing.T)
```
**Acceptance Criteria:**
- List() returns []RuntimeInfo
- Each RuntimeInfo has ID, DefinitionID, Boundary

#### Test 2: TestLifecycleService_ListEmptyWhenNoRuntimes
```go
func TestLifecycleService_ListEmptyWhenNoRuntimes(t *testing.T)
```
**Acceptance Criteria:**
- List() returns empty slice when no runtimes
- No error returned

#### Test 3: TestLifecycleService_SpawnTracksRuntime
```go
func TestLifecycleService_SpawnTracksRuntime(t *testing.T)
```
**Acceptance Criteria:**
- Spawn() adds runtime to internal tracking
- List() includes spawned runtime

#### Test 4: TestLifecycleService_ControlStart
```go
func TestLifecycleService_ControlStart(t *testing.T)
```
**Acceptance Criteria:**
- Control() with CmdStart succeeds
- Returns nil error

#### Test 5: TestLifecycleService_ControlStop
```go
func TestLifecycleService_ControlStop(t *testing.T)
```
**Acceptance Criteria:**
- Control() with CmdStop succeeds
- Returns nil error

#### Test 6: TestLifecycleService_ControlNotFoundReturnsError
```go
func TestLifecycleService_ControlNotFoundReturnsError(t *testing.T)
```
**Acceptance Criteria:**
- Control() with non-existent ID returns error
- Error indicates runtime not found

#### Test 7: TestLifecycleService_RuntimeInfoHasActiveStates
```go
func TestLifecycleService_RuntimeInfoHasActiveStates(t *testing.T)
```
**Acceptance Criteria:**
- RuntimeInfo.ActiveStates is populated
- RuntimeInfo.IsRunning is true for active runtimes

### Dependencies

#### Test Dependencies
```
Test 1 → Test 2 (List behavior)
Test 3 (Spawn tracking)
Test 4 → Test 5, Test 6 (Control behavior)
Test 7 (RuntimeInfo details)
```

### Files to Create/Modify

| File | Operation | Changes |
|------|-----------|---------|
| `pkg/services/lifecycle/service.go` | MODIFY | Update List(), add Control(), add runtime tracking |

### Implementation Order (TDD)

1. **Test 1**: Write test → verify RED → Update List() to return []RuntimeInfo → verify GREEN → commit
2. **Test 2**: Write test → verify RED → Handle empty runtime list → verify GREEN → commit
3. **Test 3**: Write test → verify RED → Add runtime tracking in Spawn() → verify GREEN → commit
4. **Test 4**: Write test → verify RED → Implement Control() with CmdStart → verify GREEN → commit
5. **Test 5**: Write test → verify RED → Implement Control() with CmdStop → verify GREEN → commit
6. **Test 6**: Write test → verify RED → Add error handling for missing runtime → verify GREEN → commit
7. **Test 7**: Write test → verify RED → Populate ActiveStates and IsRunning → verify GREEN → commit

**Total: 7 tests, 7 commits**

### Deliverables
- ✅ `List()` returns []RuntimeInfo
- ✅ `Control()` implemented
- ✅ Runtime tracking in Spawn()
- ✅ 7 commits (one per test per TDD workflow)

---

## Phase 2.5: Security Service Enhancement

### Goal
Enhance Security Service with taint propagation and boundary validation.

### Scope
- Implement `ValidateAndSanitize()` pass-through
- Implement `TaintPropagate()` stub
- Implement `ReportTaints()` stub
- Implement `PrepareContextForBoundary()` stub

### Current Status

#### Implementation Status
| Test | Status | Notes |
|------|--------|-------|
| TestSecurityService_ValidateAndSanitize | ❌ Missing | Method doesn't exist |
| TestSecurityService_TaintPropagate | ❌ Missing | Method doesn't exist |
| TestSecurityService_ReportTaints | ❌ Missing | Method doesn't exist |
| TestSecurityService_PrepareContextForBoundary | ❌ Missing | Method doesn't exist |

#### Files Status
| File | Status |
|------|--------|
| `pkg/services/security/service.go` | ⚠️ Partial - add new methods |

### Required Implementation

#### ValidateAndSanitize Method
```go
// pkg/services/security/service.go
func (s *SecurityService) ValidateAndSanitize(mail mail.Mail, sourceBoundary, targetBoundary mail.BoundaryType) (mail.Mail, error) {
    // Phase 1: pass-through
    return mail, nil
}
```

#### TaintPropagate Method
```go
// pkg/services/security/service.go
func (s *SecurityService) TaintPropagate(obj any, newTaints []string) (any, error) {
    // Phase 1: return as-is
    return obj, nil
}
```

#### ReportTaints Method
```go
// pkg/services/security/service.go
func (s *SecurityService) ReportTaints(runtimeId string) (security.TaintMap, error) {
    // Phase 1: return empty
    return security.TaintMap{}, nil
}
```

#### PrepareContextForBoundary Method
```go
// pkg/services/security/service.go
func (s *SecurityService) PrepareContextForBoundary(runtimeId string, boundary mail.BoundaryType) error {
    // Phase 1: no-op
    return nil
}
```

### Tests to Write (10 tests, 10 commits)

#### Test 1: TestSecurityService_ValidateAndSanitizePassThrough
```go
func TestSecurityService_ValidateAndSanitizePassThrough(t *testing.T)
```
**Acceptance Criteria:**
- ValidateAndSanitize() returns mail unchanged
- Returns nil error

#### Test 2: TestSecurityService_ValidateAndSanitizeWithBoundaries
```go
func TestSecurityService_ValidateAndSanitizeWithBoundaries(t *testing.T)
```
**Acceptance Criteria:**
- ValidateAndSanitize() accepts boundary types
- Phase 1: no enforcement

#### Test 3: TestSecurityService_TaintPropagateReturnsObject
```go
func TestSecurityService_TaintPropagateReturnsObject(t *testing.T)
```
**Acceptance Criteria:**
- TaintPropagate() returns object unchanged
- Returns nil error

#### Test 4: TestSecurityService_TaintPropagateWithString
```go
func TestSecurityService_TaintPropagateWithString(t *testing.T)
```
**Acceptance Criteria:**
- TaintPropagate() works with string type
- Returns nil error

#### Test 5: TestSecurityService_ReportTaintsReturnsEmptyMap
```go
func TestSecurityService_ReportTaintsReturnsEmptyMap(t *testing.T)
```
**Acceptance Criteria:**
- ReportTaints() returns TaintMap
- TaintMap is empty (Phase 1)

#### Test 6: TestSecurityService_ReportTaintsWithRuntimeID
```go
func TestSecurityService_ReportTaintsWithRuntimeID(t *testing.T)
```
**Acceptance Criteria:**
- ReportTaints() accepts runtimeId
- Returns nil error

#### Test 7: TestSecurityService_PrepareContextForBoundaryNoOp
```go
func TestSecurityService_PrepareContextForBoundaryNoOp(t *testing.T)
```
**Acceptance Criteria:**
- PrepareContextForBoundary() returns nil
- Phase 1: no-op

#### Test 8: TestSecurityService_PrepareContextForBoundaryWithInnerBoundary
```go
func TestSecurityService_PrepareContextForBoundaryWithInnerBoundary(t *testing.T)
```
**Acceptance Criteria:**
- Works with InnerBoundary type
- Returns nil error

#### Test 9: TestSecurityService_PrepareContextForBoundaryWithOuterBoundary
```go
func TestSecurityService_PrepareContextForBoundaryWithOuterBoundary(t *testing.T)
```
**Acceptance Criteria:**
- Works with OuterBoundary type
- Returns nil error

#### Test 10: TestSecurityService_AllMethodsCoexist
```go
func TestSecurityService_AllMethodsCoexist(t *testing.T)
```
**Acceptance Criteria:**
- All new methods exist on SecurityService
- No conflicts with existing methods

### Dependencies

#### Test Dependencies
```
Test 1 → Test 2 (ValidateAndSanitize)
Test 3 → Test 4 (TaintPropagate)
Test 5 → Test 6 (ReportTaints)
Test 7 → Test 8, Test 9 (PrepareContextForBoundary)
Test 10 (All methods coexist)
```

### Files to Create/Modify

| File | Operation | Changes |
|------|-----------|---------|
| `pkg/services/security/service.go` | MODIFY | Add ValidateAndSanitize(), TaintPropagate(), ReportTaints(), PrepareContextForBoundary() |

### Implementation Order (TDD)

1. **Test 1**: Write test → verify RED → Implement ValidateAndSanitize() pass-through → verify GREEN → commit
2. **Test 2**: Write test → verify RED → Accept boundary parameters → verify GREEN → commit
3. **Test 3**: Write test → verify RED → Implement TaintPropagate() stub → verify GREEN → commit
4. **Test 4**: Write test → verify RED → Handle string type in TaintPropagate() → verify GREEN → commit
5. **Test 5**: Write test → verify RED → Implement ReportTaints() stub → verify GREEN → commit
6. **Test 6**: Write test → verify RED → Accept runtimeId parameter → verify GREEN → commit
7. **Test 7**: Write test → verify RED → Implement PrepareContextForBoundary() stub → verify GREEN → commit
8. **Test 8**: Write test → verify RED → Handle InnerBoundary → verify GREEN → commit
9. **Test 9**: Write test → verify RED → Handle OuterBoundary → verify GREEN → commit
10. **Test 10**: Write test → verify RED → Verify all methods coexist → verify GREEN → commit

**Total: 10 tests, 10 commits**

### Deliverables
- ✅ `ValidateAndSanitize()` implemented (pass-through)
- ✅ `TaintPropagate()` implemented (stub)
- ✅ `ReportTaints()` implemented (stub)
- ✅ `PrepareContextForBoundary()` implemented (stub)
- ✅ 10 commits (one per test per TDD workflow)

---

## Phase 2.6: Service Registry & Integration

### Goal
Implement Service Registry and verify all services integrate correctly.

### Scope
- Implement `Register()` method
- Implement `Get()` method
- Implement `List()` method
- Add integration tests for all services

### Current Status

#### Implementation Status
| Test | Status | Notes |
|------|--------|-------|
| TestServiceRegistry_Register | ❌ Missing | Method is TODO |
| TestServiceRegistry_Get | ❌ Missing | Method is TODO |
| TestServiceRegistry_List | ❌ Missing | Method is TODO |
| TestAllServicesIntegrateViaRegistry | ❌ Missing | Integration test |

#### Files Status
| File | Status |
|------|--------|
| `pkg/services/registry.go` | ⚠️ Partial - all methods are TODO |
| `pkg/services/registry_test.go` | ❌ MISSING - must create |

### Required Implementation

#### Register Method
```go
// pkg/services/registry.go
func (r *ServiceRegistry) Register(name string, svc Service) error {
    // Register service by name
}
```

#### Get Method
```go
// pkg/services/registry.go
func (r *ServiceRegistry) Get(name string) (Service, bool) {
    // Get service by name
}
```

#### List Method
```go
// pkg/services/registry.go
func (r *ServiceRegistry) List() []string {
    // List all service names
}
```

### Tests to Write (5 tests, 5 commits)

#### Test 1: TestServiceRegistry_Register
```go
func TestServiceRegistry_Register(t *testing.T)
```
**Acceptance Criteria:**
- Register() stores service by name
- Returns nil error

#### Test 2: TestServiceRegistry_Get
```go
func TestServiceRegistry_Get(t *testing.T)
```
**Acceptance Criteria:**
- Get() returns registered service
- Returns true for found service

#### Test 3: TestServiceRegistry_List
```go
func TestServiceRegistry_List(t *testing.T)
```
**Acceptance Criteria:**
- List() returns all registered service names
- Order is deterministic

#### Test 4: TestServiceRegistry_GetNotFound
```go
func TestServiceRegistry_GetNotFound(t *testing.T)
```
**Acceptance Criteria:**
- Get() with non-existent name returns (nil, false)
- No error returned

#### Test 5: TestAllServicesIntegrateViaRegistry
```go
func TestAllServicesIntegrateViaRegistry(t *testing.T)
```
**Acceptance Criteria:**
- All 4 services can be registered
- All 4 services can be retrieved
- List() returns all 4 service names

### Dependencies

#### Test Dependencies
```
Test 1 → Test 2, Test 3, Test 4 (Registry operations)
Test 5 (Integration - all services)
```

### Files to Create/Modify

| File | Operation | Changes |
|------|-----------|---------|
| `pkg/services/registry.go` | MODIFY | Implement Register(), Get(), List() |
| `pkg/services/registry_test.go` | CREATE | Add 5 tests |

### Implementation Order (TDD)

1. **Test 1**: Write test → verify RED → Implement Register() → verify GREEN → commit
2. **Test 2**: Write test → verify RED → Implement Get() → verify GREEN → commit
3. **Test 3**: Write test → verify RED → Implement List() → verify GREEN → commit
4. **Test 4**: Write test → verify RED → Handle not found in Get() → verify GREEN → commit
5. **Test 5**: Write test → verify RED → Register all 4 services → verify GREEN → commit

**Total: 5 tests, 5 commits**

### Deliverables
- ✅ `Register()` implemented
- ✅ `Get()` implemented
- ✅ `List()` implemented
- ✅ All 4 services integrate via registry
- ✅ 5 commits (one per test per TDD workflow)

---

## Summary

### Phase Summary Table

| Phase | Tests | Files | Branch | Dependencies | Doc |
|-------|-------|-------|--------|--------------|-----|
| 2.1 | 6 | 3 | `feat/layer2-type-definitions` | None | [phase-2.1-type-definitions.md](phase-2.1-type-definitions.md) |
| 2.2 | 8 | 3 | `feat/layer2-communication` | 2.1 | [phase-2.2-communication-service.md](phase-2.2-communication-service.md) |
| 2.3 | 9 | 2 | `feat/layer2-observability` | 2.1 | [phase-2.3-observability-service.md](phase-2.3-observability-service.md) |
| 2.4 | 7 | 1 | `feat/layer2-lifecycle` | 2.1 | [phase-2.4-lifecycle-service.md](phase-2.4-lifecycle-service.md) |
| 2.5 | 10 | 1 | `feat/layer2-security` | 2.1 | [phase-2.5-security-service.md](phase-2.5-security-service.md) |
| 2.6 | 5 | 2 | `feat/layer2-registry` | 2.2, 2.3, 2.4, 2.5 | [phase-2.6-registry-integration.md](phase-2.6-registry-integration.md) |

**Total: 45 tests, 12 files, 6 phases**

### Individual Phase Documents

Each phase has a detailed implementation plan:
- [Phase 2.1: Type Definitions](phase-2.1-type-definitions.md)
- [Phase 2.2: Communication Service Enhancement](phase-2.2-communication-service.md)
- [Phase 2.3: Observability Service Enhancement](phase-2.3-observability-service.md)
- [Phase 2.4: Lifecycle Service Enhancement](phase-2.4-lifecycle-service.md)
- [Phase 2.5: Security Service Enhancement](phase-2.5-security-service.md)
- [Phase 2.6: Service Registry & Integration](phase-2.6-registry-integration.md)

### Execution Order

```
Phase 2.1 (Type Definitions)
    ↓
┌─────────────────────────────────────────┐
│     Phases 2.2-2.5 (can run in parallel)│
│  ┌──────────┬─────────────┬──────────┐  │
│  │ 2.2 Comm │ 2.3 Obs     │ 2.4 Life │  │
│  │ 2.5 Sec  │             │          │  │
│  └──────────┴─────────────┴──────────┘  │
└─────────────────────────────────────────┘
    ↓
Phase 2.6 (Registry & Integration)
```

### Files to Create

| File | Phase | Purpose |
|------|-------|---------|
| `pkg/mail/types.go` | 2.1 | Add StreamChunk |
| `pkg/services/types.go` | 2.1 | Add TraceFilters, MetricsCollector |
| `pkg/services/lifecycle/types.go` | 2.1 | Add RuntimeInfo |
| `pkg/services/communication/publisher.go` | 2.2 | Publisher interface |
| `pkg/services/communication/subscriber.go` | 2.2 | Subscriber interface |
| `pkg/services/observability/trace.go` | 2.3 | DeadLetterEntry |
| `pkg/services/registry_test.go` | 2.6 | Registry tests |

### Files to Modify

| File | Phases | Changes |
|------|--------|---------|
| `pkg/services/communication/service.go` | 2.2 | Update Publish(), add Unsubscribe(), UpgradeToStream() |
| `pkg/services/observability/service.go` | 2.3 | Update QueryTraces(), add LogDeadLetter(), GetMetrics() |
| `pkg/services/lifecycle/service.go` | 2.4 | Update List(), add Control() |
| `pkg/services/security/service.go` | 2.5 | Add ValidateAndSanitize(), TaintPropagate(), ReportTaints(), PrepareContextForBoundary() |
| `pkg/services/registry.go` | 2.6 | Implement Register(), Get(), List() |

### Next Steps

1. **Start Phase 2.1**: Create branch `feat/layer2-type-definitions`
2. **Follow TDD workflow**: Write test → verify RED → Implement → verify GREEN → commit
3. **Repeat for each phase**: One branch per phase, merge after completion
4. **Run full test suite** after Phase 2.6 to verify integration

---

## Appendix A: Layer 1 Patterns Applied

This phase breakdown follows Layer 1 patterns:
- ✅ **1:1 test-to-commit ratio** (45 tests = 45 commits)
- ✅ **One branch per phase** (6 branches total)
- ✅ **Descriptive test names** (`Test[Component]_[Behavior]_[ExpectedResult]`)
- ✅ **Small, focused phases** (5-10 tests per phase)
- ✅ **Explicit dependencies** between phases
- ✅ **Clear acceptance criteria** for each test

## Appendix B: Gap Analysis Summary

| Gap | Resolution | Phase |
|-----|------------|-------|
| `StreamChunk` in wrong location | Move to `pkg/mail/types.go` | 2.1 |
| `TraceFilters` missing | Add to `pkg/services/types.go` | 2.1 |
| `MetricsCollector` missing | Add to `pkg/services/types.go` | 2.1 |
| `RuntimeInfo` missing | Add to `pkg/services/lifecycle/types.go` | 2.1 |
| `Publish()` returns error only | Update to return Ack | 2.2 |
| `Unsubscribe()` missing | Implement | 2.2 |
| `QueryTraces()` uses string | Update to use TraceFilters | 2.3 |
| `LogDeadLetter()` missing | Implement | 2.3 |
| `List()` returns []string | Update to return []RuntimeInfo | 2.4 |
| `Control()` missing | Implement | 2.4 |
| `ValidateAndSanitize()` missing | Implement (pass-through) | 2.5 |
| `TaintPropagate()` missing | Implement (stub) | 2.5 |
| Service Registry methods stubs | Implement all | 2.6 |