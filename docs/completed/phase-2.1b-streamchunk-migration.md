# Phase 2.1B: StreamChunk Migration & Integration

## Goal
Migrate `StreamChunk` from humangateway to mail package and update MailMetadata, following strict TDD workflow.

## Scope
- Move `StreamChunk` from `humangateway` to `mail` package (breaking change)
- Update `MailMetadata.Stream` field from `bool` to `*StreamChunk` (breaking change)
- Update all humangateway imports to use mail.StreamChunk
- Verify no circular dependencies are introduced
- 1 test, 1 commit

## Current Status

### Implementation Status
| Item | Status | Notes |
|------|--------|-------|
| `StreamChunk` in mail | ✅ Complete | From Phase 2.1A |
| `StreamChunk` in humangateway | ⚠️ Exists | Must be removed |
| `MailMetadata.Stream` | ⚠️ Wrong type | Currently `bool`, must be `*StreamChunk` |
| humangateway imports | ⚠️ Wrong | Currently uses local StreamChunk |

### Files Status
| File | Status |
|------|--------|
| `pkg/mail/types.go` | ✅ Complete - StreamChunk exists |
| `pkg/mail/types.go` | ⚠️ Partial - MailMetadata.Stream needs update |
| `pkg/services/humangateway/chat.go` | ⚠️ Partial - remove local StreamChunk, import from mail |
| `pkg/services/humangateway/humangateway.go` | ⚠️ Partial - update imports if needed |

## Required Implementation

### Remove StreamChunk from humangateway
```go
// pkg/services/humangateway/chat.go - REMOVE this block:
// type StreamChunk struct {
//     Data     string
//     Sequence int
//     IsFinal  bool
//     Taints   []string
// }
```

### Update humangateway imports
```go
// pkg/services/humangateway/chat.go - UPDATE imports:
import (
    "github.com/maelstrom/v3/pkg/mail"
    // ... other imports
)

// Replace all references:
// StreamChunk{...} → mail.StreamChunk{...}
// *StreamChunk → *mail.StreamChunk
```

### Update MailMetadata.Stream field
```go
// pkg/mail/types.go - UPDATE MailMetadata:
type MailMetadata struct {
    // ... existing fields ...
    Stream *StreamChunk  // CHANGE from: bool
    // ... existing fields ...
}
```

## Tests to Write (1 test, 1 commit)

### Test 1: TestMail_StreamChunkMigrationComplete
```go
func TestMail_StreamChunkMigrationComplete(t *testing.T)
```
**Acceptance Criteria:**
- `mail.StreamChunk` type exists and is usable
- `humangateway.StreamChunk` type does NOT exist (removed from chat.go)
- humangateway imports and uses `mail.StreamChunk`
- `MailMetadata.Stream` field type is `*StreamChunk` (not `bool`)
- Can assign: `metadata.Stream = &mail.StreamChunk{Data: "test"}`
- Can assign nil: `metadata.Stream = nil`
- **NO circular dependencies** between mail and humangateway packages
- All existing tests still pass

## Dependencies

### Test Dependencies
```
Test 1 (Migration) → Phase 2.1A (StreamChunk in mail exists)
```

### Phase Dependencies
- **Phase 2.1A** - Must complete first (StreamChunk type in mail package)
- **Phase 2.2** (Communication Service) - Depends on this migration completing

### Circular Dependency Check
```
mail package → NO imports from humangateway ✅
humangateway → imports mail.StreamChunk ✅
```

## Files to Create/Modify

| File | Operation | Changes |
|------|-----------|---------|
| `pkg/mail/types.go` | MODIFY | Update MailMetadata.Stream field from `bool` to `*StreamChunk` |
| `pkg/services/humangateway/chat.go` | MODIFY | Remove local StreamChunk type, import mail.StreamChunk, update all references |
| `pkg/services/humangateway/humangateway.go` | MODIFY (if needed) | Update imports if StreamChunk referenced |

## Implementation Order (TDD)

1. **Test 1**: Write test → verify RED → 
   - Update MailMetadata.Stream field type
   - Remove StreamChunk from humangateway/chat.go
   - Update humangateway to import mail.StreamChunk
   - Update all StreamChunk references in humangateway
   - Verify no circular dependencies
   → verify GREEN → commit

**Total: 1 test, 1 commit**

## Deliverables
- ✅ `StreamChunk` removed from `pkg/services/humangateway/chat.go`
- ✅ humangateway imports and uses `mail.StreamChunk`
- ✅ `MailMetadata.Stream` field updated to `*StreamChunk` type
- ✅ No circular dependencies between mail and humangateway
- ✅ 1 commit (per TDD workflow)
- ✅ Breaking change properly scoped and documented