# Phase G1.3: Security Boundary Enforcement

**Parent**: Phase G1 (Critical Fixes)  
**Gap References**: L2-C1  
**Status**: ❌ PRIORITY

## Overview
- Tests: 2
- Commits: 2 (1:1 ratio)
- Dependencies: None

## SPEC

### Requirements
From `arch-v1.md 7.1, 14.2` - Security boundary enforcement:
- `ValidateAndSanitize` must validate boundary transitions
- Inner→Outer transitions must be allowed
- Outer→Inner transitions must add taints

### Implementation Details

**Files to create/modify**:
- `pkg/services/security/service.go` - Implement ValidateAndSanitize
- `pkg/services/security/service_test.go` - Add tests

**Functions to implement**:
```go
func (s *SecurityService) ValidateAndSanitize(data interface{}, from, to BoundaryType) (interface[], error)
```

**Test scenarios**:
1. Validate inner→outer transition (allowed, no taints)
2. Validate outer→inner transition (allowed, adds taints)

## TDD Workflow

### Iteration 1: TestSecurityService_ValidateAndSanitize_innerToOuter
1. Write test: `TestSecurityService_ValidateAndSanitize_innerToOuter` in `pkg/services/security/service_test.go`
2. Run → RED
3. Implement: Add boundary validation logic for inner→outer
4. Run → GREEN
5. Commit: `git commit -m "fix(gap/L2-C1): implement boundary enforcement inner→outer"`

### Iteration 2: TestSecurityService_ValidateAndSanitize_outerToInner
1. Write test: `TestSecurityService_ValidateAndSanitize_outerToInner`
2. Run → RED
3. Implement: Add boundary validation logic for outer→inner with taints
4. Run → GREEN
5. Commit: `git commit -m "fix(gap/L2-C1): implement boundary enforcement outer→inner with taints"`

## Deliverables
- 2 commits
- All tests passing
- Files modified: `pkg/services/security/service.go`, `pkg/services/security/service_test.go`