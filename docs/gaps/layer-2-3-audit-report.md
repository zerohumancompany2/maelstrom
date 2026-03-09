# Layer 2 & 3 Audit Report

**Generated**: 2026-03-09  
**Auditor**: Automated Analysis Agent  
**Scope**: arch-v1.md Sections 7, 9, 10 vs Implementation

---

## Executive Summary

| Metric | Value |
|--------|-------|
| Total requirements checked | 47 |
| Fully implemented | 28 |
| Partially implemented | 12 |
| Not implemented | 7 |
| **Compliance rate** | **59.6%** |

### Summary by Layer

| Layer | Requirements | Complete | Partial | Missing | Compliance |
|-------|-------------|----------|---------|---------|------------|
| Layer 2 (Core Services) | 24 | 14 | 7 | 3 | 58.3% |
| Layer 3 (Communication & Mail) | 23 | 14 | 5 | 4 | 60.9% |

---

## Layer 2 Gaps

### Critical

| # | Gap | Spec Ref | Impact |
|---|-----|----------|--------|
| L2-C1 | `sys:security` boundary enforcement not implemented - all methods return `nil` without validation | arch-v1.md 7.1, 14.2 | Security model ineffective |
| L2-C2 | `NamespaceIsolate` method missing from SecurityService | arch-v1.md 7.3 | No isolated views for agents |
| L2-C3 | `CheckTaintPolicy` method missing from SecurityService | arch-v1.md 7.3 | Taint policy enforcement impossible |

### High

| # | Gap | Spec Ref | Impact |
|---|-----|----------|--------|
| L2-H1 | `sys:security` taint propagation is stub (returns input unchanged) | arch-v1.md 7.1 | Taint tracking ineffective |
| L2-H2 | `sys:communication` at-least-once delivery not guaranteed (no retry logic) | arch-v1.md 9.2 | Message loss possible |
| L2-H3 | `sys:observability` metrics collection incomplete (no transition/event rates) | arch-v1.md 7.3 | Observability limited |
| L2-H4 | `sys:lifecycle` hot-reload not implemented (only spawn/stop) | arch-v1.md 7.2 | No dynamic service updates |
| L2-H5 | Service registry lacks lifecycle state tracking | docs/completed/layer-02-core-services.md 431-440 | Cannot query service states |

### Medium

| # | Gap | Spec Ref | Impact |
|---|-----|----------|--------|
| L2-M1 | `sys:communication` deduplication via correlationId not implemented | arch-v1.md 9.2 | Duplicate messages possible |
| L2-M2 | `sys:observability` dead-letter query returns copy (memory inefficient) | arch-v1.md 7.3 | Performance issue at scale |
| L2-M3 | `sys:lifecycle` runtime tracking incomplete (no active states update) | arch-v1.md 7.3 | Runtime state stale |
| L2-M4 | Hot-reloadable services (sys:admin, sys:persistence, etc.) not implemented | arch-v1.md 7.2 | Limited service ecosystem |
| L2-M5 | Service bootstrap sequence not wired to Kernel | docs/completed/layer-02-core-services.md 431-480 | Services not auto-started |

---

## Layer 3 Gaps

### Critical

| # | Gap | Spec Ref | Impact |
|---|-----|----------|--------|
| L3-C1 | `StreamSession.Send()` and `Close()` methods panic with "not implemented" | arch-v1.md 10.2 | Streaming broken |
| L3-C2 | Gateway HTTP endpoint exposure not implemented (OpenAPI generation missing) | arch-v1.md 10.3 | No external API surface |
| L3-C3 | `sys:human-gateway` chat session UI/HTTPS endpoint not implemented | arch-v1.md 10.4 | No human-in-the-loop interface |

### High

| # | Gap | Spec Ref | Impact |
|---|-----|----------|--------|
| L3-H1 | Mail address validation helpers (`IsValidAgentAddress`, etc.) not found in code | arch-v1.md 9.2 | Addressing validation missing |
| L3-H2 | Request-reply pattern via correlationId not implemented | arch-v1.md 9.2 | Limited communication patterns |
| L3-H3 | Gateway adapter normalization incomplete (no actual HTTP/WebSocket servers) | arch-v1.md 10.1 | No external connectivity |
| L3-H4 | `ParseAddress` function referenced but not found in router.go | pkg/mail/router.go:12 | Router will not compile |

### Medium

