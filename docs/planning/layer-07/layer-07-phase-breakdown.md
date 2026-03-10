# Layer 7: Platform Services - Phase Breakdown

## Executive Summary

Layer 7 implements the Platform Services that provide system-level capabilities for the Maelstrom runtime. Based on analysis of Layer 2-3 patterns, Layer 7 specification (arch-v1.md L449-492), and current codebase state, this document breaks down Layer 7 into **7 phases** with **~42 tests** across **~42 commits**.

### Current State (Layers 0-6 Complete/In Progress)
- ✅ All 4 hard-coded services have implementations (90% complete)
- ✅ All 8 hot-reloadable services have YAML definitions (85% complete)
- ✅ ChartRegistry service infrastructure exists
- ✅ Bootstrap sequence implemented with KERNEL_READY signaling
- ⚠️ Post-KERNEL_READY hot-reloadable service loading incomplete
- ⚠️ Full hot-reload protocol integration missing
- ⚠️ Some service-specific contracts incomplete (adminCommand 2FA, tool descriptors)

### Layer 7 Goal
Complete Phase 1 behavior per spec (arch-v1.md L449-492):
1. **Hard-coded services** (L457-460): sys:security, sys:communication, sys:observability, sys:lifecycle
2. **Hot-reloadable services** (L466-473): sys:gateway, sys:admin, sys:persistence, sys:heartbeat, sys:memory, sys:human-gateway, sys:tools, sys:datasources
3. **Platform service contract**: `handleMail(mail: Mail) → outcomeEvent` (L479)
4. **Bootstrap integration**: KERNEL_READY signaling (L800-897)
5. **Hot-reload protocol**: Quiescence detection and reload (L865-882)

---

## Phase Overview

| Phase | Name | Tests | Files | Branch | Dependencies |
|-------|------|-------|-------|--------|--------------|
| 7.1 | Hard-Coded Services Completion | 6 | 2 | `feat/layer7-hardcoded-services` | None |
| 7.2 | Hot-Reloadable Services Completion | 8 | 4 | `feat/layer7-hotreloadable` | 7.1 |
| 7.3 | Post-KERNEL_READY Loading | 6 | 3 | `feat/layer7-kernel-ready-loading` | 7.1, 7.2 |
| 7.4 | Hot-Reload Protocol Integration | 7 | 2 | `feat/layer7-hotreload-protocol` | 7.3 |
| 7.5 | Service-Specific Contracts | 8 | 3 | `feat/layer7-service-contracts` | 7.2 |
| 7.6 | Service Discovery & Runtime Registration | 4 | 2 | `feat/layer7-discovery` | 7.3, 7.5 |
| 7.7 | Integration Tests | 3 | 1 | `feat/layer7-integration` | 7.4, 7.6 |

**Total: 42 tests, 17 files, 7 phases**

---

## Phase 7.1: Hard-Coded Services Completion

### Goal
Complete the remaining 10% of hard-coded platform services (sys:security, sys:communication, sys:observability, sys:lifecycle) per spec (arch-v1.md L453-461).

### Scope
- Complete sys:security boundary enforcement (L457)
- Complete sys:communication mail pub/sub (L458)
- Complete sys:observability trace storage (L459)
- Complete sys:lifecycle spawn/stop (L460)
- Verify all 4 services integrate with bootstrap sequence

### Current Status

#### Implementation Status
| Service | Status | Notes |
|---------|--------|-------|
| `sys:security` | ⚠️ 90% complete | Taint propagation exists; needs boundary enforcement tests |
| `sys:communication` | ⚠️ 90% complete | Pub/sub exists; needs integration with bootstrap |
| `sys:observability` | ⚠️ 90% complete | Trace storage exists; needs dead-letter integration |
| `sys:lifecycle` | ⚠️ 90% complete | Spawn/stop exists; needs runtime tracking tests |

#### Files Status
| File | Status |
|------|--------|
| `pkg/services/security/service.go` | ⚠️ Partial - add boundary enforcement tests |
| `pkg/services/communication/service.go` | ⚠️ Partial - add bootstrap integration tests |
| `pkg/services/observability/service.go` | ⚠️ Partial - add dead-letter tests |
| `pkg/services/lifecycle/service.go` | ⚠️ Partial - add runtime tracking tests |

### Required Implementation

#### sys:security Boundary Enforcement (arch-v1.md L457)
```go
// pkg/services/security/service.go
func (s *SecurityService) ValidateAndSanitize(mail mail.Mail, sourceBoundary, targetBoundary mail.BoundaryType) (mail.Mail, error) {
    // Enforce boundary transitions: outer→inner (taint), inner→outer (block SECRET/PII)
}
```

#### sys:communication Pub/Sub (arch-v1.md L458)
```go
// pkg/services/communication/service.go
func (c *CommunicationService) Publish(mail mail.Mail) (ack mail.Ack, err error) {
    // Deliver to subscribers, return Ack with delivery confirmation
}
```

### Tests to Write (6 tests, 6 commits)

#### Test 1: TestHardcodedServices_SecurityBoundaryEnforcement
```go
func TestHardcodedServices_SecurityBoundaryEnforcement(t *testing.T)
```
**Acceptance Criteria:**
- `ValidateAndSanitize()` enforces outer→inner tainting (arch-v1.md L457)
- `ValidateAndSanitize()` blocks inner→outer for SECRET/PII taints
- Returns error on boundary violation

#### Test 2: TestHardcodedServices_SecurityTaintPropagation
```go
func TestHardcodedServices_SecurityTaintPropagation(t *testing.T)
```
**Acceptance Criteria:**
- `TaintPropagate()` recursively adds taints to nested maps
- `TaintPropagate()` merges existing and new taints
- Returns propagated object with `_taints` field

#### Test 3: TestHardcodedServices_CommunicationPublishAck
```go
func TestHardcodedServices_CommunicationPublishAck(t *testing.T)
```
**Acceptance Criteria:**
- `Publish()` returns Ack with MailID matching published mail (arch-v1.md L458)
- Ack.Success is true when delivered to subscriber
- Ack.DeliveredAt is set to current time

