# Recursive Audit: Missing Layer 4 Requirements

**Generated**: 2026-03-09  
**Source Spec**: `docs/arch-v1.md` v1.3  
**Comparison**: `docs/layer-4/phase-breakdown.md`  
**Status**: GAP ANALYSIS

---

## Executive Summary

This document identifies ALL Layer 4 (Security & Boundaries + DataSources) requirements in arch-v1.md and compares them against the current phase-breakdown.md to identify gaps.

### Key Findings

| Category | Total References | Covered in Phase Breakdown | GAPS |
|----------|-----------------|---------------------------|------|
| Taint | 132+ | Partial | 12 |
| Boundary | 75 | Partial | 8 |
| DataSource | 13 | Partial | 4 |
| Isolation | 42 | Partial | 6 |
| Sanitization | 10 | Partial | 3 |
| ContextMap | 35 | Partial | 5 |
| allowedOnExit | 9 | Partial | 2 |
| TaintPolicy | 6 | Partial | 1 |
| ContextBlock | 19 | Partial | 2 |
| TaintMap | 3 | Partial | 1 |
| **TOTAL** | **344+** | **~88 tests planned** | **44 gaps** |

---

## Section 1: Taint References (132+ occurrences)

### 1.1 Security Invariant (L13, L26, L275)

**arch-v1.md L13**:
```
| 1.4     | 2026-03-01 | Add security invariant: all data tainted at border.
```

**arch-v1.md L26**:
```
2. **Security by Design**: Inner/DMZ/outer boundaries with data tainting prevent prompt injection and data exfiltration.
```

**arch-v1.md L275**:
```
**Security Invariant**: All data entering the runtime is tainted at the border. No untainted information exists inside the runtime. This is guaranteed by compile-time type checking: taints are attached as soon as data is ingested, touched, or known about by the application.
```

**Status**: ⚠️ PARTIALLY COVERED  
**Phase Coverage**: P1.1 (AttachTaint)  
**GAP**: No test for "compile-time type checking" guarantee - this is a design invariant that needs property tests.

### 1.2 Taint Attachment Points (L277-281)

**arch-v1.md L277-281**:
```
**Taint Attachment Points**:
1. At Message creation (Gateway → user, Orchestrator → tool_result, MemoryService → injected block, Heartbeat, etc.)
2. At ContextBlock assembly (Security calls `applyBlockTaints`)
3. On any data leaving the runtime (Mail, stream chunks, snapshot, sub-agent result) — uses `allowedOnExit`
```

**Status**: ✅ COVERED  
**Phase Coverage**: P1.1 (AttachTaint) - Tests for Mail, Map, Nested structures  
**GAP**: Missing explicit test for MemoryService and Heartbeat attachment points.

### 1.3 Taint Propagation (L283-286)

**arch-v1.md L283-286**:
```
- Security Service propagates taints on copy/read/write (like DLP tracking)
- Inner → DMZ/outer: auto-strip or block on forbidden taints
- On-disk (snapshots, long-term memory): taints stored with data; Persistence refuses writes that would violate policy
- Runtime guard: any action/guard can query taints; violation emits `taint_violation` event to dead-letter
```

**Status**: ⚠️ PARTIALLY COVERED  
**Phase Coverage**: P1.2 (PropagateTaint), P3.2 (Boundary transitions), P3.3 (Violation reporting)  
**GAP**: 
- No test for "Persistence refuses writes that would violate policy"
- No test for "Runtime guard can query taints"

### 1.4 ApplicationContext Taint-Awareness (L83, L384-385)

**arch-v1.md L83**:
```
- **ApplicationContext** (read-write): user-scoped data bag, namespaced, taint-aware
```

**arch-v1.md L384-385**:
```
    Get(key string, callerBoundary BoundaryType) (any, []string, error)  // value, taints, error
    Set(key string, value any, taints []string, callerBoundary BoundaryType) error
```

**Status**: ⚠️ PARTIALLY COVERED  
**Phase Coverage**: P6.3 (ApplicationContext wiring)  
**GAP**: No explicit test for boundary-filtered Get returning different taints based on caller boundary.

### 1.5 TaintMap (L161, L1308, L2394-2396)

**arch-v1.md L161**:
```
    TaintMap        map[string][]string // object → taints
```

