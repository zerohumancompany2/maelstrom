# Maelstrom Implementation Scope

**Date**: 2026-03-05  
**Status**: Layer 0 Complete, Layers 1-8 Pending  
**Based on**: `docs/arch-v1.md` (v1.3)

---

## Executive Summary

Maelstrom is a zero-human, statechart-native agentic runtime where **every behavior** is expressed as a statechart. The architecture is organized into **8 architectural layers** (bottom-up), each depending on the layer below it.

**Current State**: Layer 0 (Statechart Engine) is **complete** with 73 tests passing.

**Target State**: Full MVP spanning all 8 layers through the external gateway interface.

**Dependency Chain**: Maximum depth is 5 layers (acceptable for MVP). Example: UI → Gateway → Communication → Security → Engine.

---

## Architectural Layers Overview

```
Maelstrom Runtime (MVP)
├── Layer 8: External Interfaces (Gateway + Streaming + OpenAPI)
│   └── Depends on: Layer 7 (Agents), Layer 5 (sys:gateway)
│
├── Layer 7: Agent Charts & LLM Integration
│   ├── Agent Charts (OODA loop as statecharts)
│   ├── LLM Bridge (ContextMap → LLM → tool calls)
│   └── Depends on: Layer 6 (Orchestrator), Layer 5 (sys:memory)
│
├── Layer 6: Tool Calls & Orchestration
│   ├── Tool Registry (boundary-aware schemas)
│   ├── Orchestrator (reusable sub-chart for execution)
│   └── Depends on: Layer 5 (sys:tools), Layer 4 (boundaries)
│
├── Layer 5: Hot-Reloadable Platform Services (YAML)
│   ├── sys:gateway, sys:admin, sys:persistence, sys:heartbeat
│   ├── sys:memory, sys:tools, sys:datasources, sys:human-gateway
│   └── Depends on: Layer 4 (Security enforces boundaries)
│
├── Layer 4: Security & Boundaries + DataSources
│   ├── Boundary enforcement (inner/DMZ/outer)
│   ├── Taint engine (DLP, redaction, sanitization)
│   ├── ContextMap filtering
│   └── Depends on: Layer 3 (Mail for violation reports)
│
├── Layer 3: Communication & Mail System
│   ├── Mail types (user, assistant, tool_call, etc.)
│   ├── Pub/sub backbone (agent:<id>, topic:<name>, sys:<service>)
│   └── Depends on: Layer 2 (sys:communication online)
│
├── Layer 2: Core Platform Services (Hard-Coded)
│   ├── sys:security, sys:communication, sys:observability, sys:lifecycle
│   └── Depends on: Layer 1 (Bootstrap Chart spawns them)
│
├── Layer 1: Kernel & Bootstrap Infrastructure
│   ├── Kernel (~50 lines, minimal bootstrap container)
│   ├── ChartRegistry (YAML loading, hydration, file watching)
│   └── Bootstrap Chart (hard-coded YAML, sequential compound state)
│
└── Layer 0: Statechart Engine ✓ COMPLETE
    ├── Atomic, compound, parallel states
    ├── Event dispatch, transitions, guards, actions
    ├── Snapshot/restore, hot-reload, transient sub-charts
    └── Trace callbacks for observability
```

---

## Layer 0: Statechart Engine ✓ COMPLETE

**Status**: Fully implemented, 73 tests passing  
**Files**: `pkg/statechart/*.go` (7 source files, 22 test files)  
**Reference**: `docs/arch-v1.md` Section 5

### Implemented Features

| Feature | Status | Notes |
|---------|--------|-------|
| Atomic states | ✅ | Leaf states with no substates |
| Compound states | ✅ | Sequential execution with default children |
| Parallel states | ✅ | Goroutine-per-region with hierarchical routing |
| Event dispatch | ✅ | FIFO ordering, run-to-completion semantics |
| Guards & Actions | ✅ | Registry-based, synchronous execution |
| Entry/Exit actions | ✅ | Proper ordering (exit before entry) |
| Snapshot/Restore | ✅ | State capture with region preservation |
| Hot-reload | ✅ | ReplaceDefinition control command |
| Transient sub-charts | ✅ | Lifecycle binding to parent |
| Trace callbacks | ✅ | OnTransition, OnStateEntry, OnStateExit, OnEventDispatch |
| Dynamic reclassification | ✅ | atomic ↔ compound/parallel at runtime |

