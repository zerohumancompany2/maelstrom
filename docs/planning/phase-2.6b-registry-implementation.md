# Phase 2.6B: Registry Implementation

## Goal
Implement Service Registry methods with full test coverage, following strict TDD workflow.

## Scope
- Implement `Register()` method (with duplicate detection)
- Implement `Get()` method (found and not found cases)
- Implement `List()` method (alphabetical ordering, empty case)
- Integration test with all 4 services
- 7 tests, 7 commits
- **NO breaking changes** - depends on Phase 2.6A interface fix

## Current Status

### Implementation Status
| Method | Status | Notes |
|--------|--------|-------|
| `Register()` | ❌ TODO | Method exists but returns nil unconditionally |
| `Get()` | ❌ TODO | Method exists but returns (nil, false) unconditionally |
| `List()` | ❌ TODO | Method exists but returns nil unconditionally |

### Files Status
| File | Status |
|------|--------|
| `pkg/services/registry.go` | ⚠️ Partial - all methods are TODO stubs |
| `pkg/services/registry_test.go` | ❌ MISSING - must create |

## Required Implementation

### Register Method
```go
// pkg/services/registry.go
func (sr *ServiceRegistry) Register(name string, svc Service) error {
	sr.mu.Lock()
	defer sr.mu.Unlock()
	if _, exists := sr.services[name]; exists {
		return ErrAlreadyRegistered
	}
	sr.services[name] = svc
	return nil
}
```

### Get Method
```go
// pkg/services/registry.go
func (sr *ServiceRegistry) Get(name string) (Service, bool) {
	sr.mu.RLock()
	defer sr.mu.RUnlock()
	svc, ok := sr.services[name]
	return svc, ok
}
```

### List Method
```go
// pkg/services/registry.go
func (sr *ServiceRegistry) List() []string {
	sr.mu.RLock()
	defer sr.mu.RUnlock()
	names := make([]string, 0, len(sr.services))
	for name := range sr.services {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}
```

## Tests to Write (7 tests, 7 commits)

### Test 1: TestServiceRegistry_Register
```go
func TestServiceRegistry_Register(t *testing.T)
```
**Acceptance Criteria:**
- Register() stores service by name
- Returns nil error
- Service can be retrieved after registration

### Test 2: TestServiceRegistry_RegisterDuplicate
```go
func TestServiceRegistry_RegisterDuplicate(t *testing.T)
```
**Acceptance Criteria:**
- Register() with duplicate name returns ErrAlreadyRegistered
- Original service is not overwritten

### Test 3: TestServiceRegistry_Get
```go
func TestServiceRegistry_Get(t *testing.T)
```
**Acceptance Criteria:**
- Get() returns registered service
- Returns true for found service
- Returns correct service instance

### Test 4: TestServiceRegistry_GetNotFound
```go
func TestServiceRegistry_GetNotFound(t *testing.T)
```
**Acceptance Criteria:**
- Get() with non-existent name returns (nil, false)
- No error returned (uses bool for existence check)

### Test 5: TestServiceRegistry_List
```go
func TestServiceRegistry_List(t *testing.T)
```
**Acceptance Criteria:**
- List() returns all registered service names
- Order is alphabetical (deterministic)
- Multiple services tested

### Test 6: TestServiceRegistry_ListEmpty
```go
func TestServiceRegistry_ListEmpty(t *testing.T)
```
**Acceptance Criteria:**
- List() on empty registry returns empty slice (not nil)
- No panic or error

### Test 7: TestAllServicesIntegrateViaRegistry
```go
func TestAllServicesIntegrateViaRegistry(t *testing.T)
```
**Acceptance Criteria:**
- All 4 services (security, communication, lifecycle, observability) can be registered
- All 4 services can be retrieved by their ID
- List() returns all 4 service IDs in alphabetical order
- Services implement fixed Service interface correctly

## Dependencies

### Test Dependencies
```
Test 1 (Register) → Independent
Test 2 (RegisterDuplicate) → Depends on Test 1
Test 3 (Get) → Depends on Test 1
Test 4 (GetNotFound) → Independent
Test 5 (List) → Depends on Test 1
Test 6 (ListEmpty) → Independent
Test 7 (Integration) → Depends on Tests 1-6, requires all 4 services
```

### Phase Dependencies
- **Phase 2.6A** - Service interface must be fixed first (ID, HandleMail signature)
- **Phase 2.2** - Communication Service complete
- **Phase 2.3** - Observability Service complete
- **Phase 2.4** - Lifecycle Service complete
- **Phase 2.5** - Security Service complete

## Files to Create/Modify

| File | Operation | Changes |
|------|-----------|---------|
| `pkg/services/registry.go` | MODIFY | Implement Register(), Get(), List() with thread-safety |
| `pkg/services/registry_test.go` | CREATE | Add 7 tests |

## Implementation Order (TDD)

1. **Test 1**: Write test → verify RED → Implement Register() → verify GREEN → commit
2. **Test 2**: Write test → verify RED → Add duplicate check to Register() → verify GREEN → commit
3. **Test 3**: Write test → verify RED → Implement Get() → verify GREEN → commit
4. **Test 4**: Write test → verify RED → Handle not found case in Get() → verify GREEN → commit
5. **Test 5**: Write test → verify RED → Implement List() with sorting → verify GREEN → commit
6. **Test 6**: Write test → verify RED → Handle empty list case → verify GREEN → commit
7. **Test 7**: Write test → verify RED → Register all 4 services → verify GREEN → commit

**Total: 7 tests, 7 commits**

## Deliverables
- ✅ `Register()` implemented (with duplicate detection)
- ✅ `Get()` implemented (found and not found cases)
- ✅ `List()` implemented (alphabetical ordering, empty case)
- ✅ All 4 services integrate via registry
- ✅ 7 commits (one per test per TDD workflow)
- ✅ Thread-safe operations with RWMutex