package kernel

import (
	"log"

	"github.com/maelstrom/v3/pkg/statechart"
)

func securityBootstrap(rt statechart.RuntimeContext, appCtx statechart.ApplicationContext, ev statechart.Event) error {
	log.Println("[bootstrap] Security service initialized")
	return nil
}

func communicationBootstrap(rt statechart.RuntimeContext, appCtx statechart.ApplicationContext, ev statechart.Event) error {
	log.Println("[bootstrap] Communication service initialized")
	return nil
}

func observabilityBootstrap(rt statechart.RuntimeContext, appCtx statechart.ApplicationContext, ev statechart.Event) error {
	log.Println("[bootstrap] Observability service initialized")
	return nil
}

func lifecycleBootstrap(rt statechart.RuntimeContext, appCtx statechart.ApplicationContext, ev statechart.Event) error {
	log.Println("[bootstrap] Lifecycle service initialized")
	return nil
}
