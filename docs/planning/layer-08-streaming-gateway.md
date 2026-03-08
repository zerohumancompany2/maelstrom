# Layer 8: Streaming & Gateway

**Parent Scope:** `implementation-scope.md`  
**Dependencies:** Layers 1-7  
**Status:** Planning Phase

---

## Overview

Layer 8 provides the external interface layer of Maelstrom, exposing functionality through multiple communication protocols and streaming interfaces. This layer handles all external communication, acting as the gateway between Maelstrom and the outside world.

### Core Capabilities

- **REST API**: Full RESTful interface for all operations
- **WebSocket**: Real-time bidirectional communication
- **Event Streaming**: Server-Sent Events (SSE) and event streams
- **Message Queues**: Kafka, RabbitMQ integration
- **GraphQL**: Alternative query interface
- **API Gateway**: Rate limiting, authentication, routing
- **Health & Metrics**: Monitoring and observability

---

## Dependencies

### Required from Layer 7

- `pkg/agent/agent.go` - Agent creation and management
- `pkg/agent/team.go` - Team orchestration
- `pkg/llm/interface.go` - LLM model access
- `pkg/agent/memory.go` - Memory operations
- `pkg/agent/learning.go` - Learning engine

### Required from Layer 6

- `pkg/orchestration/workflow_engine.go` - Workflow execution
- `pkg/orchestration/tool_registry.go` - Tool access
- `pkg/orchestration/context_manager.go` - Context management

### Required from Layer 5

- `pkg/platform/eventbus.go` - Event bus for streaming
- `pkg/platform/cache.go` - Response caching
- `pkg/platform/scheduler.go` - Scheduled tasks

### Required from Layer 4

- `pkg/security/auth.go` - Authentication middleware
- `pkg/security/authz.go` - Authorization checks
- `pkg/security/datasource.go` - Secure data access

### Required from Layer 1

- `pkg/kernel/bootstrap.go` - Bootstrap context
- `pkg/kernel/config.go` - Configuration access

---

## Detailed Specifications

### 8.1 REST API Layer

#### 8.1.1 API Design Principles

- **RESTful**: Resource-based design with proper HTTP methods
- **Versioned**: URL-based versioning (`/api/v1/`)
- **Consistent**: Uniform error responses, pagination, filtering
- **Documented**: OpenAPI/Swagger specifications
- **Authenticated**: All endpoints require authentication

#### 8.1.2 Core Endpoints

```go
// pkg/gateway/api/agent.go

package api

// Agent endpoints
GET    /api/v1/agents              - List all agents
POST   /api/v1/agents              - Create new agent
GET    /api/v1/agents/{id}         - Get agent details
PUT    /api/v1/agents/{id}         - Update agent
DELETE /api/v1/agents/{id}         - Delete agent
POST   /api/v1/agents/{id}/start   - Start agent
POST   /api/v1/agents/{id}/stop    - Stop agent
POST   /api/v1/agents/{id}/pause   - Pause agent

// Team endpoints
GET    /api/v1/teams               - List all teams
POST   /api/v1/teams               - Create team
GET    /api/v1/teams/{id}          - Get team details
POST   /api/v1/teams/{id}/messages - Send message to team

// LLM endpoints
POST   /api/v1/llm/generate        - Generate text
POST   /api/v1/llm/generate/stream - Stream generation
GET    /api/v1/llm/models          - List available models
POST   /api/v1/llm/embed           - Generate embeddings

// Memory endpoints
GET    /api/v1/memory              - List memory items
POST   /api/v1/memory              - Store memory
GET    /api/v1/memory/search       - Search memory
DELETE /api/v1/memory/{id}         - Delete memory

// Workflow endpoints
POST   /api/v1/workflows           - Execute workflow
GET    /api/v1/workflows/{id}      - Get workflow status
GET    /api/v1/workflows/{id}/logs - Get workflow logs
```

#### 8.1.3 Request/Response Format