### Known TODOs

| Location | Issue | Severity |
|----------|-------|----------|
| `runtime.go:143` | Quiescence detection for parallel states returns conservative `false` | LOW |

**Quiescence Definition** (from arch-v1.md Section 12.3):  
A ChartRuntime is quiescent when:

1. Event queue is empty (no pending events)
2. No active parallel regions are processing events
3. No inflight tool calls or sub-agent invocations (Orchestrator idle)

**Impact**: Conservative approach is safe; snapshot/restore work correctly. Optional enhancement if needed.

---

## Layer 1: Kernel & Bootstrap Infrastructure ❌ PENDING

**Goal**: Stand up the minimal hard-coded kernel that brings the 4 core services online

**Dependencies**: Layer 0 (Statechart Engine)  
**Reference**: `docs/arch-v1.md` Section 12, `phase1-kernel-bootstrap-implementation-plan.md`

### Components to Implement

| Component | File | Purpose | Lines (est.) |
|-----------|------|---------|--------------|
| Kernel | `cmd/maelstrom/kernel.go` | Bare binary bootstrap container | ~50 |
| ChartRegistry | `pkg/registry/registry.go` | YAML loading, hydration, file watching | ~300 |
| Bootstrap Chart | `embedded/bootstrap.yaml` | Hard-coded sequential orchestrator | ~100 |

### Bootstrap Sequence (arch-v1.md 12.1)

1. **Kernel starts** (bare binary, no Charts yet)
   - Loads minimal config (env + appVars)
   - Instantiates pure Statechart Library
   - Registers 4 bootstrap actions: `sysInit`, `securityBootstrap`, `communicationBootstrap`, `observabilityBootstrap`

2. **Kernel spawns Bootstrap Chart** (single hard-coded YAML)
   - Sequential compound state with 4 regions:
     - Region 1: Security & Boundary Service
     - Region 2: Communication Layer
     - Region 3: Observability + Persistence
     - Region 4: Lifecycle + Tools Registry
   - Kernel manually starts this one Chart (only time it touches Library directly)

3. **Bootstrap Chart runs** (inside Library)
   - Each region's entry action spawns a core service
   - Once all ready, emits `kernel_ready` event

4. **Full Maelstrom activation**
   - Kernel registers remaining actions/guards (LLM bridge, ContextMap, Orchestrator, etc.)
   - Loads YAML-defined services via ChartRegistry
   - All future Charts spawned via `sys:lifecycle`

5. **Steady state**
   - Kernel goes dormant (only listens for shutdown signals)
   - Everything is 100% Charts + Mail + Events

### Hard-Coded vs YAML-Defined (arch-v1.md 12.2)

| Category | How Defined | Loaded By | Why |
|----------|-------------|-----------|-----|
| **Core / Foundational** (Security, Communication, Observability, minimal Lifecycle/Registry, Bootstrap orchestrator) | Hard-coded Go structs (or embedded YAML) | Kernel at startup | Zero risk of missing files, deterministic bootstrap |
| **Everything else** (Gateway, Admin, Memory, Heartbeat, all Agents) | Normal YAML on disk | ChartRegistry (hot-reloadable) | Full flexibility for users |

---

## Layer 2: Core Platform Services ❌ PENDING

**Goal**: Implement the 4 hard-coded services that Bootstrap Chart brings online

**Dependencies**: Layer 1 (Bootstrap Chart spawns them)  
**Reference**: `docs/arch-v1.md` Section 7.1

### Services to Implement

| Service | File | Key Behavior | Mail Interface |
|---------|------|--------------|----------------|
| sys:security | `pkg/services/security/service.go` | Boundary enforcement, taint validation | `handleMail(mail) → outcome` |
| sys:communication | `pkg/services/communication/service.go` | Mail pub/sub backbone, topic routing | `handleMail(mail) → outcome` |
| sys:observability | `pkg/services/observability/service.go` | Traces, metrics, dead-letter queue | `emit(trace) / query(filters)` |
| sys:lifecycle | `pkg/services/lifecycle/service.go` | Spawn/stop charts post-bootstrap | `spawn(def) / stop(id)` |

### Critical Design Decision

