# Maelstrom Architecture v1.0

A zero-human, statechart-native agentic runtime. This document specifies the complete technical architecture including design decisions, semantics, behaviors, and API contracts.

---

## 1. Executive Summary

Maelstrom is a compact, secure, enterprise-grade framework for running autonomous AI agents without human intervention. The architecture is built on a single unifying abstraction: **statecharts**.

Every behavior in Maelstrom—reasoning, orchestration, lifecycle management, observability—is expressed as a statechart. Agents are statecharts. System services are statecharts. Tool execution is orchestrated through statecharts.

### Key Design Principles

1. **Single Source of Truth**: The statechart engine is the sole runtime for all control flow.
2. **Security by Design**: Inner/DMZ/outer boundaries with data tainting prevent prompt injection and data exfiltration.
3. **Declarative Everything**: All behavior is defined in YAML, hot-reloadable, versioned, and auditable.
4. **Zero-Human Operation**: Agents self-orchestrate, self-heal, and self-improve through a meta-agent feedback loop.
5. **Compact but Enterprise**: <10 functional blocks, clean APIs, comprehensive observability.

---

## 2. Core Abstractions

### 2.1 Chart

The atomic unit of everything in Maelstrom. A Chart is a state machine defined in YAML, hot-loaded, hydrated (with environment variables, application variables, and Go templates), and instantiated at runtime.

Every running Chart (ChartRuntime) has:

- **Active configuration**: current states, regions
- **Runtime context** (read-only): chart ID, start time, metrics, parent ID if attached
- **Application context** (read-write, permissioned): user-scoped data bag with namespace isolation
- **Event queue**: synchronous inside atomic/compound regions; asynchronous across parallel regions and external sources

### 2.2 Node (Unified State Primitive)

The statechart library works exclusively with **Node** objects. A Node's runtime behavior is derived automatically from its structure:

- **0 children** → atomic (leaf behavior)
- **1 child** → compound (exactly one active sub-state)
- **≥2 children** (with explicit region names) → parallel (one active sub-state per named region)

This unification is an implementation detail inside the pure statechart library. The YAML author-facing model remains simple: "states are just nodes with children."

Dynamic reclassification (atomic → compound/parallel via entry action) is atomic within a single synchronous evaluation step. Subsequent events see the new structure.

**Guardrails**: A chart-level `stabilityPolicy` limits reconfiguration depth or forces collapse after N reconfigurations to prevent runaway meta-behavior.

### 2.3 Events

Events drive all internal transitions. Events carry: `type`, `payload`, `correlationId`, `source`.

**Evaluation Order Guarantees**:
- **Synchronous ordering**: Guaranteed only within a compound/parallel boundary. Events are processed deterministically.
- **Eventually consistent**: Across parallel regions or external sources. Best-effort, at-least-once delivery with deduplication via `correlationId`.

**Event Addressing** (local to ChartRuntime only):
- `targetPath` optional, used for:
  - Parallel regions: `region:foo`
  - Sub-charts: `child:bar` or `child:bar/region:baz`
  - Self: omitted or "."
- No global addressing, no topics, no inter-Chart routing — that is 100% the Mail system's job.

### 2.4 Actions & Guards

**Actions** execute on state entry/exit or transitions. Signature:
```
func(runtimeCtx RuntimeContext, appCtx ApplicationContext, event Event)
```

- **RuntimeContext** (read-only): chart ID, active states, metrics, start time
- **ApplicationContext** (read-write): user-scoped data bag, namespaced, taint-aware
- **Initiating event**: the event that triggered the action

**Guards** determine if a transition may fire. Signature:
```
func(appCtx ApplicationContext, event Event) bool
```

- **ApplicationContext** (read-only): boundary-filtered view
- **Initiating event**

Both actions and guards may be chart-defined or system-provided (e.g., `"sendMail"`, `"logMetric"`).

### 2.5 System Service

A System Service is a specialized Chart that handles application-level concerns: chart lifecycle, timers/crons, heartbeat scheduling, resource quotas, observability, dead-letter queue. System Services run at bootstrap and are addressable by well-known IDs (e.g., `sys:heartbeat`, `sys:persistence`).

---

## 3. Data Models

### 3.1 ChartDefinition

Immutable, versioned YAML source + hydrated runtime graph.

```yaml
apiVersion: maelstrom.dev/v1
kind: Chart
metadata:
  name: string           # unique identifier
  version: "x.y.z"       # semver
  immutable: boolean     # if true, hot-reload rejected
  boundary: inner | dmz | outer
spec:
  stabilityPolicy:
    maxReconfigDepth: int
    collapseAfter: int
  taintPolicy:
    enforcement: strict | redact | audit
    allowedOnExit: [string]    # taint categories allowed to leave
    redactRules:
      - taint: string
        replacement: string
  persistence:
    enabled: boolean
    snapshotOn:
      - stateEntry: string
      - every: int messages
      - cron: string
  nodes: { ... }           # Node tree (see below)
  contextMap: [ ... ]     # Agent-only: ContextBlocks
  orchestrator: { ... }   # Agent-only: tool execution policy
  tools: [ ... ]           # Tool bindings
  expose: { ... }          # HTTP/OpenAPI exposure
```

### 3.2 ChartRuntime

Live instance of a ChartDefinition.

```go
type ChartRuntime struct {
    ID              string          // unique runtime ID
    DefinitionID    string          // pinned version
    ParentID        *string         // nil if top-level
    Boundary        BoundaryType    // inner | dmz | outer
    ActiveStates    []StatePath     // current configuration
    EventQueue      []Event         // pending events
    RuntimeContext  RuntimeContext  // read-only metrics
    ApplicationContext ApplicationContext // R/W data bag
    SessionPointer  string          // reference to session store
    TaintMap        map[string][]string // object → taints
}
```

### 3.3 Message

Immutable unit of session history.

```yaml
type: user | assistant | tool_result | tool_call | mail_received | heartbeat | error | human_feedback | partial_assistant | subagent_done | taint_violation
createdAt: timestamp
source: agent:<id> | sys:<service> | human | gateway | tool:<name>
target: agent:<id> | topic:<name> | sys:<service>
content: any
correlationId: uuid
metadata:
  tokens: int
  model: string
  cost: float
  boundary: inner | dmz | outer
  taints: [string]
  stream: boolean        # for partial_assistant
  isFinal: boolean       # last chunk marker
```

### 3.4 Session

Ordered, append-only list of Messages. Immutable history; summarization produces new Messages, never mutates old.

```yaml
session:
  id: uuid
  agentId: string        # owning agent
  messages: [Message]    # append-only
  summaryPointer: string # reference to summary Message
  createdAt: timestamp
  lastActivity: timestamp
```