#### Test 4: TestHardcodedServices_ObservabilityDeadLetter
```go
func TestHardcodedServices_ObservabilityDeadLetter(t *testing.T)
```
**Acceptance Criteria:**
- `LogDeadLetter()` stores mail and reason (arch-v1.md L459)
- `QueryDeadLetters()` returns all logged entries
- DeadLetterEntry has Logged timestamp

#### Test 5: TestHardcodedServices_LifecycleSpawnTracking
```go
func TestHardcodedServices_LifecycleSpawnTracking(t *testing.T)
```
**Acceptance Criteria:**
- `Spawn()` adds runtime to internal tracking (arch-v1.md L460)
- `List()` includes spawned runtime with RuntimeInfo
- RuntimeInfo has ID, DefinitionID, Boundary fields

#### Test 6: TestHardcodedServices_AllServicesBootstrapIntegration
```go
func TestHardcodedServices_AllServicesBootstrapIntegration(t *testing.T)
```
**Acceptance Criteria:**
- All 4 hard-coded services can be loaded via bootstrap actions (arch-v1.md L800-897)
- Services are registered in kernel service map
- KERNEL_READY event can be emitted after all services ready

### Dependencies

#### Test Dependencies
```
Test 1 → Test 2 (Security service)
Test 3 (Communication service)
Test 4 (Observability service)
Test 5 (Lifecycle service)
Test 6 (All services integration)
```

### Files to Create/Modify

| File | Operation | Changes |
|------|-----------|---------|
| `pkg/services/security/service_test.go` | MODIFY | Add boundary enforcement, taint propagation tests |
| `pkg/services/communication/service_test.go` | MODIFY | Add publish Ack tests |
| `pkg/services/observability/service_test.go` | MODIFY | Add dead-letter tests |
| `pkg/services/lifecycle/service_test.go` | MODIFY | Add spawn tracking tests |

### Implementation Order (TDD)

1. **Test 1**: Write test → verify RED → Implement boundary enforcement → verify GREEN → commit
2. **Test 2**: Write test → verify RED → Implement taint propagation → verify GREEN → commit
3. **Test 3**: Write test → verify RED → Implement Publish Ack → verify GREEN → commit
4. **Test 4**: Write test → verify RED → Implement dead-letter logging → verify GREEN → commit
5. **Test 5**: Write test → verify RED → Implement spawn tracking → verify GREEN → commit
6. **Test 6**: Write test → verify RED → Integrate all services with bootstrap → verify GREEN → commit

**Total: 6 tests, 6 commits**

### Deliverables
- ✅ `sys:security` boundary enforcement complete (arch-v1.md L457)
- ✅ `sys:communication` pub/sub complete (arch-v1.md L458)
- ✅ `sys:observability` dead-letter complete (arch-v1.md L459)
- ✅ `sys:lifecycle` spawn tracking complete (arch-v1.md L460)
- ✅ All 4 services integrate with bootstrap (arch-v1.md L800-897)
- ✅ 6 commits (one per test per TDD workflow)

---

## Phase 7.2: Hot-Reloadable Services Completion

### Goal
Complete the remaining 15% of hot-reloadable platform services (sys:gateway, sys:admin, sys:persistence, sys:heartbeat, sys:memory, sys:human-gateway, sys:tools, sys:datasources) per spec (arch-v1.md L463-474).

### Scope
- Complete sys:gateway channel adapters (L466)
- Complete sys:admin 2FA-gated commands (L467)
- Complete sys:persistence snapshot/restore (L468)
- Complete sys:heartbeat scheduled wake-ups (L469)
- Complete sys:memory vector/graph stores (L470)
- Complete sys:human-gateway chat interface (L471)
- Complete sys:tools registry and resolution (L472)
- Complete sys:datasources pluggable storage (L473)

### Current Status

#### Implementation Status
| Service | Status | Notes |
|---------|--------|-------|
| `sys:gateway` | ⚠️ 85% complete | Adapters exist; needs OpenAPI generation tests |
| `sys:admin` | ⚠️ 85% complete | 2FA exists; needs boundary enforcement tests |
| `sys:persistence` | ⚠️ 85% complete | Storage exists; needs snapshot/restore tests |
| `sys:heartbeat` | ⚠️ 85% complete | Scheduler exists; needs HEARTBEAT.md injection tests |
| `sys:memory` | ⚠️ 85% complete | Vector/graph exists; needs ContextMap injection tests |
| `sys:human-gateway` | ⚠️ 85% complete | Chat exists; needs streaming tests |
| `sys:tools` | ⚠️ 85% complete | Registry exists; needs ToolDescriptor tests |
| `sys:datasources` | ⚠️ 85% complete | Storage exists; needs tainting tests |

#### Files Status
| File | Status |
|------|--------|
| `pkg/services/gateway/service.go` | ⚠️ Partial - add OpenAPI generation |
| `pkg/services/admin/service.go` | ⚠️ Partial - add boundary enforcement |
| `pkg/services/persistence/service.go` | ⚠️ Partial - add snapshot/restore |
| `pkg/services/heartbeat/service.go` | ⚠️ Partial - add HEARTBEAT.md injection |
| `pkg/services/memory/service.go` | ⚠️ Partial - add ContextMap injection |
| `pkg/services/humangateway/service.go` | ⚠️ Partial - add streaming |
| `pkg/services/tools/service.go` | ⚠️ Partial - add ToolDescriptor tests |
| `pkg/services/datasources/service.go` | ⚠️ Partial - add tainting |

### Required Implementation

#### sys:admin 2FA-Gated Commands (arch-v1.md L467, L485)
```go
// pkg/services/admin/service.go
func (s *adminService) ExecuteCommand(cmd string, token string) error {
    // Verify 2FA token, enforce outer-boundary-only
}
```

