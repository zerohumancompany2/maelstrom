# Phase 1 Implementation Plan: Kernel & Bootstrap Infrastructure

**Date**: 2026-02-28
**Status**: Ready for Review
**Scope**: Layers 1-2 (Kernel, ChartRegistry, Bootstrap Chart, 4 Core Services)
**Estimated Effort**: ~1,200 lines (60% reuse from prior implementations)

---

## Overview

Phase 1 implements the foundation for the entire Maelstrom system: the kernel bootstrap sequence that brings the 4 hard-coded core services online and emits `kernel_ready`. This is the only code that runs outside the statechart abstraction.

**Key Constraint**: The kernel is ~50 lines and spawns exactly ONE chart directly (Bootstrap Chart). Everything after that uses the Library API through `sys:lifecycle`.

---

## Component Breakdown

### 1. Kernel (`cmd/maelstrom/kernel.go`)

**User Story**: As a system operator, I can start the Maelstrom binary and have it automatically bootstrap the core services in the correct order.

**Acceptance Criteria**:
- Kernel loads minimal config from environment variables
- Kernel instantiates the statechart Library
- Kernel registers exactly 4 bootstrap actions
- Kernel spawns Bootstrap Chart and waits for `kernel_ready`
- Kernel goes dormant after bootstrap (signal handling only)

**Public API**:
```go
// Kernel is the minimal bootstrap container (~50 lines)
type Kernel struct {
    library *statechart.Engine
    config  KernelConfig
}

// Run executes the bootstrap sequence
func (k *Kernel) Run() error

// Shutdown gracefully stops all services
func (k *Kernel) Shutdown() error
```

**Test Descriptions**:
1. `TestKernel_LoadsConfigFromEnv` - Verifies env var loading
2. `TestKernel_RegistersBootstrapActions` - All 4 actions registered before spawn
3. `TestKernel_SpawnsBootstrapChart` - Exactly one chart spawned directly
4. `TestKernel_WaitsForKernelReady` - Blocks until bootstrap complete
5. `TestKernel_GoesDormantAfterReady` - Signal handling only post-bootstrap

---

### 2. Registry Architecture (`pkg/registry/`)

**Design Decision**: Registry-per-Type with directory-partitioned Sources. Based on registry-sketch analysis: Source decoupling via channels, clone-under-lock pattern, single-threaded Service orchestration.

**Key Principle**: Source interface decouples from Registry - Registry knows nothing about file systems.

**Public API**:

#### Source Interface (`pkg/source/` - Top Level)
```go
type SourceEvent struct {
    Key       string    // Relative path, e.g., "gateway.yaml"
    Content   []byte    // Raw YAML
    Type      EventType // Created, Updated, Deleted
    Timestamp time.Time
}

type Source interface {
    Events() <-chan SourceEvent  // Receive-only: Source owns channel
    Err() error                   // After graceful shutdown
}

// FileSystemSource implementation
func NewFileSystemSource(root string, debounce time.Duration) (Source, error)
```

#### Generic Registry (`pkg/registry/registry.go`)
```go
type RegistryItem struct {
    VersionID string
    Raw       []byte  // Original YAML
    Content   any     // Hydrated object (ChartDefinition | AgentSpec | ...)
    Timestamp time.Time
}

type Registry struct {
    versions  map[string][]RegistryItem  // All versions
    latest    map[string]any             // Current hydrated
    hydrator  HydratorFunc               // Type-specific
    preHooks  []PreLoadHook
    postHooks []PostLoadHook
    mu        sync.Mutex                  // Clone-under-lock pattern
}

type HydratorFunc func(raw []byte) (any, error)
type PreLoadHook func(key string, value any) (any, error)
type PostLoadHook func(key string, value any)

func NewRegistry(hydrator HydratorFunc) *Registry
func (r *Registry) Set(key string, raw []byte) error  // Hydrates, runs hooks, stores
func (r *Registry) Get(key string) (any, error)
func (r *Registry) AddPreLoadHook(hook PreLoadHook)
func (r *Registry) AddPostLoadHook(hook PostLoadHook)
```

