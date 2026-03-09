# Phase G6.1: Deduplication

**Parent**: Phase G6 (Hot-Reload & Advanced)  
**Gap References**: L2-M1  
**Status**: ❌ PENDING

## Overview
- Tests: 4
- Commits: 4 (1:1 ratio)
- Dependencies: G2.3 (At-Least-Once Delivery)

## SPEC

### Requirements
From `arch-v1.md 9.2` - Deduplication:
- Deduplicate by correlationId
- Deduplication time window
- Old correlationIds expire
- Concurrent deduplication

### Implementation Details

**Files to create/modify**:
- `pkg/services/communication/service.go` - Add deduplication
- `pkg/services/communication/service_test.go` - Add tests

**Functions to implement**:
```go
func (s *CommunicationService) isDuplicate(correlationID string) bool
func (s *CommunicationService) expireOldCorrelationIDs(window time.Duration)
```

**Test scenarios**:
1. Deduplicate by correlationId
2. Deduplication time window
3. Old correlationIds expire
4. Concurrent deduplication

## TDD Workflow

### Iteration 1: TestCommunicationService_Deduplication
1. Write test: `TestCommunicationService_Deduplication` in `pkg/services/communication/service_test.go`
2. Run → RED
3. Implement: Add deduplication by correlationId
4. Run → GREEN
5. Commit: `git commit -m "fix(gap/L2-M1): implement deduplication by correlationId"`

### Iteration 2: TestCommunicationService_DeduplicationWindow
1. Write test: `TestCommunicationService_DeduplicationWindow`
2. Run → RED
3. Implement: Add deduplication time window
4. Run → GREEN
5. Commit: `git commit -m "fix(gap/L2-M1): implement deduplication time window"`

### Iteration 3: TestCommunicationService_DeduplicationExpiry
1. Write test: `TestCommunicationService_DeduplicationExpiry`
2. Run → RED
3. Implement: Add correlationId expiry
4. Run → GREEN
5. Commit: `git commit -m "fix(gap/L2-M1): implement correlationId expiry"`

### Iteration 4: TestCommunicationService_DeduplicationConcurrent
1. Write test: `TestCommunicationService_DeduplicationConcurrent`
2. Run → RED
3. Implement: Add concurrent deduplication support
4. Run → GREEN
5. Commit: `git commit -m "fix(gap/L2-M1): implement concurrent deduplication"`

## Deliverables
- 4 commits
- All tests passing
- Files modified: `pkg/services/communication/service.go`, `pkg/services/communication/service_test.go`