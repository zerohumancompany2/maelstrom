package kernel

import (
	"context"
	"fmt"
	"log"
	"sync"

	"github.com/maelstrom/v3/pkg/bootstrap"
	"github.com/maelstrom/v3/pkg/runtime"
	"github.com/maelstrom/v3/pkg/statechart"
)

// KernelConfig holds kernel configuration.
type KernelConfig struct {
	ChartsDir string            // Path to charts/ directory
	AppVars   map[string]string // Application variables for hydration
}

// Kernel orchestrates bootstrap and hands off to ChartRegistry.
type Kernel struct {
	engine    statechart.Library
	config    KernelConfig
	factory   *runtime.Factory
	sequence  *bootstrap.Sequence
	services  map[string]statechart.RuntimeID
	runtimes  map[string]*runtime.ChartRuntime
	mu        sync.RWMutex
	readyChan chan struct{}
}

// New creates a new Kernel.
func New() *Kernel {
	return &Kernel{
		services: make(map[string]statechart.RuntimeID),
		runtimes: make(map[string]*runtime.ChartRuntime),
	}
}

// NewWithEngine creates a new Kernel with the given statechart engine.
func NewWithEngine(engine statechart.Library) *Kernel {
	return &Kernel{
		engine:   engine,
		services: make(map[string]statechart.RuntimeID),
		runtimes: make(map[string]*runtime.ChartRuntime),
	}
}

// WithConfig sets the kernel configuration.
func (k *Kernel) WithConfig(cfg KernelConfig) *Kernel {
	k.config = cfg
	return k
}

// RegisterBootstrapActions registers the four bootstrap actions.
func (k *Kernel) RegisterBootstrapActions() {
	if k.engine == nil {
		return
	}
	k.engine.RegisterAction("securityBootstrap", securityBootstrap)
	k.engine.RegisterAction("communicationBootstrap", communicationBootstrap)
	k.engine.RegisterAction("observabilityBootstrap", observabilityBootstrap)
	k.engine.RegisterAction("lifecycleBootstrap", lifecycleBootstrap)
}

// Start begins the bootstrap sequence and transitions to runtime.
func (k *Kernel) Start(ctx context.Context) error {
	log.Println("[kernel] Starting kernel")

	// Register bootstrap actions before spawning
	k.RegisterBootstrapActions()

	// Load bootstrap chart definition
	def, err := bootstrap.LoadBootstrapChart()
	if err != nil {
		return fmt.Errorf("failed to load bootstrap chart: %w", err)
	}

	log.Printf("[kernel] Loaded bootstrap chart: %s v%s", def.ID, def.Version)

	// Spawn bootstrap runtime if engine is available
	var bootstrapRTID statechart.RuntimeID
	if k.engine != nil {
		bootstrapRTID, err = k.engine.Spawn(def, nil)
		if err != nil {
			return fmt.Errorf("failed to spawn bootstrap runtime: %w", err)
		}
		log.Printf("[kernel] Spawning bootstrap runtime: %s", bootstrapRTID)

		// Register bootstrap service
		k.mu.Lock()
		k.services["sys:bootstrap"] = bootstrapRTID
		k.mu.Unlock()

		// Start the bootstrap runtime
		if err := k.engine.Control(bootstrapRTID, statechart.CmdStart); err != nil {
			return fmt.Errorf("failed to start bootstrap runtime: %w", err)
		}
	}

	// Create bootstrap sequence
	seq := bootstrap.NewSequence()
	k.mu.Lock()
	k.sequence = seq
	k.readyChan = make(chan struct{})
	k.mu.Unlock()

	// Set up state entry handlers
	seq.OnStateEnter(func(state string) error {
		return k.onBootstrapStateEnter(ctx, state, bootstrapRTID)
	})

	// Set up completion handler
	seq.OnComplete(func() {
		log.Println("[kernel] Bootstrap complete, handing off to ChartRegistry")
		k.onBootstrapComplete()
		k.mu.Lock()
		if k.readyChan != nil {
			close(k.readyChan)
		}
		k.mu.Unlock()
	})

	// Start the sequence
	if err := seq.Start(ctx); err != nil {
		return fmt.Errorf("bootstrap failed: %w", err)
	}

	// Wait for completion or cancellation
	select {
	case <-k.readyChan:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (k *Kernel) onBootstrapStateEnter(ctx context.Context, state string, bootstrapRTID statechart.RuntimeID) error {
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

// GetServiceRuntimeID returns the RuntimeID for a service.
func (k *Kernel) GetServiceRuntimeID(name string) (statechart.RuntimeID, bool) {
	k.mu.RLock()
	defer k.mu.RUnlock()
	id, ok := k.services[name]
	return id, ok
}
