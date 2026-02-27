# Parallel States Implementation: Discussion Summary

**Date:** 2026-02-27
**Context:** Review of v4 implementation plan for CEO discussion

---

## Background

Created `parallel-states-implementation-plan-v4.md` which proposes a **recursive runtime pattern**:
- A parallel region is a `StateMachine` core wrapped in a goroutine
- `EventRouter` mediates all communication (hierarchical routing)
- Clean separation: user events vs runtime coordination protocol

---

## Open Question: Channel Types

### The Dilemma

**Option A: Two types (asymmetric)**
- Input: `eventQueue chan Event` (parent → region)
- Output: `parentChan chan RegionMessage` (region → parent)
- Different types for different concerns

**Option B: Unified `Event` type**
- Both directions use `chan Event`
- Lifecycle discrimination via prefix: `sys:done`, `sys:transition`
- Matches arch-v1's use of `sys:*` prefixes for services

### Technical Considerations

| Aspect | Two Types | Unified Event |
|--------|-----------|---------------|
| Type safety | Strong (compiler checked) | Weak (runtime prefix check) |
| Channel symmetry | Asymmetric (feels weird) | Symmetric (both `Event`) |
| arch-v1 alignment | Compliant | Compliant (bidirectional channels) |
| User event purity | Clean (no magic prefixes) | Contaminated (`sys:` convention) |
| Implementation | More complex | Simpler |

### Key Insight

The statechart library has two distinct event streams:
1. **User events** - Drive state transitions (no prefixes)
2. **Runtime coordination** - Region lifecycle, completion, errors

The question is whether these should be the same type in transport channels.

---

## Recommendation for CEO Discussion

**Primary Question:** Should we use a unified `Event` type with `sys:` prefixing, or keep separate types for user events vs runtime coordination?

**Context:**
- v4 implementation plan is otherwise solid
- This is the last open question before implementation can begin
- Decision affects channel signatures and EventRouter implementation

**Trade-offs to Discuss:**
1. Type safety vs channel symmetry
2. arch-v1 magic prefix convention vs explicit types
3. Future extensibility (sub-charts, nested parallel states)

---

## Current Blocker

Pending CEO input on channel type decision before v4 plan can be finalized.
