# Layer 1: Kernel Bootstrap - Implementation Status Report

**Generated**: 2026-03-07  
**Spec**: `docs/planning/layer-01-kernel-bootstrap.md`  
**Status**: ⚠️ PARTIALLY COMPLETE

---

## Executive Summary

| Component | Status | Notes |
|-----------|--------|-------|
| Layer 0 Dependencies (Statechart Engine) | ✅ COMPLETE | All required APIs implemented |
| Kernel Core | ⚠️ PARTIAL | Missing `Shutdown()`, `KernelConfig` |
| ChartRegistry | ✅ COMPLETE | Versioning, hooks, cloning work |
| File Watching (Source) | ✅ COMPLETE | fsnotify with debouncing |
| Bootstrap Chart | ⚠️ PARTIAL | Missing `failed` state, params |
| Core Services (sys:*) | ❌ STUBS ONLY | All 4 services are stubs |
| Test Coverage | ⚠️ GAPS | 11 placeholder tests, 8 missing spec tests |

**Critical Path**: Bootstrap cannot complete because core services are stubs and bootstrap actions are not implemented.

---

## 1. Kernel (`pkg/kernel/kernel.go`)

### What Exists
- `Kernel` struct with statechart engine
- `New()` / `NewWithEngine()` constructors
- `Start(ctx)` method (spec says `Run()`)
- `IsBootstrapComplete()` helper
- `GetRuntimes()` helper
- Main entry point at `cmd/maelstrom/main.go` with signal handling

### What's Missing
| Component | Spec Location | Status |
|-----------|---------------|--------|
| `KernelConfig` struct | Spec line 52-55 | ❌ Missing |
| `Shutdown()` method | Spec line 69 | ❌ Missing |
| `services` map in Kernel struct | Spec line 63 | ❌ Missing |
| Bootstrap action registration | Spec lines 77, 93 | ❌ Not implemented |

### API Discrepancies
| Spec | Actual | Impact |
|------|--------|--------|
| `Run()` | `Start(ctx)` | Minor - different signature |
| `KernelConfig{ChartsDir, AppVars}` | No config struct | Medium - no env-based config |

---

## 2. ChartRegistry (`pkg/registry/`)

### What Exists
| File | Status | Lines |
|------|--------|-------|
| `types.go` | ✅ Complete | Registry struct with versioning |
| `service.go` | ✅ Complete | Service orchestration |
| `registry_test.go` | ✅ Complete | 9 tests, all passing |
| `service_test.go` | ✅ Complete | 5 tests, all passing |

### What's Missing
| Component | Spec Location | Status |
|-----------|---------------|--------|
| `RegistryItem` struct | Spec lines 111-116 | ❌ Missing (VersionID, Raw, Content, Timestamp) |
| `OnChange(fn)` on Registry | Spec line 148 | ⚠️ On Service, not Registry |
| `Stop()` method on Service | Spec line 212 | ❌ Missing |

### Files Moved (Spec vs Actual)
| Spec Location | Actual Location |
|---------------|-----------------|
| `pkg/registry/source.go` | `pkg/source/types.go` |
| `pkg/registry/watcher.go` | `pkg/source/filesystem.go` |
| `pkg/registry/hydrator.go` | `pkg/statechart/types.go` (DefaultHydrator) |

---

## 3. Source Interface (`pkg/source/`)

### What Exists
| Component | Status | Notes |
|-----------|--------|-------|
| `Source` interface | ✅ Complete | `Events()`, `Err()` methods |
| `SourceEvent` struct | ✅ Complete | Key, Content, Type, Timestamp |
| `FileSystemSource` | ✅ Complete | With fsnotify |
| `ManualSource` | ✅ Complete | Test helper |
| Debouncing | ✅ Complete | Implemented |

### Test Coverage
- `filesystem_test.go`: 7 tests, all implemented

---

## 4. Bootstrap Chart (`pkg/bootstrap/chart.go`)

