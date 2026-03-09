package bootstrap

import (
	"errors"
	"fmt"
	"log"

	"github.com/maelstrom/v3/pkg/services/communication"
	"github.com/maelstrom/v3/pkg/services/lifecycle"
	"github.com/maelstrom/v3/pkg/services/observability"
	"github.com/maelstrom/v3/pkg/services/security"
	"github.com/maelstrom/v3/pkg/statechart"
)

var ErrNotImplemented = errors.New("not implemented")

const (
	ActionLoadSecurityService      = "loadSecurityService"
	ActionLoadCommunicationService = "loadCommunicationService"
	ActionLoadObservabilityService = "loadObservabilityService"
	ActionLoadLifecycleService     = "loadLifecycleService"
	ActionSignalKernelReady        = "signalKernelReady"
)

func getEngine(appCtx statechart.ApplicationContext, chartID string) (statechart.Library, error) {
	engineAny, _, err := appCtx.Get("__engine", chartID)
	if err != nil {
		return nil, fmt.Errorf("failed to get engine: %w", err)
	}
	if engineAny == nil {
		return nil, fmt.Errorf("engine not found in appCtx")
	}
	engine, ok := engineAny.(statechart.Library)
	if !ok {
		return nil, fmt.Errorf("invalid engine type in appCtx")
	}
	return engine, nil
}

// securityBootstrap is the entry action for the security state.
func securityBootstrap(runtimeCtx statechart.RuntimeContext, appCtx statechart.ApplicationContext, event statechart.Event) error {
	boundaries, _, err := appCtx.Get("boundaries", "bootstrap")
	if err != nil {
		return fmt.Errorf("failed to get boundaries param: %w", err)
	}
	if boundaries == nil {
		return fmt.Errorf("boundaries parameter is required")
	}
	log.Printf("[bootstrap] securityBootstrap executed with boundaries: %v", boundaries)
	return nil
}

// communicationBootstrap is the entry action for the communication state.
func communicationBootstrap(runtimeCtx statechart.RuntimeContext, appCtx statechart.ApplicationContext, event statechart.Event) error {
	mailBackbone, _, err := appCtx.Get("mailBackbone", "bootstrap")
	if err != nil {
		return fmt.Errorf("failed to get mailBackbone param: %w", err)
	}
	if mailBackbone == nil {
		return fmt.Errorf("mailBackbone parameter is required")
	}
	log.Printf("[bootstrap] communicationBootstrap executed with mailBackbone: %v", mailBackbone)
	return nil
}

// observabilityBootstrap is the entry action for the observability state.
func observabilityBootstrap(runtimeCtx statechart.RuntimeContext, appCtx statechart.ApplicationContext, event statechart.Event) error {
	tracing, _, err := appCtx.Get("tracing", "bootstrap")
	if err != nil {
		return fmt.Errorf("failed to get tracing param: %w", err)
	}
	if tracing == nil {
		return fmt.Errorf("tracing parameter is required")
	}
	metrics, _, err := appCtx.Get("metrics", "bootstrap")
	if err != nil {
		return fmt.Errorf("failed to get metrics param: %w", err)
	}
	if metrics == nil {
		return fmt.Errorf("metrics parameter is required")
	}
	deadLetterQueue, _, err := appCtx.Get("deadLetterQueue", "bootstrap")
	if err != nil {
		return fmt.Errorf("failed to get deadLetterQueue param: %w", err)
	}
	if deadLetterQueue == nil {
		return fmt.Errorf("deadLetterQueue parameter is required")
	}
	log.Printf("[bootstrap] observabilityBootstrap executed with tracing: %v, metrics: %v, deadLetterQueue: %v", tracing, metrics, deadLetterQueue)
	return nil
}

// lifecycleBootstrap is the entry action for the lifecycle state.
func lifecycleBootstrap(runtimeCtx statechart.RuntimeContext, appCtx statechart.ApplicationContext, event statechart.Event) error {
	enableSpawn, _, err := appCtx.Get("enableSpawn", "bootstrap")
	if err != nil {
		return fmt.Errorf("failed to get enableSpawn param: %w", err)
	}
	if enableSpawn == nil {
		return fmt.Errorf("enableSpawn parameter is required")
	}
	enableStop, _, err := appCtx.Get("enableStop", "bootstrap")
	if err != nil {
		return fmt.Errorf("failed to get enableStop param: %w", err)
	}
	if enableStop == nil {
		return fmt.Errorf("enableStop parameter is required")
	}
	toolRegistry, _, err := appCtx.Get("toolRegistry", "bootstrap")
	if err != nil {
		return fmt.Errorf("failed to get toolRegistry param: %w", err)
	}
	if toolRegistry == nil {
		return fmt.Errorf("toolRegistry parameter is required")
	}
	log.Printf("[bootstrap] lifecycleBootstrap executed with enableSpawn: %v, enableStop: %v, toolRegistry: %v", enableSpawn, enableStop, toolRegistry)
	return nil
}

