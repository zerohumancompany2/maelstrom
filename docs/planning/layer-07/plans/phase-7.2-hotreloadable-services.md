# Phase 7.2: Hot-Reloadable Services Completion

## Goal
Complete the remaining 15% of hot-reloadable platform services (sys:gateway, sys:admin, sys:persistence, sys:heartbeat, sys:memory, sys:human-gateway, sys:tools, sys:datasources) per spec (arch-v1.md L463-474).

## Scope
- Complete sys:gateway channel adapters (arch-v1.md L466)
- Complete sys:admin 2FA-gated commands (arch-v1.md L467, L485)
- Complete sys:persistence snapshot/restore (arch-v1.md L468, L486)
- Complete sys:heartbeat scheduled wake-ups (arch-v1.md L469)
- Complete sys:memory vector/graph stores (arch-v1.md L470, L488)
- Complete sys:human-gateway chat interface (arch-v1.md L471)
- Complete sys:tools registry and resolution (arch-v1.md L472, L487)
- Complete sys:datasources pluggable storage (arch-v1.md L473, L489)
- 10 tests, 10 commits

## Current Status

### Implementation Status
| Service | Status | Notes |
|---------|--------|-------|
| `sys:gateway` | ⚠️ 85% complete | Adapters exist; needs OpenAPI generation tests |
| `sys:admin` | ⚠️ 85% complete | 2FA exists; needs boundary enforcement tests |
| `sys:persistence` | ⚠️ 85% complete | Storage exists; needs snapshot/restore tests |
| `sys:heartbeat` | ⚠️ 85% complete | Scheduler exists; needs HEARTBEAT.md injection tests |
| `sys:memory` | ⚠️ 85% complete | Vector/graph exists; needs ContextMap injection tests |
| `sys:human-gateway` | ⚠️ 85% complete | Chat exists; needs streaming tests |
| `sys:tools` | ⚠️ 85% complete | Registry exists; needs ToolDescriptor tests |
| `sys:datasources` | ⚠️ 85% complete | Storage exists; needs tainting tests |

### Files Status
| File | Status |
|------|--------|
| `pkg/services/gateway/service.go` | ⚠️ Partial - add OpenAPI generation |
| `pkg/services/admin/service.go` | ⚠️ Partial - add boundary enforcement |
| `pkg/services/persistence/service.go` | ⚠️ Partial - add snapshot/restore |
| `pkg/services/heartbeat/service.go` | ⚠️ Partial - add HEARTBEAT.md injection |
| `pkg/services/memory/service.go` | ⚠️ Partial - add ContextMap injection |
| `pkg/services/humangateway/service.go` | ⚠️ Partial - add streaming |
| `pkg/services/tools/service.go` | ⚠️ Partial - add ToolDescriptor tests |
| `pkg/services/datasources/service.go` | ⚠️ Partial - add tainting |

## Required Implementation

### sys:admin 2FA-Gated Commands (arch-v1.md L467, L485)
```go
// pkg/services/admin/service.go
func (s *adminService) ExecuteCommand(cmd string, token string) error {
    // Verify 2FA token, enforce outer-boundary-only
}
```

### sys:tools ToolDescriptor Resolution (arch-v1.md L487)
```go
// pkg/services/tools/service.go
func (s *toolsService) Resolve(name string, callerBoundary string) (ToolDescriptor, error) {
    // Resolve tool by name, check boundary accessibility
}
```

### sys:persistence Snapshot/Restore (arch-v1.md L486)
```go
// pkg/services/persistence/service.go
func (s *persistenceService) Snapshot(runtimeId string) (string, error) {
    // Create snapshot with taints and session pointer
}

func (s *persistenceService) Restore(snapshotId string) error {
    // Restore runtime state from snapshot
}
```

### sys:heartbeat HEARTBEAT.md Injection (arch-v1.md L469)
```go
// pkg/services/heartbeat/service.go
func (s *heartbeatService) InjectHEARTBEAT(runtimeId string) error {
    // Inject HEARTBEAT.md content into runtime context
}
```

