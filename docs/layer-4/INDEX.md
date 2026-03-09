# Layer 4: Security & Boundaries + DataSources

**Status**: ✅ READY FOR TDD IMPLEMENTATION | **Tests**: 147 | **Phases**: 45 | **Coverage**: ~95%+

## ✅ Planning Complete

**Final coverage audit complete — all plans ready for implementation:**

- **Total Phases**: 45 phases with detailed test specifications
- **Total Tests**: 147 tests with Given/When/Then acceptance criteria
- **Spec Coverage**: ~95%+ of arch-v1.md requirements covered
- **Audit Coverage**: 18/45 phases (40%) audited, covering critical security paths

**See**: [`FINAL-COVERAGE-REPORT.md`](FINAL-COVERAGE-REPORT.md) for complete coverage analysis

---

---

## Overview

Layer 4 implements **Security & Boundaries + DataSources** — the core security enforcement layer that ensures data crossing boundaries (inner/DMZ/outer) is properly sanitized and that file/object storage operations carry appropriate taints.

### Key Guarantee

> "Never allow outer/DMZ ContextMap to contain untagged inner data"  
> — `arch-v1.md` L20

### Primary Components

| Component | Type | Responsibility |
|-----------|------|----------------|
| **Security Service** (`sys:security`) | Core Platform Service | Boundary enforcement, taint validation, ContextMap sanitization |
| **Taint Engine** | Sub-component | Attach/propagate taints, policy enforcement, bloom filter optimization |
| **DataSource Abstraction** | Pluggable Infrastructure | Filesystem/object storage with automatic tainting |
| **Boundary Enforcement** | Sub-component | Enforce inner/DMZ/outer transitions via Mail only |
| **ContextMap Filtering** | Sub-component | Redact/drop/audit per-block taintPolicy before LLM calls |

### Security Invariant

> "All data entering the runtime is tainted at the border. No untainted information exists inside the runtime."  
> — `arch-v1.md` L276

---

## Document Navigation

### Core Planning Documents

| Document | Purpose | Lines |
|----------|---------|-------|
| [spec-extraction.md](spec-extraction.md) | Complete spec extraction from `arch-v1.md` | 592 |
| [phase-breakdown.md](phase-breakdown.md) | Detailed 19-phase breakdown | 1501 |
| [implementation-assessment.md](implementation-assessment.md) | Current implementation state | 567 |
| [implementation-patterns.md](implementation-patterns.md) | Proven patterns from Layers 1-3 | 883 |
| [audit-report.md](audit-report.md) | Initial audit of 19 phase plans | 571 |

### Final Coverage Documents

| Document | Purpose | Status |
|----------|---------|--------|
| **[FINAL-COVERAGE-REPORT.md](FINAL-COVERAGE-REPORT.md)** | **Complete coverage analysis** | **✅ 45 phases, 147 tests, ~95%+ coverage** |
| [phase-breakdown.md](phase-breakdown.md) | Detailed phase breakdown | ✅ All 45 phases documented |
| [spec-extraction.md](spec-extraction.md) | Complete spec extraction from `arch-v1.md` | ✅ 38 unique arch-v1.md references |

### Phase Plans (45 total)

**[plans/](plans/)** — Individual phase plan documents

### Phase Plans (19 total)

**Phase P1: Taint Engine Core** (Foundation)
- [P1.1-AttachTaint.md](plans/P1.1-AttachTaint.md) — Attach taints at Message creation, file read (4 tests)
- [P1.2-PropagateTaint.md](plans/P1.2-PropagateTaint.md) — Recursive taint propagation (5 tests)
- [P1.3-StripTaint.md](plans/P1.3-StripTaint.md) — Strip forbidden taints on transitions (4 tests)
- [P1.4-TaintSet.md](plans/P1.4-TaintSet.md) — TaintSet operations (3 tests)

**Phase P2: DataSource** (Depends on P1)
- [P2.1-LocalDisk.md](plans/P2.1-LocalDisk.md) — xattr-based taint storage (5 tests)
- [P2.2-S3.md](plans/P2.2-S3.md) — S3 object storage (4 tests)
- [P2.3-InMemory.md](plans/P2.3-InMemory.md) — inMemoryWorkspace (3 tests)
- [P2.4-Registry.md](plans/P2.4-Registry.md) — DataSource registry (3 tests)