### 3.5 ContextBlock

Declarative slot in a ContextMap defining how context is assembled for LLM calls.

```yaml
contextMap:
  - name: string              # identifier
    source: static | session | memoryService | toolRegistry | runtime
    content: string           # for static
    strategy: lastN | summarize | RAG | full  # for session/memory
    n: int                    # for lastN
    maxTokens: int            # hard cap
    eviction: summarize | truncate | error      # what to do on overflow
    compressor: string        # name of registered compressor
    priority: int             # assembly order
    streamCommit: boolean     # false = partials don't enter next prompt
    taintPolicy:              # per-block security
      redactMode: redact | dropBlock | audit
      redactRules:
        - taint: string
          replacement: string
      allowedForBoundary: [inner | dmz | outer]
```

### 3.6 AgentSpec

ChartDefinition specialized for LLM agents.

```yaml
apiVersion: maelstrom.dev/v1
kind: AgentSpec  # extends Chart
metadata:
  name: string
  version: string
  boundary: inner | dmz | outer
spec:
  # All Chart fields plus:
  llmConfig:
    modelFamily: string
    temperature: float
    maxTokens: int
    streaming: boolean
  contextMap: [ContextBlock]
  toolBindings:
    global: [ToolBinding]
    perState:
      stateName: [ToolBinding]
  orchestratorPolicy:
    default: string
    timeoutMs: int
  subAgentPolicy:
    maxDepth: int
    inheritContext: boolean
  stabilityPolicy:
    maxReconfigDepth: int
    collapseAfter: int
```

---

## 4. Security & Boundaries

### 4.1 Boundary Model

Every Chart declares its boundary in metadata:

- **inner**: Full read/write to sensitive application-context slices (secrets, internal DBs, production APIs). Prompt injection surface is zero because no untrusted Messages ever enter an inner Chart.
- **DMZ**: Mediated access. Tools/sub-agents are wrapped; outputs are sanitized (PII redaction, length limits, schema validation). ContextMap blocks marked with forbidden taints are stripped before any LLM call.
- **outer**: Untrusted ingress only (user chat, webhooks, mail from external). Can only talk to DMZ agents via mail; cannot directly read/write inner data.

### 4.2 Data Tainting

Every Message, ContextBlock, and ApplicationContext value carries an immutable taint set (e.g., `["PII", "SECRET", "INNER_ONLY", "USER_SUPPLIED"]`).

**Taint Attachment Points**:
1. At Message creation (Gateway → user, Orchestrator → tool_result, MemoryService → injected block, Heartbeat, etc.)
2. At ContextBlock assembly (Security calls `applyBlockTaints`)
3. On any data leaving the runtime (Mail, stream chunks, snapshot, sub-agent result) — uses `allowedOnExit`

**Taint Propagation**:
- Security Service propagates taints on copy/read/write (like DLP tracking)
- Inner → DMZ/outer: auto-strip or block on forbidden taints
- On-disk (snapshots, long-term memory): taints stored with data; Persistence refuses writes that would violate policy
- Runtime guard: any action/guard can query taints; violation emits `taint_violation` event to dead-letter

### 4.3 Boundary Enforcement

**Transitions across boundaries** are explicit and always go through the Mail system (never direct event passing). This is enforced by the `sys:security` System Service Chart at runtime.

**ContextMap Assembly Enforcement**:
- Security service calls `prepareContextForBoundary(runtimeId, boundary)` inside every `assembleContextMap`
- Any block containing forbidden taint for current boundary is redacted or dropped before LLM call
- Per-block `taintPolicy` overrides global settings

**Stream Sanitization**:
- Every outgoing chunk passes through DMZ sanitizers (PII redaction, length caps, schema validation, inner-data stripping)
- Sanitization is per-chunk (stateless) so latency stays <50ms
- No buffering of entire response unless chart explicitly requests it

---

## 5. Statechart Library vs Maelstrom App

### 5.1 The Locked Seam

This is the precise boundary between the reusable, pure statechart library and the Maelstrom application layer.

**Statechart Library (pure, reusable, zero domain knowledge)**:
- Node-based graph (children-derived atomic/compound/parallel semantics)
- Local event addressing (`runtimeId`, `targetPath?`, `type`, `payload`, `correlationId`)
- Action/Guard registry + execution (`func(runtimeCtx, appCtx, event)`)
- Parallel regions & transient sub-charts (spawn/join/abort)
- Control primitives: `start | pause | resume | stop | snapshot | restore | injectEvent | replaceDefinition`
- Observability trace callbacks only

**Maelstrom App (everything else — thin, registered on top)**:
- Kernel (hard-coded, minimal)
- ChartRegistry (file watch, hydration, versioning, persistence)
- All domain actions/guards (`assembleContextMap`, `dispatchLLM`, `publishMail`, `taintPropagate`, `securityCheck`, etc.)
- Mail system, Security & Isolation & Tainting, Gateway, `sys:*` services, Agent Extensions, etc.
- Hard-coded core services (Security, Communication, Observability, minimal Lifecycle)

The Library never calls Maelstrom code except through registered actions/guards or the two context bags. Maelstrom never touches Library internals. The boundary is airtight and reusable.

### 5.2 Pure Statechart Library API

```go
type Library interface {
    // Lifecycle
    Spawn(def ChartDefinition, initialAppCtx ApplicationContext) (RuntimeID, error)
    Control(id RuntimeID, cmd ControlCmd) error  // start|pause|resume|stop|snapshot|restore|injectEvent|replaceDefinition

    // Event dispatch (local addressing only)
    Dispatch(id RuntimeID, ev Event) error

    // Extension points
    RegisterAction(name string, fn ActionFn)
    RegisterGuard(name string, fn GuardFn)

    // Persistence
    Snapshot(id RuntimeID) (Snapshot, error)
    Restore(snap Snapshot) (RuntimeID, error)
}

type ControlCmd string
const (
    CmdStart ControlCmd = "start"
    CmdPause ControlCmd = "pause"
    CmdResume ControlCmd = "resume"
    CmdStop ControlCmd = "stop"
    CmdSnapshot ControlCmd = "snapshot"
    CmdInjectEvent ControlCmd = "injectEvent"
    CmdReplaceDefinition ControlCmd = "replaceDefinition"
)

type Event struct {
    Type          string
    Payload       any
    CorrelationID string
    Source        string
    TargetPath    string  // optional: "region:foo", "child:bar", etc.
}

type RuntimeContext struct {
    ChartID       string
    RuntimeID     string
    ParentID      *string
    ActiveStates  []string
    StartTime     time.Time
    Metrics       map[string]float64
}

type ApplicationContext interface {
    Get(key string, callerBoundary BoundaryType) (any, []string, error)  // value, taints, error
    Set(key string, value any, taints []string, callerBoundary BoundaryType) error
    Namespace() string
}

type ActionFn func(runtimeCtx RuntimeContext, appCtx ApplicationContext, event Event) error
type GuardFn func(appCtx ApplicationContext, event Event) bool
```