```go
// pkg/gateway/api/types.go

type Agent struct {
    ID        string                 `json:"id"`
    Name      string                 `json:"name"`
    State     string                 `json:"state"`
    Config    AgentConfig            `json:"config"`
    CreatedAt time.Time              `json:"created_at"`
    Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

type AgentConfig struct {
    Model       string  `json:"model"`
    Temperature float64 `json:"temperature"`
    MaxTokens   int     `json:"max_tokens"`
}

type APIError struct {
    Code    string `json:"code"`
    Message string `json:"message"`
    Details string `json:"details,omitempty"`
}

type PaginatedResponse struct {
    Data     interface{} `json:"data"`
    Page     int         `json:"page"`
    PageSize int         `json:"page_size"`
    Total    int         `json:"total"`
}
```

---

### 8.2 WebSocket Layer

#### 8.2.1 Connection Management

```go
// pkg/gateway/ws/handler.go

type WebSocketHandler struct {
    Upgrader websocket.Upgrader
    AuthFunc func(*websocket.Conn) (string, error)
}

type WSMessage struct {
    Type      string      `json:"type"`
    Payload   interface{} `json:"payload"`
    Timestamp time.Time   `json:"timestamp"`
}

type WSEvent struct {
    Event   string      `json:"event"`
    Data    interface{} `json:"data"`
    AgentID string      `json:"agent_id,omitempty"`
}
```

#### 8.2.2 Message Types

- **subscribe**: Subscribe to agent/team events
- **unsubscribe**: Unsubscribe from events
- **send**: Send message to agent/team
- **command**: Execute command on agent
- **response**: Response to previous request
- **event**: Real-time event notification

#### 8.2.3 Connection Lifecycle

1. **Handshake**: WebSocket upgrade with authentication
2. **Authenticated**: Connection verified and authorized
3. **Active**: Processing messages and events
4. **Closing**: Graceful shutdown
5. **Closed**: Connection terminated

---

### 8.3 Event Streaming

#### 8.3.1 Server-Sent Events (SSE)

```go
// pkg/gateway/sse/stream.go

type SSEStream struct {
    Writer http.ResponseWriter
    Encoder *json.Encoder
}

func (s *SSEStream) Send(event string, data interface{}) error
func (s *SSEStream) Close() error
```

#### 8.3.2 Stream Types

- **AgentEvents**: Real-time agent state changes
- **WorkflowEvents**: Workflow execution progress
- **LogStream**: Agent/workflow logs
- **MemoryUpdates**: Memory changes
- **LLMStream**: LLM generation streaming

#### 8.3.3 Reconnection Support

- **Event IDs**: Track last received event
- **Automatic Retry**: Client-side reconnection
- **State Recovery**: Resume from last event

---

### 8.4 Message Queue Integration

#### 8.4.1 Queue Types

- **agent-commands**: Commands to agents
- **agent-events**: Agent-generated events
- **workflow-queue**: Workflow execution requests
- **memory-updates**: Memory synchronization
- **llm-requests**: LLM processing queue

#### 8.4.2 Queue Configuration

```go
// pkg/gateway/queue/config.go

type QueueConfig struct {
    Name        string
    Type        QueueType // kafka, rabbitmq, redis
    BrokerURL   string
    Durable     bool
    AutoDelete  bool
    MaxMessages int
}

type QueueType string

const (
    QueueKafka    QueueType = "kafka"
    QueueRabbitMQ QueueType = "rabbitmq"
    QueueRedis    QueueType = "redis"
)
```

#### 8.4.3 Producer/Consumer Pattern

- **Producers**: Gateway publishes to queues
- **Consumers**: Agents/services consume from queues
- **Acknowledgment**: Confirm message processing
- **Retry**: Handle processing failures

---

### 8.5 GraphQL Interface

#### 8.5.1 Schema Design

