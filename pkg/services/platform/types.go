// Package platform provides PlatformService types for YAML schema.
// Spec Reference: Section 13.7
package platform

import "fmt"

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