### 5.3 Library Guarantees (SHALL/SHALL NOT)

**SHALL**:
- Evaluate events synchronously inside compound/parallel boundaries
- Run parallel regions in isolated goroutines with bidirectional channels
- Fire trace callback on every transition, entry, exit, event dispatch, sub-chart spawn
- Support dynamic reclassification (atomic ↔ compound/parallel) atomically within one evaluation step
- Provide `replaceDefinition` hook for hot-reload (running instance decides acceptance)
- Support transient sub-charts with automatic lifecycle binding to parent

**SHALL NOT**:
- Know about Mail, LLM, taints, boundaries, files, or agents
- Perform any I/O (file, network, etc.)
- Implement persistence (only serialize/deserialize snapshots)
- Have opinions on security policy

---

## 6. Agent Charts & Reasoning

### 6.1 Agent Chart Specification

An Agent is a Chart whose states represent **modes of reasoning** (Observe, Orient, Decide, Act, Reflect, Escalate, etc.). The chart author encodes any desired loop (OODA, ReAct, Plan-Execute-Verify, etc.) as compound/parallel regions.

Exit from the top-level agent state is gated by:
- Success criteria (guard condition)
- Max-iterations counter
- External stop event

### 6.2 Agent Context Map

An ordered list of ContextBlocks that defines the exact prompt assembled before any LLM inference. Built declaratively in YAML so the same chart can be tuned for short-context vs long-context models or cost vs quality trade-offs.

Assembly is deterministic and happens inside a canonical `PrepareContext` action.

### 6.3 Canonical LLM Inference Loop

Every Agent Chart has an implicit or explicit `LLMReason` compound state (or transition target). When entered:

1. **ContextMap → assembled prompt** (with token budgeting)
2. **LLM call** (streaming or full)
3. **Response routed to configured Orchestrator** sub-chart
4. **Orchestrator outcome** (success, partial, failure) produces new Messages and events that drive the next transition

This makes the LLM a first-class "external action" rather than something sprinkled across states.

### 6.4 Sub-Agents

Sub-agents are invoked exactly like tools (they appear in the tool registry when declared).

**Attached**: Spawns a child ChartRuntime whose lifecycle is bound to the parent (auto-terminated on parent exit). Parent can share a scoped slice of its application context. Result returned via a `subAgentDone` event + optional Messages injected into parent session.

**Detached**: Spawns a new top-level ChartRuntime managed by `sys:lifecycle`. Fire-and-forget or fire-and-await via mail `correlationId`.

---

## 7. Platform Services

All Platform Services are Charts themselves, addressable by well-known IDs. They run on the same Statechart Engine as user agents.

### 7.1 Core Services (Hard-Coded)

These are compiled into the binary and started by the Kernel during bootstrap:

- **sys:security** — Boundary enforcement, taint validation, sanitization
- **sys:communication** — Mail pub/sub backbone
- **sys:observability** — Traces, metrics, dead-letter queue
- **sys:lifecycle** (minimal) — Basic spawn/stop for bootstrap

### 7.2 Hot-Reloadable Services (YAML-Defined)

Loaded by ChartRegistry after `kernel_ready`:

- **sys:gateway** — Channel adapters (HTTP/SSE/WS/Email/Slack/etc.), OpenAPI generation, 2FA enforcement
- **sys:admin** — k9s-style terminal/web console for debugging (outer-only, 2FA-gated)
- **sys:persistence** — Snapshots, event sourcing, version migration
- **sys:heartbeat** — Scheduled agent wake-ups with HEARTBEAT.md injection
- **sys:memory** — Long-term memory (vector/graph stores) for ContextMap injection
- **sys:human-gateway** — Chat interface for human-in-the-loop with running agents
- **sys:tools** — Tool registry and resolution
- **sys:datasources** — Pluggable FS/object storage with tainting

### 7.3 Platform Service Contract

Every `sys:*` service exposes:

```yaml
handleMail(mail: Mail) → outcomeEvent
```

Specific additions per service:

- **sys:admin**: `adminCommand(cmd: list|control|queryTaints|inject, authToken)` (2FA-gated)
- **sys:persistence**: `snapshot(runtimeId) / restore(snapshotId)`
- **sys:observability**: `emit(trace) / query(traces, filters)`
- **sys:tools**: `resolve(name, callerBoundary) → ToolDescriptor | notFound`
- **sys:memory**: `query(vector, topK, boundaryFilter) → MessageSlice`
- **sys:datasources**: `tagOnWrite(path, taints) / getTaints(path) → []string`

---

## 8. Tool Calls & Orchestration

### 8.1 Tool & Action Orchestration (Unified)

Tool calls and sub-agent calls are both "actions" that may require orchestration. Instead of four hardcoded variants, we have a single configurable **Orchestrator** sub-chart (itself a reusable Chart) parameterized by:

- **Execution mode**: `sequential | parallel`
- **Failure policy**: `fail-fast | continue-on-error`
- **Completion signaling**: `immediate` (block until done) | `async` (DONE event later)

The LLM response is always handed to the orchestrator named in the current state's binding (or chart default). Orchestrator instances run as transient parallel regions inside the `LLMReason` compound state.

### 8.2 Orchestrator YAML Shape

```yaml
orchestrator:
  default: seq_failfast          # or seq_continue, par_continue
  timeoutMs: 30000
  maxParallel: 8                  # only for par_*
  policies:
    seq_failfast:
      maxRetries: 2
      isolation: strict             # each tool in own namespace
    seq_continue:
      maxRetries: 1
      isolation: process
    par_continue:
      maxParallel: 8
      isolation: strict
```

### 8.3 Execution Patterns

**Sequential, No-Fail-On-Fail** (`seq_continue`):
```
tool call → awaiting → tool call → awaiting → ...
```
- Each tool call adds its output to the databag (scoped to calling chart)
- Includes failed tool calls
- When results re-inserted into session context, includes failed tool-call output

**Sequential, Fail-On-First-Fail** (`seq_failfast`):
- Short-circuiting AND application on tool-calls
- First failed tool call aborts all following calls
- All tool output injected back into session context (scoped to session)