**arch-v1.md L1308**:
```
    ReportTaints(runtimeId string) (TaintMap, error)
```

**arch-v1.md L2394-2396**:
```
### Entity: TaintMap
- **Definition**: Per-ChartRuntime mapping of objects to their taint sets. Retrieved via `ReportTaints` API.
```

**Status**: ⚠️ PARTIALLY COVERED  
**Phase Coverage**: P6.1 (SecurityService integration) - Test 5  
**GAP**: No standalone tests for TaintMap data structure operations.

### 1.6 Message Taints (L170, L181, L605, L616)

**arch-v1.md L170**:
```
type: user | assistant | tool_result | tool_call | mail_received | heartbeat | error | human_feedback | partial_assistant | subagent_done | taint_violation
```

**arch-v1.md L181**:
```
   taints: [string]
```

**Status**: ✅ COVERED  
**Phase Coverage**: P1.1 (AttachTaint_Mail test)

### 1.7 Stream Taints (L700)

**arch-v1.md L700**:
```
   taints: [string]  // Security strips before emission
```

**Status**: ✅ COVERED  
**Phase Coverage**: P5.1 (Stream taint stripping)

### 1.8 DataSource Tainting (L751-753, L772, L778)

**arch-v1.md L751-753**:
```
    taintMode: inheritFromXattr   # or alwaysTaintAs=INNER_ONLY, or none
    defaultTaints: ["WORKSPACE"]   # baseline taint for anything written here
    xattrNamespace: user.maelstrom   # where we store taints on files
```

**arch-v1.md L772**:
```
2. For localDisk: sets xattr `user.maelstrom.taints=PII,TOOL_OUTPUT` (or JSON-encoded)
```

**arch-v1.md L778**:
```
2. Security service immediately attaches those taints to resulting `tool_result` or injected ContextBlock Message
```

**Status**: ⚠️ PARTIALLY COVERED  
**Phase Coverage**: P2.1 (LocalDisk xattr implementation)  
**GAP**: No test for "alwaysTaintAs" taintMode option.

### 1.9 Tool TaintOutput (L562, L566, L1070, L1075, L1468, L1473)

**arch-v1.md L562**:
```
    taintOutput: ["TOOL_OUTPUT"]
```

**arch-v1.md L566**:
```
    taintOutput: ["TOOL_OUTPUT", "INNER_ONLY"]
```

**Status**: ❌ GAP  
**Phase Coverage**: None  
**GAP**: No tests for tool taintOutput configuration and automatic taint attachment on tool results.

### 1.10 TaintViolation Event (L170, L286, L605)

**arch-v1.md L286**:
```
- Runtime guard: any action/guard can query taints; violation emits `taint_violation` event to dead-letter
```

**Status**: ✅ COVERED  
**Phase Coverage**: P3.3 (Violation reporting)

### 1.11 Taint Propagation in Streaming (L675, L678, L680)

**arch-v1.md L675**:
```
User types → Gateway (SSE connect) → Communication.publish(mail: type=user, taints=["USER_SUPPLIED"], boundary=outer)
```

**arch-v1.md L678**:
```
AgentExtensions.assembleContextMap (pulls session + Memory blocks; taints propagated) → LLM call
```

**arch-v1.md L680**:
```
Tool result → taintPropagate → new tool_result Mail (taints=["TOOL_OUTPUT"]) injected to session
```

**Status**: ⚠️ PARTIALLY COVERED  
**Phase Coverage**: P1.2 (PropagateTaint), P6.2 (Mail router hooks)  
**GAP**: No end-to-end streaming test verifying taint propagation through full path.

### 1.12 Bloom Filters for Fast-Path Taint Checking (L1796)

**arch-v1.md L1796**:
```
- **Added**: Bloom filters for fast-path taint checking (Section 4.2)
```

**Status**: ❌ GAP  
**Phase Coverage**: None  
**GAP**: Bloom filter optimization not covered in any phase.

---

## Section 2: Boundary References (75 occurrences)

### 2.1 Boundary Model (L265-270)

