# Phase G5.4: Runtime Tracking

**Parent**: Phase G5 (Observability & Metrics)  
**Gap References**: L2-M3  
**Status**: ❌ PENDING

## Overview
- Tests: 3
- Commits: 3 (1:1 ratio)
- Dependencies: G3.1 (Registry State Tracking)

## SPEC

### Requirements
From `arch-v1.md 7.3` - Runtime tracking:
- Update runtime state on transition
- List runtimes with current states
- Track state history

### Implementation Details

**Files to create/modify**:
- `pkg/services/lifecycle/service.go` - Add runtime tracking
- `pkg/services/lifecycle/service_test.go` - Add tests

**Functions to implement**:
```go
func (s *LifecycleService) updateRuntimeState(runtimeID, state string) error
func (s *LifecycleService) List() []RuntimeInfo
func (s *LifecycleService) getStateHistory(runtimeID string) []StateTransition
```

**Test scenarios**:
1. Update runtime state on transition
2. List runtimes with current states
3. Track state history

## TDD Workflow

### Iteration 1: TestLifecycleService_RuntimeStateUpdate
1. Write test: `TestLifecycleService_RuntimeStateUpdate` in `pkg/services/lifecycle/service_test.go`
2. Run → RED
3. Implement: Add updateRuntimeState method
4. Run → GREEN
5. Commit: `git commit -m "fix(gap/L2-M3): implement runtime state update"`

### Iteration 2: TestLifecycleService_ListWithStates
1. Write test: `TestLifecycleService_ListWithStates`
2. Run → RED
3. Implement: Add List method with current states
4. Run → GREEN
5. Commit: `git commit -m "fix(gap/L2-M3): implement list runtimes with states"`

### Iteration 3: TestLifecycleService_StateHistory
1. Write test: `TestLifecycleService_StateHistory`
2. Run → RED
3. Implement: Add state history tracking
4. Run → GREEN
5. Commit: `git commit -m "fix(gap/L2-M3): implement state history tracking"`

## Deliverables
- 3 commits
- All tests passing
- Files modified: `pkg/services/lifecycle/service.go`, `pkg/services/lifecycle/service_test.go`