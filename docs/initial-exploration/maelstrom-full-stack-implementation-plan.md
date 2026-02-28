# Maelstrom Full Stack Implementation Plan

**Date**: 2026-02-28
**Status**: Planning Phase
**Based on**: arch-v1.md dependency tree (Section 18) and bootstrap sequence (Section 12)

---

## Executive Summary

This document maps the complete implementation path from the current state (parallel states complete) to a full Maelstrom system. It is organized by **architectural layers** (bottom-up) to maintain clean dependency chains.

**Current State**: Layer 0 (Statechart Engine) is complete with 73 tests passing.

**Target**: All layers through 8 (External Interface).

---

## Architectural Dependency Analysis

From arch-v1.md Section 18:

```
Maelstrom Runtime (MVP)
├── External Interfaces (Gateway + Streaming + OpenAPI)
│   └── Communication (Mail + Streams)
│       └── Security & Boundary + Isolation & Tainting + DataSources
├── Agent Layer
│   ├── Agent Charts (nodes + ContextMap + Orchestrator)
│   └── Agent Extensions (LLM bridge)
├── Platform Services (all Charts except minimal kernel)
│   ├── sys:security, sys:communication, sys:observability, sys:lifecycle (hard-coded)
│   ├── sys:gateway, sys:admin (2FA outer-only), sys:memory, sys:datasources
│   ├── sys:persistence, sys:heartbeat, sys:tools, sys:human-gateway
│   └── All user-defined agents
├── Statechart Engine (pure library) ✓ COMPLETE
└── ChartRegistry + Kernel (hard-coded bootstrap)
```

**Longest dependency chain**: 5 layers (acceptable for MVP)

**Example**: UI → Gateway → Communication → Security → Engine

---

## Layer 0: Statechart Engine ✓ COMPLETE

**Status**: Fully implemented, 73 tests passing

**Files**: `pkg/statechart/*.go` (22 files)

**Components**:
- Atomic, compound, parallel states with goroutine-per-region
- Event dispatch, transitions, guards, actions
- Entry/exit actions for all state types
- Snapshot/restore with region state preservation
- Hot-reload via ReplaceDefinition
- Transient sub-charts with lifecycle binding
- Trace callbacks for observability
- Dynamic reclassification (atomic ↔ compound/parallel)

**Known TODO**:
- Region quiescence detection (runtime.go:143) - conservative approach (returns false for parallel states) is safe and functional

**Deferred**:
- Sub-chart routing (`child:bar`) - partially implemented via transient sub-charts
- Nested parallel states - not required for MVP

---

## Layer 1: Kernel & Bootstrap Infrastructure

**Goal**: Stand up the minimal hard-coded kernel that can load the bootstrap chart

**Dependencies**: Layer 0

**New Components**:

| Component       | File                       | Purpose                                | Lines |
| --------------- | -------------------------- | -------------------------------------- | ----- |
| Kernel          | `cmd/maelstrom/kernel.go`  | Bare binary bootstrap                  | ~50   |
| ChartRegistry   | `pkg/registry/registry.go` | YAML loading, hydration, file watching | ~300  |
| Bootstrap Chart | `embedded/bootstrap.yaml`  | Hard-coded bootstrap orchestrator      | ~100  |

**Bootstrap Sequence Implementation** (arch-v1.md 12.1):

1. **Kernel starts** (bare binary, no Charts yet)
   - Loads minimal config (env + appVars)
   - Instantiates pure Statechart Library + Chart Definition Loader
   - Registers absolute minimum built-in actions/guards:
     - `sysInit`
     - `securityBootstrap`
     - `communicationBootstrap`
     - `observabilityBootstrap`

2. **Kernel loads & spawns Bootstrap Chart** (single special YAML file: `bootstrap.yaml`)
   - This Chart is compound with sequential regions:
     - Region 1: Security & Boundary Service (inner/DMZ/outer + tainting + namespaces)
     - Region 2: Communication Layer (mail pub/sub backbone)
     - Region 3: Observability + Persistence
     - Region 4: Lifecycle + Tools Registry
   - Kernel manually starts this one Chart (the only time it ever touches the Library directly)

3. **Bootstrap Chart runs** (now fully inside the Library)
   - Security comes online first → enforces boundaries from this point forward
   - Communication comes online → all subsequent services communicate via Mail
   - Observability, Persistence, Lifecycle spin up as sub-states
   - Once all are ready, Bootstrap Chart emits `kernel_ready` event

