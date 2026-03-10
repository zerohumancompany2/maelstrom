# Phase 7.1: Hard-Coded Services Completion

## Goal
Complete the remaining 10% of hard-coded platform services (sys:security, sys:communication, sys:observability, sys:lifecycle) per spec (arch-v1.md L453-461).

## Scope
- Complete sys:security boundary enforcement (arch-v1.md L457)
- Complete sys:communication mail pub/sub (arch-v1.md L458)
- Complete sys:observability trace storage (arch-v1.md L459)
- Complete sys:lifecycle spawn/stop (arch-v1.md L460)
- Verify all 4 services integrate with bootstrap sequence (arch-v1.md L800-847)
- 6 tests, 6 commits

## Current Status

### Implementation Status
| Service | Status | Notes |
|---------|--------|-------|
| `sys:security` | ⚠️ 90% complete | Taint propagation exists; needs boundary enforcement tests |
| `sys:communication` | ⚠️ 90% complete | Pub/sub exists; needs integration with bootstrap |
| `sys:observability` | ⚠️ 90% complete | Trace storage exists; needs dead-letter integration |
| `sys:lifecycle` | ⚠️ 90% complete | Spawn/stop exists; needs runtime tracking tests |

### Files Status
| File | Status |
|------|--------|
| `pkg/services/security/service.go` | ⚠️ Partial - add boundary enforcement tests |
| `pkg/services/communication/service.go` | ⚠️ Partial - add bootstrap integration tests |
| `pkg/services/observability/service.go` | ⚠️ Partial - add dead-letter tests |
| `pkg/services/lifecycle/service.go` | ⚠️ Partial - add runtime tracking tests |

## Required Implementation

### sys:security Boundary Enforcement (arch-v1.md L457)
```go
// pkg/services/security/service.go
func (s *SecurityService) ValidateAndSanitize(mail mail.Mail, sourceBoundary, targetBoundary mail.BoundaryType) (mail.Mail, error) {
    // Enforce boundary transitions: outer→inner (taint), inner→outer (block SECRET/PII)
}
```

### sys:communication Pub/Sub (arch-v1.md L458)
```go
// pkg/services/communication/service.go
func (c *CommunicationService) Publish(mail mail.Mail) (ack mail.Ack, err error) {
    // Deliver to subscribers, return Ack with delivery confirmation
}
```

### sys:observability Dead-Letter (arch-v1.md L459)
```go
// pkg/services/observability/service.go
func (o *ObservabilityService) LogDeadLetter(mail mail.Mail, reason string) {
    // Store mail and reason in dead-letter queue
}

func (o *ObservabilityService) QueryDeadLetters(filters TraceFilters) []DeadLetterEntry {
    // Return all logged entries matching filters
}
```

### sys:lifecycle Spawn Tracking (arch-v1.md L460)
```go
// pkg/services/lifecycle/service.go
func (l *LifecycleService) Spawn(ctx context.Context, def ChartDefinition) (RuntimeID, error) {
    // Spawn runtime and add to internal tracking
}

func (l *LifecycleService) List() []RuntimeInfo {
    // Return all tracked runtimes with RuntimeInfo
}
```

## Tests to Write (6 tests, 6 commits)

### Test 1: TestHardcodedServices_SecurityBoundaryEnforcement
```go
func TestHardcodedServices_SecurityBoundaryEnforcement(t *testing.T) {
    svc := NewSecurityService()
    
    // Test outer→inner tainting (arch-v1.md L457)
    outerMail := mail.Mail{
        ID:       "test-001",
        Source:   "agent:outer",
        Target:   "agent:inner",
        Content:  map[string]any{"data": "sensitive"},
        Metadata: mail.MailMetadata{Boundary: mail.OuterBoundary},
    }
    
    taintedMail, err := svc.ValidateAndSanitize(outerMail, mail.OuterBoundary, mail.InnerBoundary)
    if err != nil {
        t.Errorf("Expected no error for outer→inner, got %v", err)
    }
    if !slices.Contains(taintedMail.Metadata.Taints, "OUTER_BOUNDARY") {
        t.Error("Expected OUTER_BOUNDARY taint added for outer→inner transition")
    }
    
    // Test inner→outer blocks SECRET/PII (arch-v1.md L457)
    innerMail := mail.Mail{
        ID:       "test-002",
        Source:   "agent:inner",
        Target:   "agent:outer",
        Content:  map[string]any{"secret": "password123"},
        Metadata: mail.MailMetadata{
            Boundary: mail.InnerBoundary,
            Taints:   []string{"SECRET", "PII"},
        },
    }
    
    _, err = svc.ValidateAndSanitize(innerMail, mail.InnerBoundary, mail.OuterBoundary)
    if err == nil {
        t.Error("Expected error for inner→outer with SECRET/PII taints")
    }
}
```
**Acceptance Criteria:**
- `ValidateAndSanitize()` enforces outer→inner tainting (arch-v1.md L457)
- `ValidateAndSanitize()` blocks inner→outer for SECRET/PII taints
- Returns error on boundary violation

