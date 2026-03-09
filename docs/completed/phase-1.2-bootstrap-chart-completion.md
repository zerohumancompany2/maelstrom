# Phase 1.2: Bootstrap Chart Completion - TDD Implementation Plan

**Branch**: `feat/bootstrap-chart-completion`  
**Parent Spec**: `docs/planning/layer-01-kernel-bootstrap.md` Section 3  
**Status**: Ready for TDD execution

---

## Analysis Summary

### Current State vs Spec Requirements

| Aspect | Current | Required by Spec |
|--------|---------|------------------|
| Chart type | (default) | `type: compound` |
| Entry actions | Simple strings | Structured with `action` + `params` |
| Failure handling | None | `failed` state with transitions from each state |
| Event naming | UPPERCASE (SECURITY_READY) | lowercase (securityReady, securityFailed) |
| Final states | `complete` | `ready` and `failed` |
| Action names | loadSecurityService | securityBootstrap, communicationBootstrap, etc. |

---

## TDD Test Sequence

### Test 1: `TestBootstrapChart_HasCompoundType`

**What it verifies:** Chart has `type: compound` in spec  
**Expected failure:** Chart currently lacks type field, or has wrong type  
**YAML change needed:**

```yaml
spec:
  type: compound
  initial: security
```

**Minimal implementation:** Add `type: compound` to chart YAML

---

### Test 2: `TestBootstrapChart_HasFailedState`

**What it verifies:** `failed` state exists as a final state  
**Expected failure:** No `failed` state in current YAML  
**YAML change needed:**

```yaml
  failed:
    type: final
    entryActions:
      - action: logFailure
      - action: panic
```

**Minimal implementation:** Add failed state definition

---

### Test 3: `TestBootstrapChart_SecurityHasStructuredEntryAction`

**What it verifies:** Security state has structured entry action with params  
**Expected failure:** Current entryActions is `loadSecurityService` (simple string)  
**YAML change needed:**

```yaml
  security:
    type: atomic
    entryActions:
      - action: securityBootstrap
        params:
          boundaries: [inner, dmz, outer]
```

**Minimal implementation:** Restructure security state entryActions

---

### Test 4: `TestBootstrapChart_SecurityHasFailureTransition`

**What it verifies:** Security state has transition to `failed` on `securityFailed` event  
**Expected failure:** Only `SECURITY_READY` -> `communication` transition exists  
**YAML change needed:**

```yaml
    transitions:
      - event: securityReady
        target: communication
      - event: securityFailed
        target: failed
```

**Minimal implementation:** Add failure transition to security state

---

### Test 5: `TestBootstrapChart_CommunicationHasStructuredEntryAction`

**What it verifies:** Communication state has structured entry action with params  
**Expected failure:** Current entryActions is `loadCommunicationService` (simple string)  
**YAML change needed:**

```yaml
  communication:
    type: atomic
    entryActions:
      - action: communicationBootstrap
        params:
          mailBackbone: true
```

**Minimal implementation:** Restructure communication state entryActions

---

### Test 6: `TestBootstrapChart_CommunicationHasFailureTransition`

**What it verifies:** Communication state has transition to `failed` on `communicationFailed` event  
**Expected failure:** Only `COMMUNICATION_READY` -> `observability` transition exists  
**YAML change needed:**

```yaml
    transitions:
      - event: communicationReady
        target: observability
      - event: communicationFailed
        target: failed
```

**Minimal implementation:** Add failure transition to communication state

---

### Test 7: `TestBootstrapChart_ObservabilityHasStructuredEntryAction`

**What it verifies:** Observability state has structured entry action with params  
**Expected failure:** Current entryActions is `loadObservabilityService` (simple string)  
**YAML change needed:**

```yaml
  observability:
    type: atomic
    entryActions:
      - action: observabilityBootstrap
        params:
          tracing: true
          metrics: true
          deadLetterQueue: true
```

**Minimal implementation:** Restructure observability state entryActions

---

### Test 8: `TestBootstrapChart_ObservabilityHasFailureTransition`

**What it verifies:** Observability state has transition to `failed` on `observabilityFailed` event  
**Expected failure:** Only `OBSERVABILITY_READY` -> `lifecycle` transition exists  
**YAML change needed:**

```yaml
    transitions:
      - event: observabilityReady
        target: lifecycle
      - event: observabilityFailed
        target: failed
```

**Minimal implementation:** Add failure transition to observability state

---

### Test 9: `TestBootstrapChart_LifecycleHasStructuredEntryAction`

**What it verifies:** Lifecycle state has structured entry action with params  
**Expected failure:** Current entryActions is `loadLifecycleService` (simple string)  
**YAML change needed:**

```yaml
  lifecycle:
    type: atomic
    entryActions:
      - action: lifecycleBootstrap
        params:
          enableSpawn: true
          enableStop: true
          toolRegistry: true
```

**Minimal implementation:** Restructure lifecycle state entryActions

---

### Test 10: `TestBootstrapChart_LifecycleHasReadyTransition`

**What it verifies:** Lifecycle transitions to `ready` (not `handoff`) with `emitKernelReady` action  
**Expected failure:** Current transition is `LIFECYCLE_READY` -> `handoff`  
**YAML change needed:**

```yaml
    transitions:
      - event: lifecycleReady
        target: ready
        actions:
          - action: emitKernelReady
```

**Minimal implementation:** Update lifecycle transition to ready state with action

---

### Test 11: `TestBootstrapChart_HasReadyFinalState`

**What it verifies:** `ready` state exists as final state with logSuccess action  
**Expected failure:** Current final state is `complete`, not `ready`  
**YAML change needed:**

