# Phase 1.1: Kernel Completion - TDD Implementation Plan

**Branch**: `feat/kernel-completion`  
**Parent Spec**: `docs/planning/layer-01-kernel-bootstrap.md`  
**Status**: Ready for TDD execution

---

## Current State Analysis

### What Exists

- `pkg/kernel/kernel.go` - 157 lines with basic bootstrap sequence
- `pkg/kernel/kernel_test.go` - 2 working tests, 4 placeholder tests
- `cmd/maelstrom/main.go` - Entry point with signal handling
- `pkg/bootstrap/chart.go` - Bootstrap Chart YAML as Go constant
- `pkg/bootstrap/sequence.go` - Bootstrap state machine orchestrator
- `pkg/statechart/` - Complete Library interface (Layer 0)

### What's Missing (per spec)

| Component | Spec Location | Current State |
|-----------|---------------|---------------|
| `KernelConfig` struct | layer-01 line 52-55 | Missing |
| `Shutdown()` method | layer-01 line 69 | Missing |
| `services` map in Kernel | layer-01 line 63 | Missing (has `runtimes` map) |
| Bootstrap action registration | layer-01 lines 77, 93 | Not implemented |
| `failed` state in bootstrap YAML | layer-01 lines 296-299 | Missing |

---

## API Changes Required

### 1. KernelConfig Struct

```go
// KernelConfig holds minimal bootstrap configuration
type KernelConfig struct {
    ChartsDir string            // Path to charts/ directory
    AppVars   map[string]string // Application variables for hydration
}
```

### 2. Kernel Struct Updates

```go
type Kernel struct {
    engine    statechart.Library
    config    KernelConfig              // NEW: configuration
    sequence  *bootstrap.Sequence
    services  map[string]statechart.RuntimeID // NEW: service runtime IDs
    runtimes  map[string]*runtime.ChartRuntime
    mu        sync.RWMutex
}
```

### 3. New Methods

```go
// WithConfig sets the kernel configuration (fluent API)
func (k *Kernel) WithConfig(cfg KernelConfig) *Kernel

// Shutdown gracefully stops all services
func (k *Kernel) Shutdown(ctx context.Context) error

// RegisterBootstrapActions registers the 4 bootstrap actions
func (k *Kernel) RegisterBootstrapActions()
```

---

## TDD Test Plan (Strict Workflow per CLAUDE.md)

### Pre-requisite: Create Branch & Stubs

```bash
git checkout -b feat/kernel-completion
```

Create stubs for all new methods before writing tests:

```go
func (k *Kernel) WithConfig(cfg KernelConfig) *Kernel {
    raise NotImplementedError
}

func (k *Kernel) Shutdown(ctx context.Context) error {
    raise NotImplementedError
}

func (k *Kernel) RegisterBootstrapActions() {
    raise NotImplementedError
}
```

---

## Test Sequence (1 test at a time, per TDD workflow)

### Test 1: `TestKernelConfig_DefaultValues`

**What it verifies**: KernelConfig has sensible defaults when created empty.

**Expected failure**: KernelConfig type doesn't exist yet.

**Minimal implementation**:

```go
type KernelConfig struct {
    ChartsDir string
    AppVars   map[string]string
}
```

---

### Test 2: `TestKernel_WithConfig_SetsConfig`

**What it verifies**: `WithConfig` method sets the config and returns receiver for fluent API.

**Expected failure**: `WithConfig` method returns `NotImplementedError`.

**Minimal implementation**:

```go
func (k *Kernel) WithConfig(cfg KernelConfig) *Kernel {
    k.config = cfg
    return k
}
```

---

### Test 3: `TestKernel_RegisterBootstrapActions_RegistersFourActions`

**What it verifies**: Kernel registers exactly 4 bootstrap actions with the engine before spawning.

**Expected failure**: `RegisterBootstrapActions` returns `NotImplementedError`.

**Minimal implementation**:

```go
func (k *Kernel) RegisterBootstrapActions() {
    k.engine.RegisterAction("securityBootstrap", securityBootstrap)
    k.engine.RegisterAction("communicationBootstrap", communicationBootstrap)
    k.engine.RegisterAction("observabilityBootstrap", observabilityBootstrap)
    k.engine.RegisterAction("lifecycleBootstrap", lifecycleBootstrap)
}
```

**Dependencies**: Bootstrap action stubs (see "Bootstrap Action Registration Strategy" below).

---

### Test 4: `TestKernel_RegisterBootstrapActions_WithNilEngine_NoPanic`

**What it verifies**: RegisterBootstrapActions handles nil engine gracefully (for testing).

