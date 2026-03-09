# Phase 1.5: Test Completion - Implementation Plan

**Branch**: `feat/test-completion`  
**Parent Spec**: `docs/planning/layer-01-status-report.md` Section 7  
**Status**: Ready for TDD execution

---

## Objective

Implement all 11 placeholder tests from the status report with minimal code to make them pass.

---

## 1. Placeholder Tests Summary

| # | Test Name | Location | Component | Status |
|---|-----------|----------|-----------|--------|
| 1 | `TestKernel_SpawnsAllServices` | `pkg/kernel/kernel_test.go:69` | Kernel | Placeholder |
| 2 | `TestKernel_ServicesReady` | `pkg/kernel/kernel_test.go:74` | Kernel | Placeholder |
| 3 | `TestKernel_KernelReadyEvent` | `pkg/kernel/kernel_test.go:79` | Kernel | Placeholder |
| 4 | `TestKernel_MailSystemRequired` | `pkg/kernel/kernel_test.go:84` | Kernel | Placeholder |
| 5 | `TestSecurityService_HandleMail` | `pkg/services/security/service_test.go:19` | Security | Placeholder |
| 6 | `TestCommunicationService_PubSub` | `pkg/services/communication/service_test.go:19` | Communication | Placeholder |
| 7 | `TestCommunicationService_RoutesMail` | `pkg/services/communication/service_test.go:23` | Communication | Placeholder |
| 8 | `TestObservabilityService_EmitTrace` | `pkg/services/observability/service_test.go:19` | Observability | Placeholder |
| 9 | `TestObservabilityService_BoundaryInner` | `pkg/services/observability/service_test.go:23` | Observability | Placeholder |
| 10 | `TestLifecycleService_SpawnChart` | `pkg/services/lifecycle/service_test.go:19` | Lifecycle | Placeholder |
| 11 | `TestLifecycleService_BoundaryInner` | `pkg/services/lifecycle/service_test.go:23` | Lifecycle | Placeholder |

---

## 2. Tests Grouped by Component

### 2.1 Kernel Tests (4 tests)

| Test | Verifies | Expected Failure | Minimal Implementation |
|------|----------|------------------|------------------------|
| `TestKernel_SpawnsAllServices` | All 4 services spawn during bootstrap | No services map in Kernel struct | Add `services` map to track spawned service IDs |
| `TestKernel_ServicesReady` | All services emit ready events immediately | Services don't emit ready events | Track ready state per service |
| `TestKernel_KernelReadyEvent` | KERNEL_READY emitted after all services ready | No kernel ready event tracking | Track when KERNEL_READY event is emitted |
| `TestKernel_MailSystemRequired` | Mail system exists before services spawn | No mail system in kernel | Add mail system reference to kernel |

### 2.2 Security Service Tests (1 test)

| Test | Verifies | Expected Failure | Minimal Implementation |
|------|----------|------------------|------------------------|
| `TestSecurityService_HandleMail` | Service accepts mail without enforcement | No `service.go` file | Create `SecurityService` struct with `HandleMail(mail) error` method |

### 2.3 Communication Service Tests (2 tests)

| Test | Verifies | Expected Failure | Minimal Implementation |
|------|----------|------------------|------------------------|
| `TestCommunicationService_PubSub` | Basic pub/sub works | No `service.go` file | Create `CommunicationService` with `Publish()` and `Subscribe()` methods |
| `TestCommunicationService_RoutesMail` | Mail routes to correct targets | No routing logic | Add mail routing based on address |

### 2.4 Observability Service Tests (2 tests)

| Test | Verifies | Expected Failure | Minimal Implementation |
|------|----------|------------------|------------------------|
| `TestObservabilityService_EmitTrace` | Traces stored correctly | No `service.go` file | Create `ObservabilityService` with `EmitTrace(trace) error` method |
| `TestObservabilityService_BoundaryInner` | Service has "inner" boundary | No boundary concept | Add `Boundary() BoundaryType` method returning `inner` |

### 2.5 Lifecycle Service Tests (2 tests)

