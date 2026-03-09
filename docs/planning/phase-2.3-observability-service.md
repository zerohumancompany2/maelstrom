# Phase 2.3: Observability Service Enhancement

## Goal
Enhance Observability Service with trace filtering, dead-letter queue, and metrics collection.

## Scope
- Update `QueryTraces()` to use `TraceFilters`
- Implement `LogDeadLetter()` method
- Implement `GetMetrics()` method
- Add dead-letter storage
- 9 tests, 9 commits

## Current Status

### Implementation Status
| Test | Status | Notes |
|------|--------|-------|
| TestObservabilityService_QueryTracesWithFilters | ❌ Missing | Uses string instead of TraceFilters |
| TestObservabilityService_LogDeadLetter | ❌ Missing | Method doesn't exist |
| TestObservabilityService_GetMetrics | ❌ Missing | Method doesn't exist |
| TestObservabilityService_QueryTracesByEventType | ❌ Missing | No event type filtering |
| TestObservabilityService_QueryTracesByTimeRange | ❌ Missing | No time range filtering |

### Files Status
| File | Status |
|------|--------|
| `pkg/services/observability/trace.go` | ❌ MISSING - must create |
| `pkg/services/observability/service.go` | ⚠️ Partial - update QueryTraces() signature |

## Required Implementation

### DeadLetterEntry
```go
// pkg/services/observability/trace.go
type DeadLetterEntry struct {
    Mail   mail.Mail
    Reason string
    Logged time.Time
}
```

### Updated QueryTraces Method
```go
// pkg/services/observability/service.go
func (o *ObservabilityService) QueryTraces(filters TraceFilters) ([]Trace, error) {
    // Filter by RuntimeID, EventType, FromTime, ToTime
}
```

### LogDeadLetter Method
```go
// pkg/services/observability/service.go
func (o *ObservabilityService) LogDeadLetter(mail mail.Mail, reason string) error {
    // Store in dead-letter queue
}
```

### GetMetrics Method
```go
// pkg/services/observability/service.go
func (o *ObservabilityService) GetMetrics() MetricsCollector {
    // Return current metrics
}
```

## Tests to Write (9 tests, 9 commits)

### Test 1: TestObservabilityService_QueryTracesWithFilters
```go
func TestObservabilityService_QueryTracesWithFilters(t *testing.T)
```
**Acceptance Criteria:**
- QueryTraces() accepts TraceFilters
- Filters by RuntimeID correctly

### Test 2: TestObservabilityService_QueryTracesByEventType
```go
func TestObservabilityService_QueryTracesByEventType(t *testing.T)
```
**Acceptance Criteria:**
- QueryTraces() filters by EventType
- Empty EventType returns all events

### Test 3: TestObservabilityService_QueryTracesByTimeRange
```go
func TestObservabilityService_QueryTracesByTimeRange(t *testing.T)
```
**Acceptance Criteria:**
- QueryTraces() filters by FromTime and ToTime
- Returns traces within time range

### Test 4: TestObservabilityService_LogDeadLetter
```go
func TestObservabilityService_LogDeadLetter(t *testing.T)
```
**Acceptance Criteria:**
- LogDeadLetter() stores mail and reason
- Logged timestamp is set

### Test 5: TestObservabilityService_QueryDeadLetters
```go
func TestObservabilityService_QueryDeadLetters(t *testing.T)
```
**Acceptance Criteria:**
- Can query dead-letter entries
- Returns all logged entries

### Test 6: TestObservabilityService_GetMetrics
```go
func TestObservabilityService_GetMetrics(t *testing.T)
```
**Acceptance Criteria:**
- GetMetrics() returns MetricsCollector
- MetricsCollector has StateCounts map

### Test 7: TestObservabilityService_MetricsUpdateOnTrace
```go
func TestObservabilityService_MetricsUpdateOnTrace(t *testing.T)
```
**Acceptance Criteria:**
- EmitTrace() updates metrics
- EventRate increases after emit

### Test 8: TestObservabilityService_DeadLetterWithReason
```go
func TestObservabilityService_DeadLetterWithReason(t *testing.T)
```
**Acceptance Criteria:**
- DeadLetterEntry preserves reason
- Reason is non-empty string

### Test 9: TestObservabilityService_EmptyFiltersReturnsAll
```go
func TestObservabilityService_EmptyFiltersReturnsAll(t *testing.T)
```
**Acceptance Criteria:**
- Empty TraceFilters returns all traces
- No filtering applied

## Dependencies

### Test Dependencies
```
Test 1 → Test 2, Test 3, Test 9 (QueryTraces with filters)
Test 4 → Test 5, Test 8 (Dead-letter functionality)
Test 6 → Test 7 (Metrics functionality)
```

### Phase Dependencies
- **Phase 2.1** - Type definitions (TraceFilters, MetricsCollector must exist)

## Files to Create/Modify

| File | Operation | Changes |
|------|-----------|---------|
| `pkg/services/observability/trace.go` | CREATE | DeadLetterEntry type |
| `pkg/services/observability/service.go` | MODIFY | Update QueryTraces(), add LogDeadLetter(), GetMetrics() |

## Implementation Order (TDD)

1. **Test 1**: Write test → verify RED → Update QueryTraces() to use TraceFilters → verify GREEN → commit
2. **Test 2**: Write test → verify RED → Add EventType filtering → verify GREEN → commit
3. **Test 3**: Write test → verify RED → Add time range filtering → verify GREEN → commit
4. **Test 4**: Write test → verify RED → Implement LogDeadLetter() → verify GREEN → commit
5. **Test 5**: Write test → verify RED → Add QueryDeadLetters() → verify GREEN → commit
6. **Test 6**: Write test → verify RED → Implement GetMetrics() stub → verify GREEN → commit
7. **Test 7**: Write test → verify RED → Update metrics on EmitTrace() → verify GREEN → commit
8. **Test 8**: Write test → verify RED → Ensure reason preserved in dead-letter → verify GREEN → commit
9. **Test 9**: Write test → verify RED → Handle empty filters → verify GREEN → commit

**Total: 9 tests, 9 commits**

## Deliverables
- ✅ `DeadLetterEntry` type in `trace.go`
- ✅ `QueryTraces()` uses TraceFilters
- ✅ `LogDeadLetter()` implemented
- ✅ `GetMetrics()` implemented
- ✅ 9 commits (one per test per TDD workflow)