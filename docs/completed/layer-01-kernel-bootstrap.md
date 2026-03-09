# Layer 1: Kernel & Bootstrap Infrastructure

**Reference**: `docs/arch-v1.md` Section 12  
**Parent Scope**: `docs/planning/implementation-scope.md`  
**Status**: ❌ PENDING

---

## Overview

Layer 1 implements the minimal hard-coded kernel that bootstraps the entire Maelstrom system. The kernel is approximately 50 lines and is the **only non-Chart code** in the system. It spawns exactly one chart directly (the Bootstrap Chart), after which all services are managed through the statechart engine.

**Key Constraint**: The kernel manually drives the Bootstrap Chart through initialization, then hands off to Mail-based coordination once `kernel_ready` is emitted.

---

## Architectural Dependencies

```
Layer 1 depends on:
└── Layer 0 (Statechart Engine) ✓ COMPLETE
    ├── Library interface (Spawn, Control, Dispatch, RegisterAction, RegisterGuard)
    ├── ChartDefinition and ChartRuntime types
    └── Event system with FIFO ordering
```

**Why Layer 0 First**: The kernel needs the statechart engine to spawn and drive the Bootstrap Chart. Layer 0 provides the pure library API that the kernel uses.

---

## Component Breakdown

### 1. Kernel (`cmd/maelstrom/kernel.go`)

**Purpose**: Bare binary bootstrap container (~50 lines) that orchestrates initial service startup.

**Responsibilities**:

1. Load minimal config from environment variables
2. Instantiate pure Statechart Library + Chart Definition Loader
3. Register bootstrap actions/guards needed for initialization
4. Spawn Bootstrap Chart (hard-coded YAML, compiled into binary)
5. Drive Bootstrap Chart through sequential states
6. Wait for `kernel_ready` event
7. Register remaining Maelstrom actions/guards
8. Go dormant (listen only for shutdown signals)

**Public API**:

```go
// KernelConfig holds minimal bootstrap configuration
type KernelConfig struct {
    ChartsDir string  // Path to charts/ directory
    AppVars   map[string]string // Application variables for hydration
}

// Kernel is the minimal bootstrap container
type Kernel struct {
    library    statechart.Library
    config     KernelConfig
    registry   *ChartRegistry // Layer 1 component
    services   map[string]statechart.RuntimeID // Service runtimes
}

// Run executes the bootstrap sequence
func (k *Kernel) Run() error

// Shutdown gracefully stops all services
func (k *Kernel) Shutdown() error
```

**Bootstrap Sequence** (arch-v1.md 12.1):

1. **Kernel starts** (bare binary, no Charts yet)
   - Loads minimal config (env + appVars)
   - Instantiates pure Statechart Library
   - Registers 4 bootstrap actions: `sysInit`, `securityBootstrap`, `communicationBootstrap`, `observabilityBootstrap`

2. **Kernel spawns Bootstrap Chart** (single hard-coded YAML)
   - Sequential compound state with 4 atomic states:
     - State: Security & Boundary Service
     - State: Communication Layer
     - State: Observability + Persistence
     - State: Lifecycle + Tools Registry
   - Kernel manually starts this one Chart (only time it touches Library directly)

3. **Bootstrap Chart runs** (inside Library)
   - Each state's entry action spawns a core service
   - Once all ready, emits `kernel_ready` event

4. **Full Maelstrom activation**
   - Kernel registers remaining actions/guards (LLM bridge, ContextMap, Orchestrator, etc.)
   - Loads YAML-defined services via ChartRegistry
   - All future Charts spawned via `sys:lifecycle`

5. **Steady state**
   - Kernel goes dormant (only listens for shutdown signals)
   - Everything is 100% Charts + Mail + Events

---

### 2. ChartRegistry (`pkg/registry/registry.go`)

**Purpose**: YAML loading, hydration, and file watching service.

**Architecture Decision**: Registry-per-Type with directory-partitioned Sources. Source interface decouples from Registry - Registry knows nothing about file systems.

**Public API**:

