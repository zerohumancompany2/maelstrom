// Package platform provides PlatformService types for YAML schema.
// Spec Reference: Section 13.7
package platform

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/maelstrom/v3/pkg/statechart"
	"gopkg.in/yaml.v3"
)

// PlatformService represents a PlatformService YAML definition.
type PlatformService struct {
	APIVersion string   `yaml:"apiVersion"`
	Kind       string   `yaml:"kind"`
	Metadata   Metadata `yaml:"metadata"`
	Spec       Spec     `yaml:"spec"`
}

// Metadata contains metadata for a PlatformService.
type Metadata struct {
	Name        string `yaml:"name"`
	Core        bool   `yaml:"core,omitempty"`
	DisplayName string `yaml:"displayName,omitempty"`
}

// Spec contains the specification for a PlatformService.
type Spec struct {
	ChartRef               string       `yaml:"chartRef"`
	RequiredForKernelReady bool         `yaml:"requiredForKernelReady,omitempty"`
	Replicas               *int32       `yaml:"replicas,omitempty"`
	Persistence            *Persistence `yaml:"persistence,omitempty"`
	Expose                 *Expose      `yaml:"expose,omitempty"`
}

// GetReplicas returns the replica count with default of 1.
func (s Spec) GetReplicas() int32 {
	if s.Replicas == nil {
		return 1
	}
	return *s.Replicas
}

// Persistence defines persistence configuration for a PlatformService.
type Persistence struct {
	Enabled       bool   `yaml:"enabled,omitempty"`
	SnapshotEvery string `yaml:"snapshotEvery,omitempty"`
}

// Expose defines exposure configuration for a PlatformService.
type Expose struct {
	HTTP *HTTPExpose `yaml:"http,omitempty"`
	TCP  *TCPExpose  `yaml:"tcp,omitempty"`
}

// HTTPExpose defines HTTP exposure configuration.
type HTTPExpose struct {
	Port int32 `yaml:"port,omitempty"`
	TLS  bool  `yaml:"tls,omitempty"`
}

// TCPExpose defines TCP exposure configuration.
type TCPExpose struct {
	Port int32 `yaml:"port,omitempty"`
}

// Validate validates the PlatformService YAML schema.
func (ps PlatformService) Validate() error {
	if ps.APIVersion == "" {
		return fmt.Errorf("apiVersion is required")
	}
	if ps.Kind != "PlatformService" {
		return fmt.Errorf("kind must be PlatformService")
	}
	if ps.Metadata.Name == "" {
		return fmt.Errorf("metadata.name is required")
	}
	if ps.Spec.ChartRef == "" {
		return fmt.Errorf("spec.chartRef is required")
	}
	return nil
}

// ToChartDefinition converts a PlatformService to a ChartDefinition.
func (ps PlatformService) ToChartDefinition(hydrator statechart.HydratorFunc) (statechart.ChartDefinition, error) {
	chartYAML := fmt.Sprintf(`
id: %s
version: 1.0.0
root:
  id: root
  children:
    init:
      id: init
      type: atomic
      transitions:
        - event: START
          target: running
    running:
      id: running
      type: atomic
    stopped:
      id: stopped
      type: final
initialState: init
`, ps.Metadata.Name)

	return hydrator([]byte(chartYAML))
}

// ChartRegistry loads PlatformService YAMLs from a directory.
type ChartRegistry struct {
	dir string
}

// NewChartRegistry creates a new ChartRegistry for the given directory.
func NewChartRegistry(dir string) *ChartRegistry {
	return &ChartRegistry{dir: dir}
}

// LoadPlatformServices loads all PlatformService YAMLs from the registry directory.
func (cr *ChartRegistry) LoadPlatformServices() ([]PlatformService, error) {
	files, err := filepath.Glob(filepath.Join(cr.dir, "*.yaml"))
	if err != nil {
		return nil, fmt.Errorf("glob: %w", err)
	}

	if len(files) == 0 {
		return []PlatformService{}, nil
	}

	// Sort for deterministic order
	sort.Strings(files)

	services := make([]PlatformService, 0, len(files))
	for _, file := range files {
		content, err := os.ReadFile(file)
		if err != nil {
			return nil, fmt.Errorf("read file %s: %w", file, err)
		}

		var ps PlatformService
		if err := yaml.Unmarshal(content, &ps); err != nil {
			return nil, fmt.Errorf("parse yaml %s: %w", file, err)
		}

		if err := ps.Validate(); err != nil {
			return nil, fmt.Errorf("validate %s: %w", file, err)
		}

		services = append(services, ps)
	}

	return services, nil
}

// GetCoreServices returns all services marked as core.
func (cr *ChartRegistry) GetCoreServices(services []PlatformService) []PlatformService {
	var core []PlatformService
	for _, svc := range services {
		if svc.Metadata.Core {
			core = append(core, svc)
		}
	}
	return core
}

// GetNonCoreServices returns all services not marked as core.
func (cr *ChartRegistry) GetNonCoreServices(services []PlatformService) []PlatformService {
	var nonCore []PlatformService
	for _, svc := range services {
		if !svc.Metadata.Core {
			nonCore = append(nonCore, svc)
		}
	}
	return nonCore
}

// ValidateCoreServices validates core service requirements.
// Returns error if any core service is missing requiredForKernelReady.
func (cr *ChartRegistry) ValidateCoreServices(services []PlatformService) error {
	coreServices := cr.GetCoreServices(services)
	for _, svc := range coreServices {
		if !svc.Spec.RequiredForKernelReady {
			return fmt.Errorf("core service %s should have requiredForKernelReady: true", svc.Metadata.Name)
		}
	}
	return nil
}
