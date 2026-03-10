# Phase 6.2: Execution Policies

## Goal
Implement execution policies (seq_continue, seq_failfast, par_continue) following strict TDD workflow with 5 tests.

## Scope
- Define ExecutionPolicy type with Mode, MaxRetries, Isolation, MaxParallel, TimeoutMs
- Implement PolicySeqContinue policy
- Implement PolicySeqFailFast policy
- Implement PolicyParContinue policy
- Verify par_failfast is explicitly dropped
- Create 5 tests in `pkg/orchestrator/policies_test.go`

## Current Status

### Implementation Status
| Test | Status | Notes |
|------|--------|-------|
| Test 1: TestExecutionPolicy_SeqContinueStructure | ❌ Missing | PolicySeqContinue does not exist |
| Test 2: TestExecutionPolicy_SeqFailFastStructure | ❌ Missing | PolicySeqFailFast does not exist |
| Test 3: TestExecutionPolicy_ParContinueStructure | ❌ Missing | PolicyParContinue does not exist |
| Test 4: TestExecutionPolicy_ParFailFastDropped | ❌ Missing | par_failfast must be verified as dropped |
| Test 5: TestExecutionPolicy_AllPoliciesDefined | ❌ Missing | All 3 policies must be defined |

### Files Status
| File | Status |
|------|--------|
| `pkg/orchestrator/policies.go` | ❌ Missing - must create with ExecutionPolicy type and constants |
| `pkg/orchestrator/policies_test.go` | ❌ Missing - must create with 5 new tests |

## Required Implementation

### ExecutionPolicy Struct
```go
type ExecutionPolicy struct {
    Mode        string
    MaxRetries  int
    Isolation   string
    MaxParallel int
    TimeoutMs   int
}
```

### Policy Constants
```go
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

### arch-v1.md References

**Orchestrator YAML Shape** (arch-v1.md Section 8.2, lines 506-524):
```yaml
orchestrator:
  default: seq_failfast          # or seq_continue, par_continue
  timeoutMs: 30000
  maxParallel: 8                  # only for par_*
  policies:
    seq_failfast:
      maxRetries: 2
      isolation: strict             # each tool in own namespace
    seq_continue:
      maxRetries: 1
      isolation: process
    par_continue:
      maxParallel: 8
      isolation: strict
```

**seq_continue Pattern** (arch-v1.md Section 8.3, lines 527-533):
- Sequential, No-Fail-On-Fail
- Each tool call adds its output to the databag
- Includes failed tool calls

**seq_failfast Pattern** (arch-v1.md Section 8.3, lines 535-538):
- Sequential, Fail-On-First-Fail
- First failed tool call aborts all following calls
- All tool output injected back into session context

**par_continue Pattern** (arch-v1.md Section 8.3, lines 540-544):
- Parallel, No-Fail-On-Fail
- Spawns parallel states for each tool-call
- All execute concurrently
- Caller continues immediately (async completion)

**par_failfast Dropped** (arch-v1.md Section 8.3, lines 546-548):
- Dropped from MVP (too racy, not worth complexity)
- Fail-fast is almost always sequential in practice

## Tests to Write (5 tests, 5 commits)

### Test 1: TestExecutionPolicy_SeqContinueStructure
```go
func TestExecutionPolicy_SeqContinueStructure()
```
**Acceptance Criteria:**
- PolicySeqContinue has Mode="seq_continue" (arch-v1.md line 510)
- PolicySeqContinue has MaxRetries=1 (arch-v1.md line 518)
- PolicySeqContinue has Isolation="process" (arch-v1.md line 519)

### Test 2: TestExecutionPolicy_SeqFailFastStructure
```go
func TestExecutionPolicy_SeqFailFastStructure()
```
**Acceptance Criteria:**
- PolicySeqFailFast has Mode="seq_failfast" (arch-v1.md line 510)
- PolicySeqFailFast has MaxRetries=2 (arch-v1.md line 515)
- PolicySeqFailFast has Isolation="strict" (arch-v1.md line 516)

### Test 3: TestExecutionPolicy_ParContinueStructure
```go
func TestExecutionPolicy_ParContinueStructure()
```
**Acceptance Criteria:**
- PolicyParContinue has Mode="par_continue" (arch-v1.md line 510)
- PolicyParContinue has MaxParallel=8 (arch-v1.md line 521)
- PolicyParContinue has Isolation="strict" (arch-v1.md line 522)

### Test 4: TestExecutionPolicy_ParFailFastDropped
```go
func TestExecutionPolicy_ParFailFastDropped()
```
**Acceptance Criteria:**
- par_failfast policy is NOT defined (arch-v1.md lines 546-548)
- Rationale: "too racy, not worth complexity" (arch-v1.md line 547)

### Test 5: TestExecutionPolicy_AllPoliciesDefined
```go
func TestExecutionPolicy_AllPoliciesDefined()
```
**Acceptance Criteria:**
- Exactly 3 policies defined: seq_continue, seq_failfast, par_continue
- All policies have valid Mode, MaxRetries, Isolation fields
- No additional policies defined

## Dependencies

### Test Dependencies
```
Test 1, Test 2, Test 3 → Test 5 (Policy definitions)
Test 4 → Independent (Verify par_failfast dropped)
```

### Phase Dependencies
- Phase 6.1 (Tool Descriptor Enhancement) - provides ToolDescriptor type

## Files to Create/Modify

| File | Operation | Changes |
|------|-----------|---------|
| `pkg/orchestrator/policies.go` | CREATE | Define ExecutionPolicy struct and 3 policy constants |
| `pkg/orchestrator/policies_test.go` | CREATE | Add 5 new tests |

## Implementation Order (TDD)

1. **Test 1**: Write test → verify RED → Define PolicySeqContinue → verify GREEN → commit
2. **Test 2**: Write test → verify RED → Define PolicySeqFailFast → verify GREEN → commit
3. **Test 3**: Write test → verify RED → Define PolicyParContinue → verify GREEN → commit
4. **Test 4**: Write test → verify RED → Verify par_failfast not defined → verify GREEN → commit
5. **Test 5**: Write test → verify RED → Verify all policies defined → verify GREEN → commit

**Total: 5 tests, 5 commits**

## Deliverables
- Created `pkg/orchestrator/policies.go` with ExecutionPolicy type and 3 policy constants
- Created `pkg/orchestrator/policies_test.go` with 5 passing tests
- 5 commits (one per test per TDD workflow)