```go
// RegistryItem holds a versioned chart definition
type RegistryItem struct {
    VersionID string    // semver identifier
    Raw       []byte    // Original YAML
    Content   any       // Hydrated ChartDefinition
    Timestamp time.Time // Last modified
}

// Registry is a generic versioned store for YAML definitions
type Registry struct {
    versions  map[string][]RegistryItem  // All versions keyed by name
    latest    map[string]any             // Current hydrated content
    hydrator  HydratorFunc               // Type-specific hydration
    observers map[string][]func(string, any) // Change callbacks
    mu        sync.Mutex                  // Clone-under-lock pattern
}

// HydratorFunc applies env substitution and template execution
type HydratorFunc func(raw []byte) (any, error)

// NewRegistry creates a new versioned registry
func NewRegistry(hydrator HydratorFunc) *Registry

// Set stores a new version of a definition
func (r *Registry) Set(name string, raw []byte) error

// Get returns the latest hydrated definition
func (r *Registry) Get(name string) (any, error)

// GetVersion returns a specific version
func (r *Registry) GetVersion(name, versionID string) (any, error)

// ListVersions returns all versions for a name
func (r *Registry) ListVersions(name string) []RegistryItem

// OnChange registers a callback for updates
func (r *Registry) OnChange(fn func(name string, def any))

// Clone returns a thread-safe copy for processing
func (r *Registry) Clone() map[string]any
```

**Hydration Process**:

1. Load YAML from disk
2. Substitute `${ENV_VAR}` references with environment variables
3. Execute Go templates (`{{ .AppVars.key }}`) with app variables
4. Unmarshal into ChartDefinition struct
5. Store raw and hydrated versions

**File Watching**:

- Uses `fsnotify` to detect file changes
- Debounces rapid changes to avoid event storms
- Emits SourceEvent (Created, Updated, Deleted) via channel
- Single Service.Run() goroutine processes events

**Source Interface** (decoupled from Registry):

```go
// SourceEvent represents a file system change
type SourceEvent struct {
    Key       string    // Relative path, e.g., "gateway.yaml"
    Content   []byte    // Raw YAML content
    Type      EventType // Created, Updated, Deleted
    Timestamp time.Time
}

// EventType defines the kind of change
type EventType int
const (
    Created EventType = iota
    Updated
    Deleted
)

// Source emits file changes via channel
type Source interface {
    Events() <-chan SourceEvent  // Receive-only channel
    Err() error                   // Error after shutdown
}

// FileSystemSource watches a directory for changes
func NewFileSystemSource(root string, debounce time.Duration) (Source, error)
```

**Service Orchestration**:

```go
// Service manages source events and registry updates
type Service struct {
    registry *Registry
    source   Source  // Decoupled - can be file, HTTP, test mock
}

// Run starts the event loop (single goroutine)
func (s *Service) Run(ctx context.Context) error

// Stop gracefully shuts down the service
func (s *Service) Stop() error
```

---

### 3. Bootstrap Chart (`embedded/bootstrap.yaml`)

**Purpose**: Hard-coded sequential compound state that brings 4 core services online in order.

**Structure**: Sequential compound state with 4 atomic states (executed in sequence, not parallel):

```yaml
id: sys:bootstrap
version: 1.0.0
type: compound  # Sequential execution

initialState: sys:bootstrap/security

states:
  # State: Security & Boundary Service
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

  # State: Communication Layer
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

  # State: Observability + Persistence
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

  # State: Lifecycle + Tools Registry
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
```

**Bootstrap Actions** (registered by kernel before spawn):