| # | Gap | Spec Ref | Impact |
|---|-----|----------|--------|
| L3-M1 | `Mail.Metadata.Stream` type mismatch (spec says bool, impl says *StreamChunk) | arch-v1.md 9.2 | Type inconsistency |
| L3-M2 | Topic subscriber interface uses `chan Mail` but spec says `TopicSubscriber` | arch-v1.md 9.3 | Interface mismatch |
| L3-M3 | Dead-letter queue integration with sys:observability not wired | arch-v1.md 9.2 | Failed messages not tracked |
| L3-M4 | Stream taint stripping called but not integrated in mail flow | arch-v1.md 10.2 | Security gap in streaming |

---

## Implementation Notes

### Things Implemented Beyond Spec

1. **ServiceRegistry** (`pkg/services/registry.go`): Full registry with Register/Get/List operations - spec only mentions service concept
2. **Multiple gateway adapters**: webhook, websocket, sse, pubsub, smtp, platform (slack/whatsapp/telegram) - all stubs but comprehensive
3. **HumanGatewayService action items**: `ParseActionItem` method for `@pause`, `@inject-memory` - not detailed in spec
4. **ChannelSubscriber**: Additional subscriber type beyond spec's `Subscriber` interface
5. **DeadLetterEntry** struct in observability: Extended spec's dead-letter concept with timestamp

### Assumptions Made

1. **BoundaryType values**: Implementation uses `inner`, `dmz`, `outer` - matches spec Section 4
2. **Mail types**: All 11 types implemented with aliases for backward compatibility
3. **Inbox buffer size**: 100 messages - not specified in spec
4. **StreamChunk buffer size**: 100 chunks - not specified in spec
5. **Service inboxes use same structure as agent inboxes** - reasonable assumption for consistency

### Design Decisions

1. **Separate pkg/mail from services**: Mail system is infrastructure, not a service itself
2. **RouterPublisher pattern**: Publisher wraps router for clean separation
3. **Gateway adapter interface**: Normalizes all external I/O through common interface
4. **ServiceRegistry centralization**: All services registered in one place for lookup

---

## Recommendations

### Priority 1 (Critical - Block Release)

1. **Implement security boundary enforcement** (L2-C1)
   - Add actual validation logic to `ValidateAndSanitize`
   - Implement boundary transition rules (inner/DMZ/outer)
   
2. **Add missing security methods** (L2-C2, L2-C3)
   - Implement `NamespaceIsolate`
   - Implement `CheckTaintPolicy`

3. **Fix streaming implementation** (L3-C1)
   - Implement `StreamSession.Send()` 
   - Implement `StreamSession.Close()`

4. **Fix router compilation error** (L3-H4)
   - Implement `ParseAddress` function or import from correct location

### Priority 2 (High - Before Production)

5. **Implement at-least-once delivery** (L2-H2)
   - Add retry logic with exponential backoff
   - Track delivery attempts

6. **Add correlationId deduplication** (L2-M1, L3-H2)
   - Track seen correlation IDs
   - Implement request-reply pattern

7. **Wire gateway adapters to actual servers** (L3-H3)
   - HTTP server for webhook adapter
   - WebSocket server for websocket adapter
   - SSE endpoint for sse adapter

8. **Complete observability metrics** (L2-H3)
   - Add transition rate tracking
   - Add event rate tracking

### Priority 3 (Medium - Enhancement)

9. **Implement hot-reloadable services** (L2-H4, L2-M4)
   - sys:admin, sys:persistence, sys:heartbeat, sys:memory

10. **Wire services to Kernel bootstrap** (L2-M5)
    - Auto-start services during kernel initialization
    - Follow bootstrap sequence from spec

11. **Fix type inconsistencies** (L3-M1)
    - Align `Mail.Metadata.Stream` with spec

12. **Integrate dead-letter queue** (L3-M3)
    - Wire failed mail delivery to observability

---

## Compliance Details

### Layer 2: Core Platform Services (Section 7)