**Parallel, No-Fail-On-Fail** (`par_continue`):
- Spawns parallel states for each tool-call
- All execute concurrently
- Session context insertion as above
- Caller continues immediately (async completion)

**Parallel, Fail-On-First-Fail** (`par_failfast`):
- Dropped from MVP (too racy, not worth complexity)
- Fail-fast is almost always sequential in practice

### 8.4 Tool Registry

Tools are registered with boundary-aware schemas:

```yaml
tools:
  - name: webSearch
    boundary: dmz
    schema:
      input: { ... }
      output: { ... }
    isolation: container         # or process, sandbox, strict
    taintOutput: ["TOOL_OUTPUT"]
  - name: innerDbQuery
    boundary: inner
    isolation: strict
    taintOutput: ["TOOL_OUTPUT", "INNER_ONLY"]
  - name: subAgentResearch
    type: attachedSubAgent
    chartRef: research-mini-v1
    isolation: strict
```

**Tool Resolution**:
```go
resolveTool(name string, callerBoundary BoundaryType) → ToolDescriptor | notFound
```

Returns sanitized schema filtered by caller's boundary. Inner tools are invisible to outer callers.

---

## 9. Inter-Agent Communication

### 9.1 Events vs Mail

**Events**: Drive internal transitions within a Chart. Synchronous ordering guaranteed only within compound/parallel boundaries. "Eventually consistent" across parallel regions.

**Mail**: The only cross-boundary primitive. Async, fire-and-forget or request-reply via `correlationId`.

Statecharts are driven by Events. We **do not** use events to facilitate inter-agent communication. We use a different method altogether: a pub-sub mail system built on top of a standard messaging library.

### 9.2 Mail System

Only top-level agents (those expected to have long life-cycles) have inboxes. All agents can send mail to any inbox.

**Mail Addressing**:
- `agent:<id>` - specific agent inbox
- `topic:<name>` - pub-sub topic (multiple subscribers)
- `sys:<service>` - system service inbox

**Mail Format**:
```yaml
id: uuid
correlationId: uuid
type: user | assistant | tool_result | tool_call | mail_received | heartbeat | error | human_feedback | partial_assistant | subagent_done | taint_violation
createdAt: timestamp
source: agent:<id> | sys:<service> | human | gateway | tool:<name>
target: agent:<id> | topic:<name> | sys:<service>
content: any
correlationId: uuid
metadata:
  tokens: int
  model: string
  cost: float
  boundary: inner | dmz | outer
  taints: [string]
  stream: boolean
  isFinal: boolean
```

**Delivery Guarantees**:
- At-least-once delivery
- Deduplication via `correlationId`
- Dead-letter queue on permanent failure (managed by `sys:observability`)

### 9.3 Communication Layer API

```go
type Communication interface {
    Publish(mail Mail) (ack Ack, err error)
    Subscribe(address string) (mailStream chan Mail, err error)
    Unsubscribe(address string, stream chan Mail) error

    // Streaming support
    UpgradeToStream(sessionId string, lastEventId *string) (chunkChannel chan StreamChunk, err error)
}

type Ack struct {
    CorrelationID string
    DeliveredAt   time.Time
}

type StreamChunk struct {
    Data      string
    Sequence  int
    IsFinal   bool
    Taints    []string
}
```

---

## 10. Streaming & Gateway

### 10.1 Gateway Layer

The Gateway System Service (`sys:gateway`) normalizes all external I/O through pluggable Channel Adapters.

**Channel Adapters** (YAML-declared, hot-loaded):
- `webhook` - HTTP POST endpoints
- `websocket` - Full bidirectional
- `sse` - Server-Sent Events (firewall-friendly)
- `pubsub` - Message queue integration
- `smtp` - Email
- `slack`, `whatsapp`, `telegram` - Messaging platforms
- `internal_grpc` - Internal service mesh

Each adapter normalizes inbound traffic to a standard `mail_received` Message and normalizes outbound `mail_send` to the channel's format.

### 10.2 Streaming Path (End-to-End)

```
User types → Gateway (SSE connect) → Communication.publish(mail: type=user, taints=["USER_SUPPLIED"], boundary=outer)
Security.validate → DMZ Agent inbox
Engine.dispatchEvent → enters LLMReason state
AgentExtensions.assembleContextMap (pulls session + Memory blocks; taints propagated) → LLM call
LLM returns tool calls → Orchestrator (isolated sub-chart) → Tool Registry.resolve → executeIsolated (new namespace)
Tool result → taintPropagate → new tool_result Mail (taints=["TOOL_OUTPUT"]) injected to session
Engine emits partial_assistant Messages (stream: true) → Communication → Gateway → Security.stripForbiddenTaints → SSE chunks to user
```

**Stream Configuration**:
```yaml
contextMap:
  - name: conversation
    source: session
    strategy: lastN
    n: 30
    streamCommit: false  # false = partials never enter next prompt
```

**Stream Chunk Format**:
```yaml
{
  chunk: string,
  sequence: int,
  isFinal: bool,
  taints: [string]  // Security strips before emission
}
```

### 10.3 HTTP(S) Exposure & OpenAPI

Any Chart can declare HTTP endpoints:

```yaml
expose:
  http:
    path: /api/v1/agents/{id}/
    events:
      - trigger: user_query
        method: POST
        requestSchema: UserQuery
        responseSchema: AgentResponse
        stream: true
      - trigger: status
        method: GET
```

`sys:gateway` automatically registers these as OpenAPI 3.1 endpoints with auth middleware. The event surface of the chart becomes the API surface. Inner-boundary Charts are never directly exposed; only DMZ/outer ones.

### 10.4 Human Chat with Running Agents

System Service: `sys:human-gateway`

For any long-running top-level Agent, open an HTTPS chat session at `/chat/{agentId}`.

**Features**:
- Session receives read-only snapshot of Agent's current ContextMap + last N Messages (sanitized by boundary rules)
- Any message sent becomes `mail_received` (type: `human_feedback` or `user`) delivered to Agent's inbox
- Agent replies via normal mail → rendered back in chat UI
- Optional "action item" shorthand: `@pause`, `@inject-memory X`, etc. become special Mail messages

---

## 11. DataSources & Tainting

### 11.1 DataSource Abstraction

Pluggable abstraction for filesystem/object storage with automatic tainting. Owned by `sys:datasources`.

