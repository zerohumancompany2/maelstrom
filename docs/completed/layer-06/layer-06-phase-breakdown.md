# Layer 6: Tool Calls & Orchestration - Phase Breakdown

## Executive Summary

Layer 6 implements tool calls and orchestration, providing a unified abstraction for executing tools and sub-agents with configurable policies. Based on analysis of the Layer 6 specification, current codebase state, and Layer 5 dependencies, this document breaks down Layer 6 into **7 phases** with **~32 tests** across **~32 commits**.

### Current State (Layer 0-5 Pending)
- ✅ Basic tool registry exists in `pkg/services/tools/service.go` (146 lines, 17 tests)
- ✅ Basic orchestrator isolation exists in `pkg/orchestrator/isolation.go` (84 lines, 2 tests)
- ⚠️ ToolDescriptor missing fields: `TaintOutput`, `Type`, `ChartRef`, `MaxIterations`
- ⚠️ Execution policies not implemented (seq_continue, seq_failfast, par_continue)
- ⚠️ OrchestratorService not implemented
- ⚠️ Sequential and Parallel executors not implemented
- ⚠️ Sub-agent lifecycle management not implemented
- ⚠️ Tool schema sanitization not implemented

### Layer 6 Goal
Complete Section 8 behavior per arch-v1.md:
1. **Tool Registry Enhancement** - Add missing fields and schema sanitization (arch-v1.md 8.4)
2. **Execution Policies** - Implement seq_continue, seq_failfast, par_continue (arch-v1.md 8.2)
3. **Orchestrator Service** - Implement unified orchestration service (arch-v1.md 8.1)
4. **Sequential Executor** - Implement sequential tool execution (arch-v1.md 8.3)
5. **Parallel Executor** - Implement parallel tool execution (arch-v1.md 8.3)
6. **Sub-Agent Support** - Implement attached/detached sub-agents (arch-v1.md 6.4)
7. **Integration** - Wire all components together

---

## Phase Overview

| Phase | Name | Tests | Files | Branch | Dependencies |
|-------|------|-------|-------|--------|--------------|
| 6.1 | Tool Descriptor Enhancement | 4 | 2 | `feat/layer6-tool-descriptor` | None |
| 6.2 | Execution Policies | 5 | 1 | `feat/layer6-execution-policies` | 6.1 |
| 6.3 | Orchestrator Service | 6 | 2 | `feat/layer6-orchestrator-service` | 6.1, 6.2 |
| 6.4 | Sequential Executor | 5 | 1 | `feat/layer6-sequential-executor` | 6.3 |
| 6.5 | Parallel Executor | 5 | 1 | `feat/layer6-parallel-executor` | 6.3 |
| 6.6 | Sub-Agent Support | 5 | 2 | `feat/layer6-subagent` | 6.3, 6.4, 6.5 |
| 6.7 | Integration & Schema Sanitization | 2 | 2 | `feat/layer6-integration` | 6.1-6.6 |

**Total: 32 tests, 11 files, 7 phases**

---

## Phase 6.1: Tool Descriptor Enhancement

### Goal
Enhance ToolDescriptor with missing fields per arch-v1.md 8.4 and add schema sanitization.

### Scope
- Add `TaintOutput` field to ToolDescriptor (arch-v1.md 563, 566)
- Add `Type` field for sub-agent identification (arch-v1.md 568)
- Add `ChartRef` field for sub-agent chart reference (arch-v1.md 569)
- Add `MaxIterations` field for sub-agent limits (arch-v1.md 569)
- Implement `SanitizeSchema()` method (arch-v1.md 578)

### Current Status

#### Implementation Status
| Field | Status | Notes |
|-------|--------|-------|
| `Name` | ✅ Exists | In `pkg/services/tools/service.go:30` |
| `Boundary` | ✅ Exists | In `pkg/services/tools/service.go:31` |
| `Schema` | ✅ Exists | In `pkg/services/tools/service.go:32` |
| `Isolation` | ✅ Exists | In `pkg/services/tools/service.go:33` |
| `TaintOutput` | ❌ Missing | Required per arch-v1.md 563, 566 |
| `Type` | ❌ Missing | Required per arch-v1.md 568 |
| `ChartRef` | ❌ Missing | Required per arch-v1.md 569 |
| `MaxIterations` | ❌ Missing | Required per arch-v1.md 569 |

#### Files Status
| File | Status |
|------|--------|
| `pkg/services/tools/service.go` | ⚠️ Partial - update ToolDescriptor |
| `pkg/orchestrator/types.go` | ❌ MISSING - create with enhanced types |

### Required Implementation

#### Enhanced ToolDescriptor
```go
// pkg/orchestrator/types.go
type ToolDescriptor struct {
    Name          string
    Boundary      mail.BoundaryType
    Schema        JSONSchema
    Isolation     string  // container, process, sandbox, strict
    TaintOutput   []string
    Type          string  // "tool" or "attachedSubAgent"
    ChartRef      string  // for sub-agents
    MaxIterations int     // for sub-agents
}
```

