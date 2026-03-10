# Phase 6.5: Parallel Executor

## Goal
Implement parallel tool execution with par_continue pattern per arch-v1.md 8.3.

## Scope
- Implement par_continue pattern (arch-v1.md 540-544)
- Spawn parallel states for each tool-call (arch-v1.md 541)
- Execute all tools concurrently (arch-v1.md 542)
- MaxParallel enforcement (arch-v1.md 521)
- Async completion signaling (arch-v1.md 544)
- Result aggregation into session context (arch-v1.md 543)
- 5 tests, 5 commits

## Current Status

### Implementation Status
| Feature | Status | Notes |
|---------|--------|-------|
| par_continue pattern | ❌ Missing | Required per arch-v1.md 540-544 |
| Parallel state spawning | ❌ Missing | Required per arch-v1.md 541 |
| Concurrent execution | ❌ Missing | Required per arch-v1.md 542 |
| MaxParallel enforcement | ❌ Missing | Required per arch-v1.md 521 |
| Async completion signaling | ❌ Missing | Required per arch-v1.md 544 |
| Result aggregation | ❌ Missing | Required per arch-v1.md 543 |

### Files Status
| File | Status |
|------|-------|
| `pkg/orchestrator/executor.go` | ⚠️ Partial - add ParallelExecutor |

## Required Implementation

### ParallelExecutor
```go
// pkg/orchestrator/executor.go
type ParallelExecutor struct {
    policy      ExecutionPolicy
    maxParallel int
    databag     map[string]any
}

func NewParallelExecutor(policy ExecutionPolicy) *ParallelExecutor {
    return &ParallelExecutor{
        policy:      policy,
        maxParallel: policy.MaxParallel,
        databag:     make(map[string]any),
    }
}

func (e *ParallelExecutor) Execute(tools []ToolCall) (<-chan ExecutionResult, error) {
    // Spawn parallel states for each tool-call (arch-v1.md 541)
    // All execute concurrently (arch-v1.md 542)
    // Caller continues immediately (arch-v1.md 544)
    // Results aggregated via channel
}
```

### ExecutionResult
```go
type ExecutionResult struct {
    ToolName string
    Output   any
    Error    error
}
```

## Tests to Write (5 tests, 5 commits)

### Test 1: TestParallelExecutor_ParContinueSpawnsParallelStates
```go
func TestParallelExecutor_ParContinueSpawnsParallelStates(t *testing.T) {
    policy := PolicyParContinue
    executor := NewParallelExecutor(policy)
    
    tools := []ToolCall{
        {Name: "tool1", Args: map[string]any{"arg": "value1"}},
        {Name: "tool2", Args: map[string]any{"arg": "value2"}},
        {Name: "tool3", Args: map[string]any{"arg": "value3"}},
    }
    
    resultChan, err := executor.Execute(tools)
    if err != nil {
        t.Fatalf("Expected nil error, got %v", err)
    }
    
    // Verify channel returned (non-blocking)
    if resultChan == nil {
        t.Fatal("Expected non-nil result channel")
    }
    
    // Verify all tools spawned (collect results)
    results := make([]ExecutionResult, 0, 3)
    for result := range resultChan {
        results = append(results, result)
    }
    
    if len(results) != 3 {
        t.Errorf("Expected 3 results, got %d", len(results))
    }
}
```
**Acceptance Criteria:**
- par_continue spawns parallel states for each tool-call (arch-v1.md 541)
- Each tool runs in separate state
- Returns non-nil result channel immediately
- All tools produce results

### Test 2: TestParallelExecutor_ParContinueExecutesConcurrently
```go
func TestParallelExecutor_ParContinueExecutesConcurrently(t *testing.T) {
    policy := PolicyParContinue
    executor := NewParallelExecutor(policy)
    
    // Create tools that take measurable time
    tools := []ToolCall{
        {Name: "slowTool1", Args: map[string]any{"delay": 100}},
        {Name: "slowTool2", Args: map[string]any{"delay": 100}},
        {Name: "slowTool3", Args: map[string]any{"delay": 100}},
    }
    
    startTime := time.Now()
    resultChan, err := executor.Execute(tools)
    if err != nil {
        t.Fatalf("Expected nil error, got %v", err)
    }
    
    // Collect all results
    for range resultChan {
        // Wait for completion
    }
    
    elapsed := time.Since(startTime)
    
    // If sequential, would take ~300ms
    // If parallel, should take ~100ms (+ overhead)
    if elapsed > 200*time.Millisecond {
        t.Errorf("Expected concurrent execution (<200ms), took %v", elapsed)
    }
}
```
**Acceptance Criteria:**
- All tools execute concurrently (arch-v1.md 542)
- Execution time less than sequential sum
- Parallel execution verified via timing