| Test | Verifies | Expected Failure | Minimal Implementation |
|------|----------|------------------|------------------------|
| `TestLifecycleService_SpawnChart` | Can spawn charts correctly | No `service.go` file | Create `LifecycleService` with `Spawn(def) (RuntimeID, error)` method |
| `TestLifecycleService_BoundaryInner` | Service has "inner" boundary | No boundary concept | Add `Boundary() BoundaryType` method returning `inner` |

---

## 3. Test Ordering by Dependency (Simplest First)

### Order 1: Service Identity Tests (No dependencies)
These tests only verify service ID and boundary - no other components needed.

1. **`TestSecurityService_HandleMail`** - Creates SecurityService, verifies HandleMail exists
2. **`TestObservabilityService_BoundaryInner`** - Creates ObservabilityService, verifies boundary
3. **`TestLifecycleService_BoundaryInner`** - Creates LifecycleService, verifies boundary

### Order 2: Service Core Functionality (Dependencies: Order 1)
These tests verify core service functionality.

4. **`TestObservabilityService_EmitTrace`** - Verifies trace storage
5. **`TestLifecycleService_SpawnChart`** - Verifies chart spawning
6. **`TestCommunicationService_PubSub`** - Verifies pub/sub
7. **`TestCommunicationService_RoutesMail`** - Verifies mail routing

### Order 3: Kernel Integration Tests (Dependencies: Orders 1-2)
These tests require all services to be implemented.

8. **`TestKernel_MailSystemRequired`** - Verifies mail system exists
9. **`TestKernel_SpawnsAllServices`** - Verifies all services spawn
10. **`TestKernel_ServicesReady`** - Verifies ready events
11. **`TestKernel_KernelReadyEvent`** - Verifies KERNEL_READY event

---

## 4. Detailed Test Specifications

### 4.1 Security Service Tests

#### Test 1: `TestSecurityService_HandleMail`

**Location**: `pkg/services/security/service_test.go:19`

**What it verifies**: SecurityService accepts mail without enforcement (Phase 1 pass-through behavior)

**Expected failure**: No `service.go` file exists; only `bootstrap.go` with stub

**Minimal implementation needed**:
```go
// pkg/services/security/service.go
type Mail struct {
    // Minimal mail structure
}

type SecurityService struct {
    id string
}

func NewSecurityService() *SecurityService {
    return &SecurityService{id: "sys:security"}
}

func (s *SecurityService) HandleMail(mail Mail) error {
    // Phase 1: Pass-through (no enforcement)
    return nil
}
```

**Test implementation**:
```go
func TestSecurityService_HandleMail(t *testing.T) {
    svc := NewSecurityService()
    mail := Mail{} // minimal mail
    err := svc.HandleMail(mail)
    if err != nil {
        t.Errorf("HandleMail should return nil for Phase 1 pass-through, got: %v", err)
    }
}
```

---

### 4.2 Observability Service Tests

#### Test 2: `TestObservabilityService_BoundaryInner`

**Location**: `pkg/services/observability/service_test.go:23`

**What it verifies**: ObservabilityService has "inner" boundary type

**Expected failure**: No `service.go` file; no boundary concept

**Minimal implementation needed**:
```go
// pkg/services/observability/service.go
type BoundaryType string

const BoundaryInner BoundaryType = "inner"

type ObservabilityService struct {
    id       string
    boundary BoundaryType
}

func NewObservabilityService() *ObservabilityService {
    return &ObservabilityService{
        id:       "sys:observability",
        boundary: BoundaryInner,
    }
}

func (s *ObservabilityService) Boundary() BoundaryType {
    return s.boundary
}
```

**Test implementation**:
```go
func TestObservabilityService_BoundaryInner(t *testing.T) {
    svc := NewObservabilityService()
    if svc.Boundary() != BoundaryInner {
        t.Errorf("Expected boundary 'inner', got: %v", svc.Boundary())
    }
}
```

#### Test 3: `TestObservabilityService_EmitTrace`

**Location**: `pkg/services/observability/service_test.go:19`

**What it verifies**: Traces are stored correctly in memory

**Expected failure**: No trace storage mechanism

**Minimal implementation needed**:
```go
// pkg/services/observability/service.go (additions)
type Trace struct {
    ID        string
    RuntimeID string
    EventType string
    StatePath string
    Timestamp time.Time
    Payload   any
}

type ObservabilityService struct {
    id       string
    boundary BoundaryType
    traces   []Trace
    mu       sync.RWMutex
}

func (s *ObservabilityService) EmitTrace(trace Trace) error {
    s.mu.Lock()
    defer s.mu.Unlock()
    s.traces = append(s.traces, trace)
    return nil
}
```