#### SanitizeSchema Method
```go
// pkg/services/tools/service.go
func (s *toolsService) SanitizeSchema(tool ToolDescriptor, callerBoundary mail.BoundaryType) JSONSchema {
    // Return sanitized schema filtered by caller's boundary (arch-v1.md 578)
    // Inner tools are invisible to outer callers
}
```

### Tests to Write (4 tests, 4 commits)

#### Test 1: TestToolDescriptor_HasTaintOutputField
```go
func TestToolDescriptor_HasTaintOutputField(t *testing.T)
```
**Acceptance Criteria:**
- ToolDescriptor has TaintOutput field (arch-v1.md 563, 566)
- TaintOutput is []string type
- Can be set to ["TOOL_OUTPUT"] or ["TOOL_OUTPUT", "INNER_ONLY"]

#### Test 2: TestToolDescriptor_HasSubAgentFields
```go
func TestToolDescriptor_HasSubAgentFields(t *testing.T)
```
**Acceptance Criteria:**
- ToolDescriptor has Type field (arch-v1.md 568)
- ToolDescriptor has ChartRef field (arch-v1.md 569)
- ToolDescriptor has MaxIterations field (arch-v1.md 569)
- Type can be "attachedSubAgent"

#### Test 3: TestToolDescriptor_SubAgentConfiguration
```go
func TestToolDescriptor_SubAgentConfiguration(t *testing.T)
```
**Acceptance Criteria:**
- Sub-agent tool has Type="attachedSubAgent"
- Sub-agent tool has ChartRef pointing to valid chart
- Sub-agent tool has MaxIterations > 0

#### Test 4: TestToolDescriptor_AllFieldsCoexist
```go
func TestToolDescriptor_AllFieldsCoexist(t *testing.T)
```
**Acceptance Criteria:**
- All 8 fields exist on ToolDescriptor
- Fields can be set independently
- No conflicts with existing fields

### Dependencies

#### Test Dependencies
```
Test 1 → Test 4 (TaintOutput structure)
Test 2 → Test 3, Test 4 (Sub-agent fields)
Test 3 → Test 4 (Sub-agent configuration)
```

### Files to Create/Modify

| File | Operation | Changes |
|------|-----------|---------|
| `pkg/services/tools/service.go` | MODIFY | Update ToolDescriptor with 4 new fields |
| `pkg/orchestrator/types.go` | CREATE | Define enhanced ToolDescriptor and related types |

### Implementation Order (TDD)

1. **Test 1**: Write test → verify RED → Add TaintOutput field → verify GREEN → commit
2. **Test 2**: Write test → verify RED → Add Type, ChartRef, MaxIterations fields → verify GREEN → commit
3. **Test 3**: Write test → verify RED → Validate sub-agent configuration → verify GREEN → commit
4. **Test 4**: Write test → verify RED → Verify all fields coexist → verify GREEN → commit

**Total: 4 tests, 4 commits**

### Deliverables
- ✅ ToolDescriptor with TaintOutput field
- ✅ ToolDescriptor with Type, ChartRef, MaxIterations fields
- ✅ Sub-agent configuration support
- ✅ 4 commits (one per test per TDD workflow)

---

## Phase 6.2: Execution Policies

### Goal
Implement execution policies: seq_continue, seq_failfast, par_continue per arch-v1.md 8.2.

### Scope
- Define ExecutionPolicy type (arch-v1.md 513-522)
- Implement seq_continue policy (arch-v1.md 527-533)
- Implement seq_failfast policy (arch-v1.md 535-538)
- Implement par_continue policy (arch-v1.md 540-544)
- Note: par_failfast DROPPED (arch-v1.md 546-548)

### Current Status

#### Implementation Status
| Policy | Status | Notes |
|--------|--------|-------|
| ExecutionPolicy type | ❌ Missing | Required per arch-v1.md 513-522 |
| seq_continue | ❌ Missing | Required per arch-v1.md 527-533 |
| seq_failfast | ❌ Missing | Required per arch-v1.md 535-538 |
| par_continue | ❌ Missing | Required per arch-v1.md 540-544 |
| par_failfast | ✅ DROPPED | Per arch-v1.md 546-548 |

#### Files Status
| File | Status |
|------|--------|
| `pkg/orchestrator/policies.go` | ❌ MISSING - must create |

### Required Implementation

#### ExecutionPolicy Type
```go
// pkg/orchestrator/policies.go
type ExecutionPolicy struct {
    Mode        string  // seq_continue, seq_failfast, par_continue
    MaxRetries  int
    Isolation   string  // strict, process, container, sandbox
    MaxParallel int     // only for parallel modes
    TimeoutMs   int
}

var (
    PolicySeqContinue = ExecutionPolicy{
        Mode:      "seq_continue",
        MaxRetries: 1,
        Isolation:  "process",
    }
    
    PolicySeqFailFast = ExecutionPolicy{
        Mode:      "seq_failfast",
        MaxRetries: 2,
        Isolation:  "strict",
    }
    
    PolicyParContinue = ExecutionPolicy{
        Mode:        "par_continue",
        MaxRetries:  1,
        Isolation:   "strict",
        MaxParallel: 8,
    }
)
```

