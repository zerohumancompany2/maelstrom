# Phase G3.1: Service Registry State Tracking

**Parent**: Phase G3 (Integration & Wiring)  
**Gap References**: L2-H5  
**Status**: ❌ PENDING

## Overview
- Tests: 3
- Commits: 3 (1:1 ratio)
- Dependencies: G1, G2

## SPEC

### Requirements
From `docs/completed/layer-02-core-services.md 431-440` - Lifecycle state tracking:
- Register service with initial state
- Update service lifecycle state
- Query services by state

### Implementation Details

**Files to create/modify**:
- `pkg/services/registry.go` - Add state tracking
- `pkg/services/registry_test.go` - Add tests

**Functions to implement**:
```go
func (r *ServiceRegistry) RegisterWithState(service Service, initialState string) error
func (r *ServiceRegistry) UpdateState(serviceID string, newState string) error
func (r *ServiceRegistry) QueryByState(state string) []Service
```

**Test scenarios**:
1. Register service with initial state
2. Update service lifecycle state
3. Query services by state

## TDD Workflow

### Iteration 1: TestServiceRegistry_RegisterWithState
1. Write test: `TestServiceRegistry_RegisterWithState` in `pkg/services/registry_test.go`
2. Run → RED
3. Implement: Add RegisterWithState method
4. Run → GREEN
5. Commit: `git commit -m "fix(gap/L2-H5): implement service registration with state"`

### Iteration 2: TestServiceRegistry_UpdateState
1. Write test: `TestServiceRegistry_UpdateState`
2. Run → RED
3. Implement: Add UpdateState method
4. Run → GREEN
5. Commit: `git commit -m "fix(gap/L2-H5): implement service state update"`

### Iteration 3: TestServiceRegistry_QueryByState
1. Write test: `TestServiceRegistry_QueryByState`
2. Run → RED
3. Implement: Add QueryByState method
4. Run → GREEN
5. Commit: `git commit -m "fix(gap/L2-H5): implement service query by state"`

## Deliverables
- 3 commits
- All tests passing
- Files modified: `pkg/services/registry.go`, `pkg/services/registry_test.go`