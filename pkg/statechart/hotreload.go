package statechart

import (
	"time"
)

// HotReloadOptions provides configuration for hot-reload operations.
type HotReloadOptions struct {
	Timeout          time.Duration
	MaxAttempts      int
	HistoryMode      HistoryMode
	ContextTransform *ContextTransform
}

// HistoryMode specifies how to restore state during hot-reload.
type HistoryMode int

const (
	HistoryModeNone HistoryMode = iota
	HistoryModeShallow
	HistoryModeDeep
)

// ContextTransform provides template-based context migration.
type ContextTransform struct {
	Template string
}

// HotReload performs a hot-reload of the chart definition for the given runtime.
func (e *Engine) HotReload(id RuntimeID, newDef ChartDefinition, opts HotReloadOptions) error {
	// TODO: Implement hot-reload protocol (arch-v1.md L865-880)
	// 1. Signal prepareForReload
	// 2. Wait for quiescence (with timeout)
	// 3. If quiescent: stop, spawn with history
	// 4. If timeout: force-stop, cleanStart
	// 5. If maxAttempts exceeded: log failure, require admin intervention
	return nil
}