These are **Charts** but hard-coded (not YAML). They use the Library API internally but have well-known IDs for Mail addressing (`sys:security`, `sys:communication`, etc.).

**Security Enforced After**: Region 1 comes online. Before that, no boundary checks. After, ALL communication goes through Mail with boundary enforcement.

---

## Layer 3: Communication & Mail System ❌ PENDING

**Goal**: Cross-boundary async messaging (foundation for everything above)

**Dependencies**: Layer 2 (sys:communication must be online)  
**Reference**: `docs/arch-v1.md` Section 9

### Components to Implement

| Component | File | Description |
|-----------|------|-------------|
| Mail types | `pkg/mail/types.go` | Addressing: `agent:<id>`, `topic:<name>`, `sys:<service>` |
| Publisher | `pkg/mail/publisher.go` | Fire-and-forget publishing |
| Subscriber | `pkg/mail/subscriber.go` | Topic subscriptions |
| Router | `pkg/mail/router.go` | Routes mail to targets |

### Mail Types (arch-v1.md 9.2)

- `user` | `assistant` | `tool_result` | `tool_call` | `mail_received`
- `heartbeat` | `error` | `human_feedback` | `partial_assistant` | `subagent_done` | `taint_violation`

### Why This Layer Before Security?

Security (sys:security) **USES** the Mail system to report violations. Mail is the transport; Security is the policy enforcement. Bootstrap sequence brings Security online first, but it needs Mail to communicate.

---

## Layer 4: Security & Boundaries + DataSources ❌ PENDING

**Goal**: Tainting, DLP, boundary enforcement, file storage with xattr

**Dependencies**: Layer 3 (Mail for violation reports), Layer 2 (sys:security online)  
**Reference**: `docs/arch-v1.md` Section 4, 11

### Components to Implement

| Component | File | Behavior |
|-----------|------|----------|
| BoundaryService | `pkg/security/boundary.go` | Inner/DMZ/outer enforcement |
| TaintEngine | `pkg/security/taint.go` | Proactive + reactive tainting |
| ContextMap | `pkg/security/contextmap.go` | Assembly with taint filtering |
| DataSource | `pkg/datasource/datasource.go` | File/object storage with xattr tainting |
| TaintRules | `pkg/security/rules.go` | Redaction, sanitization policies |

### Key Guarantee (arch-v1.md Section 4)

"Never allow outer/DMZ ContextMap to contain untagged inner data"

**Taint Attachment Points**:

1. At Message creation (Gateway → user, Orchestrator → tool_result, MemoryService → injected block, Heartbeat, etc.)
2. At ContextBlock assembly (Security calls `applyBlockTaints`)
3. On any data leaving the runtime (Mail, stream chunks, snapshot, sub-agent result) — uses `allowedOnExit`

---

## Layer 5: Hot-Reloadable Platform Services (YAML) ❌ PENDING

**Goal**: Load remaining sys:* services from YAML via ChartRegistry

**Dependencies**: Layer 4 (Security enforces boundaries on these)  
**Reference**: `docs/arch-v1.md` Section 7.2

### Services to Implement (all YAML-defined, loaded post-kernel_ready)

| Service | Key Capability | Depends On |
|---------|---------------|------------|
| sys:gateway | HTTP/SSE/WS, 2FA, OpenAPI | sys:communication, sys:security |
| sys:admin | k9s-style terminal/web | sys:gateway (outer-only) |
| sys:persistence | Snapshots, event sourcing | sys:observability |
| sys:heartbeat | Scheduled wake-ups | sys:lifecycle |
| sys:memory | Vector/graph stores | sys:persistence |
| sys:tools | Tool registry | sys:lifecycle |
| sys:datasources | Storage plugins | sys:security, DataSources |
| sys:human-gateway | Human-in-the-loop chat | sys:gateway |

**Note**: These are Charts that happen to be system services. They use the same YAML schema as user agents.

---

## Layer 6: Tool Calls & Orchestration ❌ PENDING

**Goal**: Execute tools and sub-agents with configurable policies

**Dependencies**: Layer 5 (sys:tools, sys:lifecycle), Layer 4 (boundaries)  
**Reference**: `docs/arch-v1.md` Section 8

### Components to Implement