### Tests to Write (5 tests, 5 commits)

#### Test 1: TestExecutionPolicy_SeqContinueStructure
```go
func TestExecutionPolicy_SeqContinueStructure(t *testing.T)
```
**Acceptance Criteria:**
- PolicySeqContinue has Mode="seq_continue" (arch-v1.md 527)
- PolicySeqContinue has MaxRetries=1 (arch-v1.md 518)
- PolicySeqContinue has Isolation="process" (arch-v1.md 519)

#### Test 2: TestExecutionPolicy_SeqFailFastStructure
```go
func TestExecutionPolicy_SeqFailFastStructure(t *testing.T)
```
**Acceptance Criteria:**
- PolicySeqFailFast has Mode="seq_failfast" (arch-v1.md 535)
- PolicySeqFailFast has MaxRetries=2 (arch-v1.md 515)
- PolicySeqFailFast has Isolation="strict" (arch-v1.md 516)

#### Test 3: TestExecutionPolicy_ParContinueStructure
```go
func TestExecutionPolicy_ParContinueStructure(t *testing.T)
```
**Acceptance Criteria:**
- PolicyParContinue has Mode="par_continue" (arch-v1.md 540)
- PolicyParContinue has MaxParallel=8 (arch-v1.md 521)
- PolicyParContinue has Isolation="strict" (arch-v1.md 522)

#### Test 4: TestExecutionPolicy_ParFailFastDropped
```go
func TestExecutionPolicy_ParFailFastDropped(t *testing.T)
```
**Acceptance Criteria:**
- par_failfast policy is NOT defined (arch-v1.md 546-548)
- Rationale: "too racy, not worth complexity" (arch-v1.md 547)

#### Test 5: TestExecutionPolicy_AllPoliciesDefined
```go
func TestExecutionPolicy_AllPoliciesDefined(t *testing.T)
```
**Acceptance Criteria:**
- Exactly 3 policies defined: seq_continue, seq_failfast, par_continue
- All policies have valid Mode, MaxRetries, Isolation fields

### Dependencies

#### Test Dependencies
```
Test 1, Test 2, Test 3 → Test 5 (Policy definitions)
Test 4 → Independent (Verify par_failfast dropped)
```

### Files to Create/Modify

| File | Operation | Changes |
|------|-----------|---------|
| `pkg/orchestrator/policies.go` | CREATE | Define ExecutionPolicy type and 3 policy constants |

### Implementation Order (TDD)

1. **Test 1**: Write test → verify RED → Define PolicySeqContinue → verify GREEN → commit
2. **Test 2**: Write test → verify RED → Define PolicySeqFailFast → verify GREEN → commit
3. **Test 3**: Write test → verify RED → Define PolicyParContinue → verify GREEN → commit
4. **Test 4**: Write test → verify RED → Verify par_failfast not defined → verify GREEN → commit
5. **Test 5**: Write test → verify RED → Verify all policies defined → verify GREEN → commit

**Total: 5 tests, 5 commits**

### Deliverables
- ✅ ExecutionPolicy type defined
- ✅ PolicySeqContinue defined
- ✅ PolicySeqFailFast defined
- ✅ PolicyParContinue defined
- ✅ par_failfast explicitly dropped
- ✅ 5 commits (one per test per TDD workflow)

---

## Phase 6.3: Orchestrator Service

### Goal
Implement OrchestratorService as unified orchestration service per arch-v1.md 8.1.

### Scope
- Define OrchestratorService structure (arch-v1.md 8.1)
- Implement RegisterPolicy() method
- Implement ResolveTool() method (arch-v1.md 574-578)
- Implement Execute() method to spawn orchestrator charts (arch-v1.md 504)
- Integrate with ToolRegistry

### Current Status

#### Implementation Status
| Method | Status | Notes |
|--------|--------|-------|
| OrchestratorService struct | ❌ Missing | Required per arch-v1.md 8.1 |
| RegisterPolicy() | ❌ Missing | Required for policy management |
| ResolveTool() | ❌ Missing | Required per arch-v1.md 574-578 |
| Execute() | ❌ Missing | Required per arch-v1.md 504 |

#### Files Status
| File | Status |
|------|--------|
| `pkg/orchestrator/orchestrator.go` | ❌ MISSING - must create |
| `pkg/services/orchestrator/service.go` | ❌ MISSING - must create |

### Required Implementation

#### OrchestratorService
```go
// pkg/orchestrator/orchestrator.go
type OrchestratorService struct {
    id       string  // "sys:orchestrator"
    registry *ToolRegistry
    library  statechart.Library
    policies map[string]*ExecutionPolicy
    mu       sync.RWMutex
}

func NewOrchestratorService() *OrchestratorService

func (s *OrchestratorService) ID() string

func (s *OrchestratorService) RegisterPolicy(name string, policy ExecutionPolicy) error

func (s *OrchestratorService) ResolveTool(name string, callerBoundary mail.BoundaryType) (ToolDescriptor, error)

func (s *OrchestratorService) Execute(tools []ToolCall, policy ExecutionPolicy) (statechart.RuntimeID, error)
```

