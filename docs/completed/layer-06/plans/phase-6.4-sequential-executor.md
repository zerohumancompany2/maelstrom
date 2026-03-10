# Phase 6.4: Sequential Executor

## Goal
Implement sequential tool execution with seq_continue and seq_failfast patterns following strict TDD workflow with 5 tests.

## Scope
- Implement seq_continue pattern (arch-v1.md lines 527-533)
- Implement seq_failfast pattern (arch-v1.md lines 535-538)
- Add tool output to databag (arch-v1.md line 531)
- Include failed tool calls in results (arch-v1.md line 532)
- Short-circuit on first failure for seq_failfast (arch-v1.md line 537)

## Current Status

### Implementation Status
| Test | Status | Notes |
|------|--------|-------|
| Test 1: TestSequentialExecutor_SeqContinueExecutesAll | ❌ Missing | SequentialExecutor does not exist |
| Test 2: TestSequentialExecutor_SeqContinueIncludesFailures | ❌ Missing | Failure handling does not exist |
| Test 3: TestSequentialExecutor_SeqFailFastShortCircuits | ❌ Missing | Short-circuit logic does not exist |
| Test 4: TestSequentialExecutor_SeqFailFastInjectsOutput | ❌ Missing | Output injection does not exist |
| Test 5: TestSequentialExecutor_DatabagScopedToChart | ❌ Missing | Databag scoping does not exist |

### Files Status
| File | Status |
|------|--------|
| `pkg/orchestrator/executor.go` | ❌ Missing - must create with SequentialExecutor struct |
| `pkg/orchestrator/executor_test.go` | ❌ Missing - must create with 5 new tests |

## Required Implementation

### SequentialExecutor Struct
```go
type SequentialExecutor struct {
    policy  ExecutionPolicy
    databag map[string]any
}
```

### Execute Method
```go
func (e *SequentialExecutor) Execute(tools []ToolCall) ([]ExecutionResult, error) {
    // seq_continue: execute all, include failures (arch-v1.md lines 527-533)
    // seq_failfast: short-circuit on first failure (arch-v1.md lines 535-538)
}
```

### arch-v1.md References

**Sequential, No-Fail-On-Fail** (arch-v1.md Section 8.3, lines 527-533):
```
tool call → awaiting → tool call → awaiting → ...
```
- Each tool call adds its output to the databag (scoped to calling chart) (arch-v1.md line 531)
- Includes failed tool calls (arch-v1.md line 532)
- When results re-inserted into session context, includes failed tool-call output (arch-v1.md line 533)

**Sequential, Fail-On-First-Fail** (arch-v1.md Section 8.3, lines 535-538):
- Short-circuiting AND application on tool-calls (arch-v1.md line 536)
- First failed tool call aborts all following calls (arch-v1.md line 537)
- All tool output injected back into session context (scoped to session) (arch-v1.md line 538)

## Tests to Write (5 tests, 5 commits)

### Test 1: TestSequentialExecutor_SeqContinueExecutesAll
```go
func TestSequentialExecutor_SeqContinueExecutesAll()
```
**Acceptance Criteria:**
- seq_continue executes all tools in sequence (arch-v1.md line 527)
- Each tool adds output to databag (arch-v1.md line 531)
- Returns all execution results

### Test 2: TestSequentialExecutor_SeqContinueIncludesFailures
```go
func TestSequentialExecutor_SeqContinueIncludesFailures()
```
**Acceptance Criteria:**
- seq_continue includes failed tool calls in results (arch-v1.md line 532)
- Failed output injected into session context (arch-v1.md line 533)

### Test 3: TestSequentialExecutor_SeqFailFastShortCircuits
```go
func TestSequentialExecutor_SeqFailFastShortCircuits()
```
**Acceptance Criteria:**
- seq_failfast short-circuits on first failure (arch-v1.md line 537)
- Aborts all following calls (arch-v1.md line 537)

### Test 4: TestSequentialExecutor_SeqFailFastInjectsOutput
```go
func TestSequentialExecutor_SeqFailFastInjectsOutput()
```
**Acceptance Criteria:**
- seq_failfast injects all tool output into session context (arch-v1.md line 538)
- Output scoped to session (arch-v1.md line 538)

### Test 5: TestSequentialExecutor_DatabagScopedToChart
```go
func TestSequentialExecutor_DatabagScopedToChart()
```
**Acceptance Criteria:**
- Databag scoped to calling chart (arch-v1.md line 531)
- Each tool output accessible by name

## Dependencies

### Test Dependencies
```
Test 1 → Test 2 (seq_continue behavior)
Test 3 → Test 4 (seq_failfast behavior)
Test 5 (Databag scoping)
```

### Phase Dependencies
- Phase 6.3 (Orchestrator Service) - provides OrchestratorService and Execute() stub

## Files to Create/Modify

| File | Operation | Changes |
|------|-----------|---------|
| `pkg/orchestrator/executor.go` | CREATE | Define SequentialExecutor struct and Execute() method |
| `pkg/orchestrator/executor_test.go` | CREATE | Add 5 new tests |

## Implementation Order (TDD)

1. **Test 1**: Write test → verify RED → Implement seq_continue basic execution → verify GREEN → commit
2. **Test 2**: Write test → verify RED → Add failure handling to seq_continue → verify GREEN → commit
3. **Test 3**: Write test → verify RED → Implement seq_failfast short-circuit → verify GREEN → commit
4. **Test 4**: Write test → verify RED → Add output injection to seq_failfast → verify GREEN → commit
5. **Test 5**: Write test → verify RED → Implement databag scoping → verify GREEN → commit

**Total: 5 tests, 5 commits**

## Deliverables
- Created `pkg/orchestrator/executor.go` with SequentialExecutor struct and Execute() method
- Created `pkg/orchestrator/executor_test.go` with 5 passing tests
- 5 commits (one per test per TDD workflow)