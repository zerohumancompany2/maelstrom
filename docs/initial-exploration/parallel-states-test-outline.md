# Parallel States Test Outline

Comprehensive test specification derived from arch-v1.md Section 17 and v3 implementation plan semantics.

## Test Categories

### 1. Unit Tests: Parallel State Structure

| Test Case                              | Description                                         | Expected Result                           |
| -------------------------------------- | --------------------------------------------------- | ----------------------------------------- |
| `TestParallel_NodeTypeRecognition`     | Node with ≥2 children is recognized as parallel     | `nodeType == parallel`                    |
| `TestParallel_RegionNamesExtracted`    | Region names parsed from children keys              | `regionNames == []{"regionA", "regionB"}` |
| `TestParallel_Invalid_NoChildren`      | Parallel state must have ≥2 children                | Error on validation                       |
| `TestParallel_Invalid_SingleChild`     | Single child is compound, not parallel              | `nodeType == compound`                    |
| `TestParallel_ActiveStatesFormat`      | ActiveStates contains one StatePath per region      | `len(ActiveStates) == numRegions`         |
| `TestParallel_StatePathIncludesRegion` | Format: `root/parallelNode/regionName/currentState` | Correct path structure                    |

### 2. Unit Tests: Event Router

| Test Case                                     | Description                                               | Expected Result               |
| --------------------------------------------- | --------------------------------------------------------- | ----------------------------- |
| `TestRouter_SelfRouting_NoContext_Broadcast`  | `TargetPath="."`, no `SourceRegion` → all regions receive | All region queues get event   |
| `TestRouter_SelfRouting_WithContext_Targeted` | `TargetPath="."`, `SourceRegion="regionA"` → only A       | Only regionA queue gets event |
| `TestRouter_ExplicitRegion_Targeted`          | `TargetPath="region:regionB"` → only B                    | Only regionB queue gets event |
| `TestRouter_InvalidRegion_Error`              | `TargetPath="region:nonexistent"`                         | Error: region not found       |
| `TestRouter_ExternalEvent_Broadcast`          | External event with no TargetPath                         | Broadcast to all regions      |
| `TestRouter_CrossRegionEvent_Routed`          | RegionA emits event with `TargetPath="region:regionB"`    | Routed through parent to B    |
| `TestRouter_EventOrdering_QueueFIFO`          | Events dispatched in order                                | Regions process FIFO          |

### 3. Unit Tests: Region Runtime

| Test Case                                | Description                                | Expected Result                              |
| ---------------------------------------- | ------------------------------------------ | -------------------------------------------- |
| `TestRegion_GoroutineIsolation`          | Each region runs in separate goroutine     | Concurrent execution                         |
| `TestRegion_EventQueueBuffered`          | Events queued when region busy             | No blocking on dispatch                      |
| `TestRegion_ParentChanSignalsDone`       | Region enters final state → `done` message | Parent receives `RegionMessage{Type:"done"}` |
| `TestRegion_ParentChanSignalsTransition` | Region transitions → `transition` message  | Parent receives state update                 |
| `TestRegion_ParentChanEmitsEvent`        | Region emits event → `event` message       | Parent routes event                          |
| `TestRegion_NoDirectRegionAccess`        | Region cannot access other regions         | Compilation/panic on attempt                 |
| `TestRegion_SynchronousProcessing`       | Single event processed to completion       | Run-to-completion per region                 |

### 4. Unit Tests: Parallel Entry

| Test Case                                 | Description                            | Expected Result                   |
| ----------------------------------------- | -------------------------------------- | --------------------------------- |
| `TestEntry_AllRegionsStartSimultaneously` | All regions enter initial states       | All regions active post-entry     |
| `TestEntry_EntryActionsExecuted`          | Entry actions run for each region      | Actions executed per-region       |
| `TestEntry_InitialStateDetermination`     | Correct initial child selected         | Default initial state entered     |
| `TestEntry_InitialHistory_RespectsDeep`   | History state restoration              | Previous state restored           |
| `TestEntry_EventRouterStarted`            | EventRouter goroutine running          | Events can be routed              |
| `TestEntry_ActiveStatesPopulated`         | ActiveStates contains all region paths | `len(ActiveStates) == numRegions` |

### 5. Unit Tests: Completion & Exit

| Test Case                                  | Description                         | Expected Result                   |
| ------------------------------------------ | ----------------------------------- | --------------------------------- |
| `TestCompletion_RegionEmitsDone`           | Region reaches final state → `done` | Parent tracks completion          |
| `TestCompletion_AllRegionsDone_AllowsExit` | All `done` → exit proceeds          | Parallel state can transition out |
| `TestCompletion_PartialDone_Waits`         | 1/2 regions done → waits            | Exit blocked until all done       |
| `TestCompletion_OrderingNotGuaranteed`     | Regions complete in any order       | Exit only when all done           |
| `TestExit_QuiescenceAchieved`              | Exit waits for current events       | Quiescence before cleanup         |
| `TestExit_ExitActionsAfterQuiescence`      | Exit actions run after quiescence   | Deterministic cleanup             |
| `TestExit_AllRegionExitActionsFire`        | Every region's exit actions run     | All exit actions executed         |
| `TestExit_ExitActionOrder_NotGuaranteed`   | Region exit actions concurrent      | No ordering dependency            |

