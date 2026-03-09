# Phase G1.5: CheckTaintPolicy Method

**Parent**: Phase G1 (Critical Fixes)  
**Gap References**: L2-C3  
**Status**: ❌ PRIORITY

## Overview
- Tests: 2
- Commits: 2 (1:1 ratio)
- Dependencies: G1.3 (Security Boundary Enforcement)

## SPEC

### Requirements
From `arch-v1.md 7.3` - Taint policy enforcement:
- Check policy allows transition
- Check policy denies transition
- Enforce taint-based access control

### Implementation Details

**Files to create/modify**:
- `pkg/services/security/service.go` - Add CheckTaintPolicy method
- `pkg/services/security/service_test.go` - Add tests

**Functions to implement**:
```go
func (s *SecurityService) CheckTaintPolicy(taints []string, action string) bool
```

**Test scenarios**:
1. Check policy allows transition
2. Check policy denies transition

## TDD Workflow

### Iteration 1: TestSecurityService_CheckTaintPolicy_allowed
1. Write test: `TestSecurityService_CheckTaintPolicy_allowed` in `pkg/services/security/service_test.go`
2. Run → RED
3. Implement: Add CheckTaintPolicy method with allow logic
4. Run → GREEN
5. Commit: `git commit -m "fix(gap/L2-C3): implement CheckTaintPolicy allow logic"`

### Iteration 2: TestSecurityService_CheckTaintPolicy_denied
1. Write test: `TestSecurityService_CheckTaintPolicy_denied`
2. Run → RED
3. Implement: Add CheckTaintPolicy method with deny logic
4. Run → GREEN
5. Commit: `git commit -m "fix(gap/L2-C3): implement CheckTaintPolicy deny logic"`

## Deliverables
- 2 commits
- All tests passing
- Files modified: `pkg/services/security/service.go`, `pkg/services/security/service_test.go`