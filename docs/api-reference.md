# Maelstrom Phase 1 API Reference

This document provides comprehensive documentation for the Phase 1 Kernel & Bootstrap implementation of the Maelstrom statechart platform.

## Table of Contents

1. [pkg/source](#pkgsource) - Source interface and implementations
2. [pkg/registry](#pkgregistry) - Registry and versioning
3. [pkg/chart](#pkgchart) - Chart definitions and hydration
4. [pkg/runtime](#pkgruntime) - Chart runtime and factory
5. [pkg/bootstrap](#pkgbootstrap) - Bootstrap sequence and chart loading
6. [pkg/kernel](#pkgkernel) - Kernel orchestration

---

## pkg/source

Package `source` provides the abstraction for file change event producers. It decouples event producers (file system watchers, HTTP endpoints, etc.) from consumers (the registry system).

### Types

#### `EventType`

```go
type EventType int
```

Represents the kind of change detected by a Source.

**Constants:**

| Constant | Value | Description |
|----------|-------|-------------|
| `Created` | 0 | A new file was detected |
| `Updated` | 1 | An existing file was modified |
| `Deleted` | 2 | A file was removed |

#### `SourceEvent`

```go
type SourceEvent struct {
    Key       string    // Relative path, e.g., "gateway.yaml"
    Content   []byte    // Raw YAML content (empty for Deleted)
    Type      EventType // Created, Updated, or Deleted
    Timestamp time.Time
}
```

Represents a file change event from a Source.

**Fields:**

| Field | Type | Description |
|-------|------|-------------|
| `Key` | `string` | Relative file path (e.g., "gateway.yaml") |
| `Content` | `[]byte` | Raw file content; empty for Deleted events |
| `Type` | `EventType` | Type of change (Created, Updated, Deleted) |
| `Timestamp` | `time.Time` | When the event occurred |

#### `Source` (Interface)

```go
type Source interface {
    // Events returns a receive-only channel of file changes.
    // The Source owns this channel and closes it on shutdown.
    Events() <-chan SourceEvent

    // Err returns any error after graceful shutdown.
    // Call this after the Events channel is closed.
    Err() error
}
```

Decouples event producers from consumers. Implementations include FileSystemSource, HTTPSource, TestSource, etc.

**Methods:**

| Method | Returns | Description |
|--------|---------|-------------|
| `Events()` | `<-chan SourceEvent` | Returns a receive-only channel of file changes |
| `Err()` | `error` | Returns any error after graceful shutdown |

### FileSystemSource

```go
type FileSystemSource struct {
    // Contains unexported fields
}
```

Watches a directory for YAML file changes using filesystem notifications.

#### `NewFileSystemSource`

```go
func NewFileSystemSource(root string, debounce time.Duration) (*FileSystemSource, error)
```

Creates a new file system source.

**Parameters:**

| Parameter | Type | Description |
|-----------|------|-------------|
| `root` | `string` | Directory path to watch |
| `debounce` | `time.Duration` | Debounce duration for rapid file changes |

**Returns:**

| Return | Description |
|--------|-------------|
| `*FileSystemSource` | Configured file system source |
| `error` | Error if watcher creation fails |

**Example:**

```go
src, err := source.NewFileSystemSource("/etc/maelstrom/charts", 100*time.Millisecond)
if err != nil {
    log.Fatal(err)
}

// Start watching (blocking)
go func() {
    if err := src.Run(); err != nil {
        log.Printf("Source error: %v", err)
    }
}()

// Process events
for evt := range src.Events() {
    fmt.Printf("File %s: %s\n", evt.Key, evt.Type)
}
```

#### `(*FileSystemSource) Run`

```go
func (s *FileSystemSource) Run() error
```

Starts watching the directory (blocking). Performs an initial scan for existing YAML files, emitting Created events for each file found.

**Returns:**

| Return | Description |
|--------|-------------|
| `error` | Error if initial scan fails or watcher encounters an error |

#### `(*FileSystemSource) Stop`

```go
func (s *FileSystemSource) Stop() error
```

Gracefully shuts down the watcher.

**Returns:**

| Return | Description |
|--------|-------------|
| `error` | Error if watcher close fails |

#### `(*FileSystemSource) Events`

```go
func (s *FileSystemSource) Events() <-chan SourceEvent
```

Implements Source interface. Returns the events channel.

#### `(*FileSystemSource) Err`

```go
func (s *FileSystemSource) Err() error
```

Implements Source interface. Returns any error after graceful shutdown.

### ManualSource

```go
type ManualSource struct {
    // Contains unexported fields
}
```

A test helper that allows manual injection of events. Useful for unit testing.

#### `NewManualSource`

```go
func NewManualSource() *ManualSource
```

Creates a source for testing that can receive manual events.

**Returns:**

| Return | Description |
|--------|-------------|
| `*ManualSource` | New manual source with 10-event buffer |

**Example:**

```go
src := source.NewManualSource()

// Send events manually
src.Send(source.SourceEvent{
    Key:       "test.yaml",
    Content:   []byte("id: test\nversion: 1.0.0"),
    Type:      source.Created,
    Timestamp: time.Now(),
})

// Signal completion
src.Close(nil)

// Process in consumer
for evt := range src.Events() {
    // Handle event
}
```

#### `(*ManualSource) Send`

```go
func (m *ManualSource) Send(evt SourceEvent)
```

Delivers an event to the source (non-blocking up to buffer size).

**Parameters:**

| Parameter | Type | Description |
|-----------|------|-------------|
| `evt` | `SourceEvent` | Event to send |

#### `(*ManualSource) Close`

```go
func (m *ManualSource) Close(err error)
```

Signals end of events and stores any error.

**Parameters:**

| Parameter | Type | Description |
|-----------|------|-------------|
| `err` | `error` | Error to report, or nil |

#### `(*ManualSource) Events`

```go
func (m *ManualSource) Events() <-chan SourceEvent
```

Implements Source interface.

#### `(*ManualSource) Err`

```go
func (m *ManualSource) Err() error
```

Implements Source interface.

---

## pkg/registry

Package `registry` provides a versioned key-value store with hooks for transformation and validation.

### Types

#### `Registry`

```go
type Registry struct {
    // Contains unexported fields
}
```

Stores and retrieves values by key with version tracking. The stored values are `interface{}` but hydrators guarantee type consistency.

#### `New`

```go
func New() *Registry
```

Creates a new empty Registry.

**Returns:**

| Return | Description |
|--------|-------------|
| `*Registry` | New empty registry |

**Example:**

```go
reg := registry.New()

// Store a value
reg.Set("gateway.yaml", chartDef)

// Retrieve current version
val, err := reg.Get("gateway.yaml")
if err != nil {
    // Handle not found
}
```

#### `(*Registry) Set`

```go
func (r *Registry) Set(key string, value interface{})
```

Stores a value in the registry as the current version.

**Parameters:**

| Parameter | Type | Description |
|-----------|------|-------------|
| `key` | `string` | Unique identifier for the value |
| `value` | `interface{}` | Value to store |

#### `(*Registry) Get`

```go
func (r *Registry) Get(key string) (interface{}, error)
```

Retrieves the current version of a key.

**Parameters:**

| Parameter | Type | Description |
|-----------|------|-------------|
| `key` | `string` | Key to retrieve |

**Returns:**

| Return | Description |
|--------|-------------|
| `interface{}` | Current value |
| `error` | `ErrNotFound` if key doesn't exist |

#### `(*Registry) GetVersion`

```go
func (r *Registry) GetVersion(key string, version int) (interface{}, error)
```

Retrieves a specific version of a key.

**Parameters:**

| Parameter | Type | Description |
|-----------|------|-------------|
| `key` | `string` | Key to retrieve |
| `version` | `int` | Version index (0-based) |

**Returns:**

| Return | Description |
|--------|-------------|
| `interface{}` | Versioned value |
| `error` | `ErrNotFound` or `ErrVersionNotFound` |

#### `(*Registry) CloneUnderLock`

```go
func (r *Registry) CloneUnderLock(fn func(map[string]interface{}))
```

Executes fn with a read-locked snapshot of the registry. Useful for consistent iteration over all entries.

**Parameters:**

| Parameter | Type | Description |
|-----------|------|-------------|
| `fn` | `func(map[string]interface{})` | Function to execute with snapshot |

**Example:**

```go
reg.CloneUnderLock(func(snapshot map[string]interface{}) {
    for key, value := range snapshot {
        fmt.Printf("%s: %v\n", key, value)
    }
})
```

#### `Version`

```go
type Version struct {
    Data      interface{}
    Timestamp int64
}
```

Represents a single version of a registry entry.

**Fields:**

| Field | Type | Description |
|-------|------|-------------|
| `Data` | `interface{}` | The stored value |
| `Timestamp` | `int64` | Unix nanoseconds when version was created |

### Hooks

#### `PreLoadHook`

```go
type PreLoadHook func(key string, content []byte) ([]byte, error)
```

Transforms raw YAML before hydration.

**Parameters:**

| Parameter | Type | Description |
|-----------|------|-------------|
| `key` | `string` | File key |
| `content` | `[]byte` | Raw YAML content |

**Returns:**

| Return | Description |
|--------|-------------|
| `[]byte` | Transformed content |
| `error` | Error if transformation fails |

#### `PostLoadHook`

```go
type PostLoadHook func(key string, value interface{}) error
```

Validates/processes values after hydration.

**Parameters:**

| Parameter | Type | Description |
|-----------|------|-------------|
| `key` | `string` | File key |
| `value` | `interface{}` | Hydrated value |

**Returns:**

| Return | Description |
|--------|-------------|
| `error` | Error if validation fails |

#### `(*Registry) AddPreLoadHook`

```go
func (r *Registry) AddPreLoadHook(hook PreLoadHook)
```

Registers a hook to transform raw YAML before hydration.

#### `(*Registry) AddPostLoadHook`

```go
func (r *Registry) AddPostLoadHook(hook PostLoadHook)
```

Registers a hook to validate/process values after hydration.

#### `(*Registry) SetWithHooks`

```go
func (r *Registry) SetWithHooks(key string, content []byte, hydrator func([]byte) (interface{}, error)) error
```

Stores a value after applying all pre-load and post-load hooks.

**Parameters:**

| Parameter | Type | Description |
|-----------|------|-------------|
| `key` | `string` | File key |
| `content` | `[]byte` | Raw YAML content |
| `hydrator` | `func([]byte) (interface{}, error)` | Function to hydrate YAML |

**Returns:**

| Return | Description |
|--------|-------------|
| `error` | Error from any hook or hydrator |

**Example:**

```go
// Add a pre-load hook for decryption
reg.AddPreLoadHook(func(key string, content []byte) ([]byte, error) {
    if strings.HasSuffix(key, ".enc") {
        return decrypt(content)
    }
    return content, nil
})

// Add a post-load hook for validation
reg.AddPostLoadHook(func(key string, value interface{}) error {
    if chart, ok := value.(ChartDefinition); ok {
        if chart.ID == "" {
            return fmt.Errorf("chart ID required")
        }
    }
    return nil
})

// Store with hooks applied
err := reg.SetWithHooks("chart.yaml", content, hydrateFunc)
```

### Errors

```go
var (
    ErrNotFound         = errors.New("key not found in registry")
    ErrVersionNotFound  = errors.New("version not found")
)
```

### Service

```go
type Service struct {
    // Contains unexported fields
}
```

Orchestrates a Source and Registry, processing events and notifying observers.

#### `NewService`

```go
func NewService(src source.Source, reg *Registry) *Service
```

Creates a new Service with the given source and registry.

**Parameters:**

| Parameter | Type | Description |
|-----------|------|-------------|
| `src` | `source.Source` | Event source |
| `reg` | `*Registry` | Target registry |

**Returns:**

| Return | Description |
|--------|-------------|
| `*Service` | New service instance |

#### `(*Service) SetHydrator`

```go
func (s *Service) SetHydrator(fn func([]byte) (interface{}, error))
```

Sets the hydrator function (for testing or custom hydration).

#### `(*Service) OnChange`

```go
func (s *Service) OnChange(fn func(key string, value interface{}))
```

Registers a callback for registry updates.

**Parameters:**

| Parameter | Type | Description |
|-----------|------|-------------|
| `fn` | `func(key string, value interface{})` | Callback function |

#### `(*Service) Run`

```go
func (s *Service) Run(ctx context.Context) error
```

Processes source events until context is cancelled.

**Parameters:**

| Parameter | Type | Description |
|-----------|------|-------------|
| `ctx` | `context.Context` | Context for cancellation |

**Returns:**

| Return | Description |
|--------|-------------|
| `error` | Context error or source error |

**Example:**

```go
src := source.NewFileSystemSource("/charts", 100*time.Millisecond)
reg := registry.New()
svc := registry.NewService(src, reg)

// Set up custom hydrator
svc.SetHydrator(func(content []byte) (interface{}, error) {
    var chart ChartDefinition
    err := yaml.Unmarshal(content, &chart)
    return chart, err
})

// Listen for changes
svc.OnChange(func(key string, value interface{}) {
    fmt.Printf("Updated: %s\n", key)
})

// Run until cancelled
ctx, cancel := context.WithCancel(context.Background())
defer cancel()

// Start source in background
go src.Run()

// Process events
svc.Run(ctx)
```

---

## pkg/chart

Package `chart` provides chart definition types, hydration functions, and the ChartRegistry for loading chart definitions from YAML files.

### Types

#### `ChartDefinition`

```go
type ChartDefinition struct {
    ID      string
    Version string
    Spec    map[string]interface{}
}
```

Represents a hydrated chart ready for instantiation.

**Fields:**

| Field | Type | Description |
|-------|------|-------------|
| `ID` | `string` | Unique chart identifier |
| `Version` | `string` | Semantic version string |
| `Spec` | `map[string]interface{}` | Chart specification (states, transitions, etc.) |

#### `(*ChartDefinition) GetID`

```go
func (c ChartDefinition) GetID() string
```

Returns the chart ID (implements HasID interface).

### HydratorFunc

```go
type HydratorFunc func([]byte) (ChartDefinition, error)
```

Transforms raw YAML bytes into a hydrated ChartDefinition.

**Parameters:**

| Parameter | Type | Description |
|-----------|------|-------------|
| `content` | `[]byte` | Raw YAML content |

**Returns:**

| Return | Description |
|--------|-------------|
| `ChartDefinition` | Hydrated chart definition |
| `error` | Error if hydration fails |

### DefaultHydrator

```go
func DefaultHydrator() HydratorFunc
```

Provides environment substitution and template execution.

**Processing Steps:**
1. Apply environment variable substitution (`${VAR}` and `${VAR:-default}`)
2. Execute templates (`{{template "name"}}`)
3. Parse YAML into ChartDefinition
4. Validate (ID and Version required)

**Returns:**

| Return | Description |
|--------|-------------|
| `HydratorFunc` | Configured hydrator function |

**Example:**

```go
hydrator := chart.DefaultHydrator()

yamlContent := []byte(`
id: ${APP_NAME:-gateway}
version: 1.0.0
spec:
  initial: idle
  states:
    idle:
      type: atomic
`)

def, err := hydrator(yamlContent)
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Chart: %s v%s\n", def.ID, def.Version)
```

### Environment Substitution

The default hydrator supports environment variable substitution:

- `${VAR}` - Required variable; errors if not set
- `${VAR:-default}` - Variable with default value

**Example:**

```yaml
id: ${SERVICE_NAME}
database:
  host: ${DB_HOST:-localhost}
  port: ${DB_PORT:-5432}
```

### ChartRegistry

```go
type ChartRegistry struct {
    // Contains unexported fields
}
```

Loads and manages chart definitions from a directory with file watching.

#### `NewChartRegistry`

```go
func NewChartRegistry(dir string, hydrator HydratorFunc) (*ChartRegistry, error)
```

Creates a registry that watches dir for YAML files.

**Parameters:**

| Parameter | Type | Description |
|-----------|------|-------------|
| `dir` | `string` | Directory to watch |
| `hydrator` | `HydratorFunc` | Function to hydrate YAML into ChartDefinition |

**Returns:**

| Return | Description |
|--------|-------------|
| `*ChartRegistry` | Configured chart registry |
| `error` | Error if file system source creation fails |

#### `(*ChartRegistry) Start`

```go
func (r *ChartRegistry) Start(ctx context.Context) error
```

Begins watching and hydrating charts. Blocks until ctx is cancelled.

**Parameters:**

| Parameter | Type | Description |
|-----------|------|-------------|
| `ctx` | `context.Context` | Context for cancellation |

**Returns:**

| Return | Description |
|--------|-------------|
| `error` | Error from service or context |

#### `(*ChartRegistry) Stop`

```go
func (r *ChartRegistry) Stop() error
```

Gracefully shuts down the registry.

#### `(*ChartRegistry) Get`

```go
func (r *ChartRegistry) Get(name string) (ChartDefinition, error)
```

Retrieves the current version of a chart.

**Parameters:**

| Parameter | Type | Description |
|-----------|------|-------------|
| `name` | `string` | Chart filename (e.g., "gateway.yaml") |

**Returns:**

| Return | Description |
|--------|-------------|
| `ChartDefinition` | Current chart definition |
| `error` | Error if not found |

#### `(*ChartRegistry) GetVersion`

```go
func (r *ChartRegistry) GetVersion(name string, version int) (ChartDefinition, error)
```

Retrieves a specific version of a chart.

#### `(*ChartRegistry) OnChange`

```go
func (r *ChartRegistry) OnChange(fn func(key string, def ChartDefinition))
```

Registers a callback for chart updates.

**Example:**

```go
chartReg, err := chart.NewChartRegistry("/etc/maelstrom/charts", chart.DefaultHydrator())
if err != nil {
    log.Fatal(err)
}

// Listen for changes
chartReg.OnChange(func(key string, def ChartDefinition) {
    fmt.Printf("Chart updated: %s (v%s)\n", def.ID, def.Version)
})

// Start watching
ctx, cancel := context.WithCancel(context.Background())
defer cancel()

if err := chartReg.Start(ctx); err != nil {
    log.Fatal(err)
}
```

---

## pkg/runtime

Package `runtime` provides chart runtime management, factory for creating runtimes, and helper functions.

### Types

#### `ChartRuntime`

```go
type ChartRuntime struct {
    ID      string
    Context RuntimeContext
    // Contains unexported fields
}
```

Represents an instantiated, running chart.

**Fields:**

| Field | Type | Description |
|-------|------|-------------|
| `ID` | `string` | Unique runtime instance ID |
| `Context` | `RuntimeContext` | Read-only runtime information |

#### `RuntimeContext`

```go
type RuntimeContext struct {
    ChartID      string
    ParentID     string
    StartTime    int64
    ActiveStates []string
    Metrics      map[string]interface{}
}
```

Provides read-only access to chart runtime information.

**Fields:**

| Field | Type | Description |
|-------|------|-------------|
| `ChartID` | `string` | ID of the chart definition |
| `ParentID` | `string` | ID of parent runtime (for subcharts) |
| `StartTime` | `int64` | Unix timestamp when runtime started |
| `ActiveStates` | `[]string` | Currently active state names |
| `Metrics` | `map[string]interface{}` | Runtime metrics |

#### `Event`

```go
type Event struct {
    Type          string
    Payload       map[string]interface{}
    CorrelationID string
    Source        string
}
```

Represents a statechart event.

**Fields:**

| Field | Type | Description |
|-------|------|-------------|
| `Type` | `string` | Event type/name |
| `Payload` | `map[string]interface{}` | Event data |
| `CorrelationID` | `string` | Tracing correlation ID |
| `Source` | `string` | Event origin |

#### `HasID` (Interface)

```go
type HasID interface {
    GetID() string
}
```

Interface for types that can provide an ID.

### Creating Runtimes

#### `NewChartRuntime`

```go
func NewChartRuntime(id string, def interface{}) (*ChartRuntime, error)
```

Creates a new runtime from a hydrated chart definition. Accepts:
- `map[string]interface{}` (looks for "id" key)
- Types implementing `interface{ GetID() string }`
- `string` (uses as ID directly)

**Parameters:**

| Parameter | Type | Description |
|-----------|------|-------------|
| `id` | `string` | Runtime instance ID |
| `def` | `interface{}` | Chart definition |

**Returns:**

| Return | Description |
|--------|-------------|
| `*ChartRuntime` | New runtime instance |
| `error` | Error if creation fails |

#### `(*ChartRuntime) SendEvent`

```go
func (r *ChartRuntime) SendEvent(evt Event) error
```

Dispatches an event to the runtime (non-blocking).

**Parameters:**

| Parameter | Type | Description |
|-----------|------|-------------|
| `evt` | `Event` | Event to dispatch |

**Returns:**

| Return | Description |
|--------|-------------|
| `error` | Error if event queue is full |

#### `(*ChartRuntime) ActiveStates`

```go
func (r *ChartRuntime) ActiveStates() []string
```

Returns the currently active states.

#### `(*ChartRuntime) Start`

```go
func (r *ChartRuntime) Start(ctx context.Context) error
```

Begins the runtime's event processing loop.

**Parameters:**

| Parameter | Type | Description |
|-----------|------|-------------|
| `ctx` | `context.Context` | Context for cancellation |

**Returns:**

| Return | Description |
|--------|-------------|
| `error` | Context error or nil |

#### `(*ChartRuntime) Stop`

```go
func (r *ChartRuntime) Stop() error
```

Gracefully shuts down the runtime.

### Factory

```go
type Factory struct {
    // Contains unexported fields
}
```

Creates ChartRuntimes from ChartDefinitions stored in a Registry.

#### `NewFactory`

```go
func NewFactory(reg *registry.Registry) *Factory
```

Creates a factory bound to a registry.

**Parameters:**

| Parameter | Type | Description |
|-----------|------|-------------|
| `reg` | `*registry.Registry` | Source registry |

**Example:**

```go
reg := registry.New()
factory := runtime.NewFactory(reg)

// Create single runtime
rt, err := factory.Create("gateway.yaml")
if err != nil {
    log.Fatal(err)
}

// Create all runtimes
runtimes, err := factory.CreateAll()
if err != nil {
    log.Fatal(err)
}
```

#### `(*Factory) Create`

```go
func (f *Factory) Create(key string) (*ChartRuntime, error)
```

Instantiates a single runtime from a definition.

**Parameters:**

| Parameter | Type | Description |
|-----------|------|-------------|
| `key` | `string` | Registry key |

**Returns:**

| Return | Description |
|--------|-------------|
| `*ChartRuntime` | New runtime instance |
| `error` | Error if not found or creation fails |

#### `(*Factory) CreateAll`

```go
func (f *Factory) CreateAll() (map[string]*ChartRuntime, error)
```

Instantiates runtimes for all charts in the registry.

**Returns:**

| Return | Description |
|--------|-------------|
| `map[string]*ChartRuntime` | Map of key to runtime |
| `error` | Error if any creation fails |

#### `(*Factory) Refresh`

```go
func (f *Factory) Refresh(runtimes map[string]*ChartRuntime) error
```

Updates runtimes based on registry changes. Creates runtimes for new charts and stops runtimes for removed charts.

**Parameters:**

| Parameter | Type | Description |
|-----------|------|-------------|
| `runtimes` | `map[string]*ChartRuntime` | Current runtimes map (modified in place) |

**Returns:**

| Return | Description |
|--------|-------------|
| `error` | Error if refresh fails |

**Example:**

```go
runtimes := make(map[string]*ChartRuntime)

// Initial creation
runtimes, _ = factory.CreateAll()

// ... registry changes ...

// Refresh to sync
factory.Refresh(runtimes)
```

### Helper Functions

#### `SendEventToChart`

```go
func SendEventToChart(runtimes map[string]*ChartRuntime, chartID string, evt Event) error
```

Sends an event to a specific runtime by ID.

**Parameters:**

| Parameter | Type | Description |
|-----------|------|-------------|
| `runtimes` | `map[string]*ChartRuntime` | Map of runtimes |
| `chartID` | `string` | Target runtime ID |
| `evt` | `Event` | Event to send |

**Example:**

```go
evt := runtime.Event{
    Type:    "USER_LOGIN",
    Payload: map[string]interface{}{"user": "alice"},
    Source:  "api",
}

err := runtime.SendEventToChart(runtimes, "gateway.yaml", evt)
```

#### `FindOrphanedCharts`

```go
func FindOrphanedCharts(runtimes map[string]*ChartRuntime, registryKeys []string) []string
```

Identifies charts in registry without runtimes.

**Returns:**

| Return | Description |
|--------|-------------|
| `[]string` | Keys that need runtime creation |

#### `FindStaleRuntimes`

```go
func FindStaleRuntimes(runtimes map[string]*ChartRuntime, registryKeys []string) []string
```

Identifies runtimes not in registry.

**Returns:**

| Return | Description |
|--------|-------------|
| `[]string` | Runtime IDs that should be stopped |

---

## pkg/bootstrap

Package `bootstrap` provides the bootstrap sequence for loading core services during kernel startup.

### BootstrapChartYAML

```go
const BootstrapChartYAML = `
id: sys:bootstrap
version: 1.0.0
spec:
  initial: initializing
  states:
    initializing:
      type: atomic
      transitions:
        - event: START_BOOTSTRAP
          target: security
    security:
      type: atomic
      entryActions:
        - loadSecurityService
      transitions:
        - event: SECURITY_READY
          target: communication
    ...
`
```

The hard-coded bootstrap chart that loads 4 core services. This is compiled into the binary and executed first on startup.

**States:**
1. `initializing` - Initial state, transitions to security
2. `security` - Loads Security Service (boundaries, data tainting)
3. `communication` - Loads Communication Service (mail system, routing)
4. `observability` - Loads Observability Service (metrics, logging)
5. `lifecycle` - Loads Lifecycle Service (chart management, hot-reload)
6. `handoff` - Signals kernel ready
7. `complete` - Final state

### CoreServices

```go
var CoreServices = []struct {
    ID          string
    Name        string
    Description string
}{
    {ID: "sys:security", Name: "Security Service", Description: "Inner/DMZ/outer boundaries, data tainting"},
    {ID: "sys:communication", Name: "Communication Service", Description: "Mail system, cross-region routing"},
    {ID: "sys:observability", Name: "Observability Service", Description: "Metrics, logging, health checks"},
    {ID: "sys:lifecycle", Name: "Lifecycle Service", Description: "Chart management, hot-reload"},
}
```

Defines the 4 core services loaded by bootstrap.

### LoadBootstrapChart

```go
func LoadBootstrapChart() (chart.ChartDefinition, error)
```

Parses the hard-coded YAML into a ChartDefinition.

**Returns:**

| Return | Description |
|--------|-------------|
| `chart.ChartDefinition` | Hydrated bootstrap chart |
| `error` | Error if hydration fails |

**Example:**

```go
def, err := bootstrap.LoadBootstrapChart()
if err != nil {
    log.Fatal(err)
}
fmt.Printf("Bootstrap: %s v%s\n", def.ID, def.Version)
```

### Sequence

```go
type Sequence struct {
    // Contains unexported fields
}
```

Orchestrates the bootstrap state machine.

#### `NewSequence`

```go
func NewSequence() *Sequence
```

Creates a new bootstrap sequence starting at "initializing".

**Returns:**

| Return | Description |
|--------|-------------|
| `*Sequence` | New bootstrap sequence |

#### `(*Sequence) OnStateEnter`

```go
func (s *Sequence) OnStateEnter(fn func(state string) error)
```

Registers a callback for state entry (to load services).

**Parameters:**

| Parameter | Type | Description |
|-----------|------|-------------|
| `fn` | `func(state string) error` | Callback function |

#### `(*Sequence) OnComplete`

```go
func (s *Sequence) OnComplete(fn func())
```

Registers a callback for when bootstrap finishes.

**Parameters:**

| Parameter | Type | Description |
|-----------|------|-------------|
| `fn` | `func()` | Callback function |

#### `(*Sequence) Start`

```go
func (s *Sequence) Start(ctx context.Context) error
```

Begins the bootstrap sequence.

**Parameters:**

| Parameter | Type | Description |
|-----------|------|-------------|
| `ctx` | `context.Context` | Context for cancellation |

**Returns:**

| Return | Description |
|--------|-------------|
| `error` | Error if transition fails |

#### `(*Sequence) HandleEvent`

```go
func (s *Sequence) HandleEvent(ctx context.Context, event string) error
```

Processes an event and transitions if valid.

**Parameters:**

| Parameter | Type | Description |
|-----------|------|-------------|
| `ctx` | `context.Context` | Context for cancellation |
| `event` | `string` | Event to process |

**Returns:**

| Return | Description |
|--------|-------------|
| `error` | Error if invalid event or transition fails |

#### `(*Sequence) CurrentState`

```go
func (s *Sequence) CurrentState() string
```

Returns the current bootstrap state.

#### `(*Sequence) IsComplete`

```go
func (s *Sequence) IsComplete() bool
```

Returns true if bootstrap has finished.

**Example:**

```go
seq := bootstrap.NewSequence()

// Set up state entry handler
seq.OnStateEnter(func(state string) error {
    switch state {
    case "security":
        // Load security service
        go loadSecurityService()
    case "communication":
        // Load communication service
        go loadCommunicationService()
    // ... etc
    }
    return nil
})

// Set up completion handler
seq.OnComplete(func() {
    fmt.Println("Bootstrap complete!")
})

// Start bootstrap
ctx := context.Background()
seq.Start(ctx)

// Signal service ready (from service goroutine)
seq.HandleEvent(ctx, "SECURITY_READY")
```

---

## pkg/kernel

Package `kernel` provides the Kernel which orchestrates bootstrap and hands off to ChartRegistry.

### Kernel

```go
type Kernel struct {
    // Contains unexported fields
}
```

Orchestrates bootstrap and hands off to ChartRegistry. The Kernel is the top-level coordinator for the Maelstrom system.

#### `New`

```go
func New(chartReg *chart.ChartRegistry) *Kernel
```

Creates a new Kernel with the given ChartRegistry.

**Parameters:**

| Parameter | Type | Description |
|-----------|------|-------------|
| `chartReg` | `*chart.ChartRegistry` | Chart registry for runtime charts |

**Returns:**

| Return | Description |
|--------|-------------|
| `*Kernel` | New kernel instance |

#### `(*Kernel) Start`

```go
func (k *Kernel) Start(ctx context.Context) error
```

Begins the bootstrap sequence and transitions to runtime. Blocks until context is cancelled.

**Parameters:**

| Parameter | Type | Description |
|-----------|------|-------------|
| `ctx` | `context.Context` | Context for cancellation |

**Returns:**

| Return | Description |
|--------|-------------|
| `error` | Error if bootstrap fails |

**Bootstrap Flow:**
1. Load bootstrap chart definition
2. Create bootstrap sequence
3. Set up state entry handlers for each core service
4. Set up completion handler
5. Start the sequence
6. Wait for completion or cancellation

#### `(*Kernel) IsBootstrapComplete`

```go
func (k *Kernel) IsBootstrapComplete() bool
```

Returns true if bootstrap has finished.

#### `(*Kernel) GetRuntimes`

```go
func (k *Kernel) GetRuntimes() map[string]*runtime.ChartRuntime
```

Returns the currently active runtimes.

**Complete Example:**

```go
package main

import (
    "context"
    "log"
    "os"
    "os/signal"

    "github.com/maelstrom/v3/pkg/chart"
    "github.com/maelstrom/v3/pkg/kernel"
)

func main() {
    // Create chart registry
    chartReg, err := chart.NewChartRegistry("/etc/maelstrom/charts", chart.DefaultHydrator())
    if err != nil {
        log.Fatal(err)
    }

    // Create kernel
    k := kernel.New(chartReg)

    // Set up context with signal handling
    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()

    sigCh := make(chan os.Signal, 1)
    signal.Notify(sigCh, os.Interrupt)
    go func() {
        <-sigCh
        cancel()
    }()

    // Start chart registry in background
    go func() {
        if err := chartReg.Start(ctx); err != nil {
            log.Printf("Chart registry error: %v", err)
        }
    }()

    // Start kernel (blocks until shutdown)
    if err := k.Start(ctx); err != nil && err != context.Canceled {
        log.Fatalf("Kernel error: %v", err)
    }

    // Graceful shutdown
    chartReg.Stop()
    log.Println("Shutdown complete")
}
```

---

## Usage Patterns

### Pattern 1: File-based Chart Loading

```go
// 1. Create chart registry watching a directory
chartReg, err := chart.NewChartRegistry("/etc/maelstrom/charts", chart.DefaultHydrator())
if err != nil {
    log.Fatal(err)
}

// 2. Listen for chart changes
chartReg.OnChange(func(key string, def chart.ChartDefinition) {
    fmt.Printf("Chart %s updated to v%s\n", def.ID, def.Version)
})

// 3. Start watching
ctx := context.Background()
go chartReg.Start(ctx)
```

### Pattern 2: Registry with Hooks

```go
reg := registry.New()

// Add decryption hook
reg.AddPreLoadHook(func(key string, content []byte) ([]byte, error) {
    if strings.HasSuffix(key, ".enc") {
        return decrypt(content)
    }
    return content, nil
})

// Add validation hook
reg.AddPostLoadHook(func(key string, value interface{}) error {
    chart, ok := value.(chart.ChartDefinition)
    if !ok {
        return fmt.Errorf("invalid type")
    }
    if chart.Version == "" {
        return fmt.Errorf("version required")
    }
    return nil
})

// Store with hooks
err := reg.SetWithHooks("chart.yaml", content, func(b []byte) (interface{}, error) {
    return hydrator(b)
})
```

### Pattern 3: Runtime Lifecycle Management

```go
factory := runtime.NewFactory(reg)

// Create initial runtimes
runtimes, _ := factory.CreateAll()

// Start all runtimes
for _, rt := range runtimes {
    go rt.Start(ctx)
}

// Periodically refresh to sync with registry changes
ticker := time.NewTicker(30 * time.Second)
for range ticker.C {
    factory.Refresh(runtimes)
}
```

### Pattern 4: Event Dispatching

```go
// Send event to specific chart
evt := runtime.Event{
    Type:    "USER_ACTION",
    Payload: map[string]interface{}{"action": "click", "target": "button"},
    Source:  "web-ui",
}

err := runtime.SendEventToChart(runtimes, "ui.yaml", evt)
if err != nil {
    // Handle not found or queue full
}
```

### Pattern 5: Bootstrap Integration

```go
// Create kernel with chart registry
k := kernel.New(chartReg)

// The kernel handles bootstrap automatically:
// 1. Loads bootstrap chart (sys:bootstrap)
// 2. Sequences through core services
// 3. Transitions to handoff state
// 4. Goes dormant

// Check bootstrap status
if k.IsBootstrapComplete() {
    // Safe to interact with runtimes
    runtimes := k.GetRuntimes()
    // ...
}
```

---

## Best Practices

1. **Always handle context cancellation** - All long-running operations accept a context and should respond to cancellation promptly.

2. **Use CloneUnderLock for iteration** - When iterating over registry entries, use `CloneUnderLock` to avoid holding locks during processing.

3. **Check for nil before calling** - Observers and hooks may be nil; always check before invoking.

4. **Debounce file changes** - Use appropriate debounce duration (100ms recommended) to avoid processing rapid successive file changes.

5. **Validate charts early** - Use post-load hooks to validate chart definitions before they enter the registry.

6. **Handle errors from Source.Err()** - After the Events channel closes, check `Err()` for any shutdown errors.

7. **Use environment substitution** - Leverage `${VAR:-default}` syntax in chart YAML for configuration that varies by environment.

8. **Monitor bootstrap progress** - Use `IsBootstrapComplete()` before performing operations that depend on core services.

9. **Clean up resources** - Always call `Stop()` on FileSystemSource, ChartRegistry, and ChartRuntime instances.

10. **Use correlation IDs** - Set `CorrelationID` on events for distributed tracing across state transitions.