**Test implementation**:
```go
func TestObservabilityService_EmitTrace(t *testing.T) {
    svc := NewObservabilityService()
    trace := Trace{
        ID:        "test-trace-1",
        RuntimeID: "test-runtime",
        EventType: "transition",
        StatePath: "root/child",
        Timestamp: time.Now(),
    }
    err := svc.EmitTrace(trace)
    if err != nil {
        t.Errorf("EmitTrace should return nil, got: %v", err)
    }
    // Verify trace was stored (need getter method or check len)
}
```

---

### 4.3 Lifecycle Service Tests

#### Test 4: `TestLifecycleService_BoundaryInner`

**Location**: `pkg/services/lifecycle/service_test.go:23`

**What it verifies**: LifecycleService has "inner" boundary type

**Expected failure**: No `service.go` file; no boundary concept

**Minimal implementation needed**: Same pattern as ObservabilityService

**Test implementation**:
```go
func TestLifecycleService_BoundaryInner(t *testing.T) {
    svc := NewLifecycleService(nil) // nil library for stub
    if svc.Boundary() != BoundaryInner {
        t.Errorf("Expected boundary 'inner', got: %v", svc.Boundary())
    }
}
```

#### Test 5: `TestLifecycleService_SpawnChart`

**Location**: `pkg/services/lifecycle/service_test.go:19`

**What it verifies**: LifecycleService can spawn charts

**Expected failure**: No spawn functionality

**Minimal implementation needed**:
```go
// pkg/services/lifecycle/service.go
type LifecycleService struct {
    id       string
    boundary BoundaryType
    library  statechart.Library
    mu       sync.RWMutex
}

func NewLifecycleService(library statechart.Library) *LifecycleService {
    return &LifecycleService{
        id:       "sys:lifecycle",
        boundary: BoundaryInner,
        library:  library,
    }
}

func (s *LifecycleService) Spawn(def statechart.ChartDefinition) (statechart.RuntimeID, error) {
    if s.library == nil {
        // Stub: return fake runtime ID
        return "fake-runtime-id", nil
    }
    return s.library.Spawn(def, nil)
}
```

**Test implementation**:
```go
func TestLifecycleService_SpawnChart(t *testing.T) {
    svc := NewLifecycleService(nil) // nil library for stub
    def := statechart.ChartDefinition{
        ID:      "test-chart",
        Version: "1.0.0",
    }
    rtID, err := svc.Spawn(def)
    if err != nil {
        t.Errorf("Spawn should return nil error for stub, got: %v", err)
    }
    if rtID == "" {
        t.Error("Spawn should return non-empty runtime ID")
    }
}
```

---

### 4.4 Communication Service Tests

#### Test 6: `TestCommunicationService_PubSub`

**Location**: `pkg/services/communication/service_test.go:19`

**What it verifies**: Basic publish/subscribe works

**Expected failure**: No pub/sub implementation

**Minimal implementation needed**:
```go
// pkg/services/communication/service.go
type CommunicationService struct {
    id          string
    boundary    BoundaryType
    subscribers map[string][]chan Mail
    mu          sync.RWMutex
}

func NewCommunicationService() *CommunicationService {
    return &CommunicationService{
        id:          "sys:communication",
        boundary:    BoundaryInner,
        subscribers: make(map[string][]chan Mail),
    }
}

func (s *CommunicationService) Publish(address string, mail Mail) error {
    s.mu.RLock()
    defer s.mu.RUnlock()
    // Send to all subscribers
    for _, ch := range s.subscribers[address] {
        select {
        case ch <- mail:
        default:
        }
    }
    return nil
}

func (s *CommunicationService) Subscribe(address string) <-chan Mail {
    ch := make(chan Mail, 10)
    s.mu.Lock()
    s.subscribers[address] = append(s.subscribers[address], ch)
    s.mu.Unlock()
    return ch
}
```