### Test 3: TestParallelExecutor_CallerContinuesImmediately
```go
func TestParallelExecutor_CallerContinuesImmediately(t *testing.T) {
    policy := PolicyParContinue
    executor := NewParallelExecutor(policy)
    
    tools := []ToolCall{
        {Name: "slowTool", Args: map[string]any{"delay": 500}},
    }
    
    // Execute should return immediately
    startTime := time.Now()
    resultChan, err := executor.Execute(tools)
    callTime := time.Since(startTime)
    
    if err != nil {
        t.Fatalf("Expected nil error, got %v", err)
    }
    
    // Execute() should return immediately (async)
    if callTime > 10*time.Millisecond {
        t.Errorf("Expected immediate return (<10ms), took %v", callTime)
    }
    
    // Verify resultChan is non-nil and usable
    if resultChan == nil {
        t.Fatal("Expected non-nil result channel")
    }
    
    // Result arrives later (async completion)
    result := <-resultChan
    if result.ToolName != "slowTool" {
        t.Errorf("Expected tool 'slowTool', got '%s'", result.ToolName)
    }
}
```
**Acceptance Criteria:**
- Caller continues immediately after Execute() (arch-v1.md 544)
- Async completion via channel
- Execute() returns before tool completion

### Test 4: TestParallelExecutor_MaxParallelLimit
```go
func TestParallelExecutor_MaxParallelLimit(t *testing.T) {
    policy := ExecutionPolicy{
        Mode:        "par_continue",
        MaxParallel: 2, // Limit to 2 concurrent
    }
    executor := NewParallelExecutor(policy)
    
    // Create 4 tools that take measurable time
    tools := []ToolCall{
        {Name: "tool1", Args: map[string]any{"delay": 100}},
        {Name: "tool2", Args: map[string]any{"delay": 100}},
        {Name: "tool3", Args: map[string]any{"delay": 100}},
        {Name: "tool4", Args: map[string]any{"delay": 100}},
    }
    
    startTime := time.Now()
    resultChan, err := executor.Execute(tools)
    if err != nil {
        t.Fatalf("Expected nil error, got %v", err)
    }
    
    // Collect all results
    for range resultChan {
    }
    
    elapsed := time.Since(startTime)
    
    // With MaxParallel=2 and 4 tools at 100ms each:
    // Should take ~200ms (2 batches of 2 tools)
    // If unlimited parallel, would take ~100ms
    if elapsed < 150*time.Millisecond {
        t.Errorf("Expected MaxParallel limit enforced (>150ms), took %v", elapsed)
    }
    if elapsed > 300*time.Millisecond {
        t.Errorf("Expected reasonable execution (<300ms), took %v", elapsed)
    }
}
```
**Acceptance Criteria:**
- MaxParallel limits concurrent execution (arch-v1.md 521)
- Default MaxParallel=8 (arch-v1.md 521)
- Excess tools queued and executed in batches

### Test 5: TestParallelExecutor_SessionContextInsertion
```go
func TestParallelExecutor_SessionContextInsertion(t *testing.T) {
    policy := PolicyParContinue
    executor := NewParallelExecutor(policy)
    
    tools := []ToolCall{
        {Name: "tool1", Args: map[string]any{"value": "result1"}},
        {Name: "tool2", Args: map[string]any{"value": "result2"}},
        {Name: "tool3", Args: map[string]any{"value": "result3"}},
    }
    
    resultChan, err := executor.Execute(tools)
    if err != nil {
        t.Fatalf("Expected nil error, got %v", err)
    }
    
    // Collect all results
    var results []ExecutionResult
    for result := range resultChan {
        results = append(results, result)
    }
    
    // Verify all results collected
    if len(results) != 3 {
        t.Errorf("Expected 3 results, got %d", len(results))
    }
    
    // Verify results can be inserted into session context
    sessionContext := make(map[string]any)
    for _, result := range results {
        sessionContext[result.ToolName] = result.Output
    }
    
    if sessionContext["tool1"] == nil {
        t.Error("Expected tool1 result in session context")
    }
    if sessionContext["tool2"] == nil {
        t.Error("Expected tool2 result in session context")
    }
    if sessionContext["tool3"] == nil {
        t.Error("Expected tool3 result in session context")
    }
}
```
**Acceptance Criteria:**
- Results inserted into session context (arch-v1.md 543)
- All tool outputs included
- Session context scoped correctly

## Dependencies

