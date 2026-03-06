# Maelstrom E2E Test Plan

**Date**: 2026-03-02  
**Status**: Planning Phase  
**Based on**: Architecture docs, implementation plans, and current test suite

---

## Executive Summary

This plan defines end-to-end (e2e) test scenarios that validate the complete Maelstrom stack working together. Unlike unit tests that test individual services, e2e tests verify:

1. **Service Integration**: All platform services communicate correctly via Mail system
2. **Boundary Enforcement**: Security boundaries work across service boundaries
3. **Data Flow**: Taints propagate correctly through ContextMap, Mail, and storage
4. **Full Lifecycle**: Agents can be created, run tools, persist state, and be restored

**Current State**: 217 unit tests passing across kernel, services, security, and platform services  
**Target**: 20-25 new e2e tests covering all major integration scenarios  
**Expected Total**: ~240 tests

---

## Test Categories

### Category 1: Bootstrap & Startup (2 tests)

**Goal**: Validate complete system initialization from kernel to all services

| Test ID | Name | What It Tests | Expected Outcome |
|---------|------|---------------|------------------|
| E2E-001 | `TestE2E_FullBootstrap` | Kernel bootstrap sequence + all platform services load | All 8 services register, kernel_ready emitted, services can receive mail |
| E2E-002 | `TestE2E_ServiceCommunication` | Services communicate via Mail system | Service A can publish mail, Service B receives and processes it |

### Category 2: Agent Lifecycle (3 tests)

**Goal**: Validate complete agent lifecycle with tool usage

| Test ID | Name                              | What It Tests                                         | Expected Outcome                                                  |
| ------- | --------------------------------- | ----------------------------------------------------- | ----------------------------------------------------------------- |
| E2E-003 | `TestE2E_AgentCreation`           | Agent spawned via sys:lifecycle, ContextMap assembled | Agent created with correct boundary, ContextMap has all 4 sources |
| E2E-004 | `TestE2E_AgentToolUsage`          | Agent invokes tool via sys:tools, results flow back   | Tool isolated, result tainted, injected into agent session        |
| E2E-005 | `TestE2E_AgentSessionPersistence` | Agent state persisted and restored                    | Snapshot created, agent restored with history preserved           |

### Category 3: Security & Boundaries (4 tests)

**Goal**: Validate boundary enforcement across all layers

| Test ID | Name | What It Tests | Expected Outcome |
|---------|------|---------------|------------------|
| E2E-006 | `TestE2E_BoundaryMailValidation` | Mail validated when crossing boundaries | Outer→DMZ mail accepted, DMZ→inner mail rejected/sanitized |
| E2E-007 | `TestE2E_ContextMapTaintFiltering` | ContextMap filters forbidden taints by boundary | DMZ agent cannot see PII/SECRET blocks, only allowed taints |
| E2E-008 | `TestE2E_StreamSanitization` | Outgoing streams sanitized before emission | PII redacted, inner data stripped, length capped |
| E2E-009 | `TestE2E_TaintPropagation` | Taints propagate through all data structures | Message→ContextMap→ApplicationContext all carry taints |

### Category 4: Persistence & Migration (3 tests)

**Goal**: Validate data persistence and version migration

| Test ID | Name | What It Tests | Expected Outcome |
|---------|------|---------------|------------------|
| E2E-010 | `TestE2E_SnapshotRestore` | Agent snapshot created and restored | Snapshot contains state + events, restore recreates agent |
| E2E-011 | `TestE2E_MigrateShallowHistory` | Migration with shallowHistory preserves parent state | Agent restored to parent default sub-state |
| E2E-012 | `TestE2E_MigrateDeepHistory` | Migration with deepHistory preserves specific state | Agent restored to exact previous sub-state |

### Category 5: Memory Service (2 tests)

**Goal**: Validate long-term memory operations

| Test ID | Name | What It Tests | Expected Outcome |
|---------|------|---------------|------------------|
| E2E-013 | `TestE2E_MemoryStoreQuery` | Store vector, query by similarity | Memory stored with taints, query returns top-K results |
| E2E-014 | `TestE2E_MemoryBoundaryFilter` | Memory queries filtered by boundary | Inner memories not visible to DMZ/outer agents |

