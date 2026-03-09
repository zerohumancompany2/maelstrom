# Phase G1.1: ParseAddress Implementation

**Parent**: Phase G1 (Critical Fixes)  
**Gap References**: L3-H4  
**Status**: ❌ PRIORITY

## Overview
- Tests: 4
- Commits: 4 (1:1 ratio)
- Dependencies: None

## SPEC

### Requirements
From `arch-v1.md 9.2` - Mail addressing system:
- Parse agent:id format addresses
- Parse topic:name format addresses  
- Parse sys:service format addresses
- Return error for invalid formats

### Implementation Details

**Files to create/modify**:
- `pkg/mail/router.go` - Add ParseAddress function
- `pkg/mail/router_test.go` - Add tests

**Functions to implement**:
```go
func ParseAddress(address string) (AddressType, string, error)
```

**AddressType enum**:
```go
type AddressType int

const (
    AgentAddress AddressType = iota
    TopicAddress
    SysAddress
)
```

**Test scenarios**:
1. Parse agent:id format
2. Parse topic:name format
3. Parse sys:service format
4. Parse invalid format returns error

## TDD Workflow

### Iteration 1: TestParseAddress_agent
1. Write test: `TestParseAddress_agent` in `pkg/mail/router_test.go`
2. Run → RED
3. Implement: Add AddressType constants and ParseAddress function with agent: support
4. Run → GREEN
5. Commit: `git commit -m "fix(gap/L3-H4): add ParseAddress for agent: format"`

### Iteration 2: TestParseAddress_topic
1. Write test: `TestParseAddress_topic`
2. Run → RED
3. Implement: Add topic: format parsing
4. Run → GREEN
5. Commit: `git commit -m "fix(gap/L3-H4): add ParseAddress for topic: format"`

### Iteration 3: TestParseAddress_sys
1. Write test: `TestParseAddress_sys`
2. Run → RED
3. Implement: Add sys: format parsing
4. Run → GREEN
5. Commit: `git commit -m "fix(gap/L3-H4): add ParseAddress for sys: format"`

### Iteration 4: TestParseAddress_invalid
1. Write test: `TestParseAddress_invalid`
2. Run → RED
3. Implement: Add error handling for invalid formats
4. Run → GREEN
5. Commit: `git commit -m "fix(gap/L3-H4): add ParseAddress error handling"`

## Deliverables
- 4 commits
- All tests passing
- Files modified: `pkg/mail/router.go`, `pkg/mail/router_test.go`