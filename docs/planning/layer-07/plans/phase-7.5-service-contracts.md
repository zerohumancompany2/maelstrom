# Phase 7.5: Service-Specific Contracts

## Goal
Complete service-specific contracts for adminCommand 2FA, tool descriptors, and other service-specific APIs (arch-v1.md L483-491).

## Scope
- sys:admin adminCommand with 2FA (L485)
- sys:persistence snapshot/restore (L486)
- sys:observability emit/query (L487)
- sys:tools resolve with ToolDescriptor (L488)
- sys:memory query with boundaryFilter (L489)
- sys:datasources tagOnWrite/getTaints (L490)
- 8 tests, 8 commits
- **Branch**: `feat/layer7-service-contracts`

## Current Status

### Implementation Status
| Contract | Status | Notes |
|----------|--------|-------|
| `adminCommand` | ⚠️ Partial | 2FA exists; needs full command handling |
| `snapshot/restore` | ⚠️ Partial | Storage exists; needs full API |
| `emit/query` | ⚠️ Partial | Trace storage exists; needs query API |
| `resolve` | ⚠️ Partial | Registry exists; needs full ToolDescriptor |
| `memory query` | ⚠️ Partial | Vector/graph exists; needs boundaryFilter |
| `tagOnWrite/getTaints` | ⚠️ Partial | Storage exists; needs taint API |

### Files Status
| File | Status |
|------|--------|
| `pkg/services/admin/service.go` | ⚠️ Partial - complete adminCommand API |
| `pkg/services/persistence/service.go` | ⚠️ Partial - complete snapshot/restore API |
| `pkg/services/tools/service.go` | ⚠️ Partial - complete ToolDescriptor resolution |

## Required Implementation

### sys:admin adminCommand (arch-v1.md L485)
```go
// pkg/services/admin/service.go
func (s *adminService) AdminCommand(cmd string, authToken string) error {
    // Verify 2FA token, execute command: list|control|queryTaints|inject
}
```

### sys:persistence snapshot/restore (arch-v1.md L486)
```go
// pkg/services/persistence/service.go
func (s *persistenceService) Snapshot(runtimeId string) error {
    // Create snapshot of runtime state
}

func (s *persistenceService) Restore(snapshotId string) error {
    // Restore runtime from snapshot
}
```

### sys:observability emit/query (arch-v1.md L487)
```go
// pkg/services/observability/service.go
func (s *observabilityService) Emit(trace Trace) {
    // Store trace entry
}

func (s *observabilityService) Query(filters QueryFilters) []Trace {
    // Return filtered traces
}
```

### sys:tools resolve (arch-v1.md L488)
```go
// pkg/services/tools/service.go
func (s *toolsService) Resolve(name string, callerBoundary string) (ToolDescriptor, error) {
    // Return ToolDescriptor or notFound error
}
```

### sys:memory query (arch-v1.md L489)
```go
// pkg/services/memory/service.go
func (s *memoryService) Query(vector []float64, topK int, boundaryFilter string) []Message {
    // Return MessageSlice respecting boundaryFilter
}
```

### sys:datasources tagOnWrite/getTaints (arch-v1.md L490)
```go
// pkg/services/datasources/service.go
func (s *datasourcesService) TagOnWrite(path string, taints []string) error {
    // Tag path with taints
}

func (s *datasourcesService) GetTaints(path string) []string {
    // Return taints for path
}
```

## Tests to Write (8 tests, 8 commits)

### Test 1: TestServiceContracts_AdminCommand2FA
```go
func TestServiceContracts_AdminCommand2FA(t *testing.T)
```
**Acceptance Criteria:**
- `adminCommand(cmd, authToken)` requires valid 2FA token (arch-v1.md L485)
- Commands: list, control, queryTaints, inject (arch-v1.md L485)
- 2FA-gated access enforced

### Test 2: TestServiceContracts_PersistenceSnapshotRestore
```go
func TestServiceContracts_PersistenceSnapshotRestore(t *testing.T)
```
**Acceptance Criteria:**
- `snapshot(runtimeId)` creates snapshot (arch-v1.md L486)
- `restore(snapshotId)` restores runtime (arch-v1.md L486)

### Test 3: TestServiceContracts_ObservabilityEmitQuery
```go
func TestServiceContracts_ObservabilityEmitQuery(t *testing.T)
```
**Acceptance Criteria:**
- `emit(trace)` stores trace (arch-v1.md L487)
- `query(traces, filters)` returns filtered traces (arch-v1.md L487)

