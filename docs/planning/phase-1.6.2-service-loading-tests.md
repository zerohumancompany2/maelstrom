# Phase 1.6.2: Service Loading Tests

## Goal
Verify services load in correct order and emit ready events using channel-based synchronization.

## Scope
- Add state history tracking to Sequence
- Add event history tracking to Sequence
- Add KERNEL_READY tracking to Sequence
- Add completion tracking to Kernel
- Implement 4 tests with channel-based synchronization

## Current Status

### Implementation Status
| Test | Status | Notes |
|------|--------|-------|
| Test 4: ServicesLoadInOrder | ❌ Missing | Sequence doesn't track state history |
| Test 5: AllServicesEmitReadyEvents | ❌ Missing | Sequence doesn't track events |
| Test 6: KernelReadyEventEmitted | ❌ Missing | No KERNEL_READY tracking |
| Test 7: OnCompleteCallbackInvoked | ❌ Missing | No completion tracking |

### Files Status
| File | Status |
|------|--------|
| `pkg/bootstrap/sequence.go` | ⚠️ Partial - needs tracking fields and getters |
| `pkg/kernel/kernel.go` | ⚠️ Partial - needs completion tracking |
| `pkg/kernel/kernel_test.go` | ❌ Missing - needs 4 new tests |

## Required Implementation

### Sequence State Tracking
```go
// pkg/bootstrap/sequence.go
type Sequence struct {
    // ... existing fields
    statesEntered   []string
    statesMu        sync.RWMutex
}

func (s *Sequence) GetStatesEntered() []string {
    s.statesMu.RLock()
    defer s.statesMu.RUnlock()
    copy := make([]string, len(s.statesEntered))
    copy(copy, s.statesEntered)
    return copy
}

// In OnStateEnter callback:
s.statesMu.Lock()
s.statesEntered = append(s.statesEntered, state)
s.statesMu.Unlock()
```

### Sequence Event Tracking
```go
// pkg/bootstrap/sequence.go
type Sequence struct {
    // ... existing fields
    eventsHandled   []string
    eventsMu        sync.RWMutex
}

func (s *Sequence) GetEventsHandled() []string {
    s.eventsMu.RLock()
    defer s.eventsMu.RUnlock()
    copy := make([]string, len(s.eventsHandled))
    copy(copy, s.eventsHandled)
    return copy
}

// In HandleEvent():
s.eventsMu.Lock()
s.eventsHandled = append(s.eventsHandled, event)
s.eventsMu.Unlock()
```

### KERNEL_READY Tracking
```go
// pkg/bootstrap/sequence.go
type Sequence struct {
    // ... existing fields
    kernelReadyEmitted atomic.Bool
}

func (s *Sequence) GetKernelReadyEmitted() bool {
    return s.kernelReadyEmitted.Load()
}

// In HandleEvent():
if event == "KERNEL_READY" {
    s.kernelReadyEmitted.Store(true)
}
```

### Kernel Completion Tracking
```go
// pkg/kernel/kernel.go
type Kernel struct {
    // ... existing fields
    onCompleteCalled atomic.Bool
}

func (k *Kernel) GetCompletionStatus() bool {
    return k.onCompleteCalled.Load()
}

func (k *Kernel) onBootstrapComplete() {
    log.Println("[kernel] Kernel going dormant")
    k.onCompleteCalled.Store(true)
}
```

## Tests to Write (4 tests, 4 commits)

### Test 4: ServicesLoadInOrder
```go
func TestKernel_ServicesLoadInOrder()
```
**Acceptance Criteria:**
1. Expected order: `["security", "communication", "observability", "lifecycle", "handoff", "complete"]`
2. `GetStatesEntered()` returns states in correct order
3. All 6 states are entered

**Synchronization:** Polling with timeout (state transitions are synchronous)

### Test 5: AllServicesEmitReadyEvents
```go
func TestKernel_AllServicesEmitReadyEvents()
```
**Acceptance Criteria:**
1. Expected events: `["SECURITY_READY", "COMMUNICATION_READY", "OBSERVABILITY_READY", "LIFECYCLE_READY"]`
2. `GetEventsHandled()` contains all 4 events
3. Events received in correct order

**Synchronization:** Channel per event (events are async)

### Test 6: KernelReadyEventEmitted
```go
func TestKernel_KernelReadyEventEmitted()
```
**Acceptance Criteria:**
1. `GetKernelReadyEmitted()` returns `true` after all services ready
2. KERNEL_READY emitted after LIFECYCLE_READY

**Synchronization:** Single channel with timeout

### Test 7: OnCompleteCallbackInvoked
```go
func TestKernel_OnCompleteCallbackInvoked()
```
**Acceptance Criteria:**
1. `onComplete` callback is invoked after bootstrap completes
2. `GetCompletionStatus()` returns `true`
3. `IsBootstrapComplete()` returns `true`

**Synchronization:** Channel on callback

## Dependencies

### Test Dependencies
```
Test 4 (ServicesLoadInOrder) → No dependencies
    ↓
Test 5 (AllServicesEmitReadyEvents) → Needs state tracking
    ↓
Test 6 (KernelReadyEventEmitted) → Needs event tracking
    ↓
Test 7 (OnCompleteCallbackInvoked) → Needs completion tracking
```

### Phase Dependencies
- **Phase 1.6.1** - Observability infrastructure must be in place
- **Phase 1.4** - Bootstrap actions must be implemented

## Files to Create/Modify

| File | Operation | Changes |
|------|-----------|---------|
| `pkg/bootstrap/sequence.go` | MODIFY | Add `statesEntered`, `eventsHandled`, `kernelReadyEmitted` tracking + getters |
| `pkg/kernel/kernel.go` | MODIFY | Add `onCompleteCalled` tracking + getter |
| `pkg/kernel/kernel_test.go` | MODIFY | Add 4 new tests with channel-based synchronization |

## Implementation Order (TDD)

1. **Test 4**: Write test → verify RED → Add state tracking to Sequence → Add getter → verify GREEN → commit
2. **Test 5**: Write test → verify RED → Add event tracking to Sequence → Add getter → verify GREEN → commit
3. **Test 6**: Write test → verify RED → Add KERNEL_READY tracking → Add getter → verify GREEN → commit
4. **Test 7**: Write test → verify RED → Add completion tracking to Kernel → Add getter → verify GREEN → commit

**Total: 4 tests, 4 commits**

## Synchronization Mechanisms

| Test | Mechanism | Rationale |
|------|-----------|-----------|
| Test 4 | **Polling with timeout** | State transitions are synchronous; simple polling is sufficient |
| Test 5 | **Channel per event** | Events are async; channels provide clean signaling |
| Test 6 | **Single channel** | Single event; channel avoids polling |
| Test 7 | **Channel on callback** | Callback is synchronous; channel provides clean test signal |

**Key Principle:** Avoid `time.Sleep` - use channels or context-based waiting for deterministic behavior.

## Deliverables
- Modified `pkg/bootstrap/sequence.go` with state and event tracking
- Modified `pkg/kernel/kernel.go` with completion tracking
- Updated `pkg/kernel/kernel_test.go` with 4 passing tests
- 4 commits (one per test per TDD workflow)