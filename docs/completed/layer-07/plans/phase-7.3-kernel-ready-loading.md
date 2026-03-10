# Phase 7.3: Post-KERNEL_READY Loading

## Goal
Implement ChartRegistry loading of hot-reloadable services after KERNEL_READY event (arch-v1.md L831-838, L466-473).

## Scope
- Implement ChartRegistry Service to watch services/ directory
- Load hot-reloadable services after KERNEL_READY (arch-v1.md L831-838)
- Wire ChartRegistry to Kernel for post-bootstrap loading
- Implement service discovery via ChartRegistry
- 6 tests, 6 commits

## Current Status

### Implementation Status
| Component | Status | Notes |
|-----------|--------|-------|
| `ChartRegistry.Service` | ⚠️ Exists | Basic service infrastructure; needs KERNEL_READY integration |
| `Kernel` | ⚠️ Partial | Has bootstrap sequence; needs post-KERNEL_READY loading |
| `services/` directory | ⚠️ YAML exists | All 8 services have YAML definitions |

### Files Status
| File | Status |
|------|--------|
| `pkg/registry/service.go` | ⚠️ Partial - add KERNEL_READY integration |
| `pkg/kernel/kernel.go` | ⚠️ Partial - add post-KERNEL_READY loading |
| `var/maelstrom/services/` | ⚠️ Partial - create directory structure |

## Required Implementation

### ChartRegistry Service Integration (arch-v1.md L831-838)
```go
// pkg/kernel/kernel.go
func (k *Kernel) startChartRegistry(ctx context.Context) error {
    // After KERNEL_READY, start ChartRegistry to load hot-reloadable services
    // sys:gateway, sys:admin, sys:persistence, sys:heartbeat, sys:memory,
    // sys:human-gateway, sys:tools, sys:datasources
}
```

### Service Loading from YAML (arch-v1.md L464-473)
```go
// pkg/registry/service.go
func (s *Service) loadPlatformServices() error {
    // Load all PlatformService YAML files from services/ directory
    // Spawn each as ChartRuntime via lifecycle service
}
```

## Tests to Write (6 tests, 6 commits)

### Test 1: TestPostKernelReadyLoading_ChartRegistryStartsAfterKERNEL_READY
```go
func TestPostKernelReadyLoading_ChartRegistryStartsAfterKERNEL_READY(t *testing.T) {
    kernel := NewKernel()
    
    // Start kernel bootstrap (arch-v1.md L800-828)
    err := kernel.Start()
    if err != nil {
        t.Fatalf("Expected kernel to start, got %v", err)
    }
    defer kernel.Stop()
    
    // Verify ChartRegistry NOT started before KERNEL_READY
    registry := kernel.GetChartRegistry()
    if registry.IsRunning() {
        t.Error("Expected ChartRegistry to NOT be running before KERNEL_READY")
    }
    
    // Wait for KERNEL_READY (arch-v1.md L829)
    select {
    case <-kernel.KernelReady():
        // KERNEL_READY emitted
    case <-time.After(5 * time.Second):
        t.Fatal("Expected KERNEL_READY event within timeout")
    }
    
    // Verify ChartRegistry starts after KERNEL_READY (arch-v1.md L831-838)
    select {
    case <-time.After(1 * time.Second):
        if !registry.IsRunning() {
            t.Error("Expected ChartRegistry to start after KERNEL_READY")
        }
    case <-time.After(5 * time.Second):
        t.Fatal("Expected ChartRegistry to start within timeout")
    }
}
```
**Acceptance Criteria:**
- ChartRegistry.Service starts only after KERNEL_READY event (arch-v1.md L831-838)
- Kernel waits for KERNEL_READY before starting ChartRegistry

### Test 2: TestPostKernelReadyLoading_ServicesLoadedFromDirectory
```go
func TestPostKernelReadyLoading_ServicesLoadedFromDirectory(t *testing.T) {
    kernel := NewKernel()
    
    // Create services directory with YAML files (arch-v1.md L464-473)
    servicesDir := "var/maelstrom/services/"
    os.MkdirAll(servicesDir, 0755)
    defer os.RemoveAll(servicesDir)
    
    // Create all 8 hot-reloadable service YAML files
    serviceNames := []string{
        "sys:gateway", "sys:admin", "sys:persistence", "sys:heartbeat",
        "sys:memory", "sys:human-gateway", "sys:tools", "sys:datasources",
    }
    
    for _, name := range serviceNames {
        fileName := strings.ReplaceAll(name, ":", "-") + ".yaml"
        filePath := filepath.Join(servicesDir, fileName)
        content := fmt.Sprintf(`apiVersion: maelstrom.dev/v1
