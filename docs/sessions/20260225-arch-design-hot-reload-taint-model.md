# Session: Architectural Design - Hot-Reload, Bootstrap, and Taint Model

**Date:** 2026-02-25

## Summary

Extended review of Maelstrom architecture focusing on three critical design areas: hot-reload semantics, kernel bootstrap sequence, and the taint/capability security model.

## Source Material

- `docs/arch-v1.md` - Base architecture specification
- `docs/initial-exploration/hot-reload-and-bootstrap-design.md` - Design notes created during session
- `docs/initial-exploration/boundary-immutability-design.md` - Security invariant specification created during session

## Key Decisions Made

### 1. Hot-Reload: Load-on-Next-Start

**Replaced in-flight replacement with clean restart using history mechanisms.**

- Charts restart with `shallowHistory` or `deepHistory` on version change
- Deleted states in deep history fall back to shallow history
- In-flight work: respect `stabilityPolicy`, wait for quiescence with timeout, force-stop if necessary
- Context migration: optional `contextTransform` template; on failure → `cleanStart`

**Rationale**: Simplifies implementation, avoids complex state migration, predictable behavior.

### 2. Kernel Bootstrap State Machine

**Explicit two-phase bootstrap with manual Kernel coordination.**

```
KernelStart
├── LoadLibrary
├── RegisterBootstrapActions
├── SpawnPreSecurity → DriveToReady (manual events)
├── SpawnPreCommunication → DriveToReady (manual events)
├── SpawnBootstrapChart → WaitFor kernel_ready (via Mail)
└── HandoffToRegistry → Kernel dormant
```

- "DriveToReady" is the only place Kernel touches Library internals directly
- Pre-bootstrap charts are minimal; full services use normal Mail/events

### 3. Snapshot Consistency

**Two-phase snapshot for parallel regions:**

1. Pause dispatch (finish current, queue new)
2. Capture parent queue + per-region in-flight queues
3. Resume

Restore: parent queue to parent; per-region in-flight re-dispatched. Accepts possible reordering (consistent with eventual consistency semantics).

### 4. Boundary Immutability (Core Security Invariant)

**Boundary is permanent identity.**

- Charts born into a boundary die there
- `(name, boundary)` is immutable; version can change
- Same name + different boundary = different security identity
- No migration path; create new chart identity if "boundary change" needed
- Sub-agent inheritance: can only spawn same or stricter boundary

**Attack prevented**: Registry compromise cannot downgrade chart to exfiltrate data.

### 5. Taint/Capability Model

**Hybrid approach (simplified for MVP):**

- **Taints**: Data lineage (what the data *is*): `["inner_only", "pii", "secret", "user_supplied", "tool_output"]`
- **Capabilities**: Access control derived from boundary level
- Policy enforcement: taint redaction at boundaries based on chart's boundary capability

**Performance**: Bloom filters for fast-path taint checking.

### 6. Token Budget Eviction

**Tiered strategy:**

```yaml
eviction:
  strategy: tiered
  tiers:
    - at: 0.9
      action: truncate_oldest
    - at: 1.0
      action: summarize_async
```

Acceptable tradeoff: one turn may have degraded context, subsequent turns improve.

## Outputs Created

- `docs/initial-exploration/hot-reload-and-bootstrap-design.md` - Complete design decisions
- `docs/initial-exploration/boundary-immutability-design.md` - Security invariant specification

## Open Items

- Context transform validation (at load time?)
- Two-phase snapshot failure recovery path
- Context quality signalling to agents (metadata field)
- Bloom filter implementation details (false positive rate tuning)

## Next Steps

1. Draft Go interfaces for Library and key Maelstrom services
2. Begin implementation of pure statechart library
3. Design test suite format for YAML-defined scenarios
