package gateway

import (
	"gopkg.in/yaml.v3"
)

// HTTPExposure represents chart-declared HTTP endpoints
type HTTPExposure struct {
	Path   string      `yaml:"path"`
	Events []HTTPEvent `yaml:"events"`
}

// HTTPEvent represents an HTTP event mapping
type HTTPEvent struct {
	Trigger        string `yaml:"trigger"`
	Method         string `yaml:"method"`
	RequestSchema  string `yaml:"requestSchema"`
	ResponseSchema string `yaml:"responseSchema"`
	Stream         bool   `yaml:"stream"`
}

// Exposure represents the expose block in a chart
type Exposure struct {
	HTTP *HTTPExposure `yaml:"http"`
}

// Chart represents a chart definition with expose block
type Chart struct {
	Name     string    `yaml:"name"`
	Boundary string    `yaml:"boundary"`
	Expose   *Exposure `yaml:"expose"`
	States   []State   `yaml:"states"`
}

// State represents a state in a chart
type State struct {
	Name string                `yaml:"name"`
	On   map[string]Transition `yaml:"on"`
}

// Transition represents a state transition
type Transition struct {
	To string `yaml:"to"`
}

// ParseChart parses a chart from YAML
func ParseChart(yamlContent string) (*Chart, error) {
	var chart Chart
	if err := yaml.Unmarshal([]byte(yamlContent), &chart); err != nil {
		return nil, err
	}
	return &chart, nil
}

// ExposureService handles boundary enforcement
type ExposureService struct{}

// NewExposureService creates a new exposure service
func NewExposureService() *ExposureService {
	return &ExposureService{}
}

// CanExpose checks if a chart can be exposed based on boundary
func (e *ExposureService) CanExpose(chart Chart) bool {
	if chart.Boundary == "inner" {
		return false
	}
	return chart.Boundary == "dmz" || chart.Boundary == "outer"
}

// MapEventToAPI maps event surface to API surface
func (e *ExposureService) MapEventToAPI(chart Chart) ([]APIEndpoint, error) {
	return nil, nil
}

// APIEndpoint represents an API endpoint
type APIEndpoint struct {
	Trigger string
	Method  string
	Path    string
}
