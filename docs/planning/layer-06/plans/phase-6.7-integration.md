# Phase 6.7: Integration & Schema Sanitization

**Phase ID**: L6-P6.7  
**Title**: Integration & Schema Sanitization  
**Parent**: Layer 6 (Tool Calls & Orchestration)  
**Status**: ❌ PENDING

## Parent Requirements

From `arch-v1.md L496-505` - Tool & Action Orchestration (Unified):
1. Tool calls and sub-agent calls are both "actions" that may require orchestration
2. Single configurable Orchestrator sub-chart parameterized by execution mode, failure policy, completion signaling
3. LLM response handed to orchestrator named in current state's binding (or chart default)
4. Orchestrator instances run as transient parallel regions inside LLMReason compound state

From `arch-v1.md L578` - Tool Resolution:
- Returns sanitized schema filtered by caller's boundary
- Inner tools are invisible to outer callers

## Dependencies

- Phase 6.1: Tool Descriptor Enhancement
- Phase 6.2: Execution Policies
- Phase 6.3: Orchestrator Service
- Phase 6.4: Sequential Executor
- Phase 6.5: Parallel Executor
- Phase 6.6: Sub-Agent Support

## Satisfied Lower-Layer Requirements

- Layer 5: ToolRegistry basic structure exists (arch-v1.md L489, layer-05-phase-breakdown.md)
- Layer 4: Boundary types (inner, dmz, outer) defined (arch-v1.md L261-270)
- Layer 3: Mail system with correlationId for fire-and-await (arch-v1.md L604)
- Layer 2: sys:lifecycle for spawning charts (arch-v1.md L460)
- Layer 0: Statechart Engine with transient sub-charts (arch-v1.md L402)

## Acceptance Criteria

1. SanitizeSchema() filters schema by caller boundary
2. End-to-end orchestration flow integrates all components (ToolRegistry, OrchestratorService, Executors)
3. sys:orchestrator service registration verified
4. Full orchestration: tool calls resolved → executed → results injected

## Test Descriptions

### Test 1: TestOrchestratorService_SanitizeSchema

**Given**: A ToolDescriptor with boundary=inner and a callerBoundary=outer  
**When**: SanitizeSchema() is called with the tool and caller boundary  
**Then**: Tool is not returned (inner tools invisible to outer callers)  
**Expected Result**: Returns error or empty schema for boundary mismatch  
**Spec Reference**: arch-v1.md L578 (inner tools invisible to outer callers)

### Test 2: TestOrchestratorService_FullOrchestrationFlow

**Given**: OrchestratorService wired with ToolRegistry, SequentialExecutor, ParallelExecutor, SubAgentExecutor  
**When**: Execute() is called with tool calls and seq_failfast policy  
**Then**: Tools are resolved → executed sequentially → results injected into databag  
**Expected Result**: End-to-end orchestration flow completes successfully with all components integrated  
**Spec Reference**: arch-v1.md L496-505 (unified orchestration model)

## Implementation Plan

### Files to Modify

1. `pkg/services/tools/service.go` - Add SanitizeSchema() method
2. `pkg/orchestrator/orchestrator_test.go` - Create integration tests

### Stubs Needed

```go
// pkg/services/tools/service.go
func (s *toolsService) SanitizeSchema(tool ToolDescriptor, callerBoundary mail.BoundaryType) JSONSchema
```

### Implementation Order

1. Create `pkg/orchestrator/orchestrator_test.go` with test stubs
2. Write Test 1 → RED → Implement SanitizeSchema() → GREEN → commit
3. Write Test 2 → RED → Wire all components together → GREEN → commit

## Commit Plan

| Test | Commit Message |
|------|----------------|
| Test 1 | `feat(layer-6/orchestrator): add SanitizeSchema boundary filtering` |
| Test 2 | `feat(layer-6/orchestrator): verify end-to-end orchestration flow` |

## Deliverables

- 2 commits (1:1 test-to-commit ratio)
- All tests passing
- Files modified: `pkg/services/tools/service.go`, `pkg/orchestrator/orchestrator_test.go`
- sys:orchestrator service registration verified

---

*Document generated: 2026-03-10*  
*Based on: arch-v1.md v1.3, layer-06-phase-breakdown.md*