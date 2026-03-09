# Phase G5.3: Dead-Letter Query Optimization

**Parent**: Phase G5 (Observability & Metrics)  
**Gap References**: L2-M2  
**Status**: ❌ PENDING

## Overview
- Tests: 3
- Commits: 3 (1:1 ratio)
- Dependencies: G3.3 (Dead-Letter Integration)

## SPEC

### Requirements
From `arch-v1.md 7.3` - Dead-letter query optimization:
- Query returns references not copies
- Memory usage within bounds
- Handle large dead-letter sets efficiently

### Implementation Details

**Files to create/modify**:
- `pkg/services/observability/service.go` - Optimize dead-letter query
- `pkg/services/observability/service_test.go` - Add tests

**Functions to implement**:
```go
func (o *ObservabilityService) QueryDeadLettersNoCopy(filters *DeadLetterFilters) []*DeadLetterEntry
func (o *ObservabilityService) getMemoryUsage() uint64
```

**Test scenarios**:
1. Query returns references not copies
2. Memory usage within bounds
3. Handle large dead-letter sets efficiently

## TDD Workflow

### Iteration 1: TestObservabilityService_QueryDeadLettersNoCopy
1. Write test: `TestObservabilityService_QueryDeadLettersNoCopy` in `pkg/services/observability/service_test.go`
2. Run → RED
3. Implement: Add QueryDeadLettersNoCopy method
4. Run → GREEN
5. Commit: `git commit -m "fix(gap/L2-M2): implement dead-letter query without copy"`

### Iteration 2: TestObservabilityService_QueryDeadLettersMemory
1. Write test: `TestObservabilityService_QueryDeadLettersMemory`
2. Run → RED
3. Implement: Add memory usage tracking
4. Run → GREEN
5. Commit: `git commit -m "fix(gap/L2-M2): implement dead-letter memory bounds"`

### Iteration 3: TestObservabilityService_QueryDeadLettersLargeSet
1. Write test: `TestObservabilityService_QueryDeadLettersLargeSet`
2. Run → RED
3. Implement: Optimize for large dead-letter sets
4. Run → GREEN
5. Commit: `git commit -m "fix(gap/L2-M2): implement dead-letter large set optimization"`

## Deliverables
- 3 commits
- All tests passing
- Files modified: `pkg/services/observability/service.go`, `pkg/services/observability/service_test.go`