### Tests to Write (6 tests, 6 commits)

#### Test 1: TestOrchestratorService_NewCreatesService
```go
func TestOrchestratorService_NewCreatesService(t *testing.T)
```
**Acceptance Criteria:**
- NewOrchestratorService() returns non-nil service
- Service has ID="sys:orchestrator" (arch-v1.md 8.1)
- Service has empty policies map

#### Test 2: TestOrchestratorService_RegisterPolicy
```go
func TestOrchestratorService_RegisterPolicy(t *testing.T)
```
**Acceptance Criteria:**
- RegisterPolicy() stores policy by name
- Returns nil error on success
- Policy can be retrieved later

#### Test 3: TestOrchestratorService_ResolveTool
```go
func TestOrchestratorService_ResolveTool(t *testing.T)
```
**Acceptance Criteria:**
- ResolveTool() delegates to ToolRegistry (arch-v1.md 574)
- Returns ToolDescriptor for valid tool
- Returns error for non-existent tool (arch-v1.md 578)

#### Test 4: TestOrchestratorService_ResolveToolWithBoundary
```go
func TestOrchestratorService_ResolveToolWithBoundary(t *testing.T)
```
**Acceptance Criteria:**
- ResolveTool() filters by callerBoundary (arch-v1.md 578)
- Inner tools invisible to outer callers (arch-v1.md 578)

#### Test 5: TestOrchestratorService_ExecuteSpawnsChart
```go
func TestOrchestratorService_ExecuteSpawnsChart(t *testing.T)
```
**Acceptance Criteria:**
- Execute() spawns orchestrator chart (arch-v1.md 504)
- Returns RuntimeID for spawned chart
- Chart runs as transient parallel region (arch-v1.md 504)

#### Test 6: TestOrchestratorService_ExecuteWithPolicy
```go
func TestOrchestratorService_ExecuteWithPolicy(t *testing.T)
```
**Acceptance Criteria:**
- Execute() accepts ExecutionPolicy parameter
- Policy determines execution mode (seq_continue, seq_failfast, par_continue)

### Dependencies

#### Test Dependencies
```
Test 1 → Test 2, Test 3, Test 4, Test 5, Test 6 (Service creation)
Test 2 → Test 6 (Policy registration)
Test 3 → Test 4 (Tool resolution)
Test 5 → Test 6 (Execution with policy)
```

### Files to Create/Modify

| File | Operation | Changes |
|------|-----------|---------|
| `pkg/orchestrator/orchestrator.go` | CREATE | Define OrchestratorService struct and methods |
| `pkg/services/orchestrator/service.go` | CREATE | Implement service interface integration |

### Implementation Order (TDD)

1. **Test 1**: Write test → verify RED → Implement NewOrchestratorService() → verify GREEN → commit
2. **Test 2**: Write test → verify RED → Implement RegisterPolicy() → verify GREEN → commit
3. **Test 3**: Write test → verify RED → Implement ResolveTool() basic → verify GREEN → commit
4. **Test 4**: Write test → verify RED → Add boundary filtering to ResolveTool() → verify GREEN → commit
5. **Test 5**: Write test → verify RED → Implement Execute() stub → verify GREEN → commit
6. **Test 6**: Write test → verify RED → Add policy parameter to Execute() → verify GREEN → commit

**Total: 6 tests, 6 commits**

### Deliverables
- ✅ OrchestratorService struct defined
- ✅ RegisterPolicy() implemented
- ✅ ResolveTool() implemented with boundary filtering
- ✅ Execute() implemented
- ✅ 6 commits (one per test per TDD workflow)

---

## Phase 6.4: Sequential Executor

### Goal
Implement sequential tool execution with seq_continue and seq_failfast patterns per arch-v1.md 8.3.

### Scope
- Implement seq_continue pattern (arch-v1.md 527-533)
- Implement seq_failfast pattern (arch-v1.md 535-538)
- Add tool output to databag (arch-v1.md 531)
- Include failed tool calls in results (arch-v1.md 532)
- Short-circuit on first failure for seq_failfast (arch-v1.md 537)

### Current Status

#### Implementation Status
| Feature | Status | Notes |
|---------|--------|-------|
| seq_continue pattern | ❌ Missing | Required per arch-v1.md 527-533 |
| seq_failfast pattern | ❌ Missing | Required per arch-v1.md 535-538 |
| Databag output | ❌ Missing | Required per arch-v1.md 531 |
| Failed tool handling | ❌ Missing | Required per arch-v1.md 532 |
| Short-circuit logic | ❌ Missing | Required per arch-v1.md 537 |

#### Files Status
| File | Status |
|------|--------|
| `pkg/orchestrator/executor.go` | ❌ MISSING - must create |

### Required Implementation

#### SequentialExecutor
```go
// pkg/orchestrator/executor.go
type SequentialExecutor struct {
    policy ExecutionPolicy
    databag map[string]any
}

func (e *SequentialExecutor) Execute(tools []ToolCall) ([]ExecutionResult, error) {
    // seq_continue: execute all, include failures (arch-v1.md 527-533)
    // seq_failfast: short-circuit on first failure (arch-v1.md 535-538)
}
```

