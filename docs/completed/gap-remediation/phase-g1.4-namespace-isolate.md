# Phase G1.4: NamespaceIsolate Method

**Parent**: Phase G1 (Critical Fixes)  
**Gap References**: L2-C2  
**Status**: ❌ PRIORITY

## Overview
- Tests: 2
- Commits: 2 (1:1 ratio)
- Dependencies: G1.3 (Security Boundary Enforcement)

## SPEC

### Requirements
From `arch-v1.md 7.3` - Namespace isolation:
- Create isolated view for single agent
- Create isolated views for multiple agents
- Each agent sees only its namespace

### Implementation Details

**Files to create/modify**:
- `pkg/services/security/service.go` - Add NamespaceIsolate method
- `pkg/services/security/service_test.go` - Add tests

**Functions to implement**:
```go
func (s *SecurityService) NamespaceIsolate(data interface[], agentID string) interface{}
```

**Test scenarios**:
1. Create isolated view for single agent
2. Create isolated views for multiple agents

## TDD Workflow

### Iteration 1: TestSecurityService_NamespaceIsolate
1. Write test: `TestSecurityService_NamespaceIsolate` in `pkg/services/security/service_test.go`
2. Run → RED
3. Implement: Add NamespaceIsolate method for single agent
4. Run → GREEN
5. Commit: `git commit -m "fix(gap/L2-C2): implement NamespaceIsolate for single agent"`

### Iteration 2: TestSecurityService_NamespaceIsolate_multipleAgents
1. Write test: `TestSecurityService_NamespaceIsolate_multipleAgents`
2. Run → RED
3. Implement: Ensure NamespaceIsolate handles multiple agents
4. Run → GREEN
5. Commit: `git commit -m "fix(gap/L2-C2): implement NamespaceIsolate for multiple agents"`

## Deliverables
- 2 commits
- All tests passing
- Files modified: `pkg/services/security/service.go`, `pkg/services/security/service_test.go`