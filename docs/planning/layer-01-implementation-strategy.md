# Layer 1 Implementation Strategy

**Created:** 2026-03-08
**Status:** Draft for discussion

## Overview

This document consolidates the implementation approach for Layer 1 (Kernel Bootstrap) based on the status report analysis.

## Phase 1.1: Kernel Completion

**Status:** Straightforward implementation
**Tests:** 14 tests
**Files to modify:**
- `pkg/kernel/kernel.go` - Add KernelConfig, Shutdown, action registration
- `pkg/kernel/kernel_test.go` - Add 14 tests

### Key Changes

1. **KernelConfig struct** (new):
   - BootstrapChartPath string
   - Registry *registry.Registry
   - Sources []source.Source
   - Services map[string]*statechart.Statechart

2. **Shutdown() method**:
   - Range over services map
   - Call Stop() on each
   - Wait for context cancellation

3. **Action Registration**:
   - Register bootstrap actions in New() or Start()
   - Actions: spawnSecurity, spawnCommunication, spawnObservability, spawnLifecycle

### Test Breakdown

**Config Tests (3):**
1. NewKernel with valid config
2. NewKernel with nil registry (error)
3. KernelConfig defaults

**Lifecycle Tests (4):**
4. Start() spawns bootstrap chart
5. Start() registers actions
6. Shutdown() stops all services
7. Shutdown() after Start() cleans up

**Service Management Tests (4):**
8. GetService() returns existing service
9. GetService() returns nil for unknown
10. RegisterService() adds to map
11. Service count after registration

**Integration Tests (3):**
12. Start → RegisterService → GetService chain
13. Start → Shutdown → no leak
14. Multiple services shutdown order

---

## Phase 1.2: Bootstrap Chart Completion

**Status:** Needs breakdown into sub-phases
**Tests:** 13 tests (currently)
**Issue:** Too many dependencies, should split

### Current Phase 1.2 Scope

1. Bootstrap YAML structure update
2. Failed state addition
3. Entry action parameters
4. Bootstrap action implementations
5. Bootstrap sequence integration
6. Error handling paths

### Proposed Breakdown

#### Phase 1.2.1: Bootstrap YAML Structure (5 tests)
**Focus:** Chart definition only, no actions

- Update YAML to match spec:
  - Add `failed` state
  - Add transitions from each service spawn to failed on error
  - Add entry/exit actions with proper parameter passing
- Tests:
  1. YAML parses without error
  2. All 5 states exist (initializing, spawning_security, spawning_communication, spawning_observability, spawning_lifecycle, failed, ready)
  3. Transitions exist between states
  4. Error transitions to failed state exist
  5. Final transition to ready state exists

#### Phase 1.2.2: Bootstrap Action Stubs (4 tests)
**Focus:** Action registration and signatures only

- Create `pkg/bootstrap/actions.go`:
  - spawnSecurityService(ctx, params)
  - spawnCommunicationService(ctx, params)
  - spawnObservabilityService(ctx, params)
  - spawnLifecycleService(ctx, params)
  - handleBootstrapFailure(ctx, params)
- Tests:
  6. All 5 actions can be registered
  7. Actions have correct signatures
  8. Actions return error on invalid params
  9. Actions log execution

#### Phase 1.2.3: Bootstrap Sequence Integration (4 tests)
**Focus:** Wiring chart to kernel

- Update `pkg/bootstrap/sequence.go`:
  - Pass kernel reference to actions
  - Actions call kernel.RegisterService()
  - Error handling calls handleBootstrapFailure
- Tests:
  10. Sequence passes kernel to actions
  11. Actions receive kernel reference
  12. Error path triggers failed state
  13. Success path triggers ready state

---

## Phase 1.3: Core Services Stubs

**Status:** Large but straightforward
**Tests:** 56 tests (4 per service × 12 services + 8 shared types)
**Files to create:** 12 service files + 1 types file

### Service List

1. `pkg/services/security/service.go` - Security service stub
2. `pkg/services/communication/service.go` - Communication service stub
3. `pkg/services/observability/service.go` - Observability service stub
4. `pkg/services/lifecycle/service.go` - Lifecycle service stub
5. `pkg/services/discovery/service.go` - Discovery service stub
6. `pkg/services/scheduling/service.go` - Scheduling service stub
7. `pkg/services/runtime/service.go` - Runtime service stub
8. `pkg/services/storage/service.go` - Storage service stub
9. `pkg/services/networking/service.go` - Networking service stub
10. `pkg/services/identity/service.go` - Identity service stub
11. `pkg/services/validation/service.go` - Validation service stub
12. `pkg/services/orchestration/service.go` - Orchestration service stub
13. `pkg/services/types.go` - Shared types (Mail, BoundaryType, Trace)

### Per-Service Pattern

Each service follows identical pattern:

```go
// service.go
package security

import (
    "context"
    "github.com/maelstrom/maelstrom/pkg/statechart"
)

type SecurityService struct {
    chart *statechart.Statechart
}

func New() *SecurityService { /* return stub */ }
func (s *SecurityService) Start(ctx context.Context) error { /* return nil */ }
func (s *SecurityService) Stop(ctx context.Context) error { /* return nil */ }
func (s *SecurityService) Handle(ctx context.Context, msg interface{}) error { /* return nil */ }
```