**Phase P3: Boundary Enforcement** (Depends on P1, P2)
- [P3.1-BoundaryValidation.md](plans/P3.1-BoundaryValidation.md) — Boundary validation (5 tests)
- [P3.2-BoundaryTransitions.md](plans/P3.2-BoundaryTransitions.md) — Transition enforcement (6 tests)
- [P3.3-ViolationReporting.md](plans/P3.3-ViolationReporting.md) — Violation reporting (4 tests)

**Phase P4: ContextMap Filtering** (Depends on P1, P3)
- [P4.1-TaintPolicyModes.md](plans/P4.1-TaintPolicyModes.md) — Policy enforcement modes (5 tests)
- [P4.2-ContextBlockFiltering.md](plans/P4.2-ContextBlockFiltering.md) — Block filtering (6 tests)
- [P4.3-IsolationHooks.md](plans/P4.3-IsolationHooks.md) — Namespace isolation (3 tests)

**Phase P5: Stream Integration** (Depends on P1, P3)
- [P5.1-StreamTaintStripping.md](plans/P5.1-StreamTaintStripping.md) — Stream sanitization (4 tests)
- [P5.2-StreamBoundaryEnforcement.md](plans/P5.2-StreamBoundaryEnforcement.md) — Stream boundary enforcement (4 tests)

**Phase P6: Service Wiring** (Depends on All)
- [P6.1-SecurityServiceIntegration.md](plans/P6.1-SecurityServiceIntegration.md) — SecurityService methods (5 tests)
- [P6.2-MailRouterHooks.md](plans/P6.2-MailRouterHooks.md) — Mail router hooks (4 tests)
- [P6.3-ApplicationContextWiring.md](plans/P6.3-ApplicationContextWiring.md) — ApplicationContext wiring (4 tests)

---

## Phase Summary

| Group | Phases | Tests | Status |
|-------|--------|-------|--------|
| P1: Taint Engine Core | 5 | 19 | ✅ Planned |
| P2: DataSource Implementations | 6 | 22 | ✅ Planned |
| P3: Boundary Enforcement | 6 | 26 | ✅ Planned |
| P4: ContextMap & Isolation | 8 | 33 | ✅ Planned |
| P5: Stream Sanitization | 8 | 24 | ✅ Planned |
| P6: Integration & Wiring | 6 | 23 | ✅ Planned |
| P7: End-to-End | 6 | 12 | ✅ Planned |
| **Total** | **45** | **147** | **✅ Ready** |

**Legend**: ✅ Planned = Phase plan complete with tests specified

---

## Execution Order

```
P1 (Taint Engine Core - Foundation, 19 tests)
├── P1.1: AttachTaint (4 tests)
├── P1.2: PropagateTaint (5 tests) <── P1.1
├── P1.3: StripTaint (4 tests) <── P1.1
├── P1.4: TaintSet (3 tests)
└── P1.6: ToolTaintOutput (3 tests)
        ↓
P2 (DataSource, 22 tests)
├── P2.1: LocalDisk (5 tests) <── P1
├── P2.2: S3 (4 tests) <── P1, P2.1
├── P2.3: InMemory (3 tests) <── P1
├── P2.4: Registry (3 tests) <── P2.1-P2.3
├── P2.5: PersistenceTaintPolicy (3 tests)
└── P2.6: AlwaysTaintAsMode (2 tests)
        ↓
P3 (Boundary Enforcement, 26 tests)
├── P3.1: Validation (5 tests) <── P1, P2
├── P3.2: Transitions (6 tests) <── P3.1
├── P3.3: Violations (4 tests) <── P3.1, P3.2
├── P3.4: BoundaryImmutability (3 tests)
├── P3.5: SubAgentBoundaryInheritance (4 tests)
└── P3.6: BoundaryAwareToolResolution (4 tests)
        ↓
P4 (ContextMap & Isolation, 33 tests)
├── P4.1: TaintPolicyModes (5 tests) <── P1, P3
├── P4.2: ContextBlockFiltering (6 tests) <── P4.1
├── P4.3: IsolationHooks (3 tests) <── P4.1, P4.2
├── P4.4: ContextMapTokenBudgeting (4 tests)
├── P4.5: ContextBlockSources (5 tests)
├── P4.6: applyBlockTaints (2 tests)
├── P4.7: OrchestratorIsolatedContexts (2 tests)
└── P4.8: IsolationHooksAPI (2 tests)
        ↓
P5 (Stream Sanitization, 24 tests)
├── P5.1: NamespaceIsolation (3 tests) <── P1, P3
├── P5.1: StreamTaintStripping (4 tests)
├── P5.2: StreamBoundaryEnforcement (4 tests)
├── P5.2: TaintReporting (3 tests)
├── P5.3: ContextPreparation (3 tests)
├── P5.4: StreamLengthCaps (2 tests)
├── P5.5: StreamSchemaValidation (2 tests)
└── P5.6: sysHumanGatewaySanitization (2 tests)
        ↓
P6 (Integration & Wiring, 23 tests)
├── P6.1: SecurityServiceIntegration (5 tests) <── All
├── P6.2: MailRouterHooks (4 tests) <── P6.1
├── P6.3: ApplicationContextWiring (4 tests) <── P6.1
├── P6.4: RuntimeTaintQuery (2 tests)
├── P6.5: allowedOnExitSnapshots (2 tests)
└── P6.6: allowedOnExitSubAgentReturns (2 tests)
        ↓
P7 (End-to-End, 12 tests)
├── P7.1: DataTaintingE2E (2 tests)
├── P7.2: BoundaryEnforcementE2E (2 tests)
├── P7.3: StreamSanitizationE2E (2 tests)
├── P7.4: DataSourceE2E (2 tests)
├── P7.5: IsolationHookE2E (2 tests)
└── P7.6: ViolationReportingE2E (2 tests)
```

