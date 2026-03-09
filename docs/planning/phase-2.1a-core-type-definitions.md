# Phase 2.1A: Core Type Definitions

## Goal
Define core types required by all services without breaking changes, following strict TDD workflow.

## Scope
- Create `StreamChunk` type in mail package (new type, no migration yet)
- Add `TraceFilters` type for observability queries
- Add `MetricsCollector` type for observability metrics
- Add `RuntimeInfo` type for lifecycle listing
- 4 tests, 4 commits
- **NO breaking changes** - MailMetadata.Stream field remains unchanged

## Current Status

### Implementation Status
| Type | Status | Notes |
|------|--------|-------|
| `StreamChunk` | ❌ Missing in mail | Currently exists in `pkg/services/humangateway/chat.go` - DO NOT MOVE YET |
| `TraceFilters` | ❌ Missing | Needed for `QueryTraces()` |
| `MetricsCollector` | ❌ Missing | Needed for `GetMetrics()` |
| `RuntimeInfo` | ❌ Missing | Needed for `List()` |

### Files Status
| File | Status |
|------|--------|
| `pkg/mail/types.go` | ⚠️ Partial - add StreamChunk type only |
| `pkg/services/types.go` | ⚠️ Partial - add TraceFilters, MetricsCollector |
| `pkg/services/lifecycle/types.go` | ❌ MISSING - create with RuntimeInfo |

## Required Implementation

### StreamChunk (new type in mail package)
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

## Tests to Write (4 tests, 4 commits)

### Test 1: TestMail_StreamChunkHasRequiredFields
```go
func TestMail_StreamChunkHasRequiredFields(t *testing.T)
```
**Acceptance Criteria:**
- StreamChunk type exists in `pkg/mail/types.go`
- Has Data (string), Sequence (int), IsFinal (bool), Taints ([]string) fields
- Can instantiate with: `StreamChunk{Data: "test", Sequence: 42, IsFinal: true, Taints: []string{"test"}}`
- **StreamChunk in humangateway/chat.go remains unchanged**

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

## Dependencies

### Test Dependencies
```
Test 1 (StreamChunk) → Independent
Test 2 (TraceFilters) → Independent
Test 3 (MetricsCollector) → Independent
Test 4 (RuntimeInfo) → Independent
```

All tests are independent and can be executed in any order.

### Phase Dependencies
- **None** - This is the first sub-phase of Layer 2
- **Phase 2.1B** depends on this phase completing first

## Files to Create/Modify

| File | Operation | Changes |
|------|-----------|---------|
| `pkg/mail/types.go` | MODIFY | Add StreamChunk type definition |
| `pkg/services/types.go` | MODIFY | Add TraceFilters, MetricsCollector types |
| `pkg/services/lifecycle/types.go` | CREATE | Create file with RuntimeInfo type |

## Implementation Order (TDD)

1. **Test 1**: Write test → verify RED → Add StreamChunk to mail/types.go → verify GREEN → commit
2. **Test 2**: Write test → verify RED → Add TraceFilters to services/types.go → verify GREEN → commit
3. **Test 3**: Write test → verify RED → Add MetricsCollector to services/types.go → verify GREEN → commit
4. **Test 4**: Write test → verify RED → Create lifecycle/types.go with RuntimeInfo → verify GREEN → commit

**Total: 4 tests, 4 commits**

## Deliverables
- ✅ `StreamChunk` type in `pkg/mail/types.go`
- ✅ `TraceFilters` type in `pkg/services/types.go`
- ✅ `MetricsCollector` type in `pkg/services/types.go`
- ✅ `RuntimeInfo` type in `pkg/services/lifecycle/types.go`
- ✅ 4 commits (one per test per TDD workflow)
- ✅ NO breaking changes - MailMetadata.Stream field unchanged
- ✅ NO migration - StreamChunk still exists in humangateway/chat.go