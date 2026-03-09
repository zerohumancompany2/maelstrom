# Phase G5.1: Mail Metadata Type Fix

**Parent**: Phase G5 (Observability & Metrics)  
**Gap References**: L3-M1  
**Status**: ❌ PENDING

## Overview
- Tests: 3
- Commits: 3 (1:1 ratio)
- Dependencies: None

## SPEC

### Requirements
From `arch-v1.md 9.2` - Mail metadata type alignment:
- Stream field is boolean
- Stream chunk separate field
- All metadata fields match spec

### Implementation Details

**Files to create/modify**:
- `pkg/mail/types.go` - Fix Mail.Metadata.Stream type
- `pkg/mail/types_test.go` - Add tests

**Types to fix**:
```go
type MailMetadata struct {
    Stream     bool         // Changed from *StreamChunk
    StreamChunk *StreamChunk // New separate field
    // ... other fields
}
```

**Test scenarios**:
1. Stream field is boolean
2. Stream chunk separate field
3. All metadata fields match spec

## TDD Workflow

### Iteration 1: TestMailMetadata_StreamBool
1. Write test: `TestMailMetadata_StreamBool` in `pkg/mail/types_test.go`
2. Run → RED
3. Implement: Change Stream field to bool
4. Run → GREEN
5. Commit: `git commit -m "fix(gap/L3-M1): fix Mail.Metadata.Stream to bool"`

### Iteration 2: TestMailMetadata_StreamChunk
1. Write test: `TestMailMetadata_StreamChunk`
2. Run → RED
3. Implement: Add StreamChunk separate field
4. Run → GREEN
5. Commit: `git commit -m "fix(gap/L3-M1): add Mail.Metadata.StreamChunk field"`

### Iteration 3: TestMailMetadata_TypeAlignment
1. Write test: `TestMailMetadata_TypeAlignment`
2. Run → RED
3. Implement: Verify all metadata fields match spec
4. Run → GREEN
5. Commit: `git commit -m "fix(gap/L3-M1): verify Mail.Metadata type alignment"`

## Deliverables
- 3 commits
- All tests passing
- Files modified: `pkg/mail/types.go`, `pkg/mail/types_test.go`