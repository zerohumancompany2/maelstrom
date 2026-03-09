# Layer 4: Recursive Audit Summary

**Generated**: 2026-03-09  
**Scope**: All Layer 4 (Security & Boundaries + DataSources) requirements  
**Source Documents**: arch-v1.md v1.3, spec-extraction.md, phase-breakdown.md, all 19 phase plans

---

## 1. Executive Summary

### Overall Coverage

| Metric | Value |
|--------|-------|
| Total Requirements in Spec | 344+ references |
| Requirements Covered | ~50% |
| Total Tests Planned | 88 |
| Additional Tests Needed | ~50-70 |
| **Total Tests Required** | **~140-160** |

### Critical Gaps Count

| Category | Critical Gaps |
|----------|---------------|
| Security Interface | 3 |
| DataSource Interface | 1 |
| Workflows (E2E) | 6 |
| Constraints | 1 |
| Missing Requirements | 10 |
| **TOTAL CRITICAL GAPS** | **21** |

### Go/No-Go Recommendation

**⚠️ NO-GO FOR CURRENT PLAN** - Proceed only after revising plans to include:
1. Phase P7: E2E Workflows (12 tests)
2. Phases P1.5-P1.9: Additional Taint Engine (15 tests)
3. Phases P3.4-P3.6: Additional Boundary Enforcement (11 tests)
4. Phases P4.4-P4.12: Additional ContextMap (20 tests)
5. Phases P5.3-P5.6: Additional Stream Integration (8 tests)
6. Phases P6.4-P6.6: Additional Service Wiring (6 tests)
7. Phases P2.5-P2.6: Additional DataSource (5 tests)

**Recommendation**: Revise all plans first, then proceed with implementation.

---

## 2. Gap Summary by Category

### 2.1 Interface Gaps (Security Interface)

**Source**: recursive-audit-security-interface.md

| Security Interface Method | Status | Gap Description |
|---------------------------|--------|-----------------|
| `ValidateAndSanitize` | ⚠️ Partial | P1.3 covers StripTaint, P3.1 covers ValidateBoundary, but no integrated test for full interface method |
| `TaintPropagate` | ✅ Covered | P1.2 PropagateTaint directly implements this |
| `NamespaceIsolate` | ❌ MISSING | No phase covers NamespaceIsolation or runtime context isolation |
| `CheckTaintPolicy` | ⚠️ Partial | P4.1 covers EnforcePolicy but doesn't match exact interface signature |
| `ReportTaints` | ❌ MISSING | No phase covers taint reporting or TaintMap generation |
| `PrepareContextForBoundary` | ❌ MISSING | No phase covers context preparation for boundary transitions |

**Missing Edge Cases**:
- Concurrent taint attachment (race conditions)
- Circular reference structures in PropagateTaint
- Mail struct stripping (only map tested)
- Boundary transition validation (source → target)
- Policy inheritance or composition

**Recommended New Phases**:
- P5.1: NamespaceIsolation (3 tests)
- P5.2: TaintReporting (3 tests)
- P5.3: ContextPreparation (3 tests)

### 2.2 DataSource Gaps

**Source**: recursive-audit-datasource-interface.md

| Implementation | Method | Status |
|---------------|--------|--------|
| InMemory | ValidateAccess | ❌ GAP - No test planned for boundary validation |

**Coverage Statistics**:
- Total Interface Methods: 3
- Total Implementations: 4
- Combinations with Tests Planned: 9/10 (90%)
- Gaps: 1 (InMemory ValidateAccess)

**Recommended Fix**: Add test to P2.3-InMemory.md for ValidateAccess boundary enforcement.

### 2.3 Workflow Gaps

**Source**: recursive-audit-workflows.md

**CRITICAL FINDING**: All 6 workflows from the spec have **ZERO E2E tests**.