**YAML Declaration**:
```yaml
dataSources:
  - name: agentWorkspace
    type: localDisk
    path: "/agents/{{agentId}}/workspace"
    mountMode: readWrite
    taintMode: inheritFromXattr   # or alwaysTaintAs=INNER_ONLY, or none
    defaultTaints: ["WORKSPACE"]   # baseline taint for anything written here
    xattrNamespace: user.maelstrom   # where we store taints on files

  - name: sharedInnerRepo
    type: localDisk
    path: "/inner/repo"
    mountMode: readOnly
    taintMode: enforceInnerOnly
    allowedForBoundary: ["inner"]

  - name: s3Bucket
    type: objectStorage
    bucket: my-agent-bucket
    taintMode: inheritFromObjectTags
```

### 11.2 Behavior

**On Write** (any tool that writes a file):
1. Write action calls `sys:datasources.tagOnWrite(path, incomingTaints)`
2. For localDisk: sets xattr `user.maelstrom.taints=PII,TOOL_OUTPUT` (or JSON-encoded)
3. For object stores: sets object tags/metadata
4. For workspaces: updates sidecar manifest for fast lookup

**On Read** (any tool that reads):
1. Read tool calls `sys:datasources.getTaints(path)`
2. Security service immediately attaches those taints to resulting `tool_result` or injected ContextBlock Message
3. Existing `prepareContextForBoundary` + per-block `taintPolicy` redacts/drops/audits as configured

**Isolation Hook**:
Every agent's persistent workspace is mounted via isolated executor (container/namespace) with exact DataSource policy. No direct `open()` syscall allowed — all FS ops go through mediated tool layer.

### 11.3 DataSource Interface

```go
type DataSource interface {
    TagOnWrite(path string, taints []string) error
    GetTaints(path string) ([]string, error)
    ValidateAccess(boundary string) error
}
```

Built-ins: `localDisk` (xattr + fallback sidecar), `s3`, `gcs`, `inMemoryWorkspace`.

Roll-your-own: Register new implementation by name — works for NFS, Git repos, vector DBs, etc.

---

## 12. Bootstrap Sequence

### 12.1 The Kernel

Because Platform Services are themselves Charts, we need a tiny hard-coded Kernel (≈50 lines) that is the only non-Chart code in the entire system.

**Bootstrap Sequence**:

1. **Kernel starts** (bare binary, no Charts yet)
   - Loads minimal config (env + appVars)
   - Instantiates pure Statechart Library + Chart Definition Loader
   - Registers absolute minimum built-in actions/guards needed for bootstrap:
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
   - Once all are ready, the Bootstrap Chart emits a `kernel_ready` event

4. **Full Maelstrom activation**
   - Kernel registers all remaining Maelstrom actions/guards (LLM bridge, ContextMap, Orchestrator, `taintPropagate`, etc.)
   - Loads and spawns:
     - `sys:gateway` + Channel Adapters (including streaming SSE/WS + 2FA)
     - `sys:admin` (outer-only)
     - `sys:memory`, `sys:heartbeat`, `sys:human-gateway`
   - Any configured top-level Agent Charts (inner/DMZ/outer)
   - All future Charts (user agents, sub-agents, etc.) are spawned via the normal Library API through `sys:lifecycle`

5. **Steady state**
   - Everything after step 4 is 100% Charts + Mail + Events
   - Kernel goes dormant (only listens for shutdown signals)

This sequence is deterministic, auditable, and hot-reload safe (re-running Bootstrap Chart can gracefully upgrade services).

### 12.2 Hard-Coded vs YAML-Defined

| Category | How Defined | Loaded By | Why |
|----------|-------------|-----------|-----|
| **Core / Foundational** (Security & Boundary, Communication, Observability, minimal Lifecycle/Registry, Bootstrap orchestrator) | Hard-coded Go structs (or embedded YAML compiled into binary) | Kernel at startup (never from disk) | Zero risk of missing/corrupt files, deterministic bootstrap, cannot be accidentally hot-reloaded or tampered with |
| **Everything else** (Gateway, Admin, Memory, Heartbeat, Human-Gateway, all Agent charts, Orchestrator templates, etc.) | Normal YAML on disk | ChartRegistry (hot-reloadable) | Full flexibility for users and higher-level services |

Only the absolute minimum needed to reach `kernel_ready` is hard-coded. Once the core services are up, the Bootstrap Chart emits `kernel_ready` and the Registry takes over loading everything else.

---

## 13. YAML Schemas

### 13.1 ChartDefinition (Base for Everything)

```yaml
apiVersion: maelstrom.dev/v1
kind: Chart
metadata:
  name: my-chart
  version: "1.2.3"
  immutable: false          # for core services
  boundary: inner | dmz | outer
spec:
  stabilityPolicy:
    maxReconfigDepth: 3
    collapseAfter: 5
  taintPolicy:
    enforcement: strict | redact | audit
    allowedOnExit: ["PII", "SECRET", "TOOL_OUTPUT"]  # taints allowed to leave
    redactRules:
      - taint: "SECRET"
        replacement: "[REDACTED]"
  persistence:
    enabled: true
    snapshotOn:
      - stateEntry: "decide"
      - every: 50 messages
      - cron: "*/15 * * * *"
    storage: "persistentVolume"  # or s3, etc.
    includeTaints: true
    includeSessionPointer: true
  expose:                     # HTTP/OpenAPI exposure
    http:
      path: /api/v1/...
      events:
        - trigger: user_query
          method: POST
          requestSchema: UserQuery
          responseSchema: AgentResponse
          stream: true
  nodes: { ... }              # Node tree (see 13.2)
  contextMap: [ ... ]         # Agent-only: ContextBlocks
  orchestrator: { ... }       # Agent-only: tool execution policy
  tools: [ ... ]              # Tool bindings
  dataSources: [ ... ]         # Persistent storage with tainting
```

### 13.2 Node (Unified Primitive)

```yaml
nodes:
  root:
    type: compound            # optional hint; derived from children
    onEnter: ["assembleContextMap"]
    onExit: ["logExit"]
    children:
      observe:
        type: atomic
        transitions:
          - event: "user_query"
            guard: "hasContext"
            action: ["dispatchLLM"]
            to: "orient"
          - event: "heartbeat"
            action: ["logHeartbeat"]
            to: "observe"          # stay in place

      orient:
        type: parallel
        regions:
          - name: shortTermMemory
            child: memoryScan
          - name: longTermMemory
            child: ragQuery
        onExit: ["mergeOrientResults"]

      decide:
        type: atomic
        onEnter: ["assembleContextMap"]
        transitions:
          - event: "llm_decision"
            guard: "needsTool"
            action: ["dispatchOrchestrator"]
            to: "act"
          - event: "llm_decision"
            guard: "noActionNeeded"
            action: ["emitFinalAnswer"]
            to: "observe"

      act:
        type: atomic
        transitions:
          - event: "orchestration_complete"
            action: ["injectToolResults"]
            to: "decide"
```

