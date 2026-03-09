# Phase 2.3d: Metrics Collection

## Goal
Implement metrics collection with `GetMetrics()` method and metrics tracking on trace emission.

## Scope
- Add `MetricsCollector` type to `pkg/services/types.go`
- Add `MetricsCollector` field to ObservabilityService
- Implement `GetMetrics()` method
- Update metrics when traces are emitted via `EmitTrace()`
- 2 tests, 2 commits

## Current Status

### Implementation Status
| Test | Status | Notes |
|------|--------|-------|
| TestObservabilityService_GetMetrics | ❌ Missing | Method doesn't exist |
| TestObservabilityService_MetricsUpdateOnTrace | ❌ Missing | Metrics not tracked on EmitTrace |

### Files Status
| File | Status |
|------|--------|
| `pkg/services/types.go` | ⚠️ UPDATE - needs MetricsCollector type |
| `pkg/services/observability/service.go` | ⚠️ UPDATE - add metrics field, GetMetrics(), update EmitTrace() |
| `pkg/services/observability/service_test.go` | ⚠️ UPDATE - needs 2 new tests |

## Required Implementation

### MetricsCollector Type
**File:** `pkg/services/types.go`

```go
type MetricsCollector struct {
    StateCounts map[string]int  // count of traces by state path
    EventRate   float64         // events per second (calculated)
    LastUpdate  time.Time       // last time metrics were updated
}
```

### ObservabilityService with Metrics
**File:** `pkg/services/observability/service.go`

```go
type ObservabilityService struct {
    mu        sync.Mutex
    traces    []services.Trace
    deadLetters []DeadLetterEntry
    metrics   services.MetricsCollector
}

func NewObservabilityService() *ObservabilityService {
    return &ObservabilityService{
        metrics: services.MetricsCollector{
            StateCounts: make(map[string]int),
            LastUpdate:  time.Now(),
        },
    }
}
```

### GetMetrics Method
```go
func (o *ObservabilityService) GetMetrics() services.MetricsCollector {
    o.mu.Lock()
    defer o.mu.Unlock()
    // Return a copy to prevent external modification
    result := o.metrics
    result.StateCounts = make(map[string]int)
    for k, v := range o.metrics.StateCounts {
        result.StateCounts[k] = v
    }
    return result
}
```

### Updated EmitTrace Method
```go
func (o *ObservabilityService) EmitTrace(trace services.Trace) error {
    o.mu.Lock()
    defer o.mu.Unlock()
    o.traces = append(o.traces, trace)
    // Update metrics
    o.metrics.StateCounts[trace.StatePath]++
    o.metrics.LastUpdate = time.Now()
    return nil
}
```

## Tests to Write (2 tests, 2 commits)

### Test 1: TestObservabilityService_GetMetrics
```go
func TestObservabilityService_GetMetrics(t *testing.T)
```

**Acceptance Criteria:**
- GetMetrics() returns MetricsCollector
- MetricsCollector has StateCounts map (initialized, not nil)
- MetricsCollector has LastUpdate time set

### Test 2: TestObservabilityService_MetricsUpdateOnTrace
```go
func TestObservabilityService_MetricsUpdateOnTrace(t *testing.T)
```

**Acceptance Criteria:**
- EmitTrace() updates StateCounts for the trace's StatePath
- StatePath count increases by 1 after each emit
- LastUpdate time changes after EmitTrace()

## Dependencies

### Phase Dependencies
- **Phase 2.3a** - TraceFilters type exists (same file, no direct dependency)
- **Phase 2.3c** - DeadLetterEntry exists (no direct dependency, same phase group)

### Test Dependencies
```
Test 1 → Test 2 (GetMetrics must exist before testing updates)
```

## Files to Create/Modify

| File | Operation | Changes |
|------|-----------|---------|
| `pkg/services/types.go` | MODIFY | Add MetricsCollector struct |
| `pkg/services/observability/service.go` | MODIFY | Add metrics field, GetMetrics(), update EmitTrace() |
| `pkg/services/observability/service_test.go` | MODIFY | Add 2 new tests |

## Implementation Order (TDD)

**Step 0: Add MetricsCollector type (no test, just type definition)**
- Add MetricsCollector struct to `pkg/services/types.go`
- Commit: `feat: add MetricsCollector type for observability metrics`

**Step 1-2: Implement metrics functionality**
1. **Test 1**: Write test for GetMetrics → verify RED → Implement GetMetrics() and initialize metrics → verify GREEN → commit
2. **Test 2**: Write test for metrics update on trace → verify RED → Update EmitTrace() to update metrics → verify GREEN → commit

**Total: 1 type commit + 2 test commits = 3 commits**

## Deliverables
- ✅ `MetricsCollector` type in `pkg/services/types.go`
- ✅ `GetMetrics()` method implemented
- ✅ Metrics tracking on EmitTrace()
- ✅ StateCounts map tracking by StatePath
- ✅ LastUpdate timestamp on metric updates
- ✅ 3 commits (1 for type + one per test per TDD workflow)