### Test 2: TestHardcodedServices_SecurityTaintPropagation
```go
func TestHardcodedServices_SecurityTaintPropagation(t *testing.T) {
    svc := NewSecurityService()
    
    // Test recursive taint propagation (arch-v1.md L457)
    nestedMap := map[string]any{
        "level1": map[string]any{
            "level2": map[string]any{
                "value": "test",
            },
        },
        "direct": "value",
    }
    
    propagated := svc.TaintPropagate(nestedMap, []string{"USER_SUPPLIED"})
    
    // Check top level has _taints
    if propagated["_taints"] == nil {
        t.Error("Expected _taints field at top level")
    }
    
    // Check nested levels have _taints
    level1 := propagated["level1"].(map[string]any)
    if level1["_taints"] == nil {
        t.Error("Expected _taints field at level1")
    }
    
    level2 := level1["level2"].(map[string]any)
    if level2["_taints"] == nil {
        t.Error("Expected _taints field at level2")
    }
    
    // Test taint merging
    existingTainted := map[string]any{
        "_taints": []string{"EXISTING"},
        "data":    "value",
    }
    merged := svc.TaintPropagate(existingTainted, []string{"NEW"})
    taints := merged["_taints"].([]string)
    if !slices.Contains(taints, "EXISTING") || !slices.Contains(taints, "NEW") {
        t.Error("Expected both EXISTING and NEW taints in merged result")
    }
}
```
**Acceptance Criteria:**
- `TaintPropagate()` recursively adds taints to nested maps
- `TaintPropagate()` merges existing and new taints
- Returns propagated object with `_taints` field

### Test 3: TestHardcodedServices_CommunicationPublishAck
```go
func TestHardcodedServices_CommunicationPublishAck(t *testing.T) {
    svc := NewCommunicationService()
    
    // Setup subscriber
    received := make(chan mail.Mail, 1)
    svc.Subscribe("topic:test", func(m mail.Mail) {
        received <- m
    })
    
    // Publish mail (arch-v1.md L458)
    testMail := mail.Mail{
        ID:            "msg-001",
        CorrelationID: "corr-001",
        Type:          mail.MailTypeUser,
        Source:        "agent:sender",
        Target:        "topic:test",
        Content:       map[string]any{"text": "hello"},
        CreatedAt:     time.Now(),
    }
    
    ack, err := svc.Publish(testMail)
    if err != nil {
        t.Fatalf("Expected no error, got %v", err)
    }
    
    // Verify Ack (arch-v1.md L458)
    if ack.CorrelationID != testMail.CorrelationID {
        t.Errorf("Expected CorrelationID 'corr-001', got '%s'", ack.CorrelationID)
    }
    if ack.DeliveredAt.IsZero() {
        t.Error("Expected DeliveredAt to be set")
    }
    
    // Verify delivery
    select {
    case m := <-received:
        if m.ID != testMail.ID {
            t.Errorf("Expected mail ID 'msg-001', got '%s'", m.ID)
        }
    case <-time.After(100 * time.Millisecond):
        t.Error("Expected mail to be delivered to subscriber")
    }
}
```
**Acceptance Criteria:**
- `Publish()` returns Ack with CorrelationID matching published mail (arch-v1.md L458)
- Ack.DeliveredAt is set to current time
- Mail delivered to subscriber

### Test 4: TestHardcodedServices_ObservabilityDeadLetter
```go
func TestHardcodedServices_ObservabilityDeadLetter(t *testing.T) {
    svc := NewObservabilityService()
    
    // Log dead-letter entry (arch-v1.md L459)
    failedMail := mail.Mail{
        ID:       "failed-001",
        Type:     mail.MailTypeError,
        Source:   "agent:failed",
        Target:   "agent:target",
        Content:  map[string]any{"error": "processing failed"},
        CreatedAt: time.Now(),
    }
    
    svc.LogDeadLetter(failedMail, "Max retries exceeded")
    
    // Query dead-letters (arch-v1.md L459)
    entries := svc.QueryDeadLetters(TraceFilters{})
    if len(entries) != 1 {
        t.Fatalf("Expected 1 dead-letter entry, got %d", len(entries))
    }
    
    entry := entries[0]
    if entry.MailID != "failed-001" {
        t.Errorf("Expected MailID 'failed-001', got '%s'", entry.MailID)
    }
    if entry.Reason != "Max retries exceeded" {
        t.Errorf("Expected Reason 'Max retries exceeded', got '%s'", entry.Reason)
    }
    if entry.Logged.IsZero() {
        t.Error("Expected Logged timestamp to be set")
    }
}
```
**Acceptance Criteria:**
- `LogDeadLetter()` stores mail and reason (arch-v1.md L459)
- `QueryDeadLetters()` returns all logged entries
- DeadLetterEntry has Logged timestamp