### 6. Unit Tests: Parent Transitions

| Test Case                                | Description                            | Expected Result                  |
| ---------------------------------------- | -------------------------------------- | -------------------------------- |
| `TestParentTransition_InitiatesExit`     | Parent transition starts exit sequence | Exit signal sent to regions      |
| `TestParentTransition_WaitsForDone`      | Parent blocked until all done          | Transition completes after done  |
| `TestParentTransition_RegionAlsoHandles` | Same event in region → both fire       | Region transitions, parent exits |
| `TestParentTransition_QuiescenceFirst`   | Quiescence before exit signal          | Current events complete first    |
| `TestParentTransition_GuardEvaluation`   | Guard checked before exit              | Guard prevents transition        |

### 7. Unit Tests: Control Commands

| Test Case                                   | Description                  | Expected Result                 |
| ------------------------------------------- | ---------------------------- | ------------------------------- |
| `TestControl_Pause_AchievesQuiescence`      | `CmdPause` → quiescence      | Dispatch halted, regions finish |
| `TestControl_Pause_FreezeParentTransitions` | Parent transitions blocked   | Whole parallel state still      |
| `TestControl_Resume_FromQuiescence`         | `CmdResume` resumes dispatch | Events flow again               |
| `TestControl_Resume_QueueDrained`           | Queued events processed      | FIFO order preserved            |
| `TestControl_Stop_Hard_NoQuiescence`        | `CmdStop` immediate          | No waiting, force close         |
| `TestControl_Stop_ExitActionsBestEffort`    | Exit actions with timeout    | Attempt then abandon            |
| `TestControl_Snapshot_PauseCaptureResume`   | `CmdSnapshot` flow           | State captured at quiescence    |
| `TestControl_Snapshot_IncludesQueueState`   | In-flight events captured    | Queue state in snapshot         |

### 8. Unit Tests: Force-Stop Scenarios

| Test Case                               | Description                    | Expected Result                |
| --------------------------------------- | ------------------------------ | ------------------------------ |
| `TestForceStop_RegionNeverReachesFinal` | Region loops forever           | Force-stop after timeout       |
| `TestForceStop_MidEventProcessing`      | Force-stop during event        | Event abandoned, region closed |
| `TestForceStop_PendingEvents_Discarded` | Queue has pending events       | Events dropped                 |
| `TestForceStop_CleanupStillRuns`        | Cleanup executes despite force | Best-effort cleanup            |
| `TestForceStop_ErrorLogged`             | Force-stop reason logged       | Observability hook fired       |

### 9. Unit Tests: Quiescence Detection

| Test Case                                  | Description                      | Expected Result               |
| ------------------------------------------ | -------------------------------- | ----------------------------- |
| `TestQuiescence_EmptyQueue`                | No pending events                | Quiescence achieved           |
| `TestQuiescence_RegionMidEvent_Waits`      | Region processing event          | Quiescence blocked            |
| `TestQuiescence_CrossRegionSignalInFlight` | Parent processing region message | Quiescence blocked            |
| `TestQuiescence_AllComplete_Achieved`      | All regions idle                 | Quiescence signaled           |
| `TestQuiescence_EntryActionImmediateFinal` | Entry action → final state       | `done` emitted, quiescence OK |

### 10. Integration Tests: Full Lifecycle

| Test Case                                      | Description                         | Steps                                                     |
| ---------------------------------------------- | ----------------------------------- | --------------------------------------------------------- |
| `TestLifecycle_SpawnStartTransitionExit`       | Complete parallel lifecycle         | Spawn → Start → Events → Exit → Stop                      |
| `TestLifecycle_CrossRegionCoordination`        | RegionA signals RegionB             | A emits → Parent routes → B receives → B done → A done    |
| `TestLifecycle_GracefulExit_WithActiveRegions` | Both regions active, exit triggered | Exit signal → Regions process → Both done → Exit complete |
| `TestLifecycle_PauseSnapshotResume`            | Pause and snapshot flow             | Pause → Quiescence → Snapshot → Resume → Continue         |
| `TestLifecycle_HotReload_DuringParallel`       | ReplaceDefinition during parallel   | Quiescence → Definition swap → State migration → Resume   |

### 11. Integration Tests: Error Handling

| Test Case                             | Description                  | Expected Result                      |
| ------------------------------------- | ---------------------------- | ------------------------------------ |
| `TestError_RegionPanic_Recovered`     | Region goroutine panics      | Parent catches, logs, force-stops    |
| `TestError_RegionError_Message`       | Region emits `error` message | Parent handles, may force-stop       |
| `TestError_ActionError_Continues`     | Entry action fails           | Error logged, transition continues   |
| `TestError_ExitActionError_Continues` | Exit action fails            | Error logged, other exit actions run |
| `TestError_RouterDeadlock_Detected`   | EventRouter blocks           | Timeout, error, force-stop           |