#### Service Orchestration (`pkg/registry/service.go`)
```go
type Service struct {
    registry *Registry
    source   Source  // Decoupled - can be file, HTTP, test mock
}

func (s *Service) Run(ctx context.Context) error  // Single goroutine event loop
```

#### ChartRegistry Integration (`pkg/chart/registry.go`)
```go
type ChartRegistry struct {
    registry *registry.Registry  // Stores ChartDefinition as any
    source   source.Source       // Watches charts/ directory
    service  *registry.Service
}

func NewChartRegistry(dir string) (*ChartRegistry, error)
func (c *ChartRegistry) Start(ctx context.Context) error
func (c *ChartRegistry) Get(name string) (statechart.ChartDefinition, error)  // Type assertion
func (c *ChartRegistry) OnChange(fn func(key string, def statechart.ChartDefinition))
```

**Acceptance Criteria**:
- Source decoupling: FileSystemSource can be replaced with HTTPSource without Registry changes
- Minimal concurrency: Clone-under-lock pattern (1 mutex), single Service.Run() goroutine
- Registry-per-type: ChartRegistry holds ChartDefinitions, future AgentRegistry holds AgentSpecs
- Load YAML with env substitution and Go template execution
- Watch directories with debounced change events
- Track versions with both raw YAML and hydrated objects

**Reusable Components**:

From maelstrom2/internal/:
- CoreRegistry version storage pattern (~150 lines)
- FileSystemWatcher fsnotify implementation (~180 lines)

From registry-sketch/:
- Source interface pattern (channel decoupling)
- Clone-under-lock concurrency pattern
- Service orchestration single-goroutine event loop

**Test Descriptions**:

#### Source Tests
1. `TestFileSystemSource_EmitsCreated` - File creation detected
2. `TestFileSystemSource_EmitsUpdated` - File changes detected
3. `TestFileSystemSource_Debounces` - Rapid changes coalesced
4. `TestFileSystemSource_GracefulShutdown` - Clean stop

#### Registry Tests (Generic)
1. `TestRegistry_SetGet` - Basic storage/retrieval
2. `TestRegistry_VersionTracking` - Multiple versions stored
3. `TestRegistry_CloneUnderLock` - Hooks execute without lock
4. `TestRegistry_PreLoadHooks` - Transform pipelines
5. `TestRegistry_PostLoadHooks` - Notifications

#### Service Tests
1. `TestService_ProcessesEvents` - Source events → Registry updates
2. `TestService_ObserverNotifications` - Changes pushed to observers
3. `TestService_ContextCancellation` - Graceful shutdown

#### Hydration Tests
1. `TestHydrateChart_SimpleYAML` - Basic unmarshaling
2. `TestHydrateChart_EnvSubstitution` - ${ENV_VAR} → value
3. `TestHydrateChart_TemplateExecution` - {{ .AppVars }} → value
4. `TestHydrateChart_InvalidYAML` - Error handling

#### Integration Tests
1. `TestChartRegistry_LoadsFromDirectory` - End-to-end loading
2. `TestChartRegistry_HotReload` - Change detected, new version loaded
3. `TestChartRegistry_TypeAssertion` - Get() returns ChartDefinition

---

### 3. Bootstrap Chart (`embedded/bootstrap.yaml`)

**User Story**: As the kernel, I can spawn a single hard-coded chart that brings the 4 core services online in sequence and signals when ready.

**Acceptance Criteria**:
- Sequential compound state (Region 1 → 2 → 3 → 4)
- Each region starts a core service via entry action
- Transitions to next region on success
- Emits `kernel_ready` event when all services online
- Handles failure in any region (logs and halts)

**Chart Definition** (hard-coded YAML compiled into binary):

