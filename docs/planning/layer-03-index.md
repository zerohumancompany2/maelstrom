# Layer 3: Communication & Mail System - Index

## Overview

This document provides navigation and summary for Layer 3 implementation, which implements the Communication layer that provides cross-boundary async messaging via a pub-sub mail system.

**Reference**: `docs/arch-v1.md` Section 9  
**Parent Scope**: `docs/planning/implementation-scope.md`  
**Status**: ❌ PENDING

---

## Phase Breakdown Tree

```
Layer 3: Communication & Mail System
│
├── 3.1 Mail Core Types (6 tests)
│   ├── Mail, MailType, MailMetadata, Ack types
│   └── Address validation (agent:, topic:, sys:)
│
├── 3.2 Mail Router (5 tests)
│   └── Route to agent/topic/service
│
├── 3.3 Inboxes & Topics (6 tests)
│   ├── 3.3a AgentInbox & ServiceInbox (3 tests)
│   │   ├── AgentInbox Push/Pop/Subscribe
│   │   └── ServiceInbox Push/Pop
│   │
│   └── 3.3b Topic Pub/Sub (3 tests)
│       └── Topic Subscribe/Unsubscribe/Publish
│
├── 3.4 Publisher/Subscriber (4 tests)
│   ├── Publisher interface + RouterPublisher
│   └── Subscriber interface + ChannelSubscriber
│
├── 3.5 Streaming Support (5 tests)
│   ├── StreamChunk structure
│   ├── UpgradeToStream
│   └── StripForbiddenTaints
│
├── 3.6 Gateway Adapters (6 tests)
│   ├── 3.6a Gateway Core & Webhook (2 tests)
│   │   ├── Gateway service
│   │   └── Webhook adapter
│   │
│   ├── 3.6b SSE & WebSocket (2 tests)
│   │   ├── SSE adapter
│   │   └── WebSocket adapter
│   │
│   └── 3.6c Stub Adapters & Integration (2 tests)
│       ├── PubSub, SMTP, Platform stubs
│       └── Round-trip integration test
│
├── 3.7 Human Gateway Service (4 tests)
│   ├── HandleChat
│   ├── ParseActionItem (@pause, @inject-memory)
│   ├── Session management
│   └── Context sanitization
│
└── 3.8 Integration (2 tests)
    ├── CommunicationService wiring
    └── Full mail flow verification
```

---

## Phase Summary Table

| Phase | Name | Tests | Files | Branch | Dependencies | Document |
|-------|------|-------|-------|--------|--------------|----------|
| 3.1 | Mail Core Types | 6 | 2 | `feat/layer3-mail-types` | None | [phase-3.1-mail-types.md](phase-3.1-mail-types.md) |
| 3.2 | Mail Router | 5 | 1 | `feat/layer3-mail-router` | 3.1 | [phase-3.2-mail-router.md](phase-3.2-mail-router.md) |
| 3.3a | AgentInbox & ServiceInbox | 3 | 1 | `feat/layer3-agent-service-inbox` | 3.1 | [phase-3.3a-agent-and-service-inbox.md](phase-3.3a-agent-and-service-inbox.md) |
| 3.3b | Topic Pub/Sub | 3 | 1 | `feat/layer3-topic-pubsub` | 3.1 | [phase-3.3b-topic-pubsub.md](phase-3.3b-topic-pubsub.md) |
| 3.4 | Publisher/Subscriber | 4 | 2 | `feat/layer3-pubsub` | 3.1, 3.2, 3.3a, 3.3b | [phase-3.4-pubsub.md](phase-3.4-pubsub.md) |
| 3.5 | Streaming Support | 5 | 1 | `feat/layer3-streaming` | 3.1 | [phase-3.5-streaming.md](phase-3.5-streaming.md) |
| 3.6a | Gateway Core & Webhook | 2 | 2 | `feat/layer3-gateway-webhook` | 3.1, 3.5 | [phase-3.6a-gateway-webhook.md](phase-3.6a-gateway-webhook.md) |
| 3.6b | SSE & WebSocket | 2 | 2 | `feat/layer3-sse-websocket` | 3.1, 3.5, 3.6a | [phase-3.6b-sse-websocket.md](phase-3.6b-sse-websocket.md) |
| 3.6c | Stub Adapters & Integration | 2 | 3 | `feat/layer3-stub-adapters` | 3.6a, 3.6b | [phase-3.6c-stubs-integration.md](phase-3.6c-stubs-integration.md) |
| 3.7 | Human Gateway Service | 4 | 2 | `feat/layer3-human-gateway` | 3.1, 3.6 | [phase-3.7-human-gateway.md](phase-3.7-human-gateway.md) |
| 3.8 | Integration | 2 | 2 | `feat/layer3-integration` | 3.2, 3.3a, 3.3b, 3.4, 3.5 | [phase-3.8-integration.md](phase-3.8-integration.md) |

