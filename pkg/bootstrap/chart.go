package bootstrap

import (
	"fmt"

	"github.com/maelstrom/v3/pkg/chart"
)

// BootstrapChartYAML is the hard-coded bootstrap chart that loads 4 core services.
// This is compiled into the binary and executed first on startup.
const BootstrapChartYAML = `
id: sys:bootstrap
version: 1.0.0
spec:
  initial: initializing
  states:
    initializing:
      type: atomic
      transitions:
        - event: START_BOOTSTRAP
          target: security

    security:
      type: atomic
      entryActions:
        - loadSecurityService
      transitions:
        - event: SECURITY_READY
          target: communication

    communication:
      type: atomic
      entryActions:
        - loadCommunicationService
      transitions:
        - event: COMMUNICATION_READY
          target: observability

    observability:
      type: atomic
      entryActions:
        - loadObservabilityService
      transitions:
        - event: OBSERVABILITY_READY
          target: lifecycle

    lifecycle:
      type: atomic
      entryActions:
        - loadLifecycleService
      transitions:
        - event: LIFECYCLE_READY
          target: handoff

    handoff:
      type: atomic
      entryActions:
        - signalKernelReady
      transitions:
        - event: KERNEL_READY
          target: complete

    complete:
      type: final
`

// CoreServices defines the 4 core services loaded by bootstrap.
var CoreServices = []struct {
	ID          string
	Name        string
	Description string
}{
	{ID: "sys:security", Name: "Security Service", Description: "Inner/DMZ/outer boundaries, data tainting"},
	{ID: "sys:communication", Name: "Communication Service", Description: "Mail system, cross-region routing"},
	{ID: "sys:observability", Name: "Observability Service", Description: "Metrics, logging, health checks"},
	{ID: "sys:lifecycle", Name: "Lifecycle Service", Description: "Chart management, hot-reload"},
}

// LoadBootstrapChart parses the hard-coded YAML into a ChartDefinition.
func LoadBootstrapChart() (chart.ChartDefinition, error) {
	hydrator := chart.DefaultHydrator()
	def, err := hydrator([]byte(BootstrapChartYAML))
	if err != nil {
		return chart.ChartDefinition{}, fmt.Errorf("failed to hydrate bootstrap chart: %w", err)
	}
	return def, nil
}
