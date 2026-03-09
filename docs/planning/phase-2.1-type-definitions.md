# Phase 2.1: Type Definitions

## Goal
Define missing types required by all services before implementing functionality.

## Scope
- Move `StreamChunk` from `humangateway` to `mail` package
- Add `TraceFilters` type for observability queries
- Add `MetricsCollector` type for observability metrics
- Add `RuntimeInfo` type for lifecycle listing
- 6 tests, 6 commits

## Current Status

### Implementation Status
| Type | Status | Notes |
|------|--------|-------|
| `StreamChunk` | ⚠️ Exists in wrong location | In `pkg/services/humangateway/chat.go` |
| `TraceFilters` | ❌ Missing | Needed for `QueryTraces()` |
| `MetricsCollector` | ❌ Missing | Needed for `GetMetrics()` |
| `RuntimeInfo` | ❌ Missing | Needed for `List()` |

### Files Status
| File | Status |
|------|--------|
| `pkg/mail/types.go` | ⚠️ Partial - add StreamChunk |
| `pkg/services/types.go` | ⚠️ Partial - add TraceFilters, MetricsCollector |
| `pkg/services/lifecycle/types.go` | ❌ MISSING - create with RuntimeInfo |

## Required Implementation

### StreamChunk (move to mail package)
```go
// pkg/mail/types.go
type StreamChunk struct {
    Data     string
    Sequence int
    IsFinal  bool
    Taints   []string
}
```

### TraceFilters
```go
// pkg/services/types.go
type TraceFilters struct {
    RuntimeID string
    EventType string
    FromTime  time.Time
    ToTime    time.Time
}
```

### MetricsCollector
```go
// pkg/services/types.go
type MetricsCollector struct {
    StateCounts      map[string]int
    TransitionRate   float64
    EventRate        float64
    LastUpdate       time.Time
}
```

### RuntimeInfo
```go
// pkg/services/lifecycle/types.go
type RuntimeInfo struct {
    ID           string
    DefinitionID string
    Boundary     mail.BoundaryType
    ActiveStates []string
    IsRunning    bool
}
```

## Tests to Write (5 tests, 5 commits)

### Test 1: TestMail_StreamChunkHasRequiredFields
```go
func TestMail_StreamChunkHasRequiredFields(t *testing.T)
```
**Acceptance Criteria:**
- StreamChunk type exists in `pkg/mail/types.go`
- Has Data (string), Sequence (int), IsFinal (bool), Taints ([]string) fields
- Can instantiate with: `StreamChunk{Data: "test", Sequence: 42, IsFinal: true, Taints: []string{"test"}}`
- StreamChunk removed from `pkg/services/humangateway/chat.go`

### Test 2: TestServices_TraceFiltersHasRequiredFields
```go
func TestServices_TraceFiltersHasRequiredFields(t *testing.T)
```
**Acceptance Criteria:**
- TraceFilters type exists in `pkg/services/types.go`
- Has RuntimeID (string), EventType (string), FromTime (time.Time), ToTime (time.Time) fields
- Can instantiate with: `TraceFilters{RuntimeID: "test", EventType: "transition"}`

### Test 3: TestServices_MetricsCollectorHasRequiredFields
```go
func TestServices_MetricsCollectorHasRequiredFields(t *testing.T)
```
**Acceptance Criteria:**
- MetricsCollector type exists in `pkg/services/types.go`
- Has StateCounts (map[string]int), TransitionRate (float64), EventRate (float64) fields
- Can instantiate with: `MetricsCollector{StateCounts: map[string]int{"state1": 1}}`

### Test 4: TestLifecycle_RuntimeInfoHasRequiredFields
```go
func TestLifecycle_RuntimeInfoHasRequiredFields(t *testing.T)
```
**Acceptance Criteria:**
- RuntimeInfo type exists in `pkg/services/lifecycle/types.go`
- Has ID (string), DefinitionID (string), Boundary (mail.BoundaryType), ActiveStates ([]string), IsRunning (bool) fields
- Can instantiate with: `RuntimeInfo{ID: "test", DefinitionID: "def1", Boundary: mail.InnerBoundary}`

### Test 5: TestMail_MailMetadataStreamFieldUsesStreamChunk
```go
func TestMail_MailMetadataStreamFieldUsesStreamChunk(t *testing.T)
```
**Acceptance Criteria:**
- MailMetadata.Stream field type is changed from `bool` to `*StreamChunk`
- Can assign: `metadata.Stream = &StreamChunk{Data: "test"}`
- Can assign nil: `metadata.Stream = nil`

## Dependencies

### Test Dependencies
```
Test 1 → Test 5 (StreamChunk structure)
Test 2 → Test 6 (TraceFilters structure)
Test 3 → Test 6 (MetricsCollector structure)
Test 4 → Test 6 (RuntimeInfo structure)
```

### Phase Dependencies
- **None** - This is the first phase of Layer 2

## Files to Create/Modify

| File | Operation | Changes |
|------|-----------|---------|
| `pkg/mail/types.go` | MODIFY | Add StreamChunk type |
| `pkg/services/types.go` | MODIFY | Add TraceFilters, MetricsCollector |
| `pkg/services/lifecycle/types.go` | CREATE | Add RuntimeInfo type |

## Implementation Order (TDD)

1. **Test 1**: Write test → verify RED → Add StreamChunk to mail/types.go → verify GREEN → commit
2. **Test 2**: Write test → verify RED → Add TraceFilters to services/types.go → verify GREEN → commit
3. **Test 3**: Write test → verify RED → Add MetricsCollector to services/types.go → verify GREEN → commit
4. **Test 4**: Write test → verify RED → Create lifecycle/types.go with RuntimeInfo → verify GREEN → commit
5. **Test 5**: Write test → verify RED → Update MailMetadata.Stream field → verify GREEN → commit
6. **Test 6**: Write test → verify RED → Verify all types exported → verify GREEN → commit

**Total: 6 tests, 6 commits**

## Deliverables
- ✅ `StreamChunk` type in `pkg/mail/types.go`
- ✅ `TraceFilters` type in `pkg/services/types.go`
- ✅ `MetricsCollector` type in `pkg/services/types.go`
- ✅ `RuntimeInfo` type in `pkg/services/lifecycle/types.go`
- ✅ 6 commits (one per test per TDD workflow)