**Expected failure**: No nil check in `RegisterBootstrapActions`.

**Minimal implementation**:

```go
func (k *Kernel) RegisterBootstrapActions() {
    if k.engine == nil {
        return
    }
    // ... register actions
}
```

---

### Test 5: `TestKernel_Start_RegistersActionsBeforeSpawn`

**What it verifies**: Kernel calls `RegisterBootstrapActions` before spawning the bootstrap chart.

**Expected failure**: `Start` method doesn't call `RegisterBootstrapActions`.

**Minimal implementation**: Add call to `RegisterBootstrapActions()` at start of `Start()` method.

---

### Test 6: `TestKernel_Start_SpawnsBootstrapChart`

**What it verifies**: Kernel spawns exactly one chart (the bootstrap chart) directly via the engine.

**Expected failure**: Current implementation spawns but doesn't track service IDs.

**Minimal implementation**:

```go
// In Start():
bootstrapRTID, err := k.engine.Spawn(def, nil)
if err != nil {
    return fmt.Errorf("failed to spawn bootstrap runtime: %w", err)
}
k.mu.Lock()
k.services["sys:bootstrap"] = bootstrapRTID
k.mu.Unlock()
```

---

### Test 7: `TestKernel_Start_WaitsForKernelReady`

**What it verifies**: Kernel blocks until bootstrap completes (kernel_ready event).

**Expected failure**: Current implementation waits on context cancellation, not bootstrap completion.

**Minimal implementation**: Add channel to signal bootstrap completion and wait on it.

---

### Test 8: `TestKernel_IsBootstrapComplete_AfterBootstrap`

**What it verifies**: `IsBootstrapComplete` returns true after bootstrap sequence finishes.

**Expected failure**: Already implemented, verify it works with new structure.

**Minimal implementation**: May need to update to check services map.

---

### Test 9: `TestKernel_Shutdown_StopsAllServices`

**What it verifies**: `Shutdown` method stops all registered services.

**Expected failure**: `Shutdown` method returns `NotImplementedError`.

**Minimal implementation**:

```go
func (k *Kernel) Shutdown(ctx context.Context) error {
    k.mu.RLock()
    services := make(map[string]statechart.RuntimeID, len(k.services))
    for k, v := range k.services {
        services[k] = v
    }
    k.mu.RUnlock()
    
    for name, id := range services {
        if err := k.engine.Control(id, statechart.CmdStop); err != nil {
            log.Printf("[kernel] failed to stop %s: %v", name, err)
        }
    }
    return nil
}
```

---

### Test 10: `TestKernel_Shutdown_WithNilEngine_NoPanic`

**What it verifies**: `Shutdown` handles nil engine gracefully.

**Expected failure**: No nil check in `Shutdown`.

**Minimal implementation**:

```go
func (k *Kernel) Shutdown(ctx context.Context) error {
    if k.engine == nil {
        return nil
    }
    // ... stop services
}
```

---

### Test 11: `TestKernel_Shutdown_ContextCancellation`

**What it verifies**: `Shutdown` respects context cancellation.

**Expected failure**: `Shutdown` doesn't check context.

**Minimal implementation**: Add context check in shutdown loop.

---

### Test 12: `TestKernel_GetServiceRuntimeID_ReturnsCorrectID`

**What it verifies**: New method to retrieve service runtime ID for testing.

**Expected failure**: Method doesn't exist.

**Minimal implementation**:

```go
func (k *Kernel) GetServiceRuntimeID(name string) (statechart.RuntimeID, bool) {
    k.mu.RLock()
    defer k.mu.RUnlock()
    id, ok := k.services[name]
    return id, ok
}
```

---

### Test 13: `TestKernel_Start_WithEngine_SpawnsAndControls`

**What it verifies**: Integration test - kernel spawns bootstrap chart and sends start command.

**Expected failure**: Already partially implemented, verify with new structure.

**Minimal implementation**: May need updates to track service IDs.

---

### Test 14: `TestKernel_FullBootstrapSequence`

**What it verifies**: End-to-end test - kernel starts, bootstrap completes, kernel_ready emitted.

**Expected failure**: Bootstrap actions not implemented (placeholder test).

**Minimal implementation**: Mark as placeholder until core services implemented.

---

## Bootstrap Action Registration Strategy

### Action Signatures (per statechart.Library interface)

```go
type ActionFn func(runtimeCtx RuntimeContext, appCtx ApplicationContext, event Event) error
```

### 4 Bootstrap Actions to Register

