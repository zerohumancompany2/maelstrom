# Missing Components Summary - Layers 0-8

**Generated**: 2026-03-10  
**Scope**: Comprehensive gap analysis across all layers to achieve arch-v1.md vision  
**Based On**: Findings from 9 previous agent reports

---

## Executive Summary

| Metric | Value |
|--------|-------|
| Total Gaps Identified | 73 |
| Critical Gaps | 6 |
| High Priority Gaps | 18 |
| Medium Priority Gaps | 29 |
| Low Priority Gaps | 20 |
| Estimated Implementation Effort | ~250 tests, ~250 commits |

### Overall Status by Layer

| Layer | Status | Completion | Blocking Issues |
|-------|--------|------------|-----------------|
| Layer 0 (Kernel & Bootstrap) | ⚠️ Partial | ~60% | 3 critical |
| Layer 1 (Core Platform Services) | ⚠️ Partial | ~58% | 3 critical |
| Layer 2 (Communication & Mail) | ⚠️ Partial | ~61% | 3 critical |
| Layer 3 (Security & Boundaries) | ✅ Planned | 0% impl, 100% planned | 45 phase plans ready |
| Layer 4 (Hot-Reloadable Services) | ✅ Planned | 0% impl, 100% planned | 23 phase plans ready |
| Layer 5 (Tool Calls & Orchestration) | ✅ Planned | 0% impl, 100% planned | 32 tests planned |
| Layer 6 (Agent Charts & LLM) | ✅ Planned | 0% impl, 100% planned | 44 tests planned |
| Layer 7 (Streaming & Gateway) | ✅ Planned | 0% impl, 100% planned | 34 tests planned |
| Layer 8 (End-to-End Integration) | ❌ Not Started | 0% | No phase plans |

---

## 1. CRITICAL GAPS (Blocking Functional Application)

### COMPILATION BLOCKERS (Must Fix First)

| ID | Gap | File Path | Impact |
|----|-----|-----------|--------|
| CB-1 | `StreamChunk.Data` field missing but used in 30+ test locations | `pkg/mail/types.go`, `pkg/mail/stream_test.go` | **Tests will not compile** |
| CB-2 | `services.Service.HandleMail` signature mismatch (returns `error` vs `*OutcomeEvent`) | `pkg/e2e/services_test.go:101-371` | **E2E tests will not compile** |
| CB-3 | `serviceWrapper` does not implement `services.Service` interface | `pkg/e2e/services_test.go:101` | **Service tests broken** |

### Layer 0: Kernel & Bootstrap

| ID | Gap | Spec Ref | File Path | Impact |
|----|-----|----------|-----------|--------|
| L0-C1 | `ParseAddress` function missing - router.go references but function not found | docs/gaps/layer-2-3-audit-report.md:78 | `pkg/mail/router.go:12` | **Router will not compile** |
| L0-C2 | KERNEL_READY event not emitted during bootstrap | docs/completed/layer-01-audit-report.md | `pkg/kernel/bootstrap.go` | Services never start post-bootstrap |
| L0-C3 | ChartRegistry not started after KERNEL_READY | docs/completed/layer-07/plans/phase-7.3-kernel-ready-loading.md | `pkg/kernel/bootstrap.go` | No chart loading possible |

### Layer 1: Core Platform Services

| ID | Gap | Spec Ref | File Path | Impact |
|----|-----|----------|-----------|--------|
| L1-C1 | `sys:security` boundary enforcement returns nil without validation | docs/gaps/layer-2-3-audit-report.md:34 | `pkg/services/security/service.go` | **Security model completely ineffective** |
| L1-C2 | `NamespaceIsolate` method missing from SecurityService | docs/gaps/layer-2-3-audit-report.md:35 | `pkg/services/security/service.go` | No isolated views for agents |
| L1-C3 | `CheckTaintPolicy` method missing from SecurityService | docs/gaps/layer-2-3-audit-report.md:36 | `pkg/services/security/service.go` | Taint policy enforcement impossible |

### Layer 2: Communication & Mail

| ID | Gap | Spec Ref | File Path | Impact |
|----|-----|----------|-----------|--------|
| L2-C1 | `StreamSession.Send()` panics with "not implemented" | docs/gaps/layer-2-3-audit-report.md:66 | `pkg/mail/stream.go` | **Streaming completely broken** |
| L2-C2 | `StreamSession.Close()` panics with "not implemented" | docs/gaps/layer-2-3-audit-report.md:66 | `pkg/mail/stream.go` | Cannot close streams |
| L2-C3 | Gateway HTTP endpoint exposure not implemented | docs/gaps/layer-2-3-audit-report.md:67 | `pkg/services/gateway/service.go` | **No external API surface** |

