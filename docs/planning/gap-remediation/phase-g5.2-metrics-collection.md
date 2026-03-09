# Phase G5.2: Metrics Collection

**Parent**: Phase G5 (Observability & Metrics)  
**Gap References**: L2-H3  
**Status**: ❌ PENDING

## Overview
- Tests: 4
- Commits: 4 (1:1 ratio)
- Dependencies: G5.1 (Mail Metadata Type Fix)

## SPEC

### Requirements
From `arch-v1.md 7.3` - Metrics collection:
- Track state transition rate
- Track event processing rate
- Get metrics collector
- Aggregate metrics over time

### Implementation Details

**Files to create/modify**:
- `pkg/services/observability/service.go` - Add rate tracking
- `pkg/services/observability/service_test.go` - Add tests

**Functions to implement**:
```go
func (o *ObservabilityService) trackTransition(from, to string)
func (o *ObservabilityService) trackEvent(eventType string)
func (o *ObservabilityService) GetMetrics() *MetricsCollector
func (o *ObservabilityService) aggregateMetrics(duration time.Duration) *MetricsCollector
```

**Test scenarios**:
1. Track state transition rate
2. Track event processing rate
3. Get metrics collector
4. Aggregate metrics over time

## TDD Workflow

### Iteration 1: TestObservabilityService_TransitionRate
1. Write test: `TestObservabilityService_TransitionRate` in `pkg/services/observability/service_test.go`
2. Run → RED
3. Implement: Add transition rate tracking
4. Run → GREEN
5. Commit: `git commit -m "fix(gap/L2-H3): implement transition rate tracking"`

### Iteration 2: TestObservabilityService_EventRate
1. Write test: `TestObservabilityService_EventRate`
2. Run → RED
3. Implement: Add event rate tracking
4. Run → GREEN
5. Commit: `git commit -m "fix(gap/L2-H3): implement event rate tracking"`

### Iteration 3: TestObservabilityService_MetricsCollector
1. Write test: `TestObservabilityService_MetricsCollector`
2. Run → RED
3. Implement: Add GetMetrics method
4. Run → GREEN
5. Commit: `git commit -m "fix(gap/L2-H3): implement metrics collector"`

### Iteration 4: TestObservabilityService_MetricsAggregation
1. Write test: `TestObservabilityService_MetricsAggregation`
2. Run → RED
3. Implement: Add aggregateMetrics method
4. Run → GREEN
5. Commit: `git commit -m "fix(gap/L2-H3): implement metrics aggregation"`

## Deliverables
- 4 commits
- All tests passing
- Files modified: `pkg/services/observability/service.go`, `pkg/services/observability/service_test.go`