kind: PlatformService
metadata:
  name: %s
  boundary: outer
`, name)
        os.WriteFile(filePath, []byte(content), 0644)
    }
    
    // Start kernel and wait for KERNEL_READY
    err := kernel.Start()
    if err != nil {
        t.Fatalf("Expected kernel to start, got %v", err)
    }
    defer kernel.Stop()
    
    select {
    case <-kernel.KernelReady():
    case <-time.After(5 * time.Second):
        t.Fatal("Expected KERNEL_READY event within timeout")
    }
    
    // Wait for ChartRegistry to load services
    select {
    case <-time.After(2 * time.Second):
    case <-time.After(5 * time.Second):
        t.Fatal("Expected services to load within timeout")
    }
    
    // Verify all 8 services discovered from directory (arch-v1.md L466-473)
    registry := kernel.GetChartRegistry()
    loadedServices := registry.GetLoadedServices()
    if len(loadedServices) != 8 {
        t.Errorf("Expected 8 services loaded from directory, got %d", len(loadedServices))
    }
    
    for _, name := range serviceNames {
        found := false
        for _, loaded := range loadedServices {
            if loaded == name {
                found = true
                break
            }
        }
        if !found {
            t.Errorf("Expected service '%s' to be loaded from directory", name)
        }
    }
}
```
**Acceptance Criteria:**
- ChartRegistry loads services from `var/maelstrom/services/` directory (arch-v1.md L464-473)
- All 8 hot-reloadable service YAML files are discovered

### Test 3: TestPostKernelReadyLoading_AllHotReloadableServicesSpawn
```go
func TestPostKernelReadyLoading_AllHotReloadableServicesSpawn(t *testing.T) {
    kernel := NewKernel()
    
    // Create services directory with YAML files (arch-v1.md L464-473)
    servicesDir := "var/maelstrom/services/"
    os.MkdirAll(servicesDir, 0755)
    defer os.RemoveAll(servicesDir)
    
    // Create all 8 hot-reloadable service YAML files (arch-v1.md L466-473)
    serviceNames := []string{
        "sys:gateway", "sys:admin", "sys:persistence", "sys:heartbeat",
        "sys:memory", "sys:human-gateway", "sys:tools", "sys:datasources",
    }
    
    for _, name := range serviceNames {
        fileName := strings.ReplaceAll(name, ":", "-") + ".yaml"
        filePath := filepath.Join(servicesDir, fileName)
        content := fmt.Sprintf(`apiVersion: maelstrom.dev/v1
kind: PlatformService
metadata:
  name: %s
  boundary: outer
`, name)
        os.WriteFile(filePath, []byte(content), 0644)
    }
    
    // Start kernel and wait for KERNEL_READY
    err := kernel.Start()
    if err != nil {
        t.Fatalf("Expected kernel to start, got %v", err)
    }
    defer kernel.Stop()
    
    select {
    case <-kernel.KernelReady():
    case <-time.After(5 * time.Second):
        t.Fatal("Expected KERNEL_READY event within timeout")
    }
    
    // Wait for services to spawn as ChartRuntimes
    select {
    case <-time.After(2 * time.Second):
    case <-time.After(5 * time.Second):
        t.Fatal("Expected services to spawn within timeout")
    }
    
    // Verify all 8 services spawned as ChartRuntimes (arch-v1.md L466-473)
    lifecycle := kernel.GetLifecycleService()
    runtimes := lifecycle.List()
    
    spawnedServices := make(map[string]bool)
    for _, runtime := range runtimes {
        spawnedServices[runtime.DefinitionID] = true
    }
    
    for _, name := range serviceNames {
        if !spawnedServices[name] {
            t.Errorf("Expected service '%s' to be spawned as ChartRuntime", name)
        }
    }
}
```
**Acceptance Criteria:**
- All 8 hot-reloadable services are spawned as ChartRuntimes (arch-v1.md L466-473)
- Services: sys:gateway, sys:admin, sys:persistence, sys:heartbeat, sys:memory, sys:human-gateway, sys:tools, sys:datasources

### Test 4: TestPostKernelReadyLoading_ServicesRegisteredInKernel
```go
func TestPostKernelReadyLoading_ServicesRegisteredInKernel(t *testing.T) {
    kernel := NewKernel()
    
    // Create services directory with YAML files
    servicesDir := "var/maelstrom/services/"
    os.MkdirAll(servicesDir, 0755)
    defer os.RemoveAll(servicesDir)
    
    // Create all 8 hot-reloadable service YAML files
    serviceNames := []string{
        "sys:gateway", "sys:admin", "sys:persistence", "sys:heartbeat",
        "sys:memory", "sys:human-gateway", "sys:tools", "sys:datasources",
    }
    
    for _, name := range serviceNames {
        fileName := strings.ReplaceAll(name, ":", "-") + ".yaml"
        filePath := filepath.Join(servicesDir, fileName)
        content := fmt.Sprintf(`apiVersion: maelstrom.dev/v1