### Category 6: Heartbeat (2 tests)

**Goal**: Validate scheduled agent wake-ups

| Test ID | Name | What It Tests | Expected Outcome |
|---------|------|---------------|------------------|
| E2E-015 | `TestE2E_HeartbeatSchedule` | Cron schedule created and fires | Schedule stored, cron triggers HEARTBEAT event |
| E2E-016 | `TestE2E_HeartbeatTemplateInjection` | HEARTBEAT.md injected into session | Template rendered, message injected with correct taints |

### Category 7: Gateway Adapters (3 tests)

**Goal**: Validate external channel adapters

| Test ID | Name | What It Tests | Expected Outcome |
|---------|------|---------------|------------------|
| E2E-017 | `TestE2E_GatewayWebhookIngress` | Webhook POST normalized to mail_received | HTTP POST → mail with correct type/source/target |
| E2E-018 | `TestE2E_GatewaySSEStreaming` | SSE stream chunks sent to client | Stream opens, chunks sent with taints, final marker sent |
| E2E-019 | `TestE2E_GatewayWebSocketBidir` | WebSocket bidirectional messaging | WS connection established, messages flow both ways |

### Category 8: Human Gateway (2 tests)

**Goal**: Validate human-in-the-loop chat

| Test ID | Name | What It Tests | Expected Outcome |
|---------|------|---------------|------------------|
| E2E-020 | `TestE2E_HumanGatewaySession` | Chat session created for agent | Session opened, agent receives human_feedback mail |
| E2E-021 | `TestE2E_HumanGatewayStream` | Agent response streamed to human | Stream chunks sent, taints sanitized, isFinal marker |

### Category 9: Admin Console (2 tests)

**Goal**: Validate admin operations with 2FA

| Test ID | Name | What It Tests | Expected Outcome |
|---------|------|---------------|------------------|
| E2E-022 | `TestE2E_Admin2FARequired` | Admin endpoints require 2FA auth | Without 2FA → 401, with 2FA → 200 |
| E2E-023 | `TestE2E_AdminControlAgent` | Admin can control agents | List agents, pause/resume/stop works, taints queryable |

### Category 10: DataSource Operations (2 tests)

**Goal**: Validate file storage with tainting

| Test ID | Name | What It Tests | Expected Outcome |
|---------|------|---------------|------------------|
| E2E-024 | `TestE2E_DatasourceTagOnWrite` | Files written with taints | File written, xattr tags set, taints readable |
| E2E-025 | `TestE2E_DatasourceAccessValidation` | File access validated by boundary | Inner files not accessible from outer boundary |

---

## Test Implementation Details

### Test Infrastructure

Create `/pkg/integration/` directory with shared test helpers:

```go
pkg/integration/
├── integration.go       # Test suite setup, shared fixtures
├── agents.go            # Agent creation helpers
├── mail.go              # Mail system test helpers
├── boundary.go          # Boundary test utilities
├── persistence.go       # Snapshot/restore helpers
└── *_test.go            # Individual e2e tests
```

### Shared Test Fixtures

All e2e tests share common setup:

```go
func TestE2E_Main(t *testing.T) {
    // Setup: Initialize full Maelstrom stack
    kernel := bootstrap.NewTestKernel()
    registry := chart.NewTestRegistry()
    
    // Start all services
    services := []string{
        "sys:gateway",
        "sys:admin",
        "sys:persistence",
        "sys:heartbeat",
        "sys:memory",
        "sys:tools",
        "sys:datasources",
        "sys:humangateway",
    }
    for _, svc := range services {
        require.NoError(t, kernel.SpawnService(svc))
    }
    
    // Wait for kernel_ready
    require.Eventually(t, kernel.IsReady, 5*time.Second, 100*time.Millisecond)
    
    // Run test
    testFunc(t, kernel, registry)
    
    // Teardown
    kernel.Stop()
}
```

