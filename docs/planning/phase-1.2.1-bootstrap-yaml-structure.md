# Phase 1.2.1: Bootstrap YAML Structure

## Goal
Update the bootstrap chart YAML to match the spec with correct state names, error handling, and action parameters.

## Scope
- Update `pkg/bootstrap/chart.go` YAML structure
- Add missing `failed` state
- Add error transitions
- Fix action names and parameters
- Create 5 tests in `pkg/bootstrap/chart_test.go`

## Current Status

### Existing YAML Structure
- ID: `sys:bootstrap`
- Initial state: `initializing`
- 7 states: `initializing`, `security`, `communication`, `observability`, `lifecycle`, `handoff`, `complete`
- Linear success path only (no error handling)

### Gaps from Spec
1. **No `failed` state** - Critical for error handling
2. **No error transitions** - Each service state needs failure path to `failed`
3. **Wrong action names** - Should use `*Bootstrap` naming convention
4. **Missing action parameters** - All entry actions need params per spec
5. **Wrong state naming** - Should use `sys:bootstrap/<state>` pattern
6. **Missing `ready` state** - Spec defines `ready` as final, not `complete`
7. **Missing `logSuccess`/`logFailure`/`panic` actions** - Required in final states

## Required Changes

### States to Add
- `sys:bootstrap/failed` - Error handling state with `panic` action
- `sys:bootstrap/ready` - Final success state with `logSuccess` action

### States to Rename
- `security` → `sys:bootstrap/security`
- `communication` → `sys:bootstrap/communication`
- `observability` → `sys:bootstrap/observability`
- `lifecycle` → `sys:bootstrap/lifecycle`
- `complete` → `sys:bootstrap/ready`

### Transitions to Add (Error Paths)
```yaml
sys:bootstrap/security:
  on:
    securityFailed:
      target: sys:bootstrap/failed

sys:bootstrap/communication:
  on:
    communicationFailed:
      target: sys:bootstrap/failed

sys:bootstrap/observability:
  on:
    observabilityFailed:
      target: sys:bootstrap/failed

sys:bootstrap/lifecycle:
  on:
    lifecycleFailed:
      target: sys:bootstrap/failed
```

### Actions to Fix
| Current Name | Required Name | Required Parameters |
|--------------|---------------|---------------------|
| `loadSecurityService` | `securityBootstrap` | `boundaries: [inner, dmz, outer]` |
| `loadCommunicationService` | `communicationBootstrap` | `mailBackbone: true` |
| `loadObservabilityService` | `observabilityBootstrap` | `tracing: true`, `metrics: true`, `deadLetterQueue: true` |
| `loadLifecycleService` | `lifecycleBootstrap` | `enableSpawn: true`, `enableStop: true`, `toolRegistry: true` |
| `signalKernelReady` | `emitKernelReady` | (none) |

### Actions to Add
- `logSuccess` - Final success state entry action
- `logFailure` - Failed state entry action  
- `panic` - Failed state panic action

## Tests to Write (5 tests)

### Test 1: YAML parses without error
```go
func TestBootstrapChartYAML_ParsesWithoutError()
```
**Acceptance Criteria:**
- `BootstrapChartYAML` constant can be parsed by statechart library
- No parsing errors or warnings
- Returns valid `ChartDefinition` struct

### Test 2: All required states exist
```go
func TestBootstrapChartYAML_HasRequiredStates()
```
**Acceptance Criteria:**
- Verify these states exist: `sys:bootstrap/security`, `sys:bootstrap/communication`, `sys:bootstrap/observability`, `sys:bootstrap/lifecycle`, `sys:bootstrap/ready`, `sys:bootstrap/failed`
- `sys:bootstrap/ready` is marked as final state

### Test 3: Transitions exist between states (success path)
```go
func TestBootstrapChartYAML_HasSuccessTransitions()
```
**Acceptance Criteria:**
- Verify transitions: `security→communication`, `communication→observability`, `observability→lifecycle`, `lifecycle→ready`
- Each transition has correct event name

### Test 4: Error transitions to failed state exist
```go
func TestBootstrapChartYAML_HasErrorTransitions()
```
**Acceptance Criteria:**
- Verify: `security` has `securityFailed` transition to `failed`
- Verify: `communication` has `communicationFailed` transition to `failed`
- Verify: `observability` has `observabilityFailed` transition to `failed`
- Verify: `lifecycle` has `lifecycleFailed` transition to `failed`

### Test 5: Final transition to ready state exists
```go
func TestBootstrapChartYAML_HasFinalReadyTransition()
```
**Acceptance Criteria:**
- Verify: `lifecycle→ready` transition exists
- Verify: Transition triggers `emitKernelReady` action

## Files to Modify

| File | Operation | Changes |
|------|-----------|---------|
| `pkg/bootstrap/chart.go` | MODIFY | Update `BootstrapChartYAML` constant with correct structure |
| `pkg/bootstrap/chart_test.go` | CREATE | Add 5 tests for Phase 1.2.1 |

## Implementation Order (TDD)

1. **Test 1**: Write test → verify RED → Update YAML to valid format → verify GREEN → commit
2. **Test 2**: Write test → verify RED → Add missing states → verify GREEN → commit
3. **Test 3**: Write test → verify RED → Add success transitions → verify GREEN → commit
4. **Test 4**: Write test → verify RED → Add error transitions → verify GREEN → commit
5. **Test 5**: Write test → verify RED → Add final transition with action → verify GREEN → commit

## Dependencies
- None (Phase 1.2.1 is independent)

## Deliverables
- Updated `pkg/bootstrap/chart.go` with correct YAML structure
- New `pkg/bootstrap/chart_test.go` with 5 passing tests
- 5 commits (one per test per TDD workflow)