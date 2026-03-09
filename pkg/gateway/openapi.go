package gateway

import (
	"github.com/maelstrom/v3/pkg/mail"
	"github.com/maelstrom/v3/pkg/statechart"
)

type OpenAPISpec struct {
	OpenAPI    string                 `json:"openapi"`
	Info       Info                   `json:"info"`
	Paths      map[string]interface{} `json:"paths"`
	Components Components             `json:"components"`
}

type Info struct {
	Title   string `json:"title"`
	Version string `json:"version"`
}

type Components struct {
	SecuritySchemes map[string]interface{} `json:"securitySchemes"`
}

type HTTPSpec struct {
	Endpoints []HTTPEndpoint `json:"endpoints"`
}

type HTTPEndpoint struct {
	Path       string            `json:"path"`
	Method     string            `json:"method"`
	Boundary   mail.BoundaryType `json:"boundary"`
	Auth       bool              `json:"auth"`
	Parameters map[string]string `json:"parameters,omitempty"`
}

type Chart = statechart.ChartDefinition

func GenerateOpenAPIFromChart(chart *Chart) (*OpenAPISpec, error) {
	if chart == nil {
		return nil, nil
	}

	spec := &OpenAPISpec{
		OpenAPI: "3.0.0",
		Info: Info{
			Title:   chart.ID,
			Version: chart.Version,
		},
		Paths: make(map[string]interface{}),
		Components: Components{
			SecuritySchemes: make(map[string]interface{}),
		},
	}

	return spec, nil
}

func GenerateHTTPExposureSpec(chart *Chart) (*HTTPSpec, error) {
	return nil, nil
}