### 13.3 ContextBlock

```yaml
contextMap:
  - name: systemPrompt
    source: static
    content: "You are a secure OODA agent operating in DMZ boundary..."
    maxTokens: 2048
    compressor: none

  - name: conversation
    source: session
    strategy: lastN
    n: 30
    maxTokens: 24000
    eviction: summarize      # summarize | truncate | error
    streamCommit: false      # false = partials never enter next prompt
    taintPolicy:
      redactMode: redact     # redact | dropBlock | audit
      redactRules:
        - taint: "PII"
          replacement: "[PII REDACTED — inner data]"
        - taint: "SECRET"
          replacement: "[API KEY REDACTED]"
      allowedForBoundary: [dmz, outer]

  - name: longTermMemory
    source: memoryService
    query: "{{currentGoal}}"
    topK: 8
    maxTokens: 8000
    taintPolicy:
      redactMode: dropBlock
      allowedForBoundary: [inner]

  - name: toolsAvailable
    source: toolRegistry
    boundaryFilter: dmz       # only tools visible at this boundary
```

### 13.4 Orchestrator

```yaml
orchestrator:
  default: seq_failfast
  timeoutMs: 45000
  maxParallel: 8
  policies:
    seq_failfast:
      maxRetries: 2
      isolation: strict         # each tool in own namespace
    seq_continue:
      maxRetries: 1
      isolation: process
    par_continue:
      maxParallel: 8
      isolation: strict
```

### 13.5 Tool Binding

```yaml
tools:
  - name: webSearch
    boundary: dmz
    schema:
      input:
        type: object
        properties:
          query: { type: string }
      output:
        type: object
        properties:
          results: { type: array }
    isolation: container         # container | process | sandbox | strict
    taintOutput: ["TOOL_OUTPUT", "EXTERNAL"]

  - name: innerDbQuery
    boundary: inner
    isolation: strict
    taintOutput: ["TOOL_OUTPUT", "INNER_ONLY"]

  - name: subAgentResearch
    type: attachedSubAgent
    chartRef: research-mini-v1
    isolation: strict
    maxIterations: 10
```

### 13.6 DataSource

```yaml
dataSources:
  - name: agentWorkspace
    type: localDisk
    path: "/agents/{{.Id}}/workspace"
    mountMode: readWrite
    taintMode: inheritFromXattr   # or alwaysTaintAs, none
    defaultTaints: ["WORKSPACE"]
    xattrNamespace: user.maelstrom

  - name: sharedInnerRepo
    type: localDisk
    path: "/inner/repo"
    mountMode: readOnly
    taintMode: enforceInnerOnly
    allowedForBoundary: [inner]

  - name: s3Bucket
    type: objectStorage
    bucket: my-agent-bucket
    taintMode: inheritFromObjectTags
    region: us-east-1
```

### 13.7 PlatformService

```yaml
apiVersion: maelstrom.dev/v1
kind: PlatformService
metadata:
  name: sys:gateway
  core: true                    # treated specially by Registry
spec:
  chartRef: gateway-v1
  requiredForKernelReady: true
  replicas: 1
  persistence:
    snapshotEvery: "100 messages"
  expose:
    http:
      port: 8080
      tls: true
```

### 13.8 Persistence Config

```yaml
persistence:
  enabled: true
  snapshotOn:
    - stateEntry: "decide"
    - every: 50 messages
    - cron: "*/15 * * * *"
  storage: "persistentVolume"   # or s3, postgres, etc.
  includeTaints: true
  includeSessionPointer: true
  compression: gzip
  encryption: aes256
```

---

## 14. API Contracts & Test Specifications

### 14.1 Pure Statechart Library Contract

```go
type Library interface {
    Spawn(def ChartDefinition, initialAppCtx ApplicationContext) (RuntimeID, error)
    Control(id RuntimeID, cmd ControlCmd) error
    Dispatch(id RuntimeID, ev Event) error
    RegisterAction(name string, fn ActionFn)
    RegisterGuard(name string, fn GuardFn)
    Snapshot(id RuntimeID) (Snapshot, error)
    Restore(snap Snapshot) (RuntimeID, error)
}
```

**SHALL Statements**:
- SHALL evaluate events synchronously inside compound/parallel boundaries
- SHALL run parallel regions in isolated goroutines with bidirectional channels
- SHALL fire trace callback on every transition, entry, exit, event dispatch, sub-chart spawn
- SHALL support dynamic reclassification (atomic ↔ compound/parallel) atomically within one evaluation step
- SHALL provide `replaceDefinition` hook for hot-reload; running instance decides acceptance via guard on `definition_updated` event
- SHALL support transient sub-charts with automatic lifecycle binding to parent

**SHALL NOT Statements**:
- SHALL NOT know about Mail, LLM, taints, boundaries, files, or agents
- SHALL NOT perform any I/O (file, network, etc.)
- SHALL NOT implement persistence (only serialize/deserialize snapshots)
- SHALL NOT have opinions on security policy

### 14.2 Maelstrom Functional Unit Contracts

#### ChartRegistry

```go
type ChartRegistry interface {
    Load(defPath string, hydration HydrationContext) (ChartDefinition, error)
    Reload(id string) (ChartDefinition, error)
    Get(id string) (ChartDefinition, error)
    List() []ChartDefinition
    Watch(source Source) error
}
```

**SHALL**:
- SHALL watch sources, hydrate (env+appVars+templates), cache versioned definitions
- SHALL call `Library.replaceDefinition` on hot-reload
- SHALL reject hot-reload of immutable core charts
- SHALL enforce SHA-256 checksum verification if configured

#### Security & Boundary + Tainting + DataSources

```go
type Security interface {
    ValidateAndSanitize(mail Mail, sourceBoundary, targetBoundary BoundaryType) (Mail, error)
    TaintPropagate(obj any, newTaints []string) (any, error)
    NamespaceIsolate(runtimeId string, operation string) (IsolatedView, error)
    CheckTaintPolicy(data any, targetBoundary BoundaryType, policy TaintPolicy) (bool, error)
    ReportTaints(runtimeId string) (TaintMap, error)
    PrepareContextForBoundary(runtimeId string, boundary BoundaryType) error
}
```

**SHALL**:
- SHALL attach taints at Message creation and file read
- SHALL call `prepareContextForBoundary` inside every `assembleContextMap`
- SHALL redact/drop/audit per-block `taintPolicy` before LLM prompt
- SHALL enforce `allowedOnExit` on Mail, streams, snapshots, sub-agent returns
- SHALL use DataSource interface for all FS/object access

#### Communication

**SHALL**:
- SHALL deliver Mail at-least-once with deduplication
- SHALL upgrade to SSE/WS for streaming `partial_assistant` chunks