### Test Dependencies
```
Test 1 → Test 2 (Parallel spawning before concurrency)
Test 2 → Test 3 (Concurrency before async verification)
Test 4 (MaxParallel limit - independent)
Test 5 (Session context - depends on 1)
```

### Phase Dependencies
- **Phase 6.3** must complete first (OrchestratorService)
- **Phase 6.4** is sibling phase (SequentialExecutor)
- **Phase 6.6** depends on this phase (Sub-Agent support)

## Files to Create/Modify

| File | Operation | Changes |
|------|-----------|---------|
| `pkg/orchestrator/executor.go` | MODIFY | Add ParallelExecutor struct and Execute() method |

## Implementation Order (TDD)

1. **Test 1**: Write test → verify RED → Implement parallel state spawning → verify GREEN → commit
2. **Test 2**: Write test → verify RED → Implement concurrent execution → verify GREEN → commit
3. **Test 3**: Write test → verify RED → Implement async completion → verify GREEN → commit
4. **Test 4**: Write test → verify RED → Add MaxParallel limit → verify GREEN → commit
5. **Test 5**: Write test → verify RED → Implement session context insertion → verify GREEN → commit

**Total: 5 tests, 5 commits**

## Deliverables
- ✅ ParallelExecutor struct defined
- ✅ Parallel state spawning implemented
- ✅ Concurrent execution implemented
- ✅ MaxParallel enforcement implemented
- ✅ Async completion signaling implemented
- ✅ Result aggregation implemented
- ✅ 5 commits (one per test per TDD workflow)

## arch-v1.md Line Number References

| Feature | Line Numbers | Description |
|---------|--------------|-------------|
| par_continue pattern | 540-544 | Parallel, no-fail-on-fail execution |
| Parallel state spawning | 541 | Spawns parallel states for each tool-call |
| Concurrent execution | 542 | All execute concurrently |
| Session context insertion | 543 | Results inserted into session context |
| Async completion | 544 | Caller continues immediately |
| MaxParallel config | 521 | Default MaxParallel=8 |

---

## Appendix A: TDD Workflow Checklist

### Commit 1: TestParallelExecutor_ParContinueSpawnsParallelStates
- [ ] Create branch `feat/layer6-parallel-executor`
- [ ] Write test in `pkg/orchestrator/executor_test.go`
- [ ] Run test → verify RED (fails)
- [ ] Implement parallel state spawning in `pkg/orchestrator/executor.go`
- [ ] Run test → verify GREEN (passes)
- [ ] `git commit -m "feat(layer-06/parallel-executor): spawn parallel states for par_continue"`

### Commit 2: TestParallelExecutor_ParContinueExecutesConcurrently
- [ ] Write test in `pkg/orchestrator/executor_test.go`
- [ ] Run test → verify RED (fails)
- [ ] Implement concurrent execution with goroutines
- [ ] Run test → verify GREEN (passes)
- [ ] `git commit -m "feat(layer-06/parallel-executor): execute tools concurrently"`

### Commit 3: TestParallelExecutor_CallerContinuesImmediately
- [ ] Write test in `pkg/orchestrator/executor_test.go`
- [ ] Run test → verify RED (fails)
- [ ] Implement async completion via channel
- [ ] Run test → verify GREEN (passes)
- [ ] `git commit -m "feat(layer-06/parallel-executor): async completion signaling"`

### Commit 4: TestParallelExecutor_MaxParallelLimit
- [ ] Write test in `pkg/orchestrator/executor_test.go`
- [ ] Run test → verify RED (fails)
- [ ] Implement MaxParallel enforcement with semaphore
- [ ] Run test → verify GREEN (passes)
- [ ] `git commit -m "feat(layer-06/parallel-executor): MaxParallel enforcement"`

### Commit 5: TestParallelExecutor_SessionContextInsertion
- [ ] Write test in `pkg/orchestrator/executor_test.go`
- [ ] Run test → verify RED (fails)
- [ ] Implement result aggregation into session context
- [ ] Run test → verify GREEN (passes)
- [ ] `git commit -m "feat(layer-06/parallel-executor): result aggregation into session context"`

---

## Appendix B: Style Compliance

This phase plan follows established patterns:
- ✅ **1:1 test-to-commit ratio** (5 tests = 5 commits)
- ✅ **Descriptive test names** (`Test[Component]_[Behavior]_[ExpectedResult]`)
- ✅ **Small, focused phase** (5 tests within recommended 2-6 range)
- ✅ **Explicit dependencies** (Phase 6.3 required, Phase 6.6 depends on this)
- ✅ **Clear acceptance criteria** for each test
- ✅ **arch-v1.md line number references** (Section 8.3, lines 525-549)
- ✅ **TDD workflow explicitly documented** in Appendix A