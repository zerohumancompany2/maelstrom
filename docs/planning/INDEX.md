# Layer 2 Planning Index

This directory contains planning documents for Layer 2: Core Platform Services.

## Overview

Layer 2 implements the 4 hard-coded platform services spawned by the Bootstrap Chart:
1. **sys:security** - Boundary enforcement, taint validation, sanitization
2. **sys:communication** - Mail pub/sub backbone
3. **sys:observability** - Traces, metrics, dead-letter queue
4. **sys:lifecycle** - Basic spawn/stop for bootstrap

## Phase Breakdown (Updated)

| Phase | Name | Tests | Status |
|-------|------|-------|--------|
| 2.1a | Core Type Definitions | 4 | ⏳ Pending |
| 2.1b | StreamChunk Migration | 1 | ⏳ Pending |
| 2.2a | Update Existing Tests | 0 (update 3) | ⏳ Pending |
| 2.2b | Publish Returns Ack | 3 | ⏳ Pending |
| 2.2c | Unsubscribe Implementation | 2 | ⏳ Pending |
| 2.3a | Update Existing Tests | 0 (update 2) | ⏳ Pending |
| 2.3b | QueryTraces with Filters | 4 | ⏳ Pending |
| 2.3c | Dead-Letter Queue | 2 | ⏳ Pending |
| 2.3d | Metrics Collection | 2 | ⏳ Pending |
| 2.4a | RuntimeInfo and Engine | 2 | ⏳ Pending |
| 2.4b | List Returns RuntimeInfo | 3 | ⏳ Pending |
| 2.4c | Control Method | 3 | ⏳ Pending |
| 2.5a | ValidateAndSanitize and TaintPropagate | 2 | ⏳ Pending |
| 2.5b | ReportTaints and PrepareContextForBoundary | 2 | ⏳ Pending |
| 2.6a | Fix Service Interface | 0 (fix only) | ⏳ Pending |
| 2.6b | Registry Implementation | 7 | ⏳ Pending |

**Total: 39 tests across 16 sub-phases** (reduced from 45 tests in 6 phases)

## Sub-Phase Documents

### Phase 2.1: Type Definitions (2 sub-phases)

#### Phase 2.1a: Core Type Definitions
- **Doc**: [phase-2.1a-core-type-definitions.md](phase-2.1a-core-type-definitions.md)
- **Branch**: `feat/layer2-2.1a-core-types`
- **Tests**: 4 tests, 4 commits
- **Dependencies**: None
- **Scope**: Define TraceFilters, MetricsCollector, RuntimeInfo types

#### Phase 2.1b: StreamChunk Migration
- **Doc**: [phase-2.1b-streamchunk-migration.md](phase-2.1b-streamchunk-migration.md)
- **Branch**: `feat/layer2-2.1b-streamchunk`
- **Tests**: 1 test, 1 commit
- **Dependencies**: Phase 2.1a
- **Scope**: Move StreamChunk to mail package, update MailMetadata

### Phase 2.2: Communication Service (3 sub-phases)

#### Phase 2.2a: Update Existing Tests
- **Doc**: [phase-2.2a-update-existing-tests.md](phase-2.2a-update-existing-tests.md)
- **Branch**: `feat/layer2-2.2a-update-tests`
- **Tests**: Update 3 existing tests
- **Dependencies**: Phase 2.1b
- **Scope**: Update tests for Publish() signature change

#### Phase 2.2b: Publish Returns Ack
- **Doc**: [phase-2.2b-publish-ack.md](phase-2.2b-publish-ack.md)
- **Branch**: `feat/layer2-2.2b-publish-ack`
- **Tests**: 3 tests, 3 commits
- **Dependencies**: Phase 2.2a
- **Scope**: Implement Publish() returning Ack

#### Phase 2.2c: Unsubscribe Implementation
- **Doc**: [phase-2.2c-unsubscribe.md](phase-2.2c-unsubscribe.md)
- **Branch**: `feat/layer2-2.2c-unsubscribe`
- **Tests**: 2 tests, 2 commits
- **Dependencies**: Phase 2.2b
- **Scope**: Implement Unsubscribe() method

### Phase 2.3: Observability Service (4 sub-phases)

#### Phase 2.3a: Update Existing Tests
- **Doc**: [phase-2.3a-update-existing-tests.md](phase-2.3a-update-existing-tests.md)
- **Branch**: `feat/layer2-2.3a-update-tests`
- **Tests**: Update 2 existing tests
- **Dependencies**: Phase 2.1a
- **Scope**: Update tests for QueryTraces() signature change

#### Phase 2.3b: QueryTraces with Filters
- **Doc**: [phase-2.3b-query-filters.md](phase-2.3b-query-filters.md)
- **Branch**: `feat/layer2-2.3b-query-filters`
- **Tests**: 4 tests, 4 commits
- **Dependencies**: Phase 2.3a
- **Scope**: Implement TraceFilters-based querying

#### Phase 2.3c: Dead-Letter Queue
- **Doc**: [phase-2.3c-dead-letter.md](phase-2.3c-dead-letter.md)
- **Branch**: `feat/layer2-2.3c-dead-letter`
- **Tests**: 2 tests, 2 commits
- **Dependencies**: Phase 2.3b
- **Scope**: Implement dead-letter queue

#### Phase 2.3d: Metrics Collection
- **Doc**: [phase-2.3d-metrics.md](phase-2.3d-metrics.md)
- **Branch**: `feat/layer2-2.3d-metrics`
- **Tests**: 2 tests, 2 commits
- **Dependencies**: Phase 2.3c
- **Scope**: Implement metrics collection

