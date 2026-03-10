# Phase 7.6: Service Discovery & Runtime Registration

## Goal
Implement service discovery and runtime registration integration for platform services (arch-v1.md L451-473).

## Scope
- Service discovery via well-known IDs (sys:*)
- Runtime registration in Kernel service map
- Integration with CommunicationService for mail routing
- Service health monitoring
- 4 tests, 4 commits
- **Branch**: `feat/layer7-discovery`

## Current Status

### Implementation Status
| Component | Status | Notes |
|-----------|--------|-------|
| `Service discovery` | ⚠️ Partial | Well-known IDs exist; needs discovery API |
| `Runtime registration` | ⚠️ Partial | Kernel has service map; needs auto-registration |
| `Mail routing` | ⚠️ Partial | CommunicationService exists; needs service routing |

### Files Status
| File | Status |
|------|--------|
| `pkg/kernel/kernel.go` | ⚠️ Partial - add service discovery API |
| `pkg/services/communication/service.go` | ⚠️ Partial - add sys:* mail routing |

## Required Implementation

### Service Discovery via Well-Known IDs (arch-v1.md L451-473)
```go
// pkg/kernel/kernel.go
func (k *Kernel) DiscoverService(serviceID string) (RuntimeID, error) {
    // Return RuntimeID for sys:* service
    // Hard-coded: sys:security, sys:communication, sys:observability, sys:lifecycle (L457-460)
    // Hot-reloadable: sys:gateway, sys:admin, sys:persistence, sys:heartbeat, sys:memory, sys:human-gateway, sys:tools, sys:datasources (L466-473)
}
```

### Runtime Auto-Registration (arch-v1.md L451)
```go
// pkg/kernel/kernel.go
func (k *Kernel) RegisterService(serviceID string, runtime *ChartRuntime) {
    // Auto-register service with Kernel on spawn
    // Store in service map for discovery
}
```

### Mail Routing to sys:* Services (arch-v1.md L477-479)
```go
// pkg/services/communication/service.go
func (c *CommunicationService) routeToService(mail Mail) {
    // Dispatch to registered sys:* service handler
    // Invoke handleMail contract (L479)
}
```

### Service Health Monitoring (arch-v1.md L459)
```go
// pkg/kernel/kernel.go
func (k *Kernel) GetServiceHealth(serviceID string) ServiceHealth {
    // Report ready state for service
    // Log failed services to sys:observability dead-letter (L459)
}
```

## Tests to Write (4 tests, 4 commits)

### Test 1: TestServiceDiscovery_WellKnownIDs
```go
func TestServiceDiscovery_WellKnownIDs(t *testing.T)
```
**Acceptance Criteria:**
- All platform services discoverable by sys:* ID format (arch-v1.md L451)
- Hard-coded services: sys:security, sys:communication, sys:observability, sys:lifecycle (arch-v1.md L457-460)
- Hot-reloadable services: sys:gateway, sys:admin, sys:persistence, sys:heartbeat, sys:memory, sys:human-gateway, sys:tools, sys:datasources (arch-v1.md L466-473)

### Test 2: TestServiceDiscovery_RuntimeRegistration
```go
func TestServiceDiscovery_RuntimeRegistration(t *testing.T)
```
**Acceptance Criteria:**
- Services auto-register with Kernel on spawn (arch-v1.md L451)
- `GetServiceRuntimeID(name)` returns RuntimeID
- Service map includes all 12 platform services (arch-v1.md L457-473)

### Test 3: TestServiceDiscovery_MailRouting
```go
func TestServiceDiscovery_MailRouting(t *testing.T)
```
**Acceptance Criteria:**
- Mail to sys:* addresses routed to correct service (arch-v1.md L477-479)
- CommunicationService dispatches to registered handlers
- handleMail contract invoked for each service (arch-v1.md L479)

### Test 4: TestServiceDiscovery_HealthMonitoring
```go
func TestServiceDiscovery_HealthMonitoring(t *testing.T)
```
**Acceptance Criteria:**
- Service health trackable via sys:observability (arch-v1.md L459)
- Ready state reported for each service
- Failed services logged to dead-letter queue (arch-v1.md L459)

## Dependencies

### Test Dependencies
```
Test 1 (Well-known IDs) → Independent
Test 2 (Runtime registration) → Depends on Test 1
Test 3 (Mail routing) → Depends on Test 2
Test 4 (Health monitoring) → Depends on Tests 1-3
```

### Phase Dependencies
- **Phase 7.3** must complete before **Phase 7.6** (needs post-KERNEL_READY loading)
- **Phase 7.5** must complete before **Phase 7.6** (needs service contracts implemented)
- **Phase 7.7** depends on **Phase 7.6** (integration tests need discovery)

## Files to Create/Modify

| File | Operation | Changes |
|------|-----------|---------|
| `pkg/kernel/kernel.go` | MODIFY | Add service discovery API, auto-registration, health monitoring |
| `pkg/services/communication/service.go` | MODIFY | Add sys:* mail routing |

## Implementation Order (TDD)

1. **Test 1**: Write test → verify RED → Implement well-known ID discovery for all 12 services → verify GREEN → commit
2. **Test 2**: Write test → verify RED → Implement auto-registration with Kernel → verify GREEN → commit
3. **Test 3**: Write test → verify RED → Implement mail routing to sys:* services → verify GREEN → commit
4. **Test 4**: Write test → verify RED → Implement health monitoring with dead-letter logging → verify GREEN → commit

**Total: 4 tests, 4 commits**

## Deliverables
- ✅ Well-known ID discovery implemented (arch-v1.md L451-473)
- ✅ Runtime registration implemented (arch-v1.md L451)
- ✅ Mail routing to sys:* services implemented (arch-v1.md L477-479)
- ✅ Health monitoring implemented (arch-v1.md L459)
- ✅ 4 commits (one per test per TDD workflow)

## Recursive Analysis Decision

**Decision: Keep as-is (no split)**

**Reasoning:**
- 4 tests is ideal per CLAUDE.md (2-10 tests per phase, ideal 4-6)
- Tests cover single coherent feature: service discovery and registration
- Logical progression: discovery → registration → routing → health
- Splitting would fragment the discovery goal

**Alternative (if split needed):**
- 7.6a: Discovery + Registration - 2 tests
- 7.6b: Mail Routing + Health Monitoring - 2 tests