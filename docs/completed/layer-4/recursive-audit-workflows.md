# Layer 4: Recursive Workflow Coverage Audit

**Generated**: 2026-03-09  
**Source Specs**: 
- `docs/arch-v1.md` v1.3 (Section 4: L277-301 - workflows section)
- `docs/layer-4/spec-extraction.md` (Section 4: Key Workflows)
- `docs/layer-4/phase-breakdown.md`

**Status**: Audit complete - CRITICAL GAPS IDENTIFIED

---

## Executive Summary

**CRITICAL FINDING**: All 6 workflows from the spec have **ZERO E2E tests**. The phase-breakdown.md defines 88 tests, but all are unit or integration tests. No end-to-end workflow tests exist.

| Metric | Value |
|--------|-------|
| Total Workflows in Spec | 6 |
| Workflows with E2E Tests | 0 |
| Workflows with Only Unit Tests | 6 |
| Total Unit/Integration Tests Planned | 88 |
| Total E2E Tests Planned | 0 |

---

## Workflow Coverage Table

| Workflow | Spec Reference | Covering Phase(s) | E2E Tests | Status |
|----------|---------------|-------------------|-----------|--------|
| **Data Tainting Workflow** (3 attachment points) | spec-extraction.md L248-269<br>arch-v1.md L277-287 | P1.1 (AttachTaint)<br>P1.2 (PropagateTaint)<br>P1.3 (StripTaint) | **NONE** | ❌ MISSING |
| **Boundary Enforcement Workflow** (Mail-only transitions) | spec-extraction.md L271-281<br>arch-v1.md L288-302 | P3.1 (Boundary validation)<br>P3.2 (Boundary transitions)<br>P3.3 (Violation reporting)<br>P6.2 (Mail router hooks) | **NONE** | ❌ MISSING |
| **Stream Sanitization Workflow** (<50ms latency) | spec-extraction.md L283-287<br>arch-v1.md L298-301 | P5.1 (Stream taint stripping)<br>P5.2 (Stream boundary enforcement) | **NONE** (1 benchmark for latency) | ❌ MISSING |
| **DataSource Write Workflow** (taint attachment) | spec-extraction.md L289-296<br>arch-v1.md L757-761 | P2.1 (LocalDisk)<br>P2.2 (S3)<br>P2.3 (InMemory)<br>P2.4 (Registry) | **NONE** | ❌ MISSING |
| **DataSource Read Workflow** (taint retrieval) | spec-extraction.md L298-303<br>arch-v1.md L764-767 | P2.1 (LocalDisk)<br>P2.2 (S3)<br>P2.3 (InMemory) | **NONE** | ❌ MISSING |
| **Isolation Hook Behavior** | spec-extraction.md L305-309<br>arch-v1.md L770-771 | P4.3 (Isolation hooks) | **NONE** | ❌ MISSING |

---

## Detailed Workflow Analysis

### 1. Data Tainting Workflow

**Spec Requirements** (spec-extraction.md L248-269):
- **3 Attachment Points**:
  1. At Message creation (Gateway → user, Orchestrator → tool_result, MemoryService → injected block, Heartbeat)
  2. At ContextBlock assembly (Security calls `applyBlockTaints`)
  3. On any data leaving the runtime (Mail, stream chunks, snapshot, sub-agent result) — uses `allowedOnExit`
- **Taint Propagation**: Security Service propagates taints on copy/read/write

**Covering Phases**:
| Phase | Tests | Type |
|-------|-------|------|
| P1.1: AttachTaint functionality | 4 tests | Unit |
| P1.2: PropagateTaint recursive propagation | 5 tests | Unit |
| P1.3: StripTaint removal | 4 tests | Unit |

**Missing E2E Tests**:
- `TestDataTainting_E2E_MessageCreation` - Full flow: Gateway creates message → taints attached → propagate through system
- `TestDataTainting_E2E_ContextBlockAssembly` - ContextBlock assembled → taints applied → verified in output
- `TestDataTainting_E2E_ExitSanitization` - Data leaves runtime → allowedOnExit enforced → forbidden taints stripped

**Steps NOT Tested End-to-End**:
- ❌ Message creation → taint attachment → propagation → exit sanitization (full chain)
- ❌ All 3 attachment points exercised together
- ❌ Taint propagation across boundary transitions

---

### 2. Boundary Enforcement Workflow