### Phase 2.4: Lifecycle Service (3 sub-phases)

#### Phase 2.4a: RuntimeInfo and Engine
- **Doc**: [phase-2.4a-runtimeinfo-and-engine.md](phase-2.4a-runtimeinfo-and-engine.md)
- **Branch**: `feat/layer2-2.4a-runtimeinfo`
- **Tests**: 2 tests, 2 commits
- **Dependencies**: Phase 2.1a
- **Scope**: Add RuntimeInfo type and engine wiring

#### Phase 2.4b: List Returns RuntimeInfo
- **Doc**: [phase-2.4b-list-runtimeinfo.md](phase-2.4b-list-runtimeinfo.md)
- **Branch**: `feat/layer2-2.4b-list`
- **Tests**: 3 tests, 3 commits
- **Dependencies**: Phase 2.4a
- **Scope**: Update List() to return []RuntimeInfo

#### Phase 2.4c: Control Method
- **Doc**: [phase-2.4c-control-method.md](phase-2.4c-control-method.md)
- **Branch**: `feat/layer2-2.4c-control`
- **Tests**: 3 tests, 3 commits
- **Dependencies**: Phase 2.4b
- **Scope**: Implement Control() method

### Phase 2.5: Security Service (2 sub-phases)

#### Phase 2.5a: ValidateAndSanitize and TaintPropagate
- **Doc**: [phase-2.5a-validate-and-taint.md](phase-2.5a-validate-and-taint.md)
- **Branch**: `feat/layer2-2.5a-validate-taint`
- **Tests**: 2 tests, 2 commits
- **Dependencies**: Phase 2.1a
- **Scope**: Implement ValidateAndSanitize() and TaintPropagate()

#### Phase 2.5b: ReportTaints and PrepareContextForBoundary
- **Doc**: [phase-2.5b-report-and-prepare.md](phase-2.5b-report-and-prepare.md)
- **Branch**: `feat/layer2-2.5b-report-prepare`
- **Tests**: 2 tests, 2 commits
- **Dependencies**: Phase 2.5a
- **Scope**: Implement ReportTaints() and PrepareContextForBoundary()

### Phase 2.6: Registry Integration (2 sub-phases)

#### Phase 2.6a: Fix Service Interface
- **Doc**: [phase-2.6a-fix-service-interface.md](phase-2.6a-fix-service-interface.md)
- **Branch**: `feat/layer2-2.6a-fix-interface`
- **Tests**: 0 tests (interface fix only)
- **Dependencies**: None
- **Scope**: Fix Service interface (ID vs Name, HandleMail signature)

#### Phase 2.6b: Registry Implementation
- **Doc**: [phase-2.6b-registry-implementation.md](phase-2.6b-registry-implementation.md)
- **Branch**: `feat/layer2-2.6b-registry`
- **Tests**: 7 tests, 7 commits
- **Dependencies**: Phase 2.6a, 2.2c, 2.3d, 2.4c, 2.5b
- **Scope**: Implement full registry with all services

## Execution Order

```
Phase 2.1a (Core Types)
    ↓
Phase 2.1b (StreamChunk Migration)
    ↓
┌─────────────────────────────────────────────────────┐
│     Phases 2.2-2.5 (can run in parallel)           │
│                                                     │
│  ┌──────────────┐                                   │
│  │  2.2a        │                                   │
│  │  2.2b        │                                   │
│  │  2.2c        │                                   │
│  └──────────────┘                                   │
│                                                     │
│  ┌──────────────┐  ┌──────────────┐  ┌────────────┐│
│  │  2.3a        │  │  2.4a        │  │  2.5a      ││
│  │  2.3b        │  │  2.4b        │  │  2.5b      ││
│  │  2.3c        │  │  2.4c        │  │            ││
│  │  2.3d        │  │              │  │            ││
│  └──────────────┘  └──────────────┘  └────────────┘│
└─────────────────────────────────────────────────────┘
    ↓
Phase 2.6a (Fix Service Interface)
    ↓
Phase 2.6b (Registry Implementation)
```

## Summary

### Test Reduction

| Original Plan | Updated Plan | Reduction |
|---------------|--------------|-----------|
| 6 phases | 16 sub-phases | +10 (more granular) |
| 45 tests | 39 tests | -6 (13% reduction) |
| 45 commits | 39 commits | -6 (13% reduction) |

### Benefits of Sub-Phase Approach

1. **Breaking changes isolated** - Update existing tests in separate sub-phases
2. **Dependencies clearer** - Each sub-phase has minimal dependencies
3. **Easier rollback** - Smaller commits, easier to revert if needed
4. **Parallel execution** - More opportunities for parallel work
5. **Better test focus** - Each sub-phase tests specific functionality

## Related Documents

- [Layer 2 Specification](layer-02-core-services.md) - Original spec
- [Layer 2 Phase Breakdown](layer-02-phase-breakdown.md) - Original breakdown (6 phases)
- [Architecture v1](../arch-v1.md) - Full architecture reference

## Status

All sub-phases are planned and ready for implementation. Follow strict TDD workflow from CLAUDE.md.

## Next Steps

1. **Start with Phase 2.1a** (Core Type Definitions)
2. **Create branch** `feat/layer2-2.1a-core-types`
3. **Follow TDD workflow** for each test
4. **After Phase 2.1b complete**, phases 2.2-2.5 can run in parallel
5. **Phase 2.6a** can be done anytime before 2.6b
6. **Phase 2.6b** integrates everything

## Test Command

After each sub-phase:
```bash
go test ./pkg/services/... -v
```

After all phases:
```bash
go test ./pkg/... -v
```