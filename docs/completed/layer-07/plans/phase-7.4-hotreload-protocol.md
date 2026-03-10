# Phase 7.4: Hot-Reload Protocol Integration

## Goal
Implement full hot-reload protocol with quiescence detection and version migration (arch-v1.md L865-882).

## Scope
- Implement quiescence detection (empty queue, no active regions, no inflight tools)
- Implement hot-reload protocol (prepareForReload, quiescence wait, force-stop)
- Implement history mechanisms (shallowHistory, deepHistory, cleanStart)
- Implement context transform validation
- 7 tests, 7 commits
- **Branch**: `feat/layer7-hotreload-protocol`

## Current Status

### Implementation Status
| Component | Status | Notes |
|-----------|--------|-------|
| `Quiescence detection` | ❌ Missing | Need to implement queue/region/tool checks |
| `Hot-reload protocol` | ❌ Missing | Need prepareForReload, timeout, force-stop |
| `History mechanisms` | ❌ Missing | Need shallow/deep history, cleanStart |
| `Context transform` | ❌ Missing | Need Go template validation |

### Files Status
| File | Status |
|------|--------|
| `pkg/runtime/runtime.go` | ⚠️ Partial - add quiescence detection |
| `pkg/registry/service.go` | ⚠️ Partial - add hot-reload protocol |

## Required Implementation

### Quiescence Detection (arch-v1.md L860-861)
```go
// pkg/runtime/runtime.go
func (r *ChartRuntime) IsQuiescent() bool {
    // Event queue empty AND no active parallel regions AND no inflight tools
}
```

### Hot-Reload Protocol (arch-v1.md L865-880)
```go
// pkg/registry/service.go
func (s *Service) hotReload(chartName string, newDef ChartDefinition) error {
    // 1. Signal prepareForReload (L867)
    // 2. Wait for quiescence (with timeout) (L868)
    // 3. IF quiescent: stop, spawn with history (L869-872)
    // 4. IF timeout: force-stop, cleanStart (L873-876)
    // 5. IF maxWaitAttempts exceeded: log failure, require admin intervention (L877-879)
}
```

## Tests to Write (7 tests, 7 commits)

### Test 1: TestHotReloadProtocol_QuiescenceDetection
```go
func TestHotReloadProtocol_QuiescenceDetection(t *testing.T)
```
**Acceptance Criteria:**
- `IsQuiescent()` returns true when event queue empty (arch-v1.md L860)
- `IsQuiescent()` returns true when no active parallel regions (arch-v1.md L860)
- `IsQuiescent()` returns true when no inflight tool calls (arch-v1.md L861)

### Test 2: TestHotReloadProtocol_ProtocolFlow
```go
func TestHotReloadProtocol_ProtocolFlow(t *testing.T)
```
**Acceptance Criteria:**
- Hot-reload signals `prepareForReload` to current runtime (arch-v1.md L867)
- Runtime attempts to reach quiescence within timeoutMs (arch-v1.md L868)

### Test 3: TestHotReloadProtocol_QuiescenceReached
```go
func TestHotReloadProtocol_QuiescenceReached(t *testing.T)
```
**Acceptance Criteria:**
- IF quiescence reached: stop current runtime (arch-v1.md L869-872)
- Spawn new runtime with history (shallow/deep)
- Apply contextTransform if version changed (arch-v1.md L872)

### Test 4: TestHotReloadProtocol_TimeoutForceStop
```go
func TestHotReloadProtocol_TimeoutForceStop(t *testing.T)
```
**Acceptance Criteria:**
- IF timeout expires: force-stop current runtime (arch-v1.md L873-876)
- cleanStart with no history preserved (arch-v1.md L875)
- Increment reload attempt counter (arch-v1.md L876)

### Test 5: TestHotReloadProtocol_MaxAttemptsExceeded
```go
func TestHotReloadProtocol_MaxAttemptsExceeded(t *testing.T)
```
**Acceptance Criteria:**
- IF maxWaitAttempts exceeded: log permanent failure to sys:observability (arch-v1.md L877-879)
- Require manual intervention via sys:admin (arch-v1.md L879)

### Test 6: TestHotReloadProtocol_HistoryMechanisms
```go
func TestHotReloadProtocol_HistoryMechanisms(t *testing.T)
```
**Acceptance Criteria:**
- `shallowHistory`: restore to parent state's default sub-state (arch-v1.md L883)
- `deepHistory`: restore to specific sub-state (if still exists) (arch-v1.md L884)
- Deleted state fallback: fall back to shallow if deepHistory target deleted (arch-v1.md L885)

### Test 7: TestHotReloadProtocol_ContextTransformValidation
```go
func TestHotReloadProtocol_ContextTransformValidation(t *testing.T)
```
**Acceptance Criteria:**
- Registry validates Go template syntax at load time (arch-v1.md L890)
- Context transform executed on version change (arch-v1.md L889)
- Failure → cleanStart fallback (arch-v1.md L889)

## Dependencies

### Test Dependencies
```
Test 1 (Quiescence detection) → Independent
Test 2 (Protocol flow) → Depends on Test 1
Test 3 (Quiescence reached) → Depends on Test 2
Test 4 (Timeout force-stop) → Depends on Test 2
Test 5 (Max attempts exceeded) → Depends on Test 4
Test 6 (History mechanisms) → Depends on Test 3
Test 7 (Context transform) → Depends on Test 3
```

### Phase Dependencies
- **Phase 7.3** must complete before **Phase 7.4** (needs post-KERNEL_READY loading)
- **Phase 7.7** depends on **Phase 7.4** (integration tests need hot-reload)

## Files to Create/Modify

| File | Operation | Changes |
|------|-----------|---------|
| `pkg/runtime/runtime.go` | MODIFY | Add IsQuiescent() method |
| `pkg/registry/service.go` | MODIFY | Add hot-reload protocol implementation |

## Implementation Order (TDD)

1. **Test 1**: Write test → verify RED → Implement IsQuiescent() → verify GREEN → commit
2. **Test 2**: Write test → verify RED → Implement prepareForReload signaling → verify GREEN → commit
3. **Test 3**: Write test → verify RED → Implement quiescence path → verify GREEN → commit
4. **Test 4**: Write test → verify RED → Implement timeout force-stop → verify GREEN → commit
5. **Test 5**: Write test → verify RED → Implement max attempts handling → verify GREEN → commit
6. **Test 6**: Write test → verify RED → Implement history mechanisms → verify GREEN → commit
7. **Test 7**: Write test → verify RED → Implement context transform validation → verify GREEN → commit

**Total: 7 tests, 7 commits**

## Deliverables
- ✅ Quiescence detection implemented (arch-v1.md L860-861)
- ✅ Hot-reload protocol implemented (arch-v1.md L865-880)
- ✅ History mechanisms implemented (shallow/deep/cleanStart) (arch-v1.md L882-885)
- ✅ Context transform validation implemented (arch-v1.md L887-890)
- ✅ 7 commits (one per test per TDD workflow)