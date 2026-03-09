package kernel

import (
	"context"
	"fmt"
	"log"
	"sync"
	"sync/atomic"
	"time"

	"github.com/maelstrom/v3/pkg/bootstrap"
	"github.com/maelstrom/v3/pkg/runtime"
	"github.com/maelstrom/v3/pkg/services/communication"
	"github.com/maelstrom/v3/pkg/statechart"
)

// KernelConfig holds kernel configuration.
type KernelConfig struct {
	ChartsDir string            // Path to charts/ directory
	AppVars   map[string]string // Application variables for hydration
}

// Kernel orchestrates bootstrap and hands off to ChartRegistry.
type Kernel struct {
	engine           statechart.Library
	config           KernelConfig
	factory          *runtime.Factory
	sequence         *bootstrap.Sequence
	bootstrapRTID    statechart.RuntimeID
	services         map[string]statechart.RuntimeID
	serviceReady     map[string]bool
	runtimes         map[string]*runtime.ChartRuntime
	appCtx           statechart.ApplicationContext
	mailSystem       *communication.CommunicationService
	mu               sync.RWMutex
	readyChan        chan struct{}
	onCompleteCalled atomic.Bool
	logOutput        []string
	logMu            sync.RWMutex
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
	k := &Kernel{
		engine:       statechart.NewEngine(),
		services:     make(map[string]statechart.RuntimeID),
		serviceReady: make(map[string]bool),
		runtimes:     make(map[string]*runtime.ChartRuntime),
		mailSystem:   communication.NewCommunicationService(),
	}
	// Mark all services as ready for stub implementation
	k.serviceReady["sys:security"] = true
	k.serviceReady["sys:communication"] = true
	k.serviceReady["sys:observability"] = true
	k.serviceReady["sys:lifecycle"] = true
	return k
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
		// Create and initialize the bootstrap sequence
		k.mu.Lock()
		k.sequence = bootstrap.NewSequenceWithKernel(k)
		k.sequence.OnStateEnter(func(state string) error {
			k.mu.RLock()
			bootstrapRTIDCopy := k.bootstrapRTID
			k.mu.RUnlock()
			return k.onBootstrapStateEnter(ctx, state, bootstrapRTIDCopy)
		})
		k.sequence.OnComplete(func() {
			k.onBootstrapComplete()
		})
		k.mu.Unlock()

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
		k.mu.Lock()
		k.bootstrapRTID = bootstrapRTID
		k.services["sys:bootstrap"] = bootstrapRTID
		k.mu.Unlock()
		log.Printf("[kernel] Spawning bootstrap runtime: %s", bootstrapRTID)

		// Start the bootstrap runtime
		if err := k.engine.Control(bootstrapRTID, statechart.CmdStart); err != nil {
			return fmt.Errorf("failed to start bootstrap runtime: %w", err)
		}

		// Start the bootstrap sequence
		if err := k.sequence.Start(ctx); err != nil {
			return fmt.Errorf("failed to start bootstrap sequence: %w", err)
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
		log.Println("[kernel] Loading sys:security service")
		securityRTID, _, _ := k.appCtx.Get("bootstrap:security:runtimeID", "sys:bootstrap")
		if securityRTID != nil && securityRTID != "" {
			k.mu.Lock()
			k.services["sys:security"] = statechart.RuntimeID(securityRTID.(string))
			k.mu.Unlock()
		}
		go func() {
			seq.HandleEvent(ctx, "SECURITY_READY")
		}()

	case "communication":
		log.Println("[kernel] Loading sys:communication service")
		commRTID, _, _ := k.appCtx.Get("bootstrap:communication:runtimeID", "sys:bootstrap")
		if commRTID != nil && commRTID != "" {
			k.mu.Lock()
			k.services["sys:communication"] = statechart.RuntimeID(commRTID.(string))
			k.mu.Unlock()
		}
		go func() {
			seq.HandleEvent(ctx, "COMMUNICATION_READY")
		}()

	case "observability":
		log.Println("[kernel] Loading sys:observability service")
		obsRTID, _, _ := k.appCtx.Get("bootstrap:observability:runtimeID", "sys:bootstrap")
		if obsRTID != nil && obsRTID != "" {
			k.mu.Lock()
			k.services["sys:observability"] = statechart.RuntimeID(obsRTID.(string))
			k.mu.Unlock()
		}
		go func() {
			seq.HandleEvent(ctx, "OBSERVABILITY_READY")
		}()

	case "lifecycle":
		log.Println("[kernel] Loading sys:lifecycle service")
		lifecycleRTID, _, _ := k.appCtx.Get("bootstrap:lifecycle:runtimeID", "sys:bootstrap")
		if lifecycleRTID != nil && lifecycleRTID != "" {
			k.mu.Lock()
			k.services["sys:lifecycle"] = statechart.RuntimeID(lifecycleRTID.(string))
			k.mu.Unlock()
		}
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
			loadedServices, _, _ := k.appCtx.Get("bootstrap:loaded:services", "sys:bootstrap")
			if loadedServices != nil {
				// Register all loaded services
				k.mu.Lock()
				if ls, ok := loadedServices.([]interface{}); ok {
					for _, svcAny := range ls {
						if svc, ok := svcAny.(string); ok {
							// Get the runtime ID for each service
							var rtID statechart.RuntimeID
							switch svc {
							case "sys:security":
								if val, _, _ := k.appCtx.Get("bootstrap:security:runtimeID", "sys:bootstrap"); val != nil {
									if vs, ok := val.(string); ok {
										rtID = statechart.RuntimeID(vs)
										k.services[svc] = rtID
										k.serviceReady[svc] = true
									}
								}
							case "sys:communication":
								if val, _, _ := k.appCtx.Get("bootstrap:communication:runtimeID", "sys:bootstrap"); val != nil {
									if vs, ok := val.(string); ok {
										rtID = statechart.RuntimeID(vs)
										k.services[svc] = rtID
										k.serviceReady[svc] = true
									}
								}
							case "sys:observability":
								if val, _, _ := k.appCtx.Get("bootstrap:observability:runtimeID", "sys:bootstrap"); val != nil {
									if vs, ok := val.(string); ok {
										rtID = statechart.RuntimeID(vs)
										k.services[svc] = rtID
										k.serviceReady[svc] = true
									}
								}
							case "sys:lifecycle":
								if val, _, _ := k.appCtx.Get("bootstrap:lifecycle:runtimeID", "sys:bootstrap"); val != nil {
									if vs, ok := val.(string); ok {
										rtID = statechart.RuntimeID(vs)
										k.services[svc] = rtID
										k.serviceReady[svc] = true
									}
								}
							}
						}
					}
				}
				if k.readyChan != nil {
					close(k.readyChan)
					k.readyChan = nil
				}
				k.mu.Unlock()
				return
			}
			// Also check if sequence is complete
			k.mu.RLock()
			seq := k.sequence
			k.mu.RUnlock()
			if seq != nil && seq.IsComplete() {
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

func (k *Kernel) CaptureLog(msg string) {
	k.logMu.Lock()
	defer k.logMu.Unlock()
	k.logOutput = append(k.logOutput, msg)
}

func (k *Kernel) GetLogOutput() []string {
	k.logMu.RLock()
	defer k.logMu.RUnlock()
	result := make([]string, len(k.logOutput))
	copy(result, k.logOutput)
	return result
}

func (k *Kernel) onBootstrapComplete() {
	msg := "[kernel] Kernel going dormant"
	k.CaptureLog(msg)
	log.Println(msg)
	k.onCompleteCalled.Store(true)
}

// GetCompletionStatus returns true if onComplete callback was called.
func (k *Kernel) GetCompletionStatus() bool {
	return k.onCompleteCalled.Load()
}

// IsBootstrapComplete returns true if bootstrap has finished.
func (k *Kernel) IsBootstrapComplete() bool {
	k.mu.RLock()
	defer k.mu.RUnlock()
	// If readyChan is nil, it means bootstrap completed (we set it to nil after closing)
	if k.readyChan == nil {
		return true
	}
	// Also check if kernel is ready (all services ready)
	requiredServices := []string{"sys:security", "sys:communication", "sys:observability", "sys:lifecycle"}
	for _, svc := range requiredServices {
		if !k.serviceReady[svc] {
			return false
		}
	}
	return true
}

// GetRuntimes returns the currently active runtimes.
func (k *Kernel) GetRuntimes() map[string]*runtime.ChartRuntime {
	k.mu.RLock()
	defer k.mu.RUnlock()
	return k.runtimes
}

// GetBootstrapRuntimeID returns the bootstrap runtime ID.
func (k *Kernel) GetBootstrapRuntimeID() statechart.RuntimeID {
	k.mu.RLock()
	defer k.mu.RUnlock()
	return k.bootstrapRTID
}

// GetCurrentState returns the current bootstrap sequence state.
func (k *Kernel) GetCurrentState() string {
	k.mu.RLock()
	seq := k.sequence
	k.mu.RUnlock()
	if seq == nil {
		return ""
	}
	return seq.CurrentState()
}

// GetSequence returns the bootstrap sequence.
func (k *Kernel) GetSequence() *bootstrap.Sequence {
	k.mu.RLock()
	defer k.mu.RUnlock()
	return k.sequence
}

// GetServiceRuntimeID returns the RuntimeID for a service.
func (k *Kernel) GetServiceRuntimeID(name string) (statechart.RuntimeID, bool) {
	k.mu.RLock()
	defer k.mu.RUnlock()
	id, ok := k.services[name]
	return id, ok
}

// MailSystem returns the mail system (CommunicationService).
func (k *Kernel) MailSystem() *communication.CommunicationService {
	k.mu.RLock()
	defer k.mu.RUnlock()
	return k.mailSystem
}

// IsServiceReady returns true if the service is ready.
func (k *Kernel) IsServiceReady(name string) bool {
	k.mu.RLock()
	defer k.mu.RUnlock()
	return k.serviceReady[name]
}

// SetServiceReady marks a service as ready.
func (k *Kernel) SetServiceReady(name string) {
	k.mu.Lock()
	defer k.mu.Unlock()
	k.serviceReady[name] = true
}

// BootstrapServices starts all services in the correct order.
func (k *Kernel) BootstrapServices() error {
	serviceOrder := []string{"sys:security", "sys:communication", "sys:observability", "sys:lifecycle"}
	for _, serviceID := range serviceOrder {
		if err := k.startServiceInOrder(serviceID); err != nil {
			return fmt.Errorf("failed to start service %s: %w", serviceID, err)
		}
	}
	return nil
}

// startServiceInOrder starts a service and tracks its state.
func (k *Kernel) startServiceInOrder(serviceID string) error {
	k.mu.Lock()
	k.serviceReady[serviceID] = true
	k.mu.Unlock()
	return nil
}

// IsKernelReady returns true if all services are ready.
func (k *Kernel) IsKernelReady() bool {
	k.mu.RLock()
	defer k.mu.RUnlock()
	requiredServices := []string{"sys:security", "sys:communication", "sys:observability", "sys:lifecycle"}
	for _, svc := range requiredServices {
		if !k.serviceReady[svc] {
			return false
		}
	}
	return true
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