```yaml
id: sys:bootstrap
version: 1.0.0
type: compound  # Sequential execution

# Region 1: Security & Boundary Service
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
      - event: securityFailed
        target: sys:bootstrap/failed

  # Region 2: Communication Layer
  communication:
    type: atomic
    entryActions:
      - action: communicationBootstrap
        params:
          mailBackbone: true
    transitions:
      - event: communicationReady
        target: sys:bootstrap/observability
      - event: communicationFailed
        target: sys:bootstrap/failed

  # Region 3: Observability + Persistence
  observability:
    type: atomic
    entryActions:
      - action: observabilityBootstrap
        params:
          tracing: true
          metrics: true
          deadLetterQueue: true
    transitions:
      - event: observabilityReady
        target: sys:bootstrap/lifecycle
      - event: observabilityFailed
        target: sys:bootstrap/failed

  # Region 4: Lifecycle + Tools Registry
  lifecycle:
    type: atomic
    entryActions:
      - action: lifecycleBootstrap
        params:
          enableSpawn: true
          enableStop: true
          toolRegistry: true
    transitions:
      - event: lifecycleReady
        target: sys:bootstrap/ready
        actions:
          - action: emitKernelReady

  # Final: Kernel Ready
  ready:
    type: final  # Emits done event
    entryActions:
      - action: logSuccess
        params:
          message: "Maelstrom bootstrap complete"

  # Failure Handler
  failed:
    type: final
    entryActions:
      - action: logFailure
      - action: panic  # Kernel exits on bootstrap failure

initialState: sys:bootstrap/security
```

**Bootstrap Actions** (registered by kernel before spawn):

```go
// securityBootstrap - brings sys:security online
func securityBootstrap(rt statechart.RuntimeContext, appCtx statechart.ApplicationContext, ev statechart.Event) error {
    // 1. Spawn sys:security as hard-coded chart
    // 2. Wait for it to signal ready
    // 3. Dispatch securityReady event
}

// communicationBootstrap - brings sys:communication online
func communicationBootstrap(rt statechart.RuntimeContext, appCtx statechart.ApplicationContext, ev statechart.Event) error {
    // 1. Spawn sys:communication as hard-coded chart
    // 2. Start mail backbone
    // 3. Dispatch communicationReady event
}

// observabilityBootstrap - brings sys:observability online
func observabilityBootstrap(rt statechart.RuntimeContext, appCtx statechart.ApplicationContext, ev statechart.Event) error {
    // 1. Spawn sys:observability as hard-coded chart
    // 2. Initialize tracing and metrics
    // 3. Dispatch observabilityReady event
}

// lifecycleBootstrap - brings sys:lifecycle online
func lifecycleBootstrap(rt statechart.RuntimeContext, appCtx statechart.ApplicationContext, ev statechart.Event) error {
    // 1. Spawn sys:lifecycle as hard-coded chart
    // 2. Register remaining actions (LLM bridge, etc.)
    // 3. Dispatch lifecycleReady event
}

// emitKernelReady - signals full Maelstrom activation
func emitKernelReady(rt statechart.RuntimeContext, appCtx statechart.ApplicationContext, ev statechart.Event) error {
    // 1. Emit kernel_ready event
    // 2. Kernel wakes up and enters steady state
}
```

**Test Descriptions**:
1. `TestBootstrapChart_SequentialExecution` - Regions start in order
2. `TestBootstrapChart_SecurityFirst` - Security region is first
3. `TestBootstrapChart_TransitionsOnReady` - Each region signals next
4. `TestBootstrapChart_EmitsKernelReady` - Final event emitted
5. `TestBootstrapChart_HandlesFailure` - Failure stops bootstrap
6. `TestBootstrapChart_HardCoded` - YAML embedded in binary, not loaded from disk

---

### 4. Core Platform Services (Hard-Coded)

**User Story**: As the Bootstrap Chart, I can spawn 4 core services that provide foundation capabilities for all other charts.

**Acceptance Criteria**:
- Each service is a hard-coded Chart (not YAML)
- Services have well-known IDs for Mail addressing
- Services expose `handleMail(mail) → outcome` interface
- Services can be spawned by Bootstrap Chart entry actions

**Public API**:

