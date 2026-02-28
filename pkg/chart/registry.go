package chart

import (
	"context"
	"time"

	"github.com/maelstrom/v3/pkg/registry"
	"github.com/maelstrom/v3/pkg/source"
)

// ChartRegistry loads and manages chart definitions from a directory.
type ChartRegistry struct {
	dir       string
	hydrator  HydratorFunc
	src       source.Source
	reg       *registry.Registry
	service   *registry.Service
	observers []func(key string, def ChartDefinition)
	cancel    context.CancelFunc
}

// NewChartRegistry creates a registry that watches dir for YAML files.
func NewChartRegistry(dir string, hydrator HydratorFunc) (*ChartRegistry, error) {
	// Create file system source
	src, err := source.NewFileSystemSource(dir, 100*time.Millisecond)
	if err != nil {
		return nil, err
	}

	// Create registry and service
	reg := registry.New()
	svc := registry.NewService(src, reg)

	// Set up hydrator
	svc.SetHydrator(func(content []byte) (interface{}, error) {
		return hydrator(content)
	})

	cr := &ChartRegistry{
		dir:      dir,
		hydrator: hydrator,
		src:      src,
		reg:      reg,
		service:  svc,
	}

	// Set up observer callback
	svc.OnChange(func(key string, value interface{}) {
		if def, ok := value.(ChartDefinition); ok {
			for _, fn := range cr.observers {
				fn(key, def)
			}
		}
	})

	return cr, nil
}

// Start begins watching and hydrating charts. Blocks until ctx is cancelled.
func (r *ChartRegistry) Start(ctx context.Context) error {
	// Start the source in a goroutine
	go func() {
		r.src.(*source.FileSystemSource).Run()
	}()

	// Give source time to do initial scan
	time.Sleep(100 * time.Millisecond)

	ctx, cancel := context.WithCancel(ctx)
	r.cancel = cancel
	return r.service.Run(ctx)
}

// Stop gracefully shuts down the registry.
func (r *ChartRegistry) Stop() error {
	if r.cancel != nil {
		r.cancel()
	}
	return r.src.(*source.FileSystemSource).Stop()
}

// Get retrieves the current version of a chart.
func (r *ChartRegistry) Get(name string) (ChartDefinition, error) {
	val, err := r.reg.Get(name)
	if err != nil {
		return ChartDefinition{}, err
	}
	return val.(ChartDefinition), nil
}

// GetVersion retrieves a specific version of a chart.
func (r *ChartRegistry) GetVersion(name string, version int) (ChartDefinition, error) {
	val, err := r.reg.GetVersion(name, version)
	if err != nil {
		return ChartDefinition{}, err
	}
	return val.(ChartDefinition), nil
}

// ListVersions returns all versions of a chart.
func (r *ChartRegistry) ListVersions(name string) ([]registry.Version, error) {
	return nil, nil // TODO: expose from registry
}

// OnChange registers a callback for chart updates.
func (r *ChartRegistry) OnChange(fn func(key string, def ChartDefinition)) {
	r.observers = append(r.observers, fn)
}
