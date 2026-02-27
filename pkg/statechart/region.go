package statechart

import (
	"sync"
)

// RegionState represents the lifecycle state of a RegionRuntime.
type RegionState int

const (
	RegionStateRunning RegionState = iota
	RegionStatePaused
	RegionStateExiting
	RegionStateDone
)

// RegionRuntime represents one parallel region executing in isolation.
type RegionRuntime struct {
	name         string
	stateMachine *StateMachine

	// Unified event channels (both chan Event for symmetry)
	inputChan  chan Event // Receives from parent router
	outputChan chan Event // Sends to parent router

	state RegionState
	mu    sync.Mutex
}

// Run executes the region's event loop in its own goroutine.
func (rr *RegionRuntime) Run() {
	for {
		rr.mu.Lock()
		state := rr.state
		rr.mu.Unlock()

		if state == RegionStateDone {
			return
		}

		select {
		case ev := <-rr.inputChan:
			rr.mu.Lock()
			state := rr.state
			rr.mu.Unlock()

			if state == RegionStatePaused && ev.Type != SysResume {
				// Skip events while paused (except resume)
				continue
			}

			if state == RegionStateDone {
				// Already done, ignore events
				return
			}

			rr.handleEvent(ev)
		}
	}
}

// handleEvent processes a single event.
func (rr *RegionRuntime) handleEvent(ev Event) {
	if ev.IsSystem() {
		rr.handleSystemEvent(ev)
		return
	}

	// User event: process through state machine
	result := rr.stateMachine.ProcessEvent(ev)

	// Report transition to parent
	if result.Transitioned {
		rr.outputChan <- Event{
			Type:   SysTransition,
			Source: "region:" + rr.name,
			Payload: TransitionPayload{
				From: result.FromState,
				To:   result.ToState,
			},
		}
	}

	// Report completion if final state reached
	if result.IsFinalState {
		rr.mu.Lock()
		rr.state = RegionStateDone
		rr.mu.Unlock()

		rr.outputChan <- Event{
			Type:   SysDone,
			Source: "region:" + rr.name,
		}
	}

	// Route emitted event through parent
	if result.EmitEvent != nil {
		rr.outputChan <- Event{
			Type:       result.EmitEvent.Type,
			Payload:    result.EmitEvent.Payload,
			Source:     "region:" + rr.name,
			TargetPath: result.EmitEvent.TargetPath,
		}
	}
}

// TransitionPayload carries state change information.
type TransitionPayload struct {
	From string
	To   string
}

// handleSystemEvent handles sys:* events.
func (rr *RegionRuntime) handleSystemEvent(ev Event) {
	switch ev.Type {
	case SysEnter:
		// Execute entry actions for initial state
		rr.stateMachine.executeEntryActions(rr.stateMachine.activeState, ev)

	case SysExit:
		rr.mu.Lock()
		rr.state = RegionStateExiting
		rr.mu.Unlock()
		// In real implementation, this would trigger processing toward final state

	case SysPause:
		rr.mu.Lock()
		rr.state = RegionStatePaused
		rr.mu.Unlock()

	case SysResume:
		rr.mu.Lock()
		rr.state = RegionStateRunning
		rr.mu.Unlock()
	}
}
