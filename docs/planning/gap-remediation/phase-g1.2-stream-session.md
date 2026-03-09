# Phase G1.2: StreamSession Send/Close Implementation

**Parent**: Phase G1 (Critical Fixes)  
**Gap References**: L3-C1  
**Status**: ❌ PRIORITY

## Overview
- Tests: 4
- Commits: 4 (1:1 ratio)
- Dependencies: G1.1 (ParseAddress)

## SPEC

### Requirements
From `arch-v1.md 10.2` - Streaming protocol:
- `StreamSession.Send()` must write chunks to channel without panicking
- `StreamSession.Close()` must close channel and signal completion
- Multiple chunks must be sent in order
- Close after sends must complete cleanly

### Implementation Details

**Files to create/modify**:
- `pkg/mail/stream.go` - Implement Send() and Close() methods
- `pkg/mail/stream_test.go` - Add tests

**Functions to implement**:
```go
func (s *StreamSession) Send(chunk *StreamChunk) error
func (s *StreamSession) Close() error
```

**Test scenarios**:
1. Send single chunk to channel
2. Send multiple chunks in order
3. Close session and signal completion
4. Close after sends completes cleanly

## TDD Workflow

### Iteration 1: TestStreamSession_Send
1. Write test: `TestStreamSession_Send` in `pkg/mail/stream_test.go`
2. Run → RED
3. Implement: Add Send() method that writes chunk to channel
4. Run → GREEN
5. Commit: `git commit -m "fix(gap/L3-C1): implement StreamSession.Send"`

### Iteration 2: TestStreamSession_SendMultiple
1. Write test: `TestStreamSession_SendMultiple`
2. Run → RED
3. Implement: Ensure Send() handles multiple chunks in order
4. Run → GREEN
5. Commit: `git commit -m "fix(gap/L3-C1): implement StreamSession.Send for multiple chunks"`

### Iteration 3: TestStreamSession_Close
1. Write test: `TestStreamSession_Close`
2. Run → RED
3. Implement: Add Close() method that closes channel
4. Run → GREEN
5. Commit: `git commit -m "fix(gap/L3-C1): implement StreamSession.Close"`

### Iteration 4: TestStreamSession_CloseAfterSend
1. Write test: `TestStreamSession_CloseAfterSend`
2. Run → RED
3. Implement: Ensure Close() after sends completes cleanly
4. Run → GREEN
5. Commit: `git commit -m "fix(gap/L3-C1): implement StreamSession.Close after sends"`

## Deliverables
- 4 commits
- All tests passing
- Files modified: `pkg/mail/stream.go`, `pkg/mail/stream_test.go`