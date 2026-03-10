# Phase P6.6: Sub-Agent Lifecycle

**Phase ID**: L6-P6.6  
**Title**: Sub-Agent Lifecycle  
**Parent**: Phase 6.6 (Sub-Agent Support)  
**Status**: ❌ PENDING

## Parent Requirements

From `arch-v1.md L439-446` - Sub-Agents:

1. **Attached sub-agents** (arch-v1.md L443): Spawns a child ChartRuntime whose lifecycle is bound to the parent (auto-terminated on parent exit). Parent can share a scoped slice of its application context. Result returned via a `subAgentDone` event + optional Messages injected into parent session.

2. **Detached sub-agents** (arch-v1.md L445): Spawns a new top-level ChartRuntime managed by `sys:lifecycle`. Fire-and-forget or fire-and-await via mail `correlationId`.

From `arch-v1.md L441`: Sub-agents are invoked exactly like tools (they appear in the tool registry when declared).

## Dependencies

- Phase 6.1 (Tool Descriptor Enhancement): Type, ChartRef, MaxIterations fields for sub-agent identification
- Phase 6.3 (Orchestrator Service): Execute() method for spawning orchestrator charts
- Phase 6.4 (Sequential Executor): Sequential execution patterns for sub-agent orchestration
- Phase 6.5 (Parallel Executor): Parallel execution patterns for concurrent sub-agents

## Satisfied Lower-Layer Requirements

- Layer 5: ToolRegistry exists with basic structure (arch-v1.md L289-328)
- Layer 5: sys:lifecycle service skeleton exists for top-level runtime management (arch-v1.md L462-474)
- Layer 0: ChartRuntime type defined for spawning child runtimes (arch-v1.md Section 2)
- Layer 3: Mail system with correlationId field for fire-and-await (arch-v1.md L49)

## Acceptance Criteria

1. Attached sub-agent spawns child ChartRuntime bound to parent lifecycle
2. Attached sub-agent auto-terminated on parent exit
3. Detached sub-agent spawns top-level ChartRuntime managed by sys:lifecycle
4. Detached sub-agent supports fire-and-forget and fire-and-await via correlationId
5. subAgentDone event emitted with result and optional Messages injected into parent session

## Test Descriptions

### Test 1: TestSubAgent_AttachedLifecycleBound

**Given**: Parent agent at inner boundary spawns attached sub-agent with ChartRef  
**When**: Sub-agent is created with Type="attachedSubAgent"  
**Then**: Child ChartRuntime spawned with lifecycle bound to parent  
**Expected Result**: Attached sub-agent has parent runtime reference, lifecycle tied to parent (arch-v1.md L443)  
**Spec Reference**: arch-v1.md L443 (Attached: lifecycle bound to parent)

### Test 2: TestSubAgent_AttachedAutoTerminated

**Given**: Parent agent with attached sub-agent running  
**When**: Parent agent exits (statechart completes or is terminated)  
**Then**: Attached sub-agent is automatically terminated  
**Expected Result**: Sub-agent ChartRuntime stopped, resources cleaned up (arch-v1.md L443)  
**Spec Reference**: arch-v1.md L443 (auto-terminated on parent exit)

### Test 3: TestSubAgent_DetachedFireAndForget

**Given**: Parent agent spawns detached sub-agent with Type="detached"  
**When**: Sub-agent creation completes  
**Then**: Top-level ChartRuntime spawned, managed by sys:lifecycle  
**Expected Result**: Sub-agent independent of parent, no lifecycle binding (arch-v1.md L445)  
**Spec Reference**: arch-v1.md L445 (Detached: managed by sys:lifecycle, fire-and-forget)

### Test 4: TestSubAgent_DetachedFireAndAwait

**Given**: Parent agent spawns detached sub-agent with correlationId  
**When**: Sub-agent completes and emits result  
**Then**: Result delivered to parent via mail with matching correlationId  
**Expected Result**: Parent receives sub-agent result via mail correlation mechanism (arch-v1.md L445)  
**Spec Reference**: arch-v1.md L445 (fire-and-await via mail correlationId)

### Test 5: TestSubAgent_subAgentDoneEvent

**Given**: Attached sub-agent completes execution with result  
**When**: Sub-agent emits completion signal  
**Then**: subAgentDone event emitted with result, optional Messages injected into parent session  
**Expected Result**: Parent receives subAgentDone event with result data (arch-v1.md L443)  
**Spec Reference**: arch-v1.md L443 (Result returned via subAgentDone event + optional Messages)

## Implementation Plan

### Files to Modify

1. `pkg/orchestrator/subagent.go` - Create sub-agent lifecycle management
2. `pkg/orchestrator/types.go` - Add SubAgentType constants and SubAgentConfig struct

### Stubs Needed

```go
// pkg/orchestrator/subagent.go
type SubAgentType string

const (
    SubAgentAttached  SubAgentType = "attached"
    SubAgentDetached  SubAgentType = "detached"
)

type SubAgentConfig struct {
    Type           SubAgentType
    ChartRef       string
    MaxIterations  int
    InheritContext bool
    CorrelationId  string
}

type SubAgentExecutor struct {
    config       SubAgentConfig
    parentNs     string
    parentRuntime statechart.RuntimeID
}

func NewSubAgentExecutor(config SubAgentConfig, parentNs string, parentRuntime statechart.RuntimeID) *SubAgentExecutor
func (e *SubAgentExecutor) Execute() (any, error)
func (e *SubAgentExecutor) spawnAttached() (statechart.RuntimeID, error)
func (e *SubAgentExecutor) spawnDetached() (statechart.RuntimeID, error)
func (e *SubAgentExecutor) emitSubAgentDone(result any, messages []mail.Mail) error
func (e *SubAgentExecutor) setupAutoTermination() error
```

### Implementation Order

1. Create `pkg/orchestrator/subagent.go` with stub implementations
2. Write Test 1 → RED → Implement attached sub-agent spawning → GREEN → commit
3. Write Test 2 → RED → Implement auto-termination on parent exit → GREEN → commit
4. Write Test 3 → RED → Implement detached sub-agent spawning → GREEN → commit
5. Write Test 4 → RED → Implement fire-and-await via correlationId → GREEN → commit
6. Write Test 5 → RED → Implement subAgentDone event emission → GREEN → commit

## Commit Plan

| Test | Commit Message |
|------|----------------|
| Test 1 | `feat(layer-6/subagent): add attached sub-agent lifecycle binding` |
| Test 2 | `feat(layer-6/subagent): add auto-termination on parent exit` |
| Test 3 | `feat(layer-6/subagent): add detached sub-agent fire-and-forget` |
| Test 4 | `feat(layer-6/subagent): add fire-and-await via correlationId` |
| Test 5 | `feat(layer-6/subagent): add subAgentDone event emission` |

## Deliverables

- 5 commits (1:1 test-to-commit ratio)
- All tests passing
- Files modified: `pkg/orchestrator/subagent.go`, `pkg/orchestrator/types.go`

---

*Document generated: 2026-03-10*  
*Based on: arch-v1.md v1.3, layer-06-phase-breakdown.md*