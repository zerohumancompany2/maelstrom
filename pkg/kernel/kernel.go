package kernel

import (
	"context"
	"fmt"
	"log"
	"sync"

	"github.com/maelstrom/v3/pkg/bootstrap"
	"github.com/maelstrom/v3/pkg/runtime"
)

// Kernel orchestrates bootstrap and hands off to ChartRegistry.
type Kernel struct {
	factory  *runtime.Factory
	sequence *bootstrap.Sequence
	runtimes map[string]*runtime.ChartRuntime
	mu       sync.RWMutex
}

// New creates a new Kernel.
func New() *Kernel {
	return &Kernel{
		runtimes: make(map[string]*runtime.ChartRuntime),
	}
}

// Start begins the bootstrap sequence and transitions to runtime.
func (k *Kernel) Start(ctx context.Context) error {
	log.Println("[kernel] Starting kernel")

	// Load bootstrap chart definition
	def, err := bootstrap.LoadBootstrapChart()
	if err != nil {
		return fmt.Errorf("failed to load bootstrap chart: %w", err)
	}

	log.Printf("[kernel] Loaded bootstrap chart: %s v%s", def.ID, def.Version)

	// Create bootstrap sequence
	seq := bootstrap.NewSequence()
	k.mu.Lock()
	k.sequence = seq
	k.mu.Unlock()

	// Set up state entry handlers
	seq.OnStateEnter(func(state string) error {
		return k.onBootstrapStateEnter(ctx, state)
	})

	// Set up completion handler
	seq.OnComplete(func() {
		log.Println("[kernel] Bootstrap complete, handing off to ChartRegistry")
		k.onBootstrapComplete()
	})

	// Start the sequence
	if err := seq.Start(ctx); err != nil {
		return fmt.Errorf("bootstrap failed: %w", err)
	}

	// Wait for completion or cancellation
	<-ctx.Done()
	return ctx.Err()
}

func (k *Kernel) onBootstrapStateEnter(ctx context.Context, state string) error {
	log.Printf("[kernel] Bootstrap state: %s", state)

	k.mu.RLock()
	seq := k.sequence
	k.mu.RUnlock()

	switch state {
	case "security":
		// In real implementation, instantiate sys:security runtime
		// For now, just log and transition
		log.Println("[kernel] Loading sys:security service")
		// Simulate service ready
		go func() {
			seq.HandleEvent(ctx, "SECURITY_READY")
		}()

	case "communication":
		log.Println("[kernel] Loading sys:communication service")
		go func() {
			seq.HandleEvent(ctx, "COMMUNICATION_READY")
		}()

	case "observability":
		log.Println("[kernel] Loading sys:observability service")
		go func() {
			seq.HandleEvent(ctx, "OBSERVABILITY_READY")
		}()

	case "lifecycle":
		log.Println("[kernel] Loading sys:lifecycle service")
		go func() {
			seq.HandleEvent(ctx, "LIFECYCLE_READY")
		}()

	case "handoff":
		log.Println("[kernel] Signaling kernel_ready")
		go func() {
			seq.HandleEvent(ctx, "KERNEL_READY")
		}()
	}

	return nil
}

func (k *Kernel) onBootstrapComplete() {
	log.Println("[kernel] Kernel going dormant")
}

// IsBootstrapComplete returns true if bootstrap has finished.
func (k *Kernel) IsBootstrapComplete() bool {
	k.mu.RLock()
	seq := k.sequence
	k.mu.RUnlock()
	if seq == nil {
		return false
	}
	return seq.IsComplete()
}

// GetRuntimes returns the currently active runtimes.
func (k *Kernel) GetRuntimes() map[string]*runtime.ChartRuntime {
	k.mu.RLock()
	defer k.mu.RUnlock()
	return k.runtimes
}