**Tests per service (4):**
1. New() returns non-nil
2. Start() returns nil
3. Stop() returns nil
4. Handle() returns nil

### Shared Types (8 tests)

`pkg/services/types.go`:
- Mail struct (From, To, Type, Payload, Timestamp)
- BoundaryType enum (Internal, External, Public)
- Trace struct (TraceID, SpanID, ParentSpanID, Service, Operation, Start, Duration)

Tests:
5. Mail struct fields correct
6. Mail zero value valid
7. BoundaryType values correct
8. Trace struct fields correct

---

## Phase 1.4: Bootstrap Actions (Real Implementation)

**Status:** Depends on Phase 1.3 completion
**Tests:** 12 tests
**Files to modify:**
- `pkg/bootstrap/actions.go` - Real implementations
- `pkg/bootstrap/actions_test.go` - Tests

### Actions to Implement

1. **spawnSecurityService**
   - Create SecurityService
   - Call kernel.RegisterService("security", service)
   - Start service
   - Return error if any step fails

2. **spawnCommunicationService**
   - Same pattern as security

3. **spawnObservabilityService**
   - Same pattern as security

4. **spawnLifecycleService**
   - Same pattern as security

5. **handleBootstrapFailure**
   - Log failure with params
   - Optionally trigger kernel shutdown
   - Return nil (always succeeds in logging)

### Test Breakdown

**Spawn Actions (8 tests):**
1. spawnSecurityService creates and registers service
2. spawnSecurityService starts service
3. spawnSecurityService returns error on creation failure
4. spawnCommunicationService creates and registers service
5. spawnObservabilityService creates and registers service
6. spawnLifecycleService creates and registers service
7. All 4 services can be spawned in sequence
8. Service retrieval after spawn works

**Error Handling (4 tests):**
9. handleBootstrapFailure logs error
10. handleBootstrapFailure includes params in log
11. handleBootstrapFailure returns nil
12. Bootstrap failure transitions to failed state

---

## Phase 1.5: Test Completion

**Status:** Depends on all previous phases
**Tests:** 11 placeholder tests to implement
**Files to modify:**
- Various test files with TODO placeholders

### Placeholder Tests to Implement

From status report analysis:
- Kernel: ~4 placeholder tests
- Bootstrap: ~3 placeholder tests
- Services: ~4 placeholder tests

**Approach:** Run tests after Phase 1.4, identify remaining placeholders, implement one at a time per TDD workflow.

---

## Phase 1.6: Integration

**Status:** Final phase, depends on all previous
**Tests:** 10 E2E tests
**Files to create:**
- `pkg/integration/e2e_test.go` or update existing

### E2E Test Scenarios

1. Full kernel start → kernel_ready event
2. All 4 core services spawned
3. Services respond to GetService()
4. Bootstrap failure path triggers failed state
5. Shutdown stops all services
6. No goroutine leaks after shutdown
7. Multiple start/stop cycles work
8. Service mail delivery (stub)
9. Trace propagation (stub)
10. Boundary enforcement (stub)

---

## Recommended Execution Order

1. **Phase 1.1** (14 tests, ~1 day)
   - Straightforward, no dependencies
   - Enables service registration for later phases

2. **Phase 1.2.1** (5 tests, ~0.5 day)
   - YAML structure only
   - No action implementation needed yet

3. **Phase 1.3** (56 tests, ~2 days)
   - Parallelizable (12 independent services)
   - Enables Phase 1.4

4. **Phase 1.2.2** (4 tests, ~0.5 day)
   - Action stubs
   - Depends on 1.2.1 for YAML

5. **Phase 1.2.3** (4 tests, ~0.5 day)
   - Wiring chart to kernel
   - Depends on 1.2.2 and 1.1

6. **Phase 1.4** (12 tests, ~1 day)
   - Real action implementations
   - Depends on 1.3 and 1.2.3

7. **Phase 1.5** (11 tests, ~1 day)
   - Fill in remaining placeholders
   - Depends on all previous

8. **Phase 1.6** (10 tests, ~1 day)
   - E2E validation
   - Final integration check

**Total:** ~8 days, 130 tests

---

## Risks & Mitigations

1. **Risk:** Phase 1.3 is large (56 tests)
   - **Mitigation:** Split into batches of 3-4 services each
   - **Mitigation:** Use code generation for boilerplate

2. **Risk:** Bootstrap chart YAML syntax errors
   - **Mitigation:** Validate YAML early in Phase 1.2.1
   - **Mitigation:** Use existing YAML as template

3. **Risk:** Action parameter passing unclear
   - **Mitigation:** Check Layer 0 statechart API docs
   - **Mitigation:** Start with simple params, iterate

4. **Risk:** Test placeholders not well-documented
   - **Mitigation:** Grep for "TODO" and "t.Skip" before Phase 1.5
   - **Mitigation:** Create checklist during Phase 1.5

---

## Next Steps

1. Review and approve phase breakdown
2. Start Phase 1.1 implementation
3. After each phase, update this document with actual results
4. Track blockers and adjust plan as needed