package observability

import "github.com/maelstrom/v3/pkg/statechart"

func BootstrapChart() statechart.ChartDefinition {
	return statechart.ChartDefinition{
		ID:      "sys:observability",
		Version: "1.0.0",
	}
}
