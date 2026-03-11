package gateway

import (
	"testing"
)

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
