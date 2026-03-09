# Phase G4.1: TopicSubscriber Interface Fix

**Parent**: Phase G4 (Gateway & External APIs)  
**Gap References**: L3-M2  
**Status**: ❌ PENDING

## Overview
- Tests: 2
- Commits: 2 (1:1 ratio)
- Dependencies: G1.1 (ParseAddress)

## SPEC

### Requirements
From `arch-v1.md 9.3` - Topic subscriber interface:
- TopicSubscriber interface implemented correctly
- Subscribe via TopicSubscriber works

### Implementation Details

**Files to create/modify**:
- `pkg/mail/router.go` - Fix TopicSubscriber interface
- `pkg/mail/router_test.go` - Add tests

**Functions to implement**:
```go
type TopicSubscriber interface {
    Subscribe(topic string) error
    Unsubscribe(topic string) error
}
```

**Test scenarios**:
1. TopicSubscriber interface implemented
2. Subscribe via TopicSubscriber

## TDD Workflow

### Iteration 1: TestTopicSubscriber_Interface
1. Write test: `TestTopicSubscriber_Interface` in `pkg/mail/router_test.go`
2. Run → RED
3. Implement: Fix TopicSubscriber interface
4. Run → GREEN
5. Commit: `git commit -m "fix(gap/L3-M2): fix TopicSubscriber interface"`

### Iteration 2: TestTopicSubscriber_Subscribe
1. Write test: `TestTopicSubscriber_Subscribe`
2. Run → RED
3. Implement: Add Subscribe via TopicSubscriber
4. Run → GREEN
5. Commit: `git commit -m "fix(gap/L3-M2): implement Subscribe via TopicSubscriber"`

## Deliverables
- 2 commits
- All tests passing
- Files modified: `pkg/mail/router.go`, `pkg/mail/router_test.go`