### What Exists
- `BootstrapChartYAML` constant (embedded in Go, not file)
- `LoadBootstrapChart()` function
- States: `initializing` → `security` → `communication` → `observability` → `lifecycle` → `handoff` → `complete`
- Entry actions: `loadSecurityService`, `loadCommunicationService`, etc.

### What's Missing
| Component | Spec Location | Status |
|-----------|---------------|--------|
| `failed` state | Spec lines 296-299 | ❌ Missing |
| State params in entry actions | Spec lines 233-279 | ❌ Missing |
| Hard-coded YAML as file embed | Spec line 447 | ⚠️ Go constant instead |

### YAML Structure Differences
| Spec | Actual |
|------|--------|
| `type: compound` with `initialState` | `spec.initial` format |
| `transitions` with `event`/`target` | ✅ Present |
| `params` in entry actions | ❌ Missing |

---

## 5. Bootstrap Sequence (`pkg/bootstrap/sequence.go`)

### What Exists
- `Sequence` state machine orchestrator
- `Start(ctx)` - begins bootstrap
- `HandleEvent(ctx, event)` - processes transitions
- State callbacks (`OnStateEnter`, `OnComplete`)

### What's Missing
- Bootstrap actions (`securityBootstrap`, `communicationBootstrap`, etc.)
- Actual service spawning (kernel logs but doesn't spawn)

---

## 6. Core Platform Services

### Summary

| Service | Bootstrap Chart | Service Implementation | Tests |
|---------|-----------------|----------------------|-------|
| `sys:security` | ✅ Stub | ❌ Missing | 1 placeholder |
| `sys:communication` | ✅ Stub | ❌ Missing | 2 placeholders |
| `sys:observability` | ✅ Stub | ❌ Missing | 3 placeholders |
| `sys:lifecycle` | ✅ Stub | ❌ Missing | 2 placeholders |

### Detailed Status

#### sys:security
- **Location**: `pkg/services/security/`
- **Exists**: `bootstrap.go` (ChartDefinition stub), `service_test.go` (2 tests)
- **Missing**: `service.go`, boundary enforcement, taint tracking

#### sys:communication
- **Location**: `pkg/services/communication/`
- **Exists**: `bootstrap.go` (ChartDefinition stub), `service_test.go` (2 tests)
- **Missing**: `service.go`, pub/sub implementation, mail routing

#### sys:observability
- **Location**: `pkg/services/observability/`
- **Exists**: `bootstrap.go` (ChartDefinition stub), `service_test.go` (1 test)
- **Missing**: `service.go`, trace storage, metrics, dead-letter queue

#### sys:lifecycle
- **Location**: `pkg/services/lifecycle/`
- **Exists**: `bootstrap.go` (ChartDefinition stub), `service_test.go` (2 tests)
- **Missing**: `service.go`, spawn/stop functionality

---

## 7. Test Coverage Gaps

### Summary

| Category | Total Tests | Implemented | Placeholders | Missing |
|----------|-------------|-------------|--------------|---------|
| Kernel | 6 | 2 | 4 | 1 |
| Registry | 14 | 14 | 0 | 0 |
| Source | 7 | 7 | 0 | 0 |
| Bootstrap | 7 | 7 | 0 | 5 |
| Core Services | 14 | 7 | 8 | 8 |

### Critical Missing Tests (from Spec TDD Plan)

| Spec Test | Location | Status |
|-----------|----------|--------|
| `TestKernel_SpawnsBootstrapChart` | `pkg/kernel/kernel_test.go` | ❌ Missing |
| `TestKernel_WaitsForKernelReady` | `pkg/kernel/kernel_test.go` | ❌ Missing |
| `TestKernel_GoesDormantAfterReady` | `pkg/kernel/kernel_test.go` | ❌ Missing |
| `TestKernel_RegistersRemainingActions` | `pkg/kernel/kernel_test.go` | ❌ Missing |
| `TestBootstrapChart_SequentialExecution` | `pkg/bootstrap/chart_test.go` | ❌ Missing |
| `TestBootstrapChart_TransitionsOnReady` | `pkg/bootstrap/chart_test.go` | ❌ Missing |
| `TestBootstrapChart_EmitsKernelReady` | `pkg/bootstrap/chart_test.go` | ❌ Missing |
| `TestBootstrapChart_HandlesFailure` | `pkg/bootstrap/chart_test.go` | ❌ Missing |

### Placeholder Tests Needing Implementation

| Test | Location |
|------|----------|
| `TestKernel_SpawnsAllServices` | `pkg/kernel/kernel_test.go` |
| `TestKernel_ServicesReady` | `pkg/kernel/kernel_test.go` |
| `TestKernel_KernelReadyEvent` | `pkg/kernel/kernel_test.go` |
| `TestKernel_MailSystemRequired` | `pkg/kernel/kernel_test.go` |
| `TestSecurityService_HandleMail` | `pkg/services/security/service_test.go` |
| `TestCommunicationService_PubSub` | `pkg/services/communication/service_test.go` |
| `TestCommunicationService_RoutesMail` | `pkg/services/communication/service_test.go` |
| `TestObservabilityService_EmitTrace` | `pkg/services/observability/service_test.go` |
| `TestObservabilityService_BoundaryInner` | `pkg/services/observability/service_test.go` |
| `TestLifecycleService_SpawnChart` | `pkg/services/lifecycle/service_test.go` |
| `TestLifecycleService_BoundaryInner` | `pkg/services/lifecycle/service_test.go` |

---

## 8. File Structure Discrepancies

### Expected (per spec)
```
maelstrom/
├── cmd/maelstrom/
│   ├── kernel.go              # ~50 lines
│   ├── main.go
│   └── kernel_test.go
├── pkg/registry/
│   ├── registry.go            # ~300 lines
│   ├── hydrator.go            # ~150 lines
│   ├── source.go
│   ├── service.go
│   ├── watcher.go
│   ├── types.go
│   └── registry_test.go
└── embedded/
    └── bootstrap.yaml
```

### Actual
```
maelstrom/
├── cmd/maelstrom/
│   ├── main.go                # ✅ Present
│   └── main_test.go           # ✅ Present
├── pkg/kernel/                # ⚠️ Different location
│   ├── kernel.go              # ✅ Present (157 lines)
│   ├── kernel_test.go         # ✅ Present
│   └── kernel_chart_definition_test.go
├── pkg/registry/              # ⚠️ Incomplete
│   ├── types.go               # ✅ Present
│   ├── service.go             # ✅ Present
│   ├── registry_test.go       # ✅ Present
│   └── service_test.go        # ✅ Present
├── pkg/source/                # ⚠️ Different location
│   ├── types.go               # ✅ Present
│   ├── filesystem.go          # ✅ Present
│   ├── mock.go                # ✅ Present
│   └── filesystem_test.go     # ✅ Present
├── pkg/bootstrap/             # ⚠️ Different location
│   ├── chart.go               # ✅ Present (YAML as Go constant)
│   ├── sequence.go            # ✅ Present
│   ├── chart_test.go          # ✅ Present
│   └── sequence_test.go       # ✅ Present
├── pkg/services/              # ⚠️ Stubs only
│   ├── security/
│   │   ├── bootstrap.go       # ✅ Stub
│   │   └── service_test.go    # ✅ Partial tests
│   ├── communication/
│   │   ├── bootstrap.go       # ✅ Stub
│   │   └── service_test.go    # ✅ Partial tests
│   ├── observability/
│   │   ├── bootstrap.go       # ✅ Stub
│   │   └── service_test.go    # ✅ Partial tests
│   └── lifecycle/
│       ├── bootstrap.go       # ✅ Stub
│       └── service_test.go    # ✅ Partial tests
└── pkg/statechart/            # ✅ Layer 0 COMPLETE
    ├── engine.go
    ├── types.go
    ├── interfaces.go
    └── ...
```

### Missing Files
- `cmd/maelstrom/kernel.go` (moved to `pkg/kernel/kernel.go`)
- `pkg/registry/registry.go` (partial in `types.go`)
- `pkg/registry/hydrator.go` (in `pkg/statechart/types.go`)
- `pkg/registry/source.go` (moved to `pkg/source/types.go`)
- `pkg/registry/watcher.go` (moved to `pkg/source/filesystem.go`)
- `embedded/bootstrap.yaml` (embedded as Go constant in `pkg/bootstrap/chart.go`)

---

## 9. What Remains to Be Done

### Phase 1.1: Kernel Completion (TDD Plan)
1. Add `KernelConfig` struct with `ChartsDir` and `AppVars`
2. Add `Shutdown()` method for graceful termination
3. Implement bootstrap action registration
4. Write missing tests:
   - `TestKernel_SpawnsBootstrapChart`
   - `TestKernel_WaitsForKernelReady`
   - `TestKernel_GoesDormantAfterReady`
   - `TestKernel_RegistersRemainingActions`

### Phase 1.2: Bootstrap Chart Completion
1. Add `failed` state to bootstrap YAML
2. Add params to entry actions
3. Write missing tests:
   - `TestBootstrapChart_SequentialExecution`
   - `TestBootstrapChart_TransitionsOnReady`
   - `TestBootstrapChart_EmitsKernelReady`
   - `TestBootstrapChart_HandlesFailure`

### Phase 1.3: Core Services (Deferred to Layer 2 per spec)
**Note**: Spec says "Full implementation is documented in layer-02-core-services.md"

Minimal implementation needed for Layer 1 to work:
1. Implement `sys:security` stub (pass-through)
2. Implement `sys:communication` stub (basic pub/sub)
3. Implement `sys:observability` stub (in-memory traces)
4. Implement `sys:lifecycle` stub (basic spawn/stop)

### Phase 1.4: Bootstrap Actions
Implement or document why not needed:
- `securityBootstrap`
- `communicationBootstrap`
- `observabilityBootstrap`
- `lifecycleBootstrap`
- `emitKernelReady`

### Phase 1.5: Test Completion
Implement placeholder tests:
- 4 kernel tests
- 8 service tests

### Phase 1.6: Integration
1. Wire all components together
2. Write `TestFullBootstrapSequence`
3. Verify complete bootstrap flow

---

## 10. Blockers

### Critical Blockers
1. **Core services are stubs** - Bootstrap cannot complete without actual service implementations
2. **Bootstrap actions not implemented** - Kernel logs service names but doesn't spawn them
3. **No `failed` state** - Bootstrap cannot handle errors per spec

### Non-Critical Issues
1. File structure differs from spec (acceptable, functionality present)
2. API naming differences (`Run()` vs `Start(ctx)`)
3. Missing `KernelConfig` struct (can be added later)

---

## 11. Recommendations

### Immediate Actions (to satisfy TDD workflow)
1. Create branch `feat/layer-1-complete`
2. Write `TestKernel_SpawnsBootstrapChart` → implement minimal code → commit
3. Write `TestBootstrapChart_HandlesFailure` → add `failed` state → commit
4. Write `TestCommunicationService_PubSub` → implement minimal pub/sub → commit
5. Repeat for each test in spec's TDD plan

### Short-Term
1. Decide: implement full core services now or defer to Layer 2
2. If defer: create minimal stubs that allow bootstrap to complete
3. Add `Shutdown()` method to Kernel
4. Add `KernelConfig` if env-based config needed

### Long-Term
1. Complete Layer 2 (core services) before Layer 1 can be considered done
2. Reconcile file structure with spec (optional)
3. Update spec to reflect actual implementation decisions

---

## 12. Open Questions (from spec)

1. **Mail system scope**: Implement full Mail in Phase 1, or stub for Phase 2?
   - Current: Stub approach (communication service is stub)
   
2. **Persistence**: CoreRegistry in-memory only for Phase 1, or add disk persistence?
   - Current: In-memory only (no persistence implemented)
   
3. **Bootstrap failure**: Should Bootstrap Chart panic on failure, or retry?
   - Current: No `failed` state, behavior undefined

---

**Report Complete**