```graphql
type Agent {
  id: ID!
  name: String!
  state: AgentState!
  config: AgentConfig!
  memory: [MemoryItem!]
  team: Team
  createdAt: DateTime!
}

type Team {
  id: ID!
  name: String!
  leader: Agent!
  agents: [Agent!]!
  protocol: TeamProtocol!
}

type LLMModel {
  id: ID!
  name: String!
  provider: String!
  config: ModelConfig!
}

type MemoryItem {
  id: ID!
  content: String!
  embedding: [Float!]
  metadata: JSON
  createdAt: DateTime!
}

type Query {
  agents: [Agent!]!
  agent(id: ID!): Agent
  teams: [Team!]!
  team(id: ID!): Team
  models: [LLMModel!]!
  memory(search: String): [MemoryItem!]
}

type Mutation {
  createAgent(config: AgentConfig!): Agent!
  stopAgent(id: ID!): Agent!
  sendMessage(agentId: ID!, message: String!): Message
  storeMemory(item: MemoryInput!): MemoryItem!
}
```

#### 8.5.2 Resolver Implementation

- **Agent Resolution**: Fetch from agent service
- **Team Resolution**: Aggregate agent data
- **Memory Resolution**: Query memory store
- **LLM Resolution**: Model registry access

---

### 8.6 API Gateway

#### 8.6.1 Middleware Stack

```go
// pkg/gateway/gateway.go

type Middleware func(http.Handler) http.Handler

var middlewares = []Middleware{
    MiddlewareLogger,           // Request logging
    MiddlewareAuth,             // Authentication
    MiddlewareRateLimit,        // Rate limiting
    MiddlewareCORS,             // CORS handling
    MiddlewareRecovery,         // Panic recovery
    MiddlewareCompression,      // Response compression
    MiddlewareCache,            // Response caching
    MiddlewareRequestID,        // Request ID tracking
}
```

#### 8.6.2 Rate Limiting

```go
// pkg/gateway/middleware/ratelimit.go

type RateLimiter interface {
    Allow(identifier string) bool
    GetLimit(identifier string) RateLimit
}

type RateLimit struct {
    Requests int
    Duration time.Duration
}

// Default limits
AgentAPI:    100 req/min
TeamAPI:     50 req/min
LLMAPI:      20 req/min (higher cost)
MemoryAPI:   200 req/min
```

#### 8.6.3 Authentication

- **JWT Tokens**: Bearer token authentication
- **API Keys**: Simple key-based auth
- **OAuth 2.0**: Third-party authentication
- **Custom**: Plugin-based auth providers

#### 8.6.4 Routing

```go
// pkg/gateway/router.go

func SetupRouter() *gin.Engine {
    r := gin.Default()
    
    // API v1
    v1 := r.Group("/api/v1")
    {
        // Agent routes
        agents := v1.Group("/agents")
        {
            agents.GET("", ListAgents)
            agents.POST("", CreateAgent)
            agents.GET("/:id", GetAgent)
            agents.POST("/:id/start", StartAgent)
            // ... more routes
        }
        
        // Team routes
        teams := v1.Group("/teams")
        {
            teams.GET("", ListTeams)
            teams.POST("", CreateTeam)
            // ... more routes
        }
        
        // WebSocket
        v1.GET("/ws", WebSocketHandler)
        
        // SSE
        v1.GET("/stream/:agentId", SSEStreamHandler)
    }
    
    return r
}
```

---

### 8.7 Health & Metrics

#### 8.7.1 Health Endpoints

```go
// pkg/gateway/health.go

type HealthCheck struct {
    Status  string            `json:"status"`
    Service string            `json:"service"`
    Details map[string]string `json:"details,omitempty"`
}

// GET /health - Basic health check
// GET /health/ready - Readiness probe
// GET /health/live - Liveness probe
// GET /health/detailed - Detailed health info
```

#### 8.7.2 Metrics Export

```go
// pkg/gateway/metrics.go

type Metrics struct {
    // Request metrics
    RequestCount   prometheus.Counter
    RequestLatency prometheus.Histogram
    
    // Agent metrics
    AgentCount     prometheus.Gauge
    AgentStates    prometheus.GaugeVec
    
    // LLM metrics
    LLMRequests    prometheus.Counter
    LLMTokens      prometheus.Counter
    
    // Memory metrics
    MemoryItems    prometheus.Gauge
    MemorySize     prometheus.Gauge
    
    // Queue metrics
    QueueDepth     prometheus.Gauge
    QueueLatency   prometheus.Histogram
}
```