#### sys:tools ToolDescriptor Resolution (arch-v1.md L487)
```go
// pkg/services/tools/service.go
func (s *toolsService) Resolve(name string, callerBoundary string) (ToolDescriptor, error) {
    // Resolve tool by name, check boundary accessibility
}
```

### Tests to Write (8 tests, 8 commits)

#### Test 1: TestHotreloadableServices_Admin2FAEnforcement
```go
func TestHotreloadableServices_Admin2FAEnforcement(t *testing.T)
```
**Acceptance Criteria:**
- `ExecuteCommand()` requires non-empty 2FA token (arch-v1.md L467, L485)
- `ExecuteCommand()` rejects invalid 2FA tokens
- `ExecuteCommandOnBoundary()` enforces outer-boundary-only

#### Test 2: TestHotreloadableServices_AdminCommandTypes
```go
func TestHotreloadableServices_AdminCommandTypes(t *testing.T)
```
**Acceptance Criteria:**
- `ExecuteCommand()` accepts list, control, queryTaints, inject commands (arch-v1.md L485)
- Each command type is validated

#### Test 3: TestHotreloadableServices_ToolsToolDescriptor
```go
func TestHotreloadableServices_ToolsToolDescriptor(t *testing.T)
```
**Acceptance Criteria:**
- `Resolve()` returns ToolDescriptor with Name, Boundary, Schema, Isolation (arch-v1.md L487)
- ToolDescriptor has TaintOutput, Type, ChartRef, MaxIterations fields

#### Test 4: TestHotreloadableServices_ToolsBoundaryFiltering
```go
func TestHotreloadableServices_ToolsBoundaryFiltering(t *testing.T)
```
**Acceptance Criteria:**
- `Resolve()` checks caller boundary against tool boundary (arch-v1.md L487)
- Inner can access inner/dmz/outer tools
- DMZ can access dmz/outer tools
- Outer can access outer tools only

#### Test 5: TestHotreloadableServices_PersistenceSnapshotRestore
```go
func TestHotreloadableServices_PersistenceSnapshotRestore(t *testing.T)
```
**Acceptance Criteria:**
- `Snapshot(runtimeId)` creates snapshot (arch-v1.md L486)
- `Restore(snapshotId)` restores runtime state
- Snapshot includes taints and session pointer

#### Test 6: TestHotreloadableServices_HeartbeatHEARTBEATInjection
```go
func TestHotreloadableServices_HeartbeatHEARTBEATInjection(t *testing.T)
```
**Acceptance Criteria:**
- Heartbeat service injects HEARTBEAT.md content (arch-v1.md L469)
- Scheduled wake-ups occur at configured interval

#### Test 7: TestHotreloadableServices_MemoryContextMapInjection
```go
func TestHotreloadableServices_MemoryContextMapInjection(t *testing.T)
```
**Acceptance Criteria:**
- Memory service queries vector/graph stores (arch-v1.md L470, L488)
- Returns MessageSlice for ContextMap injection
- Respects boundaryFilter in query

#### Test 8: TestHotreloadableServices_DatasourcesTainting
```go
func TestHotreloadableServices_DatasourcesTainting(t *testing.T)
```
**Acceptance Criteria:**
- `tagOnWrite(path, taints)` tags file with taints (arch-v1.md L473, L489)
- `getTaints(path)` returns []string of taints
- Taints stored in xattr or object tags

### Dependencies

#### Test Dependencies
```
Test 1 → Test 2 (Admin service)
Test 3 → Test 4 (Tools service)
Test 5 (Persistence service)
Test 6 (Heartbeat service)
Test 7 (Memory service)
Test 8 (Datasources service)
```

### Files to Create/Modify

| File | Operation | Changes |
|------|-----------|---------|
| `pkg/services/admin/service_test.go` | MODIFY | Add 2FA enforcement, command type tests |
| `pkg/services/tools/service_test.go` | MODIFY | Add ToolDescriptor, boundary filtering tests |
| `pkg/services/persistence/service_test.go` | MODIFY | Add snapshot/restore tests |
| `pkg/services/heartbeat/service_test.go` | MODIFY | Add HEARTBEAT.md injection tests |
| `pkg/services/memory/service_test.go` | MODIFY | Add ContextMap injection tests |
| `pkg/services/datasources/service_test.go` | MODIFY | Add tainting tests |

### Implementation Order (TDD)

1. **Test 1**: Write test → verify RED → Implement 2FA enforcement → verify GREEN → commit
2. **Test 2**: Write test → verify RED → Implement command type validation → verify GREEN → commit
3. **Test 3**: Write test → verify RED → Implement ToolDescriptor → verify GREEN → commit
4. **Test 4**: Write test → verify RED → Implement boundary filtering → verify GREEN → commit
5. **Test 5**: Write test → verify RED → Implement snapshot/restore → verify GREEN → commit
6. **Test 6**: Write test → verify RED → Implement HEARTBEAT.md injection → verify GREEN → commit
7. **Test 7**: Write test → verify RED → Implement ContextMap injection → verify GREEN → commit
8. **Test 8**: Write test → verify RED → Implement datasources tainting → verify GREEN → commit

**Total: 8 tests, 8 commits**

### Deliverables
- ✅ `sys:admin` 2FA-gated commands complete (arch-v1.md L467, L485)
- ✅ `sys:tools` ToolDescriptor resolution complete (arch-v1.md L472, L487)
- ✅ `sys:persistence` snapshot/restore complete (arch-v1.md L468, L486)
- ✅ `sys:heartbeat` HEARTBEAT.md injection complete (arch-v1.md L469)
- ✅ `sys:memory` ContextMap injection complete (arch-v1.md L470, L488)
- ✅ `sys:datasources` tainting complete (arch-v1.md L473, L489)
- ✅ 8 commits (one per test per TDD workflow)

---

## Phase 7.3: Post-KERNEL_READY Loading

### Goal
Implement ChartRegistry loading of hot-reloadable services after KERNEL_READY event (arch-v1.md L800-897, L463-474).

