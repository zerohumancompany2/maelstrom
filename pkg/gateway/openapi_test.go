package gateway

import (
	"testing"

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
