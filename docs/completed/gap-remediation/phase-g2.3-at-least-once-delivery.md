# Phase G2.3: At-Least-Once Delivery

**Parent**: Phase G2 (Core Functionality)  
**Gap References**: L2-H2  
**Status**: ❌ PENDING

## Overview
- Tests: 4
- Commits: 4 (1:1 ratio)
- Dependencies: G1.1 (ParseAddress)

## SPEC

### Requirements
From `arch-v1.md 9.2` - At-least-once delivery:
- Retry delivery on failure
- Exponential backoff between retries
- Stop after max retries
- Track delivery attempts

### Implementation Details

**Files to create/modify**:
- `pkg/services/communication/service.go` - Add retry logic
- `pkg/services/communication/service_test.go` - Add tests

**Functions to implement**:
```go
func (s *CommunicationService) PublishWithRetry(mail *Mail, maxRetries int) error
func (s *CommunicationService) trackDeliveryAttempt(correlationID string)
```

**Test scenarios**:
1. Retry delivery on failure
2. Exponential backoff between retries
3. Stop after max retries
4. Track delivery attempts

## TDD Workflow

### Iteration 1: TestCommunicationService_RetryOnFailure
1. Write test: `TestCommunicationService_RetryOnFailure` in `pkg/services/communication/service_test.go`
2. Run → RED
3. Implement: Add retry logic with configurable max retries
4. Run → GREEN
5. Commit: `git commit -m "fix(gap/L2-H2): implement retry on delivery failure"`

### Iteration 2: TestCommunicationService_ExponentialBackoff
1. Write test: `TestCommunicationService_ExponentialBackoff`
2. Run → RED
3. Implement: Add exponential backoff between retries
4. Run → GREEN
5. Commit: `git commit -m "fix(gap/L2-H2): implement exponential backoff"`

### Iteration 3: TestCommunicationService_MaxRetries
1. Write test: `TestCommunicationService_MaxRetries`
2. Run → RED
3. Implement: Add max retries limit
4. Run → GREEN
5. Commit: `git commit -m "fix(gap/L2-H2): implement max retries limit"`

### Iteration 4: TestCommunicationService_DeliveryTracking
1. Write test: `TestCommunicationService_DeliveryTracking`
2. Run → RED
3. Implement: Add delivery attempt tracking
4. Run → GREEN
5. Commit: `git commit -m "fix(gap/L2-H2): implement delivery attempt tracking"`

## Deliverables
- 4 commits
- All tests passing
- Files modified: `pkg/services/communication/service.go`, `pkg/services/communication/service_test.go`