### Tests to Write (5 tests, 5 commits)

#### Test 1: TestSequentialExecutor_SeqContinueExecutesAll
```go
func TestSequentialExecutor_SeqContinueExecutesAll(t *testing.T)
```
**Acceptance Criteria:**
- seq_continue executes all tools in sequence (arch-v1.md 527)
- Each tool adds output to databag (arch-v1.md 531)
- Returns all execution results

#### Test 2: TestSequentialExecutor_SeqContinueIncludesFailures
```go
func TestSequentialExecutor_SeqContinueIncludesFailures(t *testing.T)
```
**Acceptance Criteria:**
- seq_continue includes failed tool calls in results (arch-v1.md 532)
- Failed output injected into session context (arch-v1.md 533)

#### Test 3: TestSequentialExecutor_SeqFailFastShortCircuits
```go
func TestSequentialExecutor_SeqFailFastShortCircuits(t *testing.T)
```
**Acceptance Criteria:**
- seq_failfast short-circuits on first failure (arch-v1.md 537)
- Aborts all following calls (arch-v1.md 537)

#### Test 4: TestSequentialExecutor_SeqFailFastInjectsOutput
```go
func TestSequentialExecutor_SeqFailFastInjectsOutput(t *testing.T)
```
**Acceptance Criteria:**
- seq_failfast injects all tool output into session context (arch-v1.md 538)
- Output scoped to session (arch-v1.md 538)

#### Test 5: TestSequentialExecutor_DatabagScopedToChart
```go
func TestSequentialExecutor_DatabagScopedToChart(t *testing.T)
```
**Acceptance Criteria:**
- Databag scoped to calling chart (arch-v1.md 531)
- Each tool output accessible by name

### Dependencies

#### Test Dependencies
```
Test 1 → Test 2 (seq_continue behavior)
Test 3 → Test 4 (seq_failfast behavior)
Test 5 (Databag scoping)
```

### Files to Create/Modify

| File | Operation | Changes |
|------|-----------|---------|
| `pkg/orchestrator/executor.go` | CREATE | Define SequentialExecutor and Execute() method |

### Implementation Order (TDD)

1. **Test 1**: Write test → verify RED → Implement seq_continue basic execution → verify GREEN → commit
2. **Test 2**: Write test → verify RED → Add failure handling to seq_continue → verify GREEN → commit
3. **Test 3**: Write test → verify RED → Implement seq_failfast short-circuit → verify GREEN → commit
4. **Test 4**: Write test → verify RED → Add output injection to seq_failfast → verify GREEN → commit
5. **Test 5**: Write test → verify RED → Implement databag scoping → verify GREEN → commit

**Total: 5 tests, 5 commits**

### Deliverables
- ✅ SequentialExecutor struct defined
- ✅ seq_continue pattern implemented
- ✅ seq_failfast pattern implemented
- ✅ Databag scoping implemented
- ✅ 5 commits (one per test per TDD workflow)

---

## Phase 6.5: Parallel Executor

### Goal
Implement parallel tool execution with par_continue pattern per arch-v1.md 8.3.

### Scope
- Implement par_continue pattern (arch-v1.md 540-544)
- Spawn parallel states for each tool-call (arch-v1.md 541)
- Execute all tools concurrently (arch-v1.md 542)
- Caller continues immediately (async completion) (arch-v1.md 544)

### Current Status

#### Implementation Status
| Feature | Status | Notes |
|---------|--------|-------|
| par_continue pattern | ❌ Missing | Required per arch-v1.md 540-544 |
| Parallel state spawning | ❌ Missing | Required per arch-v1.md 541 |
| Concurrent execution | ❌ Missing | Required per arch-v1.md 542 |
| Async completion | ❌ Missing | Required per arch-v1.md 544 |

#### Files Status
| File | Status |
|------|--------|
| `pkg/orchestrator/executor.go` | ⚠️ Partial - add ParallelExecutor |

### Required Implementation

#### ParallelExecutor
```go
// pkg/orchestrator/executor.go
type ParallelExecutor struct {
    policy ExecutionPolicy
    maxParallel int
}

func (e *ParallelExecutor) Execute(tools []ToolCall) (<-chan ExecutionResult, error) {
    // Spawn parallel states for each tool-call (arch-v1.md 541)
    // All execute concurrently (arch-v1.md 542)
    // Caller continues immediately (arch-v1.md 544)
}
```

### Tests to Write (5 tests, 5 commits)

#### Test 1: TestParallelExecutor_ParContinueSpawnsParallelStates
```go
func TestParallelExecutor_ParContinueSpawnsParallelStates(t *testing.T)
```
**Acceptance Criteria:**
- par_continue spawns parallel states for each tool-call (arch-v1.md 541)
- Each tool runs in separate state

#### Test 2: TestParallelExecutor_ParContinueExecutesConcurrently
```go
func TestParallelExecutor_ParContinueExecutesConcurrently(t *testing.T)
```
**Acceptance Criteria:**
- All tools execute concurrently (arch-v1.md 542)
- Execution time less than sequential sum

