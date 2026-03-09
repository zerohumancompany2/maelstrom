# Phase 1.1A: Kernel Core

## Goal
Implement core kernel functionality with 9 focused tests following strict TDD workflow.

## Scope
- Create `KernelConfig` struct and `WithConfig` method
- Create bootstrap action registration with 4 stub actions
- Implement service tracking with `services` map
- Implement basic `Shutdown` method
- Create 9 tests in `pkg/kernel/kernel_test.go`

## Current Status

### Implementation Status
| Test | Status | Notes |
|------|--------|-------|
| Test 1: KernelConfig_DefaultValues | ❌ Missing | KernelConfig struct does not exist |
| Test 2: WithConfig_SetsConfig | ❌ Missing | WithConfig method does not exist |
| Test 3: RegisterBootstrapActions_RegistersFourActions | ❌ Missing | RegisterBootstrapActions method does not exist |
| Test 4: RegisterBootstrapActions_WithNilEngine_NoPanic | ❌ Missing | Nil check needed |
| Test 5: Start_RegistersActionsBeforeSpawn | ❌ Missing | Must call RegisterBootstrapActions in Start() |
| Test 6: Start_SpawnsBootstrapChart | ⚠️ Partial | ruintimes map exists, need services map |
| Test 7: Start_WaitsForKernelReady | ⚠️ Partial | Currently waits on ctx.Done(), needs readyChan |
| Test 9: Shutdown_StopsAllServices | ❌ Missing | Shutdown method does not exist |
| Test 10: Shutdown_WithNilEngine_NoPanic | ❌ Missing | Nil check needed |

### Files Status
| File | Status |
|------|--------|
| `pkg/kernel/kernel.go` | ⚠️ Partial - needs modifications |
| `pkg/kernel/actions.go` | ❌ **MISSING** - must create |
| `pkg/kernel/kernel_test.go` | ⚠️ Partial - needs 9 new tests |

## Required Implementation

### KernelConfig Struct
```go
type KernelConfig struct {
    ChartsDir string            // Path to charts/ directory
    AppVars   map[string]string // Application variables for hydration
}
```

### Kernel Struct Modifications
```go
type Kernel struct {
    engine    statechart.Library
    config    KernelConfig                    // ADD
    sequence  *bootstrap.Sequence
    services  map[string]statechart.RuntimeID // ADD
    runtimes  map[string]*runtime.ChartRuntime
    mu        sync.RWMutex
    readyChan chan struct{}                   // ADD
}
```

### New Methods

#### WithConfig
```go
func (k *Kernel) WithConfig(cfg KernelConfig) *Kernel {
    k.config = cfg
    return k
}
```

#### RegisterBootstrapActions
```go
func (k *Kernel) RegisterBootstrapActions() {
    if k.engine == nil {
        return
    }
    k.engine.RegisterAction("securityBootstrap", securityBootstrap)
    k.engine.RegisterAction("communicationBootstrap", communicationBootstrap)
    k.engine.RegisterAction("observabilityBootstrap", observabilityBootstrap)
    k.engine.RegisterAction("lifecycleBootstrap", lifecycleBootstrap)
}
```

#### Shutdown
```go
func (k *Kernel) Shutdown(ctx context.Context) error {
    if k.engine == nil {
        return nil
    }
    k.mu.RLock()
    services := make(map[string]statechart.RuntimeID, len(k.services))
    for name, id := range k.services {
        services[name] = id
    }
    k.mu.RUnlock()
    
    for name, id := range services {
        select {
        case <-ctx.Done():
            return ctx.Err()
        default:
        }
        if err := k.engine.Control(id, statechart.CmdStop); err != nil {
            log.Printf("[kernel] failed to stop %s: %v", name, err)
        }
    }
    return nil
}
```

#### GetServiceRuntimeID
```go
func (k *Kernel) GetServiceRuntimeID(name string) (statechart.RuntimeID, bool) {
    k.mu.RLock()
    defer k.mu.RUnlock()
    id, ok := k.services[name]
    return id, ok
}
```

### Bootstrap Actions (pkg/kernel/actions.go)
```go
package kernel

import (
    "log"
    "github.com/maelstrom/v3/pkg/statechart"
)

func securityBootstrap(rt statechart.RuntimeContext, appCtx statechart.ApplicationContext, ev statechart.Event) error {
    log.Println("[bootstrap] Security service initialized")
    return nil
}

func communicationBootstrap(rt statechart.RuntimeContext, appCtx statechart.ApplicationContext, ev statechart.Event) error {
    log.Println("[bootstrap] Communication service initialized")
    return nil
}

func observabilityBootstrap(rt statechart.RuntimeContext, appCtx statechart.ApplicationContext, ev statechart.Event) error {
    log.Println("[bootstrap] Observability service initialized")
    return nil
}

func lifecycleBootstrap(rt statechart.RuntimeContext, appCtx statechart.ApplicationContext, ev statechart.Event) error {
    log.Println("[bootstrap] Lifecycle service initialized")
    return nil
}
```

