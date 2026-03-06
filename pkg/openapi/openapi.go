package openapi

// Spec represents an OpenAPI specification
type Spec struct {
	Version string
	Info    Info
	Paths   map[string]interface{}
}

// Info contains metadata about the API
type Info struct {
	Title   string
	Version string
}
