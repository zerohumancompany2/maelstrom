# Phase G3.2: Service Bootstrap

**Parent**: Phase G3 (Integration & Wiring)  
**Gap References**: L2-M5  
**Status**: ❌ PENDING

## Overview
- Tests: 4
- Commits: 4 (1:1 ratio)
- Dependencies: G3.1 (Registry State Tracking)

## SPEC

### Requirements
From `docs/completed/layer-02-core-services.md 431-480` - Service bootstrap:
- Services start during bootstrap
- Services start in correct order
- Ready events emitted after bootstrap
- Handle service startup failures

### Implementation Details

**Files to create/modify**:
- `pkg/kernel/kernel.go` - Add bootstrap sequence
- `pkg/kernel/kernel_test.go` - Add tests

**Functions to implement**:
```go
func (k *Kernel) BootstrapServices() error
func (k *Kernel) startServiceInOrder(serviceID string) error
```

**Test scenarios**:
1. Services start during bootstrap
2. Services start in correct order
3. Ready events emitted after bootstrap
4. Handle service startup failures

## TDD Workflow

### Iteration 1: TestKernel_BootstrapServices
1. Write test: `TestKernel_BootstrapServices` in `pkg/kernel/kernel_test.go`
2. Run → RED
3. Implement: Add BootstrapServices method
4. Run → GREEN
5. Commit: `git commit -m "fix(gap/L2-M5): implement kernel service bootstrap"`

### Iteration 2: TestKernel_BootstrapSequence
1. Write test: `TestKernel_BootstrapSequence`
2. Run → RED
3. Implement: Add ordered service startup
4. Run → GREEN
5. Commit: `git commit -m "fix(gap/L2-M5): implement bootstrap sequence order"`

### Iteration 3: TestKernel_ServiceReadyEvents
1. Write test: `TestKernel_ServiceReadyEvents`
2. Run → RED
3. Implement: Add ready event emission
4. Run → GREEN
5. Commit: `git commit -m "fix(gap/L2-M5): implement service ready events"`

### Iteration 4: TestKernel_BootstrapFailure
1. Write test: `TestKernel_BootstrapFailure`
2. Run → RED
3. Implement: Add failure handling
4. Run → GREEN
5. Commit: `git commit -m "fix(gap/L2-M5): implement bootstrap failure handling"`

## Deliverables
- 4 commits
- All tests passing
- Files modified: `pkg/kernel/kernel.go`, `pkg/kernel/kernel_test.go`