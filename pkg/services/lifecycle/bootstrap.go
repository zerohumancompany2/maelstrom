package lifecycle

import "github.com/maelstrom/v3/pkg/statechart"

func BootstrapChart() statechart.ChartDefinition {
	return statechart.ChartDefinition{
		ID:      "sys:lifecycle",
		Version: "1.0.0",
	}
}