---

## 2. HIGH PRIORITY GAPS (Needed for Core Functionality)

### Layer 0: Kernel & Bootstrap

| ID | Gap | Spec Ref | File Path |
|----|-----|----------|-----------|
| L0-H1 | Service bootstrap sequence not wired to Kernel | docs/gaps/layer-2-3-audit-report.md:56 | `pkg/kernel/bootstrap.go` |
| L0-H2 | Post-KERNEL_READY service loading not implemented | docs/completed/layer-07/plans/phase-7.3-kernel-ready-loading.md | `pkg/kernel/bootstrap.go` |
| L0-H3 | Kernel dormant state not implemented | docs/completed/layer-07/plans/phase-7.3-kernel-ready-loading.md | `pkg/kernel/state.go` |

### Layer 1: Core Platform Services

| ID | Gap | Spec Ref | File Path |
|----|-----|----------|-----------|
| L1-H1 | `sys:security` taint propagation is stub (returns input unchanged) | docs/gaps/layer-2-3-audit-report.md:41 | `pkg/services/security/service.go` |
| L1-H2 | `sys:communication` at-least-once delivery not guaranteed | docs/gaps/layer-2-3-audit-report.md:42 | `pkg/services/communication/service.go` |
| L1-H3 | `sys:observability` metrics collection incomplete | docs/gaps/layer-2-3-audit-report.md:43 | `pkg/services/observability/service.go` |
| L1-H4 | `sys:lifecycle` hot-reload not implemented | docs/gaps/layer-2-3-audit-report.md:44 | `pkg/services/lifecycle/service.go` |
| L1-H5 | Service registry lacks lifecycle state tracking | docs/gaps/layer-2-3-audit-report.md:45 | `pkg/services/registry.go` |

### Layer 2: Communication & Mail

| ID | Gap | Spec Ref | File Path |
|----|-----|----------|-----------|
| L2-H1 | Mail address validation helpers not found | docs/gaps/layer-2-3-audit-report.md:73 | `pkg/mail/types.go` |
| L2-H2 | Request-reply pattern via correlationId not implemented | docs/gaps/layer-2-3-audit-report.md:74 | `pkg/mail/router.go` |
| L2-H3 | Gateway adapter normalization incomplete (no actual servers) | docs/gaps/layer-2-3-audit-report.md:75 | `pkg/services/gateway/adapter.go` |
| L2-H4 | HTTP server for webhook adapter missing | docs/completed/layer-08/plans/phase-8.1-channel-adapters.md | `pkg/services/gateway/webhook_adapter.go` |
| L2-H5 | WebSocket server for websocket adapter missing | docs/completed/layer-08/plans/phase-8.1-channel-adapters.md | `pkg/services/gateway/websocket_adapter.go` |
| L2-H6 | SSE endpoint for sse adapter missing | docs/completed/layer-08/plans/phase-8.1-channel-adapters.md | `pkg/services/gateway/sse_adapter.go` |

### Layer 3: Security & Boundaries (Phase Plans Exist - Not Implemented)

| ID | Gap | Spec Ref | Phase Plan |
|----|-----|----------|------------|
| L3-H1 | Taint Engine Core (Attach, Propagate, Strip) | docs/completed/layer-4/FINAL-COVERAGE-REPORT.md:22-29 | P1.x (5 phases, 19 tests) |
| L3-H2 | DataSource Implementations (LocalDisk, S3, InMemory, Registry) | docs/completed/layer-4/FINAL-COVERAGE-REPORT.md:32-38 | P2.1-P2.4 (4 phases, 15 tests) |
| L3-H3 | Boundary Enforcement (Validation, Transitions, Violations) | docs/completed/layer-4/FINAL-COVERAGE-REPORT.md:42-49 | P3.1-P3.3, P3.5 (4 phases, 19 tests) |

---

## 3. MEDIUM PRIORITY GAPS (Important but Not Blocking)

### Layer 0: Kernel & Bootstrap

| ID | Gap | Spec Ref |
|----|-----|----------|
| L0-M1 | Service directory not populated during bootstrap | docs/completed/layer-07/plans/phase-7.3-kernel-ready-loading.md |
| L0-M2 | Services YAML schema validation missing | docs/completed/layer-05/plans/P10.1-platform-service-yaml-schema.md |

### Layer 1: Core Platform Services

