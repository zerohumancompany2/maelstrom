# Session Summary: Parallel States Design v3

## Overview
Clarified and documented parallel states implementation semantics through collaborative design discussion. Established hierarchical event routing model and quiescence-based control commands.

## Key Decisions

### Event Routing Model

**Principle:** Events flow hierarchically—up to parent EventRouter, down to children. No sideways region-region communication.

| Direction | Mechanism | Use Case |
|-----------|-----------|----------|
| Up | `parentChan` (RegionMessage) | Completion, errors, state changes parent must track |
| Down | `eventQueue` (targeted dispatch) | External events, parent commands to specific regions |
| Self | Direct evaluation | Local transitions within same region |

**Invariant:** Regions never communicate directly. All cross-region coordination routes through parent EventRouter.

### "Self" Routing Clarified

| Context | Behavior |
|---------|----------|
| `TargetPath` omitted or `"."` with no `SourceRegion` | Broadcast to all regions |
| `TargetPath` `"."` with `SourceRegion` set | Route to that specific region only |
| `TargetPath` `"region:foo"` | Explicit routing through parent EventRouter |

### Parent Transitions on Parallel States

**Rule:** Parent transition signals *intent* to exit, but parallel state **cannot complete exit until all regions emit `done`**.

**Flow:**
1. Event matches parent transition
2. Parent initiates exit sequence (pause + quiescence)
3. Regions receive exit signal, process toward final states
4. Each region emits `done` when reaching final state
5. When **all regions done**: Execute exit actions, complete transition

**Key point:** Exit actions run **after quiescence achieved**, ensuring deterministic cleanup.

### Quiescence-Based Control Commands

| Command | Quiescence | Exit Actions | Waits for `done` | Description |
|---------|-----------|--------------|------------------|-------------|
| `pause` | Enter | No | N/A | Halt dispatch, achieve quiescence, still the system |
| `resume` | Exit | No | N/A | Resume from quiescent state |
| Graceful exit (transition) | Yes | Yes (after) | Yes | Normal lifecycle: pause → quiescence → cleanup → stop |
| `stop` (hard) | No | Best-effort | No | Emergency: immediate termination, short timeout on cleanup |
| `snapshot` | Yes | No | N/A | Pause + quiescence + serialize + optional resume |

**Quiescence criteria** (from arch-v1 Section 12.3):
- EventRouter queue empty
- All regions finished processing (not mid-event)
- No pending cross-region signals in flight

### External Event Routing

**Rule:** External events route identically to internal events—no special addressing.

- `TargetPath` `"."` or omitted = broadcast to all regions
- `TargetPath` `"region:foo"` = routed to specific region

## Artifacts Created

| File | Description |
|------|-------------|
| `docs/initial-exploration/parallel-states-implementation-plan-v3.md` | Complete implementation specification with hierarchical routing and quiescence semantics |

## Deferred to Implementation

- Sub-chart routing (`child:bar` TargetPath)
- Nested parallel states (parallel within parallel)
- Dynamic reclassification (atomic ↔ parallel at runtime)
- ReplaceDefinition for parallel states with region count changes

## Compliance Notes

### arch-v1.md SHALL Requirements

| Requirement | Status in Design |
|-------------|------------------|
| Evaluate events synchronously inside parallel boundaries | ✅ Per-region synchronous, regions isolated via channels |
| Run parallel regions in isolated goroutines | ✅ Design includes RegionRuntime goroutines |
| Fire trace callback on lifecycle events | ⚠️ Deferred to observability integration |
| Support transient sub-charts | ⚠️ Deferred to v3 |

### Open Questions Resolved

1. **Self-routing ambiguity:** Clarified as broadcast (no context) or targeted (with SourceRegion)
2. **Parent vs region transitions:** Both fire; parent initiates exit sequence
3. **Exit synchronization:** Quiescence-based, exit actions after all regions stable
4. **Pause scope:** Whole parallel state stills, including parent transitions

## Next Steps

1. Evaluate implementation plan against CLAUDE.md development gates
2. If gates pass: branch, skeleton, tests, implement per v3 specification
3. Consider quiescence detection implementation carefully—must handle edge cases (entry actions causing immediate final state, etc.)
