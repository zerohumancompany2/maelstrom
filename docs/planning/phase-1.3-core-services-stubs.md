# Phase 1.3: Core Services Stubs

## Goal
Create stub implementations for the 4 missing core services (security, communication, observability, lifecycle) with pass-through methods that return nil/empty.

## Scope
- Create 4 missing service files
- Implement 26 stub methods total
- All methods return nil/empty (no actual functionality)
- Create shared types file if needed
- Write tests for stub signatures

## Current Status

### Service Implementation Status

| # | Service | Status | Methods Implemented | Notes |
|---|---------|--------|---------------------|-------|
| 1 | **tools** | ✅ **COMPLETE** | 6/6 | Full pass-through implementation |
| 2 | **security** | ❌ **MISSING** | 0/5 | Only `bootstrap.go` exists; `service.go` missing |
| 3 | **communication** | ❌ **MISSING** | 0/7 | Only `bootstrap.go` exists; `service.go` missing |
| 4 | **observability** | ❌ **MISSING** | 0/7 | Only `bootstrap.go` exists; `service.go` missing |
| 5 | **lifecycle** | ❌ **MISSING** | 0/7 | Only `bootstrap.go` exists; `service.go` missing |
| 6 | **memory** | ✅ **COMPLETE** | 5/5 | Full pass-through implementation |
| 7 | **heartbeat** | ✅ **COMPLETE** | 4/4 | Full pass-through implementation |
| 8 | **persistence** | ✅ **COMPLETE** | 6/6 | Full pass-through implementation |
| 9 | **admin** | ✅ **COMPLETE** | 4/4 | Full pass-through implementation |
| 10 | **datasources** | ✅ **COMPLETE** | 6/6 | Full pass-through implementation |
| 11 | **gateway** | ✅ **COMPLETE** | 7/7 | Full implementation |
| 12 | **humangateway** | ✅ **COMPLETE** | 6/6 | Full pass-through implementation |

### Missing Files
- `pkg/services/security/service.go`
- `pkg/services/communication/service.go`
- `pkg/services/observability/service.go`
- `pkg/services/lifecycle/service.go`

### Shared Types Status
- `Mail` struct: ✅ Exists in `pkg/mail/types.go`
- `BoundaryType` enum: ✅ Exists in `pkg/mail/types.go`
- `Trace` struct: ❌ **MISSING** (needed for observability service)
- `pkg/services/types.go`: ❌ **MISSING**

## Required Services

### 1. Security Service (5 methods)
**File:** `pkg/services/security/service.go`

```go
type SecurityService struct{}

func NewSecurityService() *SecurityService

func (s *SecurityService) ID() string
// Return: "sys:security"

func (s *SecurityService) HandleMail(mail Mail) error
// Return: nil

func (s *SecurityService) Start() error
// Return: nil

func (s *SecurityService) Stop() error
// Return: nil
```

### 2. Communication Service (7 methods)
**File:** `pkg/services/communication/service.go`

```go
type CommunicationService struct{}

func NewCommunicationService() *CommunicationService

func (c *CommunicationService) ID() string
// Return: "sys:communication"

func (c *CommunicationService) HandleMail(mail Mail) error
// Return: nil

func (c *CommunicationService) Publish(mail Mail) error
// Return: nil

func (c *CommunicationService) Subscribe(address string) (<-chan Mail, error)
// Return: nil, nil

func (c *CommunicationService) Start() error
// Return: nil

func (c *CommunicationService) Stop() error
// Return: nil
```

### 3. Observability Service (7 methods)
**File:** `pkg/services/observability/service.go`

```go
type ObservabilityService struct{}

func NewObservabilityService() *ObservabilityService

func (o *ObservabilityService) ID() string
// Return: "sys:observability"

func (o *ObservabilityService) HandleMail(mail Mail) error
// Return: nil

func (o *ObservabilityService) EmitTrace(trace Trace) error
// Return: nil

func (o *ObservabilityService) QueryTraces(runtimeID string) ([]Trace, error)
// Return: nil, nil

func (o *ObservabilityService) Start() error
// Return: nil

func (o *ObservabilityService) Stop() error
// Return: nil
```

### 4. Lifecycle Service (7 methods)
**File:** `pkg/services/lifecycle/service.go`

```go
type LifecycleService struct{}

func NewLifecycleService() *LifecycleService

func (l *LifecycleService) ID() string
// Return: "sys:lifecycle"

func (l *LifecycleService) HandleMail(mail Mail) error
// Return: nil

func (l *LifecycleService) Spawn(def statechart.ChartDefinition) (statechart.RuntimeID, error)
// Return: "", nil

func (l *LifecycleService) Stop(id statechart.RuntimeID) error
// Return: nil

func (l *LifecycleService) List() ([]string, error)
// Return: nil, nil

func (l *LifecycleService) Start() error
// Return: nil
```

### 5. Shared Types
**File:** `pkg/services/types.go`

```go
type Trace struct {
    RuntimeID string
    Event     string
    Timestamp time.Time
    Data      map[string]interface{}
}
```

## Tests to Write

### Test Pattern (per service)
Each service needs basic signature tests:

```go
func TestSecurityService_NewSecurityServiceReturnsNonNil()
func TestSecurityService_IDReturnsCorrectString()
func TestSecurityService_HandleMailReturnsNil()
func TestSecurityService_StartReturnsNil()
func TestSecurityService_StopReturnsNil()
```

Repeat for all 4 services = **20 tests total**

## Files to Create

| File | Operations |
|------|------------|
| `pkg/services/security/service.go` | CREATE - 5 stub methods |
| `pkg/services/communication/service.go` | CREATE - 7 stub methods |
| `pkg/services/observability/service.go` | CREATE - 7 stub methods |
| `pkg/services/lifecycle/service.go` | CREATE - 7 stub methods |
| `pkg/services/types.go` | CREATE - Trace struct |
| `pkg/services/security/service_test.go` | MODIFY - Add 5 tests |
| `pkg/services/communication/service_test.go` | MODIFY - Add 7 tests |
| `pkg/services/observability/service_test.go` | MODIFY - Add 7 tests |
| `pkg/services/lifecycle/service_test.go` | MODIFY - Add 7 tests |

## Implementation Order (TDD by Service)

### Security Service (5 tests, 5 commits)
1. Test `NewSecurityService()` → implement → commit
2. Test `ID()` → implement → commit
3. Test `HandleMail()` → implement → commit
4. Test `Start()` → implement → commit
5. Test `Stop()` → implement → commit

### Communication Service (7 tests, 7 commits)
6-12. Repeat pattern for 7 methods

### Observability Service (7 tests, 7 commits)
13-19. Repeat pattern for 7 methods (requires Trace type first)

### Lifecycle Service (7 tests, 7 commits)
20-26. Repeat pattern for 7 methods

**Total: 26 tests, 26 commits**

## Dependencies
- Layer 0 (statechart.RuntimeID, statechart.ChartDefinition types)
- `pkg/mail/types.go` (Mail struct, BoundaryType enum)

## Deliverables
- 4 new service files with stub implementations
- 1 new shared types file
- 26 new tests across 4 test files
- 26 commits (one per test per TDD workflow)
- All 12 core services have stub implementations