**Spec Requirements** (spec-extraction.md L271-281):
- Transitions across boundaries are **explicit and always go through Mail system** (never direct event passing)
- Enforced by `sys:security` System Service Chart at runtime
- `prepareContextForBoundary(runtimeId, boundary)` called inside every `assembleContextMap`
- Forbidden taints redacted/dropped before LLM call

**Covering Phases**:
| Phase | Tests | Type |
|-------|-------|------|
| P3.1: Boundary validation | 5 tests | Unit |
| P3.2: Boundary transitions | 6 tests | Unit |
| P3.3: Violation reporting | 4 tests | Unit |
| P6.2: Mail router hooks | 4 tests | Integration |

**Missing E2E Tests**:
- `TestBoundaryEnforcement_E2E_InnerToOuter` - Inner agent → Mail → Outer agent → verify taints stripped
- `TestBoundaryEnforcement_E2E_OuterToInner_Blocked` - Outer agent attempts forbidden transition → blocked with violation
- `TestBoundaryEnforcement_E2E_ContextMapAssembly` - ContextMap assembled → forbidden blocks filtered → LLM receives clean prompt

**Steps NOT Tested End-to-End**:
- ❌ Full Mail routing through security service with boundary enforcement
- ❌ `prepareContextForBoundary` integration with ContextMap assembly
- ❌ Violation reporting to dead-letter queue via sys:observability

---

### 3. Stream Sanitization Workflow

**Spec Requirements** (spec-extraction.md L283-287):
- Every outgoing chunk passes through DMZ sanitizers (PII redaction, length caps, schema validation, inner-data stripping)
- **Sanitization is per-chunk (stateless) so latency stays <50ms**
- No buffering of entire response unless chart explicitly requests it

**Covering Phases**:
| Phase | Tests | Type |
|-------|-------|------|
| P5.1: Stream taint stripping | 4 tests | Unit |
| P5.2: Stream boundary enforcement | 4 tests | Unit + 1 benchmark |

**Missing E2E Tests**:
- `TestStreamSanitization_E2E_FullStream` - Multi-chunk stream → each chunk sanitized → latency measured end-to-end
- `TestStreamSanitization_E2E_PII_Redaction` - Stream with PII → redacted output verified
- `TestStreamSanitization_E2E_LengthCap` - Long stream → truncated with isFinal flag

**Steps NOT Tested End-to-End**:
- ❌ Full stream pipeline with multiple sanitization stages
- ❌ Latency measurement in realistic conditions (benchmark is synthetic)
- ❌ Schema validation integrated with taint stripping

---

### 4. DataSource Write Workflow

**Spec Requirements** (spec-extraction.md L289-296):
1. Write action calls `sys:datasources.tagOnWrite(path, incomingTaints)`
2. For localDisk: sets xattr `user.maelstrom.taints=PII,TOOL_OUTPUT` (or JSON-encoded)
3. For object stores: sets object tags/metadata
4. For workspaces: updates sidecar manifest for fast lookup

**Covering Phases**:
| Phase | Tests | Type |
|-------|-------|------|
| P2.1: LocalDisk xattr implementation | 5 tests | Unit |
| P2.2: S3 DataSource | 4 tests | Unit |
| P2.3: InMemory DataSource | 3 tests | Unit |
| P2.4: DataSource Registry | 3 tests | Unit |

**Missing E2E Tests**:
- `TestDataSourceWrite_E2E_LocalDisk` - Write file → xattr set → read xattr → verify taints
- `TestDataSourceWrite_E2E_S3` - Write S3 object → tags set → read tags → verify taints
- `TestDataSourceWrite_E2E_SidecarFallback` - Non-xattr filesystem → sidecar created → taints persisted

**Steps NOT Tested End-to-End**:
- ❌ Full write → tag → read → taint retrieval cycle
- ❌ Sidecar fallback mechanism on real filesystems
- ❌ DataSource registry integration with actual writes

---

### 5. DataSource Read Workflow

**Spec Requirements** (spec-extraction.md L298-303):
1. Read tool calls `sys:datasources.getTaints(path)`
2. Security service immediately attaches those taints to resulting `tool_result` or injected ContextBlock Message
3. Existing `prepareContextForBoundary` + per-block `taintPolicy` redacts/drops/audits as configured

**Covering Phases**:
| Phase | Tests | Type |
|-------|-------|------|
| P2.1: LocalDisk xattr implementation | 5 tests (includes GetTaints) | Unit |
| P2.2: S3 DataSource | 4 tests (includes GetTaints) | Unit |
| P2.3: InMemory DataSource | 3 tests (includes GetTaints) | Unit |