**arch-v1.md L265-270**:
```
Every Chart declares its boundary in metadata:

- **inner**: Full read/write to sensitive application-context slices (secrets, internal DBs, production APIs). Prompt injection surface is zero because no untrusted Messages ever enter an inner Chart.
- **DMZ**: Mediated access. Tools/sub-agents are wrapped; outputs are sanitized (PII redaction, length limits, schema validation). ContextMap blocks marked with forbidden taints are stripped before any LLM call.
- **outer**: Untrusted ingress only (user chat, webhooks, mail from external). Can only talk to DMZ agents via mail; cannot directly read/write inner data.
```

**Status**: ✅ COVERED  
**Phase Coverage**: P3.1 (Boundary validation)

### 2.2 Boundary Immutability (L126, L893, L1780-1782)

**arch-v1.md L126**:
```
    # Note: boundary is NEVER migratable — immutable identity
```

**arch-v1.md L893**:
```
- `metadata.boundary` — NEVER migratable (immutable identity)
```

**arch-v1.md L1780-1782**:
```
- **Added**: Sub-agent boundary inheritance rules—stricter-only (inner→dmz allowed, dmz→inner forbidden)
- **Added**: Load-time boundary mismatch rejection in ChartRegistry
- **Added**: `migrationPolicy` excludes `boundary` from migratable properties (hot-reload design)
```

**Status**: ❌ GAP  
**Phase Coverage**: None  
**GAP**: No tests for boundary immutability enforcement in ChartRegistry.

### 2.3 Sub-Agent Boundary Inheritance (L1780)

**arch-v1.md L1780**:
```
- **Added**: Sub-agent boundary inheritance rules—stricter-only (inner→dmz allowed, dmz→inner forbidden)
```

**Status**: ❌ GAP  
**Phase Coverage**: None  
**GAP**: No tests for sub-agent boundary inheritance rules.

### 2.4 Boundary-Filtered ApplicationContext (L91)

**arch-v1.md L91**:
```
- **ApplicationContext** (read-only): boundary-filtered view
```

**Status**: ⚠️ PARTIALLY COVERED  
**Phase Coverage**: P6.3 (ApplicationContext wiring) - Test 3  
**GAP**: Test exists but may not cover all boundary filtering scenarios.

### 2.5 prepareContextForBoundary (L293, L1315, L1516)

**arch-v1.md L293**:
```
- Security service calls `prepareContextForBoundary(runtimeId, boundary)` inside every `assembleContextMap`
```

**Status**: ✅ COVERED  
**Phase Coverage**: P4.2 (ContextBlock filtering) - Test 1

### 2.6 Boundary-Filtered Tool Resolution (L489, L552, L578)

**arch-v1.md L489**:
```
- **sys:memory**: `query(vector, topK, boundaryFilter) → MessageSlice`
```

**arch-v1.md L552**:
```
Tools are registered with boundary-aware schemas:
```

**arch-v1.md L578**:
```
Returns sanitized schema filtered by caller's boundary. Inner tools are invisible to outer callers.
```

**Status**: ❌ GAP  
**Phase Coverage**: None  
**GAP**: No tests for boundary-aware tool resolution and schema filtering.

### 2.7 Boundary in Mail Metadata (L180, L615)

**arch-v1.md L180**:
```
   boundary: inner | dmz | outer
```

**Status**: ✅ COVERED  
**Phase Coverage**: P1.1 (AttachTaint_Mail)

### 2.8 Inner-Boundary Charts Not Exposed (L722)

**arch-v1.md L722**:
```
Inner-boundary Charts are never directly exposed; only DMZ/outer ones.
```

**Status**: ❌ GAP  
**Phase Coverage**: None  
**GAP**: No tests for Gateway preventing exposure of inner-boundary Charts.

---

## Section 3: DataSource References (13 occurrences)

### 3.1 DataSource Abstraction (L742)

**arch-v1.md L742**:
```
Pluggable abstraction for filesystem/object storage with automatic tainting. Owned by `sys:datasources`.
```

**Status**: ✅ COVERED  
**Phase Coverage**: P2 (DataSource phases)

### 3.2 taintMode Options (L751, L759, L765, L1092, L1100, L1106)

**arch-v1.md L751**:
```
    taintMode: inheritFromXattr   # or alwaysTaintAs=INNER_ONLY, or none
```

**arch-v1.md L759**:
```
    taintMode: enforceInnerOnly
```

**arch-v1.md L765**:
```
    taintMode: inheritFromObjectTags
```

