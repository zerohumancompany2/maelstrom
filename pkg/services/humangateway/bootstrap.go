package humangateway

// BootstrapChart returns the chart definition for the human gateway service
func BootstrapChart() map[string]interface{} {
	return map[string]interface{}{
		"name":        "sys:human-gateway",
		"description": "Chat interface for human-in-the-loop with running agents",
		"sessions":    true,
	}
}
