package bootstrap

import (
	"log"

	"github.com/maelstrom/v3/pkg/statechart"
)

// securityBootstrap is the entry action for the security state.
func securityBootstrap(runtimeCtx statechart.RuntimeContext, appCtx statechart.ApplicationContext, event statechart.Event) error {
	log.Println("[bootstrap] securityBootstrap executed")
	return nil
}

// communicationBootstrap is the entry action for the communication state.
func communicationBootstrap(runtimeCtx statechart.RuntimeContext, appCtx statechart.ApplicationContext, event statechart.Event) error {
	log.Println("[bootstrap] communicationBootstrap executed")
	return nil
}

// observabilityBootstrap is the entry action for the observability state.
func observabilityBootstrap(runtimeCtx statechart.RuntimeContext, appCtx statechart.ApplicationContext, event statechart.Event) error {
	log.Println("[bootstrap] observabilityBootstrap executed")
	return nil
}

// lifecycleBootstrap is the entry action for the lifecycle state.
func lifecycleBootstrap(runtimeCtx statechart.RuntimeContext, appCtx statechart.ApplicationContext, event statechart.Event) error {
	log.Println("[bootstrap] lifecycleBootstrap executed")
	return nil
}

// logSuccess logs successful bootstrap completion.
func logSuccess(runtimeCtx statechart.RuntimeContext, appCtx statechart.ApplicationContext, event statechart.Event) error {
	log.Println("[bootstrap] logSuccess executed - bootstrap completed successfully")
	return nil
}

// logFailure logs bootstrap failure.
func logFailure(runtimeCtx statechart.RuntimeContext, appCtx statechart.ApplicationContext, event statechart.Event) error {
	log.Println("[bootstrap] logFailure executed - bootstrap failed")
	return nil
}

// panicAction panics on bootstrap failure.
func panicAction(runtimeCtx statechart.RuntimeContext, appCtx statechart.ApplicationContext, event statechart.Event) error {
	log.Println("[bootstrap] panicAction executed - panicking due to bootstrap failure")
	panic("bootstrap failed")
}