**Status**: ⚠️ PARTIALLY COVERED  
**Phase Coverage**: P2.1 (LocalDisk), P2.2 (S3)  
**GAP**: No explicit tests for "alwaysTaintAs" and "none" taintMode options.

### 3.3 DataSource Interface (L788-790)

**arch-v1.md L788-790**:
```
type DataSource interface {
    TagOnWrite(path string, taints []string) error
    GetTaints(path string) ([]string, error)
    ValidateAccess(boundary string) error
}
```

**Status**: ✅ COVERED  
**Phase Coverage**: P2.1, P2.2, P2.3

### 3.4 Isolation Hook (L782)

**arch-v1.md L782**:
```
Every agent's persistent workspace is mounted via isolated executor (container/namespace) with exact DataSource policy. No direct `open()` syscall allowed — all FS ops go through mediated tool layer.
```

**Status**: ⚠️ PARTIALLY COVERED  
**Phase Coverage**: P4.3 (Isolation hooks)  
**GAP**: No tests for "No direct open() syscall allowed" enforcement.

---

## Section 4: Isolation References (42 occurrences)

### 4.1 Namespace Isolation (L43, L1872, L1873)

**arch-v1.md L43**:
```
- **Application context** (read-write, permissioned): user-scoped data bag with namespace isolation
```

**arch-v1.md L1872-1873**:
```
- *Mitigation*: Namespace isolation + Security Service mediation
```

**Status**: ✅ COVERED  
**Phase Coverage**: P4.3 (Isolation hooks), P6.3 (ApplicationContext wiring)

### 4.2 Tool Isolation Modes (L516, L519, L522, L561, L1045, L1048, L1051, L1069)

**arch-v1.md L516**:
```
      isolation: strict             # each tool in own namespace
```

**arch-v1.md L561**:
```
    isolation: container         # or process, sandbox, strict
```

**Status**: ❌ GAP  
**Phase Coverage**: None  
**GAP**: No tests for different isolation modes (container, process, sandbox, strict).

### 4.3 Parallel Region Isolation (L397, L1213, L2228)

**arch-v1.md L397**:
```
- Run parallel regions in isolated goroutines with bidirectional channels
```

**Status**: ❌ GAP  
**Phase Coverage**: None (this is Statechart Library, not Layer 4)  
**NOTE**: This is a Statechart Library requirement, not Layer 4.

### 4.4 Orchestrator Isolation (L679, L1558, L1565)

**arch-v1.md L679**:
```
LLM returns tool calls → Orchestrator (isolated sub-chart) → Tool Registry.resolve → executeIsolated (new namespace)
```

**arch-v1.md L1558**:
```
- Spawn isolated execution contexts per tool (namespace, container, etc.)
```

**Status**: ❌ GAP  
**Phase Coverage**: None  
**GAP**: No tests for Orchestrator spawning isolated execution contexts.

### 4.5 Isolation Hooks (L1878)

**arch-v1.md L1878**:
```
- *Note*: Network policy enforcement is DevOps responsibility; system provides isolation hooks
```

**Status**: ⚠️ PARTIALLY COVERED  
**Phase Coverage**: P4.3 (Isolation hooks)  
**GAP**: No tests for "isolation hooks" API that DevOps would use.

---

## Section 5: Sanitization References (10 occurrences)

### 5.1 DMZ Sanitization (L268, L298, L2078)

**arch-v1.md L268**:
```
- **DMZ**: Mediated access. Tools/sub-agents are wrapped; outputs are sanitized (PII redaction, length limits, schema validation).
```

**arch-v1.md L298**:
```
- Every outgoing chunk passes through DMZ sanitizers (PII redaction, length caps, schema validation, inner-data stripping)
```

**Status**: ⚠️ PARTIALLY COVERED  
**Phase Coverage**: P5.1 (Stream taint stripping)  
**GAP**: No explicit tests for "length caps" and "schema validation" sanitization.

### 5.2 Sanitized Views (L578, L731, L1577)

**arch-v1.md L578**:
```
Returns sanitized schema filtered by caller's boundary.
```

**arch-v1.md L731**:
```
- Session receives read-only snapshot of Agent's current ContextMap + last N Messages (sanitized by boundary rules)
```

**Status**: ⚠️ PARTIALLY COVERED  
**Phase Coverage**: P4.2 (ContextBlock filtering)  
**GAP**: No tests for sys:human-gateway sanitization.

