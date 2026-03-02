package gateway

// BootstrapChart returns the chart definition for the gateway service
func BootstrapChart() map[string]interface{} {
	return map[string]interface{}{
		"name":        "sys:gateway",
		"description": "Channel adapters (HTTP/SSE/WS/Email) for external I/O",
		"adapters":    []string{"webhook", "websocket", "sse", "smtp"},
	}
}