### sys:memory ContextMap Injection (arch-v1.md L470, L488)
```go
// pkg/services/memory/service.go
func (s *memoryService) Query(vector any, topK int, boundaryFilter string) ([]MessageSlice, error) {
    // Query vector/graph stores, return MessageSlice for ContextMap injection
}
```

### sys:datasources Tainting (arch-v1.md L473, L489)
```go
// pkg/services/datasources/service.go
func (s *datasourcesService) TagOnWrite(path string, taints []string) error {
    // Tag file with taints in xattr or object tags
}

func (s *datasourcesService) GetTaints(path string) ([]string, error) {
    // Return []string of taints for path
}
```

## Tests to Write (10 tests, 10 commits)

### Test 1: TestHotreloadableServices_Admin2FAEnforcement
```go
func TestHotreloadableServices_Admin2FAEnforcement(t *testing.T) {
    svc := NewAdminService()
    
    // Test empty token rejected (arch-v1.md L467, L485)
    err := svc.ExecuteCommand("list", "")
    if err == nil {
        t.Error("Expected error for empty 2FA token")
    }
    
    // Test invalid token rejected (arch-v1.md L467)
    err = svc.ExecuteCommand("list", "invalid-token")
    if err == nil {
        t.Error("Expected error for invalid 2FA token")
    }
    
    // Test valid token accepted (arch-v1.md L485)
    validToken := "valid-2fa-token-12345"
    err = svc.ExecuteCommand("list", validToken)
    if err != nil {
        t.Errorf("Expected no error for valid 2FA token, got %v", err)
    }
    
    // Test outer-boundary-only enforcement (arch-v1.md L467)
    err = svc.ExecuteCommandOnBoundary("list", validToken, "inner")
    if err == nil {
        t.Error("Expected error for inner boundary access")
    }
    
    err = svc.ExecuteCommandOnBoundary("list", validToken, "outer")
    if err != nil {
        t.Errorf("Expected no error for outer boundary access, got %v", err)
    }
}
```
**Acceptance Criteria:**
- `ExecuteCommand()` requires non-empty 2FA token (arch-v1.md L467, L485)
- `ExecuteCommand()` rejects invalid 2FA tokens
- `ExecuteCommandOnBoundary()` enforces outer-boundary-only

### Test 2: TestHotreloadableServices_AdminCommandTypes
```go
func TestHotreloadableServices_AdminCommandTypes(t *testing.T) {
    svc := NewAdminService()
    token := "valid-2fa-token"
    
    // Test list command (arch-v1.md L485)
    err := svc.ExecuteCommand("list", token)
    if err != nil {
        t.Errorf("Expected list command to be accepted, got %v", err)
    }
    
    // Test control command (arch-v1.md L485)
    err = svc.ExecuteCommand("control", token)
    if err != nil {
        t.Errorf("Expected control command to be accepted, got %v", err)
    }
    
    // Test queryTaints command (arch-v1.md L485)
    err = svc.ExecuteCommand("queryTaints", token)
    if err != nil {
        t.Errorf("Expected queryTaints command to be accepted, got %v", err)
    }
    
    // Test inject command (arch-v1.md L485)
    err = svc.ExecuteCommand("inject", token)
    if err != nil {
        t.Errorf("Expected inject command to be accepted, got %v", err)
    }
    
    // Test invalid command rejected
    err = svc.ExecuteCommand("invalidCommand", token)
    if err == nil {
        t.Error("Expected error for invalid command type")
    }
}
```
**Acceptance Criteria:**
- `ExecuteCommand()` accepts list, control, queryTaints, inject commands (arch-v1.md L485)
- Each command type is validated
- Invalid command types are rejected