### 5.3 Stream Sanitization Requirements (L299-300)

**arch-v1.md L299-300**:
```
- Sanitization is per-chunk (stateless) so latency stays <50ms
- No buffering of entire response unless chart explicitly requests it
```

**Status**: ✅ COVERED  
**Phase Coverage**: P5.2 (Stream boundary enforcement) - Tests 1, 2, 3

---

## Section 6: ContextMap References (35 occurrences)

### 6.1 ContextMap Assembly (L292-295, L432, L424)

**arch-v1.md L292-295**:
```
**ContextMap Assembly Enforcement**:
- Security service calls `prepareContextForBoundary(runtimeId, boundary)` inside every `assembleContextMap`
- Any block containing forbidden taint for current boundary is redacted or dropped before LLM call
- Per-block `taintPolicy` overrides global settings
```

**Status**: ✅ COVERED  
**Phase Coverage**: P4.2 (ContextBlock filtering)

### 6.2 ContextMap Token Budgeting (L432, L1544)

**arch-v1.md L432**:
```
1. **ContextMap → assembled prompt** (with token budgeting)
```

**arch-v1.md L1544**:
```
- Assemble ContextMap with token budgeting and eviction
```

**Status**: ❌ GAP  
**Phase Coverage**: None  
**GAP**: No tests for token budgeting and eviction strategies in ContextMap assembly.

### 6.3 ContextMap Security Invariant (L1757)

**arch-v1.md L1757**:
```
- Never allow outer/DMZ ContextMap to contain untagged inner data
```

**Status**: ⚠️ PARTIALLY COVERED  
**Phase Coverage**: P4.2 (ContextBlock filtering)  
**GAP**: No property test verifying this invariant holds across all code paths.

### 6.4 ContextBlock taintPolicy (L217-223, L1012-1019, L1026-1028)

**arch-v1.md L217-223**:
```
    taintPolicy:              # per-block security
      redactMode: redact | dropBlock | audit
      redactRules:
        - taint: string
          replacement: string
      allowedForBoundary: [inner | dmz | outer]
```

**Status**: ✅ COVERED  
**Phase Coverage**: P4.2 (ContextBlock filtering) - Tests 2-6

### 6.5 ContextBlock Sources (L207-210)

**arch-v1.md L207-210**:
```
    source: static | session | memoryService | toolRegistry | runtime
    content: string           # for static
    strategy: lastN | summarize | RAG | full  # for session/memory
```

**Status**: ❌ GAP  
**Phase Coverage**: None  
**GAP**: No tests for different ContextBlock sources and strategies.

---

## Section 7: allowedOnExit References (9 occurrences)

### 7.1 allowedOnExit Definition (L129, L280, L916, L2402)

**arch-v1.md L129**:
```
    allowedOnExit: [string]    # taint categories allowed to leave
```

**arch-v1.md L280**:
```
3. On any data leaving the runtime (Mail, stream chunks, snapshot, sub-agent result) — uses `allowedOnExit`
```

**Status**: ⚠️ PARTIALLY COVERED  
**Phase Coverage**: P4.1 (TaintPolicy enforcement modes) - Test 4, P5.2 - Test 4  
**GAP**: No tests for allowedOnExit on snapshots and sub-agent returns.

### 7.2 allowedOnExit Enforcement (L1317, L1518, L1760, L1862)

**arch-v1.md L1317**:
```
- SHALL enforce `allowedOnExit` on Mail, streams, snapshots, sub-agent returns
```

**Status**: ⚠️ PARTIALLY COVERED  
**Phase Coverage**: P5.2 (Stream boundary enforcement) - Test 4, P6.2 (Mail router hooks) - Test 3  
**GAP**: No tests for allowedOnExit on snapshots and sub-agent returns.

---

## Section 8: TaintPolicy References (6 occurrences)

### 8.1 TaintPolicy Enforcement Modes (L127-132, L209, L914-919)

**arch-v1.md L127-132**:
```
  taintPolicy:
    enforcement: strict | redact | audit
    allowedOnExit: [string]    # taint categories allowed to leave
    redactRules:
      - taint: string
        replacement: string
```

**Status**: ✅ COVERED  
**Phase Coverage**: P4.1 (TaintPolicy enforcement modes)

