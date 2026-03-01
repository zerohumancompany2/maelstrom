# Phase 3 Plan: Chart Definition Consolidation & GetID() Addition

## User Story
As a Maelstrom developer, I want a single authoritative `ChartDefinition` type in `pkg/statechart` with a `GetID()` method, so that all code uses the same type consistently and can extract chart IDs reliably.

## Acceptance Criteria
1. `pkg/statechart.ChartDefinition` has `GetID()` method returning `c.ID`
2. All imports of `pkg/chart.ChartDefinition` replaced with `pkg/statechart.ChartDefinition`
3. `pkg/chart` package deleted (registry.go, hydrator.go, hydrator_test.go, registry_test.go)
4. All existing tests pass (118+ tests)
5. No references to `pkg/chart.ChartDefinition` remain in codebase

## Frozen Semantics & API

### New Method on ChartDefinition
```go
// In pkg/statechart/types.go, after ChartDefinition struct:
func (c ChartDefinition) GetID() string {
    return c.ID
}
```

### Files to Modify
1. `pkg/statechart/types.go` - Add `GetID()` method
2. `pkg/kernel/kernel.go` - Replace import `pkg/chart` → `pkg/statechart`
3. `pkg/kernel/kernel_test.go` - Replace import `pkg/chart` → `pkg/statechart`
4. `pkg/bootstrap/chart.go` - Replace import `pkg/chart` → `pkg/statechart`
5. `pkg/bootstrap/chart_test.go` - Replace import `pkg/chart` → `pkg/statechart`
6. `pkg/chart/registry.go` - DELETE
7. `pkg/chart/hydrator.go` - DELETE
8. `pkg/chart/registry_test.go` - DELETE
9. `pkg/chart/hydrator_test.go` - DELETE

### Files to Review for Import Updates
- `pkg/kernel/kernel.go` (line 10) - currently imports `pkg/chart`
- `pkg/kernel/kernel_test.go` (line 8) - currently imports `pkg/chart`
- `pkg/bootstrap/chart.go` (line 6) - currently imports `pkg/chart`
- `pkg/bootstrap/chart_test.go` - check imports

## Implementation Order

1. **Add GetID() to statechart.ChartDefinition**
   - File: `pkg/statechart/types.go`
   - Add method after line ~178 (end of ChartDefinition struct)
   - Test: `go test ./pkg/statechart -run TestChartDefinition_GetID`

2. **Update pkg/kernel imports**
   - File: `pkg/kernel/kernel.go`
   - Replace: `github.com/maelstrom/v3/pkg/chart` → `github.com/maelstrom/v3/pkg/statechart`
   - Replace: `chart.ChartDefinition` → `statechart.ChartDefinition`
   - Test: `go test ./pkg/kernel`

3. **Update pkg/kernel tests**
   - File: `pkg/kernel/kernel_test.go`
   - Replace: `github.com/maelstrom/v3/pkg/chart` → `github.com/maelstrom/v3/pkg/statechart`
   - Replace: `chart.ChartDefinition` → `statechart.ChartDefinition`
   - Test: `go test ./pkg/kernel`

4. **Update pkg/bootstrap imports**
   - File: `pkg/bootstrap/chart.go`
   - Replace: `github.com/maelstrom/v3/pkg/chart` → `github.com/maelstrom/v3/pkg/statechart`
   - Replace: `chart.ChartDefinition` → `statechart.ChartDefinition`
   - Test: `go test ./pkg/bootstrap`

5. **Update pkg/bootstrap tests**
   - File: `pkg/bootstrap/chart_test.go`
   - Replace: `github.com/maelstrom/v3/pkg/chart` → `github.com/maelstrom/v3/pkg/statechart`
   - Replace: `chart.ChartDefinition` → `statechart.ChartDefinition`
   - Test: `go test ./pkg/bootstrap`

6. **Delete pkg/chart package**
   - Delete: `pkg/chart/registry.go`
   - Delete: `pkg/chart/hydrator.go`
   - Delete: `pkg/chart/registry_test.go`
   - Delete: `pkg/chart/hydrator_test.go`
   - Verify: `go test ./...` runs successfully