// logSuccess logs successful bootstrap completion.
func logSuccess(runtimeCtx statechart.RuntimeContext, appCtx statechart.ApplicationContext, event statechart.Event) error {
	log.Println("[bootstrap] logSuccess executed - bootstrap completed successfully")
	return nil
}

// logFailure logs bootstrap failure.
func logFailure(runtimeCtx statechart.RuntimeContext, appCtx statechart.ApplicationContext, event statechart.Event) error {
	errMsg, _, err := appCtx.Get("error", "bootstrap")
	if err != nil {
		return fmt.Errorf("failed to get error param: %w", err)
	}
	if errMsg == nil {
		return fmt.Errorf("error parameter is required")
	}
	log.Printf("[bootstrap] logFailure executed - bootstrap failed: %v", errMsg)
	return nil
}

// panicAction panics on bootstrap failure.
func panicAction(runtimeCtx statechart.RuntimeContext, appCtx statechart.ApplicationContext, event statechart.Event) error {
	log.Println("[bootstrap] panicAction executed - panicking due to bootstrap failure")
	panic("bootstrap failed")
}

// LoadSecurityService is the entry action for the security state.
func LoadSecurityService(runtimeCtx statechart.RuntimeContext, appCtx statechart.ApplicationContext, event statechart.Event) error {
	engine, err := getEngine(appCtx, runtimeCtx.ChartID)
	if err != nil {
		return err
	}

	// Load security chart definition
	def := security.BootstrapChart()

	// Spawn security runtime
	securityRTID, err := engine.Spawn(def, appCtx)
	if err != nil {
		return fmt.Errorf("failed to spawn security runtime: %w", err)
	}
	log.Printf("[bootstrap] Spawned security runtime: %s", securityRTID)

	// Start the runtime
	if err := engine.Control(securityRTID, statechart.CmdStart); err != nil {
		return fmt.Errorf("failed to start security runtime: %w", err)
	}
	log.Printf("[bootstrap] Started security runtime: %s", securityRTID)

	// Store security RTID in appCtx
	if err := appCtx.Set("bootstrap:security:runtimeID", string(securityRTID), nil, runtimeCtx.ChartID); err != nil {
		return fmt.Errorf("failed to store security RTID: %w", err)
	}

	// Dispatch SECURITY_READY event to bootstrap parent
	if err := engine.Dispatch(statechart.RuntimeID(runtimeCtx.RuntimeID), statechart.Event{Type: "SECURITY_READY"}); err != nil {
		return fmt.Errorf("failed to dispatch SECURITY_READY: %w", err)
	}
	log.Printf("[bootstrap] Dispatched SECURITY_READY event")

	return nil
}

// LoadCommunicationService is the entry action for the communication state.
func LoadCommunicationService(runtimeCtx statechart.RuntimeContext, appCtx statechart.ApplicationContext, event statechart.Event) error {
	engine, err := getEngine(appCtx, runtimeCtx.ChartID)
	if err != nil {
		return err
	}

	// Load communication chart definition
	def := communication.BootstrapChart()

	// Spawn communication runtime
	commRTID, err := engine.Spawn(def, appCtx)
	if err != nil {
		return fmt.Errorf("failed to spawn communication runtime: %w", err)
	}
	log.Printf("[bootstrap] Spawned communication runtime: %s", commRTID)

	// Start the runtime
	if err := engine.Control(commRTID, statechart.CmdStart); err != nil {
		return fmt.Errorf("failed to start communication runtime: %w", err)
	}
	log.Printf("[bootstrap] Started communication runtime: %s", commRTID)

	// Store communication RTID in appCtx
	if err := appCtx.Set("bootstrap:communication:runtimeID", string(commRTID), nil, runtimeCtx.ChartID); err != nil {
		return fmt.Errorf("failed to store communication RTID: %w", err)
	}

	// Dispatch COMMUNICATION_READY event to bootstrap parent
	if err := engine.Dispatch(statechart.RuntimeID(runtimeCtx.RuntimeID), statechart.Event{Type: "COMMUNICATION_READY"}); err != nil {
		return fmt.Errorf("failed to dispatch COMMUNICATION_READY: %w", err)
	}
	log.Printf("[bootstrap] Dispatched COMMUNICATION_READY event")

	return nil
}

