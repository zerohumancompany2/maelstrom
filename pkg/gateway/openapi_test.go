package gateway

import (
	"testing"

	"github.com/maelstrom/v3/pkg/mail"
	"github.com/maelstrom/v3/pkg/statechart"
)

func TestOpenAPI_GenerateFromChart(t *testing.T) {
	chart := &statechart.ChartDefinition{
		ID:      "test-chart",
		Version: "1.0.0",
	}

	spec, err := GenerateOpenAPIFromChart(chart)
	if err != nil {
		t.Fatalf("Expected nil error, got %v", err)
	}

	if spec == nil {
		t.Fatal("Expected non-nil OpenAPI spec")
	}

	if spec.Info.Title != "test-chart" {
		t.Errorf("Expected title 'test-chart', got '%s'", spec.Info.Title)
	}

	if spec.Info.Version != "1.0.0" {
		t.Errorf("Expected version '1.0.0', got '%s'", spec.Info.Version)
	}
}

func TestOpenAPI_HTTPExposure(t *testing.T) {
	chart := &statechart.ChartDefinition{
		ID:      "test-chart",
		Version: "1.0.0",
	}

	spec, err := GenerateHTTPExposureSpec(chart)
	if err != nil {
		t.Fatalf("Expected nil error, got %v", err)
	}

	if spec == nil {
		t.Fatal("Expected non-nil HTTP spec")
	}

	if len(spec.Endpoints) == 0 {
		t.Error("Expected at least one endpoint")
	}
}

func TestOpenAPI_AuthMiddleware(t *testing.T) {
	chart := &statechart.ChartDefinition{
		ID:      "test-chart",
		Version: "1.0.0",
	}

	spec, err := GenerateOpenAPIFromChart(chart)
	if err != nil {
		t.Fatalf("Expected nil error, got %v", err)
	}

	if spec == nil {
		t.Fatal("Expected non-nil OpenAPI spec")
	}

	if spec.Components.SecuritySchemes == nil {
		t.Fatal("Expected non-nil security schemes")
	}

	if _, ok := spec.Components.SecuritySchemes["bearerAuth"]; !ok {
		t.Error("Expected bearerAuth security scheme")
	}
}

func TestOpenAPI_BoundaryFiltering(t *testing.T) {
	chart := &statechart.ChartDefinition{
		ID:      "test-chart",
		Version: "1.0.0",
	}

	spec, err := GenerateHTTPExposureSpec(chart)
	if err != nil {
		t.Fatalf("Expected nil error, got %v", err)
	}

	if spec == nil {
		t.Fatal("Expected non-nil HTTP spec")
	}

	for _, endpoint := range spec.Endpoints {
		if endpoint.Boundary == mail.InnerBoundary {
			t.Error("Inner boundary endpoints should not be exposed")
		}
	}
}