### Scope
- Implement ChartRegistry Service to watch services/ directory
- Load hot-reloadable services after KERNEL_READY (arch-v1.md L840-847)
- Wire ChartRegistry to Kernel for post-bootstrap loading
- Implement service discovery via ChartRegistry

### Current Status

#### Implementation Status
| Component | Status | Notes |
|-----------|--------|-------|
| `ChartRegistry.Service` | ⚠️ Exists | Basic service infrastructure; needs KERNEL_READY integration |
| `Kernel` | ⚠️ Partial | Has bootstrap sequence; needs post-KERNEL_READY loading |
| `services/` directory | ⚠️ YAML exists | All 8 services have YAML definitions |

#### Files Status
| File | Status |
|------|--------|
| `pkg/registry/service.go` | ⚠️ Partial - add KERNEL_READY integration |
| `pkg/kernel/kernel.go` | ⚠️ Partial - add post-KERNEL_READY loading |
| `var/maelstrom/services/` | ⚠️ Partial - create directory structure |

### Required Implementation

#### ChartRegistry Service Integration (arch-v1.md L840-847)
```go
// pkg/kernel/kernel.go
func (k *Kernel) startChartRegistry(ctx context.Context) error {
    // After KERNEL_READY, start ChartRegistry to load hot-reloadable services
    // sys:gateway, sys:admin, sys:persistence, sys:heartbeat, sys:memory,
    // sys:human-gateway, sys:tools, sys:datasources
}
```

#### Service Loading from YAML (arch-v1.md L463-474)
```go
// pkg/registry/service.go
func (s *Service) loadPlatformServices() error {
    // Load all PlatformService YAML files from services/ directory
    // Spawn each as ChartRuntime via lifecycle service
}
```

### Tests to Write (6 tests, 6 commits)

#### Test 1: TestPostKernelReadyLoading_ChartRegistryStartsAfterKERNEL_READY
```go
func TestPostKernelReadyLoading_ChartRegistryStartsAfterKERNEL_READY(t *testing.T)
```
**Acceptance Criteria:**
- ChartRegistry.Service starts only after KERNEL_READY event (arch-v1.md L840-847)
- Kernel waits for KERNEL_READY before starting ChartRegistry

#### Test 2: TestPostKernelReadyLoading_ServicesLoadedFromDirectory
```go
func TestPostKernelReadyLoading_ServicesLoadedFromDirectory(t *testing.T)
```
**Acceptance Criteria:**
- ChartRegistry loads services from `var/maelstrom/services/` directory (arch-v1.md L892-897)
- All 8 hot-reloadable service YAML files are discovered

#### Test 3: TestPostKernelReadyLoading_AllHotReloadableServicesSpawn
```go
func TestPostKernelReadyLoading_AllHotReloadableServicesSpawn(t *testing.T)
```
**Acceptance Criteria:**
- All 8 hot-reloadable services are spawned as ChartRuntimes (arch-v1.md L466-473)
- Services: sys:gateway, sys:admin, sys:persistence, sys:heartbeat, sys:memory, sys:human-gateway, sys:tools, sys:datasources

#### Test 4: TestPostKernelReadyLoading_ServicesRegisteredInKernel
```go
func TestPostKernelReadyLoading_ServicesRegisteredInKernel(t *testing.T)
```
**Acceptance Criteria:**
- Spawned services are registered in Kernel service map
- `GetServiceRuntimeID()` returns RuntimeID for each service

#### Test 5: TestPostKernelReadyLoading_ServicesHandleMail
```go
func TestPostKernelReadyLoading_ServicesHandleMail(t *testing.T)
```
**Acceptance Criteria:**
- Each service implements `handleMail(mail: Mail) → outcomeEvent` contract (arch-v1.md L479)
- Services can receive and process Mail via CommunicationService

#### Test 6: TestPostKernelReadyLoading_KernelGoesDormant
```go
func TestPostKernelReadyLoading_KernelGoesDormant(t *testing.T)
```
**Acceptance Criteria:**
- Kernel enters dormant state after handoff to ChartRegistry (arch-v1.md L848-849)
- Kernel only listens for shutdown signals

### Dependencies

#### Test Dependencies
```
Test 1 → Test 2 (ChartRegistry startup)
Test 3 (Service spawning)
Test 4 (Service registration)
Test 5 (Service mail handling)
Test 6 (Kernel dormant state)
```

### Files to Create/Modify

| File | Operation | Changes |
|------|-----------|---------|
| `pkg/kernel/kernel.go` | MODIFY | Add post-KERNEL_READY ChartRegistry loading |
| `pkg/registry/service.go` | MODIFY | Add platform service loading from YAML |
| `var/maelstrom/services/` | CREATE | Directory structure for service YAML files |

### Implementation Order (TDD)

1. **Test 1**: Write test → verify RED → Implement KERNEL_READY gating → verify GREEN → commit
2. **Test 2**: Write test → verify RED → Implement directory-based loading → verify GREEN → commit
3. **Test 3**: Write test → verify RED → Implement service spawning → verify GREEN → commit
4. **Test 4**: Write test → verify RED → Implement service registration → verify GREEN → commit
5. **Test 5**: Write test → verify RED → Implement handleMail contract → verify GREEN → commit
6. **Test 6**: Write test → verify RED → Implement kernel dormant state → verify GREEN → commit

**Total: 6 tests, 6 commits**

### Deliverables
- ✅ ChartRegistry starts after KERNEL_READY (arch-v1.md L840-847)
- ✅ All 8 hot-reloadable services loaded from YAML (arch-v1.md L466-473)
- ✅ Services spawned as ChartRuntimes
- ✅ Services registered in Kernel
- ✅ Services implement handleMail contract (arch-v1.md L479)
- ✅ Kernel goes dormant after handoff (arch-v1.md L848-849)
- ✅ 6 commits (one per test per TDD workflow)

---

## Phase 7.4: Hot-Reload Protocol Integration

### Goal
Implement full hot-reload protocol with quiescence detection and version migration (arch-v1.md L865-882).