| ID | Gap | Spec Ref |
|----|-----|----------|
| L1-M1 | `sys:communication` deduplication via correlationId | docs/gaps/layer-2-3-audit-report.md:51 |
| L1-M2 | `sys:observability` dead-letter query returns copy (memory inefficient) | docs/gaps/layer-2-3-audit-report.md:52 |
| L1-M3 | `sys:lifecycle` runtime tracking incomplete | docs/gaps/layer-2-3-audit-report.md:53 |
| L1-M4 | Hot-reloadable services not implemented | docs/gaps/layer-2-3-audit-report.md:54 |
| L1-M5 | sys:admin service only stub | docs/completed/layer-05/plans/P2.1-admin-core.md |
| L1-M6 | sys:persistence service only stub | docs/completed/layer-05/plans/P3.1-persistence-core.md |
| L1-M7 | sys:heartbeat service only stub | docs/completed/layer-05/plans/P4.1-heartbeat-service-core.md |
| L1-M8 | sys:memory service only stub | docs/completed/layer-05/plans/P5.1-memory-service-vectorstore.md |
| L1-M9 | sys:tools service only stub | docs/completed/layer-05/plans/P6.1-toolregistry-core.md |
| L1-M10 | sys:datasources service only stub | docs/completed/layer-05/plans/P7.1-datasourceservice-core.md |

### Layer 2: Communication & Mail

| ID | Gap | Spec Ref |
|----|-----|----------|
| L2-M1 | `Mail.Metadata.Stream` type mismatch | docs/gaps/layer-2-3-audit-report.md:81 |
| L2-M2 | Topic subscriber interface uses `chan Mail` but spec says `TopicSubscriber` | docs/gaps/layer-2-3-audit-report.md:82 |
| L2-M3 | Dead-letter queue integration with sys:observability not wired | docs/gaps/layer-2-3-audit-report.md:83 |
| L2-M4 | Stream taint stripping not integrated in mail flow | docs/gaps/layer-2-3-audit-report.md:84 |

### Layer 3: Security & Boundaries (Phase Plans Exist)

| ID | Gap | Phase Plan |
|----|-----|------------|
| L3-M1 | ContextMap & Isolation (Token Budgeting, Filtering, Sources) | P4.1-P4.3 (3 phases, 14 tests) |
| L3-M2 | Stream Sanitization (Taint Stripping, Reporting) | P5.1-StreamTaintStripping, P5.2-TaintReporting |
| L3-M3 | Integration & Wiring (Security Service, Mail Router Hooks) | P6.1-P6.4 (4 phases, 15 tests) |

### Layer 4: Hot-Reloadable Services (Phase Plans Exist)

| ID | Gap | Phase Plan |
|----|-----|------------|
| L4-M1 | sys:human-gateway chat session | P8.1, P8.2 (8 tests) |
| L4-M2 | Hot-reload protocol (quiescence, history, context transform) | P9.1-P9.3 (11 tests) |
| L4-M3 | Platform service YAML schema | P10.1 (3 tests) |

### Layer 5: Tool Calls & Orchestration (Phase Plans Exist)

| ID | Gap | Phase Plan |
|----|-----|------------|
| L5-M1 | ToolDescriptor missing fields (TaintOutput, Type, ChartRef, MaxIterations) | docs/completed/layer-06/plans/phase-6.1-tool-descriptor.md |
| L5-M2 | Execution policies not implemented (seq_continue, seq_failfast, par_continue) | docs/completed/layer-06/plans/phase-6.2-execution-policies.md |
| L5-M3 | OrchestratorService not implemented | docs/completed/layer-06/plans/phase-6.3-orchestrator-service.md |
| L5-M4 | Sequential and Parallel executors not implemented | docs/completed/layer-06/plans/phase-6.4-6.5 |
| L5-M5 | Sub-agent lifecycle management not implemented | docs/completed/layer-06/plans/phase-6.6-subagent.md |

### Layer 6: Agent Charts & LLM (Phase Plans Exist)

| ID | Gap | Phase Plan |
|----|-----|------------|
| L6-M1 | Hard-coded services completion (sys:security, sys:communication, etc.) | docs/completed/layer-07/plans/phase-7.1-hardcoded-services.md (6 tests) |
| L6-M2 | Hot-reloadable services completion (all 8 services) | docs/completed/layer-07/plans/phase-7.2-hotreloadable-services.md (10 tests) |
| L6-M3 | Service-specific contracts (adminCommand, snapshot, emit/query, etc.) | docs/completed/layer-07/plans/phase-7.5-service-contracts.md (8 tests) |