### 12. Property Tests (Deterministic)

| Test Case                                 | Description                             | Property                       |
| ----------------------------------------- | --------------------------------------- | ------------------------------ |
| `TestProperty_DoneIdempotency`            | Region emits `done` multiple times      | Only first counts              |
| `TestProperty_QuiescenceIdempotency`      | Multiple pause commands                 | Idempotent, single quiescence  |
| `TestProperty_RegionIsolation`            | Events in one region don't affect other | State independence             |
| `TestProperty_EventOrderingPreserved`     | Events processed in dispatch order      | Per-region FIFO                |
| `TestProperty_ActiveStatesAlwaysComplete` | ActiveStates always has all regions     | Never partial during operation |

### 13. Concurrency Tests

| Test Case                                 | Description                             | Scenario                         |
| ----------------------------------------- | --------------------------------------- | -------------------------------- |
| `TestConcurrent_MultipleDispatches`       | Many events dispatched rapidly          | No races, all processed          |
| `TestConcurrent_DispatchDuringQuiescence` | Dispatch while achieving quiescence     | Rejected or queued appropriately |
| `TestConcurrent_RegionDoneDuringDispatch` | Region done while event in flight       | Proper synchronization           |
| `TestConcurrent_ParentTransitionRace`     | Exit triggered during region processing | Quiescence wins, exit follows    |
| `TestConcurrent_SnapshotDuringActivity`   | Snapshot while events flowing           | Consistent state captured        |

## Test Data: Chart Definitions

### Basic Parallel Chart

```yaml
nodes:
  root:
    type: parallel
    regions:
      - name: regionA
        child:
          idle:
            type: atomic
            transitions:
              - event: "nextA"
                to: "doneA"
          doneA:
            type: final
      - name: regionB
        child:
          idle:
            type: atomic
            transitions:
              - event: "nextB"
                to: "doneB"
          doneB:
            type: final
```

### Cross-Region Coordination Chart

```yaml
nodes:
  root:
    type: parallel
    regions:
      - name: coordinator
        child:
          dispatch:
            type: atomic
            onEnter:
              - action: emitEvent
                event:
                  type: "startWorker"
                  targetPath: "region:worker"
            transitions:
              - event: "workerDone"
                to: "complete"
          complete:
            type: final
      - name: worker
        child:
          idle:
            type: atomic
            transitions:
              - event: "startWorker"
                to: "processing"
          processing:
            type: atomic
            transitions:
              - event: "workComplete"
                action:
                  - emitEvent:
                      type: "workerDone"
                      targetPath: "region:coordinator"
                to: "done"
          done:
            type: final
```

### Parent Transition Chart

```yaml
nodes:
  root:
    type: parallel
    onExit: ["cleanupParallel"]
    transitions:
      - event: "abort"
        to: "errorState"
    regions:
      - name: regionA
        child:
          processing:
            type: atomic
            onExit: ["cleanupA"]
            transitions:
              - event: "finishA"
                to: "doneA"
              - event: "abort"
                to: "doneA"
          doneA:
            type: final
      - name: regionB
        child:
          processing:
            type: atomic
            onExit: ["cleanupB"]
            transitions:
              - event: "finishB"
                to: "doneB"
              - event: "abort"
                to: "doneB"
          doneB:
            type: final
  errorState:
    type: atomic
```

## Compliance Mapping

| arch-v1 Requirement                                            | Test Coverage                   |
| -------------------------------------------------------------- | ------------------------------- |
| SHALL evaluate events synchronously inside parallel boundaries | Unit tests 3, 4; Property tests |
| SHALL fire trace callback on lifecycle events                  | Deferred (observability)        |
| SHALL provide `replaceDefinition` hook                         | Integration test 10             |
| SHALL run parallel regions in isolated goroutines              | Unit tests 3; Concurrency tests |
| SHALL support transient sub-charts                             | Deferred                        |

## Implementation Priority

### Phase 1: Core Structure (Blocking)
1. Node type recognition
2. Region initialization
3. ActiveStates format

### Phase 2: Event Routing (Blocking)
1. EventRouter basic routing
2. Self-routing (both variants)
3. Explicit region targeting

### Phase 3: Lifecycle (Blocking)
1. Parallel entry
2. Region runtime goroutines
3. Completion tracking

### Phase 4: Exit & Quiescence (Blocking)
1. Quiescence detection
2. Graceful exit flow
3. Exit action execution

### Phase 5: Control Commands
1. Pause/resume
2. Hard stop
3. Snapshot

### Phase 6: Integration & Edge Cases
1. Cross-region coordination
2. Error handling
3. Concurrency stress tests

---

*Test outline derived from arch-v1.md Section 17 and parallel-states-implementation-plan-v3.md*