| Workflow | Spec Reference | E2E Tests | Status |
|----------|---------------|-----------|--------|
| Data Tainting Workflow | L248-269 | NONE | ❌ MISSING |
| Boundary Enforcement Workflow | L271-281 | NONE | ❌ MISSING |
| Stream Sanitization Workflow | L283-287 | NONE | ❌ MISSING |
| DataSource Write Workflow | L289-296 | NONE | ❌ MISSING |
| DataSource Read Workflow | L298-303 | NONE | ❌ MISSING |
| Isolation Hook Behavior | L305-309 | NONE | ❌ MISSING |

**Required E2E Tests (Minimum Set)**:
1. `TestE2E_DataTainting_FullChain` - Message created → taints propagated → enforced on exit
2. `TestE2E_BoundaryEnforcement_MailRouting` - Mail routed across boundaries → security validates → forbidden blocked
3. `TestE2E_StreamSanitization_Latency` - Multi-chunk stream → each chunk sanitized → latency <50ms
4. `TestE2E_DataSource_WriteReadCycle` - Write with taints → read → taints attached to result
5. `TestE2E_IsolationHook_FS_Mediation` - FS operation → routed through tool layer → direct syscall blocked
6. `TestE2E_ViolationReporting_Flow` - Violation triggered → sent to dead-letter → sys:observability receives

**Recommended Phase P7: E2E Workflows** (12 tests):
- P7.1: Data Tainting E2E (2 tests)
- P7.2: Boundary Enforcement E2E (2 tests)
- P7.3: Stream Sanitization E2E (2 tests)
- P7.4: DataSource E2E (2 tests)
- P7.5: Isolation Hook E2E (2 tests)
- P7.6: Violation Reporting E2E (2 tests)

### 2.4 Constraint Gaps

**Source**: recursive-audit-constraints.md

| Constraint | Coverage Status | Gaps |
|------------|-----------------|------|
| Security Invariant | ✅ 100% | None |
| Performance <50ms | ⚠️ Partial | No explicit benchmark test |
| Boundary Immutability | ❌ MISSING | No tests in any phase |
| TaintPolicy Modes | ✅ 100% | None |
| allowedOnExit Policy | ✅ 100% | None |
| Per-chunk Stateless | ✅ 100% | None |

**Critical Gap: Boundary Immutability**
- **Spec References**: arch-v1.md L127, L1783
- **Impact**: HIGH - Without tests, boundary could be accidentally changed during hot-reload or version migration
- **Required Tests**:
  1. `TestBoundaryImmutability_HotReload` - Verify boundary unchanged after hot-reload
  2. `TestBoundaryImmutability_VersionUpgrade` - Verify boundary unchanged across version changes
  3. `TestBoundaryImmutability_MigrationPolicy` - Verify migrationPolicy does not affect boundary

**Medium Gap: Performance Benchmark**
- No explicit benchmark test with `go test -bench` annotation
- No load testing under concurrent conditions
- No worst-case scenario testing

### 2.5 Missing Requirements Gaps

**Source**: recursive-audit-missing-requirements.md

**44 Total Gaps Identified** across 10 categories:

| Category | Total References | Covered | Gaps |
|----------|-----------------|---------|------|
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

---

## 3. Critical Issues (Must Fix Before Implementation)

| Priority | Issue | Impact | Recommended Phase |
|----------|-------|--------|-------------------|
| **P0** | No E2E tests for any of 6 workflows | Cannot verify end-to-end security guarantees | P7 |
| **P0** | Boundary immutability not tested | Security invariant could be violated during hot-reload | P3.4 |
| **P0** | Sub-agent boundary inheritance not tested | Inner data could leak to outer sub-agents | P3.5 |
| **P0** | Tool taintOutput automatic attachment not tested | Tool results may not be properly tainted | P1.6 |
| **P0** | Persistence refuses writes violating taint policy not tested | Data could be persisted with wrong taints | P2.5 |
| **P1** | Security interface methods incomplete (NamespaceIsolate, ReportTaints, PrepareContextForBoundary) | Interface cannot be fully implemented | P5.1-P5.3 |
| **P1** | Boundary-aware tool resolution not tested | Inner tools could be exposed to outer callers | P3.6 |
| **P1** | ContextMap token budgeting not tested | Prompts could exceed token limits | P4.4 |
| **P1** | Gateway blocks inner-boundary Chart exposure not tested | Inner charts could be directly exposed | P5.3 |
| **P1** | Runtime guard can query taints not tested | Violations may not be detected | P6.4 |

