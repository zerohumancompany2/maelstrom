# Layer 6: Tool Calls & Orchestration

**Reference**: `docs/arch-v1.md` Section 8  
**Parent Scope**: `docs/planning/implementation-scope.md`  
**Status**: ❌ PENDING

---

## Overview

Layer 6 implements tool calls and orchestration, providing a unified abstraction for executing tools and sub-agents with configurable policies. Instead of four hardcoded variants, there is a single configurable **Orchestrator sub-chart** (itself a reusable Chart) parameterized by execution mode, failure policy, and completion signaling.

**Key Design Principle**: Tool calls and sub-agent calls are both "actions" that may require orchestration. The Orchestrator is a reusable sub-chart that runs as transient parallel regions inside the `LLMReason` compound state.

---

## Architectural Dependencies

```
Layer 6 depends on:
├── Layer 5 (Platform Services)
│   ├── sys:tools - Tool registry for resolution
│   └── sys:lifecycle - Spawning orchestrator charts
├── Layer 4 (Security & Boundaries)
│   └── Boundaries for tool access control
└── Layer 0 (Statechart Engine) ✓ COMPLETE
    └── Sub-chart spawning and lifecycle
```

**Why Layer 5 First**: The Orchestrator depends on sys:tools for tool resolution and sys:lifecycle for spawning orchestrator charts.

---

## Orchestrator as Sub-Chart (arch-v1.md 8.1)

**Unified Orchestration Model**:
- Tool calls and sub-agent calls are both treated as "actions" that may require orchestration
- Instead of four hardcoded variants, there is a **single configurable Orchestrator sub-chart** (itself a reusable Chart)
- Orchestrator instances run as **transient parallel regions** inside the `LLMReason` compound state
- The LLM response is always handed to the orchestrator named in the current state's binding (or chart default)

**Key Design Principle**: Single abstraction for all execution patterns (sequential and parallel). All tool/sub-agent calls go through Orchestrator (never executed directly by agent).

---

## Orchestrator YAML Shape (arch-v1.md 8.2)

```yaml
orchestrator:
  default: seq_failfast          # or seq_continue, par_continue
  timeoutMs: 30000-45000
  maxParallel: 8                  # only for par_* patterns
  policies:
    seq_failfast:
      maxRetries: 2
      isolation: strict           # each tool in own namespace
    seq_continue:
      maxRetries: 1
      isolation: process
    par_continue:
      maxParallel: 8
      isolation: strict
```

### Orchestrator Parameters

| Parameter | Purpose |
|-----------|---------|
| `default` | Default execution pattern to use |
| `timeoutMs` | Maximum execution time (30-45 seconds typical) |
| `maxParallel` | Max concurrent tools (only for parallel patterns) |
| `policies` | Named execution configurations with isolation settings |

### Isolation Types

- `strict` - each tool in own namespace
- `process` - process-level isolation
- `container` - container isolation
- `sandbox` - sandboxed execution

---

## Execution Patterns (arch-v1.md 8.3)

### Sequential, No-Fail-On-Fail (`seq_continue`)

```
tool call → awaiting → tool call → awaiting → ...
```

- Each tool call adds its output to the databag (scoped to calling chart)
- **Includes failed tool calls** in results
- When results re-inserted into session context, includes failed tool-call output

### Sequential, Fail-On-First-Fail (`seq_failfast`)

- **Short-circuiting** AND application on tool-calls
- First failed tool call **aborts all following calls**
- All tool output injected back into session context (scoped to session)

### Parallel, No-Fail-On-Fail (`par_continue`)

- Spawns **parallel states** for each tool-call
- All execute **concurrently**
- Session context insertion as above
- Caller continues **immediately** (async completion)

### Parallel, Fail-On-First-Fail (`par_failfast`) - DROPPED

- **Dropped from MVP** due to complexity concerns
- Rationale: "too racy, not worth complexity"
- Key insight: "Fail-fast is almost always sequential in practice"

---

## Tool Registry & Schema (arch-v1.md 8.4)

### Tool Registration with Boundary-Aware Schemas

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

### Tool Resolution API

```go
resolveTool(name string, callerBoundary BoundaryType) → ToolDescriptor | notFound
```
- Returns **sanitized schema filtered by caller's boundary**
- Inner tools are **invisible to outer callers**

### Tool Binding Schema Details

| Field | Description |
|-------|-------------|
| `name` | Tool identifier |
| `boundary` | inner \| dmz \| outer - access level |
| `schema` | Input/output JSON schema |
| `isolation` | execution isolation mode |
| `taintOutput` | taint categories on output |
| `type` | optional: `attachedSubAgent` for sub-agents |
| `chartRef` | optional: chart reference for sub-agents |
| `maxIterations` | optional: for sub-agent limits |

