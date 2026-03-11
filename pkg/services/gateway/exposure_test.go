package gateway

import (
	"testing"
)

func TestHTTPExposure_OpenAPI31Generation(t *testing.T) {
	gen := NewOpenAPIGen()

	// Setup: Charts with expose blocks
	charts := []Chart{
		{
			Name:     "agent:dmz",
			Boundary: "dmz",
			Expose: &Exposure{
				HTTP: &HTTPExposure{
					Path: "/api/v1/agents/{id}/",
					Events: []HTTPEvent{
						{
							Trigger:        "user_query",
							Method:         "POST",
							RequestSchema:  "UserQuery",
							ResponseSchema: "AgentResponse",
							Stream:         true,
						},
						{
							Trigger: "status",
							Method:  "GET",
						},
					},
				},
			},
		},
	}

	// Auto-generate OpenAPI 3.1 spec (arch-v1.md L722)
	spec, err := gen.GenerateSpec(charts)
	if err != nil {
		t.Fatalf("Expected no error generating spec, got %v", err)
	}

	// Verify OpenAPI 3.1 version (arch-v1.md L722)
	if spec.OpenAPI != "3.1.0" {
		t.Errorf("Expected OpenAPI version '3.1.0', got '%s'", spec.OpenAPI)
	}

	// Verify info section
	if spec.Info == nil {
		t.Error("Expected spec to have info section")
	}

	// Verify paths generated (arch-v1.md L722)
	if spec.Paths == nil {
		t.Error("Expected spec to have paths")
	}

	// Verify /api/v1/agents/{id}/ path exists
	if _, exists := spec.Paths["/api/v1/agents/{id}/"]; !exists {
		t.Error("Expected path '/api/v1/agents/{id}/' in spec")
	}

	// Verify POST operation generated
	pathItem := spec.Paths["/api/v1/agents/{id}/"]
	if pathItem.Post == nil {
		t.Error("Expected POST operation on path")
	}

	// Verify GET operation generated
	if pathItem.Get == nil {
		t.Error("Expected GET operation on path")
	}

	// Verify operation IDs based on triggers
	if pathItem.Post.OperationID != "agent_dmz_user_query" {
		t.Errorf("Expected operationId 'agent_dmz_user_query', got '%s'", pathItem.Post.OperationID)
	}
	if pathItem.Get.OperationID != "agent_dmz_status" {
		t.Errorf("Expected operationId 'agent_dmz_status', got '%s'", pathItem.Get.OperationID)
	}
}

func TestHTTPExposure_ChartDeclarativeEndpoints(t *testing.T) {
	// Setup: Chart with expose block (arch-v1.md L708-720)
	chartYAML := `
name: agent:dmz
boundary: dmz
expose:
  http:
    path: /api/v1/agents/{id}/
    events:
      - trigger: user_query
        method: POST
        requestSchema: UserQuery
        responseSchema: AgentResponse
        stream: true
      - trigger: status
        method: GET
`
	chart, err := ParseChart(chartYAML)
	if err != nil {
		t.Fatalf("Expected no error parsing chart, got %v", err)
	}

	// Verify expose block parsed correctly (arch-v1.md L708-720)
	if chart.Expose == nil {
		t.Error("Expected chart to have expose block")
	}

	if chart.Expose.HTTP == nil {
		t.Error("Expected chart to have HTTP exposure")
	}

	// Verify path (arch-v1.md L711)
	if chart.Expose.HTTP.Path != "/api/v1/agents/{id}/" {
		t.Errorf("Expected path '/api/v1/agents/{id}/', got '%s'", chart.Expose.HTTP.Path)
	}

	// Verify events (arch-v1.md L712-720)
	if len(chart.Expose.HTTP.Events) != 2 {
		t.Errorf("Expected 2 events, got %d", len(chart.Expose.HTTP.Events))
	}

	// Verify first event: user_query POST (arch-v1.md L713-717)
	if chart.Expose.HTTP.Events[0].Trigger != "user_query" {
		t.Errorf("Expected trigger 'user_query', got '%s'", chart.Expose.HTTP.Events[0].Trigger)
	}
	if chart.Expose.HTTP.Events[0].Method != "POST" {
		t.Errorf("Expected method 'POST', got '%s'", chart.Expose.HTTP.Events[0].Method)
	}
	if chart.Expose.HTTP.Events[0].RequestSchema != "UserQuery" {
		t.Errorf("Expected requestSchema 'UserQuery', got '%s'", chart.Expose.HTTP.Events[0].RequestSchema)
	}
	if chart.Expose.HTTP.Events[0].ResponseSchema != "AgentResponse" {
		t.Errorf("Expected responseSchema 'AgentResponse', got '%s'", chart.Expose.HTTP.Events[0].ResponseSchema)
	}
	if !chart.Expose.HTTP.Events[0].Stream {
		t.Error("Expected stream to be true")
	}

	// Verify second event: status GET (arch-v1.md L718-720)
	if chart.Expose.HTTP.Events[1].Trigger != "status" {
		t.Errorf("Expected trigger 'status', got '%s'", chart.Expose.HTTP.Events[1].Trigger)
	}
	if chart.Expose.HTTP.Events[1].Method != "GET" {
		t.Errorf("Expected method 'GET', got '%s'", chart.Expose.HTTP.Events[1].Method)
	}
}