### Test 3: TestHotreloadableServices_ToolsToolDescriptor
```go
func TestHotreloadableServices_ToolsToolDescriptor(t *testing.T) {
    svc := NewToolsService()
    
    // Register a tool (arch-v1.md L487)
    tool := ToolDescriptor{
        Name:          "webSearch",
        Boundary:      "dmz",
        Schema:        map[string]any{"type": "object", "properties": map[string]any{"query": map[string]any{"type": "string"}}},
        Isolation:     "container",
        TaintOutput:   []string{"TOOL_OUTPUT", "EXTERNAL"},
        Type:          "tool",
        ChartRef:      "",
        MaxIterations: 0,
    }
    err := svc.Register(tool)
    if err != nil {
        t.Fatalf("Expected no error registering tool, got %v", err)
    }
    
    // Resolve tool (arch-v1.md L487)
    resolved, err := svc.Resolve("webSearch", "inner")
    if err != nil {
        t.Fatalf("Expected no error resolving tool, got %v", err)
    }
    
    // Verify ToolDescriptor fields (arch-v1.md L487)
    if resolved.Name != "webSearch" {
        t.Errorf("Expected Name 'webSearch', got '%s'", resolved.Name)
    }
    if resolved.Boundary != "dmz" {
        t.Errorf("Expected Boundary 'dmz', got '%s'", resolved.Boundary)
    }
    if resolved.Isolation != "container" {
        t.Errorf("Expected Isolation 'container', got '%s'", resolved.Isolation)
    }
    if len(resolved.TaintOutput) != 2 {
        t.Errorf("Expected 2 TaintOutput entries, got %d", len(resolved.TaintOutput))
    }
    if resolved.Type != "tool" {
        t.Errorf("Expected Type 'tool', got '%s'", resolved.Type)
    }
}
```
**Acceptance Criteria:**
- `Resolve()` returns ToolDescriptor with Name, Boundary, Schema, Isolation (arch-v1.md L487)
- ToolDescriptor has TaintOutput, Type, ChartRef, MaxIterations fields

### Test 4: TestHotreloadableServices_ToolsBoundaryFiltering
```go
func TestHotreloadableServices_ToolsBoundaryFiltering(t *testing.T) {
    svc := NewToolsService()
    
    // Register tools at different boundaries (arch-v1.md L487)
    innerTool := ToolDescriptor{Name: "innerDb", Boundary: "inner", Schema: map[string]any{}, Isolation: "strict"}
    dmzTool := ToolDescriptor{Name: "webSearch", Boundary: "dmz", Schema: map[string]any{}, Isolation: "container"}
    outerTool := ToolDescriptor{Name: "publicApi", Boundary: "outer", Schema: map[string]any{}, Isolation: "process"}
    
    svc.Register(innerTool)
    svc.Register(dmzTool)
    svc.Register(outerTool)
    
    // Test inner boundary can access all (arch-v1.md L487)
    _, err := svc.Resolve("innerDb", "inner")
    if err != nil {
        t.Errorf("Inner should access inner tool, got %v", err)
    }
    _, err = svc.Resolve("webSearch", "inner")
    if err != nil {
        t.Errorf("Inner should access dmz tool, got %v", err)
    }
    _, err = svc.Resolve("publicApi", "inner")
    if err != nil {
        t.Errorf("Inner should access outer tool, got %v", err)
    }
    
    // Test dmz boundary can access dmz/outer only (arch-v1.md L487)
    _, err = svc.Resolve("innerDb", "dmz")
    if err == nil {
        t.Error("DMZ should not access inner tool")
    }
    _, err = svc.Resolve("webSearch", "dmz")
    if err != nil {
        t.Errorf("DMZ should access dmz tool, got %v", err)
    }
    _, err = svc.Resolve("publicApi", "dmz")
    if err != nil {
        t.Errorf("DMZ should access outer tool, got %v", err)
    }
    
    // Test outer boundary can access outer only (arch-v1.md L487)
    _, err = svc.Resolve("innerDb", "outer")
    if err == nil {
        t.Error("Outer should not access inner tool")
    }
    _, err = svc.Resolve("webSearch", "outer")
    if err == nil {
        t.Error("Outer should not access dmz tool")
    }
    _, err = svc.Resolve("publicApi", "outer")
    if err != nil {
        t.Errorf("Outer should access outer tool, got %v", err)
    }
}
```
**Acceptance Criteria:**
- `Resolve()` checks caller boundary against tool boundary (arch-v1.md L487)
- Inner can access inner/dmz/outer tools
- DMZ can access dmz/outer tools
- Outer can access outer tools only

