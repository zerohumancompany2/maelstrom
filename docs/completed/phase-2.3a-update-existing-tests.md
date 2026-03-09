# Phase 2.3a: Update Existing Tests for Breaking Change

## Goal
Update 2 existing tests that call `QueryTraces(string)` to use new signature `QueryTraces(TraceFilters)`.

## Scope
- Update `TestObservabilityService_QueryTracesReturnsNilNil` to use TraceFilters
- Update `TestObservabilityService_EmitTrace` to use TraceFilters
- Change QueryTraces signature from `(runtimeID string)` to `(filters TraceFilters)`
- Add TraceFilters type to `pkg/services/types.go`
- 2 tests updated, 2 commits

## Current Status

### Implementation Status
| Test | Status | Notes |
|------|--------|-------|
| TestObservabilityService_QueryTracesReturnsNilNil | ⚠️ BREAKING | Calls QueryTraces("runtime-123") with string |
| TestObservabilityService_EmitTrace | ⚠️ BREAKING | Calls QueryTraces("test-runtime") with string |

### Files Status
| File | Status |
|------|--------|
| `pkg/services/types.go` | ⚠️ PARTIAL - needs TraceFilters type |
| `pkg/services/observability/service.go` | ⚠️ UPDATE - QueryTraces signature change |
| `pkg/services/observability/service_test.go` | ⚠️ UPDATE - 2 tests need signature update |

## Required Implementation

### TraceFilters Type
**File:** `pkg/services/types.go`

```go
type TraceFilters struct {
    RuntimeID string   // filter by runtime ID
    EventType string   // filter by event type
    FromTime  time.Time // filter traces after this time
    ToTime    time.Time // filter traces before this time
}
```

### Updated QueryTraces Method
**File:** `pkg/services/observability/service.go`

```go
func (o *ObservabilityService) QueryTraces(filters TraceFilters) ([]services.Trace, error) {
    // Filter by RuntimeID, EventType, FromTime, ToTime
}
```

## Tests to Update (2 tests, 2 commits)

### Test 1: TestObservabilityService_QueryTracesReturnsNilNil
**Current:**
```go
traces, err := svc.QueryTraces("runtime-123")
```

**Updated:**
```go
filters := services.TraceFilters{RuntimeID: "runtime-123"}
traces, err := svc.QueryTraces(filters)
```

**Acceptance Criteria:**
- QueryTraces() accepts TraceFilters instead of string
- Test still passes with same behavior

### Test 2: TestObservabilityService_EmitTrace
**Current:**
```go
traces, err := svc.QueryTraces("test-runtime")
```

**Updated:**
```go
filters := services.TraceFilters{RuntimeID: "test-runtime"}
traces, err := svc.QueryTraces(filters)
```

**Acceptance Criteria:**
- QueryTraces() accepts TraceFilters instead of string
- Test still passes with same behavior

## Dependencies

### Phase Dependencies
- **Phase 2.1** - Type definitions infrastructure exists

### Test Dependencies
```
Test 1 → Test 2 (both use TraceFilters type)
```

## Files to Create/Modify

| File | Operation | Changes |
|------|-----------|---------|
| `pkg/services/types.go` | MODIFY | Add TraceFilters struct |
| `pkg/services/observability/service.go` | MODIFY | Update QueryTraces() signature |
| `pkg/services/observability/service_test.go` | MODIFY | Update 2 tests to use TraceFilters |

## Implementation Order (TDD)

**Step 0: Add TraceFilters type (no test, just type definition)**
- Add TraceFilters struct to `pkg/services/types.go`
- Commit: `feat: add TraceFilters type for query filtering`

**Step 1-2: Update existing tests**
1. **Test 1**: Write updated test → verify RED (signature mismatch) → Update QueryTraces() signature → verify GREEN → commit
2. **Test 2**: Write updated test → verify RED → Update QueryTraces() implementation to use filters → verify GREEN → commit

**Total: 1 type commit + 2 test commits = 3 commits**

## Deliverables
- ✅ `TraceFilters` type in `pkg/services/types.go`
- ✅ `QueryTraces()` signature updated to use TraceFilters
- ✅ 2 existing tests updated and passing
- ✅ 3 commits (1 for type + one per test per TDD workflow)