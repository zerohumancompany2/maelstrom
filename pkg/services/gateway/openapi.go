package gateway

// PathItem represents an OpenAPI path item with operations
type PathItem struct {
	Post *Operation `json:"post,omitempty"`
	Get  *Operation `json:"get,omitempty"`
}

// Operation represents an OpenAPI operation
type Operation struct {
	OperationID string `json:"operationId"`
}

// OpenAPISpec represents an OpenAPI 3.1 specification
type OpenAPISpec struct {
	OpenAPI string               `json:"openapi"`
	Info    *OpenAPIInfo         `json:"info"`
	Paths   map[string]*PathItem `json:"paths"`
}

// OpenAPIInfo contains metadata about the API
type OpenAPIInfo struct {
	Title   string `json:"title"`
	Version string `json:"version"`
}

// OpenAPIGen generates OpenAPI 3.1 specifications
type OpenAPIGen struct{}

// NewOpenAPIGen creates a new OpenAPI generator
func NewOpenAPIGen() *OpenAPIGen {
	return &OpenAPIGen{}
}

// GenerateSpec generates an OpenAPI 3.1 spec from charts
func (g *OpenAPIGen) GenerateSpec(charts []Chart) (*OpenAPISpec, error) {
	spec := &OpenAPISpec{
		OpenAPI: "3.1.0",
		Info: &OpenAPIInfo{
			Title:   "Maelstrom API",
			Version: "1.0.0",
		},
		Paths: make(map[string]*PathItem),
	}

	for _, chart := range charts {
		if !canExposeChart(chart) {
			continue
		}
		if chart.Expose == nil || chart.Expose.HTTP == nil {
			continue
		}

		path := chart.Expose.HTTP.Path
		pathItem, exists := spec.Paths[path]
		if !exists {
			pathItem = &PathItem{}
			spec.Paths[path] = pathItem
		}

		chartName := chart.Name
		for _, event := range chart.Expose.HTTP.Events {
			operationID := chartName + "_" + event.Trigger
			operationID = normalizeOperationID(operationID)

			switch event.Method {
			case "POST":
				pathItem.Post = &Operation{OperationID: operationID}
			case "GET":
				pathItem.Get = &Operation{OperationID: operationID}
			}
		}
	}

	return spec, nil
}

func canExposeChart(chart Chart) bool {
	if chart.Boundary == "inner" {
		return false
	}
	return chart.Boundary == "dmz" || chart.Boundary == "outer"
}

func normalizeOperationID(id string) string {
	result := ""
	for _, c := range id {
		if c == ':' {
			result += "_"
		} else {
			result += string(c)
		}
	}
	return result
}
