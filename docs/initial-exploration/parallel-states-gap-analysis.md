# Parallel States Gap Analysis

**Date:** 2026-02-27
**Scope:** Analysis of implementation completeness against arch-v1.md and parallel-states-implementation-plan-v3.md
**Status:** INCOMPLETE - Significant gaps identified

---

## Executive Summary

The current statechart implementation (`pkg/statechart/`) implements **atomic and compound states only**. Parallel states, while defined in the type system, are **not actually implemented**. The sophisticated hierarchical event routing system described in the v3 implementation plan does not exist in the codebase.

| Component | Status | Impact |
|-----------|--------|--------|
| Atomic states | Complete | Functional |
| Compound states | Complete | Functional |
| Parallel states | **NOT IMPLEMENTED** | Major gap - core architecture feature missing |
| EventRouter | **MISSING** | No hierarchical event routing |
| RegionRuntime | **MISSING** | No per-region goroutine isolation |
| Quiescence detection | **MISSING** | Required for hot-reload per arch-v1 |

---

## Detailed Gap Analysis

### 1. Current Implementation vs v3 Plan

#### 1.1 EventRouter (CRITICAL GAP)

**v3 Plan Specification:**
```go
type EventRouter struct {
    parentID     string
    regions      map[string]*RegionRuntime
    completion   map[string]bool
    parentChan   chan RegionMessage
}
```

**Current Implementation:**
- No `EventRouter` struct exists
- Single event queue per `ChartRuntime` (not per-region)
- No hierarchical routing

**Gap Impact:** Events cannot be routed between regions or from parent to specific regions. The `TargetPath` field in `Event` is completely ignored.

#### 1.2 RegionRuntime (CRITICAL GAP)

**v3 Plan Specification:**
```go
type RegionRuntime struct {
    name         string
    currentState string
    eventQueue   chan Event       // receives from parent router
    parentChan   chan RegionMessage  // sends to parent router
    actions      map[string]ActionFn
    guards       map[string]GuardFn
}
```

**Current Implementation:**
- No `RegionRuntime` struct exists
- Regions are merely entries in a `regionStates map[string]string`
- No goroutine isolation - all regions would share the same event processing loop

**Gap Impact:** Regions cannot run concurrently. No true parallel execution.

#### 1.3 TargetPath Routing (CRITICAL GAP)

**v3 Plan Event Addressing:**

| TargetPath | Semantics | Implementation |
|------------|-----------|----------------|
| `""` or `"."` | Self (broadcast to all regions if no context) | ❌ Ignored |
| `"region:foo"` | Specific region targeting | ❌ Ignored |
| `"child:bar"` | Sub-chart routing | ❌ Ignored |

**Current Code:**
The `TargetPath` field exists in `Event` struct but is never parsed or used in `processEvent()`.

#### 1.4 Parent-Region Message Protocol (CRITICAL GAP)

**v3 Plan Specification:**
```go
type RegionMessage struct {
    Type     string      // "done", "transition", "event", "error"
    Region   string
    NewState string
    Event    Event
    Error    error
}
```

**Current Implementation:**
- No `RegionMessage` type exists
- No protocol for regions to signal completion to parent
- No mechanism for cross-region event emission

**Gap Impact:** Parent cannot track region completion. Cannot implement "all regions done" exit condition.

#### 1.5 ActiveStates Representation (MEDIUM GAP)

**arch-v1 Required Format:**
```go
// StatePath format: "root/parallelNode/regionName/currentNode"
// Example: ["root/orient/shortTerm/conversationScan", "root/orient/longTerm/ragQuery"]
type ChartRuntime struct {
    activeStates []StatePath
}
```

**Current Implementation:**
```go
type ChartRuntime struct {
    activeState  string              // Singular - only one state
    regionStates map[string]string   // Flat map, no hierarchical paths
}
```

**Gap Impact:** Cannot represent full parallel state configuration per arch-v1 specification.

---

### 2. Test Coverage Analysis

#### 2.1 Current Test Status

**Existing Tests (`engine_parallel_test.go`):**
- Only **2 tests** (73 lines total)
- Both are basic stubs with minimal assertions
- One test has a placeholder (`_ = snap`) instead of actual validation
- Neither test verifies concurrent region execution

**Test Outline (`parallel-states-test-outline.md`):**
- 13 test categories defined
- ~60+ specific test cases specified
- Comprehensive coverage of: structure, routing, lifecycle, quiescence, control commands
- **Implementation: 0%** - None of these tests exist

#### 2.2 Test Coverage Matrix

| Test Category | Outline Count | Implemented | Status |
|--------------|---------------|-------------|--------|
| Parallel State Structure | 6 | 0 | ❌ Missing |
| Event Router | 7 | 0 | ❌ Missing |
| Region Runtime | 7 | 0 | ❌ Missing |
| Parallel Entry | 6 | 0 | ❌ Partial (basic spawn only) |
| Completion & Exit | 8 | 0 | ❌ Missing |
| Parent Transitions | 5 | 0 | ❌ Missing |
| Control Commands | 8 | 0 | ❌ Partial (only basic pause/resume) |
| Force-Stop Scenarios | 5 | 0 | ❌ Missing |
| Quiescence Detection | 5 | 0 | ❌ Missing |
| Full Lifecycle Integration | 5 | 0 | ❌ Missing |
| Error Handling | 5 | 0 | ❌ Missing |
| Property Tests | 5 | 0 | ❌ Missing |
| Concurrency Tests | 5 | 0 | ❌ Missing |

---

### 3. arch-v1 SHALL Statement Compliance

### 3.1 Section 14.1: Pure Statechart Library Contract

