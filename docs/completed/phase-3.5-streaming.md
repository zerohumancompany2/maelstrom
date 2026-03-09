# Phase 3.5: Streaming Support

## Goal
Implement streaming support with StreamChunk format and taint stripping.

## Scope
- Create `pkg/mail/stream.go` with streaming support
- Implement StreamChunk structure
- Implement UpgradeToStream method
- Implement StripForbiddenTaints function
- SSE/WS upgrade logic
- 5 tests, 5 commits

## Current Status

### Implementation Status
| Component | Status | Notes |
|-----------|--------|-------|
| `StreamChunk` | ⚠️ Defined in Layer 2 | Reuse from Layer 2 |
| `UpgradeToStream()` | ❌ Missing | SSE/WS upgrade |
| `StripForbiddenTaints()` | ❌ Missing | Taint filtering |

### Files Status
| File | Status |
|------|-------|
| `pkg/mail/stream.go` | ❌ MISSING - create |

## Required Implementation

### StreamChunk
```go
// pkg/mail/stream.go
package mail

import (
    "sync"
    "time"
)

type StreamChunk struct {
    Data     string
    Sequence int
    IsFinal  bool
    Taints   []string
}

type StreamSession struct {
    ID          string
    LastEventID *string
    Chunks      chan StreamChunk
    Closed      bool
    mu          sync.RWMutex
    CreatedAt   time.Time
}

func NewStreamSession(sessionID string, lastEventID *string) *StreamSession {
    return &StreamSession{
        ID:          sessionID,
        LastEventID: lastEventID,
        Chunks:      make(chan StreamChunk, 100),
        Closed:      false,
        CreatedAt:   time.Now(),
    }
}

func (s *StreamSession) Send(chunk StreamChunk) bool {
    s.mu.RLock()
    defer s.mu.RUnlock()
    
    if s.Closed {
        return false
    }
    
    select {
    case s.Chunks <- chunk:
        return true
    default:
        return false
    }
}

func (s *StreamSession) Close() {
    s.mu.Lock()
    defer s.mu.Unlock()
    
    if !s.Closed {
        s.Closed = true
        close(s.Chunks)
    }
}

func UpgradeToStream(sessionID string, lastEventID *string) (chan StreamChunk, error) {
    session := NewStreamSession(sessionID, lastEventID)
    return session.Chunks, nil
}
```

### StripForbiddenTaints
```go
func StripForbiddenTaints(chunk StreamChunk, allowed []string) StreamChunk {
    if len(allowed) == 0 {
        chunk.Taints = nil
        return chunk
    }
    
    allowedMap := make(map[string]bool)
    for _, a := range allowed {
        allowedMap[a] = true
    }
    
    var filtered []string
    for _, taint := range chunk.Taints {
        if allowedMap[taint] {
            filtered = append(filtered, taint)
        }
    }
    chunk.Taints = filtered
    return chunk
}
```

## Tests to Write (5 tests, 5 commits)

### Test 1: TestStreamUpgrade
```go
func TestStreamUpgrade(t *testing.T) {
    sessionID := "session-001"
    lastEventID := stringPtr("event-100")
    
    ch, err := UpgradeToStream(sessionID, lastEventID)
    if err != nil {
        t.Errorf("Expected nil error, got %v", err)
    }
    
    if ch == nil {
        t.Error("Expected non-nil channel")
    }
    
    // Verify channel is buffered
    select {
    case ch <- StreamChunk{Data: "test", Sequence: 1}:
        // Success
    default:
        t.Error("Expected channel to accept chunk without blocking")
    }
}

func stringPtr(s string) *string {
    return &s
}
```
**Acceptance Criteria:**
- UpgradeToStream returns channel
- Channel is non-nil
- sessionId is used for tracking
- Channel is buffered (non-blocking send)

### Test 2: TestStreamChunkFormat
```go
func TestStreamChunkFormat(t *testing.T) {
    chunk := StreamChunk{
        Data:     "Hello, ",
        Sequence: 1,
        IsFinal:  false,
        Taints:   []string{"USER_SUPPLIED"},
    }
    
    if chunk.Data != "Hello, " {
        t.Errorf("Expected Data 'Hello, ', got '%s'", chunk.Data)
    }
    if chunk.Sequence != 1 {
        t.Errorf("Expected Sequence 1, got %d", chunk.Sequence)
    }
    if chunk.IsFinal {
        t.Error("Expected IsFinal to be false")
    }
    if len(chunk.Taints) != 1 {
        t.Errorf("Expected 1 taint, got %d", len(chunk.Taints))
    }
    
    // Test final chunk
    finalChunk := StreamChunk{
        Data:     "world!",
        Sequence: 2,
        IsFinal:  true,
        Taints:   []string{},
    }
    
    if !finalChunk.IsFinal {
        t.Error("Expected IsFinal to be true")
    }
    if finalChunk.Sequence != 2 {
        t.Errorf("Expected Sequence 2, got %d", finalChunk.Sequence)
    }
}
```
**Acceptance Criteria:**
- StreamChunk has Data, Sequence, IsFinal, Taints fields
- Can instantiate with all fields
- Sequence increments correctly

