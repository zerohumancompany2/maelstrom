# Phase G3.3: Dead-Letter Integration

**Parent**: Phase G3 (Integration & Wiring)  
**Gap References**: L3-M3  
**Status**: ❌ PENDING

## Overview
- Tests: 4
- Commits: 4 (1:1 ratio)
- Dependencies: G2.3 (At-Least-Once Delivery)

## SPEC

### Requirements
From `arch-v1.md 9.2` - Dead-letter queue integration:
- Failed delivery goes to dead-letter
- Query dead-letter entries
- Include failure reason
- Dead-letter after retries exhausted

### Implementation Details

**Files to create/modify**:
- `pkg/services/communication/service.go` - Add dead-letter callback
- `pkg/services/observability/service.go` - Wire dead-letter
- `pkg/services/*_test.go` - Add tests

**Functions to implement**:
```go
func (s *CommunicationService) sendToDeadLetter(mail *Mail, reason string)
func (o *ObservabilityService) QueryDeadLetters(filters *DeadLetterFilters) []*DeadLetterEntry
```

**Test scenarios**:
1. Failed delivery goes to dead-letter
2. Query dead-letter entries
3. Include failure reason
4. Dead-letter after retries exhausted

## TDD Workflow

### Iteration 1: TestCommunicationService_DeadLetterOnFailure
1. Write test: `TestCommunicationService_DeadLetterOnFailure` in `pkg/services/communication/service_test.go`
2. Run → RED
3. Implement: Add dead-letter callback on delivery failure
4. Run → GREEN
5. Commit: `git commit -m "fix(gap/L3-M3): implement dead-letter on delivery failure"`

### Iteration 2: TestObservabilityService_QueryDeadLetters
1. Write test: `TestObservabilityService_QueryDeadLetters` in `pkg/services/observability/service_test.go`
2. Run → RED
3. Implement: Add QueryDeadLetters method
4. Run → GREEN
5. Commit: `git commit -m "fix(gap/L3-M3): implement dead-letter query"`

### Iteration 3: TestCommunicationService_DeadLetterWithReason
1. Write test: `TestCommunicationService_DeadLetterWithReason`
2. Run → RED
3. Implement: Add failure reason to dead-letter entries
4. Run → GREEN
5. Commit: `git commit -m "fix(gap/L3-M3): implement dead-letter with failure reason"`

### Iteration 4: TestCommunicationService_DeadLetterAfterMaxRetries
1. Write test: `TestCommunicationService_DeadLetterAfterMaxRetries`
2. Run → RED
3. Implement: Add dead-letter after retries exhausted
4. Run → GREEN
5. Commit: `git commit -m "fix(gap/L3-M3): implement dead-letter after max retries"`

## Deliverables
- 4 commits
- All tests passing
- Files modified: `pkg/services/communication/service.go`, `pkg/services/observability/service.go`