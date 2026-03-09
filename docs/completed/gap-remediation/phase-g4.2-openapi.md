# Phase G4.2: OpenAPI Generation

**Parent**: Phase G4 (Gateway & External APIs)  
**Gap References**: L3-C2  
**Status**: ❌ PENDING

## Overview
- Tests: 4
- Commits: 4 (1:1 ratio)
- Dependencies: G2.5 (Gateway Servers)

## SPEC

### Requirements
From `arch-v1.md 10.3` - OpenAPI generation:
- Generate OpenAPI from chart YAML
- Generate HTTP endpoint specs
- Auth middleware in spec
- Only DMZ/outer exposed

### Implementation Details

**Files to create/modify**:
- `pkg/gateway/openapi.go` - Add OpenAPI generation
- `pkg/gateway/openapi_test.go` - Add tests

**Functions to implement**:
```go
func GenerateOpenAPIFromChart(chart *Chart) (*OpenAPISpec, error)
func GenerateHTTPExposureSpec(chart *Chart) (*HTTPSpec, error)
```

**Test scenarios**:
1. Generate OpenAPI from chart YAML
2. Generate HTTP endpoint specs
3. Auth middleware in spec
4. Only DMZ/outer exposed

## TDD Workflow

### Iteration 1: TestOpenAPI_GenerateFromChart
1. Write test: `TestOpenAPI_GenerateFromChart` in `pkg/gateway/openapi_test.go`
2. Run → RED
3. Implement: Add OpenAPI generation from charts
4. Run → GREEN
5. Commit: `git commit -m "fix(gap/L3-C2): implement OpenAPI generation from chart"`

### Iteration 2: TestOpenAPI_HTTPExposure
1. Write test: `TestOpenAPI_HTTPExposure`
2. Run → RED
3. Implement: Add HTTP endpoint spec generation
4. Run → GREEN
5. Commit: `git commit -m "fix(gap/L3-C2): implement HTTP exposure spec"`

### Iteration 3: TestOpenAPI_AuthMiddleware
1. Write test: `TestOpenAPI_AuthMiddleware`
2. Run → RED
3. Implement: Add auth middleware to spec
4. Run → GREEN
5. Commit: `git commit -m "fix(gap/L3-C2): implement auth middleware in OpenAPI"`

### Iteration 4: TestOpenAPI_BoundaryFiltering
1. Write test: `TestOpenAPI_BoundaryFiltering`
2. Run → RED
3. Implement: Add boundary filtering (DMZ/outer only)
4. Run → GREEN
5. Commit: `git commit -m "fix(gap/L3-C2): implement boundary filtering in OpenAPI"`

## Deliverables
- 4 commits
- All tests passing
- Files modified: `pkg/gateway/openapi.go`, `pkg/gateway/openapi_test.go`