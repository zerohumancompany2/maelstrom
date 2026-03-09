# Phase 1.3.4: Lifecycle Service Stub

## Goal
Create stub implementation for the lifecycle service with 7 pass-through methods for chart spawn/stop functionality.

## Scope
- Create `pkg/services/lifecycle/service.go`
- Implement 7 stub methods
- Write 7 tests in `pkg/services/lifecycle/service_test.go`
- All methods return nil/empty (no actual spawn/stop functionality)

## Current Status

### File Status
| File | Status |
|------|--------|
| `pkg/services/lifecycle/bootstrap.go` | ✅ EXISTS - BootstrapChart() stub |
| `pkg/services/lifecycle/service.go` | ❌ **MISSING** - Must create |
| `pkg/services/lifecycle/service_test.go` | ✅ EXISTS - Has placeholder tests |

### Implementation Status
- **0/7 methods implemented**
- Service struct does not exist
- Only bootstrap.go exists

## Required Implementation

### Lifecycle Service Struct
```go
package lifecycle

type LifecycleService struct{}
```

### Constructor
```go
func NewLifecycleService() *LifecycleService {
    return &LifecycleService{}
}
```

### 7 Methods

| # | Method | Signature | Return Value | Purpose |
|---|--------|-----------|--------------|---------|
| 1 | `ID()` | `func (l *LifecycleService) ID() string` | `"sys:lifecycle"` | Returns service identifier |
| 2 | `HandleMail()` | `func (l *LifecycleService) HandleMail(mail mail.Mail) error` | `nil` | Processes incoming mail |
| 3 | `Spawn()` | `func (l *LifecycleService) Spawn(def statechart.ChartDefinition) (statechart.RuntimeID, error)` | `""`, `nil` | Spawns a new chart runtime |
| 4 | `Stop()` | `func (l *LifecycleService) Stop(id statechart.RuntimeID) error` | `nil` | Stops a running chart |
| 5 | `List()` | `func (l *LifecycleService) List() ([]string, error)` | `nil`, `nil` | Lists active chart runtimes |
| 6 | `Start()` | `func (l *LifecycleService) Start() error` | `nil` | Starts the service itself |

**Note:** The spec shows 6 methods total (not 7). The `Stop()` method serves dual purpose: stopping charts (method 4) and stopping the service itself (method 6 in original spec). For clarity, we'll implement 6 methods.

## Tests to Write (6 tests, 6 commits)

### Test 1: Constructor returns non-nil
```go
func TestLifecycleService_NewLifecycleServiceReturnsNonNil()
```
**Acceptance Criteria:**
- `NewLifecycleService()` returns a non-nil `*LifecycleService`

### Test 2: ID returns correct string
```go
func TestLifecycleService_IDReturnsCorrectString()
```
**Acceptance Criteria:**
- `ID()` returns exactly `"sys:lifecycle"`

### Test 3: HandleMail returns nil
```go
func TestLifecycleService_HandleMailReturnsNil()
```
**Acceptance Criteria:**
- `HandleMail(mail.Mail{})` returns `nil` error

### Test 4: Spawn returns empty RuntimeID
```go
func TestLifecycleService_SpawnReturnsEmptyRuntimeID()
```
**Acceptance Criteria:**
- `Spawn(def)` returns `""` (empty RuntimeID) and `nil` error

### Test 5: Stop returns nil
```go
func TestLifecycleService_StopReturnsNil()
```
**Acceptance Criteria:**
- `Stop(id)` returns `nil` error

### Test 6: List returns nil slice
```go
func TestLifecycleService_ListReturnsNilSlice()
```
**Acceptance Criteria:**
- `List()` returns `nil` slice and `nil` error

### Test 7: Start returns nil
```go
func TestLifecycleService_StartReturnsNil()
```
**Acceptance Criteria:**
- `Start()` returns `nil` error

## Dependencies

| Package | Types Used | Purpose |
|---------|------------|---------|
| `github.com/maelstrom/v3/pkg/statechart` | `RuntimeID`, `ChartDefinition` | Spawn method parameters and return values |
| `github.com/maelstrom/v3/pkg/mail` | `Mail` | HandleMail method parameter |

### Key Type Definitions

**From `pkg/statechart/types.go`:**
```go
// RuntimeID uniquely identifies a ChartRuntime instance within an Engine.
type RuntimeID string

// ChartDefinition is the immutable, hydrated definition of a statechart.
type ChartDefinition struct {
    ID           string
    Version      string
    Root         *Node
    InitialState string
    Spec         map[string]interface{}
}
```

**From `pkg/mail/types.go`:**
```go
type Mail struct {
    ID            string
    CorrelationID string
    Type          MailType
    CreatedAt     time.Time
    Source        string
    Target        string
    Content       any
    Metadata      MailMetadata
}
```

## Files to Create/Modify

| File | Operation | Changes |
|------|-----------|---------|
| `pkg/services/lifecycle/service.go` | CREATE | LifecycleService struct + 6 stub methods |
| `pkg/services/lifecycle/service_test.go` | MODIFY | Add 7 tests |

## Implementation Order (TDD)

1. **Test 1**: Write test → verify RED → Implement `NewLifecycleService()` → verify GREEN → commit
2. **Test 2**: Write test → verify RED → Implement `ID()` → verify GREEN → commit
3. **Test 3**: Write test → verify RED → Implement `HandleMail()` → verify GREEN → commit
4. **Test 4**: Write test → verify RED → Implement `Spawn()` → verify GREEN → commit
5. **Test 5**: Write test → verify RED → Implement `Stop()` → verify GREEN → commit
6. **Test 6**: Write test → verify RED → Implement `List()` → verify GREEN → commit
7. **Test 7**: Write test → verify RED → Implement `Start()` → verify GREEN → commit

**Total: 7 tests, 7 commits**

## Deliverables
- New `pkg/services/lifecycle/service.go` with 6 stub methods
- Updated `pkg/services/lifecycle/service_test.go` with 7 passing tests
- 7 commits (one per test per TDD workflow)