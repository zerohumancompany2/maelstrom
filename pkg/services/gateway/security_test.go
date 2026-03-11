package gateway

import (
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/maelstrom/v3/pkg/security/sanitizers"
)

func TestGatewaySecurity_AllDataTaintedAtBorder(t *testing.T) {
	tainter := &BorderTainter{
		DefaultTaints: []string{"USER_SUPPLIED", "OUTER_BOUNDARY"},
	}

	// All data entering the runtime is tainted at the border (arch-v1.md L276)
	testData := map[string]any{
		"message": "Hello, agent!",
		"source":  "user:web",
	}

	taintedData, err := tainter.TaintInboundData(testData, "user:web")
	if err != nil {
		t.Fatalf("Expected no error tainting inbound data, got %v", err)
	}

	// Verify data is tainted (arch-v1.md L276)
	taintedMap, ok := taintedData.(map[string]any)
	if !ok {
		t.Fatal("Expected tainted data to be map")
	}

	if taintedMap["taints"] == nil {
		t.Error("Expected taints to be attached to data")
	}

	// Verify taints are attached as soon as data is ingested (arch-v1.md L276)
	taints := taintedMap["taints"].([]string)
	if !slices.Contains(taints, "USER_SUPPLIED") {
		t.Error("Expected USER_SUPPLIED taint")
	}

	if !slices.Contains(taints, "OUTER_BOUNDARY") {
		t.Error("Expected OUTER_BOUNDARY taint")
	}

	// No untainted information exists inside the runtime (arch-v1.md L276)
	if len(taints) == 0 {
		t.Error("Expected all data to have at least one taint")
	}

	// Test with different source types
	testCases := []struct {
		source        string
		expectedTaint string
	}{
		{"user:web", "USER_SUPPLIED"},
		{"user:api", "USER_SUPPLIED"},
		{"system:heartbeat", "SYSTEM_GENERATED"},
		{"tool:registry", "TOOL_OUTPUT"},
	}

	for _, tc := range testCases {
		taintedData, err := tainter.TaintInboundData(map[string]any{"data": "test"}, tc.source)
		if err != nil {
			t.Fatalf("Expected no error for source %s, got %v", tc.source, err)
		}

		taintedMap := taintedData.(map[string]any)
		taints := taintedMap["taints"].([]string)

		if !slices.Contains(taints, tc.expectedTaint) {
			t.Errorf("Expected taint %s for source %s", tc.expectedTaint, tc.source)
		}
	}
}

func TestGatewaySecurity_StreamSanitizationPerChunk(t *testing.T) {
	sanitizer := &sanitizers.StreamSanitizer{
		Redactor:          sanitizers.NewPIIRedactor(),
		LengthCapper:      sanitizers.NewLengthCapper(1000),
		SchemaValidator:   sanitizers.NewSchemaValidator(),
		InnerDataStripper: sanitizers.NewInnerDataStripper(),
	}

	// Every outgoing chunk passes through DMZ sanitizers (arch-v1.md L298)
	chunks := []sanitizers.StreamChunk{
		{
			Chunk:    "Hello, this is a normal message",
			Sequence: 1,
			IsFinal:  false,
			Taints:   []string{"USER_SUPPLIED"},
		},
		{
			Chunk:    "PII data: John Doe, email: john@example.com",
			Sequence: 2,
			IsFinal:  false,
			Taints:   []string{"PII"},
		},
		{
			Chunk:    "SECRET: api_key=abc123",
			Sequence: 3,
			IsFinal:  false,
			Taints:   []string{"SECRET", "INNER_ONLY"},
		},
		{
			Chunk:    "Final response complete",
			Sequence: 4,
			IsFinal:  true,
			Taints:   []string{},
		},
	}

	// Sanitization is per-chunk (stateless) (arch-v1.md L299)
	for i, chunk := range chunks {
		startTime := time.Now()
		sanitized, err := sanitizer.SanitizeChunk(chunk)
		elapsed := time.Since(startTime)

		if err != nil {
			t.Fatalf("Expected no error sanitizing chunk %d, got %v", i, err)
		}

		// Latency stays <50ms (arch-v1.md L299)
		if elapsed >= 50*time.Millisecond {
			t.Errorf("Expected chunk %d sanitization <50ms, got %v", i, elapsed)
		}

		// Verify PII redaction (arch-v1.md L298)
		if chunk.Taints != nil && slices.Contains(chunk.Taints, "PII") {
			if strings.Contains(sanitized.Chunk, "john@example.com") {
				t.Error("Expected PII to be redacted")
			}
		}

		// Verify inner-data stripping (arch-v1.md L298)
		if slices.Contains(chunk.Taints, "INNER_ONLY") {
			if strings.Contains(sanitized.Chunk, "api_key") {
				t.Error("Expected inner data to be stripped")
			}
		}

		// Verify length caps (arch-v1.md L298)
		if len(sanitized.Chunk) > 1000 {
			t.Error("Expected chunk to be capped at 1000 chars")
		}

		// Verify schema validation (arch-v1.md L298)
		if sanitized.Sequence != chunk.Sequence {
			t.Errorf("Expected sequence %d, got %d", chunk.Sequence, sanitized.Sequence)
		}

		if sanitized.IsFinal != chunk.IsFinal {
			t.Errorf("Expected IsFinal %v, got %v", chunk.IsFinal, sanitized.IsFinal)
		}
	}

	// No buffering of entire response unless chart explicitly requests it (arch-v1.md L299)
	// Verify each chunk is sanitized independently (stateless)
	_, _ = sanitizer.SanitizeChunk(chunks[0])
	secondChunk, _ := sanitizer.SanitizeChunk(chunks[1])

	// Sanitizing chunk 2 should not depend on chunk 1
	if secondChunk.Sequence != 2 {
		t.Error("Expected chunk 2 to be sanitized independently")
	}
}