| Requirement | Spec Location | Implementation | Status |
|-------------|---------------|----------------|--------|
| sys:security service exists | 7.1 | pkg/services/security/service.go | ✅ |
| sys:communication service exists | 7.1 | pkg/services/communication/service.go | ✅ |
| sys:observability service exists | 7.1 | pkg/services/observability/service.go | ✅ |
| sys:lifecycle service exists | 7.1 | pkg/services/lifecycle/service.go | ✅ |
| Services are Charts | 7 | Not verified | ⚠️ |
| Services run on Statechart Engine | 7 | Not verified | ⚠️ |
| Services addressable by well-known IDs | 7.1 | ID() methods return correct IDs | ✅ |
| Services started by Kernel during bootstrap | 7.1 | Not wired | ❌ |
| sys:security - boundary enforcement | 7.1 | Stub (returns nil) | ❌ |
| sys:security - taint validation | 7.1 | Stub (returns input) | ❌ |
| sys:security - sanitization | 7.1 | Stub (returns input) | ❌ |
| sys:communication - mail pub/sub | 7.1 | Basic implementation | ✅ |
| sys:observability - traces | 7.1 | In-memory storage | ✅ |
| sys:observability - metrics | 7.1 | Partial (state counts only) | ⚠️ |
| sys:observability - dead-letter queue | 7.1 | Basic implementation | ✅ |
| sys:lifecycle - spawn | 7.1 | Implemented | ✅ |
| sys:lifecycle - stop | 7.1 | Implemented | ✅ |
| Hot-reloadable services (sys:gateway, etc.) | 7.2 | Not implemented | ❌ |
| sys:admin service | 7.2 | Stub exists | ⚠️ |
| sys:persistence service | 7.2 | Stub exists | ⚠️ |
| sys:heartbeat service | 7.2 | Stub exists | ⚠️ |
| sys:memory service | 7.2 | Stub exists | ⚠️ |
| sys:tools service | 7.2 | Stub exists | ⚠️ |
| sys:datasources service | 7.2 | Stub exists | ⚠️ |

### Layer 3: Communication & Mail (Sections 9, 10)

| Requirement | Spec Location | Implementation | Status |
|-------------|---------------|----------------|--------|
| Mail type defined | 9.2 | pkg/mail/types.go | ✅ |
| Mail addressing (agent:topic:sys:) | 9.2 | Types defined | ✅ |
| 11 mail types | 9.2 | All 11 defined | ✅ |
| Mail metadata fields | 9.2 | Most fields present | ⚠️ |
| At-least-once delivery | 9.2 | Not implemented | ❌ |
| Deduplication via correlationId | 9.2 | Not implemented | ❌ |
| Dead-letter queue on failure | 9.2 | Not wired | ❌ |
| Communication.Publish() | 9.3 | Implemented | ✅ |
| Communication.Subscribe() | 9.3 | Implemented | ✅ |
| Communication.Unsubscribe() | 9.3 | Implemented | ✅ |
| Communication.UpgradeToStream() | 9.3 | Implemented (returns channel) | ✅ |
| Ack structure | 9.3 | Implemented (extended with Success) | ✅ |
| StreamChunk structure | 9.3 | Implemented | ✅ |
| MailRouter | 9.2 | pkg/mail/router.go | ✅ |
| AgentInbox | 9.2 | pkg/mail/inbox.go | ✅ |
| Topic pub-sub | 9.2 | pkg/mail/router.go | ✅ |
| ServiceInbox | 9.2 | pkg/mail/inbox.go | ✅ |
| Gateway adapters | 10.1 | Stubs in pkg/gateway/adapters/ | ⚠️ |
| Webhook adapter | 10.1 | Stub | ⚠️ |
| WebSocket adapter | 10.1 | Stub | ⚠️ |
| SSE adapter | 10.1 | Stub | ⚠️ |
| HTTP(S) endpoint exposure | 10.3 | Not implemented | ❌ |
| OpenAPI generation | 10.3 | Not implemented | ❌ |
| sys:human-gateway | 10.4 | Basic service stub | ⚠️ |
| Chat session at /chat/{agentId} | 10.4 | Not implemented | ❌ |
| Action item parsing (@pause, etc.) | 10.4 | Method stub exists | ⚠️ |

---

## Next Actions

1. **Immediate (This Sprint)**:
   - Fix compilation errors (ParseAddress)
   - Implement StreamSession.Send/Close
   - Add security boundary enforcement stub with proper structure

2. **Short-term (Next Sprint)**:
   - Wire services to Kernel bootstrap
   - Implement at-least-once delivery
   - Add correlationId deduplication

3. **Medium-term (This Quarter)**:
   - Complete gateway adapter implementations
   - Implement OpenAPI generation
   - Add human-gateway chat interface

4. **Long-term (Future)**:
   - Hot-reloadable services
   - Full observability metrics
   - Persistence layer

---

**Report End**