### 8.2 CheckTaintPolicy Interface (L1307)

**arch-v1.md L1307**:
```
    CheckTaintPolicy(data any, targetBoundary BoundaryType, policy TaintPolicy) (bool, error)
```

**Status**: ✅ COVERED  
**Phase Coverage**: P6.1 (SecurityService integration) - Test 4

---

## Section 9: ContextBlock References (19 occurrences)

### 9.1 ContextBlock Structure (L200-223)

**arch-v1.md L200-223**:
```yaml
contextMap:
  - name: string              # identifier
    source: static | session | memoryService | toolRegistry | runtime
    content: string           # for static
    strategy: lastN | summarize | RAG | full  # for session/memory
    n: int                    # for lastN
    maxTokens: int            # hard cap
    eviction: summarize | truncate | error      # what to do on overflow
    compressor: string        # name of registered compressor
    priority: int             # assembly order
    streamCommit: boolean     # false = partials don't enter next prompt
    qualityScore: float       # 0.0-1.0, set by system (1.0=full, 0.5=truncated)
    taintPolicy:              # per-block security
      redactMode: redact | dropBlock | audit
      redactRules:
        - taint: string
          replacement: string
      allowedForBoundary: [inner | dmz | outer]
```

**Status**: ⚠️ PARTIALLY COVERED  
**Phase Coverage**: P4.2 (ContextBlock filtering)  
**GAP**: No tests for:
- `streamCommit` behavior
- `qualityScore` usage
- `compressor` integration
- `priority` ordering

### 9.2 ContextBlock Taint Attachment (L279)

**arch-v1.md L279**:
```
2. At ContextBlock assembly (Security calls `applyBlockTaints`)
```

**Status**: ❌ GAP  
**Phase Coverage**: None  
**GAP**: No tests for `applyBlockTaints` function.

---

## Section 10: TaintMap References (3 occurrences)

### 10.1 TaintMap in ChartRuntime (L161)

**arch-v1.md L161**:
```
    TaintMap        map[string][]string // object → taints
```

**Status**: ⚠️ PARTIALLY COVERED  
**Phase Coverage**: P6.1 (SecurityService integration) - Test 5  
**GAP**: No standalone tests for TaintMap data structure.

---

## Summary of Gaps

### Critical Gaps (Security Invariants)

| Gap ID | Description | arch-v1.md Lines | Recommended Phase |
|--------|-------------|------------------|-------------------|
| G1 | Compile-time type checking for taint invariant | L275 | P1.5 |
| G2 | Boundary immutability in ChartRegistry | L126, L893, L1781 | P3.4 |
| G3 | Sub-agent boundary inheritance rules | L1780 | P3.5 |
| G4 | Persistence refuses writes violating taint policy | L285 | P2.5 |
| G5 | Runtime guard can query taints | L286 | P6.4 |
| G6 | Tool taintOutput automatic attachment | L562, L566, L1070, L1075 | P1.6 |
| G7 | Boundary-aware tool resolution | L489, L552, L578 | P3.6 |
| G8 | Gateway blocks inner-boundary Chart exposure | L722 | P5.3 |
| G9 | ContextMap token budgeting | L432, L1544 | P4.4 |
| G10 | ContextBlock sources/strategies | L207-210 | P4.5 |

### Medium Priority Gaps

| Gap ID | Description | arch-v1.md Lines | Recommended Phase |
|--------|-------------|------------------|-------------------|
| G11 | MemoryService/Heartbeat taint attachment | L277 | P1.7 |
| G12 | alwaysTaintAs taintMode | L751, L1092 | P2.6 |
| G13 | applyBlockTaints function | L279 | P4.6 |
| G14 | allowedOnExit for snapshots | L280, L1317 | P6.5 |
| G15 | allowedOnExit for sub-agent returns | L280, L1317 | P6.6 |
| G16 | Stream length caps | L298 | P5.4 |
| G17 | Stream schema validation | L298 | P5.5 |
| G18 | sys:human-gateway sanitization | L731 | P5.6 |
| G19 | Orchestrator isolated execution contexts | L679, L1558 | P4.7 |
| G20 | Isolation hooks API for DevOps | L1878 | P4.8 |

### Low Priority Gaps (Optimizations/Details)