### Test 5: TestHardcodedServices_LifecycleSpawnTracking
```go
func TestHardcodedServices_LifecycleSpawnTracking(t *testing.T) {
    ctx := context.Background()
    svc := NewLifecycleService()
    
    // Spawn runtime (arch-v1.md L460)
    def := ChartDefinition{
        Metadata: ChartMetadata{
            Name:     "test-agent",
            Version:  "1.0.0",
            Boundary: mail.InnerBoundary,
        },
    }
    
    runtimeID, err := svc.Spawn(ctx, def)
    if err != nil {
        t.Fatalf("Expected no error, got %v", err)
    }
    
    // List runtimes (arch-v1.md L460)
    runtimes := svc.List()
    if len(runtimes) != 1 {
        t.Fatalf("Expected 1 runtime, got %d", len(runtimes))
    }
    
    info := runtimes[0]
    if info.ID != runtimeID {
        t.Errorf("Expected runtime ID '%s', got '%s'", runtimeID, info.ID)
    }
    if info.DefinitionID != "test-agent" {
        t.Errorf("Expected DefinitionID 'test-agent', got '%s'", info.DefinitionID)
    }
    if info.Boundary != mail.InnerBoundary {
        t.Errorf("Expected Boundary InnerBoundary, got %s", info.Boundary)
    }
}
```
**Acceptance Criteria:**
- `Spawn()` adds runtime to internal tracking (arch-v1.md L460)
- `List()` includes spawned runtime with RuntimeInfo
- RuntimeInfo has ID, DefinitionID, Boundary fields

### Test 6: TestHardcodedServices_AllServicesBootstrapIntegration
```go
func TestHardcodedServices_AllServicesBootstrapIntegration(t *testing.T) {
    kernel := NewKernel()
    
    // Start kernel bootstrap (arch-v1.md L800-847)
    err := kernel.Start()
    if err != nil {
        t.Fatalf("Expected kernel to start, got %v", err)
    }
    defer kernel.Stop()
    
    // Wait for KERNEL_READY (arch-v1.md L830)
    select {
    case <-kernel.KernelReady():
        // KERNEL_READY emitted
    case <-time.After(5 * time.Second):
        t.Fatal("Expected KERNEL_READY event within timeout")
    }
    
    // Verify all 4 hard-coded services registered (arch-v1.md L457-460)
    services := []string{"sys:security", "sys:communication", "sys:observability", "sys:lifecycle"}
    for _, svcName := range services {
        runtimeID, err := kernel.GetServiceRuntimeID(svcName)
        if err != nil {
            t.Errorf("Expected service '%s' to be registered, got %v", svcName, err)
        }
        if runtimeID == "" {
            t.Errorf("Expected non-empty runtime ID for '%s'", svcName)
        }
    }
    
    // Verify handleMail contract for each service (arch-v1.md L479)
    testMail := mail.Mail{
        ID:      "test-001",
        Source:  "agent:tester",
        Target:  "sys:observability",
        Content: map[string]any{"test": "data"},
    }
    
    _, err = kernel.RouteMail(testMail)
    if err != nil {
        t.Errorf("Expected mail routing to succeed, got %v", err)
    }
}
```
**Acceptance Criteria:**
- All 4 hard-coded services can be loaded via bootstrap actions (arch-v1.md L800-847)
- Services are registered in kernel service map
- KERNEL_READY event can be emitted after all services ready

## Dependencies

### Test Dependencies
```
Test 1 → Test 2 (Security service)
Test 3 (Communication service)
Test 4 (Observability service)
Test 5 (Lifecycle service)
Test 6 (All services integration)
```

### Phase Dependencies
- **None** - This is the first phase of Layer 7

## Files to Create/Modify

| File | Operation | Changes |
|------|-----------|---------|
| `pkg/services/security/service_test.go` | MODIFY | Add boundary enforcement, taint propagation tests |
| `pkg/services/communication/service_test.go` | MODIFY | Add publish Ack tests |
| `pkg/services/observability/service_test.go` | MODIFY | Add dead-letter tests |
| `pkg/services/lifecycle/service_test.go` | MODIFY | Add spawn tracking tests |

## Implementation Order (TDD)

1. **Test 1**: Write test → verify RED → Implement boundary enforcement → verify GREEN → commit
2. **Test 2**: Write test → verify RED → Implement taint propagation → verify GREEN → commit
3. **Test 3**: Write test → verify RED → Implement Publish Ack → verify GREEN → commit
4. **Test 4**: Write test → verify RED → Implement dead-letter logging → verify GREEN → commit
5. **Test 5**: Write test → verify RED → Implement spawn tracking → verify GREEN → commit
6. **Test 6**: Write test → verify RED → Integrate all services with bootstrap → verify GREEN → commit

**Total: 6 tests, 6 commits**

## Deliverables
- ✅ `sys:security` boundary enforcement complete (arch-v1.md L457)
- ✅ `sys:communication` pub/sub complete (arch-v1.md L458)
- ✅ `sys:observability` dead-letter complete (arch-v1.md L459)
- ✅ `sys:lifecycle` spawn tracking complete (arch-v1.md L460)
- ✅ All 4 services integrate with bootstrap (arch-v1.md L800-847)
- ✅ 6 commits (one per test per TDD workflow)