| Component | File | Description |
|-----------|------|-------------|
| ToolRegistry | `pkg/tools/registry.go` | Schema, boundary-aware resolution |
| Orchestrator | `pkg/orchestrator/orchestrator.go` | Reusable sub-chart for execution |
| ExecutionPolicies | `pkg/orchestrator/policies.go` | seq_failfast, seq_continue, par_continue |

### Orchestrator as Sub-Chart

Tool calls spawn Orchestrator charts with parameters:

- `maxRetries`, `isolation` (strict/process/container), `maxParallel`

### Execution Patterns (arch-v1.md 8.3)

1. **Sequential, No-Fail-On-Fail** (`seq_continue`): Each tool call adds output to databag (scoped to calling chart)
2. **Sequential, Fail-On-First-Fail** (`seq_failfast`): First failed tool call aborts all following calls
3. **Parallel, No-Fail-On-Fail** (`par_continue`): All execute concurrently, async completion

---

## Layer 7: Agent Charts & LLM Integration ❌ PENDING

**Goal**: Actual AI agents that reason and act

**Dependencies**: Layer 6 (Orchestrator for tools), Layer 5 (sys:memory for ContextMap)  
**Reference**: `docs/arch-v1.md` Section 6, 15

### Components to Implement

| Component | File | Description |
|-----------|------|-------------|
| AgentChart | `pkg/agent/chart.go` | OODA loop as statechart |
| LLMBridge | `pkg/agent/llm.go` | ContextMap → LLM → tool calls |
| ContextMap | `pkg/agent/contextmap.go` | Assembles message history + memory |
| Extensions | `pkg/agent/extensions.go` | Tool use, streaming, sub-agents |

### OODA Loop (arch-v1.md 15)

Orient → Decide → Act → Observe as statechart states

**Agent Context Map**: An ordered list of ContextBlocks that defines the exact prompt assembled before any LLM inference. Built declaratively in YAML.

---

## Layer 8: Streaming & Gateway ❌ PENDING

**Goal**: External HTTP/SSE/WS interfaces

**Dependencies**: Layer 7 (Agents to expose), Layer 5 (sys:gateway)  
**Reference**: `docs/arch-v1.md` Section 10

### Components to Implement

| Component | File | Description |
|-----------|------|-------------|
| HTTPGateway | `pkg/gateway/http.go` | OpenAPI generation, routing |
| SSEAdapter | `pkg/gateway/sse.go` | Server-sent events for streaming |
| WSAdapter | `pkg/gateway/ws.go` | WebSocket support |
| 2FAHandler | `pkg/gateway/2fa.go` | Outer-only admin enforcement |

### Gateway Layer (arch-v1.md 10.1)

Channel Adapters (YAML-declared, hot-loaded):

- `webhook` - HTTP POST endpoints
- `websocket` - Full bidirectional
- `sse` - Server-Sent Events (firewall-friendly)
- `pubsub` - Message queue integration
- `smtp` - Email
- `slack`, `whatsapp`, `telegram` - Messaging platforms

---

## Implementation Phases

### Phase 1: Foundation (Layers 1-2)

**Milestone**: `kernel_ready` event fires. System can spawn charts.

1. Kernel bootstrap sequence
2. ChartRegistry YAML loading
3. Bootstrap Chart (hard-coded)
4. 4 core services (sys:security, sys:communication, sys:observability, sys:lifecycle)

**Reference**: `phase1-kernel-bootstrap-implementation-plan.md`

### Phase 2: Communication Backbone (Layer 3)

**Milestone**: Charts can send/receive mail.

1. Mail types and addressing
2. Pub/sub backbone
3. Router

### Phase 3: Security (Layer 4)

**Milestone**: Security enforced on all mail and data access.

1. Boundary service (inner/DMZ/outer)
2. Taint engine
3. DataSources with xattr
4. ContextMap filtering

### Phase 4: Remaining Platform Services (Layer 5)

**Milestone**: Full platform operational.

1. YAML-defined services loadable
2. sys:gateway, sys:admin, sys:persistence, sys:heartbeat
3. sys:memory, sys:tools, sys:datasources, sys:human-gateway

### Phase 5: Execution (Layer 6)

**Milestone**: Can execute tools and sub-agents.

1. Tool registry
2. Orchestrator sub-chart
3. Execution policies

### Phase 6: Agents (Layer 7)

**Milestone**: Agents can reason and use tools.

