package gateway

import (
	"net/http"
	"net/http/httptest"
	"slices"
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

func TestHTTPExposure_AuthMiddlewareApplied(t *testing.T) {
	svc := NewGatewayService()
	authMiddleware := NewAuthMiddleware()

	// Setup: Create handler with auth middleware
	baseHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	// Apply auth middleware (arch-v1.md L722)
	protectedHandler := authMiddleware.Apply(baseHandler)

	// Test without auth header
	req := httptest.NewRequest("GET", "/api/v1/agents/test/", nil)
	rr := httptest.NewRecorder()

	protectedHandler.ServeHTTP(rr, req)

	// Verify 401 Unauthorized (arch-v1.md L722)
	if rr.Code != http.StatusUnauthorized {
		t.Errorf("Expected status 401, got %d", rr.Code)
	}

	// Test with valid auth header
	req2 := httptest.NewRequest("GET", "/api/v1/agents/test/", nil)
	req2.Header.Set("Authorization", "Bearer valid-token-123")
	rr2 := httptest.NewRecorder()

	protectedHandler.ServeHTTP(rr2, req2)

	// Verify 200 OK with valid auth
	if rr2.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", rr2.Code)
	}

	// Verify sys:gateway registers all endpoints with auth middleware (arch-v1.md L722)
	charts := []Chart{
		{
			Name:     "agent:dmz",
			Boundary: "dmz",
			Expose: &Exposure{
				HTTP: &HTTPExposure{
					Path: "/api/v1/agents/{id}/",
					Events: []HTTPEvent{
						{Trigger: "user_query", Method: "POST"},
						{Trigger: "status", Method: "GET"},
					},
				},
			},
		},
	}

	// Register endpoints with auth middleware
	err := svc.RegisterEndpoints(charts)
	if err != nil {
		t.Fatalf("Expected no error registering endpoints, got %v", err)
	}

	// Verify all endpoints have auth middleware applied
	gwSvc := svc.(*gatewayService)
	if len(gwSvc.protectedEndpoints) != 2 {
		t.Errorf("Expected 2 protected endpoints, got %d", len(gwSvc.protectedEndpoints))
	}

	// Verify both POST and GET are protected
	protectedMethods := []string{}
	for method := range gwSvc.protectedEndpoints {
		protectedMethods = append(protectedMethods, method)
	}

	if !slices.Contains(protectedMethods, "POST /api/v1/agents/{id}/") {
		t.Error("Expected POST /api/v1/agents/{id}/ to be protected")
	}
	if !slices.Contains(protectedMethods, "GET /api/v1/agents/{id}/") {
		t.Error("Expected GET /api/v1/agents/{id}/ to be protected")
	}
}

func TestHTTPExposure_EventSurfaceIsAPISurface(t *testing.T) {
	svc := NewGatewayService()

	// Setup: Chart with events that become API surface (arch-v1.md L722)
	chart := Chart{
		Name:     "agent:dmz",
		Boundary: "dmz",
		States: []State{
			{
				Name: "Idle",
				On: map[string]Transition{
					"user_query": {To: "Processing"},
					"status":     {To: "Idle"},
				},
			},
			{
				Name: "Processing",
				On: map[string]Transition{
					"tool_result": {To: "Idle"},
				},
			},
		},
		Expose: &Exposure{
			HTTP: &HTTPExposure{
				Path: "/api/v1/agents/{id}/",
				Events: []HTTPEvent{
					{Trigger: "user_query", Method: "POST"},
					{Trigger: "status", Method: "GET"},
				},
			},
		},
	}

	// The event surface of the chart becomes the API surface (arch-v1.md L722)
	endpoints, err := svc.MapEventToAPI(chart)
	if err != nil {
		t.Fatalf("Expected no error mapping events to API, got %v", err)
	}

	// Verify API endpoints match event triggers
	if len(endpoints) != 2 {
		t.Errorf("Expected 2 API endpoints, got %d", len(endpoints))
	}

	// Verify user_query event maps to POST endpoint
	var userQueryEndpoint *APIEndpoint
	var statusEndpoint *APIEndpoint
	for i := range endpoints {
		ep := &endpoints[i]
		if ep.Trigger == "user_query" {
			userQueryEndpoint = ep
		}
		if ep.Trigger == "status" {
			statusEndpoint = ep
		}
	}

	if userQueryEndpoint == nil {
		t.Error("Expected user_query endpoint")
	} else {
		if userQueryEndpoint.Method != "POST" {
			t.Errorf("Expected user_query to map to POST, got %s", userQueryEndpoint.Method)
		}
		if userQueryEndpoint.Path != "/api/v1/agents/{id}/" {
			t.Errorf("Expected path '/api/v1/agents/{id}/', got '%s'", userQueryEndpoint.Path)
		}
	}

	if statusEndpoint == nil {
		t.Error("Expected status endpoint")
	} else {
		if statusEndpoint.Method != "GET" {
			t.Errorf("Expected status to map to GET, got %s", statusEndpoint.Method)
		}
		if statusEndpoint.Path != "/api/v1/agents/{id}/" {
			t.Errorf("Expected path '/api/v1/agents/{id}/', got '%s'", statusEndpoint.Path)
		}
	}

	// Verify internal events NOT exposed (tool_result not in expose block)
	var toolResultEndpoint *APIEndpoint
	for i := range endpoints {
		ep := &endpoints[i]
		if ep.Trigger == "tool_result" {
			toolResultEndpoint = ep
			break
		}
	}

	if toolResultEndpoint != nil {
		t.Error("Expected tool_result event NOT to be exposed (not in expose block)")
	}
}