### Test 5: TestHotreloadableServices_PersistenceSnapshotRestore
```go
func TestHotreloadableServices_PersistenceSnapshotRestore(t *testing.T) {
    svc := NewPersistenceService()
    runtimeId := "test-runtime-001"
    
    // Create snapshot (arch-v1.md L486)
    snapshotId, err := svc.Snapshot(runtimeId)
    if err != nil {
        t.Fatalf("Expected no error creating snapshot, got %v", err)
    }
    if snapshotId == "" {
        t.Error("Expected non-empty snapshot ID")
    }
    
    // Verify snapshot includes taints and session pointer (arch-v1.md L486)
    snapshot, err := svc.GetSnapshot(snapshotId)
    if err != nil {
        t.Fatalf("Expected no error getting snapshot, got %v", err)
    }
    if snapshot.RuntimeID != runtimeId {
        t.Errorf("Expected RuntimeID '%s', got '%s'", runtimeId, snapshot.RuntimeID)
    }
    if snapshot.Taints == nil {
        t.Error("Expected snapshot to include taints")
    }
    if snapshot.SessionPointer == "" {
        t.Error("Expected snapshot to include session pointer")
    }
    
    // Restore from snapshot (arch-v1.md L486)
    err = svc.Restore(snapshotId)
    if err != nil {
        t.Errorf("Expected no error restoring snapshot, got %v", err)
    }
    
    // Verify restored state
    restored, err := svc.GetRuntimeState(runtimeId)
    if err != nil {
        t.Fatalf("Expected no error getting runtime state, got %v", err)
    }
    if restored.RestoredFrom != snapshotId {
        t.Errorf("Expected restored from '%s', got '%s'", snapshotId, restored.RestoredFrom)
    }
}
```
**Acceptance Criteria:**
- `Snapshot(runtimeId)` creates snapshot (arch-v1.md L486)
- `Restore(snapshotId)` restores runtime state
- Snapshot includes taints and session pointer

### Test 6: TestHotreloadableServices_HeartbeatHEARTBEATInjection
```go
func TestHotreloadableServices_HeartbeatHEARTBEATInjection(t *testing.T) {
    svc := NewHeartbeatService()
    runtimeId := "test-runtime-001"
    
    // Setup scheduled wake-up (arch-v1.md L469)
    interval := 5 * time.Second
    err := svc.ScheduleWakeUp(runtimeId, interval)
    if err != nil {
        t.Fatalf("Expected no error scheduling wake-up, got %v", err)
    }
    
    // Inject HEARTBEAT.md content (arch-v1.md L469)
    err = svc.InjectHEARTBEAT(runtimeId)
    if err != nil {
        t.Errorf("Expected no error injecting HEARTBEAT, got %v", err)
    }
    
    // Verify HEARTBEAT.md content was injected
    injected, err := svc.GetInjectedContent(runtimeId)
    if err != nil {
        t.Fatalf("Expected no error getting injected content, got %v", err)
    }
    if injected.Type != "HEARTBEAT" {
        t.Errorf("Expected content type 'HEARTBEAT', got '%s'", injected.Type)
    }
    if injected.Timestamp.IsZero() {
        t.Error("Expected injection timestamp to be set")
    }
    
    // Verify scheduled wake-ups occur at configured interval
    select {
    case <-svc.NextWakeUp(runtimeId):
        // Wake-up occurred as expected
    case <-time.After(interval + 1*time.Second):
        t.Error("Expected wake-up to occur within configured interval")
    }
}
```
**Acceptance Criteria:**
- Heartbeat service injects HEARTBEAT.md content (arch-v1.md L469)
- Scheduled wake-ups occur at configured interval

