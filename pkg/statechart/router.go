package statechart

// EventRouter coordinates multiple parallel regions with unified Event I/O.
// Receives external events and routes to appropriate regions.
// Receives region events (sys:*) and handles coordination.
type EventRouter struct {
	regions    map[string]*RegionRuntime
	inputChan  chan Event // External events + region events
	outputChan chan Event // To parent (ChartRuntime)
}

// NewEventRouter creates a router for coordinating parallel regions.
func NewEventRouter(regionDefs map[string]ChartDefinition, actions map[string]ActionFn, guards map[string]GuardFn, appCtx ApplicationContext) *EventRouter {
	er := &EventRouter{
		regions:    make(map[string]*RegionRuntime),
		inputChan:  make(chan Event, 100),
		outputChan: make(chan Event, 100),
	}

	// Create region runtimes for each definition
	for name, def := range regionDefs {
		sm := &StateMachine{
			definition:  def,
			activeState: def.InitialState,
			actions:     actions,
			guards:      guards,
			appCtx:      appCtx,
		}

		// Each region gets its own channels connected to router
		regionInput := make(chan Event, 10)
		regionOutput := er.inputChan // Regions send back to router's input

		region := &RegionRuntime{
			name:         name,
			stateMachine: sm,
			inputChan:    regionInput,
			outputChan:   regionOutput,
			state:        RegionStateRunning,
		}

		er.regions[name] = region
	}

	return er
}

// Run starts the router's event loop.
func (er *EventRouter) Run() {
	// Start all region goroutines
	for _, region := range er.regions {
		go region.Run()
		// Send sys:enter to trigger initial entry actions
		region.inputChan <- Event{Type: SysEnter}
	}

	// Route events
	for ev := range er.inputChan {
		er.routeEvent(ev)
	}
}

// routeEvent directs events to appropriate destinations.
func (er *EventRouter) routeEvent(ev Event) {
	// System events from regions go to parent
	if ev.IsSystem() {
		if er.outputChan != nil {
			er.outputChan <- ev
		}
		return
	}

	// User events: route based on TargetPath
	if ev.TargetPath != "" {
		// Targeted routing: "region:name"
		targetRegion := er.parseTargetRegion(ev.TargetPath)
		if targetRegion != "" {
			if region, exists := er.regions[targetRegion]; exists {
				region.inputChan <- ev
			}
		}
	} else {
		// Broadcast to all regions
		for _, region := range er.regions {
			region.inputChan <- ev
		}
	}
}

// parseTargetRegion extracts region name from TargetPath.
// Format: "region:name" -> returns "name"
func (er *EventRouter) parseTargetRegion(targetPath string) string {
	const prefix = "region:"
	if len(targetPath) > len(prefix) && targetPath[:len(prefix)] == prefix {
		return targetPath[len(prefix):]
	}
	return ""
}

// Send injects an external event into the router.
func (er *EventRouter) Send(ev Event) {
	er.inputChan <- ev
}

// Stop gracefully shuts down all regions.
func (er *EventRouter) Stop() {
	for _, region := range er.regions {
		region.inputChan <- Event{Type: SysExit}
	}
	close(er.inputChan)
}