---

## 4. Recommended New Phases

### 4.1 Phase P1: Additional Taint Engine

| Phase | Tests | Description |
|-------|-------|-------------|
| P1.5 | 3 | Taint Invariant Property Tests (compile-time guarantees) |
| P1.6 | 3 | Tool TaintOutput automatic attachment |
| P1.7 | 2 | MemoryService/Heartbeat taint attachment |
| P1.8 | 2 | Bloom filters for fast-path taint checking |
| P1.9 | 2 | TaintMap standalone tests |

**Total P1.X Additional**: 12 tests

### 4.2 Phase P2: Additional DataSource

| Phase | Tests | Description |
|-------|-------|-------------|
| P2.5 | 3 | Persistence taint policy enforcement |
| P2.6 | 2 | alwaysTaintAs taintMode option |

**Total P2.X Additional**: 5 tests

### 4.3 Phase P3: Additional Boundary Enforcement

| Phase | Tests | Description |
|-------|-------|-------------|
| P3.4 | 3 | Boundary immutability (hot-reload, version upgrade) |
| P3.5 | 4 | Sub-agent boundary inheritance rules |
| P3.6 | 4 | Boundary-aware tool resolution |

**Total P3.X Additional**: 11 tests

### 4.4 Phase P4: Additional ContextMap

| Phase | Tests | Description |
|-------|-------|-------------|
| P4.4 | 4 | ContextMap token budgeting and eviction |
| P4.5 | 5 | ContextBlock sources (static, session, memoryService, toolRegistry, runtime) |
| P4.6 | 2 | applyBlockTaints function |
| P4.7 | 2 | Orchestrator isolated execution contexts |
| P4.8 | 2 | Isolation hooks API for DevOps |
| P4.9 | 1 | ContextBlock streamCommit behavior |
| P4.10 | 1 | ContextBlock qualityScore usage |
| P4.11 | 1 | ContextBlock compressor integration |
| P4.12 | 1 | ContextBlock priority ordering |

**Total P4.X Additional**: 19 tests

### 4.5 Phase P5: Additional Stream Integration

| Phase | Tests | Description |
|-------|-------|-------------|
| P5.3 | 2 | Gateway inner-boundary blocking |
| P5.4 | 2 | Stream length caps |
| P5.5 | 2 | Stream schema validation |
| P5.6 | 2 | sys:human-gateway sanitization |

**Total P5.X Additional**: 8 tests

### 4.6 Phase P6: Additional Service Wiring

| Phase | Tests | Description |
|-------|-------|-------------|
| P6.4 | 2 | Runtime taint query |
| P6.5 | 2 | allowedOnExit for snapshots |
| P6.6 | 2 | allowedOnExit for sub-agent returns |

**Total P6.X Additional**: 6 tests

### 4.7 Phase P7: E2E Workflows (NEW)

| Phase | Tests | Description |
|-------|-------|-------------|
| P7.1 | 2 | Data Tainting E2E |
| P7.2 | 2 | Boundary Enforcement E2E |
| P7.3 | 2 | Stream Sanitization E2E |
| P7.4 | 2 | DataSource E2E |
| P7.5 | 2 | Isolation Hook E2E |
| P7.6 | 2 | Violation Reporting E2E |

**Total P7.X Additional**: 12 tests

---

## 5. Revised Test Count