### Test 7: TestHotreloadableServices_MemoryContextMapInjection
```go
func TestHotreloadableServices_MemoryContextMapInjection(t *testing.T) {
    svc := NewMemoryService()
    
    // Insert test vectors (arch-v1.md L470, L488)
    testVector := []float32{0.1, 0.2, 0.3, 0.4}
    msg := MessageSlice{
        ID:       "msg-001",
        Content:  "test memory content",
        Boundary: "dmz",
        Taints:   []string{"USER_SUPPLIED"},
    }
    err := svc.Insert(testVector, msg)
    if err != nil {
        t.Fatalf("Expected no error inserting vector, got %v", err)
    }
    
    // Query vector store (arch-v1.md L488)
    queryVector := []float32{0.11, 0.21, 0.31, 0.41}
    results, err := svc.Query(queryVector, 5, "dmz")
    if err != nil {
        t.Fatalf("Expected no error querying, got %v", err)
    }
    
    // Verify results (arch-v1.md L488)
    if len(results) == 0 {
        t.Error("Expected at least one result from query")
    }
    
    // Verify boundaryFilter respected (arch-v1.md L488)
    if results[0].Boundary != "dmz" {
        t.Errorf("Expected boundary 'dmz', got '%s'", results[0].Boundary)
    }
    
    // Test boundaryFilter blocks inner content
    innerMsg := MessageSlice{ID: "msg-002", Content: "secret", Boundary: "inner", Taints: []string{"INNER_ONLY"}}
    svc.Insert(testVector, innerMsg)
    
    filteredResults, err := svc.Query(queryVector, 10, "outer")
    if err != nil {
        t.Fatalf("Expected no error querying with outer filter, got %v", err)
    }
    for _, r := range filteredResults {
        if r.Boundary == "inner" {
            t.Error("Expected inner boundary content to be filtered out for outer query")
        }
    }
}
```
**Acceptance Criteria:**
- Memory service queries vector/graph stores (arch-v1.md L470, L488)
- Returns MessageSlice for ContextMap injection
- Respects boundaryFilter in query

### Test 8: TestHotreloadableServices_DatasourcesTainting
```go
func TestHotreloadableServices_DatasourcesTainting(t *testing.T) {
    svc := NewDatasourcesService()
    testPath := "/tmp/test-file.txt"
    
    // Create test file
    err := os.WriteFile(testPath, []byte("test content"), 0644)
    if err != nil {
        t.Fatalf("Expected no error creating test file, got %v", err)
    }
    defer os.Remove(testPath)
    
    // Tag on write (arch-v1.md L473, L489)
    taints := []string{"USER_SUPPLIED", "PII"}
    err = svc.TagOnWrite(testPath, taints)
    if err != nil {
        t.Errorf("Expected no error tagging file, got %v", err)
    }
    
    // Get taints (arch-v1.md L489)
    retrievedTaints, err := svc.GetTaints(testPath)
    if err != nil {
        t.Fatalf("Expected no error getting taints, got %v", err)
    }
    
    // Verify taints stored correctly (arch-v1.md L473)
    if len(retrievedTaints) != 2 {
        t.Errorf("Expected 2 taints, got %d", len(retrievedTaints))
    }
    if !slices.Contains(retrievedTaints, "USER_SUPPLIED") {
        t.Error("Expected USER_SUPPLIED taint")
    }
    if !slices.Contains(retrievedTaints, "PII") {
        t.Error("Expected PII taint")
    }
    
    // Verify taints stored in xattr or object tags
    xattrTaints, err := svc.getXattrTaints(testPath)
    if err != nil {
        t.Logf("xattr not available, testing fallback storage")
    } else {
        if len(xattrTaints) == 0 {
            t.Error("Expected taints to be stored in xattr")
        }
    }
}
```
**Acceptance Criteria:**
- `tagOnWrite(path, taints)` tags file with taints (arch-v1.md L473, L489)
- `getTaints(path)` returns []string of taints
- Taints stored in xattr or object tags