```go
// PlatformService is the interface for all sys:* services
type PlatformService interface {
    // ID returns the well-known service ID (e.g., "sys:security")
    ID() string

    // HandleMail processes incoming mail
    HandleMail(mail Mail) error

    // Start initializes the service
    Start() error

    // Stop halts the service
    Stop() error
}

// Mail addressing format
type Mail struct {
    ID            string
    Type          MailType  // user, assistant, tool_call, etc.
    Source        string    // agent:<id> | sys:<service> | human | gateway
    Target        string    // agent:<id> | topic:<name> | sys:<service>
    Content       any
    CorrelationID string
    Metadata      MailMetadata
}
```

**Service Implementations**:

#### sys:security (`pkg/services/security/service.go`)

```go
type SecurityService struct {
    id        string  // "sys:security"
    boundaries BoundaryConfig
    taintEngine TaintEngine  // Stub for Phase 3
}

func (s *SecurityService) HandleMail(mail Mail) error {
    // Phase 1: Pass-through (no enforcement yet)
    // Phase 3: Add boundary enforcement
}

func (s *SecurityService) ValidateBoundary(caller, callee string) error
func (s *SecurityService) TagWithTaint(data any, taints []string) any
```

**Test Descriptions**:
1. `TestSecurityService_ID` - Returns "sys:security"
2. `TestSecurityService_HandleMail` - Accepts mail (Phase 1: no enforcement)
3. `TestSecurityService_StartStop` - Lifecycle management

#### sys:communication (`pkg/services/communication/service.go`)

```go
type CommunicationService struct {
    id          string  // "sys:communication"
    publisher   Publisher
    subscribers map[string][]Subscriber
    mu          sync.RWMutex
}

func (s *CommunicationService) HandleMail(mail Mail) error {
    // Route mail to target (agent, topic, or sys service)
}

func (s *CommunicationService) Subscribe(target string, sub Subscriber) error
func (s *CommunicationService) Publish(mail Mail) error
```

**Test Descriptions**:
1. `TestCommunicationService_ID` - Returns "sys:communication"
2. `TestCommunicationService_PublishSubscribe` - Pub/sub works
3. `TestCommunicationService_RouteToAgent` - Direct routing
4. `TestCommunicationService_RouteToTopic` - Topic broadcasting
5. `TestCommunicationService_StartStop` - Lifecycle management

#### sys:observability (`pkg/services/observability/service.go`)

```go
type ObservabilityService struct {
    id        string  // "sys:observability"
    traces    []Trace
    metrics   MetricsCollector
    deadLetter []Mail  // Failed messages
    mu        sync.RWMutex
}

func (s *ObservabilityService) HandleMail(mail Mail) error {
    // Process observability commands
}

func (s *ObservabilityService) EmitTrace(trace Trace)
func (s *ObservabilityService) QueryTraces(filters TraceFilters) []Trace
func (s *ObservabilityService) LogDeadLetter(mail Mail, reason string)
```

**Test Descriptions**:
1. `TestObservabilityService_ID` - Returns "sys:observability"
2. `TestObservabilityService_EmitTrace` - Traces stored
3. `TestObservabilityService_QueryTraces` - Filtering works
4. `TestObservabilityService_DeadLetter` - Failed messages logged
5. `TestObservabilityService_StartStop` - Lifecycle management

#### sys:lifecycle (`pkg/services/lifecycle/service.go`)

```go
type LifecycleService struct {
    id       string  // "sys:lifecycle"
    library  *statechart.Engine
    registry *ChartRegistry
}

func (s *LifecycleService) HandleMail(mail Mail) error {
    // Process spawn/stop commands
}

func (s *LifecycleService) Spawn(def statechart.ChartDefinition) (statechart.RuntimeID, error)
func (s *LifecycleService) Stop(id statechart.RuntimeID) error
func (s *LifecycleService) List() []RuntimeInfo
```

