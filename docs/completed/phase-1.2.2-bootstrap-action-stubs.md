# Phase 1.2.2: Bootstrap Action Stubs

## Goal
Create stub implementations for the 5 bootstrap actions that can be registered with the statechart library.

## Scope
- Create `pkg/bootstrap/actions.go` with 5 action stub functions
- Create `pkg/bootstrap/actions_test.go` with 4 tests
- Modify `pkg/kernel/kernel.go` to register actions
- Actions are stubs (log execution, validate params, return nil/empty)

## Current Status

### Action Implementation Status
| File | Status | Details |
|------|--------|---------|
| `pkg/bootstrap/actions.go` | ❌ **MISSING** | File does not exist |
| `pkg/bootstrap/sequence.go` | ⚠️ **PARTIAL** | Has state machine transitions but NO action implementations |
| `pkg/kernel/kernel.go` | ⚠️ **PARTIAL** | Has placeholder `onBootstrapStateEnter()` with simulated service loading, but no real action registration |

### Current Placeholder Behavior
- Lines 103-107: `security` state - logs and sends `SECURITY_READY` event (no real service spawn)
- Lines 109-112: `communication` state - logs and sends `COMMUNICATION_READY` event (no real service spawn)
- Lines 114-117: `observability` state - logs and sends `OBSERVABILITY_READY` event (no real service spawn)
- Lines 119-122: `lifecycle` state - logs and sends `LIFECYCLE_READY` event (no real service spawn)
- Lines 124-127: `handoff` state - logs and sends `KERNEL_READY` event

### Missing
- No action registration in kernel
- No action function implementations
- No parameter handling
- No error paths to `failed` state

## Required Actions (5 actions)

### 1. spawnSecurityService
```go
func spawnSecurityService(ctx context.Context, params map[string]interface{}) error
```
**Purpose:** Spawn sys:security service
**Parameters:** `boundaries: [inner, dmz, outer]`
**Behavior:** Log execution, validate params, return nil

### 2. spawnCommunicationService
```go
func spawnCommunicationService(ctx context.Context, params map[string]interface{}) error
```
**Purpose:** Spawn sys:communication service
**Parameters:** `mailBackbone: true`
**Behavior:** Log execution, validate params, return nil

### 3. spawnObservabilityService
```go
func spawnObservabilityService(ctx context.Context, params map[string]interface{}) error
```
**Purpose:** Spawn sys:observability service
**Parameters:** `tracing: true`, `metrics: true`, `deadLetterQueue: true`
**Behavior:** Log execution, validate params, return nil

### 4. spawnLifecycleService
```go
func spawnLifecycleService(ctx context.Context, params map[string]interface{}) error
```
**Purpose:** Spawn sys:lifecycle service
**Parameters:** `enableSpawn: true`, `enableStop: true`, `toolRegistry: true`
**Behavior:** Log execution, validate params, return nil

### 5. handleBootstrapFailure
```go
func handleBootstrapFailure(ctx context.Context, params map[string]interface{}) error
```
**Purpose:** Handle bootstrap failure path
**Parameters:** `error: string`
**Behavior:** Log error, return nil

## Tests to Write (4 tests)

### Test 1: All 5 actions can be registered
```go
func TestBootstrapActions_CanRegisterAllActions()
```
**Acceptance Criteria:**
- All 5 action functions can be registered with `statechart.Library.RegisterAction()`
- No registration errors
- Actions are accessible by name

### Test 2: Actions have correct signatures
```go
func TestBootstrapActions_HaveCorrectSignatures()
```
**Acceptance Criteria:**
- All actions accept `(context.Context, map[string]interface{})` as parameters
- All actions return `error`
- Action signatures match statechart action interface

### Test 3: Actions return error on invalid params
```go
func TestBootstrapActions_ReturnErrorOnInvalidParams()
```
**Acceptance Criteria:**
- `spawnSecurityService` returns error if `boundaries` param missing
- `spawnCommunicationService` returns error if `mailBackbone` param missing
- `spawnObservabilityService` returns error if required params missing
- `spawnLifecycleService` returns error if required params missing
- `handleBootstrapFailure` returns error if `error` param missing

### Test 4: Actions log execution
```go
func TestBootstrapActions_LogExecution()
```
**Acceptance Criteria:**
- Each action logs when it executes (use test logger or verify log output)
- Log includes action name and parameters
- Successful execution logs success message

## Files to Create/Modify

| File | Operation | Changes |
|------|-----------|---------|
| `pkg/bootstrap/actions.go` | CREATE | Implement 5 action stub functions with param validation and logging |
| `pkg/bootstrap/actions_test.go` | CREATE | Add 4 tests for Phase 1.2.2 |
| `pkg/kernel/kernel.go` | MODIFY | Add action registration in `Start()` method |

## Implementation Order (TDD)

1. **Test 1**: Write test → verify RED → Implement action functions → Register in kernel → verify GREEN → commit
2. **Test 2**: Write test → verify RED → Ensure correct signatures → verify GREEN → commit
3. **Test 3**: Write test → verify RED → Add parameter validation → verify GREEN → commit
4. **Test 4**: Write test → verify RED → Add logging to actions → verify GREEN → commit

## Dependencies
- Phase 1.2.1 (YAML structure must reference correct action names)
- Layer 0 (statechart.Library.RegisterAction() must be available)

## Deliverables
- New `pkg/bootstrap/actions.go` with 5 action stub implementations
- New `pkg/bootstrap/actions_test.go` with 4 passing tests
- Updated `pkg/kernel/kernel.go` with action registration
- 4 commits (one per test per TDD workflow)