### Test Naming Convention

- **Prefix**: `TestE2E_` for all e2e tests
- **Category**: `Agent`, `Boundary`, `Persistence`, etc.
- **Scenario**: `Creation`, `ToolUsage`, `SnapshotRestore`, etc.

Example: `TestE2E_AgentToolUsage` - Agent lifecycle, tool usage scenario

---

## Test Execution Order

E2e tests should run in this order due to dependencies:

1. **Bootstrap tests** (E2E-001, E2E-002) - Must run first to initialize system
2. **Agent lifecycle tests** (E2E-003, E2E-004, E2E-005) - Need bootstrap complete
3. **Security tests** (E2E-006 to E2E-009) - Need agents running
4. **Persistence tests** (E2E-010 to E2E-012) - Need agents with state
5. **Memory tests** (E2E-013, E2E-014) - Standalone, can run in parallel
6. **Heartbeat tests** (E2E-015, E2E-016) - Standalone, can run in parallel
7. **Gateway tests** (E2E-017 to E2E-019) - Standalone, can run in parallel
8. **Human Gateway tests** (E2E-020, E2E-021) - Need agents running
9. **Admin tests** (E2E-022, E2E-023) - Standalone, can run in parallel
10. **DataSource tests** (E2E-024, E2E-025) - Standalone, can run in parallel

**Parallel Execution**: Tests in categories 5-10 can run in parallel (no shared state)

---

## Expected Test Outcomes

### Success Criteria

1. ✅ All 25 e2e tests pass
2. ✅ All 217 existing unit tests still pass
3. ✅ Total test count: ~242 tests
4. ✅ Test coverage: All major integration points covered
5. ✅ Execution time: <5 minutes total (with parallelization)

### Failure Scenarios to Handle

| Scenario | Expected Behavior | Test Response |
|----------|------------------|---------------|
| Service fails to start | Bootstrap test fails early | Fail fast, log service errors |
| Mail delivery fails | Communication test fails | Check service logs, verify subscriptions |
| Boundary violation | Security test catches it | Verify taints rejected/sanitized |
| Snapshot restore fails | Persistence test fails | Verify snapshot data integrity |
| Memory query returns wrong results | Memory test fails | Check vector similarity calculation |

---

## Test Data & Fixtures

### Sample Agent Definition

```yaml
apiVersion: maelstrom.dev/v1
kind: AgentSpec
metadata:
  name: e2e-test-agent
  version: "1.0.0"
  boundary: dmz
spec:
  llmConfig:
    modelFamily: test-model
    temperature: 0.7
    maxTokens: 1000
    streaming: true
  contextMap:
    - name: systemPrompt
      source: static
      content: "You are a test agent."
      maxTokens: 500
    - name: conversation
      source: session
      strategy: lastN
      n: 10
  tools:
    - name: testTool
      boundary: dmz
      isolation: strict
```

### Sample Test Tool

```go
func testTool(ctx ApplicationContext, args map[string]any) (any, error) {
    // Simulated tool that returns tainted data
    return map[string]any{
        "result": "test output",
        "taint":  "TOOL_OUTPUT",
    }, nil
}
```

### Sample Memory Entry

```go
memoryEntry := MemoryResult{
    ID:      "mem-123",
    Content: "Important long-term memory",
    Vector:  []float32{0.1, 0.2, 0.3}, // 3-dim for testing
    Taints:  []string{"INNER_ONLY"},
    Metadata: map[string]any{
        "agentId": "agent-456",
        "timestamp": time.Now(),
    },
}
```

---

## Integration Points to Validate

### 1. Mail System Integration

**What**: All services communicate via Mail, not direct calls  
**How to test**: Service A publishes mail, Service B receives and processes  
**Verify**: Mail has correct type, source, target, taints

### 2. Boundary Propagation

**What**: Boundaries enforced across all service boundaries  
**How to test**: Outer→DMZ mail accepted, DMZ→inner mail rejected  
**Verify**: Taints checked, violations logged to dead-letter