#### 8.7.3 Prometheus Export

- **/metrics**: Prometheus metrics endpoint
- **/prometheus**: Alternative metrics path
- **Custom metrics**: All layer metrics exported

---

## File Structure

```
pkg/gateway/
├── gateway.go           # Main gateway setup
├── router.go            # HTTP routing
├── types.go             # Common types

pkg/gateway/api/
├── api.go               # REST API handlers
├── agent.go             # Agent endpoints
├── team.go              # Team endpoints
├── llm.go               # LLM endpoints
├── memory.go            # Memory endpoints
├── workflow.go          # Workflow endpoints
├── health.go            # Health endpoints
└── types.go             # API request/response types

pkg/gateway/ws/
├── handler.go           # WebSocket handler
├── connection.go        # WebSocket connection management
├── messages.go          # WebSocket message types
└── broker.go            # WebSocket event broker

pkg/gateway/sse/
├── stream.go            # SSE stream implementation
├── handler.go           # SSE HTTP handler
└── events.go            # SSE event types

pkg/gateway/queue/
├── queue.go             # Queue abstraction
├── kafka.go             # Kafka implementation
├── rabbitmq.go          # RabbitMQ implementation
├── redis.go             # Redis implementation
└── producer.go          # Message producer
└── consumer.go          # Message consumer

pkg/gateway/graphql/
├── schema.go            # GraphQL schema
├── resolver.go          # Root resolver
├── agent_resolver.go    # Agent resolvers
├── team_resolver.go     # Team resolvers
├── llm_resolver.go      # LLM resolvers
└── memory_resolver.go   # Memory resolvers

pkg/gateway/middleware/
├── auth.go              # Authentication middleware
├── ratelimit.go         # Rate limiting middleware
├── cors.go              # CORS middleware
├── logging.go           # Request logging
├── recovery.go          # Panic recovery
├── compression.go       # Response compression
└── requestid.go         # Request ID middleware

pkg/gateway/metrics/
├── metrics.go           # Metrics definitions
├── exporter.go          # Prometheus exporter
└── handlers.go          # Metrics HTTP handlers
```

---

## TDD Implementation Plan

### Phase 1: REST API Foundation (25 tests)

#### Test Suite 1: API Setup
1. TestAPI_RoutesRegisteredCorrectly
2. TestAPI_MiddlewareChainApplied
3. TestAPI_CORSHeadersSet
4. TestAPI_CompressionEnabled
5. TestAPI_RequestIDAdded

#### Test Suite 2: Agent Endpoints
6. TestAgentAPI_ListReturnsAllAgents
7. TestAgentAPI_CreateCreatesNewAgent
8. TestAgentAPI_GetReturnsAgentData
9. TestAgentAPI_UpdateModifiesAgent
10. TestAgentAPI_DeleteRemovesAgent
11. TestAgentAPI_StartTransitionsState
12. TestAgentAPI_StopTransitionsState
13. TestAgentAPI_PauseTransitionsState
14. TestAgentAPI_InvalidIDReturns404
15. TestAgentAPI_InvalidPayloadReturns400

#### Test Suite 3: Team Endpoints
16. TestTeamAPI_ListReturnsAllTeams
17. TestTeamAPI_CreateCreatesNewTeam
18. TestTeamAPI_GetReturnsTeamData
19. TestTeamAPI_SendMessageSendsToTeam
20. TestTeamAPI_InvalidIDReturns404

#### Test Suite 4: LLM Endpoints
21. TestLLMAPI_GenerateReturnsResponse
22. TestLLMAPI_GenerateStreamReturnsStream
23. TestLLMAPI_ListModelsReturnsModels
24. TestLLMAPI_EmbedGeneratesEmbedding
25. TestLLMAPI_InvalidModelReturns400