---

## AgentSpec Fields (arch-v1.md 3.6)

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

### Key AgentSpec Fields

| Field | Purpose |
|-------|---------|
| `llmConfig.modelFamily` | LLM model family identifier |
| `llmConfig.temperature` | Sampling temperature for LLM inference |
| `llmConfig.maxTokens` | Maximum tokens for LLM response |
| `llmConfig.streaming` | Enable/disable streaming responses |
| `contextMap` | Array of ContextBlock definitions |
| `toolBindings` | Global and per-state tool bindings |
| `orchestratorPolicy` | Default execution pattern + timeout |
| `subAgentPolicy` | Sub-agent depth limit + context inheritance |
| `stabilityPolicy` | Reconfiguration guards |

---

## Sub-Agents (arch-v1.md 6.4)

Sub-agents are invoked exactly like tools (they appear in the tool registry when declared).

### Attached Sub-Agents

- Spawns a **child ChartRuntime** whose lifecycle is **bound to the parent**
- **Auto-terminated on parent exit**
- Parent can **share a scoped slice** of its application context
- Result returned via a `subAgentDone` event + optional Messages injected into parent session

### Detached Sub-Agents

- Spawns a **new top-level ChartRuntime** managed by `sys:lifecycle`
- **Fire-and-forget** or **fire-and-await** via mail `correlationId`

### Key Characteristics

- Sub-agents are **invoked exactly like tools** (appear in tool registry when declared)
- Can be **attached** (lifecycle bound) vs **detached** (top-level, fire-and-forget)

---

## Service Implementation Details

### OrchestratorService

```go
type OrchestratorService struct {
    id          string  // "sys:orchestrator"
    registry    *ToolRegistry
    library     statechart.Library
    policies    map[string]*ExecutionPolicy
    mu          sync.RWMutex
}

type ExecutionPolicy struct {
    Mode          string  // seq_continue, seq_failfast, par_continue
    MaxRetries    int
    Isolation     string  // strict, process, container, sandbox
    MaxParallel   int     // only for parallel modes
    TimeoutMs     int
}

func NewOrchestratorService() *OrchestratorService

func (s *OrchestratorService) ID() string

func (s *OrchestratorService) HandleMail(mail Mail) error

func (s *OrchestratorService) ResolveTool(name string, callerBoundary BoundaryType) (ToolDescriptor, error)

func (s *OrchestratorService) Execute(tools []ToolCall, policy ExecutionPolicy) (statechart.RuntimeID, error)

func (s *OrchestratorService) RegisterPolicy(name string, policy ExecutionPolicy) error
```

### ToolRegistry

```go
type ToolRegistry struct {
    tools     map[string]ToolDescriptor
    mu        sync.RWMutex
}

type ToolDescriptor struct {
    Name        string
    Boundary    BoundaryType
    Schema      JSONSchema
    Isolation   string  // container, process, sandbox, strict
    TaintOutput []string
    Type        string  // tool, attachedSubAgent
    ChartRef    string  // for sub-agents
    MaxIterations int   // for sub-agents
}

func NewToolRegistry() *ToolRegistry

func (r *ToolRegistry) Register(tool ToolDescriptor) error

func (r *ToolRegistry) Resolve(name string, callerBoundary BoundaryType) (ToolDescriptor, error)

func (r *ToolRegistry) ListTools(boundary BoundaryType) []ToolDescriptor

func (r *ToolRegistry) SanitizeSchema(tool ToolDescriptor, callerBoundary BoundaryType) JSONSchema
```

### ToolCall Structure

```go
type ToolCall struct {
    Name      string
    Arguments any
    TimeoutMs int
}
```

---

## End-to-End Execution Flow (arch-v1.md 10.2)

```
User types → Gateway (SSE connect) → Communication.publish(mail: type=user, taints=["USER_SUPPLIED"], boundary=outer)
Security.validate → DMZ Agent inbox
Engine.dispatchEvent → enters LLMReason state
AgentExtensions.assembleContextMap (pulls session + Memory blocks; taints propagated) → LLM call
LLM returns tool calls → Orchestrator (isolated sub-chart) → Tool Registry.resolve → executeIsolated (new namespace)
Tool result → taintPropagate → new tool_result Mail (taints=["TOOL_OUTPUT"]) injected to session
Engine emits partial_assistant Messages (stream: true) → Communication → Gateway → Security.stripForbiddenTaints → SSE chunks to user
```

---

## Orchestrator Chart Definition

