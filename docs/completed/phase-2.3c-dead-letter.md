# Phase 2.3c: Dead-Letter Queue

## Goal
Implement dead-letter queue functionality with `LogDeadLetter()` and `QueryDeadLetters()` methods.

## Scope
- Create `DeadLetterEntry` type in `pkg/services/observability/trace.go`
- Implement `LogDeadLetter(mail, reason)` method
- Implement `QueryDeadLetters()` method
- Add dead-letter storage to ObservabilityService
- 2 tests, 2 commits

## Current Status

### Implementation Status
| Test | Status | Notes |
|------|--------|-------|
| TestObservabilityService_LogDeadLetter | ❌ Missing | Method doesn't exist |
| TestObservabilityService_QueryDeadLetters | ❌ Missing | Method doesn't exist |

### Files Status
| File | Status |
|------|--------|
| `pkg/services/observability/trace.go` | ❌ MISSING - must create with DeadLetterEntry type |
| `pkg/services/observability/service.go` | ⚠️ UPDATE - add dead-letter storage and methods |
| `pkg/services/observability/service_test.go` | ⚠️ UPDATE - needs 2 new tests |

## Required Implementation

### DeadLetterEntry Type
**File:** `pkg/services/observability/trace.go`

```go
package observability

import (
    "github.com/maelstrom/v3/pkg/mail"
    "time"
)

type DeadLetterEntry struct {
    Mail   mail.Mail   // the mail that was dead-lettered
    Reason string      // reason for dead-lettering
    Logged time.Time   // timestamp when logged
}
```

### ObservabilityService with Dead-Letter Storage
**File:** `pkg/services/observability/service.go`

```go
type ObservabilityService struct {
    mu              sync.Mutex
    traces          []services.Trace
    deadLetters     []DeadLetterEntry
}
```

### LogDeadLetter Method
```go
func (o *ObservabilityService) LogDeadLetter(mail mail.Mail, reason string) error {
    o.mu.Lock()
    defer o.mu.Unlock()
    entry := DeadLetterEntry{
        Mail:   mail,
        Reason: reason,
        Logged: time.Now(),
    }
    o.deadLetters = append(o.deadLetters, entry)
    return nil
}
```

### QueryDeadLetters Method
```go
func (o *ObservabilityService) QueryDeadLetters() ([]DeadLetterEntry, error) {
    o.mu.Lock()
    defer o.mu.Unlock()
    result := make([]DeadLetterEntry, len(o.deadLetters))
    copy(result, o.deadLetters)
    return result, nil
}
```

## Tests to Write (2 tests, 2 commits)

### Test 1: TestObservabilityService_LogDeadLetter
```go
func TestObservabilityService_LogDeadLetter(t *testing.T)
```

**Acceptance Criteria:**
- LogDeadLetter() stores mail and reason
- Logged timestamp is set (not zero time)
- Returns nil error on success

### Test 2: TestObservabilityService_QueryDeadLetters
```go
func TestObservabilityService_QueryDeadLetters(t *testing.T)
```

**Acceptance Criteria:**
- QueryDeadLetters() returns all logged entries
- Returns empty slice when no dead-letters exist
- Returns nil error on success
- Entry contains correct Mail, Reason, and Logged values

## Dependencies

### Phase Dependencies
- **Phase 2.3a** - TraceFilters type exists (no direct dependency, same phase group)

### Test Dependencies
```
Test 1 → Test 2 (must log before querying)
```

## Files to Create/Modify

| File | Operation | Changes |
|------|-----------|---------|
| `pkg/services/observability/trace.go` | CREATE | DeadLetterEntry type |
| `pkg/services/observability/service.go` | MODIFY | Add deadLetters field, LogDeadLetter(), QueryDeadLetters() |
| `pkg/services/observability/service_test.go` | MODIFY | Add 2 new tests |

## Implementation Order (TDD)

**Step 0: Create DeadLetterEntry type (no test, just type definition)**
- Create `pkg/services/observability/trace.go` with DeadLetterEntry struct
- Commit: `feat: add DeadLetterEntry type for dead-letter queue`

**Step 1-2: Implement dead-letter methods**
1. **Test 1**: Write test for LogDeadLetter → verify RED → Implement LogDeadLetter() → verify GREEN → commit
2. **Test 2**: Write test for QueryDeadLetters → verify RED → Implement QueryDeadLetters() → verify GREEN → commit

**Total: 1 type commit + 2 test commits = 3 commits**

## Deliverables
- ✅ `DeadLetterEntry` type in `pkg/services/observability/trace.go`
- ✅ `LogDeadLetter()` method implemented
- ✅ `QueryDeadLetters()` method implemented
- ✅ Dead-letter storage in ObservabilityService
- ✅ 3 commits (1 for type + one per test per TDD workflow)