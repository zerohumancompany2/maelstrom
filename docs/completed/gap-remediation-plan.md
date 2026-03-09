# Gap Remediation Plan

**Generated**: 2026-03-09  
**Scope**: Layer 2 & 3 Audit Gaps (arch-v1.md Sections 7, 9, 10)  
**Parent**: `docs/planning/implementation-scope.md`

---

## Overview

| Metric | Value |
|--------|-------|
| Total gaps | 23 |
| P0 (Critical) | 6 |
| P1 (High) | 7 |
| P2 (Medium) | 8 |
| P3 (Low) | 2 |
| Pre-existing failures | 4 |

### Gap Distribution by Layer

| Layer | P0 | P1 | P2 | P3 | Total |
|-------|----|----|----|----|-------|
| Layer 2 (Core Services) | 3 | 3 | 5 | 0 | 11 |
| Layer 3 (Communication & Mail) | 3 | 4 | 3 | 2 | 12 |

---

## Phase Breakdown

### Phase G1: Critical Fixes (P0)

**Goal**: Fix compilation errors and critical security gaps that block all other work

**Gaps addressed**:
- L3-H4: `ParseAddress` function missing (router.go:12) - compilation error
- L3-C1: `StreamSession.Send()` and `Close()` panic - streaming broken
- L2-C1: `sys:security` boundary enforcement - all methods return nil
- L2-C2: `NamespaceIsolate` method missing from SecurityService
- L2-C3: `CheckTaintPolicy` method missing from SecurityService
- L3-C3: `sys:human-gateway` chat session endpoint missing

**Dependencies**: None

**Estimated**: 12 tests, 12 commits

**Files to modify**:
- `pkg/mail/router.go` - Add ParseAddress
- `pkg/mail/stream.go` - Implement Send/Close
- `pkg/services/security/service.go` - Implement boundary enforcement, add missing methods
- `pkg/services/humangateway/service.go` - Add chat endpoint stub

---

### Phase G2: Core Functionality (P1)

**Goal**: Implement missing core functionality for security and communication

**Gaps addressed**:
- L2-H1: `sys:security` taint propagation is stub
- L2-H2: `sys:communication` at-least-once delivery not guaranteed
- L3-H1: Mail address validation helpers missing
- L3-H2: Request-reply pattern via correlationId not implemented
- L3-H3: Gateway adapter normalization incomplete (no actual servers)

**Dependencies**: G1

**Estimated**: 18 tests, 18 commits

**Files to modify**:
- `pkg/services/security/service.go` - Implement taint propagation
- `pkg/services/communication/service.go` - Add retry logic, delivery tracking
- `pkg/mail/address.go` - Add validation helpers
- `pkg/services/communication/service.go` - Add correlationId tracking for request-reply
- `pkg/gateway/adapters/webhook.go` - Add HTTP server
- `pkg/gateway/adapters/websocket.go` - Add WebSocket server
- `pkg/gateway/adapters/sse.go` - Add SSE endpoint

---

### Phase G3: Integration & Wiring (P1)

**Goal**: Wire services together, fix integration issues

**Gaps addressed**:
- L2-H5: Service registry lacks lifecycle state tracking
- L3-M3: Dead-letter queue integration with sys:observability not wired
- L3-M4: Stream taint stripping not integrated in mail flow
- L2-M5: Service bootstrap sequence not wired to Kernel

**Dependencies**: G1, G2

**Estimated**: 14 tests, 14 commits

**Files to modify**:
- `pkg/services/registry.go` - Add lifecycle state tracking
- `pkg/services/observability/service.go` - Wire dead-letter to communication failures
- `pkg/mail/stream.go` - Integrate taint stripping in mail flow
- `pkg/kernel/kernel.go` - Wire service bootstrap sequence
- `pkg/services/communication/service.go` - Add dead-letter on delivery failure

---

### Phase G4: Gateway & External APIs (P2)

**Goal**: Complete gateway adapters and HTTP endpoints

**Gaps addressed**:
- L3-C2: Gateway HTTP endpoint exposure not implemented (OpenAPI generation missing)
- L2-M4: Hot-reloadable services not implemented (sys:admin, sys:persistence, etc.)
- L3-M2: Topic subscriber interface uses chan Mail but spec says TopicSubscriber

**Dependencies**: G1, G2, G3

**Estimated**: 20 tests, 20 commits