### Scope
- Implement quiescence detection (empty queue, no active regions, no inflight tools)
- Implement hot-reload protocol (prepareForReload, quiescence wait, force-stop)
- Implement history mechanisms (shallowHistory, deepHistory, cleanStart)
- Implement context transform validation

### Current Status

#### Implementation Status
| Component | Status | Notes |
|-----------|--------|-------|
| `Quiescence detection` | ❌ Missing | Need to implement queue/region/tool checks |
| `Hot-reload protocol` | ❌ Missing | Need prepareForReload, timeout, force-stop |
| `History mechanisms` | ❌ Missing | Need shallow/deep history, cleanStart |
| `Context transform` | ❌ Missing | Need Go template validation |

#### Files Status
| File | Status |
|------|--------|
| `pkg/runtime/runtime.go` | ⚠️ Partial - add quiescence detection |
| `pkg/registry/service.go` | ⚠️ Partial - add hot-reload protocol |

### Required Implementation

#### Quiescence Detection (arch-v1.md L865-868)
```go
// pkg/runtime/runtime.go
func (r *ChartRuntime) IsQuiescent() bool {
    // Event queue empty AND no active parallel regions AND no inflight tools
}
```

#### Hot-Reload Protocol (arch-v1.md L870-882)
```go
// pkg/registry/service.go
func (s *Service) hotReload(chartName string, newDef ChartDefinition) error {
    // 1. Signal prepareForReload
    // 2. Wait for quiescence (with timeout)
    // 3. IF quiescent: stop, spawn with history
    // 4. IF timeout: force-stop, cleanStart
    // 5. IF maxWaitAttempts exceeded: log failure, require admin intervention
}
```

### Tests to Write (7 tests, 7 commits)

#### Test 1: TestHotReloadProtocol_QuiescenceDetection
```go
func TestHotReloadProtocol_QuiescenceDetection(t *testing.T)
```
**Acceptance Criteria:**
- `IsQuiescent()` returns true when event queue empty (arch-v1.md L865)
- `IsQuiescent()` returns true when no active parallel regions
- `IsQuiescent()` returns true when no inflight tool calls

#### Test 2: TestHotReloadProtocol_ProtocolFlow
```go
func TestHotReloadProtocol_ProtocolFlow(t *testing.T)
```
**Acceptance Criteria:**
- Hot-reload signals `prepareForReload` to current runtime (arch-v1.md L870)
- Runtime attempts to reach quiescence within timeoutMs

#### Test 3: TestHotReloadProtocol_QuiescenceReached
```go
func TestHotReloadProtocol_QuiescenceReached(t *testing.T)
```
**Acceptance Criteria:**
- IF quiescence reached: stop current runtime (arch-v1.md L873-875)
- Spawn new runtime with history (shallow/deep)
- Apply contextTransform if version changed

#### Test 4: TestHotReloadProtocol_TimeoutForceStop
```go
func TestHotReloadProtocol_TimeoutForceStop(t *testing.T)
```
**Acceptance Criteria:**
- IF timeout expires: force-stop current runtime (arch-v1.md L876-878)
- cleanStart with no history preserved
- Increment reload attempt counter

#### Test 5: TestHotReloadProtocol_MaxAttemptsExceeded
```go
func TestHotReloadProtocol_MaxAttemptsExceeded(t *testing.T)
```
**Acceptance Criteria:**
- IF maxWaitAttempts exceeded: log permanent failure to sys:observability (arch-v1.md L879-881)
- Require manual intervention via sys:admin

#### Test 6: TestHotReloadProtocol_HistoryMechanisms
```go
func TestHotReloadProtocol_HistoryMechanisms(t *testing.T)
```
**Acceptance Criteria:**
- `shallowHistory`: restore to parent state's default sub-state (arch-v1.md L884)
- `deepHistory`: restore to specific sub-state (if exists)
- Deleted state fallback: fall back to shallow if deepHistory target deleted

#### Test 7: TestHotReloadProtocol_ContextTransformValidation
```go
func TestHotReloadProtocol_ContextTransformValidation(t *testing.T)
```
**Acceptance Criteria:**
- Registry validates Go template syntax at load time (arch-v1.md L888)
- Context transform executed on version change
- Failure → cleanStart fallback

### Dependencies

#### Test Dependencies
```
Test 1 → Test 2 (Quiescence detection)
Test 3 (Quiescence reached)
Test 4 (Timeout force-stop)
Test 5 (Max attempts exceeded)
Test 6 (History mechanisms)
Test 7 (Context transform)
```

### Files to Create/Modify

| File | Operation | Changes |
|------|-----------|---------|
| `pkg/runtime/runtime.go` | MODIFY | Add IsQuiescent() method |
| `pkg/registry/service.go` | MODIFY | Add hot-reload protocol implementation |

### Implementation Order (TDD)

1. **Test 1**: Write test → verify RED → Implement IsQuiescent() → verify GREEN → commit
2. **Test 2**: Write test → verify RED → Implement prepareForReload signaling → verify GREEN → commit
3. **Test 3**: Write test → verify RED → Implement quiescence path → verify GREEN → commit
4. **Test 4**: Write test → verify RED → Implement timeout force-stop → verify GREEN → commit
5. **Test 5**: Write test → verify RED → Implement max attempts handling → verify GREEN → commit
6. **Test 6**: Write test → verify RED → Implement history mechanisms → verify GREEN → commit
7. **Test 7**: Write test → verify RED → Implement context transform validation → verify GREEN → commit

**Total: 7 tests, 7 commits**

### Deliverables
- ✅ Quiescence detection implemented (arch-v1.md L865-868)
- ✅ Hot-reload protocol implemented (arch-v1.md L870-882)
- ✅ History mechanisms implemented (shallow/deep/cleanStart)
- ✅ Context transform validation implemented
- ✅ 7 commits (one per test per TDD workflow)

---

## Phase 7.5: Service-Specific Contracts

### Goal
Complete service-specific contracts for adminCommand 2FA, tool descriptors, and other service-specific APIs (arch-v1.md L483-491).