1. Agent Charts (OODA loop)
2. LLM bridge
3. ContextMap assembly

### Phase 7: External Interface (Layer 8)

**Milestone**: External clients can interact with agents.

1. HTTP gateway
2. SSE/WS streaming
3. 2FA for admin

---

## Target File Structure

```
maelstrom/
├── cmd/maelstrom/
│   ├── kernel.go              # ~50 lines, bootstrap sequence
│   ├── main.go                # Entry point
│   └── kernel_test.go         # Kernel tests
│
├── pkg/registry/
│   ├── registry.go            # ChartRegistry orchestration (~300 lines)
│   ├── hydrator.go            # ChartHydrator (~150 lines)
│   ├── types.go               # Registry types and interfaces
│   └── registry_test.go       # Registry tests
│
├── pkg/services/
│   ├── security/
│   │   ├── service.go         # SecurityService
│   │   └── service_test.go
│   ├── communication/
│   │   ├── service.go         # CommunicationService
│   │   ├── publisher.go       # Mail publisher
│   │   ├── subscriber.go      # Mail subscriber
│   │   └── service_test.go
│   ├── observability/
│   │   ├── service.go         # ObservabilityService
│   │   ├── trace.go           # Trace types
│   │   └── service_test.go
│   └── lifecycle/
│       ├── service.go         # LifecycleService
│       └── service_test.go
│
├── pkg/mail/
│   ├── types.go               # Mail types and addressing
│   ├── publisher.go           # Fire-and-forget publishing
│   ├── subscriber.go          # Topic subscriptions
│   └── router.go              # Routes mail to targets
│
├── pkg/security/
│   ├── boundary.go            # Inner/DMZ/outer enforcement
│   ├── taint.go               # Proactive + reactive tainting
│   ├── contextmap.go          # Assembly with taint filtering
│   ├── rules.go               # Redaction, sanitization policies
│   └── service_test.go
│
├── pkg/datasource/
│   └── datasource.go          # File/object storage with xattr tainting
│
├── pkg/orchestrator/
│   ├── orchestrator.go        # Reusable sub-chart for execution
│   ├── policies.go            # seq_failfast, seq_continue, par_continue
│   └── orchestrator_test.go
│
├── pkg/tools/
│   └── registry.go            # Schema, boundary-aware resolution
│
├── pkg/agent/
│   ├── chart.go               # OODA loop as statechart
│   ├── llm.go                 # ContextMap → LLM → tool calls
│   ├── contextmap.go          # Assembles message history + memory
│   └── extensions.go          # Tool use, streaming, sub-agents
│
├── pkg/gateway/
│   ├── http.go                # OpenAPI generation, routing
│   ├── sse.go                 # Server-sent events for streaming
│   ├── ws.go                  # WebSocket support
│   └── 2fa.go                 # Outer-only admin enforcement
│
├── embedded/
│   └── bootstrap.yaml         # Hard-coded bootstrap chart
│
└── internal/testutil/
    └── mock_context.go        # Mock ApplicationContext for testing
```

---

## Key Design Decisions

### 1. Bootstrap Sequence (arch-v1.md 12.1)

- **Kernel** is ~50 lines and spawns exactly ONE chart directly (Bootstrap Chart)
- Everything after uses Library API through `sys:lifecycle`
- Bootstrap Chart is the ONLY chart kernel spawns directly
- Kernel goes dormant after `kernel_ready` event

### 2. Hot-Reload Protocol (arch-v1.md 12.3)

- **Load-on-next-start**: Replace in-flight chart with clean start using history mechanisms
- **Quiescence**: Wait for empty queue + no active regions + no inflight work
- **History**: `shallowHistory` (parent default), `deepHistory` (specific sub-state), `cleanStart`
- **Context Transform**: Go template for schema migration on version change

### 3. Parallel States (arch-v1.md Section 5)

- **Hierarchical routing**: Events flow up to parent EventRouter, down to children
- **No sideways communication**: Regions never talk directly; all via parent
- **Quiescence-based control**: Pause, resume, graceful exit, snapshot
- **Goroutine-per-region**: Isolated execution with bidirectional channels

### 4. Security by Design (arch-v1.md Section 4)