```yaml
  ready:
    type: final
    entryActions:
      - action: logSuccess
        params:
          message: "Maelstrom bootstrap complete"
```

**Minimal implementation:** Add ready state as final state

---

### Test 12: `TestBootstrapChart_HasLowercaseEvents`

**What it verifies:** All events use lowercase camelCase naming (securityReady, not SECURITY_READY)  
**Expected failure:** Current events are UPPERCASE  
**YAML change needed:** Change all events from UPPERCASE to lowercase  
**Minimal implementation:** Update all event names in transitions

---

### Test 13: `TestBootstrapChart_NoInitializingOrHandoffStates`

**What it verifies:** Removing unnecessary states (`initializing`, `handoff`)  
**Expected failure:** These states currently exist  
**YAML change needed:** Remove these states entirely  
**Minimal implementation:** Delete state definitions

---

## Complete Target YAML Structure

After all tests pass, the YAML should be:

```yaml
id: sys:bootstrap
version: 1.0.0
spec:
  type: compound
  initial: security
  states:
    security:
      type: atomic
      entryActions:
        - action: securityBootstrap
          params:
            boundaries: [inner, dmz, outer]
      transitions:
        - event: securityReady
          target: communication
        - event: securityFailed
          target: failed

    communication:
      type: atomic
      entryActions:
        - action: communicationBootstrap
          params:
            mailBackbone: true
      transitions:
        - event: communicationReady
          target: observability
        - event: communicationFailed
          target: failed

    observability:
      type: atomic
      entryActions:
        - action: observabilityBootstrap
          params:
            tracing: true
            metrics: true
            deadLetterQueue: true
      transitions:
        - event: observabilityReady
          target: lifecycle
        - event: observabilityFailed
          target: failed

    lifecycle:
      type: atomic
      entryActions:
        - action: lifecycleBootstrap
          params:
            enableSpawn: true
            enableStop: true
            toolRegistry: true
      transitions:
        - event: lifecycleReady
          target: ready
          actions:
            - action: emitKernelReady

    ready:
      type: final
      entryActions:
        - action: logSuccess
          params:
            message: "Maelstrom bootstrap complete"

    failed:
      type: final
      entryActions:
        - action: logFailure
        - action: panic
```

---

## State Transition Diagram

```
                    ┌─────────────┐
                    │   security  │
                    │ (boundaries)│
                    └──────┬──────┘
          securityReady    │    securityFailed
               ┌───────────┼──────────────────────┐
               │           │                      │
               ▼           │                      │
        ┌─────────────┐    │                      │
        │communication│    │                      │
        │  (mail)     │    │                      │
        └──────┬──────┘    │                      │
 communicationReady  │    │                      │
      ┌──────────────┼────┼──────────────────────┤
      │              │    │                      │
      ▼              │    │                      │
 ┌─────────────┐     │    │                      │
 │observability│     │    │                      │
 │(trace/metrics)│    │                      │
 └──────┬──────┘     │    │                      │
      │              │    │                      │
      ▼              │    │                      │
 ┌─────────────┐     │    │                      │
 │  lifecycle  │     │    │                      │
 │(spawn/stop) │     │    │                      │
 └──────┬──────┘     │    │                      │
        │            │    │                      │
        │lifecycleReady        │                      │
        │            │    │                      │
        ▼            │    │                      │
    ┌────────┐       │    │                      │
    │ ready  │◄──────┘    │                      │
    │(final) │            │                      │
    └────────┘            │                      │
                          │                      │
                          ▼                      ▼
                      ┌─────────┐            ┌─────────┐
                      │ failed  │            │ failed  │
                      │ (final) │            │ (final) │
                      └─────────┘            └─────────┘
```

---

## Error Handling Strategy

### 1. **Fail-Fast on Bootstrap Failure**

- Any service load failure triggers immediate transition to `failed` state
- `failed` state is final - no recovery path
- `panic` action ensures kernel exits on bootstrap failure

### 2. **Event-Driven Failure Signaling**

- Each state emits either `XReady` (success) or `XFailed` (failure)
- Sequence handler must support both event types
- Failure events must route to `failed` state

### 3. **Structured Error Data**

- `logFailure` action should capture:
  - Which state failed
  - Error message from entry action
  - Stack trace for debugging
- `panic` action terminates with non-zero exit code

### 4. **Transition Validation**

- Invalid events in any state should be rejected (current behavior)
- `failed` and `ready` states have no outgoing transitions

---

## sequence.go Updates Required

The `sequence.go` file needs these changes to support the new chart:

1. **Add failure event handling:**

```go
transitions := map[string]map[string]string{
    "security":      {"securityReady": "communication", "securityFailed": "failed"},
    "communication": {"communicationReady": "observability", "communicationFailed": "failed"},
    "observability": {"observabilityReady": "lifecycle", "observabilityFailed": "failed"},
    "lifecycle":     {"lifecycleReady": "ready"},
}
```

1. **Support structured entry actions with params:**

- `onStateEnter` callback needs to receive action params
- May need to parse YAML action structure and pass params to service loaders

1. **Handle final states:**

- `ready` state triggers `onComplete` callback
- `failed` state triggers panic/hard exit

---

## Summary

| Item | Count |
|------|-------|
| Tests to write | 13 |
| States to modify | 4 (security, communication, observability, lifecycle) |
| States to add | 2 (ready, failed) |
| States to remove | 2 (initializing, handoff) |
| Transitions to add | 4 (failure transitions) |
| Event names to rename | 7 |

---

**End of Phase 1.2 TDD Plan**