#### Test 3: TestParallelExecutor_CallerContinuesImmediately
```go
func TestParallelExecutor_CallerContinuesImmediately(t *testing.T)
```
**Acceptance Criteria:**
- Caller continues immediately after Execute() (arch-v1.md 544)
- Async completion via channel

#### Test 4: TestParallelExecutor_MaxParallelLimit
```go
func TestParallelExecutor_MaxParallelLimit(t *testing.T)
```
**Acceptance Criteria:**
- MaxParallel limits concurrent execution (arch-v1.md 521)
- Default MaxParallel=8 (arch-v1.md 521)

#### Test 5: TestParallelExecutor_SessionContextInsertion
```go
func TestParallelExecutor_SessionContextInsertion(t *testing.T)
```
**Acceptance Criteria:**
- Results inserted into session context (arch-v1.md 543)
- All tool outputs included

### Dependencies

#### Test Dependencies
```
Test 1 → Test 2 (Parallel spawning)
Test 2 → Test 3 (Concurrent execution)
Test 4 (MaxParallel limit)
Test 5 (Session context)
```

### Files to Create/Modify

| File | Operation | Changes |
|------|-----------|---------|
| `pkg/orchestrator/executor.go` | MODIFY | Add ParallelExecutor and Execute() method |

### Implementation Order (TDD)

1. **Test 1**: Write test → verify RED → Implement parallel state spawning → verify GREEN → commit
2. **Test 2**: Write test → verify RED → Implement concurrent execution → verify GREEN → commit
3. **Test 3**: Write test → verify RED → Implement async completion → verify GREEN → commit
4. **Test 4**: Write test → verify RED → Add MaxParallel limit → verify GREEN → commit
5. **Test 5**: Write test → verify RED → Implement session context insertion → verify GREEN → commit

**Total: 5 tests, 5 commits**

### Deliverables
- ✅ ParallelExecutor struct defined
- ✅ Parallel state spawning implemented
- ✅ Concurrent execution implemented
- ✅ Async completion implemented
- ✅ 5 commits (one per test per TDD workflow)

---

## Phase 6.6: Sub-Agent Support

### Goal
Implement attached and detached sub-agent support per arch-v1.md 6.4.

### Scope
- Implement attached sub-agents (arch-v1.md 6.4)
- Implement detached sub-agents (arch-v1.md 6.4)
- Lifecycle bound to parent for attached (arch-v1.md 6.4)
- Auto-terminated on parent exit (arch-v1.md 6.4)
- subAgentDone event emission (arch-v1.md 6.4)
- Fire-and-forget for detached (arch-v1.md 6.4)

### Current Status

#### Implementation Status
| Feature | Status | Notes |
|---------|--------|-------|
| Attached sub-agents | ❌ Missing | Required per arch-v1.md 6.4 |
| Detached sub-agents | ❌ Missing | Required per arch-v1.md 6.4 |
| Lifecycle binding | ❌ Missing | Required per arch-v1.md 6.4 |
| Auto-termination | ❌ Missing | Required per arch-v1.md 6.4 |
| subAgentDone event | ❌ Missing | Required per arch-v1.md 6.4 |

#### Files Status
| File | Status |
|------|--------|
| `pkg/orchestrator/subagent.go` | ❌ MISSING - must create |
| `pkg/orchestrator/types.go` | ⚠️ Partial - add SubAgent types |

### Required Implementation

#### SubAgent Types
```go
// pkg/orchestrator/subagent.go
type SubAgentType string

const (
    SubAgentAttached  SubAgentType = "attached"
    SubAgentDetached  SubAgentType = "detached"
)

type SubAgentConfig struct {
    Type          SubAgentType
    ChartRef      string
    MaxIterations int
    InheritContext bool
}

type SubAgentExecutor struct {
    config SubAgentConfig
    parentNs string
}

func (e *SubAgentExecutor) Execute() (any, error) {
    // Attached: spawn child ChartRuntime bound to parent (arch-v1.md 6.4)
    // Detached: spawn top-level ChartRuntime via sys:lifecycle (arch-v1.md 6.4)
}
```

### Tests to Write (5 tests, 5 commits)

#### Test 1: TestSubAgent_AttachedLifecycleBound
```go
func TestSubAgent_AttachedLifecycleBound(t *testing.T)
```
**Acceptance Criteria:**
- Attached sub-agent spawns child ChartRuntime (arch-v1.md 6.4)
- Lifecycle bound to parent (arch-v1.md 6.4)

#### Test 2: TestSubAgent_AttachedAutoTerminated
```go
func TestSubAgent_AttachedAutoTerminated(t *testing.T)
```
**Acceptance Criteria:**
- Attached sub-agent auto-terminated on parent exit (arch-v1.md 6.4)
- Parent exit triggers sub-agent termination

#### Test 3: TestSubAgent_DetachedFireAndForget
```go
func TestSubAgent_DetachedFireAndForget(t *testing.T)
```
**Acceptance Criteria:**
- Detached sub-agent spawns top-level ChartRuntime (arch-v1.md 6.4)
- Managed by sys:lifecycle (arch-v1.md 6.4)
- Fire-and-forget behavior

