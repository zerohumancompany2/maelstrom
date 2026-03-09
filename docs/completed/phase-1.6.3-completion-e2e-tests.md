# Phase 1.6.3: Completion and E2E Tests

## Goal
Verify complete bootstrap flow with comprehensive end-to-end testing using channel-based synchronization.

## Scope
- Add log capture mechanism to Kernel
- Implement channel-based completion waiting
- Implement comprehensive E2E test combining all observability
- Implement 3 tests with proper synchronization

## Current Status

### Implementation Status
| Test | Status | Notes |
|------|--------|-------|
| Test 8: BootstrapCompletes | ⚠️ Partial | `IsBootstrapComplete()` exists but uses `time.Sleep` |
| Test 9: GoesDormant | ❌ Missing | No log capture mechanism |
| Test 10: FullE2EBootstrap | ❌ Missing | All components exist but not integrated |

### Files Status
| File | Status |
|------|--------|
| `pkg/kernel/kernel.go` | ⚠️ Partial - needs log capture |
| `pkg/kernel/kernel_test.go` | ❌ Missing - needs 3 new tests |

## Required Implementation

### Kernel Log Capture
```go
// pkg/kernel/kernel.go
type Kernel struct {
    // ... existing fields
    logOutput       []string
    logMu           sync.RWMutex
}

func (k *Kernel) CaptureLog(msg string) {
    k.logMu.Lock()
    defer k.logMu.Unlock()
    k.logOutput = append(k.logOutput, msg)
}

func (k *Kernel) GetLogOutput() []string {
    k.logMu.RLock()
    defer k.logMu.RUnlock()
    copy := make([]string, len(k.logOutput))
    copy(copy, k.logOutput)
    return copy
}

func (k *Kernel) onBootstrapComplete() {
    msg := "[kernel] Kernel going dormant"
    k.CaptureLog(msg)
    log.Println(msg)
    k.onCompleteCalled.Store(true)
}
```

### Channel-Based Completion Waiting
```go
// Helper function for tests
func waitForBootstrapComplete(t *testing.T, kernel *kernel.Kernel, timeout time.Duration) {
    t.Helper()
    ctx, cancel := context.WithTimeout(context.Background(), timeout)
    defer cancel()
    
    done := make(chan bool)
    go func() {
        for {
            select {
            case <-ctx.Done():
                return
            default:
            }
            if kernel.IsBootstrapComplete() {
                done <- true
                return
            }
            time.Sleep(10 * time.Millisecond)
        }
    }()
    
    select {
    case <-done:
        // Success
    case <-ctx.Done():
        t.Fatal("Bootstrap did not complete within timeout")
    }
}
```

## Tests to Write (3 tests, 3 commits)

### Test 8: BootstrapCompletes
```go
func TestKernel_BootstrapCompletes()
```
**Acceptance Criteria:**
1. `IsBootstrapComplete()` returns `true` after bootstrap completes
2. Completion occurs within timeout (2 seconds)
3. Uses channel-based waiting, not `time.Sleep`

**Synchronization:** Channel + context with timeout

### Test 9: GoesDormant
```go
func TestKernel_GoesDormant()
```
**Acceptance Criteria:**
1. Kernel logs "going dormant" after bootstrap completes
2. Log message is captured and retrievable
3. Log appears after completion

**Synchronization:** Channel on callback (log capture is synchronous)

### Test 10: FullE2EBootstrap
```go
func TestKernel_FullE2EBootstrap()
```
**Acceptance Criteria:**
1. States enter in correct order: `["security", "communication", "observability", "lifecycle", "handoff", "complete"]`
2. All 4 READY events received: `["SECURITY_READY", "COMMUNICATION_READY", "OBSERVABILITY_READY", "LIFECYCLE_READY"]`
3. KERNEL_READY event emitted
4. `onComplete` callback invoked
5. `IsBootstrapComplete()` returns `true`
6. "going dormant" logged

**Synchronization:** Channels + context for overall coordination

## Dependencies

### Test Dependencies
```
Test 8 (BootstrapCompletes) → Needs completion tracking from 1.6.2
    ↓
Test 9 (GoesDormant) → Needs log capture
    ↓
Test 10 (FullE2EBootstrap) → Needs ALL observability from 1.6.1 and 1.6.2
```

### Phase Dependencies
- **Phase 1.6.1** - Bootstrap runtime ID and state tracking
- **Phase 1.6.2** - State history, event tracking, completion tracking
- **Phase 1.4** - Bootstrap actions must be fully implemented
- **Phase 1.3** - All 4 core services must have stubs

## Files to Create/Modify

| File | Operation | Changes |
|------|-----------|---------|
| `pkg/kernel/kernel.go` | MODIFY | Add log capture mechanism (`logOutput`, `CaptureLog()`, `GetLogOutput()`) |
| `pkg/kernel/kernel_test.go` | MODIFY | Add 3 new tests with channel-based synchronization |

## Implementation Order (TDD)

1. **Test 8**: Write test → verify RED → Implement channel-based completion waiting → verify GREEN → commit
2. **Test 9**: Write test → verify RED → Add log capture to Kernel → Update `onBootstrapComplete()` → verify GREEN → commit
3. **Test 10**: Write test → verify RED → Combine all observability → Implement comprehensive E2E test → verify GREEN → commit

**Total: 3 tests, 3 commits**

## Synchronization Mechanisms

| Test | Mechanism | Rationale |
|------|-----------|-----------|
| Test 8 | **Channel + context** | Existing `IsBootstrapComplete()` can be polled with channel |
| Test 9 | **Channel on callback** | Log capture is synchronous with callback |
| Test 10 | **Channels + context** | E2E test needs coordination of all signals |

**Key Principle:** All tests use channel-based synchronization - NO `time.Sleep` patterns.

## Helper Functions

### assertStatesInOrder
```go
func assertStatesInOrder(t *testing.T, actual []string) {
    t.Helper()
    expected := []string{"security", "communication", "observability", "lifecycle", "handoff", "complete"}
    assert.Equal(t, expected, actual)
}
```

### assertAllEventsReceived
```go
func assertAllEventsReceived(t *testing.T, actual []string) {
    t.Helper()
    expectedEvents := []string{"SECURITY_READY", "COMMUNICATION_READY", "OBSERVABILITY_READY", "LIFECYCLE_READY"}
    for _, expected := range expectedEvents {
        found := false
        for _, actualEvent := range actual {
            if actualEvent == expected {
                found = true
                break
            }
        }
        assert.True(t, found, "Expected event %s not found", expected)
    }
}
```

### assertDormantLogged
```go
func assertDormantLogged(t *testing.T, logs []string) {
    t.Helper()
    found := false
    for _, log := range logs {
        if strings.Contains(log, "going dormant") {
            found = true
            break
        }
    }
    assert.True(t, found, "Expected 'going dormant' in logs")
}
```

## Deliverables
- Modified `pkg/kernel/kernel.go` with log capture mechanism
- Updated `pkg/kernel/kernel_test.go` with 3 passing tests
- 3 commits (one per test per TDD workflow)
- Complete E2E bootstrap verification