7. **Run full test suite**
   - Command: `go test ./...`
   - Target: All 118+ tests pass
   - Verify: No references to `pkg/chart.ChartDefinition` remain

8. **Update documentation**
   - Update API docs to reference `pkg/statechart.ChartDefinition`
   - Update usage examples to show correct import path
   - Update integration guide references

## Test Plan

### Test 1: GetID() Method
```go
func TestChartDefinition_GetID(t *testing.T) {
    def := statechart.ChartDefinition{
        ID:      "test-chart",
        Version: "1.0.0",
    }
    if def.GetID() != "test-chart" {
        t.Errorf("GetID() = %q, want %q", def.GetID(), "test-chart")
    }
}
```

### Test 2: Kernel Uses statechart.ChartDefinition
```go
func TestKernel_UsesStatechartChartDefinition(t *testing.T) {
    // Verify kernel can use statechart.ChartDefinition without errors
    // Check that LoadBootstrapChart returns statechart.ChartDefinition
}
```

### Test 3: Bootstrap Uses statechart.ChartDefinition
```go
func TestBootstrap_LoadChart_UsesStatechart(t *testing.T) {
    // Verify bootstrap.LoadBootstrapChart works with statechart.ChartDefinition
    // Ensure no pkg/chart imports remain
}
```

### Test 4: Full Test Suite Passes
```bash
go test ./...
# All 118+ tests should pass
```

### Test 5: No pkg/chart References Remain
```bash
grep -r "pkg/chart" .
# Should only find documentation, not code
```

## Architecture Decisions

1. **Single Source of Truth**: `pkg/statechart.ChartDefinition` becomes the authoritative type
2. **API Consistency**: All ChartDefinitions now have `GetID()` method
3. **Backward Compatibility**: Factory already uses `statechart.ChartDefinition`, so no breaking changes
4. **Test Coverage**: Existing tests validate the transition works correctly

## Files to Review Before Implementation

- `pkg/statechart/types.go` - Confirm ChartDefinition struct location
- `pkg/kernel/kernel.go` - Confirm current imports and usage
- `pkg/kernel/kernel_test.go` - Confirm test imports
- `pkg/bootstrap/chart.go` - Confirm current imports and usage
- `pkg/bootstrap/chart_test.go` - Confirm test imports
- `pkg/chart/registry.go` - Confirm it's only used by kernel/bootstrap
- `pkg/chart/hydrator.go` - Confirm it's only used by kernel/bootstrap

## Success Criteria

1. ✅ `GetID()` method exists on `statechart.ChartDefinition`
2. ✅ All imports of `pkg/chart` replaced with `pkg/statechart`
3. ✅ `pkg/chart` package deleted
4. ✅ All 118+ tests pass
5. ✅ No code references to `pkg/chart.ChartDefinition` remain
6. ✅ Documentation updated to reflect new structure

## Potential Issues & Mitigations

1. **Issue**: Import path inconsistency (v3 vs v4)
   - **Mitigation**: Use consistent import paths as specified in existing files

2. **Issue**: Tests may fail during transition
   - **Mitigation**: Update imports incrementally, test after each change

3. **Issue**: Documentation references may be outdated
   - **Mitigation**: Update all docs after code changes complete

## Verification Commands

```bash
# Add GetID() method
go test ./pkg/statechart -run TestChartDefinition_GetID

# Test kernel package
go test ./pkg/kernel

# Test bootstrap package
go test ./pkg/bootstrap

# Delete chart package
rm pkg/chart/registry.go pkg/chart/hydrator.go pkg/chart/registry_test.go pkg/chart/hydrator_test.go

# Full test suite
go test ./...

# Verify no chart references remain
grep -r "pkg/chart" . --exclude-dir=.git --exclude="*.md"
```

## Implementation Notes

1. **Order matters**: Add `GetID()` first, then update imports, then delete old package
2. **Test after each step**: Don't wait until all changes are made
3. **Commit incrementally**: Each successful step gets its own commit
4. **Documentation updates**: Do these last, after all code changes pass tests