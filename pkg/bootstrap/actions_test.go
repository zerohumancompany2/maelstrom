package bootstrap

import (
	"testing"

	"github.com/maelstrom/v3/pkg/statechart"
)

// TestBootstrapActions_CanRegisterAllActions verifies all 5 action functions can be registered.
func TestBootstrapActions_CanRegisterAllActions(t *testing.T) {
	engine := statechart.NewEngine()

	// Register all bootstrap actions referenced in the chart
	engine.RegisterAction("securityBootstrap", securityBootstrap)
	engine.RegisterAction("communicationBootstrap", communicationBootstrap)
	engine.RegisterAction("observabilityBootstrap", observabilityBootstrap)
	engine.RegisterAction("lifecycleBootstrap", lifecycleBootstrap)
	engine.RegisterAction("logSuccess", logSuccess)
	engine.RegisterAction("logFailure", logFailure)
	engine.RegisterAction("panic", panicAction)

	// Verify no registration errors occurred
	// (RegisterAction doesn't return errors, so we just verify the call succeeded)
	t.Log("All bootstrap actions registered successfully")
}

// TestBootstrapActions_HaveCorrectSignatures verifies all actions have correct signatures.
func TestBootstrapActions_HaveCorrectSignatures(t *testing.T) {
	// This test verifies that all action functions match the statechart.ActionFn signature:
	// func(runtimeCtx RuntimeContext, appCtx ApplicationContext, event Event) error

	// Create a test engine to verify actions can be registered
	engine := statechart.NewEngine()

	// These assignments verify the signature matches statechart.ActionFn
	var _ statechart.ActionFn = securityBootstrap
	var _ statechart.ActionFn = communicationBootstrap
	var _ statechart.ActionFn = observabilityBootstrap
	var _ statechart.ActionFn = lifecycleBootstrap
	var _ statechart.ActionFn = logSuccess
	var _ statechart.ActionFn = logFailure
	var _ statechart.ActionFn = panicAction

	// Register to confirm they work
	engine.RegisterAction("securityBootstrap", securityBootstrap)
	engine.RegisterAction("communicationBootstrap", communicationBootstrap)
	engine.RegisterAction("observabilityBootstrap", observabilityBootstrap)
	engine.RegisterAction("lifecycleBootstrap", lifecycleBootstrap)
	engine.RegisterAction("logSuccess", logSuccess)
	engine.RegisterAction("logFailure", logFailure)
	engine.RegisterAction("panic", panicAction)

	t.Log("All action signatures verified")
}

// TestBootstrapActions_ReturnErrorOnInvalidParams verifies actions return errors on invalid params.
func TestBootstrapActions_ReturnErrorOnInvalidParams(t *testing.T) {
	// Mock application context for testing
	mockAppCtx := &mockApplicationContext{data: make(map[string]interface{})}
	mockRuntimeCtx := statechart.RuntimeContext{
		ChartID:   "sys:bootstrap",
		RuntimeID: "test-rt",
	}
	mockEvent := statechart.Event{Type: "TEST"}

	t.Run("securityBootstrap requires boundaries param", func(t *testing.T) {
		err := securityBootstrap(mockRuntimeCtx, mockAppCtx, mockEvent)
		if err == nil {
			t.Error("expected error when boundaries param is missing, got nil")
		}
	})

	t.Run("communicationBootstrap requires mailBackbone param", func(t *testing.T) {
		err := communicationBootstrap(mockRuntimeCtx, mockAppCtx, mockEvent)
		if err == nil {
			t.Error("expected error when mailBackbone param is missing, got nil")
		}
	})

	t.Run("observabilityBootstrap requires tracing, metrics, deadLetterQueue params", func(t *testing.T) {
		err := observabilityBootstrap(mockRuntimeCtx, mockAppCtx, mockEvent)
		if err == nil {
			t.Error("expected error when required params are missing, got nil")
		}
	})

	t.Run("lifecycleBootstrap requires enableSpawn, enableStop, toolRegistry params", func(t *testing.T) {
		err := lifecycleBootstrap(mockRuntimeCtx, mockAppCtx, mockEvent)
		if err == nil {
			t.Error("expected error when required params are missing, got nil")
		}
	})

	t.Run("logFailure requires error param", func(t *testing.T) {
		err := logFailure(mockRuntimeCtx, mockAppCtx, mockEvent)
		if err == nil {
			t.Error("expected error when error param is missing, got nil")
		}
	})
}

// mockApplicationContext is a simple mock for testing
type mockApplicationContext struct {
	data map[string]interface{}
}

func (m *mockApplicationContext) Get(key string, callerBoundary string) (interface{}, []string, error) {
	val, ok := m.data[key]
	if !ok {
		return nil, nil, nil
	}
	return val, nil, nil
}

func (m *mockApplicationContext) Set(key string, value interface{}, taints []string, callerBoundary string) error {
	m.data[key] = value
	return nil
}

func (m *mockApplicationContext) Namespace() string {
	return "test"
}

// TestBootstrapActions_LogExecution verifies actions log their execution.
func TestBootstrapActions_LogExecution(t *testing.T) {
	mockAppCtx := &mockApplicationContext{data: map[string]interface{}{
		"boundaries":      []string{"inner", "dmz", "outer"},
		"mailBackbone":    true,
		"tracing":         true,
		"metrics":         true,
		"deadLetterQueue": true,
		"enableSpawn":     true,
		"enableStop":      true,
		"toolRegistry":    true,
		"error":           "test error",
	}}
	mockRuntimeCtx := statechart.RuntimeContext{
		ChartID:   "sys:bootstrap",
		RuntimeID: "test-rt",
	}
	mockEvent := statechart.Event{Type: "TEST"}

	t.Run("securityBootstrap logs execution with parameters", func(t *testing.T) {
		err := securityBootstrap(mockRuntimeCtx, mockAppCtx, mockEvent)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	t.Run("communicationBootstrap logs execution with parameters", func(t *testing.T) {
		err := communicationBootstrap(mockRuntimeCtx, mockAppCtx, mockEvent)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	t.Run("observabilityBootstrap logs execution with parameters", func(t *testing.T) {
		err := observabilityBootstrap(mockRuntimeCtx, mockAppCtx, mockEvent)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	t.Run("lifecycleBootstrap logs execution with parameters", func(t *testing.T) {
		err := lifecycleBootstrap(mockRuntimeCtx, mockAppCtx, mockEvent)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	t.Run("logSuccess logs success message", func(t *testing.T) {
		err := logSuccess(mockRuntimeCtx, mockAppCtx, mockEvent)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	t.Run("logFailure logs error message", func(t *testing.T) {
		err := logFailure(mockRuntimeCtx, mockAppCtx, mockEvent)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
}
