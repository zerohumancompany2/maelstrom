package bootstrap

import (
	"fmt"

	"github.com/maelstrom/v3/pkg/statechart"
)

type ChartDefinition = statechart.ChartDefinition

func DefaultHydrator() statechart.HydratorFunc {
	return statechart.DefaultHydrator()
}

// BootstrapChartYAML is the hard-coded bootstrap chart that loads 4 core services.
// This is compiled into the binary and executed first on startup.
const BootstrapChartYAML = `
id: sys:bootstrap
version: 1.0.0
spec:
  initial: sys:bootstrap/init
  states:
    sys:bootstrap/init:
      type: atomic
      transitions:
        - event: START_BOOTSTRAP
          target: sys:bootstrap/security

    sys:bootstrap/security:
      type: atomic
      entryActions:
        - loadSecurityService
      transitions:
        - event: SECURITY_READY
          target: sys:bootstrap/communication
        - event: securityFailed
          target: sys:bootstrap/failed

    sys:bootstrap/communication:
      type: atomic
      entryActions:
        - loadCommunicationService
      transitions:
        - event: COMMUNICATION_READY
          target: sys:bootstrap/observability
        - event: communicationFailed
          target: sys:bootstrap/failed

    sys:bootstrap/observability:
      type: atomic
      entryActions:
        - loadObservabilityService
      transitions:
        - event: OBSERVABILITY_READY
          target: sys:bootstrap/lifecycle
        - event: observabilityFailed
          target: sys:bootstrap/failed

    sys:bootstrap/lifecycle:
      type: atomic
      entryActions:
        - loadLifecycleService
      transitions:
        - event: LIFECYCLE_READY
          target: sys:bootstrap/ready
        - event: lifecycleFailed
          target: sys:bootstrap/failed

    sys:bootstrap/ready:
      type: final
      entryActions:
        - signalKernelReady
        - logSuccess

    sys:bootstrap/failed:
      type: final
      entryActions:
        - logFailure
        - panic
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
func LoadBootstrapChart() (statechart.ChartDefinition, error) {
	hydrator := statechart.DefaultHydrator()
	def, err := hydrator([]byte(BootstrapChartYAML))
	if err != nil {
		return statechart.ChartDefinition{}, fmt.Errorf("failed to hydrate bootstrap chart: %w", err)
	}
	return def, nil
}