```yaml
id: sys:orchestrator
version: 1.0.0
type: parallel
regions:
  - name: sequential
    child: sequentialExecutor  # for seq_* patterns
  - name: parallel
    child: parallelExecutor    # for par_* patterns

states:
  sequentialExecutor:
    type: compound
    children:
      execute:
        type: atomic
        onEnter: ["executeTool"]
        transitions:
          - event: "toolDone"
            action: ["injectResult"]
            to: "nextOrComplete"
      nextOrComplete:
        type: atomic
        transitions:
          - event: "moreTools"
            to: "execute"
          - event: "complete"
            to: "done"
            actions: ["emitOrchestrationComplete"]

  parallelExecutor:
    type: parallel
    regions:
      - name: tools
        child: parallelToolExecutor
    onExit: ["waitAllComplete"]

  done:
    type: final
```

---

## File Structure

```
maelstrom/
├── pkg/orchestrator/
│   ├── orchestrator.go      # OrchestratorService (~200 lines)
│   ├── policies.go          # Execution policies (~100 lines)
│   ├── executor.go          # Tool executors (~150 lines)
│   ├── registry.go          # ToolRegistry (~150 lines)
│   └── orchestrator_test.go # Tests (~300 lines)
│
├── pkg/tools/
│   └── types.go             # Tool types (~50 lines)
│
└── pkg/services/
    └── tools/
        └── service.go       # ToolRegistry integration (~100 lines)
```

---

## TDD Implementation Plan

### Phase 6.1: Tool Types

**Test**: `TestToolCall_Structure`
- ToolCall structure correct

**Test**: `TestToolDescriptor`
- ToolDescriptor fields correct

**Implementation**:
- ~50 lines types.go

### Phase 6.2: ToolRegistry

**Test**: `TestToolRegistry_RegisterResolve`
- Registration and resolution work

**Test**: `TestToolRegistry_BoundaryFiltering`
- Boundary filtering works

**Test**: `TestToolRegistry_SanitizeSchema`
- Schema sanitization works

**Implementation**:
- ~150 lines registry.go

### Phase 6.3: Execution Policies

**Test**: `TestExecutionPolicy_SeqContinue`
- Sequential continue policy works

**Test**: `TestExecutionPolicy_SeqFailFast`
- Sequential fail-fast policy works

**Test**: `TestExecutionPolicy_ParContinue`
- Parallel continue policy works

**Implementation**:
- ~100 lines policies.go

### Phase 6.4: OrchestratorService

**Test**: `TestOrchestratorService_Execute`
- Execution works correctly

**Test**: `TestOrchestratorService_Isolation`
- Isolation works correctly

**Test**: `TestOrchestratorService_Timeout`
- Timeout enforcement works

**Implementation**:
- ~200 lines orchestrator.go

### Phase 6.5: Executors

**Test**: `TestSequentialExecutor`
- Sequential execution works

**Test**: `TestParallelExecutor`
- Parallel execution works

**Test**: `TestFailFastExecutor`
- Fail-fast behavior works

**Implementation**:
- ~150 lines executor.go

### Phase 6.6: Sub-Agents

**Test**: `TestAttachedSubAgent`
- Attached sub-agent lifecycle bound

**Test**: `TestDetachedSubAgent`
- Detached sub-agent fire-and-forget works

**Implementation**:
- ~100 lines subagent.go

### Phase 6.7: Integration

**Test**: `TestFullOrchestrationFlow`
- End-to-end orchestration works

**Test**: `TestOrchestratorService_Integration`
- Service integrated with sys:tools

**Implementation**:
- Wire all components together

---

## Dependencies & External Libraries

- `github.com/maelstrom/v3/pkg/statechart` - Library interface for spawning orchestrator charts
- `github.com/maelstrom/v3/pkg/mail` - Mail types
- `sync` - Mutex for thread-safe state

---

## Risk Assessment

| Risk | Likelihood | Impact | Mitigation |
|------|-----------|--------|------------|
| Orchestration deadlock | Low | High | Comprehensive tests with timeouts |
| Isolation bypass | Medium | High | Strict namespace isolation tests |
| Policy misconfiguration | Medium | Medium | Validation at service start |
| Sub-agent recursion | Low | Medium | MaxDepth enforcement |

---

## Open Questions

1. **Execution engine**: In-process, container, or sandbox?
2. **Result aggregation**: How to aggregate results from parallel tools?
3. **Error handling**: Retry strategies, circuit breakers?
4. **Sub-agent limits**: MaxDepth, maxIterations defaults?

---

## References

- `docs/arch-v1.md` Section 8 - Tool Calls & Orchestration
- `docs/arch-v1.md` Section 3.6 - AgentSpec Fields
- `docs/arch-v1.md` Section 6.4 - Sub-Agents
- `docs/arch-v1.md` Section 14.2 - Orchestrator SHALL Statements
- `docs/planning/layer-05-platform-services.md` - Layer 5 (dependency)
- `docs/planning/implementation-scope.md` - Parent scope document

---