### 3. ContextMap Assembly

**What**: All 4 sources (static, session, memory, tool) work together  
**How to test**: Agent ContextMap has blocks from all sources  
**Verify**: Blocks ordered by priority, filtered by boundary

### 4. Service Registration

**What**: Services register in bootstrap.go  
**How to test**: All 8 services available after bootstrap  
**Verify**: Services addressable by well-known IDs (sys:*)

### 5. Hot-Reload Compatibility

**What**: Services can reload without downtime  
**How to test**: Replace service definition, system continues running  
**Verify**: State preserved, events queued correctly

---

## Risk Assessment

| Risk | Likelihood | Impact | Mitigation |
|------|-----------|--------|------------|
| Tests too slow | Medium | High | Parallelize independent tests, use mocks |
| Flaky tests | Medium | Medium | Use retries, avoid timing-dependent assertions |
| Test data cleanup | Low | Medium | Teardown each test, use unique IDs |
| Integration gaps | Medium | High | Cover all major integration points |
| Test maintenance | Low | Low | Document test purposes, keep focused |

---

## Test Execution Commands

### Run All E2E Tests

```bash
go test -v ./pkg/integration/...
```

### Run Specific Test

```bash
go test -v -run TestE2E_AgentToolUsage ./pkg/integration/...
```

### Run with Coverage

```bash
go test -v -cover ./pkg/integration/...
```

### Parallel Execution (Default)

```bash
go test -v -parallel 8 ./pkg/integration/...
```

---

## Deferred E2E Scenarios

### Phase 5: Execution Layer E2E

- Tool execution policies (seq_failfast, seq_continue, par_continue)
- Orchestrator sub-chart behavior
- Sub-agent invocation with context inheritance

**Milestone**: Agents can execute complex tool sequences

### Phase 6: Agent Reasoning E2E

- OODA loop execution (observe, orient, decide, act, reflect)
- LLM bridge integration with streaming
- Context compression strategies (summarize vs truncate vs RAG)

**Milestone**: Agents can reason autonomously

### Phase 7: External Interface E2E

- HTTP gateway with OpenAPI spec generation
- SSE/WS streaming with auth
- 2FA enforcement on admin endpoints

**Milestone**: External clients can interact with agents

---

## Success Metrics

### Coverage Metrics

- **Integration Points**: All 5 major integration points covered
- **Services Tested**: All 8 platform services exercised
- **Boundaries Tested**: All 3 boundary types (inner, dmz, outer)
- **Data Types Tested**: Mail, ContextMap, Memory, Snapshots

### Quality Metrics

- **Test Pass Rate**: 100% (no flaky tests)
- **Execution Time**: <5 minutes total
- **Code Coverage**: >80% on integration code
- **Documentation**: Each test has clear purpose documented

### Business Metrics

- **Confidence**: High confidence in system integration
- **Regression Detection**: Fast detection of integration breaks
- **Developer Experience**: Clear test failures with actionable messages

---

## Questions for Review

1. **Test Count**: 25 e2e tests - is this appropriate given scope?
2. **Test Categories**: Are the 10 categories comprehensive?
3. **Execution Order**: Is the proposed order correct?
4. **Parallelization**: Which tests can safely run in parallel?
5. **Mock vs Real**: Should tests use real services or mocks?

---

## Next Steps

1. **Review**: Get approval on test plan
2. **Implement**: Create `/pkg/integration/` directory structure
3. **Write**: Implement tests one category at a time
4. **Validate**: Run all tests, ensure 100% pass rate
5. **Document**: Update docs with e2e test results

---

## References

- `/home/albert/git/maelstrom-v4/docs/arch-v1.md` - Architecture specification
- `/home/albert/git/maelstrom-v4/docs/initial-exploration/phase4-services-plan.md` - Service implementation plan
- `/home/albert/git/maelstrom-v4/docs/initial-exploration/security-layer-implementation-plan.md` - Security layer plan
- `/home/albert/git/maelstrom-v4/docs/integration-guide.md` - Integration guide