**Test implementation**:
```go
func TestCommunicationService_PubSub(t *testing.T) {
    svc := NewCommunicationService()
    ch := svc.Subscribe("test-topic")
    
    mail := Mail{Source: "test", Target: "test-topic"}
    err := svc.Publish("test-topic", mail)
    if err != nil {
        t.Errorf("Publish should return nil, got: %v", err)
    }
    
    select {
    case received := <-ch:
        if received.Source != mail.Source {
            t.Errorf("Expected source %s, got %s", mail.Source, received.Source)
        }
    case <-time.After(100 * time.Millisecond):
        t.Error("Timeout waiting for mail")
    }
}
```

#### Test 7: `TestCommunicationService_RoutesMail`

**Location**: `pkg/services/communication/service_test.go:23`

**What it verifies**: Mail routes to correct targets based on address

**Expected failure**: No routing logic

**Minimal implementation needed**: Add routing based on address prefixes (agent:, topic:, sys:)

**Test implementation**:
```go
func TestCommunicationService_RoutesMail(t *testing.T) {
    svc := NewCommunicationService()
    
    // Subscribe to different address types
    agentCh := svc.Subscribe("agent:test-agent")
    topicCh := svc.Subscribe("topic:test-topic")
    sysCh := svc.Subscribe("sys:security")
    
    // Publish to each and verify routing
    err := svc.Publish("agent:test-agent", Mail{Source: "test"})
    if err != nil {
        t.Errorf("Publish to agent failed: %v", err)
    }
    
    select {
    case <-agentCh:
        // OK
    case <-time.After(100 * time.Millisecond):
        t.Error("Timeout waiting for agent mail")
    }
}
```

---

### 4.5 Kernel Tests

#### Test 8: `TestKernel_MailSystemRequired`

**Location**: `pkg/kernel/kernel_test.go:84`

**What it verifies**: Mail system (CommunicationService) exists before services spawn

**Expected failure**: No mail system reference in Kernel

**Minimal implementation needed**:
```go
// pkg/kernel/kernel.go (additions)
type Kernel struct {
    engine          statechart.Library
    factory         *runtime.Factory
    sequence        *bootstrap.Sequence
    runtimes        map[string]*runtime.ChartRuntime
    services        map[string]string // serviceID -> runtimeID
    mailSystem      *communication.CommunicationService
    mu              sync.RWMutex
}

func New() *Kernel {
    return &Kernel{
        runtimes:   make(map[string]*runtime.ChartRuntime),
        services:   make(map[string]string),
        mailSystem: communication.NewCommunicationService(),
    }
}
```

**Test implementation**:
```go
func TestKernel_MailSystemRequired(t *testing.T) {
    kernel := New()
    // Verify mail system exists
    if kernel.MailSystem() == nil {
        t.Error("Mail system should exist in kernel")
    }
}
```

#### Test 9: `TestKernel_SpawnsAllServices`

**Location**: `pkg/kernel/kernel_test.go:69`

**What it verifies**: All 4 services spawn during bootstrap

**Expected failure**: Kernel logs service names but doesn't actually spawn them

**Minimal implementation needed**: Track spawned services in the services map

**Test implementation**:
```go
func TestKernel_SpawnsAllServices(t *testing.T) {
    kernel := New()
    
    ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
    defer cancel()
    
    go kernel.Start(ctx)
    time.Sleep(500 * time.Millisecond)
    
    // Verify all 4 services were spawned
    expectedServices := []string{"sys:security", "sys:communication", 
                                  "sys:observability", "sys:lifecycle"}
    for _, svc := range expectedServices {
        if _, ok := kernel.GetServiceRuntime(svc); !ok {
            t.Errorf("Service %s should be spawned", svc)
        }
    }
}
```

#### Test 10: `TestKernel_ServicesReady`

**Location**: `pkg/kernel/kernel_test.go:74`

**What it verifies**: All services emit ready events immediately

**Expected failure**: Services don't emit ready events

**Minimal implementation needed**: Track ready state per service

**Test implementation**:
```go
func TestKernel_ServicesReady(t *testing.T) {
    kernel := New()
    
    ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
    defer cancel()
    
    go kernel.Start(ctx)
    time.Sleep(800 * time.Millisecond)
    
    expectedServices := []string{"sys:security", "sys:communication", 
                                  "sys:observability", "sys:lifecycle"}
    for _, svc := range expectedServices {
        if !kernel.IsServiceReady(svc) {
            t.Errorf("Service %s should be ready", svc)
        }
    }
}
```