**Test Descriptions**:
1. `TestLifecycleService_ID` - Returns "sys:lifecycle"
2. `TestLifecycleService_Spawn` - Can spawn charts
3. `TestLifecycleService_Stop` - Can stop charts
4. `TestLifecycleService_List` - Returns runtime list
5. `TestLifecycleService_StartStop` - Lifecycle management

---

## File Structure

```
cmd/maelstrom/
├── kernel.go              # ~50 lines, bootstrap sequence
├── main.go                # Entry point
└── kernel_test.go         # Kernel tests

pkg/registry/
├── registry.go            # ChartRegistry orchestration (~300 lines)
├── core.go                # CoreRegistry (reused from maelstrom2)
├── watcher.go             # FileSystemWatcher (reused from maelstrom2)
├── hydrator.go            # ChartHydrator (~150 lines new)
├── types.go               # Registry types and interfaces
└── registry_test.go       # Registry tests

pkg/services/
├── security/
│   ├── service.go         # SecurityService
│   └── service_test.go
├── communication/
│   ├── service.go         # CommunicationService
│   ├── publisher.go       # Mail publisher
│   ├── subscriber.go      # Mail subscriber
│   └── service_test.go
├── observability/
│   ├── service.go         # ObservabilityService
│   ├── trace.go           # Trace types
│   └── service_test.go
└── lifecycle/
    ├── service.go         # LifecycleService
    └── service_test.go

embedded/
└── bootstrap.yaml         # Hard-coded bootstrap chart
```

---

## Implementation Order (TDD Workflow per CLAUDE.md)

### Sprint 1: Kernel & Bootstrap Chart
1. `git checkout -b feat/kernel-bootstrap`
2. Stub kernel types and methods
3. Test: `TestKernel_SpawnsBootstrapChart`
4. Implement kernel bootstrap sequence
5. Test: `TestBootstrapChart_SequentialExecution`
6. Implement Bootstrap Chart YAML and actions
7. Commit: `feat: kernel bootstrap sequence`

### Sprint 2: ChartRegistry Core
1. Copy CoreRegistry from maelstrom2/internal/registry/core/
2. Test: `TestChartRegistry_LoadsYAML`
3. Implement ChartHydrator with env substitution
4. Test: `TestChartRegistry_SubstitutesEnvVars`
5. Commit: `feat: chart registry core`

### Sprint 3: File Watching
1. Copy FileSystemWatcher from maelstrom2/internal/ingestion/filesystem/
2. Test: `TestChartRegistry_WatchesDirectory`
3. Integrate watcher with registry
4. Test: `TestChartRegistry_EmitsEvents`
5. Commit: `feat: file watching and hot-reload events`

### Sprint 4: Core Services (1 per service)
1. SecurityService stub + tests
2. CommunicationService stub + tests
3. ObservabilityService stub + tests
4. LifecycleService stub + tests
5. Commit per service: `feat: sys:security`, etc.

### Sprint 5: Integration
1. End-to-end bootstrap test
2. Full bootstrap sequence verification
3. Commit: `feat: phase 1 integration`

---

## Dependencies & External Libraries

- `github.com/fsnotify/fsnotify` - File watching (already used in maelstrom2)
- `gopkg.in/yaml.v3` - YAML parsing (already in statechart)
- `text/template` - Go templates for hydration (stdlib)

---

## Risk Assessment

| Risk | Likelihood | Impact | Mitigation |
|------|-----------|--------|------------|
| Prior code doesn't fit v4 architecture | Low | Medium | Copy adapters only, not full implementations |
| Bootstrap sequence deadlocks | Low | High | Comprehensive tests with timeouts |
| File watching too resource-intensive | Low | Low | Debouncing already implemented |
| Service startup order bugs | Medium | Medium | Explicit sequential compound state |

---

## Questions for Review

1. Should we implement full Mail system in Phase 1, or stub it for Phase 2?
2. Do we need persistence for CoreRegistry in Phase 1, or in-memory only?
3. Should Bootstrap Chart failure panic (current plan) or retry?
4. Any services to defer to Phase 2 (e.g., observability metrics)?