**Missing E2E Tests**:
- `TestDataSourceRead_E2E_TaintAttachment` - Read file → taints retrieved → attached to tool_result → verified
- `TestDataSourceRead_E2E_ContextBlockInjection` - Read file → taints attached to ContextBlock → boundary filtering applied
- `TestDataSourceRead_E2E_RedactionFlow` - Read tainted file → redact policy applied → clean output

**Steps NOT Tested End-to-End**:
- ❌ Read → taint retrieval → attachment → boundary filtering complete chain
- ❌ Integration with ContextMap assembly and taintPolicy enforcement
- ❌ Redaction flow from DataSource through to LLM prompt

---

### 6. Isolation Hook Behavior

**Spec Requirements** (spec-extraction.md L305-309):
- Every agent's persistent workspace is mounted via isolated executor (container/namespace) with exact DataSource policy
- No direct `open()` syscall allowed — all FS ops go through mediated tool layer

**Covering Phases**:
| Phase | Tests | Type |
|-------|-------|------|
| P4.3: Isolation hooks | 3 tests | Unit |

**Missing E2E Tests**:
- `TestIsolationHook_E2E_NamespaceIsolation` - Agent in namespace → FS ops mediated → direct syscall blocked
- `TestIsolationHook_E2E_ToolLayerMediation` - File operation → routed through tool layer → policy enforced
- `TestIsolationHook_E2E_WorkspaceMount` - Workspace mounted → isolated view → DataSource policy applied

**Steps NOT Tested End-to-End**:
- ❌ Full isolation enforcement with container/namespace
- ❌ Mediated tool layer blocking direct syscalls
- ❌ Workspace mount with DataSource policy integration

---

## Missing Workflows Summary

### Critical Gaps

| Gap | Impact | Severity |
|-----|--------|----------|
| **No E2E tests for any workflow** | Cannot verify end-to-end security guarantees | CRITICAL |
| **No cross-phase integration tests** | Individual phases may work but integration may fail | HIGH |
| **No realistic latency testing** | <50ms requirement only tested with benchmarks | HIGH |
| **No violation-to-dead-letter flow** | Security violations may not be properly reported | HIGH |

### Required E2E Tests (Minimum Set)

```
E2E Test Suite for Layer 4:

1. TestE2E_DataTainting_FullChain
   - Message created with taints
   - Taints propagated through system
   - Taints enforced on exit

2. TestE2E_BoundaryEnforcement_MailRouting
   - Mail routed across boundaries
   - Security service validates
   - Forbidden transitions blocked

3. TestE2E_StreamSanitization_Latency
   - Multi-chunk stream processed
   - Each chunk sanitized
   - Total latency <50ms

4. TestE2E_DataSource_WriteReadCycle
   - Write with taints
   - Read and retrieve taints
   - Taints attached to result

5. TestE2E_IsolationHook_FS_Mediation
   - Agent attempts FS operation
   - Routed through tool layer
   - Direct syscall blocked

6. TestE2E_ViolationReporting_Flow
   - Violation triggered
   - Sent to dead-letter
   - sys:observability receives
```

---

## Recommendations

### Immediate Actions

1. **Add E2E test phase (P7)**: Create new phase with 6+ E2E tests covering full workflows
2. **Add integration tests**: Between phases to verify cross-phase behavior
3. **Add realistic latency tests**: Measure actual latency in production-like conditions

### Phase P7: E2E Workflows (Recommended)

| Sub-phase | Tests | Description |
|-----------|-------|-------------|
| P7.1 | 2 | Data Tainting E2E |
| P7.2 | 2 | Boundary Enforcement E2E |
| P7.3 | 2 | Stream Sanitization E2E |
| P7.4 | 2 | DataSource E2E |
| P7.5 | 2 | Isolation Hook E2E |
| P7.6 | 2 | Violation Reporting E2E |
| **Total** | **12** | **New E2E tests** |

---

## Conclusion

**CRITICAL**: The current phase-breakdown.md defines 88 unit/integration tests but **ZERO E2E tests** for the 6 workflows specified in the architecture. This is a significant gap that must be addressed before Layer 4 can be considered complete.

**Action Required**: Add Phase P7 with E2E workflow tests to verify end-to-end security guarantees.

---

*Audit completed: 2026-03-09*  
*Auditor: File Search Specialist*  
*Based on: arch-v1.md v1.3, spec-extraction.md, phase-breakdown.md*
