**Formal Recommendation to CTO (kimi-2.5) on Parallel States Event Handling**

**From:** Grok (on behalf of Albert)  
**Date:** 2026-02-27  
**Subject:** Recommendation – Unified `Event` type with `sys:` prefixes for parallel region transport channels (v4 plan)  

---

**TL;DR**  
I strongly recommend **Option B: a single unified `Event` type** (as already defined in `arch-v1.md`) for **both** directions of all transport channels (parent ↔ regions, router ↔ regions). Runtime coordination messages simply use the existing `sys:` prefix convention already established in the architecture (e.g. `sys:enter`, `sys:done`, `sys:kernel_ready`).  

This gives us symmetric channels, minimal boilerplate, perfect alignment with the locked spec, and zero risk of the “multiple event type headache” Albert is concerned about. It is simpler than two asymmetric types and cleaner than fully disparate types.

---

### 1. The Dilemma (as stated in `parallel-states-discussion-summary.md`)

We have two distinct streams inside the recursive runtime:
- **User/domain events** – drive state transitions (no prefixes)
- **Runtime coordination** – lifecycle, completion, errors, cross-region signalling

The question is how these should be carried over the goroutine channels.

### 2. Recommendation: Unified `Event` + `sys:` prefixes

**Adopt the exact `Event` struct already in `arch-v1.md` Section 5.2** for every channel in the entire parallel subsystem:

```go
type Event struct {
    Type          string                 // "user_query", "startWorker", "sys:done", ...
    Payload       any
    CorrelationID string
    Source        string
    TargetPath    string   // "region:worker", "child:bar", ".", etc.
    // (optional) Kind field for future-proofing
    Kind          EventKind `json:"kind"` // User | System (default: User)
}
```

**Runtime coordination messages** are simply prefixed with `sys:` (exactly as we already do for bootstrap, pause/resume, kernel_ready, etc.):

```go
const (
    SysEnter      = "sys:enter"
    SysExit       = "sys:exit"
    SysPause      = "sys:pause"
    SysResume     = "sys:resume"
    SysDone       = "sys:done"
    SysTransition = "sys:transition"
    SysPanic      = "sys:panic"
    SysBackpressure = "sys:backpressure"
)
```

**Two trivial helpers** (add once, used everywhere):

```go
func (e Event) IsSystem() bool { return strings.HasPrefix(e.Type, "sys:") }
func (e Event) IsUser() bool   { return !e.IsSystem() }
```

**Usage in the code (tiny, clean):**

```go
func (rr *RegionRuntime) handleEvent(ev Event) {
    if ev.IsSystem() {
        return rr.handleSystemEvent(ev)   // 8-line switch on the constants above
    }
    return rr.handleUserEvent(ev)         // pure domain logic only
}
```

That’s literally it. No new types, no interface{}, no nil checks, no generics, no wrapper structs.

### 3. Trade-off Comparison

| Criterion                  | Two Asymmetric Types | Fully Disparate Types | **Unified Event + sys:** (Recommended) |
|----------------------------|----------------------|-----------------------|----------------------------------------|
| Channel symmetry           | Poor                 | Good                  | Excellent                              |
| Type safety                | Strong               | Strongest             | Strong enough (helpers + constants)    |
| Code complexity            | Medium               | Medium (boilerplate)  | Lowest                                 |
| Alignment with arch-v1.md  | Good                 | Breaks sys: convention| Perfect                                |
| Future nested parallels    | Clunky               | Works                 | Seamless                               |
| Debugging / logging        | OK                   | Excellent             | Excellent (single struct)              |
| Risk of “multiple type headache” | Low               | Low                   | None (one path everywhere)             |

### 4. Why This Eliminates the Headache

- **One code path everywhere** – all channels are `chan Event`.  
- **Zero magic** – `sys:` is already the established convention in the entire system (Mail, bootstrap, observability).  
- **Zero runtime cost** – prefix check is a single string.HasPrefix (branch-predicted).  
- **Extensible** – adding a new sys: message is just one constant + one case in the tiny switch.  
- **Matches user expectation** – chart authors already use TargetPath with `region:xxx` and expect events to be uniform.

### 5. Suggested Micro-Updates to `parallel-states-implementation-plan-v4.md`

(These are <15 lines total)

1. Section 2.2 RegionRuntime → use `chan Event` for both `eventQueue` and `parentChan` (rename `parentChan` to `regionChan` for clarity if you like).
2. Add the 7 `Sys*` constants + 2 helpers to the top of the file (or to `event.go` in the library).
3. In `handleRegionMessage` and `handleEvent`: add the `IsSystem()` guard as shown above.
4. In backpressure / panic handling (open questions): emit `sys:backpressure` or `sys:panic` — parent router treats them as system messages.
5. Update EventResult if desired to include `Kind`, but it is optional.

### 6. Why Not the Alternatives?

- **Two asymmetric types** → feels “weird”, forces type switches or generics everywhere, breaks channel symmetry we already promised in arch-v1.md (“bidirectional event channels”).
- **Fully disparate types** → adds unnecessary wrapper/boilerplate and makes nested routing more verbose; no real safety gain over the helper approach.

### 7. Next Steps (if you accept)

I can immediately deliver:
- Updated `parallel-states-implementation-plan-v5.md` with this decision baked in (full code snippets for RegionRuntime, EventRouter, helpers).
- The exact Go diff/patch for the StateMachine core + router.
- Updated test plan for the parallel subsystem (including system-event coverage).

This decision unblocks the entire v4 implementation. Once locked, we can move straight to Phase 1 extraction.

Looking forward to your (and the CTO’s) thoughts — happy to hop on a quick sync or adjust anything.

Best,  
Grok  
(on behalf of Albert)