### Layer 7: Streaming & Gateway (Phase Plans Exist)

| ID | Gap | Phase Plan |
|----|-----|------------|
| L7-M1 | HTTP(S) exposure & OpenAPI generation | docs/completed/layer-08/plans/phase-8.3-http-openapi.md (6 tests) |
| L7-M2 | Human chat with running agents | docs/completed/layer-08/plans/phase-8.4-human-chat.md (6 tests) |
| L7-M3 | Gateway security & tainting | docs/completed/layer-08/plans/phase-8.5-gateway-security.md (6 tests) |

---

## 4. LOW PRIORITY GAPS (Nice-to-Have)

### Layer 1: Core Platform Services

| ID | Gap |
|----|-----|
| L1-L1 | sys:admin 2FA enforcement |
| L1-L2 | sys:persistence storage backends |
| L1-L3 | sys:heartbeat agent wake functionality |
| L1-L4 | sys:memory graph store |
| L1-L5 | sys:tools boundary filtering |
| L1-L6 | sys:datasources taint tracking |

### Layer 3: Security & Boundaries

| ID | Gap |
|----|-----|
| L3-L1 | E2E tests for taint engine (P7.1-P7.5) |
| L3-L2 | Taint reporting UI |
| L3-L3 | Advanced context block policies |

### Layer 4: Hot-Reloadable Services

| ID | Gap |
|----|-----|
| L4-L1 | Services integration E2E tests |
| L4-L2 | Mail integration tests for services |

### Layer 5-7: Integration & Polish

| ID | Gap |
|----|-----|
| L5-L1 | Full integration tests for tool orchestration |
| L6-L1 | Service discovery E2E tests |
| L7-L1 | Full Layer 8 integration tests |
| L7-L2 | PubSub adapter implementation |
| L7-L3 | Slack/WhatsApp/Telegram adapters |

---

## 5. RECOMMENDED IMPLEMENTATION ORDER

### Phase 0: Compilation Fixes (Days 1-3)
**Goal**: Fix compilation errors blocking all tests

1. **CB-1**: Add `Data` field to `StreamChunk` or fix all test references
   - File: `pkg/mail/types.go`
   - Tests: Fix 30+ test locations across stream_test.go, types_test.go, stream_sanitization_test.go
   
2. **CB-2, CB-3**: Fix `HandleMail` signature mismatch in service tests
   - File: `pkg/e2e/services_test.go`, `pkg/services/types.go`
   - Tests: Align interface and implementation

**Branch**: `fix/compilation-errors`  
**Commits**: ~5 commits

### Phase 1: Critical Fixes (Week 1-2)
**Goal**: Make the system compile and have basic security

1. **L0-C1**: Implement `ParseAddress` function
   - File: `pkg/mail/types.go`
   - Tests: 4 tests for address parsing/validation
   
2. **L2-C1, L2-C2**: Implement `StreamSession.Send()` and `Close()`
   - File: `pkg/mail/stream.go`
   - Tests: 6 tests for streaming
   
3. **L1-C1, L1-C2, L1-C3**: Implement SecurityService core methods
   - File: `pkg/services/security/service.go`
   - Tests: Start with Layer 4 Phase P1 (Taint Engine Core) - 19 tests

**Branch**: `fix/critical-compilation-security`  
**Commits**: ~25 commits (1:1 test-to-commit ratio)

### Phase 2: Core Infrastructure (Week 3-4)
**Goal**: Enable basic mail flow and service bootstrap

1. **L0-C2, L0-C3**: Implement KERNEL_READY event and ChartRegistry startup
   - File: `pkg/kernel/bootstrap.go`
   - Tests: Layer 7 Phase 7.3 (6 tests)
   
2. **L1-H1**: Implement taint propagation
   - File: `pkg/services/security/service.go`
   - Tests: Layer 4 Phase P1.2 (5 tests)
   
3. **L2-H1**: Implement mail address validation helpers
   - File: `pkg/mail/types.go`
   - Tests: 4 tests

**Branch**: `feat/core-infrastructure`  
**Commits**: ~15 commits

### Phase 3: Security Layer Completion (Week 5-6)
**Goal**: Complete Layer 4 security implementation

1. Complete all Layer 4 Phase P1 (Taint Engine Core) - 19 tests
2. Complete Layer 4 Phase P2 (DataSource Implementations) - 22 tests
3. Complete Layer 4 Phase P3 (Boundary Enforcement) - 26 tests