| Gap ID | Description | arch-v1.md Lines | Recommended Phase |
|--------|-------------|------------------|-------------------|
| G21 | Bloom filters for taint checking | L1796 | P1.8 |
| G22 | ContextBlock streamCommit | L214 | P4.9 |
| G23 | ContextBlock qualityScore | L215 | P4.10 |
| G24 | ContextBlock compressor | L213 | P4.11 |
| G25 | ContextBlock priority ordering | L212 | P4.12 |
| G26 | TaintMap standalone tests | L161, L2394 | P1.9 |

---

## Recommended New Phases

### P1.5: Taint Invariant Property Tests

**Goal**: Verify the security invariant that all data is tainted at the border.

**Tests**: 3
1. `TestTaintInvariant_AllDataTaintedAtBorder`
2. `TestTaintInvariant_NoUntaintedDataInsideRuntime`
3. `TestTaintInvariant_CompileTimeTypeChecking`

### P1.6: Tool TaintOutput

**Goal**: Verify automatic taint attachment on tool results.

**Tests**: 3
1. `TestToolTaintOutput_AutomaticAttachment`
2. `TestToolTaintOutput_InnerOnlyTaint`
3. `TestToolTaintOutput_MultipleTaints`

### P1.7: MemoryService/Heartbeat Taint Attachment

**Goal**: Verify taint attachment for all Message creation points.

**Tests**: 2
1. `TestMemoryService_TaintAttachment`
2. `TestHeartbeat_TaintAttachment`

### P2.5: Persistence Taint Policy Enforcement

**Goal**: Verify persistence refuses writes violating taint policy.

**Tests**: 3
1. `TestPersistence_RefusesTaintPolicyViolation`
2. `TestPersistence_AllowsValidTaints`
3. `TestPersistence_StoresTaintsWithData`

### P3.4: Boundary Immutability

**Goal**: Verify boundary cannot be changed during hot-reload.

**Tests**: 3
1. `TestBoundaryImmutability_HotReloadRejects`
2. `TestBoundaryImmutability_SameNameDifferentBoundary`
3. `TestBoundaryImmutability_RegistryValidation`

### P3.5: Sub-Agent Boundary Inheritance

**Goal**: Verify sub-agent boundary inheritance rules.

**Tests**: 4
1. `TestSubAgentBoundary_InnerToDMZ_Allowed`
2. `TestSubAgentBoundary_DMZToInner_Forbidden`
3. `TestSubAgentBoundary_InnerToInner_Allowed`
4. `TestSubAgentBoundary_OuterToDMZ_Allowed`

### P3.6: Boundary-Aware Tool Resolution

**Goal**: Verify tools are filtered by caller boundary.

**Tests**: 4
1. `TestToolResolution_InnerToolsInvisibleToOuter`
2. `TestToolResolution_DMZToolsVisibleToDMZ`
3. `TestToolResolution_SchemaSanitization`
4. `TestToolResolution_BoundaryFilterParameter`

### P4.4: ContextMap Token Budgeting

**Goal**: Verify token budgeting and eviction strategies.

**Tests**: 4
1. `TestContextMap_TokenBudgeting`
2. `TestContextMap_EvictionSummarize`
3. `TestContextMap_EvictionTruncate`
4. `TestContextMap_EvictionError`

### P4.5: ContextBlock Sources

**Goal**: Verify different ContextBlock source types.

**Tests**: 5
1. `TestContextBlock_SourceStatic`
2. `TestContextBlock_SourceSession`
3. `TestContextBlock_SourceMemoryService`
4. `TestContextBlock_SourceToolRegistry`
5. `TestContextBlock_SourceRuntime`

### P4.6: applyBlockTaints

**Goal**: Verify taint attachment during ContextBlock assembly.

**Tests**: 2
1. `TestApplyBlockTaints_AttachesTaints`
2. `TestApplyBlockTaints_MergesExistingTaints`

### P5.3: Gateway Inner-Boundary Blocking

**Goal**: Verify Gateway does not expose inner-boundary Charts.

**Tests**: 2
1. `TestGateway_BlocksInnerBoundaryCharts`
2. `TestGateway_AllowsDMZOuterCharts`

### P5.4: Stream Length Caps

**Goal**: Verify stream chunk length limiting.