#### Agent Extensions

**SHALL**:
- SHALL assemble ContextMap → call LLM → route to Orchestrator
- SHALL inject tainted tool/sub-agent results as Messages

#### sys:admin

**SHALL**:
- SHALL expose k9s-style commands only on outer boundary with 2FA

#### Kernel (bootstrap)

**SHALL**:
- SHALL hard-code only Security, Communication, Observability, minimal Lifecycle
- SHALL spawn Bootstrap Chart, then hand off to ChartRegistry

---

## 15. Complete OODA Agent Example

This example exercises every layer of the architecture:

```yaml
apiVersion: maelstrom.dev/v1
kind: Chart
metadata:
  name: ooda-agent-v1
  version: "1.1.0"
  boundary: dmz                     # can call inner tools but never leaks inner data
spec:
  stabilityPolicy:
    maxReconfigDepth: 3
    collapseAfter: 5

  taintPolicy:
    enforcement: strict
    allowedOnExit: ["TOOL_OUTPUT", "SYSTEM", "USER_SUPPLIED"]   # PII/SECRET/INNER_ONLY blocked

  persistence:
    enabled: true
    snapshotOn:
      - stateEntry: "decide"
      - every: 20 messages

  nodes:
    root:
      type: compound
      onEnter: ["rehydrateIfRestored"]          # Maelstrom action
      children:
        observe:
          type: atomic
          onEnter: ["assembleContextMap"]
          transitions:
            - event: "user_query"               # from human chat / heartbeat / mail
              guard: "hasFreshInput"
              action: ["dispatchLLM"]
              to: "orient"
            - event: "heartbeat"
              action: ["logHeartbeat"]
              to: "observe"                     # stay awake

        orient:
          type: parallel
          regions:
            - name: shortTerm
              child: conversationScan
            - name: longTerm
              child: memoryQuery
          onExit: ["mergeOrientResults"]

        decide:
          type: atomic
          onEnter: ["assembleContextMap"]       # re-assemble after orient
          transitions:
            - event: "llm_decision"
              guard: "needsTool"
              action: ["dispatchOrchestrator"]
              to: "act"
            - event: "llm_decision"
              guard: "noActionNeeded"
              action: ["emitFinalAnswer"]
              to: "observe"

        act:
          type: atomic
          transitions:
            - event: "orchestration_complete"
              action: ["injectToolResults"]     # tainted results
              to: "decide"

  contextMap:
    - name: systemPrompt
      source: static
      content: "You are a secure OODA agent operating in DMZ boundary..."
      maxTokens: 2048
      compressor: none

    - name: conversation
      source: session
      strategy: lastN
      n: 30
      maxTokens: 24000
      eviction: summarize
      streamCommit: false                   # partials never enter next prompt
      taintPolicy:                          # proactive block-level policy
        redactMode: redact
        redactRules:
          - taint: "PII"
            replacement: "[PII REDACTED — inner data]"
          - taint: "SECRET"
            replacement: "[API KEY REDACTED]"
        allowedForBoundary: [dmz, outer]

    - name: longTermMemory
      source: memoryService
      query: "{{currentGoal}}"
      topK: 8
      maxTokens: 8000
      taintPolicy:
        redactMode: dropBlock
        allowedForBoundary: [inner]

    - name: toolsAvailable
      source: toolRegistry
      boundaryFilter: dmz

  orchestrator:
    default: seq_failfast
    timeoutMs: 45000
    policies:
      seq_failfast:
        maxRetries: 2
        isolation: strict
      par_continue:
        maxParallel: 8
        isolation: strict

  tools:
    - name: webSearch
      boundary: dmz
      isolation: container
      taintOutput: ["TOOL_OUTPUT", "EXTERNAL"]

    - name: innerDbQuery
      boundary: inner
      isolation: strict
      taintOutput: ["TOOL_OUTPUT", "INNER_ONLY"]

    - name: subAgentResearch
      type: attachedSubAgent
      chartRef: research-mini-v1
      isolation: strict

  expose:
    http:
      path: /chat/{sessionId}/stream
      events:
        - trigger: user_query
          method: POST
          stream: true
```

---

## 16. API Contracts & Test Specifications

### 16.1 Functional Unit Contracts

Each functional unit has explicit SHALL/SHALL NOT statements for test specification.

#### ChartRegistry

**SHALL**:
- Watch configured sources for ChartDefinition changes
- Hydrate definitions with env vars, app vars, and Go templates
- Cache versioned definitions with unique IDs
- Call `Library.replaceDefinition` on hot-reload
- Reject hot-reload of charts marked `immutable: true`
- Enforce SHA-256 checksum verification when configured

**SHALL NOT**:
- Modify running ChartRuntime state directly
- Bypass Security validation on definition load

#### Security & Boundary Service

**SHALL**:
- Attach taints at Message creation time
- Propagate taints on every copy, read, and write operation
- Call `prepareContextForBoundary` inside every `assembleContextMap`
- Redact, drop, or audit per-block `taintPolicy` before LLM prompt assembly
- Enforce `allowedOnExit` on all Mail, stream chunks, snapshots, and sub-agent returns
- Validate boundary crossings through Mail only (never direct events)
- Strip forbidden taints from stream chunks at Gateway

**SHALL NOT**:
- Allow inner data to reach outer/DMZ LLM prompts without redaction
- Permit boundary crossing via Events (only Mail)
- Bypass taint checks on file read operations

#### Communication Layer

**SHALL**:
- Deliver Mail with at-least-once guarantee
- Deduplicate via `correlationId`
- Route to `agent:<id>`, `topic:<name>`, or `sys:<service>`
- Support request-reply pattern via `correlationId` correlation
- Upgrade connections to SSE/WS for streaming `partial_assistant` chunks
- Maintain dead-letter queue for permanent failures

**SHALL NOT**:
- Deliver Events (Events are internal to ChartRuntime only)
- Guarantee ordering across different Mail targets

#### Agent Extensions

**SHALL**:
- Assemble ContextMap with token budgeting and eviction
- Call LLM with assembled prompt
- Route LLM response (tool calls) to Orchestrator
- Inject tainted tool/sub-agent results as Messages into session
- Respect per-block `taintPolicy` during assembly

**SHALL NOT**:
- Execute tool calls directly (always through Orchestrator)
- Bypass ContextMap assembly for LLM calls

#### Orchestrator

**SHALL**:
- Execute tool/sub-agent calls per policy (sequential/parallel, fail-fast/continue)
- Spawn isolated execution contexts per tool (namespace, container, etc.)
- Collect all results (success and failure) in `seq_continue` mode
- Abort remaining calls on first failure in `fail-fast` mode
- Return `orchestration_complete` or `orchestration_failed` event
- Inject all results as tainted Messages into parent session

