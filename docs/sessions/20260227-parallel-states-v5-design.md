# Session Summary: Parallel States v5 Design - Unified Event Transport

**Date:** 2026-02-27
**Participants:** Albert (CEO), Grok (CTO)
**Topic:** Finalize parallel states architecture with recursive runtime pattern and unified event channels

---

## Overview

Completed the parallel states design by resolving the event transport question. Established **recursive runtime pattern** with **unified `Event` type** for all channel communication, using `sys:` prefixes for runtime coordination.

---

## Key Decisions

### 1. Gap Analysis Completed

Analyzed current implementation against arch-v1.md requirements:

| Component | Status | Gap |
|-----------|--------|-----|
| Atomic states | ✅ Complete | None |
| Compound states | ✅ Complete | None |
| Parallel states | ❌ Not implemented | No EventRouter, RegionRuntime |
| Test coverage | ❌ 0% of test outline | Only 2 stub tests exist |

**Critical finding:** Current code treats all states as single active state. Parallel regions require goroutine isolation and hierarchical event routing.

### 2. Recursive Runtime Pattern Approved

**Core insight:** A parallel region is just a runtime wrapped in a goroutine.

```
ChartRuntime (parallel parent)
├── EventRouter (coordinates regions)
├── RegionRuntime A (goroutine + StateMachine core)
├── RegionRuntime B (goroutine + StateMachine core)
└── RegionRuntime N (goroutine + StateMachine core)
```

**Benefits:**
- Same `StateMachine` core for top-level and regions
- Clean separation: state logic vs coordination
- Sub-charts (future) follow same pattern

### 3. Unified Event Transport (Major Decision)

**Resolved:** All channels use `chan Event` (symmetric).

**Discrimination via `sys:` prefix** (matching arch-v1 convention):

```go
// System events (runtime coordination)
SysEnter      = "sys:enter"
SysExit       = "sys:exit"
SysDone       = "sys:done"
SysTransition = "sys:transition"
SysPanic      = "sys:panic"

// Helpers
func (e Event) IsSystem() bool { return strings.HasPrefix(e.Type, "sys:") }
func (e Event) IsUser() bool   { return !e.IsSystem() }
```

**Rejected alternative:** Two asymmetric types (`chan Event` + `chan RegionMessage`)

| Criterion | Unified Event (Chosen) | Two Types |
|-----------|----------------------|-----------|
| Channel symmetry | ✅ Excellent | ❌ Poor |
| Type safety | Good (helpers) | Strong |
| Code complexity | ✅ Lowest | Medium |
| Alignment with arch-v1 | ✅ Perfect (sys: convention) | Breaks convention |

### 4. EventRouter Simplification

**v4 approach:** Two input channels (`incoming chan Event`, `regionChan chan RegionMessage`)
**v5 approach:** Single channel (`inputChan chan Event`)

Router distinguishes by checking `IsSystem()` + `Source` field:
- System events from regions → coordination handling
- User events → routing based on `TargetPath`

### 5. System Event Protocol

| Direction | Events | Purpose |
|-----------|--------|---------|
| Parent → Region | `sys:enter`, `sys:exit`, `sys:pause`, `sys:resume` | Lifecycle control |
| Region → Parent | `sys:done`, `sys:transition`, `sys:panic` | Coordination signals |

---

## Artifacts Created

| File | Description |
|------|-------------|
| `docs/initial-exploration/parallel-states-gap-analysis.md` | Comprehensive analysis of implementation gaps vs arch-v1 |
| `docs/initial-exploration/parallel-states-implementation-plan-v4.md` | Recursive runtime pattern with asymmetric channels |
| `docs/initial-exploration/parallel-states-discussion-summary.md` | Summary of event transport dilemma for CEO discussion |
| `docs/initial-exploration/parallel-states-discussion-review.md` | CTO recommendation for unified Event approach |
| `docs/initial-exploration/parallel-states-implementation-plan-v5.md` | **Final approved design** with unified Event transport |

---

## Implementation Path Forward

### Phase 1: State Machine Core Extraction
Extract reusable `StateMachine` from `engine.go`:
- `ProcessEvent()` method
- `executeEntryActions()`, `executeExitActions()`
- Returns `EventResult` for coordination

### Phase 2: RegionRuntime
- Goroutine with symmetric `chan Event` I/O
- System event handling (`handleSystemEvent()`)
- User event processing through StateMachine core

### Phase 3: EventRouter
- Single input channel design
- `IsSystem()` discrimination
- TargetPath-based routing

### Phase 4: Integration
- Parallel state detection in `ChartRuntime`
- `enterParallelState()` with region spawning
- Quiescence-based exit

---

## Compliance with arch-v1.md

| SHALL Requirement | v5 Design Status |
|-------------------|------------------|
| Evaluate events synchronously inside parallel boundaries | ✅ Per-region synchronous via StateMachine core |
| Run parallel regions in isolated goroutines | ✅ RegionRuntime per goroutine |
| Bidirectional channels | ✅ Unified `chan Event` (symmetric) |
| Fire trace callback on lifecycle events | ⚠️ Deferred to observability phase |
| Support transient sub-charts | ⚠️ Deferred to v3+ |

---

## Open Questions Resolved

1. **Event transport type:** Unified `Event` with `sys:` prefixes ✅
2. **Channel symmetry:** Both directions use `chan Event` ✅
3. **Region→Parent protocol:** System events via output channel ✅
4. **Backpressure:** `sys:backpressure` event (implementation detail) ✅

---

## Next Steps

1. **Begin Phase 1:** Extract StateMachine core from `engine.go`
2. **Branch:** Create feature branch for parallel states implementation
3. **Test-first:** Implement test outline from `parallel-states-test-outline.md`
4. **Incremental:** Phase 1 → 2 → 3 → 4 with tests at each phase

**Ready to begin implementation.**