### Test 3: TestTaintStripping
```go
func TestTaintStripping(t *testing.T) {
    chunk := StreamChunk{
        Data:   "test data",
        Taints: []string{"USER_SUPPLIED", "TOOL_OUTPUT", "INNER_BOUNDARY"},
    }
    
    // Test with allowed taints
    allowed := []string{"USER_SUPPLIED", "TOOL_OUTPUT"}
    stripped := StripForbiddenTaints(chunk, allowed)
    
    if len(stripped.Taints) != 2 {
        t.Errorf("Expected 2 taints, got %d", len(stripped.Taints))
    }
    
    // Verify correct taints preserved
    hasUserSupplied := false
    hasToolOutput := false
    for _, t := range stripped.Taints {
        if t == "USER_SUPPLIED" {
            hasUserSupplied = true
        }
        if t == "TOOL_OUTPUT" {
            hasToolOutput = true
        }
    }
    if !hasUserSupplied {
        t.Error("Expected USER_SUPPLIED to be preserved")
    }
    if !hasToolOutput {
        t.Error("Expected TOOL_OUTPUT to be preserved")
    }
    
    // Test with empty allowed list (strip all)
    chunk2 := StreamChunk{
        Data:   "test data",
        Taints: []string{"USER_SUPPLIED"},
    }
    stripped2 := StripForbiddenTaints(chunk2, []string{})
    if len(stripped2.Taints) != 0 {
        t.Errorf("Expected 0 taints with empty allowed list, got %d", len(stripped2.Taints))
    }
}
```
**Acceptance Criteria:**
- StripForbiddenTaints removes disallowed taints
- Allowed taints preserved
- Empty allowed list strips all taints

### Test 4: TestStreamChunk_IsFinal
```go
func TestStreamChunk_IsFinal(t *testing.T) {
    // Create a sequence of chunks
    chunks := []StreamChunk{
        {Data: "Part 1", Sequence: 1, IsFinal: false},
        {Data: "Part 2", Sequence: 2, IsFinal: false},
        {Data: "Part 3", Sequence: 3, IsFinal: true},
    }
    
    // Verify only last chunk is final
    for i, chunk := range chunks {
        expectedFinal := (i == len(chunks)-1)
        if chunk.IsFinal != expectedFinal {
            t.Errorf("Chunk %d: Expected IsFinal=%v, got %v", i, expectedFinal, chunk.IsFinal)
        }
    }
    
    // Test stream completion detection
    isComplete := false
    for _, chunk := range chunks {
        if chunk.IsFinal {
            isComplete = true
            break
        }
    }
    if !isComplete {
        t.Error("Expected to detect stream completion via IsFinal")
    }
}
```
**Acceptance Criteria:**
- IsFinal flag indicates last chunk
- Can be used to detect stream completion

### Test 5: TestStreamChunk_Sequence
```go
func TestStreamChunk_Sequence(t *testing.T) {
    // Test sequence uniqueness
    chunks := []StreamChunk{
        {Data: "A", Sequence: 1},
        {Data: "B", Sequence: 2},
        {Data: "C", Sequence: 3},
    }
    
    seen := make(map[int]bool)
    for _, chunk := range chunks {
        if seen[chunk.Sequence] {
            t.Errorf("Duplicate sequence number: %d", chunk.Sequence)
        }
        seen[chunk.Sequence] = true
    }
    
    // Test sequence ordering
    for i := 1; i < len(chunks); i++ {
        if chunks[i].Sequence <= chunks[i-1].Sequence {
            t.Errorf("Sequence not in order: %d <= %d", chunks[i].Sequence, chunks[i-1].Sequence)
        }
    }
    
    // Test sequence starts at 1
    if chunks[0].Sequence != 1 {
        t.Errorf("Expected first sequence to be 1, got %d", chunks[0].Sequence)
    }
}
```
**Acceptance Criteria:**
- Sequence numbers are unique
- Sequence can be used for ordering
- Sequence starts at 1

## Dependencies

### Test Dependencies
```
Test 1 → Independent
Test 2 → Test 3, Test 4, Test 5 (Chunk structure)
```

### Phase Dependencies
- **Phase 3.1** must complete first (Mail types)
- **Phase 3.5** can run in parallel with 3.2, 3.3a, 3.3b, 3.4

## Files to Create/Modify

| File | Operation | Changes |
|------|-----------|---------|
| `pkg/mail/stream.go` | CREATE | StreamChunk, StreamSession, UpgradeToStream, StripForbiddenTaints |

## Implementation Order (TDD)

1. **Test 1**: Write test → verify RED → Implement UpgradeToStream and StreamSession → verify GREEN → commit
2. **Test 2**: Write test → verify RED → Define StreamChunk structure → verify GREEN → commit
3. **Test 3**: Write test → verify RED → Implement StripForbiddenTaints → verify GREEN → commit
4. **Test 4**: Write test → verify RED → Add IsFinal flag handling → verify GREEN → commit
5. **Test 5**: Write test → verify RED → Add sequence number tracking → verify GREEN → commit

**Total: 5 tests, 5 commits**

## Deliverables
- ✅ StreamChunk in `pkg/mail/stream.go`
- ✅ StreamSession for session management
- ✅ UpgradeToStream function
- ✅ StripForbiddenTaints function
- ✅ 5 commits (one per test per TDD workflow)

## Recursive Analysis Decision

**Decision: Keep as-is (no split)**

**Reasoning:**
- 5 tests is within recommended range (2-5)
- Tests are tightly coupled around streaming functionality
- Single coherent feature: Streaming support
- Splitting would create unnecessary fragmentation

**Alternative (if split needed):**
- 3.5a: StreamChunk structure + UpgradeToStream - 2 tests
- 3.5b: Taint stripping + IsFinal/Sequence - 3 tests