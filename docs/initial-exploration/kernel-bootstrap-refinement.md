# Bootstrap Design Notes

Discussion date: 2026-02-25
Status: Design decisions for implementation

---

## 2. Kernel Bootstrap State Machine

### Problem
Core services (Security, Communication, Observability, Lifecycle) are Charts but need to function before the full bootstrap completes.

### Solution
Kernel manually drives pre-bootstrap charts through initialization, then hands off to Mail-based coordination.

### Detailed State Sequence

```
KernelStart
├── LoadLibrary              # Instantiate pure statechart library
├── RegisterBootstrapActions # sysInit, securityBootstrap, etc.
│
├── SpawnPreSecurity         # Minimal Security chart (hard-coded YAML)
│   └── DriveToReady         # Kernel manually pushes events:
│       - init → boundary_check → boundary_ready
│
├── SpawnPreCommunication    # Minimal Communication chart
│   └── DriveToReady         # Kernel manually pushes events:
│       - init → transport_setup → transport_ready
│
├── SpawnBootstrapChart      # Full YAML-defined bootstrap chart
│   └── WaitForEvent         # Now Mail works; kernel listens
│
└── HandoffToRegistry        # On kernel_ready event, go dormant
```

### Key Points
- "DriveToReady" is the ONLY place Kernel touches Library internals directly
- Pre-bootstrap charts are minimal—just enough for Mail to function
- All subsequent services use normal Mail/events
- Kernel listens for shutdown signals while dormant

---

## 3. Snapshot Consistency (Two-Phase)

### Problem
- Parent-only queue loses in-flight events dispatched to parallel regions
- Per-region queues have ordering hazards on restore

### Solution: Two-Phase Snapshot

#### Phase 1: Pause
- Finish processing current event in each region
- Queue (don't dispatch) new incoming events

#### Phase 2: Capture
- Parent event queue
- Per-region *in-flight* queues (events dispatched but not yet processed)

#### Phase 3: Resume
- Normal operation continues

### Restore Behavior
- Parent queue → restored to parent
- Per-region in-flight → re-dispatched to those regions
- **Acceptable hazard**: Relative ordering between parent and child may shift
- Consistent with "no cross-region ordering guarantee" semantics

---

## 4. Taint Performance: Bloom Filters

### Opportunity
Bloom filters provide fast-path taint checking:
- "Definitely not in set" → skip expensive scan
- "Probably in set" → do full scan
- False positives are safe (just mean extra work)

### Application
- Pre-compute bloom filters for ContextBlocks at assembly time
- Per-path filter for each data source (session, file, context block)
- Check filter before expensive taint propagation

---

## 5. Token Budget: Tiered Eviction

### Problem
Summarization may blow budget while trying to save it.

### Solution: Configurable Tiered Strategy

```yaml
eviction:
  strategy: tiered
  tiers:
    - at: 0.9              # 90% of maxTokens
      action: truncate_oldest
    - at: 1.0
      action: summarize_async  # map-reduce in background
  fallback: error
```

### Tradeoff Acceptance
- One turn may have "not great" context (truncated)
- Subsequent turns get better context (summary ready)
- Context quality signalled in metadata for agent adaptation

---

## Open Questions

1. Should context transforms be validated at chart load time?
2. Should we version the context shape independently of chart version?
3. What's the recovery path if two-phase snapshot fails mid-capture?
4. How do we signal context quality degradation to the agent?

---

## Related Sections in arch-v1.md

- Section 3.1 (ChartDefinition) - migrationPolicy addition
- Section 12 (Bootstrap Sequence) - refined state machine
- Section 4.2 (Data Tainting) - Bloom filter optimization note
- Section 3.5 (ContextBlock) - tiered eviction expansion
