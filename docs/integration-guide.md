# Maelstrom Phase 1 Integration Guide

This guide walks you through wiring the full Maelstrom Phase 1 stack from scratch, including the bootstrap sequence, ChartRegistry with hot-reload, and the kernel handoff.

## Table of Contents

1. [Prerequisites](#1-prerequisites)
2. [Creating a ChartRegistry with DefaultHydrator](#2-creating-a-chartregistry-with-defaulthydrator)
3. [Creating and Starting the Kernel](#3-creating-and-starting-the-kernel)
4. [The Bootstrap Sequence in Action](#4-the-bootstrap-sequence-in-action)
5. [After kernel_ready - ChartRegistry Takes Over](#5-after-kernel_ready---chartregistry-takes-over)
6. [Hot-Reload Demonstration](#6-hot-reload-demonstration)

---

## 1. Prerequisites

### Directory Setup

Create the following directory structure for your Maelstrom application:

```bash
mkdir -p my-maelstrom-app/{charts,agents,services,cmd/kernel}
cd my-maelstrom-app
go mod init my-maelstrom-app
```

Directory purposes:
- `charts/` - Chart definitions (YAML files) that will be hot-reloaded
- `agents/` - Agent-specific chart definitions
- `services/` - Service chart definitions
- `cmd/kernel/` - Main kernel executable

### Dependencies

Add the required dependencies to your `go.mod`:

```go
module my-maelstrom-app

go 1.23

require (
    github.com/maelstrom/v3 v0.0.0
    github.com/fsnotify/fsnotify v1.9.0
    gopkg.in/yaml.v3 v3.0.1
)

replace github.com/maelstrom/v3 => /path/to/maelstrom-v4
```

Or if using the local module:

```go
replace github.com/maelstrom/v3 => /home/albert/git/maelstrom-v4
```

Then run:

```bash
go mod tidy
```

---

## 2. Creating a ChartRegistry with DefaultHydrator

The ChartRegistry watches a directory for YAML files, hydrates them (environment substitution, templates), and maintains a versioned registry.

### Example: Basic ChartRegistry Setup

Create `cmd/kernel/main.go`:

```go
package main

import (
    "context"
    "fmt"
    "log"
    "os"
    "os/signal"
    "syscall"
    "time"

    "github.com/maelstrom/v3/pkg/chart"
)

func main() {
    // Create the charts directory if it doesn't exist
    chartsDir := "./charts"
    if err := os.MkdirAll(chartsDir, 0755); err != nil {
        log.Fatalf("Failed to create charts directory: %v", err)
    }

    // Create a ChartRegistry with the DefaultHydrator
    // The hydrator performs: env substitution -> template execution -> YAML parsing -> validation
    registry, err := chart.NewChartRegistry(chartsDir, chart.DefaultHydrator())
    if err != nil {
        log.Fatalf("Failed to create ChartRegistry: %v", err)
    }

    // Register an observer to be notified when charts change
    registry.OnChange(func(key string, def chart.ChartDefinition) {
        log.Printf("[registry] Chart updated: %s (ID=%s, Version=%s)",
            key, def.ID, def.Version)
    })

    // Create a context that can be cancelled
    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()

    // Handle graceful shutdown
    sigChan := make(chan os.Signal, 1)
    signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

    // Start the registry in a goroutine
    go func() {
        log.Println("[main] Starting ChartRegistry...")
        if err := registry.Start(ctx); err != nil && err != context.Canceled {
            log.Printf("[main] Registry error: %v", err)
        }
    }()

    // Give it time to do initial scan
    time.Sleep(200 * time.Millisecond)

    // List any initially loaded charts
    log.Println("[main] Initial charts loaded:")
    // (We'll add chart loading in the next section)

    // Wait for shutdown signal
    <-sigChan
    log.Println("[main] Shutting down...")

    // Graceful shutdown
    if err := registry.Stop(); err != nil {
        log.Printf("[main] Error stopping registry: %v", err)
    }
}
```

### Expected Output (Empty Directory)

```
2025/02/28 10:00:00 [main] Starting ChartRegistry...
2025/02/28 10:00:00 [main] Initial charts loaded:
^C
2025/02/28 10:00:05 [main] Shutting down...
```

### Creating Your First Chart

Create `charts/gateway.yaml`:

```yaml
id: gateway
version: 1.0.0
spec:
  initial: idle
  states:
    idle:
      type: atomic
      transitions:
        - event: start
          target: active
    active:
      type: atomic
```

Run the kernel again:

```bash
go run cmd/kernel/main.go
```

### Expected Output (With Chart)

```
2025/02/28 10:00:00 [main] Starting ChartRegistry...
2025/02/28 10:00:00 [registry] Chart updated: gateway.yaml (ID=gateway, Version=1.0.0)
2025/02/28 10:00:00 [main] Initial charts loaded:
^C
2025/02/28 10:00:05 [main] Shutting down...
```

---

## 3. Creating and Starting the Kernel

The kernel orchestrates the bootstrap sequence and then hands off control to the ChartRegistry.

### Complete Kernel Implementation

Replace `cmd/kernel/main.go` with the full kernel:

```go
package main

import (
    "context"
    "fmt"
    "log"
    "os"
    "os/signal"
    "path/filepath"
    "syscall"
    "time"

    "github.com/maelstrom/v3/pkg/bootstrap"
    "github.com/maelstrom/v3/pkg/chart"
)

// Kernel orchestrates bootstrap and chart registry
type Kernel struct {
    bootstrapSeq *bootstrap.Sequence
    chartReg     *chart.ChartRegistry
    chartsDir    string
    state        string
}

// NewKernel creates a new kernel instance
func NewKernel(chartsDir string) (*Kernel, error) {
    return &Kernel{
        chartsDir: chartsDir,
        state:     "initializing",
    }, nil
}

// Start begins the kernel operation
func (k *Kernel) Start(ctx context.Context) error {
    log.Println("[kernel] Starting kernel...")

    // Step 1: Initialize bootstrap sequence
    k.bootstrapSeq = bootstrap.NewSequence()

    // Step 2: Set up state entry callbacks for loading services
    k.bootstrapSeq.OnStateEnter(func(state string) error {
        log.Printf("[kernel] Entering bootstrap state: %s", state)

        switch state {
        case "security":
            return k.loadSecurityService()
        case "communication":
            return k.loadCommunicationService()
        case "observability":
            return k.loadObservabilityService()
        case "lifecycle":
            return k.loadLifecycleService()
        case "handoff":
            return k.signalKernelReady()
        case "complete":
            log.Println("[kernel] Bootstrap complete")
        }
        return nil
    })

    // Step 3: Set up completion callback
    k.bootstrapSeq.OnComplete(func() {
        log.Println("[kernel] Bootstrap sequence complete, starting ChartRegistry...")
        k.state = "running"

        // Start the ChartRegistry
        go func() {
            if err := k.chartReg.Start(ctx); err != nil && err != context.Canceled {
                log.Printf("[kernel] ChartRegistry error: %v", err)
            }
        }()
    })

    // Step 4: Create ChartRegistry (but don't start it yet)
    var err error
    k.chartReg, err = chart.NewChartRegistry(k.chartsDir, chart.DefaultHydrator())
    if err != nil {
        return fmt.Errorf("failed to create ChartRegistry: %w", err)
    }

    // Register chart change observer
    k.chartReg.OnChange(func(key string, def chart.ChartDefinition) {
        log.Printf("[kernel] Chart loaded/updated: %s (ID=%s, Version=%s)",
            key, def.ID, def.Version)
    })

    // Step 5: Start the bootstrap sequence
    if err := k.bootstrapSeq.Start(ctx); err != nil {
        return fmt.Errorf("bootstrap start failed: %w", err)
    }

    // Step 6: Simulate service ready events (in real implementation,
    // these would come from actual service initialization)
    events := []string{
        "SECURITY_READY",
        "COMMUNICATION_READY",
        "OBSERVABILITY_READY",
        "LIFECYCLE_READY",
        "KERNEL_READY",
    }

    for _, event := range events {
        select {
        case <-ctx.Done():
            return ctx.Err()
        default:
            // Simulate async service initialization time
            time.Sleep(100 * time.Millisecond)
            if err := k.bootstrapSeq.HandleEvent(ctx, event); err != nil {
                return fmt.Errorf("bootstrap event %s failed: %w", event, err)
            }
        }
    }

    // Wait for context cancellation
    <-ctx.Done()
    return ctx.Err()
}

// Stop gracefully shuts down the kernel
func (k *Kernel) Stop() error {
    log.Println("[kernel] Stopping kernel...")
    if k.chartReg != nil {
        return k.chartReg.Stop()
    }
    return nil
}

// Service loading methods (stubs for demonstration)
func (k *Kernel) loadSecurityService() error {
    log.Println("[kernel]   -> Loading security service...")
    // Load sys:security service
    return nil
}

func (k *Kernel) loadCommunicationService() error {
    log.Println("[kernel]   -> Loading communication service...")
    // Load sys:communication service
    return nil
}

func (k *Kernel) loadObservabilityService() error {
    log.Println("[kernel]   -> Loading observability service...")
    // Load sys:observability service
    return nil
}

func (k *Kernel) loadLifecycleService() error {
    log.Println("[kernel]   -> Loading lifecycle service...")
    // Load sys:lifecycle service
    return nil
}

func (k *Kernel) signalKernelReady() error {
    log.Println("[kernel]   -> Signaling kernel ready...")
    return nil
}

func main() {
    // Get the charts directory
    chartsDir := "./charts"
    if len(os.Args) > 1 {
        chartsDir = os.Args[1]
    }

    // Ensure charts directory exists
    if err := os.MkdirAll(chartsDir, 0755); err != nil {
        log.Fatalf("Failed to create charts directory: %v", err)
    }

    absPath, _ := filepath.Abs(chartsDir)
    log.Printf("[main] Using charts directory: %s", absPath)

    // Create and start kernel
    kernel, err := NewKernel(chartsDir)
    if err != nil {
        log.Fatalf("Failed to create kernel: %v", err)
    }

    // Set up graceful shutdown
    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()

    sigChan := make(chan os.Signal, 1)
    signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

    go func() {
        <-sigChan
        log.Println("[main] Received shutdown signal")
        cancel()
    }()

    // Start the kernel (blocks until context cancelled)
    if err := kernel.Start(ctx); err != nil && err != context.Canceled {
        log.Fatalf("Kernel error: %v", err)
    }

    // Graceful shutdown
    if err := kernel.Stop(); err != nil {
        log.Printf("[main] Error during shutdown: %v", err)
    }

    log.Println("[main] Kernel stopped")
}
```

### Expected Output

```
2025/02/28 10:00:00 [main] Using charts directory: /home/user/my-maelstrom-app/charts
2025/02/28 10:00:00 [kernel] Starting kernel...
2025/02/28 10:00:00 [bootstrap] Starting bootstrap sequence
2025/02/28 10:00:00 [bootstrap] Transitioning: initializing -> security
2025/02/28 10:00:00 [kernel] Entering bootstrap state: security
2025/02/28 10:00:00 [kernel]   -> Loading security service...
2025/02/28 10:00:00 [bootstrap] Received event: SECURITY_READY in state: security
2025/02/28 10:00:00 [bootstrap] Transitioning: security -> communication
2025/02/28 10:00:00 [kernel] Entering bootstrap state: communication
2025/02/28 10:00:00 [kernel]   -> Loading communication service...
2025/02/28 10:00:00 [bootstrap] Received event: COMMUNICATION_READY in state: communication
2025/02/28 10:00:00 [bootstrap] Transitioning: communication -> observability
2025/02/28 10:00:00 [kernel] Entering bootstrap state: observability
2025/02/28 10:00:00 [kernel]   -> Loading observability service...
2025/02/28 10:00:00 [bootstrap] Received event: OBSERVABILITY_READY in state: observability
2025/02/28 10:00:00 [bootstrap] Transitioning: observability -> lifecycle
2025/02/28 10:00:00 [kernel] Entering bootstrap state: lifecycle
2025/02/28 10:00:00 [kernel]   -> Loading lifecycle service...
2025/02/28 10:00:00 [bootstrap] Received event: LIFECYCLE_READY in state: lifecycle
2025/02/28 10:00:00 [bootstrap] Transitioning: lifecycle -> handoff
2025/02/28 10:00:00 [kernel] Entering bootstrap state: handoff
2025/02/28 10:00:00 [kernel]   -> Signaling kernel ready...
2025/02/28 10:00:00 [bootstrap] Received event: KERNEL_READY in state: handoff
2025/02/28 10:00:00 [bootstrap] Transitioning: handoff -> complete
2025/02/28 10:00:00 [kernel] Entering bootstrap state: complete
2025/02/28 10:00:00 [kernel] Bootstrap complete
2025/02/28 10:00:00 [kernel] Bootstrap sequence complete, starting ChartRegistry...
^C
2025/02/28 10:00:05 [main] Received shutdown signal
2025/02/28 10:00:05 [kernel] Stopping kernel...
2025/02/28 10:00:05 [main] Kernel stopped
```

---

## 4. The Bootstrap Sequence in Action

The bootstrap sequence is a state machine with 6 states:

```
initializing -> security -> communication -> observability -> lifecycle -> handoff -> complete
```

### Bootstrap Chart YAML

The bootstrap sequence is defined by a hard-coded chart in `/home/albert/git/maelstrom-v4/pkg/bootstrap/chart.go`:

```yaml
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

    communication:
      type: atomic
      entryActions:
        - loadCommunicationService
      transitions:
        - event: COMMUNICATION_READY
          target: observability

    observability:
      type: atomic
      entryActions:
        - loadObservabilityService
      transitions:
        - event: OBSERVABILITY_READY
          target: lifecycle

    lifecycle:
      type: atomic
      entryActions:
        - loadLifecycleService
      transitions:
        - event: LIFECYCLE_READY
          target: handoff

    handoff:
      type: atomic
      entryActions:
        - signalKernelReady
      transitions:
        - event: KERNEL_READY
          target: complete

    complete:
      type: final
```

### Using the Bootstrap Sequence Directly

If you want to use just the bootstrap sequence without the full kernel:

```go
package main

import (
    "context"
    "fmt"
    "log"

    "github.com/maelstrom/v3/pkg/bootstrap"
)

func main() {
    ctx := context.Background()

    // Create the sequence
    seq := bootstrap.NewSequence()

    // Track state entries
    var states []string
    seq.OnStateEnter(func(state string) error {
        states = append(states, state)
        log.Printf("Entering state: %s", state)
        return nil
    })

    // Track completion
    done := make(chan bool)
    seq.OnComplete(func() {
        log.Println("Bootstrap complete!")
        done <- true
    })

    // Start the sequence
    if err := seq.Start(ctx); err != nil {
        log.Fatalf("Start failed: %v", err)
    }

    // Simulate service ready events
    events := []string{
        "SECURITY_READY",
        "COMMUNICATION_READY",
        "OBSERVABILITY_READY",
        "LIFECYCLE_READY",
        "KERNEL_READY",
    }

    for _, event := range events {
        if err := seq.HandleEvent(ctx, event); err != nil {
            log.Fatalf("Event %s failed: %v", event, err)
        }
    }

    // Wait for completion
    <-done

    fmt.Printf("Visited states: %v\n", states)
    fmt.Printf("IsComplete: %v\n", seq.IsComplete())
}
```

### Expected Output

```
2025/02/28 10:00:00 [bootstrap] Starting bootstrap sequence
2025/02/28 10:00:00 [bootstrap] Transitioning: initializing -> security
2025/02/28 10:00:00 Entering state: security
2025/02/28 10:00:00 [bootstrap] Received event: SECURITY_READY in state: security
2025/02/28 10:00:00 [bootstrap] Transitioning: security -> communication
2025/02/28 10:00:00 Entering state: communication
2025/02/28 10:00:00 [bootstrap] Received event: COMMUNICATION_READY in state: communication
2025/02/28 10:00:00 [bootstrap] Transitioning: communication -> observability
2025/02/28 10:00:00 Entering state: observability
2025/02/28 10:00:00 [bootstrap] Received event: OBSERVABILITY_READY in state: observability
2025/02/28 10:00:00 [bootstrap] Transitioning: observability -> lifecycle
2025/02/28 10:00:00 Entering state: lifecycle
2025/02/28 10:00:00 [bootstrap] Received event: LIFECYCLE_READY in state: lifecycle
2025/02/28 10:00:00 [bootstrap] Transitioning: lifecycle -> handoff
2025/02/28 10:00:00 Entering state: handoff
2025/02/28 10:00:00 [bootstrap] Received event: KERNEL_READY in state: handoff
2025/02/28 10:00:00 [bootstrap] Transitioning: handoff -> complete
2025/02/28 10:00:00 Entering state: complete
2025/02/28 10:00:00 [bootstrap] Bootstrap complete
2025/02/28 10:00:00 Bootstrap complete!
Visited states: [security communication observability lifecycle handoff complete]
IsComplete: true
```

---

## 5. After kernel_ready - ChartRegistry Takes Over

Once bootstrap reaches `complete`, the kernel hands off control to the ChartRegistry. The ChartRegistry:

1. Watches the `charts/` directory for YAML files
2. Hydrates charts (env substitution, templates, validation)
3. Maintains version history
4. Notifies observers of changes

### Loading Charts from Code

To retrieve a loaded chart:

```go
// Get the current version of a chart
def, err := registry.Get("gateway.yaml")
if err != nil {
    log.Fatalf("Failed to get chart: %v", err)
}

fmt.Printf("Chart ID: %s\n", def.ID)
fmt.Printf("Version: %s\n", def.Version)
fmt.Printf("Spec: %+v\n", def.Spec)

// Get a specific version (version 0 is first version)
oldDef, err := registry.GetVersion("gateway.yaml", 0)
if err != nil {
    log.Printf("Version not found: %v", err)
}
```

### Complete Example: Bootstrap + ChartRegistry Integration

Create `cmd/kernel/integrated.go`:

```go
package main

import (
    "context"
    "fmt"
    "log"
    "os"
    "os/signal"
    "syscall"
    "time"

    "github.com/maelstrom/v3/pkg/bootstrap"
    "github.com/maelstrom/v3/pkg/chart"
)

type IntegratedKernel struct {
    seq       *bootstrap.Sequence
    registry  *chart.ChartRegistry
    chartsDir string
    charts    map[string]chart.ChartDefinition
}

func NewIntegratedKernel(chartsDir string) (*IntegratedKernel, error) {
    return &IntegratedKernel{
        chartsDir: chartsDir,
        charts:    make(map[string]chart.ChartDefinition),
    }, nil
}

func (k *IntegratedKernel) Start(ctx context.Context) error {
    // Create ChartRegistry early (but don't start it)
    var err error
    k.registry, err = chart.NewChartRegistry(k.chartsDir, chart.DefaultHydrator())
    if err != nil {
        return err
    }

    // Set up chart change tracking
    k.registry.OnChange(func(key string, def chart.ChartDefinition) {
        k.charts[key] = def
        log.Printf("[registry] Chart loaded: %s -> %s@%s", key, def.ID, def.Version)
    })

    // Set up bootstrap sequence
    k.seq = bootstrap.NewSequence()

    k.seq.OnStateEnter(func(state string) error {
        log.Printf("[bootstrap] State: %s", state)
        switch state {
        case "lifecycle":
            log.Println("[bootstrap]   -> Preparing ChartRegistry...")
        case "handoff":
            log.Println("[bootstrap]   -> Ready to handoff")
        case "complete":
            log.Println("[bootstrap]   -> Bootstrap complete, starting ChartRegistry")
            // Start ChartRegistry when bootstrap completes
            go func() {
                if err := k.registry.Start(ctx); err != nil && err != context.Canceled {
                    log.Printf("[registry] Error: %v", err)
                }
            }()
        }
        return nil
    })

    // Start bootstrap
    if err := k.seq.Start(ctx); err != nil {
        return err
    }

    // Drive bootstrap to completion
    events := []string{
        "SECURITY_READY",
        "COMMUNICATION_READY",
        "OBSERVABILITY_READY",
        "LIFECYCLE_READY",
        "KERNEL_READY",
    }

    for _, event := range events {
        select {
        case <-ctx.Done():
            return ctx.Err()
        default:
            time.Sleep(50 * time.Millisecond)
            if err := k.seq.HandleEvent(ctx, event); err != nil {
                return fmt.Errorf("event %s: %w", event, err)
            }
        }
    }

    // Give ChartRegistry time to scan
    time.Sleep(200 * time.Millisecond)

    // Show loaded charts
    log.Printf("[kernel] Loaded %d chart(s)", len(k.charts))
    for key, def := range k.charts {
        log.Printf("[kernel]   - %s: %s@%s", key, def.ID, def.Version)
    }

    // Wait for shutdown
    <-ctx.Done()
    return ctx.Err()
}

func (k *IntegratedKernel) Stop() error {
    if k.registry != nil {
        return k.registry.Stop()
    }
    return nil
}

func main() {
    chartsDir := "./charts"
    if len(os.Args) > 1 {
        chartsDir = os.Args[1]
    }

    os.MkdirAll(chartsDir, 0755)

    kernel, err := NewIntegratedKernel(chartsDir)
    if err != nil {
        log.Fatalf("Failed to create kernel: %v", err)
    }

    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()

    sigChan := make(chan os.Signal, 1)
    signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

    go func() {
        <-sigChan
        log.Println("[main] Shutting down...")
        cancel()
    }()

    if err := kernel.Start(ctx); err != nil && err != context.Canceled {
        log.Fatalf("Kernel error: %v", err)
    }

    kernel.Stop()
    log.Println("[main] Stopped")
}
```

### Test Setup

Create some test charts:

```bash
mkdir -p charts

# Create gateway.yaml
cat > charts/gateway.yaml << 'EOF'
id: gateway
version: 1.0.0
spec:
  initial: idle
  states:
    idle:
      type: atomic
    active:
      type: atomic
EOF

# Create api.yaml
cat > charts/api.yaml << 'EOF'
id: api-service
version: 2.1.0
spec:
  replicas: 3
  routes:
    - /v1/users
    - /v1/orders
EOF
```

### Expected Output

```
2025/02/28 10:00:00 [bootstrap] Starting bootstrap sequence
2025/02/28 10:00:00 [bootstrap] Transitioning: initializing -> security
2025/02/28 10:00:00 [bootstrap] State: security
2025/02/28 10:00:00 [bootstrap] Received event: SECURITY_READY in state: security
2025/02/28 10:00:00 [bootstrap] Transitioning: security -> communication
2025/02/28 10:00:00 [bootstrap] State: communication
2025/02/28 10:00:00 [bootstrap] Received event: COMMUNICATION_READY in state: communication
2025/02/28 10:00:00 [bootstrap] Transitioning: communication -> observability
2025/02/28 10:00:00 [bootstrap] State: observability
2025/02/28 10:00:00 [bootstrap] Received event: OBSERVABILITY_READY in state: observability
2025/02/28 10:00:00 [bootstrap] Transitioning: observability -> lifecycle
2025/02/28 10:00:00 [bootstrap] State: lifecycle
2025/02/28 10:00:00 [bootstrap]   -> Preparing ChartRegistry...
2025/02/28 10:00:00 [bootstrap] Received event: LIFECYCLE_READY in state: lifecycle
2025/02/28 10:00:00 [bootstrap] Transitioning: lifecycle -> handoff
2025/02/28 10:00:00 [bootstrap] State: handoff
2025/02/28 10:00:00 [bootstrap]   -> Ready to handoff
2025/02/28 10:00:00 [bootstrap] Received event: KERNEL_READY in state: handoff
2025/02/28 10:00:00 [bootstrap] Transitioning: handoff -> complete
2025/02/28 10:00:00 [bootstrap] State: complete
2025/02/28 10:00:00 [bootstrap]   -> Bootstrap complete, starting ChartRegistry
2025/02/28 10:00:00 [bootstrap] Bootstrap complete
2025/02/28 10:00:00 [registry] Chart loaded: gateway.yaml -> gateway@1.0.0
2025/02/28 10:00:00 [registry] Chart loaded: api.yaml -> api-service@2.1.0
2025/02/28 10:00:00 [kernel] Loaded 2 chart(s)
2025/02/28 10:00:00 [kernel]   - api.yaml: api-service@2.1.0
2025/02/28 10:00:00 [kernel]   - gateway.yaml: gateway@1.0.0
^C
2025/02/28 10:00:05 [main] Shutting down...
2025/02/28 10:00:05 [main] Stopped
```

---

## 6. Hot-Reload Demonstration

The FileSystemSource watches for file changes and emits events (Created, Updated, Deleted). The RegistryService hydrates and stores the updated charts.

### How Hot-Reload Works

```
File Change -> fsnotify Event -> Debounce (100ms) -> SourceEvent ->
Hydrator -> Registry.Set() -> Notify Observers
```

### Hot-Reload Demo Script

Create a demo program that shows hot-reload in action:

```go
package main

import (
    "context"
    "fmt"
    "log"
    "os"
    "path/filepath"
    "time"

    "github.com/maelstrom/v3/pkg/chart"
)

func main() {
    chartsDir := "./charts"
    os.MkdirAll(chartsDir, 0755)

    log.Println("=== Hot-Reload Demo ===")
    log.Printf("Charts directory: %s\n", chartsDir)

    // Create registry
    registry, err := chart.NewChartRegistry(chartsDir, chart.DefaultHydrator())
    if err != nil {
        log.Fatalf("Failed to create registry: %v", err)
    }

    // Track changes
    changeCount := 0
    registry.OnChange(func(key string, def chart.ChartDefinition) {
        changeCount++
        log.Printf("[CHANGE #%d] %s -> %s@%s", changeCount, key, def.ID, def.Version)
    })

    // Start registry
    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()

    go func() {
        registry.Start(ctx)
    }()

    // Wait for initial scan
    time.Sleep(200 * time.Millisecond)

    log.Println("\n--- Step 1: Create first chart ---")
    writeChart(chartsDir, "service.yaml", `
id: my-service
version: 1.0.0
spec:
  port: 8080
`)
    time.Sleep(300 * time.Millisecond)

    log.Println("\n--- Step 2: Update the chart ---")
    writeChart(chartsDir, "service.yaml", `
id: my-service
version: 1.1.0
spec:
  port: 9090
  replicas: 3
`)
    time.Sleep(300 * time.Millisecond)

    log.Println("\n--- Step 3: Create another chart ---")
    writeChart(chartsDir, "worker.yaml", `
id: background-worker
version: 2.0.0
spec:
  queue: tasks
  concurrency: 10
`)
    time.Sleep(300 * time.Millisecond)

    log.Println("\n--- Step 4: Update first chart again ---")
    writeChart(chartsDir, "service.yaml", `
id: my-service
version: 1.2.0
spec:
  port: 8080
  replicas: 5
  tls: true
`)
    time.Sleep(300 * time.Millisecond)

    log.Println("\n--- Step 5: Delete a chart ---")
    os.Remove(filepath.Join(chartsDir, "worker.yaml"))
    time.Sleep(300 * time.Millisecond)

    // Show final state
    log.Println("\n--- Final State ---")
    if def, err := registry.Get("service.yaml"); err == nil {
        log.Printf("service.yaml: %s@%s", def.ID, def.Version)
        log.Printf("  Spec: %+v", def.Spec)
    }

    // Try to get deleted chart (should fail or return nil)
    if _, err := registry.Get("worker.yaml"); err != nil {
        log.Printf("worker.yaml: %v", err)
    }

    log.Printf("\nTotal changes observed: %d", changeCount)

    registry.Stop()
}

func writeChart(dir, name, content string) {
    path := filepath.Join(dir, name)
    log.Printf("Writing %s...", name)
    if err := os.WriteFile(path, []byte(content), 0644); err != nil {
        log.Printf("Error writing %s: %v", name, err)
    }
}
```

### Expected Output

```
2025/02/28 10:00:00 === Hot-Reload Demo ===
2025/02/28 10:00:00 Charts directory: ./charts

--- Step 1: Create first chart ---
2025/02/28 10:00:00 Writing service.yaml...
2025/02/28 10:00:00 [CHANGE #1] service.yaml -> my-service@1.0.0

--- Step 2: Update the chart ---
2025/02/28 10:00:00 Writing service.yaml...
2025/02/28 10:00:01 [CHANGE #2] service.yaml -> my-service@1.1.0

--- Step 3: Create another chart ---
2025/02/28 10:00:01 Writing worker.yaml...
2025/02/28 10:00:01 [CHANGE #3] worker.yaml -> background-worker@2.0.0

--- Step 4: Update first chart again ---
2025/02/28 10:00:01 Writing service.yaml...
2025/02/28 10:00:02 [CHANGE #4] service.yaml -> my-service@1.2.0

--- Step 5: Delete a chart ---
2025/02/28 10:00:02 [CHANGE #5] worker.yaml -> @(deleted)

--- Final State ---
2025/02/28 10:00:02 service.yaml: my-service@1.2.0
2025/02/28 10:00:02   Spec: map[port:8080 replicas:5 tls:true]
2025/02/28 10:00:02 worker.yaml: key not found in registry

Total changes observed: 5
```

### Interactive Hot-Reload Demo

For a fully interactive demo, create this program and then manually edit files:

```go
package main

import (
    "context"
    "fmt"
    "log"
    "os"
    "os/signal"
    "syscall"

    "github.com/maelstrom/v3/pkg/chart"
)

func main() {
    chartsDir := "./charts"
    os.MkdirAll(chartsDir, 0755)

    fmt.Println("╔════════════════════════════════════════════════════════════╗")
    fmt.Println("║        Maelstrom Hot-Reload Interactive Demo               ║")
    fmt.Println("╠════════════════════════════════════════════════════════════╣")
    fmt.Printf("║ Charts directory: %-40s ║\n", chartsDir)
    fmt.Println("║                                                            ║")
    fmt.Println("║ Try these commands in another terminal:                    ║")
    fmt.Printf("║   echo 'id: test\\nversion: 1.0.0' > %s/test.yaml    ║\n", chartsDir)
    fmt.Printf("║   echo 'id: test\\nversion: 2.0.0' > %s/test.yaml    ║\n", chartsDir)
    fmt.Printf("║   rm %s/test.yaml                                        ║\n", chartsDir)
    fmt.Println("║                                                            ║")
    fmt.Println("║ Press Ctrl+C to exit                                       ║")
    fmt.Println("╚════════════════════════════════════════════════════════════╝")
    fmt.Println()

    registry, err := chart.NewChartRegistry(chartsDir, chart.DefaultHydrator())
    if err != nil {
        log.Fatalf("Failed to create registry: %v", err)
    }

    registry.OnChange(func(key string, def chart.ChartDefinition) {
        if def.ID == "" {
            log.Printf("[DELETED] %s", key)
        } else {
            log.Printf("[LOADED]  %s -> %s@%s", key, def.ID, def.Version)
        }
    })

    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()

    sigChan := make(chan os.Signal, 1)
    signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

    go func() {
        <-sigChan
        fmt.Println("\nShutting down...")
        cancel()
    }()

    registry.Start(ctx)
}
```

Run it and then in another terminal:

```bash
# Create a chart
echo -e "id: gateway\nversion: 1.0.0\nspec:\n  port: 8080" > charts/gateway.yaml

# Update it
echo -e "id: gateway\nversion: 1.1.0\nspec:\n  port: 9090" > charts/gateway.yaml

# Delete it
rm charts/gateway.yaml
```

---

## Environment Variable Substitution

The DefaultHydrator supports environment variable substitution:

```yaml
id: database-service
version: 1.0.0
spec:
  host: ${DB_HOST}
  port: ${DB_PORT:-5432}
  password: ${DB_PASSWORD}
```

Required variables (no default) will error if not set:

```bash
# This will fail - DB_PASSWORD not set
go run cmd/kernel/main.go

# This will work
export DB_HOST=localhost
export DB_PASSWORD=secret
go run cmd/kernel/main.go
```

---

## Reference: Key Types and Interfaces

### ChartDefinition

```go
type ChartDefinition struct {
    ID      string
    Version string
    Spec    map[string]interface{}
}
```

### ChartRegistry Methods

```go
// Create a new registry watching the given directory
func NewChartRegistry(dir string, hydrator HydratorFunc) (*ChartRegistry, error)

// Start watching (blocks until context cancelled)
func (r *ChartRegistry) Start(ctx context.Context) error

// Stop watching
func (r *ChartRegistry) Stop() error

// Get current version of a chart
func (r *ChartRegistry) Get(name string) (ChartDefinition, error)

// Get specific version (0-indexed)
func (r *ChartRegistry) GetVersion(name string, version int) (ChartDefinition, error)

// Register change callback
func (r *ChartRegistry) OnChange(fn func(key string, def ChartDefinition))
```

### Bootstrap Sequence Methods

```go
// Create new sequence
func NewSequence() *Sequence

// Start the sequence
func (s *Sequence) Start(ctx context.Context) error

// Handle an event (triggers transitions)
func (s *Sequence) HandleEvent(ctx context.Context, event string) error

// Get current state
func (s *Sequence) CurrentState() string

// Check if complete
func (s *Sequence) IsComplete() bool

// Register state entry callback
func (s *Sequence) OnStateEnter(fn func(state string) error)

// Register completion callback
func (s *Sequence) OnComplete(fn func())
```

---

## File Locations

Key source files in the Maelstrom codebase:

- `/home/albert/git/maelstrom-v4/pkg/bootstrap/sequence.go` - Bootstrap sequence implementation
- `/home/albert/git/maelstrom-v4/pkg/bootstrap/chart.go` - Bootstrap chart YAML definition
- `/home/albert/git/maelstrom-v4/pkg/chart/registry.go` - ChartRegistry implementation
- `/home/albert/git/maelstrom-v4/pkg/chart/hydrator.go` - DefaultHydrator with env substitution
- `/home/albert/git/maelstrom-v4/pkg/source/filesystem.go` - FileSystemSource with fsnotify
- `/home/albert/git/maelstrom-v4/pkg/source/types.go` - Source interface and event types
- `/home/albert/git/maelstrom-v4/pkg/registry/service.go` - RegistryService orchestration
- `/home/albert/git/maelstrom-v4/pkg/registry/types.go` - Registry with version tracking

---

## Summary

You now have a complete working integration of the Maelstrom Phase 1 stack:

1. **Bootstrap Sequence**: A state machine that loads 4 core services in order (security -> communication -> observability -> lifecycle)
2. **ChartRegistry**: Watches `charts/` directory, hydrates YAML, maintains versions
3. **Hot-Reload**: File changes are detected, debounced, and loaded automatically
4. **Kernel Integration**: Bootstrap completes, then hands off to ChartRegistry

The architecture is ready for Phase 2: instantiating statecharts from the loaded ChartDefinitions and connecting them to the statechart engine.
