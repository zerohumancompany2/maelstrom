# Gap Remediation Phase Breakdown

**Generated**: 2026-03-09  
**Parent**: `docs/planning/gap-remediation-plan.md`

---

## Phase G1: Critical Fixes (P0)

**Goal**: Fix compilation errors and critical security gaps that block all other work

### Gap Details

| Gap | Spec Ref | File | Line | Issue |
|-----|----------|------|------|-------|
| L3-H4 | arch-v1.md 9.2 | pkg/mail/router.go | 12 | ParseAddress function referenced but not found |
| L3-C1 | arch-v1.md 10.2 | pkg/mail/stream.go | 32-38 | Send() and Close() panic with "not implemented" |
| L2-C1 | arch-v1.md 7.1, 14.2 | pkg/services/security/service.go | 28-30 | ValidateAndSanitize returns nil without validation |
| L2-C2 | arch-v1.md 7.3 | pkg/services/security/service.go | - | NamespaceIsolate method missing |
| L2-C3 | arch-v1.md 7.3 | pkg/services/security/service.go | - | CheckTaintPolicy method missing |
| L3-C3 | arch-v1.md 10.4 | pkg/services/humangateway/service.go | - | Chat session UI/HTTPS endpoint not implemented |

### Test Count: 12 tests

#### L3-H4: ParseAddress (2 tests, 2 commits)
1. `TestParseAddress_agent` - Parse agent:id format
2. `TestParseAddress_topic` - Parse topic:name format
3. `TestParseAddress_sys` - Parse sys:service format
4. `TestParseAddress_invalid` - Parse invalid format returns error

#### L3-C1: StreamSession Send/Close (4 tests, 4 commits)
5. `TestStreamSession_Send` - Send chunk to channel
6. `TestStreamSession_SendMultiple` - Send multiple chunks in order
7. `TestStreamSession_Close` - Close session and signal completion
8. `TestStreamSession_CloseAfterSend` - Close after sends completes cleanly

#### L2-C1: Security Boundary Enforcement (2 tests, 2 commits)
9. `TestSecurityService_ValidateAndSanitize_innerToOuter` - Validate inner->outer transition
10. `TestSecurityService_ValidateAndSanitize_outerToInner` - Validate outer->inner transition with taints

#### L2-C2: NamespaceIsolate (2 tests, 2 commits)
11. `TestSecurityService_NamespaceIsolate` - Create isolated view for agent
12. `TestSecurityService_NamespaceIsolate_multipleAgents` - Multiple isolated views

#### L2-C3: CheckTaintPolicy (2 tests, 2 commits)
13. `TestSecurityService_CheckTaintPolicy_allowed` - Check policy allows transition
14. `TestSecurityService_CheckTaintPolicy_denied` - Check policy denies transition

#### L3-C3: Human Gateway Chat (2 tests, 2 commits)
15. `TestHumanGatewayService_ChatEndpoint` - Chat endpoint exists and responds
16. `TestHumanGatewayService_ChatSession` - Chat session creates context

### File Modifications

| File | Changes |
|------|---------|
| pkg/mail/router.go | Add ParseAddress function, AddressType enum |
| pkg/mail/stream.go | Implement Send() to write to channel, Close() to close channel |
| pkg/services/security/service.go | Implement ValidateAndSanitize with boundary logic, add NamespaceIsolate, add CheckTaintPolicy |
| pkg/services/humangateway/service.go | Add chat endpoint handler stub |

### Dependencies
- None

### Execution Order within Phase
1. L3-H4 (ParseAddress) - Unblocks compilation
2. L3-C1 (StreamSession) - Unblocks streaming
3. L2-C1 (Boundary Enforcement) - Core security
4. L2-C2 (NamespaceIsolate) - Security extension
5. L2-C3 (CheckTaintPolicy) - Security extension
6. L3-C3 (Chat Endpoint) - Human gateway stub

---

## Phase G2: Core Functionality (P1)

**Goal**: Implement missing core functionality for security and communication

### Gap Details