**Files to modify**:
- `pkg/gateway/openapi.go` - Add OpenAPI generation
- `pkg/gateway/gateway.go` - Add HTTP endpoint registration
- `pkg/services/admin/service.go` - Implement admin service
- `pkg/services/persistence/service.go` - Implement persistence service
- `pkg/services/heartbeat/service.go` - Implement heartbeat service
- `pkg/services/memory/service.go` - Implement memory service
- `pkg/mail/router.go` - Fix TopicSubscriber interface

---

### Phase G5: Observability & Metrics (P2)

**Goal**: Complete observability features

**Gaps addressed**:
- L2-H3: `sys:observability` metrics collection incomplete (no transition/event rates)
- L2-M2: `sys:observability` dead-letter query returns copy (memory inefficient)
- L2-M3: `sys:lifecycle` runtime tracking incomplete (no active states update)
- L3-M1: `Mail.Metadata.Stream` type mismatch (spec says bool, impl says *StreamChunk)

**Dependencies**: G1, G2

**Estimated**: 13 tests, 13 commits

**Files to modify**:
- `pkg/services/observability/service.go` - Add transition rate tracking, event rate tracking, fix dead-letter query
- `pkg/services/lifecycle/service.go` - Update runtime tracking on state changes
- `pkg/mail/types.go` - Fix Mail.Metadata.Stream type

---

### Phase G6: Hot-Reload & Advanced Features (P3)

**Goal**: Implement hot-reload and advanced features

**Gaps addressed**:
- L2-H4: `sys:lifecycle` hot-reload not implemented (only spawn/stop)
- L2-M1: `sys:communication` deduplication via correlationId not implemented

**Dependencies**: All above

**Estimated**: 8 tests, 8 commits

**Files to modify**:
- `pkg/services/lifecycle/service.go` - Add hot-reload functionality
- `pkg/services/communication/service.go` - Add correlationId deduplication

---

### Pre-existing: Mail System Integration Tests

**Goal**: Fix known integration test failures (non-blocking)

**Gaps addressed**:
- TestMailSystem_PublishDeliversMail
- TestMailSystem_SubscribeReceivesMail
- TestMailSystem_UnsubscribeRemovesSubscriber
- TestMailSystem_ConcurrentPublish

**Dependencies**: G1, G3

**Estimated**: 4 tests (fix existing), 4 commits

**Files to modify**:
- `pkg/mail/system_test.go` - Fix integration test setup
- `pkg/mail/mailbox.go` - Potential fixes
- `pkg/mail/subscriber.go` - Potential fixes
- `pkg/kernel/kernel.go` - Potential fixes

---

## Summary

| Phase | Priority | Gaps | Tests | Commits | Dependencies |
|-------|----------|------|-------|---------|--------------|
| G1 | P0 | 6 | 12 | 12 | None |
| G2 | P1 | 5 | 18 | 18 | G1 |
| G3 | P1 | 4 | 14 | 14 | G1, G2 |
| G4 | P2 | 3 | 20 | 20 | G1, G2, G3 |
| G5 | P2 | 4 | 13 | 13 | G1, G2 |
| G6 | P3 | 2 | 8 | 8 | All above |
| Mail Integration | Pre-existing | 4 | 4 | 4 | G1, G3 |
| **Total** | | **24** | **89** | **89** | |

---

## Execution Order

1. **G1** (Critical) - Immediate action required
2. **G2** (Core Functionality) - After G1
3. **G3** (Integration) - After G1, G2
4. **G5** (Observability) - Can run parallel to G4 after G1, G2
5. **G4** (Gateway) - After G1, G2, G3
6. **G6** (Hot-Reload) - After all above
7. **Mail Integration** - After G1, G3 (non-blocking)

---

## P0 Gaps (Immediate Action)

| Gap | File | Issue |
|-----|------|-------|
| L3-H4 | pkg/mail/router.go | ParseAddress missing - compilation error |
| L3-C1 | pkg/mail/stream.go | Send/Close panic - streaming broken |
| L2-C1 | pkg/services/security/service.go | Boundary enforcement returns nil |
| L2-C2 | pkg/services/security/service.go | NamespaceIsolate method missing |
| L2-C3 | pkg/services/security/service.go | CheckTaintPolicy method missing |
| L3-C3 | pkg/services/humangateway/service.go | Chat endpoint not implemented |

---

**Plan End**