### Phase 2: WebSocket & Streaming (20 tests)

#### Test Suite 5: WebSocket
26. TestWebSocket_HandshakeSucceeds
27. TestWebSocket_AuthenticationRequired
28. TestWebSocket_SendMessageDeliversMessage
29. TestWebSocket_BroadcastSendsToAll
30. TestWebSocket_SubscribeAddsSubscription
31. TestWebSocket_UnsubscribeRemovesSubscription
32. TestWebSocket_CloseGracefullyTerminates
33. TestWebSocket_ReconnectResumesConnection

#### Test Suite 6: SSE
34. TestSSE_StreamStartsCorrectly
35. TestSSE_SendEventDeliversEvent
36. TestSSE_ContentTypesetCorrectly
37. TestSSE_ReconnectionSupported
38. TestSSE_CloseStopsStream

#### Test Suite 7: Event Streaming
39. TestEvents_AgentEventsPublished
40. TestEvents_WorkflowEventsPublished
41. TestEvents_LogStreamWorks
42. TestEvents_MemoryUpdatesPublished

### Phase 3: Message Queues (15 tests)

#### Test Suite 8: Queue Abstraction
43. TestQueue_KafkaCreatesConnection
44. TestQueue_RabbitMQCreatesConnection
45. TestQueue_ProducerSendsMessage
46. TestQueue_ConsumerReceivesMessage
47. TestQueue_AcknowledgesMessage

#### Test Suite 9: Queue Implementations
48. TestKafkaQueue_ConnectsToBroker
49. TestKafkaQueue_ProducesMessages
50. TestKafkaQueue_ConsumesMessages
51. TestRabbitMQQueue_ConnectsToBroker
52. TestRabbitMQQueue_ProducesMessages
53. TestRedisQueue_ConnectsToRedis

#### Test Suite 10: Queue Integration
54. TestQueue_AgentCommandsQueued
55. TestQueue_WorkflowQueueProcesses
56. TestQueue_MemorySyncQueued

### Phase 4: GraphQL (10 tests)

#### Test Suite 11: GraphQL
57. TestGraphQL_QueryAgentsReturnsAgents
58. TestGraphQL_QueryAgentReturnsSingle
59. TestGraphQL_QueryTeamsReturnsTeams
60. TestGraphQL_MutationCreateAgentCreates
61. TestGraphQL_MutationStopAgentStops
62. TestGraphQL_MutationSendMessageSends
63. TestGraphQL_InvalidQueryReturnsError
64. TestGraphQL_AuthenticationRequired

### Phase 5: Gateway Features (20 tests)

#### Test Suite 12: Authentication
65. TestAuth_JWTValidTokenAuthorized
66. TestAuth_JWTInvalidTokenRejected
67. TestAuth_APIKeyValidAuthorized
68. TestAuth_APIKeyInvalidRejected
69. TestAuth_NoTokenRejected
70. TestAuth_ExpiredTokenRejected

#### Test Suite 13: Rate Limiting
71. TestRateLimit_AllowsWithinLimit
72. TestRateLimit_RejectsAboveLimit
73. TestRateLimit_ResetAfterDuration
74. TestRateLimit_DifferentIdentifiers
75. TestRateLimit-HeadersIncluded

#### Test Suite 14: Health & Metrics
76. TestHealth_BasicReturnsHealthy
77. TestHealth_ReadyReturnsReady
78. TestHealth_LiveReturnsAlive
79. TestHealth_DetailedReturnsDetails
80. TestMetrics_PrometheusExportsMetrics
81. TestMetrics_RequestCountIncrements
82. TestMetrics_AgentCountUpdates

#### Test Suite 15: Error Handling
83. TestError_404ForNotFound
84. TestError_400ForBadRequest
85. TestError_401ForUnauthorized
86. TestError_403ForForbidden
87. TestError_429ForRateLimited
88. TestError_500ForInternalServerError
89. TestError_ConsistentFormat

