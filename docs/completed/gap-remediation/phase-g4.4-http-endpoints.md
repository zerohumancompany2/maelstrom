# Phase G4.4: HTTP Endpoint Exposure

**Parent**: Phase G4 (Gateway & External APIs)  
**Gap References**: L3-C2  
**Status**: ❌ PENDING

## Overview
- Tests: 4
- Commits: 4 (1:1 ratio)
- Dependencies: G4.2 (OpenAPI Generation), G2.5 (Gateway Servers)

## SPEC

### Requirements
From `arch-v1.md 10.3` - HTTP endpoint exposure:
- Register HTTP endpoint
- Endpoint handler works
- Endpoint in OpenAPI spec
- Inner boundary not exposed

### Implementation Details

**Files to create/modify**:
- `pkg/gateway/gateway.go` - Add HTTP endpoint registration
- `pkg/gateway/gateway_test.go` - Add tests

**Functions to implement**:
```go
func (g *GatewayService) RegisterHTTPEndpoint(path string, handler http.Handler) error
func (g *GatewayService) checkBoundaryExposure(boundary BoundaryType) bool
```

**Test scenarios**:
1. Register HTTP endpoint
2. Endpoint handler works
3. Endpoint in OpenAPI spec
4. Inner boundary not exposed

## TDD Workflow

### Iteration 1: TestGatewayService_RegisterHTTPEndpoint
1. Write test: `TestGatewayService_RegisterHTTPEndpoint` in `pkg/gateway/gateway_test.go`
2. Run → RED
3. Implement: Add RegisterHTTPEndpoint method
4. Run → GREEN
5. Commit: `git commit -m "fix(gap/L3-C2): implement HTTP endpoint registration"`

### Iteration 2: TestGatewayService_HTTPEndpointHandler
1. Write test: `TestGatewayService_HTTPEndpointHandler`
2. Run → RED
3. Implement: Add endpoint handler functionality
4. Run → GREEN
5. Commit: `git commit -m "fix(gap/L3-C2): implement HTTP endpoint handler"`

### Iteration 3: TestGatewayService_OpenAPIRegistration
1. Write test: `TestGatewayService_OpenAPIRegistration`
2. Run → RED
3. Implement: Add endpoint to OpenAPI spec
4. Run → GREEN
5. Commit: `git commit -m "fix(gap/L3-C2): implement OpenAPI endpoint registration"`

### Iteration 4: TestGatewayService_BoundaryExposure
1. Write test: `TestGatewayService_BoundaryExposure`
2. Run → RED
3. Implement: Add boundary exposure check
4. Run → GREEN
5. Commit: `git commit -m "fix(gap/L3-C2): implement boundary exposure filtering"`

## Deliverables
- 4 commits
- All tests passing
- Files modified: `pkg/gateway/gateway.go`, `pkg/gateway/gateway_test.go`