### Test 9: TestHumanGateway_HandleMail_ChatInterface
```go
func TestHumanGateway_HandleMail_ChatInterface(t *testing.T) {
    svc := NewHumanGatewayService()
    agentID := "test-agent-001"
    
    // Create chat session (arch-v1.md L471)
    session, err := svc.CreateSession(agentID)
    if err != nil {
        t.Fatalf("Expected no error creating session, got %v", err)
    }
    if session.AgentID != agentID {
        t.Errorf("Expected AgentID '%s', got '%s'", agentID, session.AgentID)
    }
    
    // Handle mail to open chat interface (arch-v1.md L471)
    mail := Mail{
        ID:      "mail-001",
        Type:    "human_feedback",
        Source:  "human",
        Target:  "agent:" + agentID,
        Content: "Hello, agent!",
    }
    err = svc.HandleMail(mail)
    if err != nil {
        t.Errorf("Expected no error handling mail, got %v", err)
    }
    
    // Verify session updated with message
    updatedSession, err := svc.GetSession(agentID)
    if err != nil {
        t.Fatalf("Expected no error getting session, got %v", err)
    }
    if len(updatedSession.Messages) == 0 {
        t.Error("Expected session to have messages")
    }
    
    // Verify boundary sanitization (arch-v1.md L471)
    if updatedSession.Messages[0].Type != "human_feedback" {
        t.Errorf("Expected message type 'human_feedback', got '%s'", updatedSession.Messages[0].Type)
    }
}
```
**Acceptance Criteria:**
- `HandleMail(mail)` processes chat interface mail (arch-v1.md L471)
- `CreateSession(agentID)` creates chat session for agent
- Session receives read-only snapshot of Agent's ContextMap + last N Messages
- Messages sanitized by boundary rules

### Test 10: TestHumanGateway_HandleMail_HumanInTheLoop
```go
func TestHumanGateway_HandleMail_HumanInTheLoop(t *testing.T) {
    svc := NewHumanGatewayService()
    agentID := "test-agent-001"
    
    // Create session for human-in-the-loop (arch-v1.md L471)
    session, err := svc.CreateSession(agentID)
    if err != nil {
        t.Fatalf("Expected no error creating session, got %v", err)
    }
    
    // Send human feedback message (arch-v1.md L471)
    feedbackMail := Mail{
        ID:      "feedback-001",
        Type:    "human_feedback",
        Source:  "human",
        Target:  "agent:" + agentID,
        Content: "@pause - need to review",
    }
    err = svc.SendMessage(agentID, feedbackMail)
    if err != nil {
        t.Errorf("Expected no error sending message, got %v", err)
    }
    
    // Verify message delivered to agent inbox (arch-v1.md L471)
    err = svc.HandleMail(feedbackMail)
    if err != nil {
        t.Errorf("Expected no error handling mail, got %v", err)
    }
    
    // Verify action item shorthand parsed (arch-v1.md L471)
    updatedSession, err := svc.GetSession(agentID)
    if err != nil {
        t.Fatalf("Expected no error getting session, got %v", err)
    }
    if len(updatedSession.Messages) == 0 {
        t.Error("Expected session to have messages")
    }
    
    // Verify @pause action item becomes special mail
    if updatedSession.Messages[0].Content != "@pause - need to review" {
        t.Errorf("Expected action item content, got '%s'", updatedSession.Messages[0].Content)
    }
    
    // Test agent reply via normal mail (arch-v1.md L471)
    agentReply := Mail{
        ID:      "reply-001",
        Type:    "assistant",
        Source:  "agent:" + agentID,
        Target:  "human",
        Content: "Paused as requested",
    }
    err = svc.HandleMail(agentReply)
    if err != nil {
        t.Errorf("Expected no error handling agent reply, got %v", err)
    }
    
    // Verify reply rendered in chat UI
    finalSession, err := svc.GetSession(agentID)
    if err != nil {
        t.Fatalf("Expected no error getting session, got %v", err)
    }
    if len(finalSession.Messages) < 2 {
        t.Error("Expected session to have both human and agent messages")
    }
}
```
**Acceptance Criteria:**
- `HandleMail(mail)` processes human-in-the-loop mail (arch-v1.md L471)
- Messages sent become `mail_received` (type: `human_feedback` or `user`) delivered to Agent's inbox
- Agent replies via normal mail rendered back in chat UI
- Optional "action item" shorthand (`@pause`, `@inject-memory X`) become special Mail messages