## Tests to Write (9 tests, 9 commits)

### Test 1: KernelConfig_DefaultValues
```go
func TestKernelConfig_DefaultValues()
```
**Acceptance Criteria:**
- Empty `KernelConfig` has `ChartsDir=""`
- `AppVars` is nil or empty map

### Test 2: WithConfig_SetsConfig
```go
func TestKernel_WithConfig_SetsConfig()
```
**Acceptance Criteria:**
- `WithConfig` sets config field
- Returns receiver for chaining

### Test 3: RegisterBootstrapActions_RegistersFourActions
```go
func TestKernel_RegisterBootstrapActions_RegistersFourActions()
```
**Acceptance Criteria:**
- Exactly 4 actions registered: `securityBootstrap`, `communicationBootstrap`, `observabilityBootstrap`, `lifecycleBootstrap`

### Test 4: RegisterBootstrapActions_WithNilEngine_NoPanic
```go
func TestKernel_RegisterBootstrapActions_WithNilEngine_NoPanic()
```
**Acceptance Criteria:**
- Calling `RegisterBootstrapActions()` with `nil` engine does NOT panic

### Test 5: Start_RegistersActionsBeforeSpawn
```go
func TestKernel_Start_RegistersActionsBeforeSpawn()
```
**Acceptance Criteria:**
- `Start()` calls `RegisterBootstrapActions()` BEFORE `engine.Spawn()`

### Test 6: Start_SpawnsBootstrapChart
```go
func TestKernel_Start_SpawnsBootstrapChart()
```
**Acceptance Criteria:**
- After `Start()`, `services["sys:bootstrap"]` contains valid RuntimeID

### Test 7: Start_WaitsForKernelReady
```go
func TestKernel_Start_WaitsForKernelReady()
```
**Acceptance Criteria:**
- `Start()` blocks until bootstrap completes (kernel_ready event)
- Uses readyChan, not just ctx.Done()

### Test 9: Shutdown_StopsAllServices
```go
func TestKernel_Shutdown_StopsAllServices()
```
**Acceptance Criteria:**
- `Shutdown()` calls `engine.Control(id, CmdStop)` for each service in `services` map

### Test 10: Shutdown_WithNilEngine_NoPanic
```go
func TestKernel_Shutdown_WithNilEngine_NoPanic()
```
**Acceptance Criteria:**
- Calling `Shutdown()` with `nil` engine returns `nil` without panic

## Dependencies

### Test Dependencies
```
Test 1 → Test 2 → (config field in Kernel)
                              ↓
              ┌───────────────┼───────────────┐
              ↓               ↓               ↓
         Test 6            Test 9          Test 3
          (services         (Shutdown      (Register
           map)             needs          Actions)
                            services)              ↓
              ↑                              Test 4
              │                              (nil check)
              │                                   ↓
              └───────────────────────────── Test 5
                                          (Start calls
                                           RegisterActions)
```

### Phase Dependencies
- **Phase 1.2.1** - Bootstrap YAML must have correct action names
- **Layer 0** - statechart.Library.RegisterAction() must be available

## Files to Create/Modify

| File | Operation | Changes |
|------|-----------|---------|
| `pkg/kernel/kernel.go` | MODIFY | Add `KernelConfig` struct, `config` field, `services` map, `readyChan`, `WithConfig()`, `RegisterBootstrapActions()`, `Shutdown()`, `GetServiceRuntimeID()` |
| `pkg/kernel/actions.go` | CREATE | Four bootstrap action stub functions |
| `pkg/kernel/kernel_test.go` | MODIFY | Add 9 new tests |

## Implementation Order (TDD)

1. **Test 1**: Write test → verify RED → Implement `KernelConfig` struct → verify GREEN → commit
2. **Test 2**: Write test → verify RED → Implement `WithConfig()` → verify GREEN → commit
3. **Test 3**: Write test → verify RED → Create `actions.go` → Implement `RegisterBootstrapActions()` → verify GREEN → commit
4. **Test 4**: Write test → verify RED → Add nil check to `RegisterBootstrapActions()` → verify GREEN → commit
5. **Test 5**: Write test → verify RED → Call `RegisterBootstrapActions()` in `Start()` → verify GREEN → commit
6. **Test 6**: Write test → verify RED → Add `services` map → Update `Start()` to populate it → verify GREEN → commit
7. **Test 7**: Write test → verify RED → Add `readyChan` → Update `Start()` to wait for completion → verify GREEN → commit
8. **Test 9**: Write test → verify RED → Implement `Shutdown()` → verify GREEN → commit
9. **Test 10**: Write test → verify RED → Add nil check to `Shutdown()` → verify GREEN → commit

**Total: 9 tests, 9 commits**

## Deliverables
- Modified `pkg/kernel/kernel.go` with new fields and methods
- New `pkg/kernel/actions.go` with 4 bootstrap action stubs
- Updated `pkg/kernel/kernel_test.go` with 9 passing tests
- 9 commits (one per test per TDD workflow)