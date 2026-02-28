# Maelstrom Phase 1 Usage Examples

This document provides practical usage examples for the Phase 1 kernel bootstrap implementation.

## Table of Contents

1. [File Watching & Hot Reload](#1-file-watching--hot-reload)
2. [Registry Operations](#2-registry-operations)
3. [Runtime Management](#3-runtime-management)
4. [Bootstrap Extension](#4-bootstrap-extension)
5. [Testing Patterns](#5-testing-patterns)

---

## 1. File Watching & Hot Reload

This example demonstrates creating a ChartRegistry, adding observer callbacks, and watching charts get loaded and updated in real-time.

```go
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/maelstrom/v3/pkg/chart"
)

// ExampleFileWatchingAndHotReload demonstrates watching chart files
// and reacting to changes (create, update, delete).
func ExampleFileWatchingAndHotReload() error {
	// Create a temporary directory for charts
	chartDir, err := os.MkdirTemp("", "maelstrom-charts-*")
	if err != nil {
		return fmt.Errorf("failed to create temp dir: %w", err)
	}
	defer os.RemoveAll(chartDir)

	// Write an initial chart file
	chartContent := `
id: example-gateway
version: 1.0.0
spec:
  initial: idle
  states:
    idle:
      type: atomic
      transitions:
        - event: start
          target: running
    running:
      type: atomic
`
	chartPath := chartDir + "/gateway.yaml"
	if err := os.WriteFile(chartPath, []byte(chartContent), 0644); err != nil {
		return fmt.Errorf("failed to write chart: %w", err)
	}

	// Create the registry with default hydrator
	reg, err := chart.NewChartRegistry(chartDir, chart.DefaultHydrator())
	if err != nil {
		return fmt.Errorf("failed to create registry: %w", err)
	}

	// Register an observer that logs all changes
	reg.OnChange(func(key string, def chart.ChartDefinition) {
		log.Printf("[observer] Chart '%s' updated: %s v%s", key, def.ID, def.Version)
	})

	// Start watching in background
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		if err := reg.Start(ctx); err != nil && err != context.Canceled {
			log.Printf("[registry] Error: %v", err)
		}
	}()

	// Wait for initial load
	time.Sleep(200 * time.Millisecond)

	// Get the loaded chart
	def, err := reg.Get("gateway.yaml")
	if err != nil {
		return fmt.Errorf("failed to get chart: %w", err)
	}
	fmt.Printf("Loaded chart: %s v%s\n", def.ID, def.Version)

	// Simulate hot reload by updating the file
	updatedContent := `
id: example-gateway
version: 1.1.0
spec:
  initial: standby
  states:
    standby:
      type: atomic
      transitions:
        - event: activate
          target: active
    active:
      type: atomic
`
	if err := os.WriteFile(chartPath, []byte(updatedContent), 0644); err != nil {
		return fmt.Errorf("failed to update chart: %w", err)
	}

	// Wait for the change to be detected
	time.Sleep(200 * time.Millisecond)

	// Get the updated chart
	updated, err := reg.Get("gateway.yaml")
	if err != nil {
		return fmt.Errorf("failed to get updated chart: %w", err)
	}
	fmt.Printf("Updated chart: %s v%s\n", updated.ID, updated.Version)

	// Verify we can access version history
	oldVersion, err := reg.GetVersion("gateway.yaml", 0)
	if err != nil {
		return fmt.Errorf("failed to get old version: %w", err)
	}
	fmt.Printf("Old version: %s v%s\n", oldVersion.ID, oldVersion.Version)

	return nil
}
```

---

## 2. Registry Operations

This example shows low-level registry operations including Set/Get, version tracking, pre/post load hooks, and CloneUnderLock iteration.

```go
package main

import (
	"fmt"
	"log"

	"github.com/maelstrom/v3/pkg/registry"
)

// ExampleRegistryOperations demonstrates core registry features.
func ExampleRegistryOperations() error {
	// Create a new registry
	reg := registry.New()

	// --- Basic Set/Get Operations ---
	fmt.Println("=== Basic Operations ===")

	// Store values
	reg.Set("service-a", map[string]string{"status": "running", "port": "8080"})
	reg.Set("service-b", map[string]string{"status": "stopped", "port": "8081"})

	// Retrieve current value
	val, err := reg.Get("service-a")
	if err != nil {
		return fmt.Errorf("failed to get service-a: %w", err)
	}
	fmt.Printf("service-a: %v\n", val)

	// --- Version Tracking ---
	fmt.Println("\n=== Version Tracking ===")

	// Update a value multiple times to create versions
	reg.Set("config", "v1")
	reg.Set("config", "v2")
	reg.Set("config", "v3")

	// Access specific versions
	v0, _ := reg.GetVersion("config", 0)
	v1, _ := reg.GetVersion("config", 1)
	v2, _ := reg.GetVersion("config", 2)

	fmt.Printf("config version 0: %v\n", v0)
	fmt.Printf("config version 1: %v\n", v1)
	fmt.Printf("config version 2 (current): %v\n", v2)

	// --- Pre-load Hooks ---
	fmt.Println("\n=== Pre-load Hooks ===")

	// Add a hook that transforms YAML before hydration
	reg.AddPreLoadHook(func(key string, content []byte) ([]byte, error) {
		log.Printf("[pre-load] Processing %s (%d bytes)", key, len(content))
		// Could modify content here (e.g., inject defaults)
		return content, nil
	})

	// --- Post-load Hooks ---
	fmt.Println("\n=== Post-load Hooks ===")

	// Add a hook that validates after hydration
	reg.AddPostLoadHook(func(key string, value interface{}) error {
		log.Printf("[post-load] Validating %s", key)
		// Could perform validation here
		if key == "" {
			return fmt.Errorf("empty key not allowed")
		}
		return nil
	})

	// Use SetWithHooks to apply all hooks
	hydrator := func(content []byte) (interface{}, error) {
		return string(content), nil
	}

	err = reg.SetWithHooks("hooked-entry", []byte("test data"), hydrator)
	if err != nil {
		return fmt.Errorf("failed to set with hooks: %w", err)
	}

	// --- CloneUnderLock Iteration ---
	fmt.Println("\n=== CloneUnderLock Iteration ===")

	// Add more entries
	reg.Set("entry-1", "value-1")
	reg.Set("entry-2", "value-2")
	reg.Set("entry-3", "value-3")

	// Iterate over all entries with read lock held
	reg.CloneUnderLock(func(snapshot map[string]interface{}) {
		fmt.Printf("Registry contains %d entries:\n", len(snapshot))
		for key, value := range snapshot {
			fmt.Printf("  - %s: %v\n", key, value)
		}
	})

	// --- Error Handling ---
	fmt.Println("\n=== Error Handling ===")

	// Try to get a non-existent key
	_, err = reg.Get("non-existent")
	if err == registry.ErrNotFound {
		fmt.Println("Correctly returned ErrNotFound for missing key")
	}

	// Try to get a non-existent version
	_, err = reg.GetVersion("config", 100)
	if err == registry.ErrVersionNotFound {
		fmt.Println("Correctly returned ErrVersionNotFound for invalid version")
	}

	return nil
}
```

---

## 3. Runtime Management

This example demonstrates creating runtimes from registry definitions, sending events, and refreshing runtimes when the registry changes.

```go
package main

import (
	"context"
	"fmt"
	"time"

	"github.com/maelstrom/v3/pkg/registry"
	"github.com/maelstrom/v3/pkg/runtime"
)

// ExampleRuntimeManagement demonstrates creating and managing chart runtimes.
func ExampleRuntimeManagement() error {
	// Create a registry and populate it with chart definitions
	reg := registry.New()

	// Store some chart definitions
	reg.Set("gateway", map[string]interface{}{
		"id":      "api-gateway",
		"version": "1.0.0",
		"states":  []string{"idle", "running"},
	})

	reg.Set("processor", map[string]interface{}{
		"id":      "event-processor",
		"version": "2.0.0",
		"states":  []string{"waiting", "processing", "completed"},
	})

	// Create a factory bound to the registry
	factory := runtime.NewFactory(reg)

	// --- Create a Single Runtime ---
	fmt.Println("=== Creating Single Runtime ===")

	rt, err := factory.Create("gateway")
	if err != nil {
		return fmt.Errorf("failed to create gateway runtime: %w", err)
	}

	fmt.Printf("Created runtime: ID=%s, ChartID=%s\n", rt.ID, rt.Context.ChartID)
	fmt.Printf("Start time: %d\n", rt.Context.StartTime)
	fmt.Printf("Initial active states: %v\n", rt.ActiveStates())

	// --- Create All Runtimes ---
	fmt.Println("\n=== Creating All Runtimes ===")

	runtimes, err := factory.CreateAll()
	if err != nil {
		return fmt.Errorf("failed to create all runtimes: %w", err)
	}

	fmt.Printf("Created %d runtimes:\n", len(runtimes))
	for key, rt := range runtimes {
		fmt.Printf("  - %s: %s\n", key, rt.Context.ChartID)
	}

	// --- Send Events to Runtime ---
	fmt.Println("\n=== Sending Events ===")

	// Create a context for the runtime
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Start the runtime's event loop in background
	go func() {
		if err := rt.Start(ctx); err != nil && err != context.Canceled {
			fmt.Printf("Runtime stopped: %v\n", err)
		}
	}()

	// Send some events
	events := []runtime.Event{
		{Type: "start", Payload: map[string]interface{}{"mode": "normal"}},
		{Type: "configure", Payload: map[string]interface{}{"timeout": 30}},
		{Type: "status_check", CorrelationID: "req-123"},
	}

	for _, evt := range events {
		if err := rt.SendEvent(evt); err != nil {
			return fmt.Errorf("failed to send event %s: %w", evt.Type, err)
		}
		fmt.Printf("Sent event: %s\n", evt.Type)
	}

	// Give events time to process
	time.Sleep(100 * time.Millisecond)

	// Check metrics (updated by event processing)
	fmt.Printf("Runtime metrics: %v\n", rt.Context.Metrics)

	// --- Refresh Runtimes on Registry Change ---
	fmt.Println("\n=== Refreshing Runtimes ===")

	// Add a new chart to the registry (simulating hot reload)
	reg.Set("new-service", map[string]interface{}{
		"id":      "new-service",
		"version": "1.0.0",
		"states":  []string{"init"},
	})

	// Refresh runtimes - adds new ones, removes stale ones
	beforeCount := len(runtimes)
	if err := factory.Refresh(runtimes); err != nil {
		return fmt.Errorf("failed to refresh runtimes: %w", err)
	}
	afterCount := len(runtimes)

	fmt.Printf("Runtime count: %d -> %d\n", beforeCount, afterCount)

	// Verify the new runtime was created
	if newRt, exists := runtimes["new-service"]; exists {
		fmt.Printf("New runtime created: %s\n", newRt.Context.ChartID)
	}

	// Stop all runtimes
	for key, rt := range runtimes {
		if err := rt.Stop(); err != nil {
			fmt.Printf("Error stopping %s: %v\n", key, err)
		}
	}

	return nil
}
```

---

## 4. Bootstrap Extension

This example shows how to hook into the bootstrap sequence, add custom service initialization, and handle the kernel_ready event.

```go
package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/maelstrom/v3/pkg/bootstrap"
)

// ExampleBootstrapExtension demonstrates extending the bootstrap sequence.
func ExampleBootstrapExtension() error {
	// Create a new bootstrap sequence
	seq := bootstrap.NewSequence()

	// Track which states we've entered
	stateHistory := []string{}

	// --- Hook into State Transitions ---
	seq.OnStateEnter(func(state string) error {
		log.Printf("[custom-hook] Entering state: %s", state)
		stateHistory = append(stateHistory, state)

		// Perform custom initialization based on state
		switch state {
		case "security":
			// Custom security service initialization
			fmt.Println("  -> Initializing custom security policies...")
			// Load certificates, configure ACLs, etc.

		case "communication":
			// Custom communication setup
			fmt.Println("  -> Configuring custom message handlers...")
			// Register custom message types

		case "observability":
			// Custom metrics and logging
			fmt.Println("  -> Registering custom metrics collectors...")
			// Set up application-specific metrics

		case "lifecycle":
			// Custom lifecycle management
			fmt.Println("  -> Loading custom chart validators...")
			// Register chart validation rules

		case "handoff":
			// Final preparations before kernel_ready
			fmt.Println("  -> Performing pre-handoff checks...")
			// Verify all services are healthy
		}

		return nil
	})

	// --- Handle Bootstrap Completion ---
	completionSignal := make(chan struct{})
	seq.OnComplete(func() {
		log.Println("[custom-hook] Bootstrap complete!")
		fmt.Println("\n=== Bootstrap Summary ===")
		fmt.Printf("States visited: %v\n", stateHistory)
		close(completionSignal)
	})

	// --- Run the Bootstrap Sequence ---
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Start the sequence in background
	go func() {
		if err := seq.Start(ctx); err != nil {
			log.Printf("[bootstrap] Error: %v", err)
		}
	}()

	// Simulate services becoming ready
	// In real usage, these events would come from actual service initialization
	events := []string{
		"SECURITY_READY",
		"COMMUNICATION_READY",
		"OBSERVABILITY_READY",
		"LIFECYCLE_READY",
		"KERNEL_READY",
	}

	for _, event := range events {
		time.Sleep(100 * time.Millisecond) // Simulate work
		if err := seq.HandleEvent(ctx, event); err != nil {
			return fmt.Errorf("failed to handle event %s: %w", event, err)
		}
	}

	// Wait for completion
	select {
	case <-completionSignal:
		fmt.Println("\nBootstrap completed successfully!")
	case <-ctx.Done():
		return fmt.Errorf("bootstrap timed out")
	}

	// Verify final state
	if seq.IsComplete() {
		fmt.Printf("Final state: %s (complete)\n", seq.CurrentState())
	}

	return nil
}

// ExampleBootstrapWithCustomService shows how to inject a custom service
// into the bootstrap sequence.
func ExampleBootstrapWithCustomService() error {
	seq := bootstrap.NewSequence()

	// Custom service registry
	services := make(map[string]interface{})

	seq.OnStateEnter(func(state string) error {
		ctx := context.Background()

		switch state {
		case "security":
			// Initialize and register custom security service
			services["security"] = &CustomSecurityService{
				policies: []string{"strict-tls", "require-auth"},
			}
			fmt.Println("Custom security service initialized")
			seq.HandleEvent(ctx, "SECURITY_READY")

		case "communication":
			// Initialize with dependency on security service
			if sec, ok := services["security"].(*CustomSecurityService); ok {
				services["communication"] = &CustomCommunicationService{
					security: sec,
				}
				fmt.Println("Custom communication service initialized")
			}
			seq.HandleEvent(ctx, "COMMUNICATION_READY")

		case "observability":
			services["observability"] = &CustomObservabilityService{
				metrics: make(map[string]int),
			}
			fmt.Println("Custom observability service initialized")
			seq.HandleEvent(ctx, "OBSERVABILITY_READY")

		case "lifecycle":
			services["lifecycle"] = &CustomLifecycleService{
				charts: make(map[string]string),
			}
			fmt.Println("Custom lifecycle service initialized")
			seq.HandleEvent(ctx, "LIFECYCLE_READY")

		case "handoff":
			fmt.Println("All custom services ready, signaling kernel_ready")
			seq.HandleEvent(ctx, "KERNEL_READY")
		}

		return nil
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	seq.Start(ctx)

	// Wait for completion
	for !seq.IsComplete() {
		time.Sleep(50 * time.Millisecond)
	}

	fmt.Printf("\nInitialized %d custom services\n", len(services))
	return nil
}

// Custom service types for demonstration
type CustomSecurityService struct {
	policies []string
}

type CustomCommunicationService struct {
	security *CustomSecurityService
}

type CustomObservabilityService struct {
	metrics map[string]int
}

type CustomLifecycleService struct {
	charts map[string]string
}
```

---

## 5. Testing Patterns

This example demonstrates best practices for testing with ManualSource, mocking runtime dependencies, and testing bootstrap state transitions.

```go
package main

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/maelstrom/v3/pkg/bootstrap"
	"github.com/maelstrom/v3/pkg/registry"
	"github.com/maelstrom/v3/pkg/source"
)

// ExampleManualSource demonstrates using ManualSource for controlled testing.
func ExampleManualSource() error {
	// Create a manual source for testing
	manualSrc := source.NewManualSource()

	// Create registry and service
	reg := registry.New()
	svc := registry.NewService(manualSrc, reg)

	// Set up a simple hydrator
	svc.SetHydrator(func(content []byte) (interface{}, error) {
		return string(content), nil
	})

	// Track changes
	var changes []string
	var mu sync.Mutex
	svc.OnChange(func(key string, value interface{}) {
		mu.Lock()
		defer mu.Unlock()
		changes = append(changes, fmt.Sprintf("%s=%v", key, value))
	})

	// Run service in background
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		svc.Run(ctx)
	}()

	// --- Simulate file events manually ---
	fmt.Println("=== Injecting Events ===")

	// Simulate file creation
	manualSrc.Send(source.SourceEvent{
		Key:       "config.yaml",
		Content:   []byte("setting: value1"),
		Type:      source.Created,
		Timestamp: time.Now(),
	})

	// Simulate file update
	manualSrc.Send(source.SourceEvent{
		Key:       "config.yaml",
		Content:   []byte("setting: value2"),
		Type:      source.Updated,
		Timestamp: time.Now(),
	})

	// Simulate another file
	manualSrc.Send(source.SourceEvent{
		Key:       "service.yaml",
		Content:   []byte("name: test-service"),
		Type:      source.Created,
		Timestamp: time.Now(),
	})

	// Give time for processing
	time.Sleep(100 * time.Millisecond)

	// --- Verify results ---
	mu.Lock()
	fmt.Printf("Observed %d changes:\n", len(changes))
	for _, change := range changes {
		fmt.Printf("  - %s\n", change)
	}
	mu.Unlock()

	// Verify registry state
	val, err := reg.Get("config.yaml")
	if err != nil {
		return fmt.Errorf("failed to get config.yaml: %w", err)
	}
	fmt.Printf("Current config.yaml value: %v\n", val)

	// Simulate deletion
	manualSrc.Send(source.SourceEvent{
		Key:       "service.yaml",
		Content:   nil,
		Type:      source.Deleted,
		Timestamp: time.Now(),
	})

	time.Sleep(50 * time.Millisecond)

	// Clean shutdown
	manualSrc.Close(nil)

	return nil
}

// ExampleTestingBootstrapStateTransitions demonstrates testing bootstrap
// state machine behavior.
func ExampleTestingBootstrapStateTransitions() error {
	// Test state transitions
	testCases := []struct {
		name          string
		initialState  string
		event         string
		expectedState string
		shouldError   bool
	}{
		{
			name:          "security to communication",
			initialState:  "security",
			event:         "SECURITY_READY",
			expectedState: "communication",
			shouldError:   false,
		},
		{
			name:         "invalid event in security state",
			initialState: "security",
			event:        "INVALID_EVENT",
			shouldError:  true,
		},
		{
			name:         "event in final state",
			initialState: "complete",
			event:        "SECURITY_READY",
			shouldError:  true,
		},
	}

	for _, tc := range testCases {
		fmt.Printf("\nTest: %s\n", tc.name)

		seq := bootstrap.NewSequence()
		ctx := context.Background()

		// Override the initial state by transitioning
		seq.OnStateEnter(func(state string) error {
			return nil // No-op for testing
		})

		// Start to get to initial state
		go seq.Start(ctx)
		time.Sleep(50 * time.Millisecond)

		// Attempt the transition
		err := seq.HandleEvent(ctx, tc.event)

		if tc.shouldError {
			if err != nil {
				fmt.Printf("  ✓ Got expected error: %v\n", err)
			} else {
				fmt.Printf("  ✗ Expected error but got none\n")
			}
		} else {
			if err != nil {
				fmt.Printf("  ✗ Unexpected error: %v\n", err)
			} else {
				fmt.Printf("  ✓ Transition successful\n")
			}
		}
	}

	return nil
}

// ExampleMockingRuntimeDependencies shows how to mock dependencies for unit tests.
func ExampleMockingRuntimeDependencies() {
	// Mock registry for testing
	mockReg := registry.New()
	mockReg.Set("test-chart", MockChartDefinition{
		ID:       "test-chart",
		Version:  "1.0.0",
		MockData: "test-data",
	})

	// Test using mock
	val, _ := mockReg.Get("test-chart")
	fmt.Printf("Mock chart: %v\n", val)
}

type MockChartDefinition struct {
	ID       string
	Version  string
	MockData string
}

// ExampleTableDrivenRegistryTests demonstrates table-driven testing patterns.
func ExampleTableDrivenRegistryTests() {
	tests := []struct {
		name        string
		operations  func(*registry.Registry) error
		expectError bool
	}{
		{
			name: "set and get",
			operations: func(r *registry.Registry) error {
				r.Set("key", "value")
				val, err := r.Get("key")
				if err != nil {
					return err
				}
				if val != "value" {
					return fmt.Errorf("expected 'value', got %v", val)
				}
				return nil
			},
			expectError: false,
		},
		{
			name: "version history",
			operations: func(r *registry.Registry) error {
				r.Set("key", "v1")
				r.Set("key", "v2")
				v0, _ := r.GetVersion("key", 0)
				v1, _ := r.GetVersion("key", 1)
				if v0 != "v1" || v1 != "v2" {
					return fmt.Errorf("version mismatch")
				}
				return nil
			},
			expectError: false,
		},
		{
			name: "get non-existent",
			operations: func(r *registry.Registry) error {
				_, err := r.Get("missing")
				if err != registry.ErrNotFound {
					return fmt.Errorf("expected ErrNotFound, got %v", err)
				}
				return nil
			},
			expectError: false, // We expect the test to pass (error is expected behavior)
		},
	}

	for _, tt := range tests {
		reg := registry.New()
		err := tt.operations(reg)

		passed := (err == nil) != tt.expectError
		status := "✓"
		if !passed {
			status = "✗"
		}

		if err != nil {
			fmt.Printf("%s %s: %v\n", status, tt.name, err)
		} else {
			fmt.Printf("%s %s\n", status, tt.name)
		}
	}
}

// RunTests is a helper to run all test examples.
func RunTests() error {
	fmt.Println("=== Running ManualSource Example ===")
	if err := ExampleManualSource(); err != nil {
		return fmt.Errorf("ManualSource example failed: %w", err)
	}

	fmt.Println("\n=== Running Bootstrap State Transition Tests ===")
	if err := ExampleTestingBootstrapStateTransitions(); err != nil {
		return fmt.Errorf("Bootstrap tests failed: %w", err)
	}

	fmt.Println("\n=== Running Table-Driven Tests ===")
	ExampleTableDrivenRegistryTests()

	return nil
}

// Helper to simulate testing.T for examples
type mockT struct {
	errors []string
}

func (m *mockT) Errorf(format string, args ...interface{}) {
	m.errors = append(m.errors, fmt.Sprintf(format, args...))
}

func (m *mockT) Logf(format string, args ...interface{}) {
	fmt.Printf(format+"\n", args...)
}
```

---

## Additional Patterns

### Environment Variable Substitution

Charts support environment variable substitution in their YAML:

```yaml
id: ${SERVICE_ID:-default-service}
version: ${VERSION:-1.0.0}
spec:
  port: ${PORT:-8080}
```

```go
// The DefaultHydrator automatically substitutes environment variables
hydrator := chart.DefaultHydrator()
def, err := hydrator([]byte(yamlContent))
// ${VAR} is replaced with os.Getenv("VAR")
// ${VAR:-default} uses "default" if VAR is not set
```

### Concurrent Access Patterns

The registry is safe for concurrent use:

```go
// Multiple goroutines can read
for i := 0; i < 10; i++ {
    go func() {
        val, _ := reg.Get("key")
        _ = val
    }()
}

// Writes are also safe
for i := 0; i < 10; i++ {
    go func(n int) {
        reg.Set("key", n)
    }(i)
}
```

### Integration Example

Putting it all together - a complete integration:

```go
func ExampleCompleteIntegration() error {
    ctx := context.Background()

    // 1. Create chart registry
    chartReg, _ := chart.NewChartRegistry("./charts", chart.DefaultHydrator())

    // 2. Create kernel
    kernel := kernel.New(chartReg)

    // 3. Start everything
    go chartReg.Start(ctx)
    go kernel.Start(ctx)

    // 4. Wait for bootstrap
    for !kernel.IsBootstrapComplete() {
        time.Sleep(100 * time.Millisecond)
    }

    // 5. Get runtimes and send events
    runtimes := kernel.GetRuntimes()
    for _, rt := range runtimes {
        rt.SendEvent(runtime.Event{Type: "initialize"})
    }

    return nil
}
```

---

*For more details, see the [architecture documentation](./arch-v1.md).*