---

## Status

### Planning: ✅ COMPLETE

**Final Planning (147 tests, 45 phases):**
- ✅ All 45 phase plans created
- ✅ 147 tests specified with Given/When/Then format
- ✅ 38 unique arch-v1.md references covered
- ✅ Dependencies validated (DAG verified)
- ✅ Test counts within 2-10 range per CLAUDE.md
- ✅ 18/45 phases (40%) audited, covering critical security paths

**Coverage Summary**:
- **Phase P1**: Taint Engine Core (5 phases, 19 tests)
- **Phase P2**: DataSource Implementations (6 phases, 22 tests)
- **Phase P3**: Boundary Enforcement (6 phases, 26 tests)
- **Phase P4**: ContextMap & Isolation (8 phases, 33 tests)
- **Phase P5**: Stream Sanitization (8 phases, 24 tests)
- **Phase P6**: Integration & Wiring (6 phases, 23 tests)
- **Phase P7**: End-to-End (6 phases, 12 tests)

### Implementation: ✅ READY TO BEGIN

**Current State** (from implementation-assessment.md):

| Category | Status |
|----------|--------|
| Core type definitions | ✅ Complete |
| DataSource interface and registry | ✅ Complete |
| Security service skeleton | ✅ Complete |
| TaintEngine interface (stubbed) | ✅ Complete |
| Boundary enforcement logic | ❌ Missing |
| Taint tracking/propagation | ❌ Missing |
| xattr-based taint storage | ❌ Missing |
| ContextMap filtering/redaction | ❌ Missing |
| `allowedOnExit` policy enforcement | ❌ Missing |

**Risk**: Current implementation provides **zero security guarantees** — all enforcement methods are stubs.

### Audit Results: ✅ READY

| Audit | Result | Finding |
|-------|--------|---------|
| Initial Audit (19 plans) | ✅ PASS | 1 minor issue |
| **Final Coverage Audit (45 plans)** | **✅ READY** | **~95%+ coverage, 147 tests** |

**Recommendation**: Begin TDD implementation per CLAUDE.md rules.

---

## ✅ READY FOR IMPLEMENTATION

**All planning complete — begin TDD implementation.**

### Next Steps

1. **Review Coverage Report**: Read [`FINAL-COVERAGE-REPORT.md`](FINAL-COVERAGE-REPORT.md)
2. **Start with Phase P1**: Taint Engine Core (foundational phase)
3. **Follow CLAUDE.md TDD Workflow**:
   - Stub every public symbol
   - Write ONE test → RED
   - Implement minimal code → GREEN
   - Commit (1:1 test-to-commit ratio)
4. **Track Progress**: Update phase status as each completes

### Implementation Order

