# Phase 6.3: Orchestrator Service

## Goal
Implement OrchestratorService as unified orchestration service following strict TDD workflow with 6 tests.

## Scope
- Define OrchestratorService struct with registry, library, policies
- Implement RegisterPolicy() method
- Implement ResolveTool() method
- Implement Execute() method stub
- Create 6 tests in `pkg/orchestrator/orchestrator_test.go`

## Current Status

### Implementation Status
| Test | Status | Notes |
|------|--------|-------|
| Test 1: TestOrchestratorService_NewCreatesService | ❌ Missing | OrchestratorService does not exist |
| Test 2: TestOrchestratorService_RegisterPolicy | ❌ Missing | RegisterPolicy() does not exist |
| Test 3: TestOrchestratorService_ResolveTool | ❌ Missing | ResolveTool() does not exist |
| Test 4: TestOrchestratorService_ResolveToolWithBoundary | ❌ Missing | Boundary filtering does not exist |
| Test 5: TestOrchestratorService_ExecuteSpawnsChart | ❌ Missing | Execute() does not exist |
| Test 6: TestOrchestratorService_ExecuteWithPolicy | ❌ Missing | Policy parameter does not exist |

### Files Status
| File | Status |
|------|--------|
| `pkg/orchestrator/orchestrator.go` | ❌ Missing - must create with OrchestratorService struct |
| `pkg/orchestrator/orchestrator_test.go` | ❌ Missing - must create with 6 new tests |

## Required Implementation

### OrchestratorService Struct
```go
type OrchestratorService struct {
    id       string
    registry *ToolRegistry
    library  statechart.Library
    policies map[string]*ExecutionPolicy
    mu       sync.RWMutex
}
```

### Constructor and Methods
```go
func NewOrchestratorService() *OrchestratorService

func (s *OrchestratorService) ID() string

func (s *OrchestratorService) RegisterPolicy(name string, policy ExecutionPolicy) error

func (s *OrchestratorService) ResolveTool(name string, callerBoundary mail.BoundaryType) (ToolDescriptor, error)

func (s *OrchestratorService) Execute(tools []ToolCall, policy ExecutionPolicy) (statechart.RuntimeID, error)
```

### arch-v1.md References

**Tool & Action Orchestration** (arch-v1.md Section 8.1, lines 494-505):
- Tool calls and sub-agent calls are both "actions" that may require orchestration
- Single configurable Orchestrator sub-chart parameterized by execution mode, failure policy, completion signaling
- LLM response handed to orchestrator named in current state's binding
- Orchestrator instances run as transient parallel regions inside LLMReason compound state

**Orchestrator YAML Shape** (arch-v1.md Section 8.2, lines 506-524):
```yaml
orchestrator:
  default: seq_failfast          # or seq_continue, par_continue
  timeoutMs: 30000
  maxParallel: 8                  # only for par_*
  policies:
    seq_failfast:
      maxRetries: 2
      isolation: strict
    seq_continue:
      maxRetries: 1
      isolation: process
    par_continue:
      maxParallel: 8
      isolation: strict
```

**Tool Resolution** (arch-v1.md Section 8.4, lines 574-579):
```go
resolveTool(name string, callerBoundary BoundaryType) → ToolDescriptor | notFound
```
- Returns sanitized schema filtered by caller's boundary
- Inner tools are invisible to outer callers

## Tests to Write (6 tests, 6 commits)

### Test 1: TestOrchestratorService_NewCreatesService
```go
func TestOrchestratorService_NewCreatesService()
```
**Acceptance Criteria:**
- NewOrchestratorService() returns non-nil service
- Service has ID="sys:orchestrator" (arch-v1.md line 504)
- Service has empty policies map

### Test 2: TestOrchestratorService_RegisterPolicy
```go
func TestOrchestratorService_RegisterPolicy()
```
**Acceptance Criteria:**
- RegisterPolicy() stores policy by name
- Returns nil error on success
- Policy can be retrieved later

### Test 3: TestOrchestratorService_ResolveTool
```go
func TestOrchestratorService_ResolveTool()
```
**Acceptance Criteria:**
- ResolveTool() delegates to ToolRegistry (arch-v1.md line 574)
- Returns ToolDescriptor for valid tool
- Returns error for non-existent tool (arch-v1.md line 578)

### Test 4: TestOrchestratorService_ResolveToolWithBoundary
```go
func TestOrchestratorService_ResolveToolWithBoundary()
```
**Acceptance Criteria:**
- ResolveTool() filters by callerBoundary (arch-v1.md line 578)
- Inner tools invisible to outer callers (arch-v1.md line 578)

### Test 5: TestOrchestratorService_ExecuteSpawnsChart
```go
func TestOrchestratorService_ExecuteSpawnsChart()
```
**Acceptance Criteria:**
- Execute() spawns orchestrator chart (arch-v1.md line 504)
- Returns RuntimeID for spawned chart
- Chart runs as transient parallel region (arch-v1.md line 504)

### Test 6: TestOrchestratorService_ExecuteWithPolicy
```go
func TestOrchestratorService_ExecuteWithPolicy()
```
**Acceptance Criteria:**
- Execute() accepts ExecutionPolicy parameter
- Policy determines execution mode (seq_continue, seq_failfast, par_continue)

## Dependencies

### Test Dependencies
```
Test 1 → Test 2, Test 3, Test 4, Test 5, Test 6 (Service creation)
Test 2 → Test 6 (Policy registration)
Test 3 → Test 4 (Tool resolution)
Test 5 → Test 6 (Execution with policy)
```

### Phase Dependencies
- Phase 6.1 (Tool Descriptor Enhancement) - provides ToolDescriptor type
- Phase 6.2 (Execution Policies) - provides ExecutionPolicy type

## Files to Create/Modify

| File | Operation | Changes |
|------|-----------|---------|
| `pkg/orchestrator/orchestrator.go` | CREATE | Define OrchestratorService struct and methods |
| `pkg/orchestrator/orchestrator_test.go` | CREATE | Add 6 new tests |

## Implementation Order (TDD)

1. **Test 1**: Write test → verify RED → Implement NewOrchestratorService() → verify GREEN → commit
2. **Test 2**: Write test → verify RED → Implement RegisterPolicy() → verify GREEN → commit
3. **Test 3**: Write test → verify RED → Implement ResolveTool() basic → verify GREEN → commit
4. **Test 4**: Write test → verify RED → Add boundary filtering to ResolveTool() → verify GREEN → commit
5. **Test 5**: Write test → verify RED → Implement Execute() stub → verify GREEN → commit
6. **Test 6**: Write test → verify RED → Add policy parameter to Execute() → verify GREEN → commit

**Total: 6 tests, 6 commits**

## Deliverables
- Created `pkg/orchestrator/orchestrator.go` with OrchestratorService struct and methods
- Created `pkg/orchestrator/orchestrator_test.go` with 6 passing tests
- 6 commits (one per test per TDD workflow)