# Phase 1.3.3: Observability Service Stub

## Goal
Create stub implementation for the observability service with 7 pass-through methods and required Trace type.

## Scope
- Create `pkg/services/types.go` with Trace struct
- Create `pkg/services/observability/service.go`
- Implement 7 stub methods
- Write 7 tests in `pkg/services/observability/service_test.go`
- All methods return nil/empty (no actual tracing functionality)

## Current Status

### File Status
| File | Status |
|------|--------|
| `pkg/services/observability/bootstrap.go` | ✅ EXISTS - BootstrapChart() stub |
| `pkg/services/observability/service.go` | ❌ **MISSING** - Must create |
| `pkg/services/observability/service_test.go` | ✅ EXISTS - Has placeholder tests |
| `pkg/services/types.go` | ❌ **MISSING** - Must create with Trace struct |

### Implementation Status
- **0/7 methods implemented**
- Service struct does not exist
- Trace type does not exist anywhere in codebase

## Required Implementation

### Trace Struct (Shared Type)
**File:** `pkg/services/types.go`

```go
package services

import "time"

type Trace struct {
    ID        string    // unique trace ID
    RuntimeID string    // originating ChartRuntime
    EventType string    // transition, entry, exit, event
    StatePath string    // state path at event time
    Timestamp time.Time // event time
    Payload   any       // event payload
}
```

### Observability Service Struct
**File:** `pkg/services/observability/service.go`

```go
package observability

import "github.com/maelstrom/v3/pkg/services"

type ObservabilityService struct{}
```

### Constructor
```go
func NewObservabilityService() *ObservabilityService {
    return &ObservabilityService{}
}
```

### 7 Methods

| # | Method | Signature | Return Value |
|---|--------|-----------|--------------|
| 1 | `ID()` | `func (o *ObservabilityService) ID() string` | `"sys:observability"` |
| 2 | `HandleMail()` | `func (o *ObservabilityService) HandleMail(mail mail.Mail) error` | `nil` |
| 3 | `EmitTrace()` | `func (o *ObservabilityService) EmitTrace(trace services.Trace) error` | `nil` |
| 4 | `QueryTraces()` | `func (o *ObservabilityService) QueryTraces(runtimeID string) ([]services.Trace, error)` | `nil, nil` |
| 5 | `Start()` | `func (o *ObservabilityService) Start() error` | `nil` |
| 6 | `Stop()` | `func (o *ObservabilityService) Stop() error` | `nil` |

## Tests to Write (7 tests, 7 commits)

### Test 1: Constructor returns non-nil
```go
func TestObservabilityService_NewObservabilityServiceReturnsNonNil()
```
**Acceptance Criteria:**
- `NewObservabilityService()` returns a non-nil `*ObservabilityService`

### Test 2: ID returns correct string
```go
func TestObservabilityService_IDReturnsCorrectString()
```
**Acceptance Criteria:**
- `ID()` returns exactly `"sys:observability"`

### Test 3: HandleMail returns nil
```go
func TestObservabilityService_HandleMailReturnsNil()
```
**Acceptance Criteria:**
- `HandleMail(mail)` returns `nil` error for any input

### Test 4: EmitTrace returns nil
```go
func TestObservabilityService_EmitTraceReturnsNil()
```
**Acceptance Criteria:**
- `EmitTrace(trace)` returns `nil` error for any input

### Test 5: QueryTraces returns nil slice and error
```go
func TestObservabilityService_QueryTracesReturnsNilNil()
```
**Acceptance Criteria:**
- `QueryTraces(runtimeID)` returns `(nil, nil)`

### Test 6: Start returns nil
```go
func TestObservabilityService_StartReturnsNil()
```
**Acceptance Criteria:**
- `Start()` returns `nil` error

### Test 7: Stop returns nil
```go
func TestObservabilityService_StopReturnsNil()
```
**Acceptance Criteria:**
- `Stop()` returns `nil` error

## Dependencies

| Package | Types Used | Purpose |
|---------|------------|---------|
| `github.com/maelstrom/v3/pkg/mail` | `Mail` | Mail struct for `HandleMail` method |
| `github.com/maelstrom/v3/pkg/services` | `Trace` | Trace struct for `EmitTrace` and `QueryTraces` methods |
| `time` | `Time` | Timestamp field in `Trace` struct |

## Files to Create/Modify

| File | Operation | Changes |
|------|-----------|---------|
| `pkg/services/types.go` | CREATE | Trace struct definition |
| `pkg/services/observability/service.go` | CREATE | ObservabilityService struct + 7 stub methods |
| `pkg/services/observability/service_test.go` | MODIFY | Add 7 tests |

## Implementation Order (TDD)

**Step 0: Create Trace type (no test, just type definition)**
- Create `pkg/services/types.go` with Trace struct
- Commit: `feat: add Trace type for observability`

**Step 1-7: Implement service methods**
1. **Test 1**: Write test → verify RED → Implement `NewObservabilityService()` → verify GREEN → commit
2. **Test 2**: Write test → verify RED → Implement `ID()` → verify GREEN → commit
3. **Test 3**: Write test → verify RED → Implement `HandleMail()` → verify GREEN → commit
4. **Test 4**: Write test → verify RED → Implement `EmitTrace()` → verify GREEN → commit
5. **Test 5**: Write test → verify RED → Implement `QueryTraces()` → verify GREEN → commit
6. **Test 6**: Write test → verify RED → Implement `Start()` → verify GREEN → commit
7. **Test 7**: Write test → verify RED → Implement `Stop()` → verify GREEN → commit

**Total: 1 type commit + 7 test commits = 8 commits**

## Deliverables
- New `pkg/services/types.go` with Trace struct
- New `pkg/services/observability/service.go` with 7 stub methods
- Updated `pkg/services/observability/service_test.go` with 7 passing tests
- 8 commits (1 for type + one per test per TDD workflow)