kind: PlatformService
metadata:
  name: %s
  boundary: outer
`, name)
        os.WriteFile(filePath, []byte(content), 0644)
    }
    
    // Start kernel and wait for KERNEL_READY
    err := kernel.Start()
    if err != nil {
        t.Fatalf("Expected kernel to start, got %v", err)
    }
    defer kernel.Stop()
    
    select {
    case <-kernel.KernelReady():
    case <-time.After(5 * time.Second):
        t.Fatal("Expected KERNEL_READY event within timeout")
    }
    
    // Wait for services to register
    select {
    case <-time.After(2 * time.Second):
    case <-time.After(5 * time.Second):
        t.Fatal("Expected services to register within timeout")
    }
    
    // Verify all services registered in Kernel service map
    for _, name := range serviceNames {
        runtimeID, err := kernel.GetServiceRuntimeID(name)
        if err != nil {
            t.Errorf("Expected service '%s' to be registered, got %v", name, err)
        }
        if runtimeID == "" {
            t.Errorf("Expected non-empty RuntimeID for service '%s'", name)
        }
    }
}
```
**Acceptance Criteria:**
- Spawned services are registered in Kernel service map
- `GetServiceRuntimeID()` returns RuntimeID for each service

### Test 5: TestPostKernelReadyLoading_ServicesHandleMail
```go
func TestPostKernelReadyLoading_ServicesHandleMail(t *testing.T) {
    kernel := NewKernel()
    
    // Create services directory with YAML files
    servicesDir := "var/maelstrom/services/"
    os.MkdirAll(servicesDir, 0755)
    defer os.RemoveAll(servicesDir)
    
    // Create all 8 hot-reloadable service YAML files
    serviceNames := []string{
        "sys:gateway", "sys:admin", "sys:persistence", "sys:heartbeat",
        "sys:memory", "sys:human-gateway", "sys:tools", "sys:datasources",
    }
    
    for _, name := range serviceNames {
        fileName := strings.ReplaceAll(name, ":", "-") + ".yaml"
        filePath := filepath.Join(servicesDir, fileName)
        content := fmt.Sprintf(`apiVersion: maelstrom.dev/v1
kind: PlatformService
metadata:
  name: %s
  boundary: outer
`, name)
        os.WriteFile(filePath, []byte(content), 0644)
    }
    
    // Start kernel and wait for KERNEL_READY
    err := kernel.Start()
    if err != nil {
        t.Fatalf("Expected kernel to start, got %v", err)
    }
    defer kernel.Stop()
    
    select {
    case <-kernel.KernelReady():
    case <-time.After(5 * time.Second):
        t.Fatal("Expected KERNEL_READY event within timeout")
    }
    
    // Wait for services to be ready
    select {
    case <-time.After(2 * time.Second):
    case <-time.After(5 * time.Second):
        t.Fatal("Expected services to be ready within timeout")
    }
    
    // Verify each service implements handleMail contract (arch-v1.md L479)
    for _, name := range serviceNames {
        testMail := mail.Mail{
            ID:       fmt.Sprintf("test-%s", name),
            Source:   "agent:tester",
            Target:   name,
            Content:  map[string]any{"test": "data"},
            CreatedAt: time.Now(),
        }
        
        // Route mail to service via CommunicationService (arch-v1.md L479)
        outcome, err := kernel.RouteMail(testMail)
        if err != nil {
            t.Errorf("Expected mail routing to '%s' to succeed, got %v", name, err)
        }
        
        // Verify outcome event returned (arch-v1.md L479)
        if outcome == nil {
            t.Errorf("Expected outcome event from '%s' handleMail", name)
        }
    }
}
```
**Acceptance Criteria:**
- Each service implements `handleMail(mail: Mail) → outcomeEvent` contract (arch-v1.md L479)
- Services can receive and process Mail via CommunicationService

### Test 6: TestPostKernelReadyLoading_KernelGoesDormant
```go
func TestPostKernelReadyLoading_KernelGoesDormant(t *testing.T) {
    kernel := NewKernel()
    
    // Create services directory with YAML files
    servicesDir := "var/maelstrom/services/"
    os.MkdirAll(servicesDir, 0755)
    defer os.RemoveAll(servicesDir)
    
    // Create all 8 hot-reloadable service YAML files
    serviceNames := []string{
        "sys:gateway", "sys:admin", "sys:persistence", "sys:heartbeat",
        "sys:memory", "sys:human-gateway", "sys:tools", "sys:datasources",
    }
    
    for _, name := range serviceNames {
        fileName := strings.ReplaceAll(name, ":", "-") + ".yaml"
        filePath := filepath.Join(servicesDir, fileName)
        content := fmt.Sprintf(`apiVersion: maelstrom.dev/v1