| SHALL Statement | Status | Evidence |
|-----------------|--------|----------|
| "Evaluate events synchronously inside parallel boundaries" | ❌ FAIL | No parallel boundaries implemented |
| "Run parallel regions in isolated goroutines" | ❌ FAIL | No RegionRuntime, no goroutine isolation |
| "Fire trace callback on every transition, entry, exit" | ❌ FAIL | `TraceCallback` defined but never invoked |
| "Support dynamic reclassification" | ❌ FAIL | Node type is computed but never changes |
| "Provide `replaceDefinition` hook" | ⚠️ PARTIAL | `CmdReplaceDefinition` exists but not implemented |
| "Support transient sub-charts" | ❌ FAIL | No sub-chart support |

### 3.2 Section 5.3: Library Guarantees (SHALL NOT)

| SHALL NOT Statement | Status | Evidence |
|---------------------|--------|----------|
| "SHALL NOT know about Mail, LLM, taints, boundaries" | ✅ PASS | Clean separation maintained |
| "SHALL NOT perform any I/O" | ✅ PASS | No I/O in library |
| "SHALL NOT implement persistence" | ✅ PASS | Only serialize/deserialize |
| "SHALL NOT have opinions on security policy" | ✅ PASS | No security logic |

---

### 4. Architecture Alignment Issues

#### 4.1 Compliant with arch-v1

- ✅ Node type derivation (0/1/≥2 children)
- ✅ Basic Library interface (Spawn, Control, Dispatch, Snapshot)
- ✅ Event type structure with TargetPath field
- ✅ ActionFn/GuardFn signatures match specification
- ✅ ApplicationContext interface definition

#### 4.2 Non-Compliant with arch-v1

| Requirement | arch-v1 Specification | Current Implementation | Gap |
|-------------|----------------------|------------------------|-----|
| ActiveStates | `[]StatePath` with format `"root/node/region/state"` | `activeState string` (singular) + `regionStates map` | Cannot represent parallel configuration |
| Quiescence | Defined in Section 12.3 for hot-reload | Simple mutex state check | No quiescence semantics |
| Trace callbacks | SHALL fire on all lifecycle events | Interface defined, never called | No observability |
| Event ordering | Synchronous within parallel boundaries | Single event queue | No parallel boundaries |

---

### 5. Critical Implementation Gaps (Priority Order)

#### Priority 1: Blocking - Core Parallel State Support

1. **EventRouter Implementation**
   - Hierarchical event routing (up/down)
   - TargetPath parsing (`region:foo`, `child:bar`)
   - Broadcast vs targeted dispatch

2. **RegionRuntime Implementation**
   - Per-region goroutines
   - Separate event queues per region
   - Parent channel communication

3. **State Path Refactoring**
   - Change `activeState string` to `activeStates []string`
   - Implement hierarchical path format
   - Update all path resolution logic

#### Priority 2: Required for Compliance

4. **Quiescence Detection**
   - Empty queue detection
   - In-flight event tracking
   - Pause/resume with quiescence

5. **Completion Tracking**
   - `done` message protocol
   - Final state detection
   - "All regions done" coordination

6. **Trace Callback Integration**
   - Invoke on every transition
   - Invoke on state entry/exit
   - Invoke on event dispatch

#### Priority 3: Advanced Features

7. **Dynamic Reclassification**
   - Atomic entry action can add children
   - Node type changes atomically
   - Stability policy enforcement

8. **Transient Sub-Charts**
   - Sub-chart lifecycle binding
   - Parent context sharing

---

### 6. Recommendations

#### 6.1 Immediate Actions (Before Parallel States)

1. **Complete compound state tests** - Ensure solid foundation
2. **Implement TraceCallback integration** - Required for observability
3. **Refactor ActiveStates to []string** - Data structure alignment

#### 6.2 Parallel States Implementation Order

**Phase 1: Infrastructure**
1. Implement `RegionMessage` type
2. Implement `EventRouter` with basic routing
3. Implement `TargetPath` parser

**Phase 2: Region Runtime**
1. Implement `RegionRuntime` struct
2. Implement per-region goroutines
3. Implement parent-region message protocol

**Phase 3: Integration**
1. Update `ChartRuntime` for parallel support
2. Implement parallel state entry/exit
3. Implement completion tracking

**Phase 4: Quiescence & Control**
1. Implement quiescence detection
2. Implement graceful exit
3. Implement snapshot with queue state

**Phase 5: Tests**
1. Implement all tests from test outline
2. Add property-based tests
3. Add concurrency stress tests

---

### 7. Files Requiring Changes

| File | Changes Required |
|------|------------------|
| `types.go` | Add `RegionMessage`, `RegionRuntime`, `EventRouter` types |
| `runtime.go` | Refactor `ChartRuntime` for parallel support |
| `engine.go` | Implement routing, quiescence, completion tracking |
| `engine_parallel_test.go` | Replace stub tests with full test suite |

---

## Conclusion

The current implementation provides a solid foundation for atomic and compound states but **does not implement parallel states** as specified in arch-v1 and the v3 implementation plan. The gap is significant:

- **0% of v3 EventRouter implemented**
- **0% of v3 RegionRuntime implemented**
- **0% of test outline implemented**
- **~17% of arch-v1 SHALL statements for parallel states passing**

To achieve compliance with the architecture specification, a substantial implementation effort is required focusing on hierarchical event routing, region goroutine isolation, and quiescence detection.

---

*Analysis generated: 2026-02-27*
*Based on: arch-v1.md Section 14, parallel-states-implementation-plan-v3.md, parallel-states-test-outline.md*