### Scope
- sys:admin adminCommand with 2FA (L485)
- sys:persistence snapshot/restore (L486)
- sys:observability emit/query (L487)
- sys:tools resolve with ToolDescriptor (L488)
- sys:memory query with boundaryFilter (L489)
- sys:datasources tagOnWrite/getTaints (L490)

### Current Status

#### Implementation Status
| Contract | Status | Notes |
|----------|--------|-------|
| `adminCommand` | ⚠️ Partial | 2FA exists; needs full command handling |
| `snapshot/restore` | ⚠️ Partial | Storage exists; needs full API |
| `emit/query` | ⚠️ Partial | Trace storage exists; needs query API |
| `resolve` | ⚠️ Partial | Registry exists; needs full ToolDescriptor |
| `memory query` | ⚠️ Partial | Vector/graph exists; needs boundaryFilter |
| `tagOnWrite/getTaints` | ⚠️ Partial | Storage exists; needs taint API |

### Tests to Write (8 tests, 8 commits)

#### Test 1: TestServiceContracts_AdminCommand2FA
```go
func TestServiceContracts_AdminCommand2FA(t *testing.T)
```
**Acceptance Criteria:**
- `adminCommand(cmd, authToken)` requires valid 2FA token (arch-v1.md L485)
- Commands: list, control, queryTaints, inject
- 2FA-gated access enforced

#### Test 2: TestServiceContracts_PersistenceSnapshotRestore
```go
func TestServiceContracts_PersistenceSnapshotRestore(t *testing.T)
```
**Acceptance Criteria:**
- `snapshot(runtimeId)` creates snapshot (arch-v1.md L486)
- `restore(snapshotId)` restores runtime

#### Test 3: TestServiceContracts_ObservabilityEmitQuery
```go
func TestServiceContracts_ObservabilityEmitQuery(t *testing.T)
```
**Acceptance Criteria:**
- `emit(trace)` stores trace (arch-v1.md L487)
- `query(traces, filters)` returns filtered traces

#### Test 4: TestServiceContracts_ToolsResolve
```go
func TestServiceContracts_ToolsResolve(t *testing.T)
```
**Acceptance Criteria:**
- `resolve(name, callerBoundary)` returns ToolDescriptor or notFound (arch-v1.md L488)
- Boundary accessibility enforced

#### Test 5: TestServiceContracts_MemoryQuery
```go
func TestServiceContracts_MemoryQuery(t *testing.T)
```
**Acceptance Criteria:**
- `query(vector, topK, boundaryFilter)` returns MessageSlice (arch-v1.md L489)
- boundaryFilter respected in results

#### Test 6: TestServiceContracts_DatasourcesTagOnWrite
```go
func TestServiceContracts_DatasourcesTagOnWrite(t *testing.T)
```
**Acceptance Criteria:**
- `tagOnWrite(path, taints)` tags path with taints (arch-v1.md L490)

#### Test 7: TestServiceContracts_DatasourcesGetTaints
```go
func TestServiceContracts_DatasourcesGetTaints(t *testing.T)
```
**Acceptance Criteria:**
- `getTaints(path)` returns []string of taints (arch-v1.md L490)

#### Test 8: TestServiceContracts_AllContractsImplemented
```go
func TestServiceContracts_AllContractsImplemented(t *testing.T)
```
**Acceptance Criteria:**
- All service-specific contracts from arch-v1.md L483-491 are implemented
- Each service exposes its contract methods

### Dependencies

#### Test Dependencies
```
Test 1 (Admin contract)
Test 2 (Persistence contract)
Test 3 (Observability contract)
Test 4 (Tools contract)
Test 5 (Memory contract)
Test 6 → Test 7 (Datasources contract)
Test 8 (All contracts)
```

### Files to Create/Modify

| File | Operation | Changes |
|------|-----------|---------|
| `pkg/services/admin/service.go` | MODIFY | Complete adminCommand API |
| `pkg/services/persistence/service.go` | MODIFY | Complete snapshot/restore API |
| `pkg/services/tools/service.go` | MODIFY | Complete ToolDescriptor resolution |

### Implementation Order (TDD)

1. **Test 1**: Write test → verify RED → Implement adminCommand 2FA → verify GREEN → commit
2. **Test 2**: Write test → verify RED → Implement snapshot/restore → verify GREEN → commit
3. **Test 3**: Write test → verify RED → Implement emit/query → verify GREEN → commit
4. **Test 4**: Write test → verify RED → Implement resolve ToolDescriptor → verify GREEN → commit
5. **Test 5**: Write test → verify RED → Implement memory query with boundaryFilter → verify GREEN → commit
6. **Test 6**: Write test → verify RED → Implement tagOnWrite → verify GREEN → commit
7. **Test 7**: Write test → verify RED → Implement getTaints → verify GREEN → commit
8. **Test 8**: Write test → verify RED → Verify all contracts → verify GREEN → commit

**Total: 8 tests, 8 commits**

### Deliverables
- ✅ adminCommand 2FA-gated API complete (arch-v1.md L485)
- ✅ snapshot/restore API complete (arch-v1.md L486)
- ✅ emit/query API complete (arch-v1.md L487)
- ✅ resolve ToolDescriptor API complete (arch-v1.md L488)
- ✅ memory query with boundaryFilter complete (arch-v1.md L489)
- ✅ tagOnWrite/getTaints API complete (arch-v1.md L490)
- ✅ 8 commits (one per test per TDD workflow)

---

## Phase 7.6: Service Discovery & Runtime Registration

### Goal
Implement service discovery and runtime registration integration for platform services.

### Scope
- Service discovery via well-known IDs (sys:*)
- Runtime registration in Kernel service map
- Integration with CommunicationService for mail routing
- Service health monitoring

### Current Status

#### Implementation Status
| Component | Status | Notes |
|-----------|--------|-------|
| `Service discovery` | ⚠️ Partial | Well-known IDs exist; needs discovery API |
| `Runtime registration` | ⚠️ Partial | Kernel has service map; needs auto-registration |
| `Mail routing` | ⚠️ Partial | CommunicationService exists; needs service routing |

