package observability

import (
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/maelstrom/v3/pkg/services"
)

type MetricsReporter struct {
	mu        sync.Mutex
	collected services.MetricsCollector
	started   time.Time
}

func NewMetricsReporter() *MetricsReporter {
	return &MetricsReporter{
		collected: services.MetricsCollector{
			StateCounts: make(map[string]int),
			LastUpdate:  time.Now(),
		},
		started: time.Now(),
	}
}

func (r *MetricsReporter) Collect() services.MetricsCollector {
	r.mu.Lock()
	defer r.mu.Unlock()

	result := r.collected
	result.StateCounts = make(map[string]int)
	for k, v := range r.collected.StateCounts {
		result.StateCounts[k] = v
	}

	duration := time.Since(r.started).Seconds()
	if duration > 0 {
		result.TransitionRate = float64(r.collected.TransitionRate) / duration
		result.EventRate = float64(r.collected.EventRate) / duration
	}

	result.LastUpdate = time.Now()
	return result
}

func (r *MetricsReporter) Report(format string) error {
	metrics := r.Collect()

	switch format {
	case "json":
		return r.reportJSON(metrics)
	case "text", "":
		return r.reportText(metrics)
	default:
		return fmt.Errorf("unsupported format: %s", format)
	}
}

func (r *MetricsReporter) Reset() {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.collected = services.MetricsCollector{
		StateCounts: make(map[string]int),
		LastUpdate:  time.Now(),
	}
	r.started = time.Now()
}

func (r *MetricsReporter) reportJSON(metrics services.MetricsCollector) error {
	data, err := json.MarshalIndent(metrics, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal metrics: %w", err)
	}
	fmt.Println(string(data))
	return nil
}

func (r *MetricsReporter) reportText(metrics services.MetricsCollector) error {
	var sb strings.Builder

	sb.WriteString("=== Metrics Report ===\n")
	sb.WriteString(fmt.Sprintf("Generated: %s\n", metrics.LastUpdate.Format(time.RFC3339)))
	sb.WriteString("\n--- State Counts ---\n")
	for state, count := range metrics.StateCounts {
		sb.WriteString(fmt.Sprintf("  %s: %d\n", state, count))
	}
	sb.WriteString("\n--- Rates (per second) ---\n")
	sb.WriteString(fmt.Sprintf("  Transition Rate: %.2f\n", metrics.TransitionRate))
	sb.WriteString(fmt.Sprintf("  Event Rate: %.2f\n", metrics.EventRate))
	sb.WriteString("\n--- Mail Delivery ---\n")
	sb.WriteString(fmt.Sprintf("  Delivered: %d\n", metrics.MailDelivered))
	sb.WriteString(fmt.Sprintf("  Failed: %d\n", metrics.MailFailed))
	sb.WriteString(fmt.Sprintf("  Retried: %d\n", metrics.MailRetried))

	fmt.Print(sb.String())
	return nil
}