// LoadObservabilityService is the entry action for the observability state.
func LoadObservabilityService(runtimeCtx statechart.RuntimeContext, appCtx statechart.ApplicationContext, event statechart.Event) error {
	engine, err := getEngine(appCtx, runtimeCtx.ChartID)
	if err != nil {
		return err
	}

	// Load observability chart definition
	def := observability.BootstrapChart()

	// Spawn observability runtime
	obsRTID, err := engine.Spawn(def, appCtx)
	if err != nil {
		return fmt.Errorf("failed to spawn observability runtime: %w", err)
	}
	log.Printf("[bootstrap] Spawned observability runtime: %s", obsRTID)

	// Start the runtime
	if err := engine.Control(obsRTID, statechart.CmdStart); err != nil {
		return fmt.Errorf("failed to start observability runtime: %w", err)
	}
	log.Printf("[bootstrap] Started observability runtime: %s", obsRTID)

	// Store observability RTID in appCtx
	if err := appCtx.Set("bootstrap:observability:runtimeID", string(obsRTID), nil, runtimeCtx.ChartID); err != nil {
		return fmt.Errorf("failed to store observability RTID: %w", err)
	}

	// Dispatch OBSERVABILITY_READY event to bootstrap parent
	if err := engine.Dispatch(statechart.RuntimeID(runtimeCtx.RuntimeID), statechart.Event{Type: "OBSERVABILITY_READY"}); err != nil {
		return fmt.Errorf("failed to dispatch OBSERVABILITY_READY: %w", err)
	}
	log.Printf("[bootstrap] Dispatched OBSERVABILITY_READY event")

	return nil
}

// LoadLifecycleService is the entry action for the lifecycle state.
func LoadLifecycleService(runtimeCtx statechart.RuntimeContext, appCtx statechart.ApplicationContext, event statechart.Event) error {
	engine, err := getEngine(appCtx, runtimeCtx.ChartID)
	if err != nil {
		return err
	}

	// Load lifecycle chart definition
	def := lifecycle.BootstrapChart()

	// Spawn lifecycle runtime
	lifecycleRTID, err := engine.Spawn(def, appCtx)
	if err != nil {
		return fmt.Errorf("failed to spawn lifecycle runtime: %w", err)
	}
	log.Printf("[bootstrap] Spawned lifecycle runtime: %s", lifecycleRTID)

	// Start the runtime
	if err := engine.Control(lifecycleRTID, statechart.CmdStart); err != nil {
		return fmt.Errorf("failed to start lifecycle runtime: %w", err)
	}
	log.Printf("[bootstrap] Started lifecycle runtime: %s", lifecycleRTID)

	// Store lifecycle RTID in appCtx
	if err := appCtx.Set("bootstrap:lifecycle:runtimeID", string(lifecycleRTID), nil, runtimeCtx.ChartID); err != nil {
		return fmt.Errorf("failed to store lifecycle RTID: %w", err)
	}

	// Dispatch LIFECYCLE_READY event to bootstrap parent
	if err := engine.Dispatch(statechart.RuntimeID(runtimeCtx.RuntimeID), statechart.Event{Type: "LIFECYCLE_READY"}); err != nil {
		return fmt.Errorf("failed to dispatch LIFECYCLE_READY: %w", err)
	}
	log.Printf("[bootstrap] Dispatched LIFECYCLE_READY event")

	return nil
}

// SignalKernelReady is the entry action for the ready state.
func SignalKernelReady(runtimeCtx statechart.RuntimeContext, appCtx statechart.ApplicationContext, event statechart.Event) error {
	engine, err := getEngine(appCtx, runtimeCtx.ChartID)
	if err != nil {
		return err
	}

	// Read all service runtime IDs from appCtx
	var loadedServices []string

	securityRTID, _, _ := appCtx.Get("bootstrap:security:runtimeID", runtimeCtx.ChartID)
	if securityRTID != "" && securityRTID != nil {
		loadedServices = append(loadedServices, "sys:security")
	}

	commRTID, _, _ := appCtx.Get("bootstrap:communication:runtimeID", runtimeCtx.ChartID)
	if commRTID != "" && commRTID != nil {
		loadedServices = append(loadedServices, "sys:communication")
	}

	obsRTID, _, _ := appCtx.Get("bootstrap:observability:runtimeID", runtimeCtx.ChartID)
	if obsRTID != "" && obsRTID != nil {
		loadedServices = append(loadedServices, "sys:observability")
	}

	lifecycleRTID, _, _ := appCtx.Get("bootstrap:lifecycle:runtimeID", runtimeCtx.ChartID)
	if lifecycleRTID != "" && lifecycleRTID != nil {
		loadedServices = append(loadedServices, "sys:lifecycle")
	}

	// Store loaded services list in appCtx
	if err := appCtx.Set("bootstrap:loaded:services", loadedServices, nil, runtimeCtx.ChartID); err != nil {
		return fmt.Errorf("failed to store loaded services: %w", err)
	}
	log.Printf("[bootstrap] Stored loaded services: %v", loadedServices)

	// Dispatch KERNEL_READY event to bootstrap parent
	if err := engine.Dispatch(statechart.RuntimeID(runtimeCtx.RuntimeID), statechart.Event{Type: "KERNEL_READY"}); err != nil {
		return fmt.Errorf("failed to dispatch KERNEL_READY: %w", err)
	}
	log.Printf("[bootstrap] Dispatched KERNEL_READY event")

	return nil
}

// Aliases for backward compatibility with tests
var (
	loadSecurityService      = LoadSecurityService
	loadCommunicationService = LoadCommunicationService
	loadObservabilityService = LoadObservabilityService
	loadLifecycleService     = LoadLifecycleService
	signalKernelReady        = SignalKernelReady
)