4. **Full Maelstrom activation**
   - Kernel registers all remaining Maelstrom actions/guards (LLM bridge, ContextMap, Orchestrator, taintPropagate, etc.)
   - Loads and spawns:
     - `sys:gateway` + Channel Adapters (including streaming SSE/WS + 2FA)
     - `sys:admin` (outer-only)
     - `sys:memory`, `sys:heartbeat`, `sys:human-gateway`
   - Any configured top-level Agent Charts (inner/DMZ/outer)
   - All future Charts (user agents, sub-agents, etc.) are spawned via `sys:lifecycle`

5. **Steady state**
   - Everything after step 4 is 100% Charts + Mail + Events
   - Kernel goes dormant (only listens for shutdown signals)

**Key Insight**: Bootstrap Chart is the ONLY chart the kernel ever spawns directly. Everything else goes through sys:lifecycle.

**Hard-Coded vs YAML-Defined** (arch-v1.md 12.2):

| Category                                                                                                                        | How Defined                                                   | Loaded By                           | Why                                                                                                               |
| ------------------------------------------------------------------------------------------------------------------------------- | ------------------------------------------------------------- | ----------------------------------- | ----------------------------------------------------------------------------------------------------------------- |
| **Core / Foundational** (Security & Boundary, Communication, Observability, minimal Lifecycle/Registry, Bootstrap orchestrator) | Hard-coded Go structs (or embedded YAML compiled into binary) | Kernel at startup (never from disk) | Zero risk of missing/corrupt files, deterministic bootstrap, cannot be accidentally hot-reloaded or tampered with |
| **Everything else** (Gateway, Admin, Memory, Heartbeat, Human-Gateway, all Agent charts, Orchestrator templates, etc.)          | Normal YAML on disk                                           | ChartRegistry (hot-reloadable)      | Full flexibility for users and higher-level services                                                              |

---

## Layer 2: Core Platform Services (Hard-Coded)

**Goal**: Implement the 4 hard-coded services that Bootstrap Chart brings online

**Dependencies**: Layer 1 (Bootstrap Chart spawns them)

**Services**:

| Service | File | Key Behavior | Mail Interface |
|---------|------|--------------|----------------|
| sys:security | `pkg/services/security/service.go` | Boundary enforcement, taint validation | `handleMail(mail) → outcome` |
| sys:communication | `pkg/services/communication/service.go` | Mail pub/sub backbone, topic routing | `handleMail(mail) → outcome` |
| sys:observability | `pkg/services/observability/service.go` | Traces, metrics, dead-letter queue | `emit(trace) / query(filters)` |
| sys:lifecycle | `pkg/services/lifecycle/service.go` | Spawn/stop charts post-bootstrap | `spawn(def) / stop(id)` |

**Critical Design Decision**: These are Charts but hard-coded (not YAML). They use the Library API internally but have well-known IDs for Mail addressing.

**Security Enforced After**: Region 1 comes online. Before that, no boundary checks. After, ALL communication goes through Mail with boundary enforcement.

---

## Layer 3: Communication & Mail System

**Goal**: Cross-boundary async messaging (foundation for everything above)

**Dependencies**: Layer 2 (sys:communication must be online)

**Components**:

| Component | File | Description |
|-----------|------|-------------|
| Mail | `pkg/mail/types.go` | Addressing: `agent:<id>`, `topic:<name>`, `sys:<service>` |
| Publisher | `pkg/mail/publisher.go` | Fire-and-forget publishing |
| Subscriber | `pkg/mail/subscriber.go` | Topic subscriptions |
| Router | `pkg/mail/router.go` | Routes mail to targets |

**Mail Types** (arch-v1.md 9.2):
- user | assistant | tool_result | tool_call | mail_received
- heartbeat | error | human_feedback | partial_assistant | subagent_done | taint_violation

**Why This Layer Before Security?**
Security (sys:security) USES the Mail system to report violations. Mail is the transport; Security is the policy enforcement. Bootstrap sequence brings Security online first, but it needs Mail to communicate.

---

## Layer 4: Security & Boundaries + DataSources

**Goal**: Tainting, DLP, boundary enforcement, file storage with xattr

**Dependencies**: Layer 3 (Mail for violation reports), Layer 2 (sys:security online)

**Components**:

| Component | File | Behavior |
|-----------|------|----------|
| BoundaryService | `pkg/security/boundary.go` | Inner/DMZ/outer enforcement |
| TaintEngine | `pkg/security/taint.go` | Proactive + reactive tainting |
| ContextMap | `pkg/security/contextmap.go` | Assembly with taint filtering |
| DataSource | `pkg/datasource/datasource.go` | File/object storage with xattr tainting |
| TaintRules | `pkg/security/rules.go` | Redaction, sanitization policies |

