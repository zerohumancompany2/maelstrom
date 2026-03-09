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