#### Test 11: `TestKernel_KernelReadyEvent`

**Location**: `pkg/kernel/kernel_test.go:79`

**What it verifies**: KERNEL_READY emitted after all services ready

**Expected failure**: No kernel ready event tracking

**Minimal implementation needed**: Track when KERNEL_READY event is emitted

**Test implementation**:
```go
func TestKernel_KernelReadyEvent(t *testing.T) {
    kernel := New()
    
    ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
    defer cancel()
    
    done := make(chan bool)
    go func() {
        kernel.Start(ctx)
        done <- true
    }()
    
    time.Sleep(1000 * time.Millisecond)
    
    if !kernel.IsKernelReady() {
        t.Error("KERNEL_READY event should be emitted after all services ready")
    }
}
```

---

## 5. Files to Create/Modify

### 5.1 New Files to Create

| File | Purpose | Lines |
|------|---------|-------|
| `pkg/services/security/service.go` | SecurityService implementation | ~30 |
| `pkg/services/communication/service.go` | CommunicationService with pub/sub | ~60 |
| `pkg/services/observability/service.go` | ObservabilityService with trace storage | ~50 |
| `pkg/services/lifecycle/service.go` | LifecycleService with spawn | ~40 |
| `pkg/services/types.go` | Shared types (Mail, BoundaryType, Trace) | ~30 |

### 5.2 Files to Modify

| File | Changes | Lines |
|------|---------|-------|
| `pkg/kernel/kernel.go` | Add services map, mail system, ready tracking | +30 |
| `pkg/services/security/service_test.go` | Implement HandleMail test | +15 |
| `pkg/services/communication/service_test.go` | Implement PubSub and RoutesMail tests | +30 |
| `pkg/services/observability/service_test.go` | Implement EmitTrace and BoundaryInner tests | +25 |
| `pkg/services/lifecycle/service_test.go` | Implement SpawnChart and BoundaryInner tests | +25 |
| `pkg/kernel/kernel_test.go` | Implement all 4 kernel tests | +50 |

---

## 6. TDD Workflow for Each Test

Following the strict TDD workflow from CLAUDE.md:

### For Each Test:

1. **Write the test** - Add test function with expected behavior
2. **Run test** - Confirm it fails (RED) - either compile error or assertion failure
3. **Write minimal implementation** - Just enough code to make test pass (GREEN)
4. **Run test** - Confirm it passes
5. **Commit** - `git commit -m "test: <test-name>"` or `feat: <implementation>`
6. **Repeat** - Move to next test

---

## 7. Summary of Minimal Implementation

### Types (shared across services)

```go
// pkg/services/types.go
type BoundaryType string
const (
    BoundaryInner BoundaryType = "inner"
    BoundaryDMZ   BoundaryType = "dmz"
    BoundaryOuter BoundaryType = "outer"
)

type Mail struct {
    Source    string
    Target    string
    Type      string
    Payload   any
    Timestamp time.Time
}

type Trace struct {
    ID        string
    RuntimeID string
    EventType string
    StatePath string
    Timestamp time.Time
    Payload   any
}
```

### Service Interface (common pattern)

```go
type PlatformService interface {
    ID() string
    Boundary() BoundaryType
    HandleMail(mail Mail) error
}
```

---

## 8. Risk Assessment

| Risk | Likelihood | Impact | Mitigation |
|------|-----------|--------|------------|
| Tests pass but don't verify actual behavior | Medium | Medium | Review test assertions carefully |
| Service implementations too minimal | Low | Low | Phase 1 is intentionally minimal |
| Kernel integration issues | Medium | Medium | Test in order, verify dependencies |

---

## 9. Estimated Effort

| Component | Tests | Implementation Lines | Time Estimate |
|-----------|-------|---------------------|---------------|
| Security Service | 1 | ~30 | 30 min |
| Observability Service | 2 | ~50 | 45 min |
| Lifecycle Service | 2 | ~40 | 45 min |
| Communication Service | 2 | ~60 | 60 min |
| Kernel Integration | 4 | ~30 | 60 min |
| **Total** | **11** | **~210** | **~4 hours** |

---

**End of Phase 1.5 Implementation Plan**