# Phase 2.6: Service Registry & Integration

## Goal
Implement Service Registry and verify all services integrate correctly.

## Scope
- Implement `Register()` method
- Implement `Get()` method
- Implement `List()` method
- Add integration tests for all services
- 5 tests, 5 commits

## Current Status

### Implementation Status
| Test | Status | Notes |
|------|--------|-------|
| TestServiceRegistry_Register | ❌ Missing | Method is TODO |
| TestServiceRegistry_Get | ❌ Missing | Method is TODO |
| TestServiceRegistry_List | ❌ Missing | Method is TODO |
| TestAllServicesIntegrateViaRegistry | ❌ Missing | Integration test |

### Files Status
| File | Status |
|------|--------|
| `pkg/services/registry.go` | ⚠️ Partial - all methods are TODO |
| `pkg/services/registry_test.go` | ❌ MISSING - must create |

## Required Implementation

### Register Method
```go
// pkg/services/registry.go
func (r *ServiceRegistry) Register(name string, svc Service) error {
    // Register service by name
}
```

### Get Method
```go
// pkg/services/registry.go
func (r *ServiceRegistry) Get(name string) (Service, bool) {
    // Get service by name
}
```

### List Method
```go
// pkg/services/registry.go
func (r *ServiceRegistry) List() []string {
    // List all service names
}
```

## Tests to Write (5 tests, 5 commits)

### Test 1: TestServiceRegistry_Register
```go
func TestServiceRegistry_Register(t *testing.T)
```
**Acceptance Criteria:**
- Register() stores service by name
- Returns nil error

### Test 2: TestServiceRegistry_Get
```go
func TestServiceRegistry_Get(t *testing.T)
```
**Acceptance Criteria:**
- Get() returns registered service
- Returns true for found service

### Test 3: TestServiceRegistry_List
```go
func TestServiceRegistry_List(t *testing.T)
```
**Acceptance Criteria:**
- List() returns all registered service names
- Order is deterministic

### Test 4: TestServiceRegistry_GetNotFound
```go
func TestServiceRegistry_GetNotFound(t *testing.T)
```
**Acceptance Criteria:**
- Get() with non-existent name returns (nil, false)
- No error returned

### Test 5: TestAllServicesIntegrateViaRegistry
```go
func TestAllServicesIntegrateViaRegistry(t *testing.T)
```
**Acceptance Criteria:**
- All 4 services can be registered
- All 4 services can be retrieved
- List() returns all 4 service names

## Dependencies

### Test Dependencies
```
Test 1 → Test 2, Test 3, Test 4 (Registry operations)
Test 5 (Integration - all services)
```

### Phase Dependencies
- **Phase 2.2** - Communication Service complete
- **Phase 2.3** - Observability Service complete
- **Phase 2.4** - Lifecycle Service complete
- **Phase 2.5** - Security Service complete

## Files to Create/Modify

| File | Operation | Changes |
|------|-----------|---------|
| `pkg/services/registry.go` | MODIFY | Implement Register(), Get(), List() |
| `pkg/services/registry_test.go` | CREATE | Add 5 tests |

## Implementation Order (TDD)

1. **Test 1**: Write test → verify RED → Implement Register() → verify GREEN → commit
2. **Test 2**: Write test → verify RED → Implement Get() → verify GREEN → commit
3. **Test 3**: Write test → verify RED → Implement List() → verify GREEN → commit
4. **Test 4**: Write test → verify RED → Handle not found in Get() → verify GREEN → commit
5. **Test 5**: Write test → verify RED → Register all 4 services → verify GREEN → commit

**Total: 5 tests, 5 commits**

## Deliverables
- ✅ `Register()` implemented
- ✅ `Get()` implemented
- ✅ `List()` implemented
- ✅ All 4 services integrate via registry
- ✅ 5 commits (one per test per TDD workflow)