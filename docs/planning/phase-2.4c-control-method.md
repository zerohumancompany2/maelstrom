# Phase 2.4C: Control Method

## Goal
Implement Control() method to forward lifecycle commands to the engine, following strict TDD workflow.

## Scope
- Implement Control() with CmdStart support
- Implement Control() with CmdStop support
- Implement error handling for not found runtimes
- 3 tests, 3 commits
- **NO breaking changes** - new method addition

## Current Status

### Implementation Status
| Item | Status | Notes |
|------|--------|-------|
| Control() method | ❌ Missing | Method doesn't exist |
| CmdStart forwarding | ❌ Missing | No implementation |
| CmdStop forwarding | ❌ Missing | No implementation |
| NotFound error handling | ❌ Missing | No error handling |

### Files Status
| File | Status |
|------|--------|
| `pkg/services/lifecycle/service.go` | ⚠️ Partial - add Control() method |

## Required Implementation

### Control Method
```go
// pkg/services/lifecycle/service.go
func (l *LifecycleService) Control(id statechart.RuntimeID, cmd statechart.ControlCmd) error {
    if l.engine == nil {
        return statechart.ErrRuntimeNotFound
    }
    
    return l.engine.Control(id, cmd)
}
```

### ControlCmd Type (from statechart package)
```go
// pkg/statechart/types.go (already exists)
type ControlCmd string

const (
    CmdStart         ControlCmd = "start"
    CmdStop          ControlCmd = "stop"
    CmdPause         ControlCmd = "pause"
    CmdResume        ControlCmd = "resume"
    CmdSnapshot      ControlCmd = "snapshot"
    CmdInjectEvent   ControlCmd = "injectEvent"
    CmdReplaceDef    ControlCmd = "replaceDefinition"
)
```

## Tests to Write (3 tests, 3 commits)

### Test 1: TestLifecycleService_ControlStart
```go
func TestLifecycleService_ControlStart(t *testing.T)
```
**Acceptance Criteria:**
- Control() with CmdStart succeeds when engine is wired
- Returns nil error
- Command is forwarded to engine

### Test 2: TestLifecycleService_ControlStop
```go
func TestLifecycleService_ControlStop(t *testing.T)
```
**Acceptance Criteria:**
- Control() with CmdStop succeeds when engine is wired
- Returns nil error
- Command is forwarded to engine

### Test 3: TestLifecycleService_ControlNotFoundReturnsError
```go
func TestLifecycleService_ControlNotFoundReturnsError(t *testing.T)
```
**Acceptance Criteria:**
- Control() with non-existent runtime ID returns error
- Error is statechart.ErrRuntimeNotFound or similar
- Service without engine returns error immediately

## Dependencies

### Test Dependencies
```
Test 1 (Control CmdStart) → Independent
Test 2 (Control CmdStop) → Independent
Test 3 (Control NotFound) → Independent
```

All tests are independent and can be executed in any order.

### Phase Dependencies
- **Phase 2.4A** must complete first (provides engine wiring)
- **Phase 2.4B** can run in parallel (List() is independent of Control())

## Files to Create/Modify

| File | Operation | Changes |
|------|-----------|---------|
| `pkg/services/lifecycle/service.go` | MODIFY | Add Control() method that forwards to engine |

## Implementation Order (TDD)

1. **Test 1**: Write test → verify RED → Implement Control() with CmdStart forwarding → verify GREEN → commit
2. **Test 2**: Write test → verify RED → Implement Control() with CmdStop forwarding → verify GREEN → commit
3. **Test 3**: Write test → verify RED → Add error handling for missing runtime → verify GREEN → commit

**Total: 3 tests, 3 commits**

## Test Implementation Examples

### Test 1: ControlStart
```go
func TestLifecycleService_ControlStart(t *testing.T) {
    engine := statechart.NewEngine()
    svc := NewLifecycleService(engine)
    
    // Spawn a runtime first
    def := statechart.ChartDefinition{
        ID:         "test-chart",
        Version:    "1.0.0",
        InitialState: "idle",
    }
    rtID, err := svc.Spawn(def)
    if err != nil {
        t.Fatalf("Spawn failed: %v", err)
    }
    
    // Test Control with CmdStart
    err = svc.Control(rtID, statechart.CmdStart)
    if err != nil {
        t.Errorf("Expected Control(CmdStart) to return nil, got %v", err)
    }
}
```

### Test 2: ControlStop
```go
func TestLifecycleService_ControlStop(t *testing.T) {
    engine := statechart.NewEngine()
    svc := NewLifecycleService(engine)
    
    def := statechart.ChartDefinition{
        ID:         "test-chart",
        Version:    "1.0.0",
        InitialState: "idle",
    }
    rtID, err := svc.Spawn(def)
    if err != nil {
        t.Fatalf("Spawn failed: %v", err)
    }
    
    err = svc.Control(rtID, statechart.CmdStop)
    if err != nil {
        t.Errorf("Expected Control(CmdStop) to return nil, got %v", err)
    }
}
```

### Test 3: ControlNotFound
```go
func TestLifecycleService_ControlNotFoundReturnsError(t *testing.T) {
    engine := statechart.NewEngine()
    svc := NewLifecycleService(engine)
    
    // Use non-existent runtime ID
    err := svc.Control(statechart.RuntimeID("non-existent"), statechart.CmdStart)
    if err == nil {
        t.Error("Expected Control with non-existent ID to return error")
    }
    
    // Also test without engine
    svcNoEngine := NewLifecycleService()
    err = svcNoEngine.Control(statechart.RuntimeID("any"), statechart.CmdStart)
    if err == nil {
        t.Error("Expected Control without engine to return error")
    }
}
```

## Deliverables
- ✅ `Control()` method implemented
- ✅ CmdStart forwarding to engine
- ✅ CmdStop forwarding to engine
- ✅ Error handling for not found runtimes
- ✅ 3 commits (one per test per TDD workflow)
- ✅ NO breaking changes - new method addition