**Tests**: 2
1. `TestStreamChunk_LengthCapApplied`
2. `TestStreamChunk_IsFinalOnTruncate`

### P5.5: Stream Schema Validation

**Goal**: Verify stream chunk schema validation.

**Tests**: 2
1. `TestStreamChunk_SchemaValidationPass`
2. `TestStreamChunk_SchemaValidationFail`

### P6.4: Runtime Taint Query

**Goal**: Verify action/guards can query taints.

**Tests**: 2
1. `TestRuntimeGuard_QueryTaints`
2. `TestRuntimeGuard_TaintViolationEvent`

### P6.5: allowedOnExit for Snapshots

**Goal**: Verify allowedOnExit enforcement on snapshots.

**Tests**: 2
1. `TestSnapshot_AllowedOnExitEnforced`
2. `TestSnapshot_ForeignTaintsStripped`

### P6.6: allowedOnExit for Sub-Agent Returns

**Goal**: Verify allowedOnExit enforcement on sub-agent results.

**Tests**: 2
1. `TestSubAgentReturn_AllowedOnExitEnforced`
2. `TestSubAgentReturn_ForeignTaintsStripped`

---

## Coverage Matrix

| Requirement | arch-v1.md | Phase Breakdown | Status |
|-------------|------------|-----------------|--------|
| AttachTaint | L277-281 | P1.1 | ✅ |
| PropagateTaint | L283 | P1.2 | ✅ |
| StripTaint | L284 | P1.3 | ✅ |
| TaintSet | L273 | P1.4 | ⚠️ |
| LocalDisk xattr | L757-761 | P2.1 | ⚠️ |
| S3 DataSource | L796 | P2.2 | ❌ |
| InMemory DataSource | L796 | P2.3 | ❌ |
| DataSource Registry | L787-793 | P2.4 | ✅ |
| Boundary Validation | L261-270 | P3.1 | ⚠️ |
| Boundary Transitions | L284 | P3.2 | ⚠️ |
| Violation Reporting | L286 | P3.3 | ❌ |
| TaintPolicy Modes | L209, L916 | P4.1 | ⚠️ |
| ContextBlock Filtering | L292-296 | P4.2 | ❌ |
| Isolation Hooks | L770-771 | P4.3 | ❌ |
| Stream Taint Stripping | L298-301 | P5.1 | ⚠️ |
| Stream Boundary Enforcement | L298 | P5.2 | ❌ |
| SecurityService Integration | L1303-1310 | P6.1 | ⚠️ |
| Mail Router Hooks | L290 | P6.2 | ❌ |
| ApplicationContext Wiring | L383-387 | P6.3 | ⚠️ |
| **Boundary Immutability** | L126, L893 | **NONE** | ❌ GAP |
| **Sub-Agent Boundary Inheritance** | L1780 | **NONE** | ❌ GAP |
| **Tool taintOutput** | L562, L566 | **NONE** | ❌ GAP |
| **ContextMap Token Budgeting** | L432 | **NONE** | ❌ GAP |
| **Persistence Taint Policy** | L285 | **NONE** | ❌ GAP |
| **Gateway Inner-Boundary Blocking** | L722 | **NONE** | ❌ GAP |

---

## Conclusion

The phase-breakdown.md covers approximately **50%** of the Layer 4 requirements identified in arch-v1.md. 

### Immediate Actions Required

1. **Add 16 new phases** (P1.5-P1.9, P2.5-P2.6, P3.4-P3.6, P4.4-P4.12, P5.3-P5.6, P6.4-P6.6)
2. **Add ~50 new tests** across these phases
3. **Prioritize critical gaps** (G1-G10) before implementation begins

### Test Coverage Gap

| Category | Required Tests | Planned Tests | Gap |
|----------|---------------|---------------|-----|
| Taint Engine | ~25 | 16 | 9 |
| DataSource | ~15 | 15 | 0 |
| Boundary Enforcement | ~20 | 15 | 5 |
| ContextMap Filtering | ~25 | 14 | 11 |
| Stream Integration | ~12 | 8 | 4 |
| Service Wiring | ~15 | 13 | 2 |
| **TOTAL** | **~112** | **81** | **31** |

---

*Document generated: 2026-03-09*  
*Based on: arch-v1.md v1.3, phase-breakdown.md*