func TestHTTPExposure_InnerBoundaryNotExposed(t *testing.T) {
	svc := NewGatewayService()

	// Setup: Inner-boundary chart with expose block (arch-v1.md L723)
	innerChart := Chart{
		Name:     "orchestrator:inner",
		Boundary: "inner",
		Expose: &Exposure{
			HTTP: &HTTPExposure{
				Path: "/api/v1/orchestrator/",
				Events: []HTTPEvent{
					{Trigger: "execute", Method: "POST"},
				},
			},
		},
	}

	// Inner-boundary Charts are never directly exposed (arch-v1.md L723)
	canExpose := svc.CanExpose(innerChart)
	if canExpose {
		t.Error("Expected inner-boundary chart to NOT be exposable")
	}

	// Verify inner chart not registered even with expose block
	charts := []Chart{innerChart}
	err := svc.RegisterEndpoints(charts)
	if err != nil {
		t.Fatalf("Expected no error registering endpoints, got %v", err)
	}

	// Verify no endpoints registered for inner chart
	gwSvc := svc.(*gatewayService)
	if len(gwSvc.protectedEndpoints) != 0 {
		t.Errorf("Expected 0 endpoints for inner chart, got %d", len(gwSvc.protectedEndpoints))
	}

	// Verify OpenAPI spec excludes inner chart
	gen := NewOpenAPIGen()
	spec, err := gen.GenerateSpec(charts)
	if err != nil {
		t.Fatalf("Expected no error generating spec, got %v", err)
	}

	// Verify /api/v1/orchestrator/ NOT in spec (arch-v1.md L723)
	if _, exists := spec.Paths["/api/v1/orchestrator/"]; exists {
		t.Error("Expected inner-boundary path NOT in OpenAPI spec")
	}

	// Verify spec has no paths from inner chart
	if len(spec.Paths) != 0 {
		t.Errorf("Expected 0 paths in spec for inner chart, got %d", len(spec.Paths))
	}
}

func TestHTTPExposure_DMZOuterOnlyExposed(t *testing.T) {
	svc := NewGatewayService()
	gen := NewOpenAPIGen()

	// Setup: Charts with different boundaries (arch-v1.md L723)
	charts := []Chart{
		{
			Name:     "agent:dmz",
			Boundary: "dmz",
			Expose: &Exposure{
				HTTP: &HTTPExposure{
					Path: "/api/v1/agents/{id}/",
					Events: []HTTPEvent{
						{Trigger: "user_query", Method: "POST"},
					},
				},
			},
		},
		{
			Name:     "gateway:outer",
			Boundary: "outer",
			Expose: &Exposure{
				HTTP: &HTTPExposure{
					Path: "/api/v1/gateway/",
					Events: []HTTPEvent{
						{Trigger: "health", Method: "GET"},
					},
				},
			},
		},
		{
			Name:     "orchestrator:inner",
			Boundary: "inner",
			Expose: &Exposure{
				HTTP: &HTTPExposure{
					Path: "/api/v1/orchestrator/",
					Events: []HTTPEvent{
						{Trigger: "execute", Method: "POST"},
					},
				},
			},
		},
	}

	// Only DMZ/outer ones are exposed (arch-v1.md L723)
	err := svc.RegisterEndpoints(charts)
	if err != nil {
		t.Fatalf("Expected no error registering endpoints, got %v", err)
	}

	// Verify only 2 endpoints registered (dmz + outer, not inner)
	gwSvc := svc.(*gatewayService)
	if len(gwSvc.protectedEndpoints) != 2 {
		t.Errorf("Expected 2 endpoints (dmz + outer), got %d", len(gwSvc.protectedEndpoints))
	}

	// Verify DMZ endpoint registered
	if !gwSvc.protectedEndpoints["POST /api/v1/agents/{id}/"] {
		t.Error("Expected DMZ endpoint to be registered")
	}

	// Verify outer endpoint registered
	if !gwSvc.protectedEndpoints["GET /api/v1/gateway/"] {
		t.Error("Expected outer endpoint to be registered")
	}

	// Verify inner endpoint NOT registered
	if gwSvc.protectedEndpoints["POST /api/v1/orchestrator/"] {
		t.Error("Expected inner endpoint NOT to be registered")
	}

	// Generate OpenAPI spec
	spec, err := gen.GenerateSpec(charts)
	if err != nil {
		t.Fatalf("Expected no error generating spec, got %v", err)
	}

	// Verify spec contains only DMZ and outer paths (arch-v1.md L723)
	if len(spec.Paths) != 2 {
		t.Errorf("Expected 2 paths in spec (dmz + outer), got %d", len(spec.Paths))
	}

	// Verify DMZ path exists
	if _, exists := spec.Paths["/api/v1/agents/{id}/"]; !exists {
		t.Error("Expected DMZ path in spec")
	}

	// Verify outer path exists
	if _, exists := spec.Paths["/api/v1/gateway/"]; !exists {
		t.Error("Expected outer path in spec")
	}

	// Verify inner path NOT in spec
	if _, exists := spec.Paths["/api/v1/orchestrator/"]; exists {
		t.Error("Expected inner path NOT in spec")
	}
}