### Phase 6: Integration & Edge Cases (5 tests)

90. TestGateway_AgentFlowComplete
91. TestGateway_TeamCollaborationFlow
92. TestGateway_WebSocketUnderLoad
93. TestGateway_QueueBackpressure
94. TestGateway_FailoverScenarios

---

## Dependencies

### External Dependencies

- **Web Framework**: Gin, Echo, or Fiber (Go HTTP framework)
- **WebSocket**: gorilla/websocket or similar
- **GraphQL**: gqlgen or similar
- **Message Queues**: Kafka, RabbitMQ, or Redis
- **Metrics**: Prometheus client library
- **CORS**: CORS middleware

### Internal Dependencies

- Layer 7: Agent, team, LLM, memory services
- Layer 6: Orchestration engine
- Layer 5: Event bus
- Layer 4: Authentication/Authorization
- Layer 1: Bootstrap context

---

## Risk Assessment

### Technical Risks

1. **API Performance Under Load**
   - **Mitigation**: Connection pooling, caching, async processing
   - **Impact**: High - Gateway is entry point

2. **WebSocket Connection Management**
   - **Mitigation**: Connection pooling, heartbeat, graceful shutdown
   - **Impact**: High - Real-time communication critical

3. **Message Queue Reliability**
   - **Mitigation**: Retry logic, dead letter queues, monitoring
   - **Impact**: Medium - Async communication reliability

4. **GraphQL Complexity**
   - **Mitigation**: Query depth limiting, complexity analysis
   - **Impact**: Medium - Query performance

5. **Rate Limiting Accuracy**
   - **Mitigation**: Distributed rate limiting, proper counters
   - **Impact**: Medium - API protection

### Operational Risks

1. **Security**
   - **Mitigation**: Strong auth, input validation, rate limiting
   - **Impact**: High - Gateway is attack surface

2. **Monitoring**
   - **Mitigation**: Comprehensive metrics, alerting
   - **Impact**: Medium - Observability critical

3. **Scalability**
   - **Mitigation**: Stateless design, horizontal scaling
   - **Impact**: Medium - Performance at scale

---

## Open Questions

1. **Which HTTP framework to use?**
   - Options: Gin, Echo, Fiber, standard net/http
   - Decision criteria: Performance, features, community

2. **GraphQL vs REST only?**
   - Options: Both, REST only, GraphQL only
   - Decision criteria: Client needs, complexity

3. **Message queue strategy?**
   - Options: Kafka, RabbitMQ, Redis, multiple
   - Decision criteria: Existing infrastructure, requirements

4. **Authentication providers?**
   - Options: JWT, API keys, OAuth, custom
   - Decision criteria: Security needs, user base

5. **Caching strategy?**
   - Options: Redis, in-memory, CDN
   - Decision criteria: Performance needs, infrastructure

---

## Testing Strategy

### Unit Tests
- API endpoint handlers
- WebSocket message handling
- SSE stream management
- Queue producer/consumer
- GraphQL resolvers
- Middleware implementations

### Integration Tests
- Full API workflows
- WebSocket communication
- Queue integration
- GraphQL queries/mutations
- Authentication flows

### End-to-End Tests
- Complete agent lifecycle via API
- Multi-agent team workflows
- Real-time event streaming
- Queue-based communication
- High-load scenarios

### Performance Tests
- 1000+ concurrent REST requests
- 500+ WebSocket connections
- 10K+ messages/second via queues
- GraphQL query complexity limits
- Rate limiting under load

---

## Success Criteria

- [ ] All REST endpoints functional and tested
- [ ] WebSocket connections stable and authenticated
- [ ] Event streaming works reliably
- [ ] Message queues integrated and tested
- [ ] GraphQL interface functional
- [ ] Authentication/authorization working
- [ ] Rate limiting effective
- [ ] Health checks pass
- [ ] Metrics exported correctly
- [ ] All 94 tests pass
- [ ] Performance meets requirements
- [ ] Security audit passed

---
