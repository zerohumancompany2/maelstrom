# Phase 7.7: Integration Tests

**Phase ID**: L7-P7.7  
**Title**: Integration Tests  
**Parent**: Layer 7 (Platform Services)  
**Status**: ❌ PENDING

## Parent Requirements

From `arch-v1.md L449-492` - Platform Services:
1. All Platform Services are Charts themselves, addressable by well-known IDs (L451)
2. Core Services (Hard-Coded): sys:security, sys:communication, sys:observability, sys:lifecycle (L453-460)
3. Hot-Reloadable Services: sys:gateway, sys:admin, sys:persistence, sys:heartbeat, sys:memory, sys:human-gateway, sys:tools, sys:datasources (L462-473)
4. Platform Service Contract: `handleMail(mail: Mail) → outcomeEvent` (L479-480)

From `arch-v1.md L800-897` - Bootstrap Sequence:
1. KERNEL_READY emitted after hard-coded services ready (L800-847)
2. Hot-reloadable services loaded after KERNEL_READY (L840-847)
3. Kernel goes dormant after handoff to ChartRegistry (L848-849)

From `arch-v1.md L865-882` - Hot-Reload Protocol:
1. Quiescence detection: empty queue, no active regions, no inflight tools (L865-868)
2. Hot-reload protocol: prepareForReload, quiescence wait, force-stop (L870-882)

## Dependencies

- Phase 7.1: Hard-Coded Services Completion
- Phase 7.2: Hot-Reloadable Services Completion
- Phase 7.3: Post-KERNEL_READY Loading
- Phase 7.4: Hot-Reload Protocol Integration
- Phase 7.5: Service-Specific Contracts
- Phase 7.6: Service Discovery & Runtime Registration

## Satisfied Lower-Layer Requirements

- Layer 5: ChartRegistry basic structure exists (arch-v1.md L463, layer-05-phase-breakdown.md)
- Layer 4: Boundary types (inner, dmz, outer) defined (arch-v1.md L261-270)
- Layer 3: Mail system with correlationId for fire-and-await (arch-v1.md L604)
- Layer 2: Core service infrastructure (arch-v1.md L453-461)
- Layer 0: Statechart Engine with transient sub-charts (arch-v1.md L402)

## Acceptance Criteria

1. All 12 platform services (4 hard-coded + 8 hot-reloadable) start during bootstrap
2. KERNEL_READY emitted after all hard-coded services ready
3. All hot-reloadable services loaded after KERNEL_READY
4. Mail can be sent to any sys:* service with handleMail invoked correctly
5. Hot-reload protocol completes for hot-reloadable service with quiescence detection and history preservation
6. All service contracts from arch-v1.md L483-491 implemented and verified

## Test Descriptions

### Test 1: TestIntegration_AllServicesBootstrap

**Given**: Kernel with all platform service definitions (4 hard-coded + 8 hot-reloadable)  
**When**: Bootstrap sequence executes from STARTUP to KERNEL_READY to dormant  
**Then**: All 12 platform services are spawned and registered in Kernel service map  
**Expected Result**: 
- sys:security, sys:communication, sys:observability, sys:lifecycle started before KERNEL_READY (arch-v1.md L457-460)
- sys:gateway, sys:admin, sys:persistence, sys:heartbeat, sys:memory, sys:human-gateway, sys:tools, sys:datasources started after KERNEL_READY (arch-v1.md L466-473)
- KERNEL_READY event emitted after hard-coded services ready (arch-v1.md L840-847)
- Kernel enters dormant state after handoff (arch-v1.md L848-849)  
**Spec Reference**: arch-v1.md L449-473, L800-849

### Test 2: TestIntegration_MailRoutingAllServices

**Given**: All 12 platform services registered with handleMail handlers  
**When**: Mail sent to each sys:* address (sys:security, sys:communication, sys:observability, sys:lifecycle, sys:gateway, sys:admin, sys:persistence, sys:heartbeat, sys:memory, sys:human-gateway, sys:tools, sys:datasources)  
**Then**: Each service's handleMail invoked with correct Mail, response routed back via CommunicationService  
**Expected Result**: 
- Mail routing to sys:* addresses dispatches to correct service handler (arch-v1.md L451)
- Each service implements handleMail(mail: Mail) → outcomeEvent contract (arch-v1.md L479-480)
- CommunicationService delivers mail to all registered subscribers (arch-v1.md L458)
- Response mail routed back with correlationId preserved (arch-v1.md L604)  
**Spec Reference**: arch-v1.md L451, L458, L479-480, L604

### Test 3: TestIntegration_HotReloadEndToEnd

**Given**: Hot-reloadable service (e.g., sys:gateway) running with active ChartRuntime  
**When**: New version of service YAML deployed, triggering hot-reload protocol  
**Then**: Quiescence detected, history preserved, new runtime spawned, service continues handling mail  
**Expected Result**: 
- prepareForReload signal sent to current runtime (arch-v1.md L870)
- IsQuiescent() returns true when event queue empty, no active regions, no inflight tools (arch-v1.md L865-868)
- Current runtime stopped, new runtime spawned with history (shallow/deep) (arch-v1.md L873-875)
- Service continues handling mail after reload without data loss (arch-v1.md L865-882)
- ContextTransform applied if version changed (arch-v1.md L888)  
**Spec Reference**: arch-v1.md L865-882

## Implementation Plan

### Files to Modify

1. `pkg/services/integration/integration_test.go` - Create comprehensive integration tests

### Stubs Needed

```go
// pkg/services/integration/integration_test.go
package integration

func TestIntegration_AllServicesBootstrap(t *testing.T)
func TestIntegration_MailRoutingAllServices(t *testing.T)
func TestIntegration_HotReloadEndToEnd(t *testing.T)
```

### Implementation Order

1. Create `pkg/services/integration/integration_test.go` with test stubs
2. Write Test 1 → RED → Wire all 12 services through bootstrap → GREEN → commit
3. Write Test 2 → RED → Verify mail routing to all sys:* services → GREEN → commit
4. Write Test 3 → RED → Verify hot-reload end-to-end flow → GREEN → commit

## Commit Plan

| Test | Commit Message |
|------|----------------|
| Test 1 | `feat(layer-7/platform-services): verify all 12 services bootstrap correctly` |
| Test 2 | `feat(layer-7/platform-services): verify mail routing to all sys:* services` |
| Test 3 | `feat(layer-7/platform-services): verify hot-reload end-to-end flow` |

## Deliverables

- 3 commits (1:1 test-to-commit ratio per CLAUDE.md)
- All tests passing
- Files modified: `pkg/services/integration/integration_test.go`
- All 12 platform services verified working end-to-end
- Full bootstrap sequence verified (hard-coded → KERNEL_READY → hot-reloadable → dormant)
- Mail routing verified for all sys:* services
- Hot-reload protocol verified end-to-end

---

*Document generated: 2026-03-10*  
*Based on: arch-v1.md v1.3, layer-07-phase-breakdown.md*