```go
// securityBootstrap - brings sys:security online
func securityBootstrap(rt statechart.RuntimeContext, appCtx statechart.ApplicationContext, ev statechart.Event) error {
    // 1. Spawn sys:security as hard-coded chart
    // 2. Wait for it to signal ready via event
    // 3. Dispatch securityReady event to transition
}

// communicationBootstrap - brings sys:communication online
func communicationBootstrap(rt statechart.RuntimeContext, appCtx statechart.ApplicationContext, ev statechart.Event) error {
    // 1. Spawn sys:communication as hard-coded chart
    // 2. Start mail backbone pub/sub
    // 3. Dispatch communicationReady event
}

// observabilityBootstrap - brings sys:observability online
func observabilityBootstrap(rt statechart.RuntimeContext, appCtx statechart.ApplicationContext, ev statechart.Event) error {
    // 1. Spawn sys:observability as hard-coded chart
    // 2. Initialize tracing and metrics collection
    // 3. Set up dead-letter queue
    // 4. Dispatch observabilityReady event
}

// lifecycleBootstrap - brings sys:lifecycle online
func lifecycleBootstrap(rt statechart.RuntimeContext, appCtx statechart.ApplicationContext, ev statechart.Event) error {
    // 1. Spawn sys:lifecycle as hard-coded chart
    // 2. Register remaining actions (LLM bridge, ContextMap, etc.)
    // 3. Dispatch lifecycleReady event
}

// emitKernelReady - signals full Maelstrom activation
func emitKernelReady(rt statechart.RuntimeContext, appCtx statechart.ApplicationContext, ev statechart.Event) error {
    // 1. Emit kernel_ready event (triggers kernel to wake)
    // 2. Kernel registers remaining Maelstrom actions/guards
    // 3. Kernel loads YAML-defined services
    // 4. Kernel goes dormant
}
```

---

### 4. Core Platform Services (Deferred to Layer 2)

The Bootstrap Chart spawns 4 hard-coded services during kernel initialization. Full implementation is documented in `docs/planning/layer-02-core-services.md`.

**Services**:

- `sys:security` - Boundary enforcement, taint validation (Phase 1: pass-through)
- `sys:communication` - Mail pub/sub backbone
- `sys:observability` - Traces, metrics, dead-letter queue
- `sys:lifecycle` - Basic spawn/stop for bootstrap

**Integration**: Layer 1's Bootstrap Chart triggers these services via entry actions (`securityBootstrap`, `communicationBootstrap`, `observabilityBootstrap`, `lifecycleBootstrap`). Each service must implement the `PlatformService` interface for integration.

See `docs/planning/layer-02-core-services.md` for complete API specifications and test descriptions.

---

## Hard-Coded vs YAML-Defined (arch-v1.md 12.2)

| Category | How Defined | Loaded By | Why |
|----------|-------------|-----------|-----|
| **Core / Foundational** (Security & Boundary, Communication, Observability, minimal Lifecycle/Registry, Bootstrap orchestrator) | Hard-coded Go structs (or embedded YAML compiled into binary) | Kernel at startup (never from disk) | Zero risk of missing/corrupt files, deterministic bootstrap, cannot be accidentally hot-reloaded or tampered with |
| **Everything else** (Gateway, Admin, Memory, Heartbeat, Human-Gateway, all Agent charts, Orchestrator templates, etc.) | Normal YAML on disk | ChartRegistry (hot-reloadable) | Full flexibility for users and higher-level services |

**Key Insight**: Only the absolute minimum needed to reach `kernel_ready` is hard-coded. Once core services are up, Bootstrap Chart emits `kernel_ready` and Registry takes over loading everything else.

---

## Hot-Reload & Quiescence (arch-v1.md 12.3)

### Quiescence Definition

A ChartRuntime is quiescent when:

1. Event queue is empty (no pending events)
2. No active parallel regions are processing events
3. No inflight tool calls or sub-agent invocations (Orchestrator idle)

### Hot-Reload Protocol (load-on-next-start)

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

### History Mechanisms

- **shallowHistory**: Restore to parent state's default sub-state
- **deepHistory**: Restore to specific sub-state (if still exists)
- **Deleted state fallback**: If deepHistory target deleted, fall back to shallow

### Context Transform

- Go template with access to `oldContext`, `newVersion`, `contextVersion`
- Executed on version change; failure → cleanStart fallback
- Registry SHALL validate template syntax at load time

```go
// Example context transform template
{{if .oldContext.deprecatedKey}}
newKey: {{.oldContext.deprecatedKey}}
{{end}}
```

### Non-Migratable Properties

- `metadata.boundary` — NEVER migratable (immutable identity)
- `metadata.name` — preserved (identity)

---

## File Structure