**P0 (Foundation - Start Here):**
1. P1: Taint Engine Core (19 tests) — All other phases depend on this

**P1 (Infrastructure):**
2. P2: DataSource Implementations (22 tests) — Depends on P1

**P2 (Security Enforcement):**
3. P3: Boundary Enforcement (26 tests) — Depends on P1, P2
4. P4: ContextMap & Isolation (33 tests) — Depends on P1, P3

**P3 (Integration):**
5. P5: Stream Sanitization (24 tests) — Depends on P1, P3
6. P6: Integration & Wiring (23 tests) — Depends on all

**P4 (Verification):**
7. P7: End-to-End (12 tests) — Final verification

---

## Quick Start

### Prerequisites

1. Review [`FINAL-COVERAGE-REPORT.md`](FINAL-COVERAGE-REPORT.md)
2. Review [`phase-breakdown.md`](phase-breakdown.md) for all 45 phases
3. Ensure test environment ready (Go toolchain, test dependencies)

### First Phase: P1.1 - AttachTaint

```bash
# 1. Create branch
git checkout -b feat/layer4-taint-attach

# 2. Read the phase plan
cat docs/layer-4/plans/P1.1-AttachTaint.md

# 3. Follow TDD workflow (1 test at a time)
#    - Write TestTaintEngine_AttachTaint_Mail → RED
#    - Implement minimal code → GREEN
#    - git commit -m "feat(layer-4/taint): add AttachTaint for Mail"
#    - Repeat for remaining 3 tests

# 4. Verify all tests pass
go test ./pkg/security/...

# 5. Commit summary
git commit -m "feat(layer-4/taint): complete AttachTaint functionality"
```

### Branch Naming Convention

```
feat/layer4-taint-attach        # P1.1
feat/layer4-taint-propagate     # P1.2
feat/layer4-taint-strip         # P1.3
feat/layer4-taintset            # P1.4
feat/layer4-datasource-local    # P2.1
feat/layer4-datasource-s3       # P2.2
...
```

### Commit Message Format

```
feat(layer-4/[component]): [one-line description]
```

Examples:
- `feat(layer-4/taint): add AttachTaint for Mail`
- `feat(layer-4/taint): add AttachTaint for nested maps`
- `feat(layer-4/boundary): add InnerToDMZ transition enforcement`

### Key Files to Create

```
pkg/security/
├── taint.go              # P1.1-P1.3
├── taint_test.go         # P1.1-P1.3
├── boundary.go           # P3.1-P3.2
├── boundary_test.go      # P3.1-P3.2
├── violation.go          # P3.3
├── violation_test.go     # P3.3
├── policy.go             # P4.1
├── policy_test.go        # P4.1
├── contextmap.go         # P4.2
├── contextmap_test.go    # P4.2
├── isolation.go          # P4.3
└── isolation_test.go     # P4.3

pkg/datasource/
├── local_disk.go         # P2.1 (enhance)
├── local_disk_test.go    # P2.1 (enhance)
├── s3.go                 # P2.2
├── s3_test.go            # P2.2
├── in_memory.go          # P2.3
└── in_memory_test.go     # P2.3
```

### TDD Workflow Reminder (from CLAUDE.md)

1. Stub every public symbol — zero implementation, only signatures
2. Write **one** test that exercises a single behavior
3. Run test suite → confirm it **fails** (RED)
4. Write minimal code that makes **only that test** pass (GREEN)
5. `git commit -m "feat(layer-4/[component]): <one-sentence>"`
6. Repeat until every test is green

**RULES**:
- Never write production code before a failing test
- Never write more than one test at a time
- Never commit failing tests or untested code
- 1:1 test-to-commit ratio

---

## References

- **Architecture Spec**: `docs/arch-v1.md` (v1.3, 2456 lines)
- **Security SHALL Statements**: `arch-v1.md` L1312-1317
- **Security SHALL NOT Statements**: `arch-v1.md` L1320-1322
- **Boundary Model**: `arch-v1.md` L261-270
- **DataSource Interface**: `arch-v1.md` L787-793
- **Security Interface**: `arch-v1.md` L1303-1309

---

*Last Updated: 2026-03-09*  
*Status: ✅ Planning Complete — 45 phases, 147 tests, ~95%+ coverage — Ready for TDD Implementation*