# Phase G2.1: Address Validation Helpers

**Parent**: Phase G2 (Core Functionality)  
**Gap References**: L3-H1  
**Status**: ❌ PENDING

## Overview
- Tests: 3
- Commits: 3 (1:1 ratio)
- Dependencies: G1.1 (ParseAddress)

## SPEC

### Requirements
From `arch-v1.md 9.2` - Mail address validation:
- Validate agent:id format
- Validate topic:name format
- Validate sys:service format

### Implementation Details

**Files to create/modify**:
- `pkg/mail/address.go` - Add validation helper functions
- `pkg/mail/address_test.go` - Add tests

**Functions to implement**:
```go
func IsValidAgentAddress(address string) bool
func IsValidTopicAddress(address string) bool
func IsValidSysAddress(address string) bool
```

**Test scenarios**:
1. Validate agent:id format
2. Validate topic:name format
3. Validate sys:service format

## TDD Workflow

### Iteration 1: TestIsValidAgentAddress
1. Write test: `TestIsValidAgentAddress` in `pkg/mail/address_test.go`
2. Run → RED
3. Implement: Add IsValidAgentAddress function
4. Run → GREEN
5. Commit: `git commit -m "fix(gap/L3-H1): add IsValidAgentAddress validation"`

### Iteration 2: TestIsValidTopicAddress
1. Write test: `TestIsValidTopicAddress`
2. Run → RED
3. Implement: Add IsValidTopicAddress function
4. Run → GREEN
5. Commit: `git commit -m "fix(gap/L3-H1): add IsValidTopicAddress validation"`

### Iteration 3: TestIsValidSysAddress
1. Write test: `TestIsValidSysAddress`
2. Run → RED
3. Implement: Add IsValidSysAddress function
4. Run → GREEN
5. Commit: `git commit -m "fix(gap/L3-H1): add IsValidSysAddress validation"`

## Deliverables
- 3 commits
- All tests passing
- Files modified: `pkg/mail/address.go`, `pkg/mail/address_test.go`