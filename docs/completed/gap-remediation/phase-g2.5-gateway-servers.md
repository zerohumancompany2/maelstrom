# Phase G2.5: Gateway Servers

**Parent**: Phase G2 (Core Functionality)  
**Gap References**: L3-H3  
**Status**: ❌ PENDING

## Overview
- Tests: 4
- Commits: 4 (1:1 ratio)
- Dependencies: G1.1 (ParseAddress)

## SPEC

### Requirements
From `arch-v1.md 10.1` - Gateway adapters:
- HTTP server for webhook adapter
- WebSocket server for websocket adapter
- SSE endpoint for sse adapter
- Inbound/outbound normalization

### Implementation Details

**Files to create/modify**:
- `pkg/gateway/adapters/webhook.go` - Add HTTP server
- `pkg/gateway/adapters/websocket.go` - Add WebSocket server
- `pkg/gateway/adapters/sse.go` - Add SSE endpoint
- `pkg/gateway/adapters/*_test.go` - Add tests

**Functions to implement**:
```go
func (a *WebhookAdapter) StartServer(addr string) error
func (a *WebSocketAdapter) StartServer(addr string) error
func (a *SSEAdapter) StartServer(addr string) error
```

**Test scenarios**:
1. HTTP server starts and responds
2. WebSocket server accepts connections
3. SSE endpoint streams events
4. Inbound/outbound normalization works

## TDD Workflow

### Iteration 1: TestWebhookAdapter_HTTPServer
1. Write test: `TestWebhookAdapter_HTTPServer` in `pkg/gateway/adapters/webhook_test.go`
2. Run → RED
3. Implement: Add HTTP server to WebhookAdapter
4. Run → GREEN
5. Commit: `git commit -m "fix(gap/L3-H3): implement webhook HTTP server"`

### Iteration 2: TestWebSocketAdapter_WSConnection
1. Write test: `TestWebSocketAdapter_WSConnection` in `pkg/gateway/adapters/websocket_test.go`
2. Run → RED
3. Implement: Add WebSocket server to WebSocketAdapter
4. Run → GREEN
5. Commit: `git commit -m "fix(gap/L3-H3): implement websocket server"`

### Iteration 3: TestSSEAdapter_SSEEndpoint
1. Write test: `TestSSEAdapter_SSEEndpoint` in `pkg/gateway/adapters/sse_test.go`
2. Run → RED
3. Implement: Add SSE endpoint to SSEAdapter
4. Run → GREEN
5. Commit: `git commit -m "fix(gap/L3-H3): implement SSE endpoint"`

### Iteration 4: TestGatewayAdapter_Normalization
1. Write test: `TestGatewayAdapter_Normalization`
2. Run → RED
3. Implement: Add inbound/outbound normalization
4. Run → GREEN
5. Commit: `git commit -m "fix(gap/L3-H3): implement gateway normalization"`

## Deliverables
- 4 commits
- All tests passing
- Files modified: `pkg/gateway/adapters/webhook.go`, `pkg/gateway/adapters/websocket.go`, `pkg/gateway/adapters/sse.go`