### Tests to Write (4 tests, 4 commits)

#### Test 1: TestServiceDiscovery_WellKnownIDs
```go
func TestServiceDiscovery_WellKnownIDs(t *testing.T)
```
**Acceptance Criteria:**
- All platform services discoverable by sys:* ID format
- Hard-coded services: sys:security, sys:communication, sys:observability, sys:lifecycle
- Hot-reloadable services: sys:gateway, sys:admin, sys:persistence, sys:heartbeat, sys:memory, sys:human-gateway, sys:tools, sys:datasources

#### Test 2: TestServiceDiscovery_RuntimeRegistration
```go
func TestServiceDiscovery_RuntimeRegistration(t *testing.T)
```
**Acceptance Criteria:**
- Services auto-register with Kernel on spawn
- `GetServiceRuntimeID(name)` returns RuntimeID
- Service map includes all 12 platform services

#### Test 3: TestServiceDiscovery_MailRouting
```go
func TestServiceDiscovery_MailRouting(t *testing.T)
```
**Acceptance Criteria:**
- Mail to sys:* addresses routed to correct service
- CommunicationService dispatches to registered handlers
- handleMail contract invoked for each service

#### Test 4: TestServiceDiscovery_HealthMonitoring
```go
func TestServiceDiscovery_HealthMonitoring(t *testing.T)
```
**Acceptance Criteria:**
- Service health trackable via sys:observability
- Ready state reported for each service
- Failed services logged to dead-letter queue

### Dependencies

#### Test Dependencies
```
Test 1 → Test 2 (Discovery)
Test 3 (Mail routing)
Test 4 (Health monitoring)
```

### Files to Create/Modify

| File | Operation | Changes |
|------|-----------|---------|
| `pkg/kernel/kernel.go` | MODIFY | Add service discovery API |
| `pkg/services/communication/service.go` | MODIFY | Add sys:* mail routing |

### Implementation Order (TDD)

1. **Test 1**: Write test → verify RED → Implement well-known ID discovery → verify GREEN → commit
2. **Test 2**: Write test → verify RED → Implement auto-registration → verify GREEN → commit
3. **Test 3**: Write test → verify RED → Implement mail routing → verify GREEN → commit
4. **Test 4**: Write test → verify RED → Implement health monitoring → verify GREEN → commit

**Total: 4 tests, 4 commits**

### Deliverables
- ✅ Well-known ID discovery implemented
- ✅ Runtime registration implemented
- ✅ Mail routing to sys:* services implemented
- ✅ Health monitoring implemented
- ✅ 4 commits (one per test per TDD workflow)

---

## Phase 7.7: Integration Tests

### Goal
Verify all platform services integrate correctly via comprehensive integration tests.

### Scope
- End-to-end bootstrap sequence with all services
- Mail routing between all services
- Hot-reload end-to-end flow
- Service contract compliance

### Tests to Write (3 tests, 3 commits)

#### Test 1: TestIntegration_AllServicesBootstrap
```go
func TestIntegration_AllServicesBootstrap(t *testing.T)
```
**Acceptance Criteria:**
- All 12 platform services (4 hard-coded + 8 hot-reloadable) start during bootstrap
- KERNEL_READY emitted after all hard-coded services ready
- All hot-reloadable services loaded after KERNEL_READY

#### Test 2: TestIntegration_MailRoutingAllServices
```go
func TestIntegration_MailRoutingAllServices(t *testing.T)
```
**Acceptance Criteria:**
- Mail can be sent to any sys:* service
- Each service's handleMail invoked correctly
- Response mail routed back via CommunicationService

#### Test 3: TestIntegration_HotReloadEndToEnd
```go
func TestIntegration_HotReloadEndToEnd(t *testing.T)
```
**Acceptance Criteria:**
- Hot-reload protocol completes for hot-reloadable service
- Quiescence detected, history preserved, new runtime spawned
- Service continues handling mail after reload

### Dependencies

#### Test Dependencies
```
Test 1 (Bootstrap integration)
Test 2 (Mail routing integration)
Test 3 (Hot-reload integration)
```

### Files to Create/Modify

| File | Operation | Changes |
|------|-----------|---------|
| `pkg/services/integration/integration_test.go` | MODIFY | Add comprehensive integration tests |

### Implementation Order (TDD)

1. **Test 1**: Write test → verify RED → Wire all services → verify GREEN → commit
2. **Test 2**: Write test → verify RED → Verify mail routing → verify GREEN → commit
3. **Test 3**: Write test → verify RED → Verify hot-reload end-to-end → verify GREEN → commit

**Total: 3 tests, 3 commits**

### Deliverables
- ✅ All 12 services bootstrap correctly
- ✅ Mail routing works between all services
- ✅ Hot-reload end-to-end flow verified
- ✅ 3 commits (one per test per TDD workflow)

---

## Summary

### Phase Summary Table

| Phase | Tests | Files | Branch | Dependencies | Doc |
|-------|-------|-------|--------|--------------|-----|
| 7.1 | 6 | 2 | `feat/layer7-hardcoded-services` | None | [phase-7.1-hardcoded-services.md](phase-7.1-hardcoded-services.md) |
| 7.2 | 8 | 4 | `feat/layer7-hotreloadable` | 7.1 | [phase-7.2-hotreloadable-services.md](phase-7.2-hotreloadable-services.md) |
| 7.3 | 6 | 3 | `feat/layer7-kernel-ready-loading` | 7.1, 7.2 | [phase-7.3-kernel-ready-loading.md](phase-7.3-kernel-ready-loading.md) |
| 7.4 | 7 | 2 | `feat/layer7-hotreload-protocol` | 7.3 | [phase-7.4-hotreload-protocol.md](phase-7.4-hotreload-protocol.md) |
| 7.5 | 8 | 3 | `feat/layer7-service-contracts` | 7.2 | [phase-7.5-service-contracts.md](phase-7.5-service-contracts.md) |
| 7.6 | 4 | 2 | `feat/layer7-discovery` | 7.3, 7.5 | [phase-7.6-discovery.md](phase-7.6-discovery.md) |
| 7.7 | 3 | 1 | `feat/layer7-integration` | 7.4, 7.6 | [phase-7.7-integration.md](phase-7.7-integration.md) |

