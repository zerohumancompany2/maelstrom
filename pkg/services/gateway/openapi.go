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
	return nil, nil
}