| Category | Current | Additional | New Total |
|----------|---------|------------|-----------|
| **P1: Taint Engine Core** | 16 | 12 | 28 |
| **P2: DataSource** | 15 | 5 | 20 |
| **P3: Boundary Enforcement** | 15 | 11 | 26 |
| **P4: ContextMap Filtering** | 14 | 19 | 33 |
| **P5: Stream Integration** | 8 | 8 | 16 |
| **P6: Service Wiring** | 13 | 6 | 19 |
| **P7: E2E Workflows** | 0 | 12 | 12 |
| **TOTAL** | **88** | **73** | **161** |

---

## 6. Revised Execution Plan

### Updated Phase Order with Dependencies

```
Phase P1 (Taint Engine Core) - Foundational
├── P1.1: AttachTaint (4 tests) - NO DEPS
├── P1.4: TaintSet (3 tests) - NO DEPS
├── P1.2: PropagateTaint (5 tests) - DEPS: P1.1
├── P1.3: StripTaint (4 tests) - DEPS: P1.1
├── P1.5: Taint Invariant Property Tests (3 tests) - DEPS: P1.1, P1.2
├── P1.6: Tool TaintOutput (3 tests) - DEPS: P1.1, P1.2
├── P1.7: MemoryService/Heartbeat Taint (2 tests) - DEPS: P1.1
├── P1.8: Bloom Filters (2 tests) - DEPS: P1.4
└── P1.9: TaintMap Standalone (2 tests) - DEPS: P1.1

Phase P2 (DataSource) - Foundational
├── P2.1: LocalDisk (5 tests) - DEPS: P1
├── P2.2: S3 (4 tests) - DEPS: P1, P2.1
├── P2.3: InMemory (4 tests) - DEPS: P1 (add ValidateAccess test)
├── P2.4: Registry (3 tests) - DEPS: P2.1, P2.2, P2.3
├── P2.5: Persistence Taint Policy (3 tests) - DEPS: P2.1, P2.4
└── P2.6: alwaysTaintAs Mode (2 tests) - DEPS: P2.1, P2.2

Phase P3 (Boundary Enforcement) - Foundational
├── P3.1: Boundary Validation (5 tests) - DEPS: P1, P2
├── P3.2: Boundary Transitions (6 tests) - DEPS: P3.1, P1
├── P3.3: Violation Reporting (4 tests) - DEPS: P3.1, P3.2
├── P3.4: Boundary Immutability (3 tests) - DEPS: P3.1
├── P3.5: Sub-Agent Boundary Inheritance (4 tests) - DEPS: P3.1, P3.2
└── P3.6: Boundary-Aware Tool Resolution (4 tests) - DEPS: P3.1, P3.5

Phase P4 (ContextMap Filtering) - Foundational
├── P4.1: TaintPolicy Modes (5 tests) - DEPS: P1, P3
├── P4.2: ContextBlock Filtering (6 tests) - DEPS: P4.1
├── P4.3: Isolation Hooks (3 tests) - DEPS: P4.1, P4.2
├── P4.4: ContextMap Token Budgeting (4 tests) - DEPS: P4.2
├── P4.5: ContextBlock Sources (5 tests) - DEPS: P4.2
├── P4.6: applyBlockTaints (2 tests) - DEPS: P4.2, P1.2
├── P4.7: Orchestrator Isolated Contexts (2 tests) - DEPS: P4.3
├── P4.8: Isolation Hooks API (2 tests) - DEPS: P4.3
├── P4.9: ContextBlock streamCommit (1 test) - DEPS: P4.2
├── P4.10: ContextBlock qualityScore (1 test) - DEPS: P4.2
├── P4.11: ContextBlock compressor (1 test) - DEPS: P4.2
└── P4.12: ContextBlock priority (1 test) - DEPS: P4.2

Phase P5 (Stream Integration) - Integration
├── P5.1: Stream Taint Stripping (4 tests) - DEPS: P1, P3
├── P5.2: Stream Boundary Enforcement (4 tests) - DEPS: P5.1, P3
├── P5.3: Gateway Inner-Boundary Blocking (2 tests) - DEPS: P3.1, P3.6
├── P5.4: Stream Length Caps (2 tests) - DEPS: P5.1
├── P5.5: Stream Schema Validation (2 tests) - DEPS: P5.1
└── P5.6: sys:human-gateway Sanitization (2 tests) - DEPS: P5.1, P4.2

Phase P6 (Service Wiring) - Integration
├── P6.1: SecurityService Integration (5 tests) - DEPS: P1-P5
├── P6.2: Mail Router Hooks (4 tests) - DEPS: P6.1, P3
├── P6.3: ApplicationContext Wiring (4 tests) - DEPS: P6.1, P1
├── P6.4: Runtime Taint Query (2 tests) - DEPS: P6.1
├── P6.5: allowedOnExit Snapshots (2 tests) - DEPS: P6.1
└── P6.6: allowedOnExit Sub-Agent (2 tests) - DEPS: P6.1

Phase P7 (E2E Workflows) - Final Validation
├── P7.1: Data Tainting E2E (2 tests) - DEPS: P1, P6
├── P7.2: Boundary Enforcement E2E (2 tests) - DEPS: P3, P6
├── P7.3: Stream Sanitization E2E (2 tests) - DEPS: P5, P6
├── P7.4: DataSource E2E (2 tests) - DEPS: P2, P6
├── P7.5: Isolation Hook E2E (2 tests) - DEPS: P4, P6
└── P7.6: Violation Reporting E2E (2 tests) - DEPS: P3, P6
```

