# Phase G1.6: Human Gateway Chat Endpoint

**Parent**: Phase G1 (Critical Fixes)  
**Gap References**: L3-C3  
**Status**: ❌ PRIORITY

## Overview
- Tests: 2
- Commits: 2 (1:1 ratio)
- Dependencies: G1.1 (ParseAddress)

## SPEC

### Requirements
From `arch-v1.md 10.4` - Human gateway:
- Chat session UI/HTTPS endpoint must exist
- Chat session must create context for agent

### Implementation Details

**Files to create/modify**:
- `pkg/services/humangateway/service.go` - Add chat endpoint handler
- `pkg/services/humangateway/service_test.go` - Add tests

**Functions to implement**:
```go
func (s *HumanGatewayService) CreateChatSession(agentID string) (*ChatSession, error)
func (s *HumanGatewayService) ChatEndpoint(w http.ResponseWriter, r *http.Request)
```

**Test scenarios**:
1. Chat endpoint exists and responds
2. Chat session creates context

## TDD Workflow

### Iteration 1: TestHumanGatewayService_ChatEndpoint
1. Write test: `TestHumanGatewayService_ChatEndpoint` in `pkg/services/humangateway/service_test.go`
2. Run → RED
3. Implement: Add ChatEndpoint handler stub
4. Run → GREEN
5. Commit: `git commit -m "fix(gap/L3-C3): add chat endpoint handler"`

### Iteration 2: TestHumanGatewayService_ChatSession
1. Write test: `TestHumanGatewayService_ChatSession`
2. Run → RED
3. Implement: Add CreateChatSession method
4. Run → GREEN
5. Commit: `git commit -m "fix(gap/L3-C3): implement chat session creation"`

## Deliverables
- 2 commits
- All tests passing
- Files modified: `pkg/services/humangateway/service.go`, `pkg/services/humangateway/service_test.go`