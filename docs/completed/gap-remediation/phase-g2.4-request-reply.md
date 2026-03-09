# Phase G2.4: Request-Reply Pattern

**Parent**: Phase G2 (Core Functionality)  
**Gap References**: L3-H2  
**Status**: ❌ PENDING

## Overview
- Tests: 4
- Commits: 4 (1:1 ratio)
- Dependencies: G2.3 (At-Least-Once Delivery)

## SPEC

### Requirements
From `arch-v1.md 9.2` - Request-reply pattern:
- Send request and receive reply
- Match reply to request via correlationId
- Timeout on no reply
- Handle multiple concurrent requests

### Implementation Details

**Files to create/modify**:
- `pkg/services/communication/service.go` - Add request-reply tracking
- `pkg/services/communication/service_test.go` - Add tests

**Functions to implement**:
```go
func (s *CommunicationService) Request(replyChan chan *Mail, timeout time.Duration) (*Mail, error)
func (s *CommunicationService) matchReply(correlationID string, mail *Mail) bool
```

**Test scenarios**:
1. Send request and receive reply
2. Match reply to request via correlationId
3. Timeout on no reply
4. Handle multiple concurrent requests

## TDD Workflow

### Iteration 1: TestCommunicationService_RequestReply
1. Write test: `TestCommunicationService_RequestReply` in `pkg/services/communication/service_test.go`
2. Run → RED
3. Implement: Add Request method with correlationId tracking
4. Run → GREEN
5. Commit: `git commit -m "fix(gap/L3-H2): implement request-reply pattern"`

### Iteration 2: TestCommunicationService_CorrelationIdMatching
1. Write test: `TestCommunicationService_CorrelationIdMatching`
2. Run → RED
3. Implement: Add correlationId matching logic
4. Run → GREEN
5. Commit: `git commit -m "fix(gap/L3-H2): implement correlationId matching"`

### Iteration 3: TestCommunicationService_RequestTimeout
1. Write test: `TestCommunicationService_RequestTimeout`
2. Run → RED
3. Implement: Add timeout handling
4. Run → GREEN
5. Commit: `git commit -m "fix(gap/L3-H2): implement request timeout"`

### Iteration 4: TestCommunicationService_MultipleRequests
1. Write test: `TestCommunicationService_MultipleRequests`
2. Run → RED
3. Implement: Ensure multiple concurrent requests handled
4. Run → GREEN
5. Commit: `git commit -m "fix(gap/L3-H2): implement concurrent request handling"`

## Deliverables
- 4 commits
- All tests passing
- Files modified: `pkg/services/communication/service.go`, `pkg/services/communication/service_test.go`