# Phase G3.4: Stream Taint Integration

**Parent**: Phase G3 (Integration & Wiring)  
**Gap References**: L3-M4  
**Status**: ❌ PENDING

## Overview
- Tests: 3
- Commits: 3 (1:1 ratio)
- Dependencies: G1.3 (Security Boundary Enforcement), G1.2 (StreamSession)

## SPEC

### Requirements
From `arch-v1.md 10.2` - Stream taint stripping:
- Taints stripped in stream flow
- Taints propagated correctly
- Security boundary check in stream

### Implementation Details

**Files to create/modify**:
- `pkg/mail/stream.go` - Integrate taint stripping
- `pkg/mail/stream_test.go` - Add tests

**Functions to implement**:
```go
func (s *StreamSession) stripTaints(chunk *StreamChunk) *StreamChunk
func (s *StreamSession) checkSecurityBoundary(from, to BoundaryType) error
```

**Test scenarios**:
1. Taints stripped in stream flow
2. Taints propagated correctly
3. Security boundary check in stream

## TDD Workflow

### Iteration 1: TestMailStream_TaintStripping
1. Write test: `TestMailStream_TaintStripping` in `pkg/mail/stream_test.go`
2. Run → RED
3. Implement: Add taint stripping in stream flow
4. Run → GREEN
5. Commit: `git commit -m "fix(gap/L3-M4): implement taint stripping in stream"`

### Iteration 2: TestMailStream_TaintPropagation
1. Write test: `TestMailStream_TaintPropagation`
2. Run → RED
3. Implement: Add taint propagation in stream
4. Run → GREEN
5. Commit: `git commit -m "fix(gap/L3-M4): implement taint propagation in stream"`

### Iteration 3: TestMailStream_SecurityCheck
1. Write test: `TestMailStream_SecurityCheck`
2. Run → RED
3. Implement: Add security boundary check in stream
4. Run → GREEN
5. Commit: `git commit -m "fix(gap/L3-M4): implement security check in stream"`

## Deliverables
- 3 commits
- All tests passing
- Files modified: `pkg/mail/stream.go`, `pkg/mail/stream_test.go`