package runtime

import "fmt"

// SendEventToChart sends an event to a specific runtime by ID.
func SendEventToChart(runtimes map[string]*ChartRuntime, chartID string, evt Event) error {
	rt, ok := runtimes[chartID]
	if !ok {
		return fmt.Errorf("runtime %s not found", chartID)
	}
	return rt.SendEvent(evt)
}

// FindOrphanedCharts identifies charts in registry without runtimes.
func FindOrphanedCharts(runtimes map[string]*ChartRuntime, registryKeys []string) []string {
	var orphaned []string
	for _, key := range registryKeys {
		if _, exists := runtimes[key]; !exists {
			orphaned = append(orphaned, key)
		}
	}
	return orphaned
}

// FindStaleRuntimes identifies runtimes not in registry.
func FindStaleRuntimes(runtimes map[string]*ChartRuntime, registryKeys []string) []string {
	var stale []string
	for key := range runtimes {
		found := false
		for _, regKey := range registryKeys {
			if regKey == key {
				found = true
				break
			}
		}
		if !found {
			stale = append(stale, key)
		}
	}
	return stale
}
