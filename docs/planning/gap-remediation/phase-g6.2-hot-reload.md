# Phase G6.2: Hot-Reload

**Parent**: Phase G6 (Hot-Reload & Advanced)  
**Gap References**: L2-H4  
**Status**: ❌ PENDING

## Overview
- Tests: 4
- Commits: 4 (1:1 ratio)
- Dependencies: G3.2 (Service Bootstrap), G5.4 (Runtime Tracking)

## SPEC

### Requirements
From `arch-v1.md 7.2` - Hot-reload:
- Hot-reload running service
- State preserved during reload
- Handle reload failure
- Rollback on failure

### Implementation Details

**Files to create/modify**:
- `pkg/services/lifecycle/service.go` - Add hot-reload
- `pkg/services/lifecycle/service_test.go` - Add tests

**Functions to implement**:
```go
func (s *LifecycleService) HotReload(serviceID string) error
func (s *LifecycleService) preserveState(serviceID string) error
func (s *LifecycleService) rollbackReload(serviceID string) error
```

**Test scenarios**:
1. Hot-reload running service
2. State preserved during reload
3. Handle reload failure
4. Rollback on failure

## TDD Workflow

### Iteration 1: TestLifecycleService_HotReload
1. Write test: `TestLifecycleService_HotReload` in `pkg/services/lifecycle/service_test.go`
2. Run → RED
3. Implement: Add HotReload method
4. Run → GREEN
5. Commit: `git commit -m "fix(gap/L2-H4): implement hot-reload"`

### Iteration 2: TestLifecycleService_HotReloadStatePreservation
1. Write test: `TestLifecycleService_HotReloadStatePreservation`
2. Run → RED
3. Implement: Add state preservation during reload
4. Run → GREEN
5. Commit: `git commit -m "fix(gap/L2-H4): implement hot-reload state preservation"`

### Iteration 3: TestLifecycleService_HotReloadFailure
1. Write test: `TestLifecycleService_HotReloadFailure`
2. Run → RED
3. Implement: Add reload failure handling
4. Run → GREEN
5. Commit: `git commit -m "fix(gap/L2-H4): implement hot-reload failure handling"`

### Iteration 4: TestLifecycleService_HotReloadRollback
1. Write test: `TestLifecycleService_HotReloadRollback`
2. Run → RED
3. Implement: Add rollback on failure
4. Run → GREEN
5. Commit: `git commit -m "fix(gap/L2-H4): implement hot-reload rollback"`

## Deliverables
- 4 commits
- All tests passing
- Files modified: `pkg/services/lifecycle/service.go`, `pkg/services/lifecycle/service_test.go`