**Key Guarantee** (arch-v1.md 20): "Never allow outer/DMZ ContextMap to contain untagged inner data"

---

## Layer 5: Hot-Reloadable Platform Services (YAML)

**Goal**: Load remaining sys:* services from YAML via ChartRegistry

**Dependencies**: Layer 4 (Security enforces boundaries on these)

**Services** (all YAML-defined, loaded post-kernel_ready):

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

## Layer 6: Tool Calls & Orchestration

**Goal**: Execute tools and sub-agents with configurable policies

**Dependencies**: Layer 5 (sys:tools, sys:lifecycle), Layer 4 (boundaries)

**Components**:

| Component | File | Description |
|-----------|------|-------------|
| ToolRegistry | `pkg/tools/registry.go` | Schema, boundary-aware resolution |
| Orchestrator | `pkg/orchestrator/orchestrator.go` | Reusable sub-chart for execution |
| ExecutionPolicies | `pkg/orchestrator/policies.go` | seq_failfast, seq_continue, par_continue |

**Orchestrator as Sub-Chart**: Tool calls spawn Orchestrator charts with parameters:
- `maxRetries`, `isolation` (strict/process/container), `maxParallel`

---

## Layer 7: Agent Charts & LLM Integration

**Goal**: Actual AI agents that reason and act

**Dependencies**: Layer 6 (Orchestrator for tools), Layer 5 (sys:memory for ContextMap)

**Components**:

| Component | File | Description |
|-----------|------|-------------|
| AgentChart | `pkg/agent/chart.go` | OODA loop as statechart |
| LLMBridge | `pkg/agent/llm.go` | ContextMap → LLM → tool calls |
| ContextMap | `pkg/agent/contextmap.go` | Assembles message history + memory |
| Extensions | `pkg/agent/extensions.go` | Tool use, streaming, sub-agents |

**OODA Loop** (arch-v1.md 15): Orient → Decide → Act → Observe as statechart states

---

## Layer 8: Streaming & Gateway

**Goal**: External HTTP/SSE/WS interfaces

**Dependencies**: Layer 7 (Agents to expose), Layer 5 (sys:gateway)

**Components**:

| Component | File | Description |
|-----------|------|-------------|
| HTTPGateway | `pkg/gateway/http.go` | OpenAPI generation, routing |
| SSEAdapter | `pkg/gateway/sse.go` | Server-sent events for streaming |
| WSAdapter | `pkg/gateway/ws.go` | WebSocket support |
| 2FAHandler | `pkg/gateway/2fa.go` | Outer-only admin enforcement |

---

## Implementation Order Recommendation

### Phase 1: Foundation (Layers 1-2)
1. Kernel bootstrap sequence
2. ChartRegistry YAML loading
3. Bootstrap Chart (hard-coded)
4. 4 core services (sys:security, sys:communication, sys:observability, sys:lifecycle)

**Milestone**: `kernel_ready` event fires. System can spawn charts.

### Phase 2: Communication Backbone (Layer 3)
1. Mail types and addressing
2. Pub/sub backbone
3. Router

**Milestone**: Charts can send/receive mail.

### Phase 3: Security (Layer 4)
1. Boundary service (inner/DMZ/outer)
2. Taint engine
3. DataSources with xattr
4. ContextMap filtering

**Milestone**: Security enforced on all mail and data access.

### Phase 4: Remaining Platform Services (Layer 5)
1. YAML-defined services loadable
2. sys:gateway, sys:admin, sys:persistence, sys:heartbeat
3. sys:memory, sys:tools, sys:datasources, sys:human-gateway

**Milestone**: Full platform operational.

### Phase 5: Execution (Layer 6)
1. Tool registry
2. Orchestrator sub-chart
3. Execution policies

**Milestone**: Can execute tools and sub-agents.

### Phase 6: Agents (Layer 7)
1. Agent Charts (OODA loop)
2. LLM bridge
3. ContextMap assembly

**Milestone**: Agents can reason and use tools.

### Phase 7: External Interface (Layer 8)
1. HTTP gateway
2. SSE/WS streaming
3. 2FA for admin

**Milestone**: External clients can interact with agents.

---

## Questions for User

1. **Scope**: Implement all layers or stop at a specific milestone?
2. **Priority**: Any specific layer that blocks other work?
3. **Testing**: Acceptance criteria for each layer?
4. **Deferred**: Any features to intentionally skip (e.g., streaming, 2FA)?