**Total: 11 phases, 42 tests, 19 files**

---

## Execution Order

### Phase 1: Foundation (Sequential)
```
3.1 Mail Core Types
    ↓
┌─────────────────────────────────────┐
│     Phase 2: Core Components        │
│  (Can run in parallel)              │
│                                     │
│  3.2 Mail Router                    │
│  3.3a AgentInbox & ServiceInbox     │
│  3.3b Topic Pub/Sub                 │
│  3.5 Streaming Support              │
└─────────────────────────────────────┘
    ↓
3.4 Publisher/Subscriber
    ↓
┌─────────────────────────────────────┐
│     Phase 3: Gateway Layer          │
│  (Can run in parallel)              │
│                                     │
│  3.6a Gateway Core & Webhook        │
│  3.6b SSE & WebSocket               │
│  3.6c Stub Adapters                 │
│  3.7 Human Gateway                  │
└─────────────────────────────────────┘
    ↓
3.8 Integration
```

### Detailed Execution Plan

1. **Week 1**: Phase 3.1 (Mail Core Types)
   - Branch: `feat/layer3-mail-types`
   - 6 tests, 6 commits

2. **Week 2**: Phases 3.2, 3.3a, 3.3b, 3.5 (Parallel)
   - Branch: `feat/layer3-mail-router`
   - Branch: `feat/layer3-agent-service-inbox`
   - Branch: `feat/layer3-topic-pubsub`
   - Branch: `feat/layer3-streaming`
   - 16 tests total, 16 commits

3. **Week 3**: Phase 3.4 (Publisher/Subscriber)
   - Branch: `feat/layer3-pubsub`
   - 4 tests, 4 commits

4. **Week 4**: Phases 3.6a, 3.6b, 3.6c, 3.7 (Parallel)
   - Branch: `feat/layer3-gateway-webhook`
   - Branch: `feat/layer3-sse-websocket`
   - Branch: `feat/layer3-stub-adapters`
   - Branch: `feat/layer3-human-gateway`
   - 10 tests total, 10 commits

5. **Week 5**: Phase 3.8 (Integration)
   - Branch: `feat/layer3-integration`
   - 2 tests, 2 commits

**Total Estimated Time: 5 weeks**

---

## Test Count by Category

| Category | Tests | Phases |
|----------|-------|--------|
| Type Definitions | 6 | 3.1 |
| Routing | 5 | 3.2 |
| Inboxes | 6 | 3.3a, 3.3b |
| Pub/Sub Interfaces | 4 | 3.4 |
| Streaming | 5 | 3.5 |
| Gateway Adapters | 6 | 3.6a, 3.6b, 3.6c |
| Human Gateway | 4 | 3.7 |
| Integration | 2 | 3.8 |
| **Total** | **42** | **11** |

---

## File Structure

```
maelstrom/
├── pkg/mail/
│   ├── types.go           # 3.1: Mail, MailType, MailMetadata, Ack
│   ├── address.go         # 3.1: Address validation
│   ├── router.go          # 3.2: MailRouter
│   ├── inbox.go           # 3.3a, 3.3b: AgentInbox, ServiceInbox, Topic
│   ├── publisher.go       # 3.4: Publisher interface
│   ├── subscriber.go      # 3.4: Subscriber interface
│   ├── stream.go          # 3.5: StreamChunk, streaming
│   └── mail_test.go       # 3.8: Integration tests
│
├── pkg/gateway/
│   ├── gateway.go         # 3.6a: Gateway service
│   └── adapters/
│       ├── webhook.go     # 3.6a: Webhook adapter
│       ├── sse.go         # 3.6b: SSE adapter
│       ├── websocket.go   # 3.6b: WebSocket adapter
│       ├── pubsub.go      # 3.6c: PubSub stub
│       ├── smtp.go        # 3.6c: SMTP stub
│       └── platform.go    # 3.6c: Platform stubs
│
└── pkg/services/
    ├── communication/
    │   └── service.go     # 3.8: CommunicationService integration
    └── humangateway/
        ├── service.go     # 3.7: HumanGatewayService
        └── chat.go        # 3.7: ChatSession, ActionItem
```

---

## Recursive Split Summary

### Phases That Were Split

| Original Phase | Split Into | Reason |
|----------------|------------|--------|
| 3.3 | 3.3a, 3.3b | Queue vs pub-sub semantics |
| 3.6 | 3.6a, 3.6b, 3.6c | Core vs streaming vs stub adapters |

### Phases That Were Not Split

| Phase | Tests | Reason |
|-------|-------|--------|
| 3.1 | 6 | Tightly coupled types, single feature |
| 3.2 | 5 | Within range, single feature |
| 3.4 | 4 | Within range, single feature |
| 3.5 | 5 | Within range, single feature |
| 3.7 | 4 | Within range, single feature |
| 3.8 | 2 | Integration tests, keep together |

---

## Dependencies Graph