### Execution Timeline

| Phase | Tests | Estimated Duration |
|-------|-------|-------------------|
| P1 (Taint Engine) | 28 | 2-3 days |
| P2 (DataSource) | 20 | 2-3 days |
| P3 (Boundary) | 26 | 2-3 days |
| P4 (ContextMap) | 33 | 3-4 days |
| P5 (Stream) | 16 | 1-2 days |
| P6 (Service) | 19 | 2-3 days |
| P7 (E2E) | 12 | 2-3 days |
| **TOTAL** | **161** | **14-21 days** |

---

## 7. Final Recommendation

### Option A: Proceed with Current Plans, Then Add P7+

**Pros**:
- Faster initial progress
- Can demonstrate unit/integration test coverage quickly
- Lower initial complexity

**Cons**:
- Will need to refactor/revisit phases later
- E2E gaps will remain unaddressed
- Critical security invariants untested
- May discover integration issues that require phase changes

### Option B: Revise All Plans First, Then Implement

**Pros**:
- Complete test coverage from the start
- All security invariants verified
- No need to revisit phases
- E2E tests guide implementation decisions
- Follows CLAUDE.md "Complete spec before implementation" rule

**Cons**:
- Slower initial progress
- More planning work upfront
- Longer time to first commit

### Recommendation: **Option B - Revise All Plans First**

**Rationale**:
1. CLAUDE.md mandates "Complete spec (user stories + acceptance criteria)" before any implementation
2. 21 critical gaps identified that must be fixed before implementation
3. E2E tests are essential for verifying security guarantees
4. Current 88 tests only cover ~50% of Layer 4 requirements
5. Revising plans now prevents costly refactoring later

**Action Items**:
1. Create all new phase plans (P1.5-P1.9, P2.5-P2.6, P3.4-P3.6, P4.4-P4.12, P5.3-P5.6, P6.4-P6.6, P7.1-P7.6)
2. Fix P4.1 line reference (L209 → L914-919)
3. Add ValidateAccess test to P2.3-InMemory.md
4. Update phase-breakdown.md with all new phases
5. Run recursive audit again to verify completeness
6. Proceed with implementation only after all 161 tests are specified

---

*Document generated: 2026-03-09*  
*Based on: arch-v1.md v1.3, recursive-audit-security-interface.md, recursive-audit-datasource-interface.md, recursive-audit-workflows.md, recursive-audit-constraints.md, recursive-audit-missing-requirements.md, audit-report.md*