package kernel

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

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
	appCtx    statechart.ApplicationContext
	mu        sync.RWMutex
	readyChan chan struct{}
}

// kernelApplicationContext provides application context with kernel engine access.
type kernelApplicationContext struct {
	kernel *Kernel
	data   map[string]interface{}
	mu     sync.RWMutex
}

func (k *kernelApplicationContext) Get(key string, callerBoundary string) (interface{}, []string, error) {
	k.mu.RLock()
	defer k.mu.RUnlock()
	if key == "__engine" {
		return k.kernel.engine, nil, nil
	}
	val, ok := k.data[key]
	if !ok {
		return nil, nil, nil
	}
	return val, nil, nil
}

func (k *kernelApplicationContext) Set(key string, value interface{}, taints []string, callerBoundary string) error {
	k.mu.Lock()
	defer k.mu.Unlock()
	k.data[key] = value
	return nil
}

func (k *kernelApplicationContext) Namespace() string {
	return "sys:kernel"
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

// RegisterBootstrapActions registers the bootstrap actions.
func (k *Kernel) RegisterBootstrapActions() {
	if k.engine == nil {
		return
	}
	// Register the new service-loading actions
	k.engine.RegisterAction(bootstrap.ActionLoadSecurityService, bootstrap.LoadSecurityService)
	k.engine.RegisterAction(bootstrap.ActionLoadCommunicationService, bootstrap.LoadCommunicationService)
	k.engine.RegisterAction(bootstrap.ActionLoadObservabilityService, bootstrap.LoadObservabilityService)
	k.engine.RegisterAction(bootstrap.ActionLoadLifecycleService, bootstrap.LoadLifecycleService)
	k.engine.RegisterAction(bootstrap.ActionSignalKernelReady, bootstrap.SignalKernelReady)
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
		// Create appCtx with engine reference for actions to use
		appCtx := &kernelApplicationContext{
			kernel: k,
			data:   make(map[string]interface{}),
		}
		k.appCtx = appCtx
		bootstrapRTID, err = k.engine.Spawn(def, appCtx)
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

		// Dispatch START_BOOTSTRAP event to begin the bootstrap flow
		if err := k.engine.Dispatch(bootstrapRTID, statechart.Event{Type: "START_BOOTSTRAP"}); err != nil {
			return fmt.Errorf("failed to dispatch START_BOOTSTRAP: %w", err)
		}
		log.Println("[kernel] Dispatched START_BOOTSTRAP event")

		// Wait for bootstrap to complete or context cancellation
		// The bootstrap chart will emit KERNEL_READY when complete
		k.readyChan = make(chan struct{})
		go k.waitForKernelReady(ctx, bootstrapRTID)

		select {
		case <-k.readyChan:
			log.Println("[kernel] Bootstrap complete, handing off to ChartRegistry")
			k.onBootstrapComplete()
			return nil
		case <-ctx.Done():
			return ctx.Err()
		}
	}
	return nil
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

func (k *Kernel) waitForKernelReady(ctx context.Context, bootstrapRTID statechart.RuntimeID) {
	ticker := time.NewTicker(10 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			// Check if bootstrap has reached the ready state
			// by checking if KERNEL_READY was processed
			services, _, _ := k.appCtx.Get("bootstrap:loaded:services", "sys:bootstrap")
			if services != nil {
				k.mu.Lock()
				if k.readyChan != nil {
					close(k.readyChan)
					k.readyChan = nil
				}
				k.mu.Unlock()
				return
			}
		}
	}
}

func (k *Kernel) onBootstrapComplete() {
	log.Println("[kernel] Kernel going dormant")
}

// IsBootstrapComplete returns true if bootstrap has finished.
func (k *Kernel) IsBootstrapComplete() bool {
	k.mu.RLock()
	defer k.mu.RUnlock()
	// If readyChan is nil, it means bootstrap completed (we set it to nil after closing)
	return k.readyChan == nil
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

// Shutdown stops all services.
func (k *Kernel) Shutdown(ctx context.Context) error {
	if k.engine == nil {
		return nil
	}
	k.mu.RLock()
	services := make(map[string]statechart.RuntimeID, len(k.services))
	for name, id := range k.services {
		services[name] = id
	}
	k.mu.RUnlock()

	for name, id := range services {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
		if err := k.engine.Control(id, statechart.CmdStop); err != nil {
			log.Printf("[kernel] failed to stop %s: %v", name, err)
		}
	}
	return nil
}