```
                    ┌──────────┐
                    │   3.1    │ Mail Core Types
                    └────┬─────┘
                         │
         ┌───────────────┼───────────────┐
         ↓               ↓               ↓
    ┌──────────┐   ┌──────────┐   ┌──────────┐
    │   3.2    │   │   3.3a   │   │   3.3b   │
    │ MailRouter│  │AgentInbox│  │  Topic   │
    └────┬─────┘   └────┬─────┘   └────┬─────┘
         │              │              │
         └──────────────┼──────────────┘
                        ↓
                   ┌──────────┐
                   │   3.4    │ Publisher/Subscriber
                   └────┬─────┘
                        │
         ┌──────────────┴──────────────┐
         ↓                             ↓
    ┌──────────┐                  ┌──────────┐
    │   3.5    │                  │   3.6a   │
    │Streaming │                  │ Gateway  │
    └────┬─────┘                  └────┬─────┘
         │                             │
         └──────────────┬──────────────┘
                        │
         ┌──────────────┼──────────────┐
         ↓              ↓              ↓
    ┌──────────┐ ┌──────────┐ ┌──────────┐
    │   3.6b   │ │   3.6c   │ │   3.7    │
    │  SSE/WS  │ │  Stubs   │ │ HumanGW  │
    └────┬─────┘ └────┬─────┘ └────┬─────┘
         │            │            │
         └────────────┼────────────┘
                      ↓
                 ┌──────────┐
                 │   3.8    │ Integration
                 └──────────┘
```

---

## Key Types Reference

### Mail Types (11 total)
- `user` - User input / external message
- `assistant` - Assistant response
- `tool_result` - Tool execution result
- `tool_call` - Tool invocation request
- `mail_received` - Mail arrival notification
- `heartbeat` - Scheduled agent wake-up
- `error` - Error condition
- `human_feedback` - Human-in-the-loop feedback
- `partial_assistant` - Streaming LLM output chunk
- `subagent_done` - Sub-agent completion signal
- `taint_violation` - Security policy violation

### Address Formats
- `agent:<id>` - Specific agent inbox
- `topic:<name>` - Pub-sub topic
- `sys:<service>` - System service inbox

### Boundary Types
- `inner` - Inner boundary (trusted)
- `dmz` - DMZ boundary (semi-trusted)
- `outer` - Outer boundary (untrusted)

---

## Next Steps

1. **Start Phase 3.1**: Create branch `feat/layer3-mail-types`
2. **Follow TDD workflow**: Write test → verify RED → Implement → verify GREEN → commit
3. **Repeat for each phase**: One branch per phase, merge after completion
4. **Run full test suite** after Phase 3.8 to verify integration

---

## References

- [Layer 3 Phase Breakdown](layer-03-phase-breakdown.md) - High-level breakdown
- [Layer 3 Specification](layer-03-communication-mail.md) - Full specification
- [Layer 2 Phase Breakdown](../completed/layer-02-phase-breakdown.md) - Reference pattern
- [Architecture v1](../arch-v1.md) - Section 9 (Inter-Agent Communication)

---

## Summary Report

### Deliverables Created

1. ✅ `docs/planning/layer-03-phase-breakdown.md` - High-level breakdown
2. ✅ Individual phase documents (11 files):
   - `phase-3.1-mail-types.md`
   - `phase-3.2-mail-router.md`
   - `phase-3.3a-agent-and-service-inbox.md`
   - `phase-3.3b-topic-pubsub.md`
   - `phase-3.4-pubsub.md`
   - `phase-3.5-streaming.md`
   - `phase-3.6a-gateway-webhook.md`
   - `phase-3.6b-sse-websocket.md`
   - `phase-3.6c-stubs-integration.md`
   - `phase-3.7-human-gateway.md`
   - `phase-3.8-integration.md`
3. ✅ `docs/planning/layer-03-index.md` - This navigation document

### Statistics

| Metric | Value |
|--------|-------|
| Total Phases | 11 |
| Total Tests | 42 |
| Total Commits | 42 (1:1 ratio) |
| Total Files | 19 |
| Phases Split Recursively | 2 (3.3 → 3.3a/3.3b, 3.6 → 3.6a/3.6b/3.6c) |
| Estimated Duration | 5 weeks |

### Execution Order

1. Phase 3.1 (Foundation)
2. Phases 3.2, 3.3a, 3.3b, 3.5 (Parallel)
3. Phase 3.4 (Integration point)
4. Phases 3.6a, 3.6b, 3.6c, 3.7 (Parallel)
5. Phase 3.8 (Final integration)

### TDD Compliance

- ✅ 1:1 test-to-commit ratio maintained
- ✅ One branch per phase
- ✅ Descriptive test names (`Test[Component]_[Behavior]`)
- ✅ Small, focused phases (2-6 tests each)
- ✅ Explicit dependencies between phases
- ✅ Clear acceptance criteria for each test