## Dependencies

### Test Dependencies
```
Test 1 → Test 2 (Admin service)
Test 3 → Test 4 (Tools service)
Test 5 (Persistence service)
Test 6 (Heartbeat service)
Test 7 (Memory service)
Test 8 (Datasources service)
Test 9 → Test 10 (HumanGateway service)
```

### Phase Dependencies
- **Phase 7.1** - Hard-coded services completion (provides service infrastructure)

## Files to Create/Modify

| File | Operation | Changes |
|------|-----------|---------|
| `pkg/services/admin/service_test.go` | MODIFY | Add 2FA enforcement, command type tests |
| `pkg/services/tools/service_test.go` | MODIFY | Add ToolDescriptor, boundary filtering tests |
| `pkg/services/persistence/service_test.go` | MODIFY | Add snapshot/restore tests |
| `pkg/services/heartbeat/service_test.go` | MODIFY | Add HEARTBEAT.md injection tests |
| `pkg/services/memory/service_test.go` | MODIFY | Add ContextMap injection tests |
| `pkg/services/datasources/service_test.go` | MODIFY | Add tainting tests |
| `pkg/services/humangateway/service_test.go` | MODIFY | Add chat interface, human-in-the-loop tests |

## Implementation Order (TDD)

1. **Test 1**: Write test → verify RED → Implement 2FA enforcement → verify GREEN → commit
2. **Test 2**: Write test → verify RED → Implement command type validation → verify GREEN → commit
3. **Test 3**: Write test → verify RED → Implement ToolDescriptor → verify GREEN → commit
4. **Test 4**: Write test → verify RED → Implement boundary filtering → verify GREEN → commit
5. **Test 5**: Write test → verify RED → Implement snapshot/restore → verify GREEN → commit
6. **Test 6**: Write test → verify RED → Implement HEARTBEAT.md injection → verify GREEN → commit
7. **Test 7**: Write test → verify RED → Implement ContextMap injection → verify GREEN → commit
8. **Test 8**: Write test → verify RED → Implement datasources tainting → verify GREEN → commit
9. **Test 9**: Write test → verify RED → Implement chat interface → verify GREEN → commit
10. **Test 10**: Write test → verify RED → Implement human-in-the-loop → verify GREEN → commit

**Total: 10 tests, 10 commits**

## Deliverables
- ✅ `sys:admin` 2FA-gated commands complete (arch-v1.md L467, L485)
- ✅ `sys:tools` ToolDescriptor resolution complete (arch-v1.md L472, L487)
- ✅ `sys:persistence` snapshot/restore complete (arch-v1.md L468, L486)
- ✅ `sys:heartbeat` HEARTBEAT.md injection complete (arch-v1.md L469)
- ✅ `sys:memory` ContextMap injection complete (arch-v1.md L470, L488)
- ✅ `sys:human-gateway` chat interface complete (arch-v1.md L471)
- ✅ `sys:datasources` tainting complete (arch-v1.md L473, L489)
- ✅ 10 commits (one per test per TDD workflow)