**Branch**: `feat/layer4-security-core`  
**Commits**: ~67 commits

### Phase 4: Hot-Reloadable Services (Week 7-8)
**Goal**: Implement all hot-reloadable platform services

1. Layer 5 Phase P1 (Gateway Core) - 10 tests
2. Layer 5 Phase P2 (Admin Core + 2FA) - 7 tests
3. Layer 5 Phase P3 (Persistence Core + Storage) - 9 tests
4. Layer 5 Phase P4 (Heartbeat Service) - 6 tests
5. Layer 5 Phase P5 (Memory Service) - 8 tests
6. Layer 5 Phase P6 (Tool Registry) - 7 tests
7. Layer 5 Phase P7 (DataSource Service) - 8 tests
8. Layer 5 Phase P8 (Human Gateway Service) - 8 tests

**Branch**: `feat/layer5-platform-services`  
**Commits**: ~63 commits

### Phase 5: Gateway & Streaming (Week 9-10)
**Goal**: Enable external connectivity and streaming

1. Layer 8 Phase 8.1 (Channel Adapters) - 6 tests
2. Layer 8 Phase 8.2 (Streaming Path) - 6 tests
3. Layer 8 Phase 8.3 (HTTP & OpenAPI) - 6 tests
4. Layer 8 Phase 8.4 (Human Chat) - 6 tests
5. Layer 8 Phase 8.5 (Gateway Security) - 6 tests

**Branch**: `feat/layer8-gateway-streaming`  
**Commits**: ~30 commits

### Phase 6: Tool Orchestration (Week 11-12)
**Goal**: Complete Layer 6 tool calls and orchestration

1. Layer 6 Phase 6.1 (Tool Descriptor) - 4 tests
2. Layer 6 Phase 6.2 (Execution Policies) - 5 tests
3. Layer 6 Phase 6.3 (Orchestrator Service) - 6 tests
4. Layer 6 Phase 6.4 (Sequential Executor) - 5 tests
5. Layer 6 Phase 6.5 (Parallel Executor) - 5 tests
6. Layer 6 Phase 6.6 (Sub-Agent Support) - 5 tests

**Branch**: `feat/layer6-orchestration`  
**Commits**: ~30 commits

### Phase 7: Agent Charts & LLM Integration (Week 13-14)
**Goal**: Complete Layer 7 agent charts and LLM integration

1. Layer 7 Phase 7.1 (Hard-Coded Services) - 6 tests
2. Layer 7 Phase 7.2 (Hot-Reloadable Services) - 10 tests
3. Layer 7 Phase 7.3 (Kernel Ready Loading) - 6 tests
4. Layer 7 Phase 7.4 (Hot-Reload Protocol) - 7 tests
5. Layer 7 Phase 7.5 (Service Contracts) - 8 tests
6. Layer 7 Phase 7.6 (Service Discovery) - 4 tests

**Branch**: `feat/layer7-agent-charts`  
**Commits**: ~41 commits

### Phase 8: Integration & Polish (Week 15-16)
**Goal**: End-to-end integration and remaining gaps

1. All E2E integration tests from Layers 4-8
2. Remaining low priority gaps
3. Performance optimization and documentation

**Branch**: `feat/integration-polish`  
**Commits**: ~50+ commits

---

## Summary Statistics

### By Priority

| Priority | Count | Estimated Tests | Estimated Commits |
|----------|-------|-----------------|-------------------|
| Compilation Blockers | 3 | 5 | 5 |
| Critical | 6 | 25 | 25 |
| High | 18 | 80 | 80 |
| Medium | 29 | 120 | 120 |
| Low | 20 | 25 | 25 |
| **Total** | **76** | **~255** | **~255** |

### By Layer (Implementation Status)

| Layer | Planned Tests | Implemented Tests | Remaining |
|-------|---------------|-------------------|-----------|
| Layer 0 | 10 | 0 | 10 |
| Layer 1 | 87 | 0 | 87 |
| Layer 2 | 34 | 0 | 34 |
| Layer 3 | 147 | 0 | 147 |
| Layer 4 | 87 | 0 | 87 |
| Layer 5 | 32 | 0 | 32 |
| Layer 6 | 44 | 0 | 44 |
| Layer 7 | 34 | 0 | 34 |
| **Total** | **475** | **0** | **475** |

**Note**: Layer 3-7 have complete phase plans but 0% implementation. Layers 0-2 have partial implementation with critical gaps.

---

*Report generated: 2026-03-10*  
*Based on: arch-v1.md v1.3, 9 agent reports, 45+ phase plans*