### Test 4: TestServiceContracts_ToolsResolve
```go
func TestServiceContracts_ToolsResolve(t *testing.T)
```
**Acceptance Criteria:**
- `resolve(name, callerBoundary)` returns ToolDescriptor or notFound (arch-v1.md L488)
- Boundary accessibility enforced (arch-v1.md L488)

### Test 5: TestServiceContracts_MemoryQuery
```go
func TestServiceContracts_MemoryQuery(t *testing.T)
```
**Acceptance Criteria:**
- `query(vector, topK, boundaryFilter)` returns MessageSlice (arch-v1.md L489)
- boundaryFilter respected in results (arch-v1.md L489)

### Test 6: TestServiceContracts_DatasourcesTagOnWrite
```go
func TestServiceContracts_DatasourcesTagOnWrite(t *testing.T)
```
**Acceptance Criteria:**
- `tagOnWrite(path, taints)` tags path with taints (arch-v1.md L490)

### Test 7: TestServiceContracts_DatasourcesGetTaints
```go
func TestServiceContracts_DatasourcesGetTaints(t *testing.T)
```
**Acceptance Criteria:**
- `getTaints(path)` returns []string of taints (arch-v1.md L490)

### Test 8: TestServiceContracts_AllContractsImplemented
```go
func TestServiceContracts_AllContractsImplemented(t *testing.T)
```
**Acceptance Criteria:**
- All service-specific contracts from arch-v1.md L483-491 are implemented
- Each service exposes its contract methods

## Dependencies

### Test Dependencies
```
Test 1 (Admin contract) → Independent
Test 2 (Persistence contract) → Independent
Test 3 (Observability contract) → Independent
Test 4 (Tools contract) → Independent
Test 5 (Memory contract) → Independent
Test 6 (Datasources tagOnWrite) → Independent
Test 7 (Datasources getTaints) → Depends on Test 6
Test 8 (All contracts) → Depends on Tests 1-7
```

### Phase Dependencies
- **Phase 7.2** must complete before **Phase 7.5** (needs hot-reloadable services)
- **Phase 7.6** depends on **Phase 7.5** (discovery needs contract compliance)

## Files to Create/Modify

| File | Operation | Changes |
|------|-----------|---------|
| `pkg/services/admin/service.go` | MODIFY | Complete adminCommand API |
| `pkg/services/persistence/service.go` | MODIFY | Complete snapshot/restore API |
| `pkg/services/tools/service.go` | MODIFY | Complete ToolDescriptor resolution |

## Implementation Order (TDD)

1. **Test 1**: Write test → verify RED → Implement adminCommand 2FA → verify GREEN → commit
2. **Test 2**: Write test → verify RED → Implement snapshot/restore → verify GREEN → commit
3. **Test 3**: Write test → verify RED → Implement emit/query → verify GREEN → commit
4. **Test 4**: Write test → verify RED → Implement resolve ToolDescriptor → verify GREEN → commit
5. **Test 5**: Write test → verify RED → Implement memory query with boundaryFilter → verify GREEN → commit
6. **Test 6**: Write test → verify RED → Implement tagOnWrite → verify GREEN → commit
7. **Test 7**: Write test → verify RED → Implement getTaints → verify GREEN → commit
8. **Test 8**: Write test → verify RED → Verify all contracts → verify GREEN → commit

**Total: 8 tests, 8 commits**

## Deliverables
- ✅ adminCommand 2FA-gated API complete (arch-v1.md L485)
- ✅ snapshot/restore API complete (arch-v1.md L486)
- ✅ emit/query API complete (arch-v1.md L487)
- ✅ resolve ToolDescriptor API complete (arch-v1.md L488)
- ✅ memory query with boundaryFilter complete (arch-v1.md L489)
- ✅ tagOnWrite/getTaints API complete (arch-v1.md L490)
- ✅ 8 commits (one per test per TDD workflow)

## Recursive Analysis Decision

**Decision: Keep as-is (no split)**

**Reasoning:**
- 8 tests is within acceptable range (2-10 per CLAUDE.md)
- Tests cover distinct service contracts (admin, persistence, observability, tools, memory, datasources)
- Single coherent feature: service-specific API contracts
- Splitting would fragment the contract completion goal

**Alternative (if split needed):**
- 7.5a: Admin + Persistence contracts - 3 tests
- 7.5b: Observability + Tools contracts - 3 tests
- 7.5c: Memory + Datasources contracts - 2 tests