- **Boundary Model**: inner (full access), dmz (mediated), outer (untrusted ingress)
- **Data Tainting**: Every Message, ContextBlock, value carries immutable taint set
- **ContextMap Assembly**: `prepareContextForBoundary` called inside every `assembleContextMap`
- **Stream Sanitization**: Every outgoing chunk passes through DMZ sanitizers

---

## References

### Primary Architecture Spec

- `docs/arch-v1.md` - Complete technical architecture v1.3

### Planning Documents

- `docs/initial-exploration/phase1-kernel-bootstrap-implementation-plan.md` - Phase 1 detailed scope
- `docs/initial-exploration/maelstrom-full-stack-implementation-plan.md` - Full 8-layer roadmap
- `docs/initial-exploration/kernel-bootstrap-refinement.md` - Bootstrap design notes
- `docs/initial-exploration/hot-reload-and-bootstrap-design.md` - Hot-reload methodology

### Design Session Notes

- `docs/sessions/20260226-parallel-states-design.md` - Parallel states semantics
- `docs/sessions/20260226-statechart-library-v1.md` - Core library implementation notes
- `docs/sessions/20260227-parallel-states-test-coverage.md` - Test coverage completion
- `docs/sessions/20260225-arch-design-hot-reload-taint-model.md` - Taint model design

### Existing Implementation

- `pkg/statechart/` - Layer 0 (73 tests passing)

---

## Next Steps

Per CLAUDE.md workflow:

1. **Create feature branch**: `git checkout -b feat/phase1-kernel-bootstrap`
2. **Define tests**: One test per behavior from spec
3. **Write stubs**: Zero implementation, only signatures
4. **Run tests**: Confirm they fail (RED)
5. **Implement**: Minimal code to make tests pass (GREEN)
6. **Commit**: One commit per behavior with descriptive message
7. **Repeat** until all Phase 1 tests green
8. **Refactor** while keeping tests green

**Recommended First Sprint**: Kernel & Bootstrap Chart

- Stub kernel types/methods
- Test: `TestKernel_SpawnsBootstrapChart`
- Implement bootstrap sequence
- Test: `TestBootstrapChart_SequentialExecution`
- Commit: `feat: kernel bootstrap sequence`

---

## Open Questions

1. **Mail system scope**: Implement full Mail in Phase 1, or stub for Phase 2?
2. **Persistence**: CoreRegistry in-memory only for Phase 1, or add disk persistence?
3. **Bootstrap failure**: Should Bootstrap Chart panic on failure, or retry?
4. **Service deferment**: Any services to defer to Phase 2 (e.g., observability metrics)?
5. **Quiescence detection**: Pursue optional enhancement for parallel states, or accept conservative approach?

---

## Gaps Identified (Cross-Cutting)

### Critical Architectural Issues

| Gap | Description | Severity |
|-----|-------------|----------|
| Layer 7/8 drift | These planning docs significantly deviate from arch-v1.md specifications - need reconciliation | CRITICAL |
| Dependency depth | Says "max depth is 5" but Layer 8 depends on Layer 7, 5, 1, 0 - that's depth 8 | CRITICAL |
| Missing type definitions | `IsolatedView`, `OpenAPISpec` used but never defined | CRITICAL |

### API Consistency Issues

| Gap | Description | Severity |
|-----|-------------|----------|
| ToolRegistry | Defined in both Layer 5 and Layer 6 - which is source of truth? | HIGH |
| Mail types | Layer 3 mentions 11, arch-v1.md Section 3.3 defines 10 | HIGH |
| Agent interface | Layer 7 defines Go Agent, arch-v1.md defines AgentSpec YAML - inconsistent | HIGH |
| Gateway scope | Layer 5 IS sys:gateway, but Layer 8 redefines completely | HIGH |

### Error Handling Gaps

| Gap | Description | Severity |
|-----|-------------|----------|
| Guard errors | No spec for guard throwing errors | HIGH |
| Action errors | No spec for action throwing errors | HIGH |
| Context transform | Template failure only says "cleanStart fallback" - no logging | HIGH |

### Interface Completeness

| Gap | Description | Severity |
|-----|-------------|----------|
| ApplicationContext | Missing Delete, List methods | HIGH |
| StorageBackend | Missing DeleteSnapshot method | HIGH |
| VectorStore/GraphStore | Missing Close() for resource cleanup | MEDIUM |
| DataSource | Missing List() for discovery | MEDIUM |
