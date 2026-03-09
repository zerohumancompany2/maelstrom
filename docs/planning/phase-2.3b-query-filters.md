# Phase 2.3b: QueryTraces with Filters

## Goal
Implement full filtering logic in `QueryTraces()` using `TraceFilters` with RuntimeID, EventType, and time range filtering.

## Scope
- Implement RuntimeID filtering in QueryTraces()
- Implement EventType filtering in QueryTraces()
- Implement time range filtering (FromTime, ToTime) in QueryTraces()
- Handle empty filters (return all traces)
- 4 tests, 4 commits

## Current Status

### Implementation Status
| Test | Status | Notes |
|------|--------|-------|
| TestObservabilityService_QueryTracesWithFilters | ❌ Missing | RuntimeID filtering |
| TestObservabilityService_QueryTracesByEventType | ❌ Missing | EventType filtering |
| TestObservabilityService_QueryTracesByTimeRange | ❌ Missing | Time range filtering |
| TestObservabilityService_EmptyFiltersReturnsAll | ❌ Missing | Empty filters return all |

### Files Status
| File | Status |
|------|--------|
| `pkg/services/observability/service.go` | ⚠️ PARTIAL - QueryTraces signature updated, needs filtering logic |
| `pkg/services/observability/service_test.go` | ⚠️ UPDATE - needs 4 new tests |

## Required Implementation

### QueryTraces with Full Filtering
**File:** `pkg/services/observability/service.go`

```go
func (o *ObservabilityService) QueryTraces(filters services.TraceFilters) ([]services.Trace, error) {
    o.mu.Lock()
    defer o.mu.Unlock()
    var result []services.Trace
    for _, trace := range o.traces {
        // Filter by RuntimeID
        if filters.RuntimeID != "" && trace.RuntimeID != filters.RuntimeID {
            continue
        }
        // Filter by EventType
        if filters.EventType != "" && trace.EventType != filters.EventType {
            continue
        }
        // Filter by FromTime
        if !filters.FromTime.IsZero() && trace.Timestamp.Before(filters.FromTime) {
            continue
        }
        // Filter by ToTime
        if !filters.ToTime.IsZero() && trace.Timestamp.After(filters.ToTime) {
            continue
        }
        result = append(result, trace)
    }
    return result, nil
}
```

## Tests to Write (4 tests, 4 commits)

### Test 1: TestObservabilityService_QueryTracesWithFilters
```go
func TestObservabilityService_QueryTracesWithFilters(t *testing.T)
```

**Acceptance Criteria:**
- QueryTraces() filters by RuntimeID correctly
- Only traces matching RuntimeID are returned
- Multiple traces with same RuntimeID all returned

### Test 2: TestObservabilityService_QueryTracesByEventType
```go
func TestObservabilityService_QueryTracesByEventType(t *testing.T)
```

**Acceptance Criteria:**
- QueryTraces() filters by EventType
- Only traces matching EventType are returned
- Empty EventType returns all events

### Test 3: TestObservabilityService_QueryTracesByTimeRange
```go
func TestObservabilityService_QueryTracesByTimeRange(t *testing.T)
```

**Acceptance Criteria:**
- QueryTraces() filters by FromTime and ToTime
- Returns traces within time range (inclusive)
- Traces before FromTime are excluded
- Traces after ToTime are excluded

### Test 4: TestObservabilityService_EmptyFiltersReturnsAll
```go
func TestObservabilityService_EmptyFiltersReturnsAll(t *testing.T)
```

**Acceptance Criteria:**
- Empty TraceFilters (all zero values) returns all traces
- No filtering applied when filters are empty
- Returns complete trace list

## Dependencies

### Phase Dependencies
- **Phase 2.3a** - TraceFilters type and QueryTraces signature updated

### Test Dependencies
```
Test 1 → Test 2, Test 3, Test 4 (all build on QueryTraces implementation)
```

## Files to Create/Modify

| File | Operation | Changes |
|------|-----------|---------|
| `pkg/services/observability/service.go` | MODIFY | Implement full filtering logic in QueryTraces() |
| `pkg/services/observability/service_test.go` | MODIFY | Add 4 new tests |

## Implementation Order (TDD)

1. **Test 1**: Write test for RuntimeID filtering → verify RED → Implement RuntimeID filter → verify GREEN → commit
2. **Test 2**: Write test for EventType filtering → verify RED → Implement EventType filter → verify GREEN → commit
3. **Test 3**: Write test for time range filtering → verify RED → Implement FromTime/ToTime filters → verify GREEN → commit
4. **Test 4**: Write test for empty filters → verify RED → Ensure empty filter handling → verify GREEN → commit

**Total: 4 tests, 4 commits**

## Deliverables
- ✅ RuntimeID filtering implemented in QueryTraces()
- ✅ EventType filtering implemented in QueryTraces()
- ✅ Time range filtering (FromTime/ToTime) implemented in QueryTraces()
- ✅ Empty filters return all traces
- ✅ 4 commits (one per test per TDD workflow)