```
maelstrom/
├── cmd/maelstrom/
│   ├── kernel.go              # ~50 lines, bootstrap sequence
│   ├── main.go                # Entry point
│   └── kernel_test.go         # Kernel tests
│
├── pkg/registry/
│   ├── registry.go            # ChartRegistry orchestration (~300 lines)
│   ├── hydrator.go            # ChartHydrator (~150 lines)
│   ├── source.go              # Source interface + FileSystemSource
│   ├── service.go             # Service orchestration
│   ├── watcher.go             # FileSystemWatcher (fsnotify)
│   ├── types.go               # Registry types and interfaces
│   └── registry_test.go       # Registry tests
│
├── embedded/
    └── bootstrap.yaml         # Hard-coded bootstrap chart
```

---

## TDD Implementation Plan

### Phase 1.1: Kernel Skeleton

**Test**: `TestKernel_SpawnsBootstrapChart`

- Stub Kernel struct and methods
- Verify exactly one chart spawned directly

**Test**: `TestKernel_WaitsForKernelReady`

- Verify Kernel blocks until kernel_ready event

**Test**: `TestKernel_GoesDormantAfterReady`

- Verify Kernel enters signal-handling mode post-bootstrap

**Implementation**:

- ~50 line kernel.go with minimal config loading
- Register 4 bootstrap actions
- Spawn and drive Bootstrap Chart

### Phase 1.2: ChartRegistry Core

**Test**: `TestRegistry_SetGet`

- Basic storage/retrieval

**Test**: `TestRegistry_VersionTracking`

- Multiple versions stored and retrievable

**Test**: `TestRegistry_HydratesYAML`

- Env substitution and template execution

**Implementation**:

- ~300 line registry.go with versioned storage
- Clone-under-lock pattern for concurrency
- OnChange callbacks for observers

### Phase 1.3: File Watching

**Test**: `TestFileSystemSource_EmitsCreated`

- File creation detected

**Test**: `TestFileSystemSource_EmitsUpdated`

- File changes detected

**Test**: `TestFileSystemSource_Debounces`

- Rapid changes coalesced

**Test**: `TestChartRegistry_WatchesDirectory`

- End-to-end file watching

**Implementation**:

- ~180 line watcher.go using fsnotify
- Single Service.Run() goroutine event loop

### Phase 1.4: Core Services

**Test**: `TestSecurityService_ID`

- Returns "sys:security"

**Test**: `TestCommunicationService_PublishSubscribe`

- Pub/sub works

**Test**: `TestObservabilityService_EmitTrace`

- Traces stored

**Test**: `TestLifecycleService_Spawn`

- Can spawn charts

**Implementation**:

- ~100 lines per service
- Phase 1: pass-through behavior
- Full implementation in Layers 2-5

### Phase 1.5: Bootstrap Chart

**Test**: `TestBootstrapChart_SequentialExecution`

- States start in order

**Test**: `TestBootstrapChart_TransitionsOnReady`

- Each state signals next

**Test**: `TestBootstrapChart_EmitsKernelReady`

- Final event emitted

**Test**: `TestBootstrapChart_HandlesFailure`

- Failure stops bootstrap

**Implementation**:

- Hard-coded YAML embedded in binary
- Sequential compound state with 4 atomic states

### Phase 1.6: Integration

**Test**: `TestFullBootstrapSequence`

- End-to-end bootstrap verification

**Test**: `TestKernel_RegistersRemainingActions`

- Actions registered after kernel_ready

**Implementation**:

- Wire all components together
- Verify complete bootstrap flow

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

## Open Questions

1. **Mail system scope**: Implement full Mail in Phase 1, or stub for Phase 2?
2. **Persistence**: CoreRegistry in-memory only for Phase 1, or add disk persistence?
3. **Bootstrap failure**: Should Bootstrap Chart panic on failure, or retry?

---

## References

- `docs/arch-v1.md` Section 12 - Bootstrap Sequence
- `docs/arch-v1.md` Section 5 - Statechart Library API
- `docs/arch-v1.md` Section 7.1 - Core Platform Services
- `docs/arch-v1.md` Section 9.2 - Mail System
- `docs/arch-v1.md` Section 14.2 - Maelstrom Functional Unit Contracts
- `docs/planning/layer-02-core-services.md` - Core services (deferred from Layer 1)
- `docs/planning/implementation-scope.md` - Parent scope document

---
