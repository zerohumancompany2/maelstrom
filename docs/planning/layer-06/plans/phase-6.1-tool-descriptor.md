# Phase 6.1: Tool Descriptor Extensions

## Goal
Enhance ToolDescriptor with 4 new fields following strict TDD workflow with 4 tests.

## Scope
- Add `TaintOutput` field to ToolDescriptor
- Add `Type` field for sub-agent identification
- Add `ChartRef` field for sub-agent chart reference
- Add `MaxIterations` field for sub-agent iteration limits
- Create 4 tests in `pkg/services/tools/service_test.go`

## Current Status

### Implementation Status
| Test | Status | Notes |
|------|--------|-------|
| Test 1: TestToolDescriptor_HasTaintOutputField | ❌ Missing | TaintOutput field does not exist |
| Test 2: TestToolDescriptor_HasSubAgentFields | ❌ Missing | Type, ChartRef, MaxIterations fields do not exist |
| Test 3: TestToolDescriptor_SubAgentConfiguration | ❌ Missing | Sub-agent configuration not supported |
| Test 4: TestToolDescriptor_AllFieldsCoexist | ❌ Missing | All 8 fields must coexist |

### Files Status
| File | Status |
|------|--------|
| `pkg/services/tools/service.go` | ⚠️ Partial - needs 4 new fields in ToolDescriptor |
| `pkg/services/tools/service_test.go` | ⚠️ Partial - needs 4 new tests |

## Required Implementation

### Enhanced ToolDescriptor Struct
```go
type ToolDescriptor struct {
    Name          string
    Boundary      string
    Schema        map[string]any
    Isolation     string
    TaintOutput   []string
    Type          string
    ChartRef      string
    MaxIterations int
}
```

### arch-v1.md References

**TaintOutput Field** (arch-v1.md Section 8.4, lines 562, 566):
```yaml
tools:
  - name: webSearch
    boundary: dmz
    isolation: container
    taintOutput: ["TOOL_OUTPUT"]
  - name: innerDbQuery
    boundary: inner
    isolation: strict
    taintOutput: ["TOOL_OUTPUT", "INNER_ONLY"]
```

**Sub-Agent Fields** (arch-v1.md Section 8.4, lines 568-569):
```yaml
  - name: subAgentResearch
    type: attachedSubAgent
    chartRef: research-mini-v1
    isolation: strict
```

**Tool Resolution** (arch-v1.md Section 8.4, lines 574-578):
```go
resolveTool(name string, callerBoundary BoundaryType) → ToolDescriptor | notFound
```
Returns sanitized schema filtered by caller's boundary. Inner tools are invisible to outer callers.

## Tests to Write (4 tests, 4 commits)

### Test 1: TestToolDescriptor_HasTaintOutputField
```go
func TestToolDescriptor_HasTaintOutputField()
```
**Acceptance Criteria:**
- ToolDescriptor has TaintOutput field (arch-v1.md line 562, 566)
- TaintOutput is []string type
- Can be set to ["TOOL_OUTPUT"] or ["TOOL_OUTPUT", "INNER_ONLY"]

### Test 2: TestToolDescriptor_HasSubAgentFields
```go
func TestToolDescriptor_HasSubAgentFields()
```
**Acceptance Criteria:**
- ToolDescriptor has Type field (arch-v1.md line 568)
- ToolDescriptor has ChartRef field (arch-v1.md line 569)
- ToolDescriptor has MaxIterations field (arch-v1.md line 569)
- Type can be "attachedSubAgent"

### Test 3: TestToolDescriptor_SubAgentConfiguration
```go
func TestToolDescriptor_SubAgentConfiguration()
```
**Acceptance Criteria:**
- Sub-agent tool has Type="attachedSubAgent"
- Sub-agent tool has ChartRef pointing to valid chart
- Sub-agent tool has MaxIterations > 0

### Test 4: TestToolDescriptor_AllFieldsCoexist
```go
func TestToolDescriptor_AllFieldsCoexist()
```
**Acceptance Criteria:**
- All 8 fields exist on ToolDescriptor
- Fields can be set independently
- No conflicts with existing fields

## Dependencies

### Test Dependencies
```
Test 1 → Test 4 (TaintOutput structure)
Test 2 → Test 3, Test 4 (Sub-agent fields)
Test 3 → Test 4 (Sub-agent configuration)
```

### Phase Dependencies
- None - Phase 6.1 has no dependencies

## Files to Create/Modify

| File | Operation | Changes |
|------|-----------|---------|
| `pkg/services/tools/service.go` | MODIFY | Add TaintOutput, Type, ChartRef, MaxIterations fields to ToolDescriptor |
| `pkg/services/tools/service_test.go` | MODIFY | Add 4 new tests |

## Implementation Order (TDD)

1. **Test 1**: Write test → verify RED → Add TaintOutput field → verify GREEN → commit
2. **Test 2**: Write test → verify RED → Add Type, ChartRef, MaxIterations fields → verify GREEN → commit
3. **Test 3**: Write test → verify RED → Validate sub-agent configuration → verify GREEN → commit
4. **Test 4**: Write test → verify RED → Verify all fields coexist → verify GREEN → commit

**Total: 4 tests, 4 commits**

## Deliverables
- Modified `pkg/services/tools/service.go` with 4 new fields in ToolDescriptor
- Updated `pkg/services/tools/service_test.go` with 4 passing tests
- 4 commits (one per test per TDD workflow)