**SHALL NOT**:
- Execute tools in parent ChartRuntime context (must isolate)
- Modify parent state directly (only via events)

#### sys:admin

**SHALL**:
- Expose k9s-style commands via Gateway
- Require 2FA/MFA for all access
- Operate only on outer boundary
- Provide: list/pause/resume/stop Charts, view session/state/taints, inject events

**SHALL NOT**:
- Access inner-boundary data (sanitized views only)
- Bypass 2FA requirements
- Modify Chart definitions directly

#### Kernel (Bootstrap)

**SHALL**:
- Hard-code only Security, Communication, Observability, minimal Lifecycle
- Instantiate pure Statechart Library
- Register minimal bootstrap actions
- Spawn hard-coded Bootstrap Chart
- Hand off to ChartRegistry on `kernel_ready`

**SHALL NOT**:
- Load any YAML from disk before `kernel_ready`
- Implement domain logic (LLM, ContextMap, etc.)
- Remain active after bootstrap (goes dormant)

---

## 17. Test Specifications

### 17.1 Unit Tests (Per-Block, Fast, Deterministic)

**Statechart Engine**:
- SHALL evaluate state transitions correctly
- SHALL handle dynamic reclassification atomically
- SHALL isolate parallel regions with proper event ordering
- SHALL fire trace callbacks on all lifecycle events

**Security & Isolation & Tainting**:
- SHALL block boundary violations
- SHALL propagate taints on copy/read/write
- SHALL enforce DLP-style policies
- SHALL prevent namespace leakage

**ContextMap**:
- SHALL assemble context correctly
- SHALL enforce token budgeting
- SHALL apply eviction strategies
- SHALL preserve compressor round-trips

**Orchestrator**:
- SHALL execute seq/parallel modes correctly
- SHALL handle fail-fast vs continue policies
- SHALL isolate tool execution

**Gateway Adapters**:
- SHALL normalize Mail correctly
- SHALL handle SSE/WS streaming
- SHALL enforce 2FA on admin paths

**Platform Services**:
- SHALL handle their specific commands correctly
- SHALL emit proper events

### 17.2 Integration Tests (Key Flows)

**Full LLMReason Loop**:
```
ContextMap → LLM → Orchestrator → tool isolation → tainted result injection → next state
```

**Streaming Chat**:
```
human → Agent → partial_assistant chunks (with taint stripping) → reconnect/resume
```

**Boundary Crossing**:
```
outer human mail → DMZ Agent → inner tool (verify taints blocked/stripped)
```

**Persistence**:
```
start → snapshot → crash → resume (verify taints, isolation, session integrity)
```

**Admin Console**:
```
authenticated query → control commands → observe live state/taints (2FA path)
```

**Heartbeat**:
```
heartbeat → long-running Agent (isolation + taint on injected message)
```

### 17.3 Security & Property Tests

**Fuzz Tests**:
- Malformed Mail handling
- Tainted data exfil attempts
- Prompt-injection simulations

**Chaos Tests**:
- Kill Engine mid-orchestration → recovery
- Taint preservation across crashes

**Runaway Prevention**:
- Max iterations guards
- stabilityPolicy enforcement
- Resource quota enforcement

**Isolation Audit**:
- Attempt cross-session memory read → must fail
- Attempt inner data access from outer → must fail

**Taint Coverage**:
- Every data path (in-mem, on-disk, stream, mail) must propagate and enforce

### 17.4 E2E / Acceptance (YAML-Defined Scenarios)

**"Safe tool use in DMZ" chart**:
- Agent in DMZ calls tool
- Tool returns tainted data
- Data redacted before reaching outer boundary

**"Long-running OODA with human chat + admin debug"**:
- Agent runs OODA loop
- Human chats via `/chat/{agentId}`
- Admin inspects state via `sys:admin`
- All boundaries enforced

**"Tainted PII never leaves inner boundary"**:
- File with PII tag read from workspace
- Taint propagates through session
- Attempt to send to outer agent → blocked/redacted

---

## 18. Architecture Dependency Tree

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
├── Statechart Engine (pure library)
└── ChartRegistry + Kernel (hard-coded bootstrap)
```

**Longest dependency chain**: 5 layers (acceptable for MVP)

**Example**: Streaming UI / Human Chat
```
UI → Gateway → Communication → Security → Engine
```

**No cycles** after bootstrap. All implicit clusters (Safety, Intelligence, I/O) are explicit Platform Services.

---

## 19. Design Decision Summary

| Decision | Semantics | Rationale |
|----------|-----------|-----------|
| **Single source of truth = Statechart** | All control flow via nodes, events, actions/guards | Unified reasoning, lifecycle, and orchestration model |
| **Unified Node primitive** | State type derived from #children (0=atomic, 1=compound, ≥2=parallel) | Simpler YAML, cleaner implementation, dynamic reclassification |
| **Inner / DMZ / Outer boundaries** | Hard partition of data & execution | Prevent prompt injection, data exfiltration |
| **Proactive + reactive tainting** | Taints at creation, enforcement at boundary crossing + ContextMap assembly | Defense in depth, automatic redaction |
| **DataSource tainting** | xattr/object tags on write, propagate on read | File/workspace data protected same as Messages |
| **Hard-coded core vs YAML** | Only bootstrap kernel is hard-coded; everything else YAML | Deterministic startup, flexible runtime |
| **Orchestrator as sub-chart** | All tool/sub-agent calls through reusable Orchestrator | Single abstraction for all execution patterns |
| **Mail vs Events** | Mail for cross-boundary, Events for internal | Clean separation, security enforceable |
| **Streaming via Gateway** | SSE/WS adapters normalize to Mail + partial_assistant | Standard LLM chat UX out of the box |

---

## 20. Prompt-Injection & Exfil Prevention Guarantees (MVP)

**SHALL**:
- Never allow outer/DMZ ContextMap to contain untagged inner data
- Redact or drop any block with PII/SECRET/INNER_ONLY taints for DMZ/outer charts
- File reads from workspaces inject taints before entering session
- Enforce `allowedOnExit` on all outbound data paths

**SHALL NOT**:
- Any LLM prompt in DMZ/outer ever contain raw inner data
- Permit Events to cross boundary (only Mail)
- Allow tool execution without Orchestrator isolation

**DevOps SHALL**:
- Enforce container sandbox (Landlock/seccomp), network policies, read-only volumes for full exfil protection

---

*End of Maelstrom Architecture v1.0*