#### Test 4: TestSubAgent_DetachedFireAndAwait
```go
func TestSubAgent_DetachedFireAndAwait(t *testing.T)
```
**Acceptance Criteria:**
- Detached sub-agent supports fire-and-await (arch-v1.md 6.4)
- Via mail correlationId (arch-v1.md 6.4)

#### Test 5: TestSubAgent_subAgentDoneEvent
```go
func TestSubAgent_subAgentDoneEvent(t *testing.T)
```
**Acceptance Criteria:**
- Result returned via subAgentDone event (arch-v1.md 6.4)
- Optional Messages injected into parent session (arch-v1.md 6.4)

### Dependencies

#### Test Dependencies
```
Test 1 → Test 2 (Attached lifecycle)
Test 3 → Test 4 (Detached behavior)
Test 5 (Event emission)
```

### Files to Create/Modify

| File | Operation | Changes |
|------|-----------|---------|
| `pkg/orchestrator/subagent.go` | CREATE | Define SubAgent types and executor |
| `pkg/orchestrator/types.go` | MODIFY | Add SubAgentType constants |

### Implementation Order (TDD)

1. **Test 1**: Write test → verify RED → Implement attached sub-agent spawning → verify GREEN → commit
2. **Test 2**: Write test → verify RED → Implement auto-termination on parent exit → verify GREEN → commit
3. **Test 3**: Write test → verify RED → Implement detached sub-agent spawning → verify GREEN → commit
4. **Test 4**: Write test → verify RED → Implement fire-and-await via correlationId → verify GREEN → commit
5. **Test 5**: Write test → verify RED → Implement subAgentDone event emission → verify GREEN → commit

**Total: 5 tests, 5 commits**

### Deliverables
- ✅ SubAgentType constants defined
- ✅ Attached sub-agent support implemented
- ✅ Detached sub-agent support implemented
- ✅ subAgentDone event emission implemented
- ✅ 5 commits (one per test per TDD workflow)

---

## Phase 6.7: Integration & Schema Sanitization

### Goal
Integrate all components and implement schema sanitization per arch-v1.md 8.4.

### Scope
- Implement SanitizeSchema() method (arch-v1.md 578)
- Wire OrchestratorService with ToolRegistry
- Wire OrchestratorService with executors
- End-to-end orchestration flow test

### Current Status

#### Implementation Status
| Feature | Status | Notes |
|---------|--------|-------|
| SanitizeSchema() | ❌ Missing | Required per arch-v1.md 578 |
| OrchestratorService integration | ❌ Missing | Required for full flow |
| End-to-end orchestration | ❌ Missing | Required for validation |

#### Files Status
| File | Status |
|------|--------|
| `pkg/services/tools/service.go` | ⚠️ Partial - add SanitizeSchema() |
| `pkg/orchestrator/orchestrator_test.go` | ❌ MISSING - must create |

### Required Implementation

#### SanitizeSchema Method
```go
// pkg/services/tools/service.go
func (s *toolsService) SanitizeSchema(tool ToolDescriptor, callerBoundary mail.BoundaryType) JSONSchema {
    // Return sanitized schema filtered by caller's boundary (arch-v1.md 578)
    // Inner tools are invisible to outer callers (arch-v1.md 578)
}
```

### Tests to Write (2 tests, 2 commits)

#### Test 1: TestOrchestratorService_SanitizeSchema
```go
func TestOrchestratorService_SanitizeSchema(t *testing.T)
```
**Acceptance Criteria:**
- SanitizeSchema() filters schema by caller boundary (arch-v1.md 578)
- Inner tools invisible to outer callers (arch-v1.md 578)
- Returns sanitized schema with appropriate fields

#### Test 2: TestOrchestratorService_FullOrchestrationFlow
```go
func TestOrchestratorService_FullOrchestrationFlow(t *testing.T)
```
**Acceptance Criteria:**
- End-to-end orchestration flow works (arch-v1.md 8.1)
- Tool calls resolved → executed → results injected
- All components integrated: ToolRegistry, OrchestratorService, Executors

### Dependencies

#### Test Dependencies
```
Test 1 (Schema sanitization)
Test 2 (Full integration - depends on all previous phases)
```

### Files to Create/Modify

| File | Operation | Changes |
|------|-----------|---------|
| `pkg/services/tools/service.go` | MODIFY | Add SanitizeSchema() method |
| `pkg/orchestrator/orchestrator_test.go` | CREATE | Add integration tests |

### Implementation Order (TDD)

1. **Test 1**: Write test → verify RED → Implement SanitizeSchema() → verify GREEN → commit
2. **Test 2**: Write test → verify RED → Wire all components together → verify GREEN → commit

**Total: 2 tests, 2 commits**

### Deliverables
- ✅ SanitizeSchema() implemented
- ✅ All components integrated
- ✅ End-to-end orchestration flow validated
- ✅ 2 commits (one per test per TDD workflow)

---

## Summary

### Phase Summary Table