**Total: 42 tests, 17 files, 7 phases**

### Individual Phase Documents

Each phase has a detailed implementation plan:
- [Phase 7.1: Hard-Coded Services Completion](phase-7.1-hardcoded-services.md)
- [Phase 7.2: Hot-Reloadable Services Completion](phase-7.2-hotreloadable-services.md)
- [Phase 7.3: Post-KERNEL_READY Loading](phase-7.3-kernel-ready-loading.md)
- [Phase 7.4: Hot-Reload Protocol Integration](phase-7.4-hotreload-protocol.md)
- [Phase 7.5: Service-Specific Contracts](phase-7.5-service-contracts.md)
- [Phase 7.6: Service Discovery & Runtime Registration](phase-7.6-discovery.md)
- [Phase 7.7: Integration Tests](phase-7.7-integration.md)

### Execution Order

```
Phase 7.1 (Hard-Coded Services Completion)
     ↓
Phase 7.2 (Hot-Reloadable Services Completion)
     ↓                    ↘
     ↓              Phase 7.5 (Service Contracts)
Phase 7.3 (Post-KERNEL_READY Loading)
     ↓
Phase 7.4 (Hot-Reload Protocol)
     ↓
Phase 7.6 (Service Discovery) ──────┐
     ↓                              ↓
Phase 7.7 (Integration Tests) <─────┘
```

### Files to Create

| File | Phase | Purpose |
|------|-------|---------|
| `var/maelstrom/services/` | 7.3 | Directory for service YAML files |
| `pkg/runtime/runtime.go` | 7.4 | Add IsQuiescent() method |

### Files to Modify

| File | Phases | Changes |
|------|--------|---------|
| `pkg/services/security/service.go` | 7.1 | Add boundary enforcement |
| `pkg/services/communication/service.go` | 7.1, 7.6 | Add Ack, mail routing |
| `pkg/services/observability/service.go` | 7.1, 7.5 | Add dead-letter, query API |
| `pkg/services/lifecycle/service.go` | 7.1 | Add spawn tracking |
| `pkg/services/admin/service.go` | 7.2, 7.5 | Add 2FA, adminCommand API |
| `pkg/services/tools/service.go` | 7.2, 7.5 | Add ToolDescriptor, resolve API |
| `pkg/services/persistence/service.go` | 7.2, 7.5 | Add snapshot/restore |
| `pkg/services/heartbeat/service.go` | 7.2 | Add HEARTBEAT.md injection |
| `pkg/services/memory/service.go` | 7.2, 7.5 | Add ContextMap injection, query API |
| `pkg/services/datasources/service.go` | 7.2, 7.5 | Add tainting API |
| `pkg/kernel/kernel.go` | 7.3, 7.6 | Add ChartRegistry loading, discovery |
| `pkg/registry/service.go` | 7.3, 7.4 | Add platform service loading, hot-reload |

### Next Steps

1. **Start Phase 7.1**: Create branch `feat/layer7-hardcoded-services`
2. **Follow TDD workflow**: Write test → verify RED → Implement → verify GREEN → commit
3. **Repeat for each phase**: One branch per phase, merge after completion
4. **Run full test suite** after Phase 7.7 to verify integration

---

## Appendix A: Patterns Applied

This phase breakdown follows established patterns from Layer 2-3:
- ✅ **1:1 test-to-commit ratio** (42 tests = 42 commits)
- ✅ **One branch per phase** (7 branches total)
- ✅ **Descriptive test names** (`Test[Component]_[Behavior]_[ExpectedResult]`)
- ✅ **Small, focused phases** (3-8 tests per phase, ideal 4-6)
- ✅ **Explicit dependencies** between phases
- ✅ **Clear acceptance criteria** for each test
- ✅ **arch-v1.md line number references** for all requirements

## Appendix B: Gap Analysis Summary

| Gap | Resolution | Phase | Spec Reference |
|-----|------------|-------|----------------|
| Hard-coded services 10% incomplete | Complete boundary enforcement, tests | 7.1 | L453-461 |
| Hot-reloadable services 15% incomplete | Complete 2FA, ToolDescriptor, etc. | 7.2 | L463-474 |
| Post-KERNEL_READY loading missing | Implement ChartRegistry loading | 7.3 | L840-847 |
| Hot-reload protocol missing | Implement quiescence, reload flow | 7.4 | L865-882 |
| Service contracts incomplete | Complete adminCommand, resolve, etc. | 7.5 | L483-491 |
| Service discovery missing | Implement well-known ID discovery | 7.6 | L457-473 |
| Integration tests missing | Add end-to-end tests | 7.7 | All |

## Appendix C: Recursive Analysis

### Layer Dependencies
- **Layer 0 (Statechart Engine)**: Complete - provides runtime foundation
- **Layer 1 (Mail Types)**: Complete - provides Mail, Ack types
- **Layer 2 (Core Services)**: Complete - provides service infrastructure
- **Layer 3 (Communication)**: Complete - provides pub/sub
- **Layer 4 (Security)**: Complete - provides taint/boundary types
- **Layer 5 (Registry)**: In Progress - provides ChartRegistry
- **Layer 6 (Tools/Orchestration)**: In Progress - provides tool execution

### Layer 7 Completes
- Platform service contract (`handleMail`)
- Bootstrap sequence (KERNEL_READY)
- Hot-reload protocol
- Service discovery

### Layers 8+ Build On Layer 7
- Layer 8 (Streaming Gateway): Uses sys:gateway
- Layer 9 (Agent Charts): Uses sys:tools, sys:memory
- Layer 10 (LLM Integration): Uses sys:communication, sys:observability