kind: PlatformService
metadata:
  name: %s
  boundary: outer
`, name)
        os.WriteFile(filePath, []byte(content), 0644)
    }
    
    // Start kernel and wait for KERNEL_READY
    err := kernel.Start()
    if err != nil {
        t.Fatalf("Expected kernel to start, got %v", err)
    }
    defer kernel.Stop()
    
    select {
    case <-kernel.KernelReady():
    case <-time.After(5 * time.Second):
        t.Fatal("Expected KERNEL_READY event within timeout")
    }
    
    // Wait for ChartRegistry handoff and kernel dormancy (arch-v1.md L839-840)
    select {
    case <-kernel.IsDormant():
        // Kernel entered dormant state
    case <-time.After(5 * time.Second):
        t.Fatal("Expected kernel to go dormant within timeout")
    }
    
    // Verify kernel is dormant (arch-v1.md L839-840)
    if !kernel.IsDormant() {
        t.Error("Expected kernel to be in dormant state after handoff to ChartRegistry")
    }
    
    // Verify kernel only listens for shutdown signals (arch-v1.md L839-840)
    if kernel.IsProcessingEvents() {
        t.Error("Expected dormant kernel to NOT process events")
    }
    
    // Verify shutdown signal handling still works
    shutdownChan := make(chan struct{})
    kernel.RegisterShutdownHandler(func() {
        close(shutdownChan)
    })
    
    kernel.SendShutdownSignal()
    select {
    case <-shutdownChan:
        // Shutdown handler called
    case <-time.After(1 * time.Second):
        t.Error("Expected shutdown handler to be called")
    }
}
```
**Acceptance Criteria:**
- Kernel enters dormant state after handoff to ChartRegistry (arch-v1.md L839-840)
- Kernel only listens for shutdown signals

## Dependencies

### Test Dependencies
```
Test 1 → Test 2 (ChartRegistry startup)
Test 3 (Service spawning)
Test 4 (Service registration)
Test 5 (Service mail handling)
Test 6 (Kernel dormant state)
```

### Phase Dependencies
- **Phase 7.1** - Hard-coded services must be complete first
- **Phase 7.2** - Hot-reloadable services must be complete first
- **Phase 7.4** depends on this phase completing

## Files to Create/Modify

| File | Operation | Changes |
|------|-----------|---------|
| `pkg/kernel/kernel.go` | MODIFY | Add post-KERNEL_READY ChartRegistry loading |
| `pkg/registry/service.go` | MODIFY | Add platform service loading from YAML |
| `var/maelstrom/services/` | CREATE | Directory structure for service YAML files |

## Implementation Order (TDD)

1. **Test 1**: Write test → verify RED → Implement KERNEL_READY gating → verify GREEN → commit
2. **Test 2**: Write test → verify RED → Implement directory-based loading → verify GREEN → commit
3. **Test 3**: Write test → verify RED → Implement service spawning → verify GREEN → commit
4. **Test 4**: Write test → verify RED → Implement service registration → verify GREEN → commit
5. **Test 5**: Write test → verify RED → Implement handleMail contract → verify GREEN → commit
6. **Test 6**: Write test → verify RED → Implement kernel dormant state → verify GREEN → commit

**Total: 6 tests, 6 commits**

## Deliverables
- ✅ ChartRegistry starts after KERNEL_READY (arch-v1.md L831-838)
- ✅ All 8 hot-reloadable services loaded from YAML (arch-v1.md L466-473)
- ✅ Services spawned as ChartRuntimes
- ✅ Services registered in Kernel
- ✅ Services implement handleMail contract (arch-v1.md L479)
- ✅ Kernel goes dormant after handoff (arch-v1.md L839-840)
- ✅ 6 commits (one per test per TDD workflow)

## Recursive Analysis Decision

**Decision: Keep as-is (no split)**

**Reasoning:**
- 6 tests is within acceptable range (2-5 recommended, 6 is close)
- Tests are tightly coupled (ChartRegistry → loading → spawning → registration → mail handling → dormancy)
- Single coherent feature: Post-KERNEL_READY loading flow
- Splitting would create unnecessary fragmentation across the loading sequence

**Alternative (if split needed):**
- 7.3a: ChartRegistry startup and service loading - 3 tests
- 7.3b: Service registration, mail handling, and kernel dormancy - 3 tests