| Phase | Tests | Files | Branch | Dependencies |
|-------|-------|-------|--------|--------------|
| 6.1 | 4 | 2 | `feat/layer6-tool-descriptor` | None |
| 6.2 | 5 | 1 | `feat/layer6-execution-policies` | 6.1 |
| 6.3 | 6 | 2 | `feat/layer6-orchestrator-service` | 6.1, 6.2 |
| 6.4 | 5 | 1 | `feat/layer6-sequential-executor` | 6.3 |
| 6.5 | 5 | 1 | `feat/layer6-parallel-executor` | 6.3 |
| 6.6 | 5 | 2 | `feat/layer6-subagent` | 6.3, 6.4, 6.5 |
| 6.7 | 2 | 2 | `feat/layer6-integration` | 6.1-6.6 |

**Total: 32 tests, 11 files, 7 phases**

### Execution Order

```
Phase 6.1 (Tool Descriptor Enhancement)
    ↓
Phase 6.2 (Execution Policies)
    ↓
Phase 6.3 (Orchestrator Service)
    ↓
┌─────────────────────────────────────────┐
│     Phases 6.4-6.6 (can run in parallel)│
│  ┌──────────────┬─────────────────────┐ │
│  │ 6.4 Sequential│ 6.5 Parallel       │ │
│  │ 6.6 Sub-Agent │                    │ │
│  └──────────────┴─────────────────────┘ │
└─────────────────────────────────────────┘
    ↓
Phase 6.7 (Integration & Schema Sanitization)
```

### Files to Create

| File | Phase | Purpose |
|------|-------|---------|
| `pkg/orchestrator/types.go` | 6.1 | Enhanced ToolDescriptor and related types |
| `pkg/orchestrator/policies.go` | 6.2 | ExecutionPolicy type and constants |
| `pkg/orchestrator/orchestrator.go` | 6.3 | OrchestratorService implementation |
| `pkg/services/orchestrator/service.go` | 6.3 | Service interface integration |
| `pkg/orchestrator/executor.go` | 6.4, 6.5 | SequentialExecutor and ParallelExecutor |
| `pkg/orchestrator/subagent.go` | 6.6 | SubAgent types and executor |
| `pkg/orchestrator/orchestrator_test.go` | 6.7 | Integration tests |

### Files to Modify

| File | Phases | Changes |
|------|--------|---------|
| `pkg/services/tools/service.go` | 6.1, 6.7 | Add 4 new fields to ToolDescriptor, add SanitizeSchema() |
| `pkg/orchestrator/types.go` | 6.6 | Add SubAgentType constants |

### Satisfied Requirements from Layers 0-5

| Requirement | Layer | Reference |
|-------------|-------|-----------|
| Statechart Engine | Layer 0 | arch-v1.md Section 2 |
| ToolRegistry basic structure | Layer 5 | arch-v1.md Section 7.2, line 289-319 |
| Boundary types (inner, dmz, outer) | Layer 4 | arch-v1.md Section 3, lines 261-270 |
| Mail types | Layer 3 | arch-v1.md Section 4 |
| sys:lifecycle for spawning | Layer 2 | arch-v1.md Section 5 |
| ChartRegistry for YAML loading | Layer 1 | arch-v1.md Section 3.1 |

### Dependencies on Layer 5

Layer 6 depends on Layer 5 for:
- `sys:tools` service as ToolRegistry backend (arch-v1.md 7.2, line 48)
- `sys:lifecycle` for spawning orchestrator charts (arch-v1.md 7.2)

### Next Steps

1. **Start Phase 6.1**: Create branch `feat/layer6-tool-descriptor`
2. **Follow TDD workflow**: Write test → verify RED → Implement → verify GREEN → commit
3. **Repeat for each phase**: One branch per phase, merge after completion
4. **Run full test suite** after Phase 6.7 to verify integration

---

## Appendix A: arch-v1.md Line Number References

| Feature | Line Numbers | Description |
|---------|--------------|-------------|
| Orchestrator unified model | 498-504 | Tool calls and sub-agent calls as actions |
| Orchestrator YAML shape | 508-523 | Execution mode, failure policy, completion signaling |
| seq_continue pattern | 527-533 | Sequential, no-fail-on-fail |
| seq_failfast pattern | 535-538 | Sequential, fail-on-first-fail |
| par_continue pattern | 540-544 | Parallel, no-fail-on-fail |
| par_failfast dropped | 546-548 | Too racy, not worth complexity |
| Tool registry schema | 554-573 | Tool registration with boundary-aware schemas |
| Tool resolution API | 574-579 | resolveTool(name, callerBoundary) |
| Sub-agents | 6.4 (Section 6) | Attached and detached sub-agents |

## Appendix B: Layer 1-5 Patterns Applied

This phase breakdown follows established patterns:
- ✅ **1:1 test-to-commit ratio** (32 tests = 32 commits)
- ✅ **One branch per phase** (7 branches total)
- ✅ **Descriptive test names** (`Test[Component]_[Behavior]_[ExpectedResult]`)
- ✅ **Small, focused phases** (2-6 tests per phase)
- ✅ **Explicit dependencies** between phases
- ✅ **Clear acceptance criteria** for each test
- ✅ **arch-v1.md line number references** for all requirements

---