| Gap | Spec Ref | File | Issue |
|-----|----------|------|-------|
| L2-H1 | arch-v1.md 7.1 | pkg/services/security/service.go | Taint propagation returns input unchanged |
| L2-H2 | arch-v1.md 9.2 | pkg/services/communication/service.go | No retry logic for at-least-once delivery |
| L3-H1 | arch-v1.md 9.2 | pkg/mail/address.go | Validation helpers not found |
| L3-H2 | arch-v1.md 9.2 | pkg/services/communication/service.go | Request-reply via correlationId not implemented |
| L3-H3 | arch-v1.md 10.1 | pkg/gateway/adapters/*.go | No actual HTTP/WebSocket/SSE servers |

### Test Count: 18 tests

#### L2-H1: Taint Propagation (3 tests, 3 commits)
1. `TestSecurityService_TaintPropagate_addTaints` - Add new taints to object
2. `TestSecurityService_TaintPropagate_mergeTaints` - Merge existing and new taints
3. `TestSecurityService_TaintPropagate_nestedObjects` - Propagate through nested structures

#### L2-H2: At-Least-Once Delivery (4 tests, 4 commits)
4. `TestCommunicationService_RetryOnFailure` - Retry delivery on failure
5. `TestCommunicationService_ExponentialBackoff` - Backoff between retries
6. `TestCommunicationService_MaxRetries` - Stop after max retries
7. `TestCommunicationService_DeliveryTracking` - Track delivery attempts

#### L3-H1: Address Validation (3 tests, 3 commits)
8. `TestIsValidAgentAddress` - Validate agent:id format
9. `TestIsValidTopicAddress` - Validate topic:name format
10. `TestIsValidSysAddress` - Validate sys:service format

#### L3-H2: Request-Reply Pattern (4 tests, 4 commits)
11. `TestCommunicationService_RequestReply` - Send request and receive reply
12. `TestCommunicationService_CorrelationIdMatching` - Match reply to request
13. `TestCommunicationService_RequestTimeout` - Timeout on no reply
14. `TestCommunicationService_MultipleRequests` - Handle multiple concurrent requests

#### L3-H3: Gateway Servers (4 tests, 4 commits)
15. `TestWebhookAdapter_HTTPServer` - HTTP server starts and responds
16. `TestWebSocketAdapter_WSConnection` - WebSocket server accepts connections
17. `TestSSEAdapter_SSEEndpoint` - SSE endpoint streams events
18. `TestGatewayAdapter_Normalization` - Inbound/outbound normalization works

### File Modifications

| File | Changes |
|------|---------|
| pkg/services/security/service.go | Implement TaintPropagate with actual taint merging |
| pkg/services/communication/service.go | Add retry logic with exponential backoff, delivery tracking |
| pkg/mail/address.go | Add IsValidAgentAddress, IsValidTopicAddress, IsValidSysAddress |
| pkg/services/communication/service.go | Add correlationId tracking for request-reply |
| pkg/gateway/adapters/webhook.go | Add HTTP server implementation |
| pkg/gateway/adapters/websocket.go | Add WebSocket server implementation |
| pkg/gateway/adapters/sse.go | Add SSE endpoint implementation |

### Dependencies
- G1 (ParseAddress for address validation)

### Execution Order within Phase
1. L3-H1 (Address Validation) - Foundation for other work
2. L2-H1 (Taint Propagation) - Security functionality
3. L2-H2 (Delivery Guarantees) - Communication reliability
4. L3-H2 (Request-Reply) - Communication patterns
5. L3-H3 (Gateway Servers) - External connectivity

---

## Phase G3: Integration & Wiring (P1)

**Goal**: Wire services together, fix integration issues

### Gap Details

| Gap | Spec Ref | File | Issue |
|-----|----------|------|-------|
| L2-H5 | docs/completed/layer-02-core-services.md 431-440 | pkg/services/registry.go | No lifecycle state tracking |
| L3-M3 | arch-v1.md 9.2 | pkg/services/observability/service.go | Dead-letter not wired to communication |
| L3-M4 | arch-v1.md 10.2 | pkg/mail/stream.go | Taint stripping not in mail flow |
| L2-M5 | docs/completed/layer-02-core-services.md 431-480 | pkg/kernel/kernel.go | Services not auto-started |

### Test Count: 14 tests

#### L2-H5: Service Registry State (3 tests, 3 commits)
1. `TestServiceRegistry_RegisterWithState` - Register service with initial state
2. `TestServiceRegistry_UpdateState` - Update service lifecycle state
3. `TestServiceRegistry_QueryByState` - Query services by state

#### L3-M3: Dead-Letter Integration (4 tests, 4 commits)
4. `TestCommunicationService_DeadLetterOnFailure` - Failed delivery goes to dead-letter
5. `TestObservabilityService_QueryDeadLetters` - Query dead-letter entries
6. `TestCommunicationService_DeadLetterWithReason` - Include failure reason
7. `TestCommunicationService_DeadLetterAfterMaxRetries` - Dead-letter after retries exhausted

#### L3-M4: Stream Taint Integration (3 tests, 3 commits)
8. `TestMailStream_TaintStripping` - Taints stripped in stream flow
9. `TestMailStream_TaintPropagation` - Taints propagated correctly
10. `TestMailStream_SecurityCheck` - Security boundary check in stream

#### L2-M5: Service Bootstrap (4 tests, 4 commits)
11. `TestKernel_BootstrapServices` - Services start during bootstrap
12. `TestKernel_BootstrapSequence` - Services start in correct order
13. `TestKernel_ServiceReadyEvents` - Ready events emitted after bootstrap
14. `TestKernel_BootstrapFailure` - Handle service startup failures

### File Modifications

| File | Changes |
|------|---------|
| pkg/services/registry.go | Add state field, UpdateState method, QueryByState method |
| pkg/services/observability/service.go | Wire dead-letter to communication service failures |
| pkg/mail/stream.go | Integrate taint stripping in mail flow |
| pkg/kernel/kernel.go | Add service bootstrap sequence, auto-start services |
| pkg/services/communication/service.go | Add dead-letter callback on delivery failure |

### Dependencies
- G1 (Security service for taint stripping)
- G2 (Communication service for dead-letter integration)

### Execution Order within Phase
1. L2-H5 (Registry State) - Foundation for bootstrap
2. L2-M5 (Service Bootstrap) - Wire services to kernel
3. L3-M3 (Dead-Letter) - Wire observability
4. L3-M4 (Stream Taints) - Wire security into mail flow

---

## Phase G4: Gateway & External APIs (P2)

**Goal**: Complete gateway adapters and HTTP endpoints

### Gap Details

| Gap | Spec Ref | File | Issue |
|-----|----------|------|-------|
| L3-C2 | arch-v1.md 10.3 | pkg/gateway/openapi.go | OpenAPI generation not implemented |
| L2-M4 | arch-v1.md 7.2 | pkg/services/* | Hot-reloadable services not implemented |
| L3-M2 | arch-v1.md 9.3 | pkg/mail/router.go | TopicSubscriber interface mismatch |

### Test Count: 20 tests

#### L3-C2: OpenAPI Generation (4 tests, 4 commits)
1. `TestOpenAPI_GenerateFromChart` - Generate OpenAPI from chart YAML
2. `TestOpenAPI_HTTPExposure` - Generate HTTP endpoint specs
3. `TestOpenAPI_AuthMiddleware` - Auth middleware in spec
4. `TestOpenAPI_BoundaryFiltering` - Only DMZ/outer exposed

#### L2-M4: Hot-Reloadable Services (10 tests, 10 commits)
5. `TestAdminService_AdminCommand` - Execute admin command
6. `TestAdminService_2FAGate` - 2FA required for commands
7. `TestPersistenceService_Snapshot` - Take snapshot
8. `TestPersistenceService_Restore` - Restore from snapshot
9. `TestHeartbeatService_Schedule` - Schedule wake-ups
10. `TestHeartbeatService_Trigger` - Trigger heartbeat
11. `TestMemoryService_Store` - Store in memory
12. `TestMemoryService_Query` - Query memory
13. `TestToolsService_Register` - Register tool
14. `TestToolsService_Resolve` - Resolve tool

#### L3-M2: TopicSubscriber Interface (2 tests, 2 commits)
15. `TestTopicSubscriber_Interface` - TopicSubscriber interface implemented
16. `TestTopicSubscriber_Subscribe` - Subscribe via TopicSubscriber

#### L3-C2: HTTP Endpoint Exposure (4 tests, 4 commits)
17. `TestGatewayService_RegisterHTTPEndpoint` - Register HTTP endpoint
18. `TestGatewayService_HTTPEndpointHandler` - Endpoint handler works
19. `TestGatewayService_OpenAPIRegistration` - Endpoint in OpenAPI spec
20. `TestGatewayService_BoundaryExposure` - Inner boundary not exposed

### File Modifications

| File | Changes |
|------|---------|
| pkg/gateway/openapi.go | Add OpenAPI generation from charts |
| pkg/gateway/gateway.go | Add HTTP endpoint registration |
| pkg/services/admin/service.go | Implement admin commands with 2FA |
| pkg/services/persistence/service.go | Implement snapshot/restore |
| pkg/services/heartbeat/service.go | Implement scheduled wake-ups |
| pkg/services/memory/service.go | Implement vector/graph stores |
| pkg/services/tools/service.go | Implement tool registry |
| pkg/mail/router.go | Fix TopicSubscriber interface |

### Dependencies
- G1 (Security for boundary enforcement)
- G2 (Gateway servers for HTTP endpoints)
- G3 (Service bootstrap for hot-reloadable services)

### Execution Order within Phase
1. L3-M2 (TopicSubscriber) - Fix interface first
2. L3-C2 (OpenAPI) - API generation
3. L2-M4 (Hot-Reloadable Services) - Implement services

---

## Phase G5: Observability & Metrics (P2)

**Goal**: Complete observability features

### Gap Details

| Gap | Spec Ref | File | Issue |
|-----|----------|------|-------|
| L2-H3 | arch-v1.md 7.3 | pkg/services/observability/service.go | No transition/event rate tracking |
| L2-M2 | arch-v1.md 7.3 | pkg/services/observability/service.go | Dead-letter query returns copy (memory inefficient) |
| L2-M3 | arch-v1.md 7.3 | pkg/services/lifecycle/service.go | Runtime tracking not updated |
| L3-M1 | arch-v1.md 9.2 | pkg/mail/types.go | Mail.Metadata.Stream type mismatch |

### Test Count: 13 tests

#### L2-H3: Metrics Collection (4 tests, 4 commits)
1. `TestObservabilityService_TransitionRate` - Track state transition rate
2. `TestObservabilityService_EventRate` - Track event processing rate
3. `TestObservabilityService_MetricsCollector` - Get metrics collector
4. `TestObservabilityService_MetricsAggregation` - Aggregate metrics over time

#### L2-M2: Dead-Letter Query Optimization (3 tests, 3 commits)
5. `TestObservabilityService_QueryDeadLettersNoCopy` - Query returns references not copies
6. `TestObservabilityService_QueryDeadLettersMemory` - Memory usage within bounds
7. `TestObservabilityService_QueryDeadLettersLargeSet` - Handle large dead-letter sets efficiently

#### L2-M3: Runtime Tracking (3 tests, 3 commits)
8. `TestLifecycleService_RuntimeStateUpdate` - Update runtime state on transition
9. `TestLifecycleService_ListWithStates` - List runtimes with current states
10. `TestLifecycleService_StateHistory` - Track state history

#### L3-M1: Mail Metadata Type (3 tests, 3 commits)
11. `TestMailMetadata_StreamBool` - Stream field is boolean
12. `TestMailMetadata_StreamChunk` - Stream chunk separate field
13. `TestMailMetadata_TypeAlignment` - All metadata fields match spec

### File Modifications

| File | Changes |
|------|---------|
| pkg/services/observability/service.go | Add transition rate tracking, event rate tracking, fix dead-letter query to return references |
| pkg/services/lifecycle/service.go | Update runtime tracking on state changes |
| pkg/mail/types.go | Fix Mail.Metadata.Stream to be bool, move StreamChunk elsewhere |

### Dependencies
- G1 (Observability service exists)
- G2 (Communication service for event tracking)

### Execution Order within Phase
1. L3-M1 (Mail Metadata) - Fix type first
2. L2-H3 (Metrics) - Add rate tracking
3. L2-M2 (Dead-Letter Optimization) - Optimize queries
4. L2-M3 (Runtime Tracking) - Update lifecycle tracking

---

## Phase G6: Hot-Reload & Advanced Features (P3)

**Goal**: Implement hot-reload and advanced features

### Gap Details

| Gap | Spec Ref | File | Issue |
|-----|----------|------|-------|
| L2-H4 | arch-v1.md 7.2 | pkg/services/lifecycle/service.go | Hot-reload not implemented |
| L2-M1 | arch-v1.md 9.2 | pkg/services/communication/service.go | Deduplication via correlationId not implemented |

### Test Count: 8 tests

#### L2-H4: Hot-Reload (4 tests, 4 commits)
1. `TestLifecycleService_HotReload` - Hot-reload running service
2. `TestLifecycleService_HotReloadStatePreservation` - State preserved during reload
3. `TestLifecycleService_HotReloadFailure` - Handle reload failure
4. `TestLifecycleService_HotReloadRollback` - Rollback on failure

#### L2-M1: Deduplication (4 tests, 4 commits)
5. `TestCommunicationService_Deduplication` - Deduplicate by correlationId
6. `TestCommunicationService_DeduplicationWindow` - Deduplication time window
7. `TestCommunicationService_DeduplicationExpiry` - Old correlationIds expire
8. `TestCommunicationService_DeduplicationConcurrent` - Concurrent deduplication

### File Modifications

| File | Changes |
|------|---------|
| pkg/services/lifecycle/service.go | Add HotReload method with state preservation |
| pkg/services/communication/service.go | Add correlationId deduplication with time window |

### Dependencies
- All above phases

### Execution Order within Phase
1. L2-M1 (Deduplication) - Foundation for reliable delivery
2. L2-H4 (Hot-Reload) - Advanced lifecycle management

---

## Pre-existing: Mail System Integration Tests

**Goal**: Fix known integration test failures (non-blocking)

### Gap Details

| Test | File | Line | Issue |
|------|------|------|-------|
| TestMailSystem_PublishDeliversMail | pkg/mail/system_test.go | 45-47 | Mail not delivered to subscriber |
| TestMailSystem_SubscribeReceivesMail | pkg/mail/system_test.go | 89-91 | Subscription not registered |
| TestMailSystem_UnsubscribeRemovesSubscriber | pkg/mail/system_test.go | 134-136 | Subscriber still receiving after unsubscribe |
| TestMailSystem_ConcurrentPublish | pkg/mail/system_test.go | 178-180 | Race condition in concurrent publish |

### Test Count: 4 tests (fix existing)

#### Integration Test Fixes (4 commits)
1. Fix TestMailSystem_PublishDeliversMail - Ensure subscriber registration before publish
2. Fix TestMailSystem_SubscribeReceivesMail - Ensure subscription is active
3. Fix TestMailSystem_UnsubscribeRemovesSubscriber - Ensure unsubscribe removes from router
4. Fix TestMailSystem_ConcurrentPublish - Add synchronization for concurrent access

### File Modifications

| File | Changes |
|------|---------|
| pkg/mail/system_test.go | Fix test setup and synchronization |
| pkg/mail/mailbox.go | Potential fixes for delivery |
| pkg/mail/subscriber.go | Potential fixes for subscription management |
| pkg/kernel/kernel.go | Potential fixes for service spawning |

### Dependencies
- G1 (ParseAddress for routing)
- G3 (Service bootstrap for proper initialization)

---

## Summary Table

| Phase | Priority | Gaps | Tests | Commits | Files Modified | Dependencies |
|-------|----------|------|-------|---------|----------------|--------------|
| G1 | P0 | 6 | 12 | 12 | 4 | None |
| G2 | P1 | 5 | 18 | 18 | 7 | G1 |
| G3 | P1 | 4 | 14 | 14 | 5 | G1, G2 |
| G4 | P2 | 3 | 20 | 20 | 8 | G1, G2, G3 |
| G5 | P2 | 4 | 13 | 13 | 3 | G1, G2 |
| G6 | P3 | 2 | 8 | 8 | 2 | All |
| Mail Integration | Pre-existing | 4 | 4 | 4 | 4 | G1, G3 |
| **Total** | | **28** | **89** | **89** | **33** | |

---

## Gap to Requirement Mapping

| Gap | Requirement (arch-v1.md) | Phase |
|-----|--------------------------|-------|
| L2-C1 | 7.1, 14.2 - Boundary enforcement | G1 |
| L2-C2 | 7.3 - NamespaceIsolate | G1 |
| L2-C3 | 7.3 - CheckTaintPolicy | G1 |
| L2-H1 | 7.1 - Taint propagation | G2 |
| L2-H2 | 9.2 - At-least-once delivery | G2 |
| L2-H3 | 7.3 - Metrics collection | G5 |
| L2-H4 | 7.2 - Hot-reload | G6 |
| L2-H5 | layer-02-core-services.md 431-440 - Lifecycle state | G3 |
| L2-M1 | 9.2 - Deduplication | G6 |
| L2-M2 | 7.3 - Dead-letter query optimization | G5 |
| L2-M3 | 7.3 - Runtime tracking | G5 |
| L2-M4 | 7.2 - Hot-reloadable services | G4 |
| L2-M5 | layer-02-core-services.md 431-480 - Bootstrap | G3 |
| L3-C1 | 10.2 - StreamSession Send/Close | G1 |
| L3-C2 | 10.3 - OpenAPI generation | G4 |
| L3-C3 | 10.4 - Chat endpoint | G1 |
| L3-H1 | 9.2 - Address validation | G2 |
| L3-H2 | 9.2 - Request-reply | G2 |
| L3-H3 | 10.1 - Gateway servers | G2 |
| L3-H4 | 9.2 - ParseAddress | G1 |
| L3-M1 | 9.2 - Mail metadata type | G5 |
| L3-M2 | 9.3 - TopicSubscriber interface | G4 |
| L3-M3 | 9.2 - Dead-letter integration | G3 |
| L3-M4 | 10.2 - Stream taint stripping | G3 |

---

**Phase Breakdown End**