```go
// securityBootstrap - brings sys:security online
func securityBootstrap(rt statechart.RuntimeContext, appCtx statechart.ApplicationContext, ev statechart.Event) error {
    // Phase 1: Stub - just log and signal ready
    log.Println("[bootstrap] Security service initialized")
    return nil
}

// communicationBootstrap - brings sys:communication online  
func communicationBootstrap(rt statechart.RuntimeContext, appCtx statechart.ApplicationContext, ev statechart.Event) error {
    log.Println("[bootstrap] Communication service initialized")
    return nil
}

// observabilityBootstrap - brings sys:observability online
func observabilityBootstrap(rt statechart.RuntimeContext, appCtx statechart.ApplicationContext, ev statechart.Event) error {
    log.Println("[bootstrap] Observability service initialized")
    return nil
}

// lifecycleBootstrap - brings sys:lifecycle online
func lifecycleBootstrap(rt statechart.RuntimeContext, appCtx statechart.ApplicationContext, ev statechart.Event) error {
    log.Println("[bootstrap] Lifecycle service initialized")
    return nil
}
```

### File Location

Create `pkg/kernel/actions.go`:

```go
package kernel

import (
    "log"
    "github.com/maelstrom/v3/pkg/statechart"
)

// securityBootstrap ...
func securityBootstrap(rt statechart.RuntimeContext, appCtx statechart.ApplicationContext, ev statechart.Event) error {
    // ...
}

// ... other actions
```

---

## Dependencies on Other Components

### 1. pkg/statechart (Layer 0) - COMPLETE

- `Library` interface - Spawn, Control, RegisterAction
- `RuntimeContext` type
- `Event` type
- `ControlCmd` constants

### 2. pkg/bootstrap - PARTIAL

- `Sequence` - Already implemented
- `LoadBootstrapChart` - Already implemented
- Missing: `failed` state in bootstrap YAML

### 3. pkg/registry - COMPLETE

- `Registry` - Already implemented
- `Service` - Already implemented
- `Source` interface - Already implemented

### 4. pkg/services - STUBS

- All 4 core services are stubs
- Bootstrap actions will be stubs for Phase 1.1

---

## Dependencies Between Tests

```
Test 1 (KernelConfig) ──┐
                        ├──→ Test 2 (WithConfig)
Test 3 (RegisterActions)┘
        │
        ├──→ Test 4 (NilEngine)
        │
        └──→ Test 5 (Start calls Register)
                │
                ├──→ Test 6 (Spawns Bootstrap)
                │
                ├──→ Test 7 (Waits for Ready)
                │
                └──→ Test 13 (Integration)
                        │
                        └──→ Test 14 (Full Sequence)

Test 9 (Shutdown) ──┐
                    ├──→ Test 10 (NilEngine)
Test 12 (GetServiceID)┘
                    │
                    └──→ Test 11 (ContextCancel)
```

---

## File Changes Summary

### New Files

1. `pkg/kernel/actions.go` - Bootstrap action implementations

### Modified Files

1. `pkg/kernel/kernel.go` - Add KernelConfig, services map, Shutdown, RegisterBootstrapActions
2. `pkg/kernel/kernel_test.go` - Add 14 tests per TDD plan
3. `pkg/bootstrap/chart.go` - Add `failed` state to YAML (Phase 1.2)

---

## Commit Strategy (per CLAUDE.md)

### Commit 1: KernelConfig and WithConfig

```bash
git commit -m "feat: add KernelConfig struct and WithConfig method"
```

### Commit 2: Bootstrap Action Registration

```bash
git commit -m "feat: add RegisterBootstrapActions method"
```

### Commit 3: Services Map and Tracking

```bash
git commit -m "feat: add services map to track runtime IDs"
```

### Commit 4: Shutdown Method

```bash
git commit -m "feat: add Shutdown method for graceful termination"
```

### Commit 5: Integration Tests

```bash
git commit -m "feat: add full bootstrap sequence integration test"
```

---

## Risk Assessment

| Risk | Likelihood | Impact | Mitigation |
|------|-----------|--------|------------|
| Bootstrap actions block startup | Low | High | Use stubs that just log |
| Service tracking causes race conditions | Medium | Medium | Use existing RWMutex pattern |
| Shutdown doesn't stop all services | Low | Medium | Comprehensive test coverage |

---

## Open Questions

1. **Should KernelConfig be optional?** - Yes, WithConfig is fluent API
2. **Should bootstrap actions be in kernel package?** - Yes, they're kernel-specific
3. **Should Shutdown use context timeout?** - Yes, for graceful timeout

---

**End of Phase 1.1 TDD Plan**
