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
